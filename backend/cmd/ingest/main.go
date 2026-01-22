package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gofrs/uuid/v5"
	_ "github.com/lib/pq"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/emailreminder"
	"github.com/rj-davidson/greenrats/ent/golfer"
	"github.com/rj-davidson/greenrats/ent/league"
	"github.com/rj-davidson/greenrats/ent/leaguemembership"
	"github.com/rj-davidson/greenrats/ent/pick"
	"github.com/rj-davidson/greenrats/ent/syncstatus"
	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/ent/tournamententry"
	"github.com/rj-davidson/greenrats/ent/user"
	"github.com/rj-davidson/greenrats/internal/config"
	"github.com/rj-davidson/greenrats/internal/email"
	"github.com/rj-davidson/greenrats/internal/external/balldontlie"
	"github.com/rj-davidson/greenrats/internal/external/googlemaps"
	"github.com/rj-davidson/greenrats/internal/external/pgatour"
	"github.com/rj-davidson/greenrats/internal/external/scrapedo"
	"github.com/rj-davidson/greenrats/internal/features/fields"
	"github.com/rj-davidson/greenrats/internal/sync"
)

const (
	tournamentSyncInterval  = 24 * time.Hour
	leaderboardSyncInterval = 1 * time.Hour
	playerSyncInterval      = 7 * 24 * time.Hour
	reminderCheckInterval   = 1 * time.Hour
	earningsCheckInterval   = 6 * time.Hour
	fieldSyncInterval       = 12 * time.Hour
)

type Ingester struct {
	db            *ent.Client
	config        *config.Config
	ballDontLie   *balldontlie.Client
	pgatourClient *pgatour.Client
	syncService   *sync.Service
	fieldsService *fields.Service
	majorScraper  *fields.MajorScraper
	email         *email.Client
	logger        *slog.Logger
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logLevel := slog.LevelInfo
	if cfg.IsDevelopment() {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	if cfg.SentryDSN != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              cfg.SentryDSN,
			Environment:      cfg.Env,
			EnableTracing:    true,
			TracesSampleRate: 0.1,
		}); err != nil {
			log.Printf("Sentry initialization failed: %v", err)
		} else {
			defer sentry.Flush(2 * time.Second)
			log.Println("Sentry initialized")
		}
	}

	db, err := ent.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	bdlClient := balldontlie.New(cfg.BallDontLieAPIKey, cfg.BallDontLieBaseURL, logger)
	emailClient := email.New(cfg)
	pgatourClient := pgatour.New(cfg.PGATourAPIKey, logger)
	scrapeDoClient := scrapedo.New(cfg.ScrapeDoAPIKey, logger)
	fieldsService := fields.NewService(db, pgatourClient, logger)
	majorScraper := fields.NewMajorScraper(scrapeDoClient, logger)

	var gmapsClient *googlemaps.Client
	if cfg.GoogleMapsAPIKey != "" {
		var err error
		gmapsClient, err = googlemaps.New(cfg.GoogleMapsAPIKey, logger)
		if err != nil {
			log.Printf("Warning: failed to create Google Maps client: %v", err)
		}
	}

	syncService := sync.NewService(db, gmapsClient, logger)

	ingester := &Ingester{
		db:            db,
		config:        cfg,
		ballDontLie:   bdlClient,
		pgatourClient: pgatourClient,
		syncService:   syncService,
		fieldsService: fieldsService,
		majorScraper:  majorScraper,
		email:         emailClient,
		logger:        logger,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go ingester.runScheduledJobs(ctx)

	log.Println("GreenRats Ingest Service started")
	log.Printf("Sync intervals: tournaments=%v, leaderboards=%v, players=%v, fields=%v",
		tournamentSyncInterval, leaderboardSyncInterval, playerSyncInterval, fieldSyncInterval)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down ingest service...")
	cancel()

	time.Sleep(2 * time.Second)
	log.Println("Ingest service exited gracefully")
	return nil
}

