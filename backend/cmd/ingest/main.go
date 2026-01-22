package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
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
)

// Sync intervals based on cost optimization plan:
// - Tournaments: Daily (24 hours) via BallDontLie
// - Leaderboards: Hourly during active tournaments via BallDontLie
// - Players: Weekly via BallDontLie
// - Earnings: Every 6 hours via BallDontLie (included in tournament results)
const (
	tournamentSyncInterval  = 24 * time.Hour
	leaderboardSyncInterval = 1 * time.Hour
	playerSyncInterval      = 7 * 24 * time.Hour // Weekly
	reminderCheckInterval   = 1 * time.Hour      // Check for pick reminders
	earningsCheckInterval   = 6 * time.Hour      // Check for earnings to sync
)

type Ingester struct {
	db          *ent.Client
	config      *config.Config
	ballDontLie *balldontlie.Client
	email       *email.Client
	logger      *slog.Logger
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

	ingester := &Ingester{
		db:          db,
		config:      cfg,
		ballDontLie: bdlClient,
		email:       emailClient,
		logger:      logger,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go ingester.runScheduledJobs(ctx)

	log.Println("GreenRats Ingest Service started")
	log.Printf("Sync intervals: tournaments=%v, leaderboards=%v, players=%v",
		tournamentSyncInterval, leaderboardSyncInterval, playerSyncInterval)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down ingest service...")
	cancel()

	time.Sleep(2 * time.Second)
	log.Println("Ingest service exited gracefully")
	return nil
}

// runScheduledJobs runs periodic data ingestion tasks.
func (i *Ingester) runScheduledJobs(ctx context.Context) {
	// Tournament sync - daily via BallDontLie
	tournamentTicker := time.NewTicker(tournamentSyncInterval)
	defer tournamentTicker.Stop()

	// Live leaderboard sync - hourly during active tournaments via BallDontLie
	leaderboardTicker := time.NewTicker(leaderboardSyncInterval)
	defer leaderboardTicker.Stop()

	// Player sync - weekly via BallDontLie
	playerTicker := time.NewTicker(playerSyncInterval)
	defer playerTicker.Stop()

	// Pick reminder check - hourly
	reminderTicker := time.NewTicker(reminderCheckInterval)
	defer reminderTicker.Stop()

	// Earnings check - every 6 hours to sync earnings for recently completed tournaments
	earningsTicker := time.NewTicker(earningsCheckInterval)
	defer earningsTicker.Stop()

	// Run initial syncs only if data is stale
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

// syncTournaments fetches and stores tournament data from BallDontLie.
// Runs daily to minimize API calls.
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
		if err := i.upsertTournament(ctx, &tournaments[idx]); err != nil {
			log.Printf("failed to upsert tournament %s: %v", tournaments[idx].Name, err)
			i.captureJobError("sync_tournaments", err)
			continue
		}
	}

	i.recordSync(ctx, "tournaments")
	log.Println("Tournament sync completed")
}

func parseEndDatePtr(endDateStr *string, startDate time.Time) time.Time {
	if endDateStr == nil || *endDateStr == "" {
		return startDate.AddDate(0, 0, 4).Add(6 * time.Hour)
	}
	return parseEndDate(*endDateStr, startDate)
}

func parseEndDate(endDateStr string, startDate time.Time) time.Time {
	const iso8601Millis = "2006-01-02T15:04:05.000Z"

	// Try ISO 8601 format first
	if endDate, err := time.Parse(iso8601Millis, endDateStr); err == nil {
		return endDate.AddDate(0, 0, 1).Add(6 * time.Hour)
	}

	// Try parsing display format like "Jan 15 - 18" or "Jan 15 - Feb 2"
	endDateStr = strings.TrimSpace(endDateStr)
	if idx := strings.LastIndex(endDateStr, " - "); idx != -1 {
		endPart := strings.TrimSpace(endDateStr[idx+3:])

		// Check if endPart is just a day number (e.g., "18")
		if day, err := strconv.Atoi(endPart); err == nil && day >= 1 && day <= 31 {
			endDate := time.Date(startDate.Year(), startDate.Month(), day, 0, 0, 0, 0, time.UTC)
			// Handle month rollover (e.g., start Jan 30, end day 2 means Feb 2)
			if endDate.Before(startDate) {
				endDate = endDate.AddDate(0, 1, 0)
			}
			return endDate.AddDate(0, 0, 1).Add(6 * time.Hour)
		}

		// Try parsing "Feb 2" format
		formats := []string{"Jan 2", "January 2"}
		for _, format := range formats {
			if parsed, err := time.Parse(format, endPart); err == nil {
				endDate := time.Date(startDate.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)
				// Handle year rollover (e.g., start Dec 30, end Jan 2)
				if endDate.Before(startDate) {
					endDate = endDate.AddDate(1, 0, 0)
				}
				return endDate.AddDate(0, 0, 1).Add(6 * time.Hour)
			}
		}
	}

	// Fallback: standard 4-day tournament (start + 4 days at 06:00 UTC)
	return startDate.AddDate(0, 0, 4).Add(6 * time.Hour)
}