func (i *Ingester) runScheduledJobs(ctx context.Context) {
	tournamentTicker := time.NewTicker(tournamentSyncInterval)
	defer tournamentTicker.Stop()

	leaderboardTicker := time.NewTicker(leaderboardSyncInterval)
	defer leaderboardTicker.Stop()

	playerTicker := time.NewTicker(playerSyncInterval)
	defer playerTicker.Stop()

	reminderTicker := time.NewTicker(reminderCheckInterval)
	defer reminderTicker.Stop()

	earningsTicker := time.NewTicker(earningsCheckInterval)
	defer earningsTicker.Stop()

	fieldTicker := time.NewTicker(fieldSyncInterval)
	defer fieldTicker.Stop()

	log.Println("Checking for required initial syncs...")
	if i.shouldSync(ctx, "tournaments", tournamentSyncInterval) {
		i.syncTournaments(ctx)
	} else {
		log.Println("Skipping tournament sync (data is fresh)")
	}
	if i.shouldSync(ctx, "players", playerSyncInterval) {
		i.syncPlayers(ctx)
	} else {
		log.Println("Skipping player sync (data is fresh)")
	}
	if i.shouldSync(ctx, "leaderboards", leaderboardSyncInterval) {
		i.syncLeaderboards(ctx)
	} else {
		log.Println("Skipping leaderboard sync (data is fresh)")
	}
	if i.shouldSync(ctx, "earnings", earningsCheckInterval) {
		i.syncEarnings(ctx)
	} else {
		log.Println("Skipping earnings sync (data is fresh)")
	}
	if i.shouldSync(ctx, "fields", fieldSyncInterval) {
		i.syncFields(ctx)
	} else {
		log.Println("Skipping field sync (data is fresh)")
	}
	i.sendPickReminders(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-tournamentTicker.C:
			i.syncTournaments(ctx)
		case <-leaderboardTicker.C:
			i.syncLeaderboards(ctx)
		case <-playerTicker.C:
			i.syncPlayers(ctx)
		case <-earningsTicker.C:
			i.syncEarnings(ctx)
		case <-fieldTicker.C:
			i.syncFields(ctx)
		case <-reminderTicker.C:
			i.sendPickReminders(ctx)
		}
	}
}

func (i *Ingester) captureJobError(job string, err error) {
	if err == nil {
		return
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetTag("job", job)
		sentry.CaptureException(err)
	})
}

func (i *Ingester) shouldSync(ctx context.Context, syncType string, interval time.Duration) bool {
	status, err := i.db.SyncStatus.Query().
		Where(syncstatus.SyncTypeEQ(syncType)).
		Only(ctx)

	if ent.IsNotFound(err) {
		return true
	}
	if err != nil {
		log.Printf("failed to check sync status for %s: %v", syncType, err)
		return true
	}

	return time.Now().After(status.LastSyncAt.Add(interval))
}

func (i *Ingester) recordSync(ctx context.Context, syncType string) {
	now := time.Now()

	existing, err := i.db.SyncStatus.Query().
		Where(syncstatus.SyncTypeEQ(syncType)).
		Only(ctx)

	if ent.IsNotFound(err) {
		_, err = i.db.SyncStatus.Create().
			SetSyncType(syncType).
			SetLastSyncAt(now).
			Save(ctx)
	} else if err == nil {
		_, err = existing.Update().
			SetLastSyncAt(now).
			Save(ctx)
	}

	if err != nil {
		log.Printf("failed to record sync for %s: %v", syncType, err)
	}
}

func (i *Ingester) syncTournaments(ctx context.Context) {
	log.Println("Starting tournament sync...")

	tournaments, err := i.ballDontLie.GetTournaments(ctx, i.config.CurrentSeason)
	if err != nil {
		log.Printf("failed to fetch tournaments from BallDontLie: %v", err)
		i.captureJobError("sync_tournaments", err)
		return
	}

	log.Printf("Fetched %d tournaments for %d season", len(tournaments), i.config.CurrentSeason)

	for idx := range tournaments {
		result, err := i.syncService.UpsertTournament(ctx, &tournaments[idx])
		if err != nil {
			log.Printf("failed to upsert tournament %s: %v", tournaments[idx].Name, err)
			i.captureJobError("sync_tournaments", err)
			continue
		}

		if result.Created {
			log.Printf("Created tournament: %s", tournaments[idx].Name)
		} else {
			log.Printf("Updated tournament: %s (status: %s)", tournaments[idx].Name, result.CurrentStatus)
			if result.PreviousStatus != tournament.StatusCompleted && result.CurrentStatus == tournament.StatusCompleted {
				i.sendTournamentResultsEmails(ctx, result.Tournament)
			}
		}
	}

	i.recordSync(ctx, "tournaments")
	log.Println("Tournament sync completed")
}

func (i *Ingester) syncPlayers(ctx context.Context) {
	log.Println("Starting player sync...")

	players, err := i.ballDontLie.GetPlayers(ctx)
	if err != nil {
		log.Printf("failed to fetch players from BallDontLie: %v", err)
		i.captureJobError("sync_players", err)
		return
	}

	log.Printf("Fetched %d players", len(players))

	for idx := range players {
		if err := i.syncService.UpsertPlayer(ctx, &players[idx]); err != nil {
			log.Printf("failed to upsert player %s: %v", players[idx].DisplayName, err)
			i.captureJobError("sync_players", err)
			continue
		}
	}

	i.recordSync(ctx, "players")
	log.Println("Player sync completed")
}

func (i *Ingester) syncLeaderboards(ctx context.Context) {
	log.Println("Checking for active tournament leaderboards...")

	tournaments, err := i.db.Tournament.Query().
		Where(tournament.StatusEQ(tournament.StatusActive)).
		All(ctx)
	if err != nil {
		log.Printf("failed to query active tournaments: %v", err)
		i.captureJobError("sync_leaderboards", err)
		return
	}

	if len(tournaments) == 0 {
		log.Println("No active tournaments found")
		return
	}

	for _, t := range tournaments {
		if t.BdlID == nil {
			log.Printf("Tournament %s has no BallDontLie ID, skipping leaderboard sync", t.Name)
			continue
		}

		if err := i.syncTournamentLeaderboard(ctx, t); err != nil {
			log.Printf("failed to sync leaderboard for tournament %s: %v", t.Name, err)
			i.captureJobError("sync_leaderboards", err)
		}
	}

	i.recordSync(ctx, "leaderboards")
	log.Println("Leaderboard sync completed")
}

func (i *Ingester) syncTournamentLeaderboard(ctx context.Context, t *ent.Tournament) error {
	log.Printf("Syncing leaderboard for tournament: %s", t.Name)

	results, err := i.ballDontLie.GetTournamentResults(ctx, *t.BdlID)
	if err != nil {
		return fmt.Errorf("failed to fetch tournament results: %w", err)
	}

	log.Printf("Fetched %d results for tournament %s", len(results), t.Name)

	for idx := range results {
		if err := i.syncService.UpsertTournamentEntry(ctx, t, &results[idx]); err != nil {
			log.Printf("failed to upsert result for player %s: %v", results[idx].Player.DisplayName, err)
			i.captureJobError("sync_tournament_leaderboard", err)
			continue
		}
	}

	return nil
}

func (i *Ingester) syncEarnings(ctx context.Context) {
	log.Println("Checking for tournaments needing earnings sync...")

	now := time.Now().UTC()
	currentYear := now.Year()

	tournaments, err := i.db.Tournament.Query().
		Where(
			tournament.StatusEQ(tournament.StatusCompleted),
			tournament.SeasonYearGTE(currentYear-1),
		).
		All(ctx)
	if err != nil {
		log.Printf("failed to query completed tournaments: %v", err)
		i.captureJobError("sync_earnings", err)
		return
	}

	if len(tournaments) == 0 {
		log.Println("No completed tournaments found")
		return
	}

	for _, t := range tournaments {
		if t.BdlID == nil {
			log.Printf("Tournament %s has no BallDontLie ID, skipping earnings sync", t.Name)
			continue
		}

		hasEarnings, err := i.tournamentHasEarnings(ctx, t)
		if err != nil {
			log.Printf("failed to check earnings for tournament %s: %v", t.Name, err)
			i.captureJobError("sync_earnings", err)
			continue
		}

		daysSinceEnd := int(now.Sub(t.EndDate).Hours() / 24)
		shouldSync := !hasEarnings || daysSinceEnd == 1 || daysSinceEnd == 2 || daysSinceEnd == 7

		if shouldSync {
			if !hasEarnings {
				log.Printf("Tournament %s missing earnings, syncing...", t.Name)
			} else {
				log.Printf("Refreshing earnings for recently completed tournament %s (day %d)", t.Name, daysSinceEnd)
			}

			if err := i.syncTournamentLeaderboard(ctx, t); err != nil {
				log.Printf("failed to sync earnings for tournament %s: %v", t.Name, err)
				i.captureJobError("sync_earnings", err)
			}
		}
	}

	i.recordSync(ctx, "earnings")
	log.Println("Earnings sync completed")
}