// upsertTournament creates or updates a tournament from BallDontLie data.
func (i *Ingester) upsertTournament(ctx context.Context, t *balldontlie.Tournament) error {
	const iso8601Millis = "2006-01-02T15:04:05.000Z"
	startDate, err := time.Parse(iso8601Millis, t.StartDate)
	if err != nil {
		return fmt.Errorf("failed to parse start date: %w", err)
	}

	endDate := parseEndDatePtr(t.EndDate, startDate)

	// Determine status based on dates
	now := time.Now().UTC()
	status := tournament.StatusUpcoming
	if now.After(endDate) {
		status = tournament.StatusCompleted
	} else if now.After(startDate) || now.Equal(startDate) {
		status = tournament.StatusActive
	}

	// Try to find existing tournament by BallDontLie ID
	existing, err := i.db.Tournament.Query().
		Where(tournament.BdlID(t.ID)).
		Only(ctx)

	switch {
	case ent.IsNotFound(err):
		// Create new tournament
		builder := i.db.Tournament.Create().
			SetBdlID(t.ID).
			SetName(t.Name).
			SetStartDate(startDate).
			SetEndDate(endDate).
			SetStatus(status).
			SetSeasonYear(t.Season)

		if t.CourseName != nil && *t.CourseName != "" {
			builder.SetCourse(*t.CourseName)
		}
		if t.City != nil && *t.City != "" {
			builder.SetLocation(*t.City)
		}
		if t.Purse != nil && *t.Purse != "" {
			if purse, err := strconv.Atoi(*t.Purse); err == nil && purse > 0 {
				builder.SetPurse(purse)
			}
		}

		_, err = builder.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create tournament: %w", err)
		}
		log.Printf("Created tournament: %s", t.Name)
	case err != nil:
		return fmt.Errorf("failed to query tournament: %w", err)
	default:
		previousStatus := existing.Status

		// Update existing tournament
		updater := existing.Update().
			SetName(t.Name).
			SetStartDate(startDate).
			SetEndDate(endDate).
			SetStatus(status)

		if t.CourseName != nil && *t.CourseName != "" {
			updater.SetCourse(*t.CourseName)
		}
		if t.City != nil && *t.City != "" {
			updater.SetLocation(*t.City)
		}
		if t.Purse != nil && *t.Purse != "" {
			if purse, err := strconv.Atoi(*t.Purse); err == nil && purse > 0 {
				updater.SetPurse(purse)
			}
		}

		_, err = updater.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to update tournament: %w", err)
		}
		log.Printf("Updated tournament: %s (status: %s)", t.Name, status)

		if previousStatus != tournament.StatusCompleted && status == tournament.StatusCompleted {
			i.sendTournamentResultsEmails(ctx, existing)
		}
	}

	return nil
}

// syncPlayers fetches and stores player data from BallDontLie.
// Runs weekly to minimize API calls.
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
		if err := i.upsertPlayer(ctx, &players[idx]); err != nil {
			log.Printf("failed to upsert player %s: %v", players[idx].DisplayName, err)
			i.captureJobError("sync_players", err)
			continue
		}
	}

	i.recordSync(ctx, "players")
	log.Println("Player sync completed")
}