func (i *Ingester) tournamentHasEarnings(ctx context.Context, t *ent.Tournament) (bool, error) {
	return i.db.TournamentEntry.Query().
		Where(
			tournamententry.HasTournamentWith(tournament.IDEQ(t.ID)),
			tournamententry.EarningsGT(0),
		).
		Exist(ctx)
}

func (i *Ingester) syncFields(ctx context.Context) {
	log.Println("Checking for tournaments needing field sync...")

	now := time.Now().UTC()
	windowStart := now
	windowEnd := now.Add(7 * 24 * time.Hour)

	tournaments, err := i.db.Tournament.Query().
		Where(
			tournament.StatusEQ(tournament.StatusUpcoming),
			tournament.StartDateGTE(windowStart),
			tournament.StartDateLTE(windowEnd),
		).
		All(ctx)
	if err != nil {
		log.Printf("failed to query upcoming tournaments: %v", err)
		i.captureJobError("sync_fields", err)
		return
	}

	if len(tournaments) == 0 {
		log.Println("No upcoming tournaments in sync window")
		i.recordSync(ctx, "fields")
		return
	}

	for _, t := range tournaments {
		if t.PgaTourID == nil || *t.PgaTourID == "" {
			log.Printf("Tournament %s has no PGA Tour ID, skipping field sync", t.Name)
			continue
		}

		hasField, err := i.tournamentHasField(ctx, t)
		if err != nil {
			log.Printf("failed to check field for tournament %s: %v", t.Name, err)
			i.captureJobError("sync_fields", err)
			continue
		}

		if hasField {
			log.Printf("Tournament %s already has field data, skipping", t.Name)
			continue
		}

		log.Printf("Syncing field for tournament: %s", t.Name)
		result, err := i.fieldsService.SyncTournamentField(ctx, t.ID)
		if err != nil {
			log.Printf("failed to sync field for tournament %s: %v", t.Name, err)
			i.captureJobError("sync_fields", err)
			continue
		}

		log.Printf("Field sync for %s: total=%d, matched=%d, new=%d, updated=%d",
			t.Name, result.TotalPlayers, result.MatchedPlayers, result.NewEntries, result.UpdatedEntries)
	}

	i.recordSync(ctx, "fields")
	log.Println("Field sync completed")
}

func (i *Ingester) tournamentHasField(ctx context.Context, t *ent.Tournament) (bool, error) {
	count, err := i.db.TournamentEntry.Query().
		Where(tournamententry.HasTournamentWith(tournament.IDEQ(t.ID))).
		Count(ctx)
	if err != nil {
		return false, err
	}
	return count >= 50, nil
}

func (i *Ingester) sendPickReminders(ctx context.Context) {
	log.Println("Checking for pick reminders to send...")

	now := time.Now().UTC()
	reminderWindowStart := now.Add(24 * time.Hour)
	reminderWindowEnd := now.Add(27 * time.Hour)

	tournaments, err := i.db.Tournament.Query().
		Where(
			tournament.StatusEQ(tournament.StatusUpcoming),
			tournament.StartDateGTE(reminderWindowStart),
			tournament.StartDateLTE(reminderWindowEnd),
		).
		All(ctx)
	if err != nil {
		log.Printf("failed to query upcoming tournaments: %v", err)
		i.captureJobError("send_pick_reminders", err)
		return
	}

	if len(tournaments) == 0 {
		log.Println("No tournaments starting within reminder window")
		return
	}

	for _, t := range tournaments {
		i.sendRemindersForTournament(ctx, t)
	}

	log.Println("Pick reminder check completed")
}

func (i *Ingester) sendRemindersForTournament(ctx context.Context, t *ent.Tournament) {
	log.Printf("Sending pick reminders for tournament: %s", t.Name)

	leagues, err := i.db.League.Query().
		Where(league.SeasonYearEQ(t.SeasonYear)).
		All(ctx)
	if err != nil {
		log.Printf("failed to query leagues: %v", err)
		i.captureJobError("send_pick_reminders", err)
		return
	}

	for _, l := range leagues {
		i.sendRemindersForLeagueTournament(ctx, l, t)
	}
}

func (i *Ingester) sendRemindersForLeagueTournament(ctx context.Context, l *ent.League, t *ent.Tournament) {
	memberships, err := i.db.LeagueMembership.Query().
		Where(leaguemembership.HasLeagueWith(league.IDEQ(l.ID))).
		WithUser().
		All(ctx)
	if err != nil {
		log.Printf("failed to query league memberships: %v", err)
		i.captureJobError("send_pick_reminders", err)
		return
	}

	for _, m := range memberships {
		if m.Edges.User == nil {
			continue
		}
		u := m.Edges.User

		if u.DisplayName == nil {
			continue
		}

		hasPick, err := i.db.Pick.Query().
			Where(
				pick.HasUserWith(user.IDEQ(u.ID)),
				pick.HasTournamentWith(tournament.IDEQ(t.ID)),
				pick.HasLeagueWith(league.IDEQ(l.ID)),
			).
			Exist(ctx)
		if err != nil {
			log.Printf("failed to check pick: %v", err)
			i.captureJobError("send_pick_reminders", err)
			continue
		}
		if hasPick {
			continue
		}

		alreadySent, err := i.db.EmailReminder.Query().
			Where(
				emailreminder.HasUserWith(user.IDEQ(u.ID)),
				emailreminder.HasTournamentWith(tournament.IDEQ(t.ID)),
				emailreminder.HasLeagueWith(league.IDEQ(l.ID)),
				emailreminder.ReminderTypeEQ(emailreminder.ReminderTypePickReminder),
			).
			Exist(ctx)
		if err != nil {
			log.Printf("failed to check reminder status: %v", err)
			i.captureJobError("send_pick_reminders", err)
			continue
		}
		if alreadySent {
			continue
		}

		if err := i.sendPickReminderEmail(ctx, u, l, t); err != nil {
			log.Printf("failed to send reminder to %s: %v", u.Email, err)
			i.captureJobError("send_pick_reminders", err)
			continue
		}

		_, err = i.db.EmailReminder.Create().
			SetUserID(u.ID).
			SetTournamentID(t.ID).
			SetLeagueID(l.ID).
			SetReminderType(emailreminder.ReminderTypePickReminder).
			Save(ctx)
		if err != nil {
			log.Printf("failed to record reminder: %v", err)
			i.captureJobError("send_pick_reminders", err)
		}
	}
}

func (i *Ingester) sendPickReminderEmail(ctx context.Context, u *ent.User, l *ent.League, t *ent.Tournament) error {
	displayName := ""
	if u.DisplayName != nil {
		displayName = *u.DisplayName
	}

	deadline := t.StartDate.Format("Monday, January 2 at 3:04 PM MST")

	data := email.PickReminderData{
		DisplayName:    displayName,
		TournamentName: t.Name,
		LeagueName:     l.Name,
		Deadline:       deadline,
	}

	return i.email.SendPickReminder(u.Email, data)
}