func (i *Ingester) upsertPlayer(ctx context.Context, p *balldontlie.Player) error {
	name := p.DisplayName
	if name == "" && p.FirstName != nil && p.LastName != nil {
		name = fmt.Sprintf("%s %s", *p.FirstName, *p.LastName)
	}

	countryCode := "UNK"
	if p.CountryCode != nil && *p.CountryCode != "" {
		countryCode = *p.CountryCode
	}

	existing, err := i.db.Golfer.Query().
		Where(golfer.BdlID(p.ID)).
		Only(ctx)

	if ent.IsNotFound(err) {
		existing, err = i.db.Golfer.Query().
			Where(golfer.Name(name)).
			Only(ctx)
	}

	switch {
	case ent.IsNotFound(err):
		builder := i.db.Golfer.Create().
			SetBdlID(p.ID).
			SetName(name).
			SetCountryCode(countryCode).
			SetActive(p.Active)

		if p.FirstName != nil {
			builder.SetFirstName(*p.FirstName)
		}
		if p.LastName != nil {
			builder.SetLastName(*p.LastName)
		}
		if p.Country != nil && *p.Country != "" {
			builder.SetCountry(*p.Country)
		}
		if p.OWGR != nil && *p.OWGR > 0 {
			builder.SetOwgr(*p.OWGR)
		}

		_, err = builder.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create golfer: %w", err)
		}
		log.Printf("Created golfer: %s", name)
	case err != nil:
		return fmt.Errorf("failed to query golfer: %w", err)
	default:
		updater := existing.Update().
			SetName(name).
			SetCountryCode(countryCode).
			SetActive(p.Active).
			SetBdlID(p.ID)

		if p.FirstName != nil {
			updater.SetFirstName(*p.FirstName)
		}
		if p.LastName != nil {
			updater.SetLastName(*p.LastName)
		}
		if p.Country != nil && *p.Country != "" {
			updater.SetCountry(*p.Country)
		}
		if p.OWGR != nil && *p.OWGR > 0 {
			updater.SetOwgr(*p.OWGR)
		}

		_, err = updater.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to update golfer: %w", err)
		}
	}

	return nil
}

// syncLeaderboards fetches live leaderboard data for active tournaments from BallDontLie.
// Runs hourly (vs the previous 2-minute interval) to optimize API usage.
func (i *Ingester) syncLeaderboards(ctx context.Context) {
	log.Println("Checking for active tournament leaderboards...")

	// Find active tournaments
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

// syncTournamentLeaderboard fetches and stores leaderboard data for a single tournament.
func (i *Ingester) syncTournamentLeaderboard(ctx context.Context, t *ent.Tournament) error {
	log.Printf("Syncing leaderboard for tournament: %s", t.Name)

	results, err := i.ballDontLie.GetTournamentResults(ctx, *t.BdlID)
	if err != nil {
		return fmt.Errorf("failed to fetch tournament results: %w", err)
	}

	log.Printf("Fetched %d results for tournament %s", len(results), t.Name)

	for idx := range results {
		if err := i.upsertTournamentEntry(ctx, t, &results[idx]); err != nil {
			log.Printf("failed to upsert result for player %s: %v", results[idx].Player.DisplayName, err)
			i.captureJobError("sync_tournament_leaderboard", err)
			continue
		}
	}

	return nil
}

func (i *Ingester) upsertTournamentEntry(ctx context.Context, t *ent.Tournament, r *balldontlie.TournamentResult) error {
	g, err := i.db.Golfer.Query().
		Where(golfer.BdlID(r.Player.ID)).
		Only(ctx)

	if ent.IsNotFound(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to query golfer: %w", err)
	}

	status := tournamententry.StatusActive
	cut := false
	if r.Tournament.Status != nil {
		switch *r.Tournament.Status {
		case "COMPLETED":
			status = tournamententry.StatusFinished
		case "IN_PROGRESS":
			status = tournamententry.StatusActive
		}
	}

	position := 0
	if r.PositionNumeric != nil {
		position = *r.PositionNumeric
	}
	if r.Position != nil && *r.Position == "CUT" {
		cut = true
		status = tournamententry.StatusFinished
	}

	score := 0
	if r.TotalScore != nil {
		score = *r.TotalScore
	}
	totalStrokes := 0

	earnings := 0
	if r.Earnings != nil {
		earnings = *r.Earnings
	}

	existing, err := i.db.TournamentEntry.Query().
		Where(
			tournamententry.HasTournamentWith(tournament.ID(t.ID)),
			tournamententry.HasGolferWith(golfer.ID(g.ID)),
		).
		Only(ctx)

	switch {
	case ent.IsNotFound(err):
		_, err = i.db.TournamentEntry.Create().
			SetTournament(t).
			SetGolfer(g).
			SetPosition(position).
			SetCut(cut).
			SetScore(score).
			SetTotalStrokes(totalStrokes).
			SetStatus(status).
			SetEarnings(earnings).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create tournament entry: %w", err)
		}
	case err != nil:
		return fmt.Errorf("failed to query tournament entry: %w", err)
	default:
		_, err = existing.Update().
			SetPosition(position).
			SetCut(cut).
			SetScore(score).
			SetTotalStrokes(totalStrokes).
			SetStatus(status).
			SetEarnings(earnings).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to update tournament entry: %w", err)
		}
	}

	return nil
}

// syncEarnings syncs earnings for completed tournaments via the BallDontLie API.
// The tournament results endpoint now includes earnings data directly.
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