func (i *Ingester) sendTournamentResultsEmails(ctx context.Context, t *ent.Tournament) {
	log.Printf("Sending tournament results emails for: %s", t.Name)

	winnerEntry, err := i.db.TournamentEntry.Query().
		Where(
			tournamententry.HasTournamentWith(tournament.IDEQ(t.ID)),
			tournamententry.PositionEQ(1),
		).
		WithGolfer().
		First(ctx)

	tournamentWinner := "Unknown"
	if err == nil && winnerEntry.Edges.Golfer != nil {
		tournamentWinner = winnerEntry.Edges.Golfer.Name
	}

	picks, err := i.db.Pick.Query().
		Where(pick.HasTournamentWith(tournament.IDEQ(t.ID))).
		WithUser().
		WithGolfer().
		WithLeague().
		All(ctx)
	if err != nil {
		log.Printf("failed to query picks: %v", err)
		i.captureJobError("send_tournament_results", err)
		return
	}

	for _, p := range picks {
		if p.Edges.User == nil || p.Edges.League == nil || p.Edges.Golfer == nil {
			continue
		}
		if p.Edges.User.DisplayName == nil {
			continue
		}

		alreadySent, err := i.db.EmailReminder.Query().
			Where(
				emailreminder.HasUserWith(user.IDEQ(p.Edges.User.ID)),
				emailreminder.HasTournamentWith(tournament.IDEQ(t.ID)),
				emailreminder.HasLeagueWith(league.IDEQ(p.Edges.League.ID)),
				emailreminder.ReminderTypeEQ(emailreminder.ReminderTypeTournamentResults),
			).
			Exist(ctx)
		if err != nil {
			log.Printf("failed to check reminder status: %v", err)
			i.captureJobError("send_tournament_results", err)
			continue
		}
		if alreadySent {
			continue
		}

		golferEntry, _ := i.db.TournamentEntry.Query().
			Where(
				tournamententry.HasTournamentWith(tournament.IDEQ(t.ID)),
				tournamententry.HasGolferWith(golfer.IDEQ(p.Edges.Golfer.ID)),
			).
			Only(ctx)

		position := "N/A"
		earnings := "$0"
		if golferEntry != nil {
			if golferEntry.Cut {
				position = "CUT"
			} else if golferEntry.Position > 0 {
				position = fmt.Sprintf("%d", golferEntry.Position)
			}
			earnings = formatCurrency(golferEntry.Earnings)
		}

		userRank, totalEarnings := i.calculateLeagueStandings(ctx, p.Edges.User.ID, p.Edges.League.ID, t.SeasonYear)

		data := &email.TournamentResultsData{
			DisplayName:      *p.Edges.User.DisplayName,
			TournamentName:   t.Name,
			TournamentWinner: tournamentWinner,
			LeagueName:       p.Edges.League.Name,
			GolferName:       p.Edges.Golfer.Name,
			GolferPosition:   position,
			GolferEarnings:   earnings,
			UserRank:         userRank,
			TotalEarnings:    formatCurrency(totalEarnings),
		}

		if err := i.email.SendTournamentResults(p.Edges.User.Email, data); err != nil {
			log.Printf("failed to send results to %s: %v", p.Edges.User.Email, err)
			i.captureJobError("send_tournament_results", err)
			continue
		}

		_, err = i.db.EmailReminder.Create().
			SetUserID(p.Edges.User.ID).
			SetTournamentID(t.ID).
			SetLeagueID(p.Edges.League.ID).
			SetReminderType(emailreminder.ReminderTypeTournamentResults).
			Save(ctx)
		if err != nil {
			log.Printf("failed to record reminder: %v", err)
			i.captureJobError("send_tournament_results", err)
		}
	}

	log.Printf("Finished sending tournament results emails for: %s", t.Name)
}

func (i *Ingester) calculateLeagueStandings(ctx context.Context, userID, leagueID uuid.UUID, seasonYear int) (rank, totalEarnings int) {
	type userEarnings struct {
		userID   uuid.UUID
		earnings int
	}

	picks, err := i.db.Pick.Query().
		Where(
			pick.HasLeagueWith(league.IDEQ(leagueID)),
			pick.SeasonYearEQ(seasonYear),
		).
		WithUser().
		WithGolfer().
		WithTournament().
		All(ctx)
	if err != nil {
		return 0, 0
	}

	earningsMap := make(map[uuid.UUID]int)
	for _, p := range picks {
		if p.Edges.User == nil || p.Edges.Golfer == nil || p.Edges.Tournament == nil {
			continue
		}
		entry, err := i.db.TournamentEntry.Query().
			Where(
				tournamententry.HasTournamentWith(tournament.IDEQ(p.Edges.Tournament.ID)),
				tournamententry.HasGolferWith(golfer.IDEQ(p.Edges.Golfer.ID)),
			).
			Only(ctx)
		if err == nil {
			earningsMap[p.Edges.User.ID] += entry.Earnings
		}
	}

	var allEarnings []userEarnings
	for uid, e := range earningsMap {
		allEarnings = append(allEarnings, userEarnings{userID: uid, earnings: e})
	}

	for i := 0; i < len(allEarnings); i++ {
		for j := i + 1; j < len(allEarnings); j++ {
			if allEarnings[j].earnings > allEarnings[i].earnings {
				allEarnings[i], allEarnings[j] = allEarnings[j], allEarnings[i]
			}
		}
	}

	userTotal := earningsMap[userID]
	userRank := 1
	for _, e := range allEarnings {
		if e.userID == userID {
			break
		}
		if e.earnings > userTotal {
			userRank++
		}
	}

	return userRank, userTotal
}

func formatCurrency(amount int) string {
	if amount >= 1000000 {
		return fmt.Sprintf("$%.2fM", float64(amount)/1000000)
	}
	if amount >= 1000 {
		return fmt.Sprintf("$%dK", amount/1000)
	}
	return fmt.Sprintf("$%d", amount)
}
