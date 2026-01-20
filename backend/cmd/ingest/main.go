package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"

	"github.com/getsentry/sentry-go"
	"github.com/gofrs/uuid/v5"
	_ "github.com/lib/pq"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/emailreminder"
	"github.com/rj-davidson/greenrats/ent/golfer"
	"github.com/rj-davidson/greenrats/ent/league"
	"github.com/rj-davidson/greenrats/ent/leaguemembership"
	"github.com/rj-davidson/greenrats/ent/pick"
	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/ent/tournamententry"
	"github.com/rj-davidson/greenrats/ent/user"
	"github.com/rj-davidson/greenrats/internal/config"
	"github.com/rj-davidson/greenrats/internal/email"
	"github.com/rj-davidson/greenrats/internal/external/balldontlie"
	"github.com/rj-davidson/greenrats/internal/external/livegolfdata"
	"github.com/rj-davidson/greenrats/internal/external/pgatour"
)

// Sync intervals based on cost optimization plan:
// - Tournaments: Daily (24 hours) via BallDontLie
// - Leaderboards: Hourly during active tournaments via BallDontLie
// - Players: Weekly via BallDontLie
// - Tournament Field: On-demand (3 days before) via Live Golf Data (250 req/mo limit)
const (
	tournamentSyncInterval  = 24 * time.Hour
	leaderboardSyncInterval = 1 * time.Hour
	playerSyncInterval      = 7 * 24 * time.Hour // Weekly
	fieldCheckInterval      = 6 * time.Hour      // Check for upcoming tournaments needing field sync
	reminderCheckInterval   = 1 * time.Hour      // Check for pick reminders
	earningsCheckInterval   = 6 * time.Hour      // Check for earnings to sync
	daysBeforeFieldSync     = 5                  // Sync field 5 days before tournament starts
	pickWindowDays          = 3                  // Pick window opens this many days before tournament
)

type Ingester struct {
	db           *ent.Client
	config       *config.Config
	liveGolfData *livegolfdata.Client
	ballDontLie  *balldontlie.Client
	pgaTour      *pgatour.Client
	email        *email.Client
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

	lgdClient := livegolfdata.New(cfg.LiveGolfDataAPIKey, cfg.LiveGolfDataBaseURL)
	bdlClient := balldontlie.New(cfg.BallDontLieAPIKey, cfg.BallDontLieBaseURL)
	pgaTourClient := pgatour.New(cfg.PGATourAPIKey, cfg.PGATourBaseURL)
	emailClient := email.New(cfg)

	ingester := &Ingester{
		db:           db,
		config:       cfg,
		liveGolfData: lgdClient,
		ballDontLie:  bdlClient,
		pgaTour:      pgaTourClient,
		email:        emailClient,
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

	// Field check - every 6 hours to find tournaments needing field sync
	fieldCheckTicker := time.NewTicker(fieldCheckInterval)
	defer fieldCheckTicker.Stop()

	// Pick reminder check - hourly
	reminderTicker := time.NewTicker(reminderCheckInterval)
	defer reminderTicker.Stop()

	// Earnings check - every 6 hours to sync earnings for recently completed tournaments
	earningsTicker := time.NewTicker(earningsCheckInterval)
	defer earningsTicker.Stop()

	// Run initial syncs on startup
	log.Println("Running initial data syncs...")
	i.syncTournaments(ctx)
	i.syncPlayers(ctx)
	i.syncUpcomingTournamentFields(ctx)
	i.syncLeaderboards(ctx)
	i.syncEarnings(ctx)
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
		case <-fieldCheckTicker.C:
			i.syncUpcomingTournamentFields(ctx)
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

// syncTournaments fetches and stores tournament data from BallDontLie.
// Runs daily to minimize API calls.
func (i *Ingester) syncTournaments(ctx context.Context) {
	log.Println("Starting tournament sync...")

	currentYear := time.Now().Year()
	tournaments, err := i.ballDontLie.GetTournaments(ctx, currentYear)
	if err != nil {
		log.Printf("failed to fetch tournaments from BallDontLie: %v", err)
		i.captureJobError("sync_tournaments", err)
		return
	}

	log.Printf("Fetched %d tournaments for %d season", len(tournaments), currentYear)

	for idx := range tournaments {
		if err := i.upsertTournament(ctx, &tournaments[idx]); err != nil {
			log.Printf("failed to upsert tournament %s: %v", tournaments[idx].Name, err)
			i.captureJobError("sync_tournaments", err)
			continue
		}
	}

	log.Println("Tournament sync completed")
}

// parseEndDate attempts to parse the tournament end date from various formats.
// Returns the end date with a buffer (06:00 UTC the next day) to account for
// late finishes and timezone differences (e.g., Hawaii is UTC-10).
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

	endDate := parseEndDate(t.EndDate, startDate)

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

		if t.Course != "" {
			builder.SetCourse(t.Course)
		}
		if t.Location != "" {
			builder.SetLocation(t.Location)
		}
		if t.Purse != "" {
			if purse, err := strconv.Atoi(t.Purse); err == nil && purse > 0 {
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

		if t.Course != "" {
			updater.SetCourse(t.Course)
		}
		if t.Location != "" {
			updater.SetLocation(t.Location)
		}
		if t.Purse != "" {
			if purse, err := strconv.Atoi(t.Purse); err == nil && purse > 0 {
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
			log.Printf("failed to upsert player %s %s: %v", players[idx].FirstName, players[idx].LastName, err)
			i.captureJobError("sync_players", err)
			continue
		}
	}

	log.Println("Player sync completed")
}

// upsertPlayer creates or updates a golfer from BallDontLie data.
func (i *Ingester) upsertPlayer(ctx context.Context, p *balldontlie.Player) error {
	// Use display_name from API if available, otherwise construct from first/last
	name := p.DisplayName
	if name == "" {
		name = fmt.Sprintf("%s %s", p.FirstName, p.LastName)
	}

	// Use country_code with fallback to "UNK"
	countryCode := p.CountryCode
	if countryCode == "" {
		countryCode = "UNK"
	}

	// Try to find existing golfer by BallDontLie ID
	existing, err := i.db.Golfer.Query().
		Where(golfer.BdlID(p.ID)).
		Only(ctx)

	if ent.IsNotFound(err) {
		// Try to find by name match (for linking with Live Golf Data)
		existing, err = i.db.Golfer.Query().
			Where(golfer.Name(name)).
			Only(ctx)
	}

	switch {
	case ent.IsNotFound(err):
		// Create new golfer
		builder := i.db.Golfer.Create().
			SetBdlID(p.ID).
			SetName(name).
			SetFirstName(p.FirstName).
			SetLastName(p.LastName).
			SetCountryCode(countryCode).
			SetActive(p.Active)

		if p.Country != "" {
			builder.SetCountry(p.Country)
		}
		if p.OWGR > 0 {
			builder.SetOwgr(p.OWGR)
		}

		_, err = builder.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create golfer: %w", err)
		}
		log.Printf("Created golfer: %s", name)
	case err != nil:
		return fmt.Errorf("failed to query golfer: %w", err)
	default:
		// Update existing golfer
		updater := existing.Update().
			SetName(name).
			SetFirstName(p.FirstName).
			SetLastName(p.LastName).
			SetCountryCode(countryCode).
			SetActive(p.Active).
			SetBdlID(p.ID)

		if p.Country != "" {
			updater.SetCountry(p.Country)
		}
		if p.OWGR > 0 {
			updater.SetOwgr(p.OWGR)
		}

		_, err = updater.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to update golfer: %w", err)
		}
	}

	return nil
}

// syncUpcomingTournamentFields checks for tournaments starting within 3 days
// and syncs their field from Live Golf Data if not already synced.
// Uses Live Golf Data's tournament-field endpoint (counts toward 250/month limit).
func (i *Ingester) syncUpcomingTournamentFields(ctx context.Context) {
	log.Println("Checking for tournaments needing field sync...")

	// Find tournaments starting within the next 3 days that need field sync
	now := time.Now()
	threshold := now.AddDate(0, 0, daysBeforeFieldSync)

	tournaments, err := i.db.Tournament.Query().
		Where(
			tournament.StatusEQ(tournament.StatusUpcoming),
			tournament.StartDateLTE(threshold),
			tournament.StartDateGTE(now),
		).
		All(ctx)
	if err != nil {
		log.Printf("failed to query upcoming tournaments: %v", err)
		i.captureJobError("sync_upcoming_fields", err)
		return
	}

	for _, t := range tournaments {
		count, err := t.QueryEntries().Count(ctx)
		if err != nil {
			log.Printf("failed to count entries for tournament %s: %v", t.Name, err)
			i.captureJobError("sync_upcoming_fields", err)
			continue
		}

		if count > 0 {
			log.Printf("Tournament %s already has %d entries in field", t.Name, count)
			continue
		}

		if err := i.syncTournamentField(ctx, t); err != nil {
			log.Printf("failed to sync field for tournament %s: %v", t.Name, err)
			i.captureJobError("sync_upcoming_fields", err)
		}
	}

	log.Println("Field sync check completed")
}

func (i *Ingester) syncTournamentField(ctx context.Context, t *ent.Tournament) error {
	if t.ScratchgolfID != nil {
		log.Printf("Syncing field for tournament: %s (Live Golf Data ID: %s)", t.Name, *t.ScratchgolfID)

		golfers, err := i.liveGolfData.GetTournamentField(ctx, *t.ScratchgolfID)
		if err == nil && len(golfers) > 0 {
			log.Printf("Fetched %d golfers for tournament %s", len(golfers), t.Name)

			var entryCount int
			for idx := range golfers {
				golferEnt, err := i.upsertGolferFromLiveGolfData(ctx, &golfers[idx])
				if err != nil {
					log.Printf("failed to upsert golfer %s: %v", golfers[idx].Name, err)
					i.captureJobError("sync_tournament_field", err)
					continue
				}

				_, err = i.db.TournamentEntry.Create().
					SetTournament(t).
					SetGolfer(golferEnt).
					SetStatus(tournamententry.StatusPending).
					Save(ctx)
				if err != nil {
					log.Printf("failed to create entry for golfer %s: %v", golfers[idx].Name, err)
					i.captureJobError("sync_tournament_field", err)
					continue
				}
				entryCount++
			}

			log.Printf("Created %d entries for tournament %s", entryCount, t.Name)
			return nil
		}

		if err != nil {
			log.Printf("Live Golf Data field fetch failed for %s: %v", t.Name, err)
			i.captureJobError("sync_tournament_field", err)
		} else {
			log.Printf("Live Golf Data field empty for %s, falling back to PGA Tour", t.Name)
		}
	}

	if i.pgaTour == nil {
		return fmt.Errorf("PGA Tour client not configured")
	}

	fieldID, err := i.findPgaTourTournamentID(ctx, t)
	if err != nil {
		return err
	}

	field, err := i.pgaTour.GetField(ctx, fieldID, false, false)
	if err != nil {
		return fmt.Errorf("failed to fetch PGA Tour field: %w", err)
	}

	if len(field.Players) == 0 {
		log.Printf("PGA Tour field empty for tournament %s", t.Name)
		return nil
	}

	log.Printf("Fetched %d PGA Tour golfers for tournament %s", len(field.Players), t.Name)

	var entryCount int
	for idx := range field.Players {
		player := field.Players[idx]
		golferEnt, err := i.upsertGolferFromPgaTour(ctx, &player)
		if err != nil {
			log.Printf("failed to upsert golfer %s: %v", player.DisplayName, err)
			i.captureJobError("sync_tournament_field", err)
			continue
		}

		_, err = i.db.TournamentEntry.Create().
			SetTournament(t).
			SetGolfer(golferEnt).
			SetStatus(tournamententry.StatusPending).
			Save(ctx)
		if err != nil {
			log.Printf("failed to create entry for golfer %s: %v", player.DisplayName, err)
			i.captureJobError("sync_tournament_field", err)
			continue
		}
		entryCount++
	}

	log.Printf("Created %d entries for tournament %s", entryCount, t.Name)
	return nil
}

func (i *Ingester) findPgaTourTournamentID(ctx context.Context, t *ent.Tournament) (string, error) {
	if i.pgaTour == nil {
		return "", fmt.Errorf("PGA Tour client not configured")
	}

	year := t.SeasonYear
	if year == 0 {
		year = t.StartDate.Year()
	}

	tournaments, err := i.pgaTour.GetUpcomingSchedule(ctx, pgatour.DefaultTourCode, year)
	if err != nil {
		return "", fmt.Errorf("failed to fetch PGA Tour schedule: %w", err)
	}

	target := normalizeTournamentName(t.Name)
	var fallback string
	for idx := range tournaments {
		name := normalizeTournamentName(tournaments[idx].TournamentName)
		if name == target {
			return tournaments[idx].ID, nil
		}
		if fallback == "" && (strings.Contains(name, target) || strings.Contains(target, name)) {
			fallback = tournaments[idx].ID
		}
	}

	if fallback != "" {
		return fallback, nil
	}

	return "", fmt.Errorf("no PGA Tour schedule match for tournament %s", t.Name)
}

func normalizeTournamentName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var builder strings.Builder
	builder.Grow(len(name))
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
			builder.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(builder.String()), " ")
}

func (i *Ingester) upsertGolferFromPgaTour(ctx context.Context, p *pgatour.FieldPlayer) (*ent.Golfer, error) {
	name := p.DisplayName
	if name == "" {
		name = p.ShortName
	}
	if name == "" && p.FirstName != "" && p.LastName != "" {
		name = fmt.Sprintf("%s %s", p.FirstName, p.LastName)
	}
	if name == "" {
		return nil, fmt.Errorf("missing golfer name")
	}

	existing, err := i.findMatchingGolfer(ctx, name, p.FirstName, p.LastName)

	if errors.Is(err, errNoMatch) {
		builder := i.db.Golfer.Create().
			SetName(name).
			SetActive(true)

		if p.FirstName != "" {
			builder.SetFirstName(p.FirstName)
		}
		if p.LastName != "" {
			builder.SetLastName(p.LastName)
		}
		if p.Country != "" {
			builder.SetCountry(p.Country)
		}
		if p.OWGR.Valid && p.OWGR.Int > 0 {
			builder.SetOwgr(p.OWGR.Int)
		}

		return builder.Save(ctx)
	} else if err != nil {
		return nil, fmt.Errorf("failed to query golfer: %w", err)
	}

	updater := existing.Update()
	if p.FirstName != "" {
		updater.SetFirstName(p.FirstName)
	}
	if p.LastName != "" {
		updater.SetLastName(p.LastName)
	}
	if p.Country != "" {
		updater.SetCountry(p.Country)
	}
	if p.OWGR.Valid && p.OWGR.Int > 0 {
		updater.SetOwgr(p.OWGR.Int)
	}

	return updater.Save(ctx)
}

var errNoMatch = errors.New("no golfer match")

func (i *Ingester) findMatchingGolfer(ctx context.Context, displayName, firstName, lastName string) (*ent.Golfer, error) {
	if displayName != "" {
		existing, err := i.db.Golfer.Query().
			Where(golfer.NameEqualFold(displayName)).
			Only(ctx)
		if err == nil || !ent.IsNotFound(err) {
			return existing, err
		}
	}

	if firstName != "" && lastName != "" {
		existing, err := i.db.Golfer.Query().
			Where(
				golfer.FirstNameEqualFold(firstName),
				golfer.LastNameEqualFold(lastName),
			).
			Only(ctx)
		if err == nil || !ent.IsNotFound(err) {
			return existing, err
		}
	}

	if lastName == "" {
		return nil, errNoMatch
	}

	candidates, err := i.db.Golfer.Query().
		Where(golfer.LastNameEqualFold(lastName)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	bestScore := 0
	var best *ent.Golfer
	for _, candidate := range candidates {
		candidateName := candidate.Name
		if candidateName == "" && candidate.FirstName != nil && candidate.LastName != nil {
			candidateName = fmt.Sprintf("%s %s", *candidate.FirstName, *candidate.LastName)
		}
		score := scoreNameMatch(displayName, candidateName, firstName, lastName)
		if score > bestScore {
			bestScore = score
			best = candidate
		}
	}

	if best != nil && bestScore >= 60 {
		return best, nil
	}

	return nil, errNoMatch
}

func scoreNameMatch(displayName, candidateName, firstName, lastName string) int {
	normalizedDisplay := normalizePersonName(displayName)
	normalizedCandidate := normalizePersonName(candidateName)

	if normalizedDisplay == "" || normalizedCandidate == "" {
		return 0
	}

	if normalizedDisplay == normalizedCandidate {
		return 100
	}

	if strings.Contains(normalizedDisplay, normalizedCandidate) || strings.Contains(normalizedCandidate, normalizedDisplay) {
		return 80
	}

	displayTokens := strings.Fields(normalizedDisplay)
	candidateTokens := strings.Fields(normalizedCandidate)
	if len(displayTokens) == 0 || len(candidateTokens) == 0 {
		return 0
	}

	displayLast := displayTokens[len(displayTokens)-1]
	candidateLast := candidateTokens[len(candidateTokens)-1]
	if displayLast != candidateLast {
		return 0
	}

	displayFirst := displayTokens[0]
	candidateFirst := candidateTokens[0]

	if displayFirst == candidateFirst {
		return 90
	}

	if strings.HasPrefix(displayFirst, candidateFirst) || strings.HasPrefix(candidateFirst, displayFirst) {
		return 70
	}

	if displayFirst[:1] == candidateFirst[:1] {
		return 60
	}

	if firstName != "" && lastName != "" {
		normalizedFirst := normalizePersonName(firstName)
		normalizedLast := normalizePersonName(lastName)
		if normalizedLast == candidateLast && normalizedFirst != "" && candidateFirst != "" &&
			normalizedFirst[:1] == candidateFirst[:1] {
			return 60
		}
	}

	return 0
}

func normalizePersonName(name string) string {
	if name == "" {
		return ""
	}

	if strings.Contains(name, ",") {
		parts := strings.SplitN(name, ",", 2)
		left := strings.TrimSpace(parts[0])
		right := strings.TrimSpace(parts[1])
		if left != "" && right != "" {
			name = right + " " + left
		}
	}

	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(name))
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
			builder.WriteRune(r)
			continue
		}
		if r == '-' || r == '.' || r == '\'' {
			builder.WriteRune(' ')
		}
	}

	normalized := removeDiacritics(builder.String())
	parts := strings.Fields(normalized)
	if len(parts) == 0 {
		return ""
	}

	suffixes := map[string]struct{}{
		"jr": {}, "sr": {}, "ii": {}, "iii": {}, "iv": {}, "v": {},
	}
	for len(parts) > 0 {
		if _, ok := suffixes[parts[len(parts)-1]]; ok {
			parts = parts[:len(parts)-1]
			continue
		}
		break
	}

	return strings.Join(parts, " ")
}

func removeDiacritics(input string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '\u00e0', '\u00e1', '\u00e2', '\u00e3', '\u00e4', '\u00e5',
			'\u0101', '\u0103', '\u0105':
			return 'a'
		case '\u00e7', '\u0107', '\u0109', '\u010b', '\u010d':
			return 'c'
		case '\u00e8', '\u00e9', '\u00ea', '\u00eb', '\u0113', '\u0115',
			'\u0117', '\u0119', '\u011b':
			return 'e'
		case '\u00ec', '\u00ed', '\u00ee', '\u00ef', '\u012b', '\u012d',
			'\u012f', '\u0131':
			return 'i'
		case '\u00f1', '\u0144', '\u0146', '\u0148':
			return 'n'
		case '\u00f2', '\u00f3', '\u00f4', '\u00f5', '\u00f6', '\u00f8',
			'\u014d', '\u014f', '\u0151':
			return 'o'
		case '\u00f9', '\u00fa', '\u00fb', '\u00fc', '\u016b', '\u016d',
			'\u016f', '\u0171', '\u0173':
			return 'u'
		case '\u00fd', '\u00ff', '\u0177':
			return 'y'
		case '\u00df':
			return 's'
		case '\u0142':
			return 'l'
		default:
			return r
		}
	}, input)
}

// upsertGolferFromLiveGolfData creates or updates a golfer from Live Golf Data.
func (i *Ingester) upsertGolferFromLiveGolfData(ctx context.Context, g *livegolfdata.Golfer) (*ent.Golfer, error) {
	// Try to find existing golfer by Live Golf Data ID
	existing, err := i.db.Golfer.Query().
		Where(golfer.ScratchgolfID(g.ID)).
		Only(ctx)

	if ent.IsNotFound(err) {
		// Try to find by name match (for linking with BallDontLie data)
		name := g.Name
		if name == "" && g.FirstName != "" && g.LastName != "" {
			name = fmt.Sprintf("%s %s", g.FirstName, g.LastName)
		}

		existing, err = i.db.Golfer.Query().
			Where(golfer.Name(name)).
			Only(ctx)

		if ent.IsNotFound(err) {
			// Create new golfer
			builder := i.db.Golfer.Create().
				SetScratchgolfID(g.ID).
				SetName(name)

			if g.Country != "" {
				builder.SetCountry(g.Country)
			}
			if g.FirstName != "" {
				builder.SetFirstName(g.FirstName)
			}
			if g.LastName != "" {
				builder.SetLastName(g.LastName)
			}
			if g.WorldRanking > 0 {
				builder.SetOwgr(g.WorldRanking)
			}
			if g.ImageURL != "" {
				builder.SetImageURL(g.ImageURL)
			}

			return builder.Save(ctx)
		} else if err != nil {
			return nil, fmt.Errorf("failed to query golfer: %w", err)
		}

		// Found by name, update with Live Golf Data ID
		return existing.Update().
			SetScratchgolfID(g.ID).
			Save(ctx)
	} else if err != nil {
		return nil, fmt.Errorf("failed to query golfer: %w", err)
	}

	// Update existing golfer
	updater := existing.Update()
	if g.WorldRanking > 0 {
		updater.SetOwgr(g.WorldRanking)
	}
	if g.ImageURL != "" {
		updater.SetImageURL(g.ImageURL)
	}

	return updater.Save(ctx)
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
	switch r.Tournament.Status {
	case "COMPLETED":
		status = tournamententry.StatusFinished
	case "IN_PROGRESS":
		status = tournamententry.StatusActive
	}

	position := r.PositionNumeric
	if r.Position == "CUT" {
		cut = true
		status = tournamententry.StatusFinished
	}

	score := r.TotalScore // API's total_score is actually score relative to par
	totalStrokes := 0     // API doesn't provide actual total strokes

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
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to update tournament entry: %w", err)
		}
	}

	return nil
}

// syncEarnings fetches earnings data for completed tournaments.
// Syncs any completed tournament missing earnings, plus recently completed
// tournaments on days 1, 2, and 7 to catch late updates.
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
		hasEarnings, err := i.tournamentHasEarnings(ctx, t)
		if err != nil {
			log.Printf("failed to check earnings for tournament %s: %v", t.Name, err)
			i.captureJobError("sync_earnings", err)
			continue
		}

		if !hasEarnings {
			log.Printf("Tournament %s missing earnings, syncing...", t.Name)
			if err := i.syncTournamentEarnings(ctx, t); err != nil {
				log.Printf("failed to sync earnings for tournament %s: %v", t.Name, err)
				i.captureJobError("sync_earnings", err)
			}
			continue
		}

		daysSinceEnd := int(now.Sub(t.EndDate).Hours() / 24)
		shouldRefresh := daysSinceEnd == 1 || daysSinceEnd == 2 || daysSinceEnd == 7
		if shouldRefresh {
			log.Printf("Refreshing earnings for recently completed tournament %s (day %d)", t.Name, daysSinceEnd)
			if err := i.syncTournamentEarnings(ctx, t); err != nil {
				log.Printf("failed to sync earnings for tournament %s: %v", t.Name, err)
				i.captureJobError("sync_earnings", err)
			}
		}
	}

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

func (i *Ingester) syncTournamentEarnings(ctx context.Context, t *ent.Tournament) error {
	log.Printf("Syncing earnings for tournament: %s", t.Name)

	lgdTournamentID, err := i.findLiveGolfDataTournamentID(ctx, t)
	if err != nil {
		return fmt.Errorf("failed to find Live Golf Data tournament ID: %w", err)
	}

	year := t.SeasonYear
	if year == 0 {
		year = t.EndDate.Year()
	}

	earnings, err := i.liveGolfData.GetEarnings(ctx, lgdTournamentID, year)
	if err != nil {
		return fmt.Errorf("failed to fetch earnings: %w", err)
	}

	if earnings == nil {
		log.Printf("Earnings not yet available for tournament %s", t.Name)
		return nil
	}

	log.Printf("Fetched %d earnings entries for tournament %s", len(earnings), t.Name)

	var updated int
	for _, e := range earnings {
		golferName := fmt.Sprintf("%s %s", e.FirstName, e.LastName)
		g, err := i.findMatchingGolfer(ctx, golferName, e.FirstName, e.LastName)
		if errors.Is(err, errNoMatch) {
			continue
		}
		if err != nil {
			log.Printf("failed to find golfer %s: %v", golferName, err)
			i.captureJobError("sync_earnings", err)
			continue
		}

		entry, err := i.db.TournamentEntry.Query().
			Where(
				tournamententry.HasTournamentWith(tournament.IDEQ(t.ID)),
				tournamententry.HasGolferWith(golfer.IDEQ(g.ID)),
			).
			Only(ctx)
		if ent.IsNotFound(err) {
			continue
		}
		if err != nil {
			log.Printf("failed to query tournament entry for %s: %v", golferName, err)
			i.captureJobError("sync_earnings", err)
			continue
		}

		if entry.Earnings != e.Earnings {
			_, err = entry.Update().SetEarnings(e.Earnings).Save(ctx)
			if err != nil {
				log.Printf("failed to update earnings for %s: %v", golferName, err)
				i.captureJobError("sync_earnings", err)
				continue
			}
			updated++
		}
	}

	log.Printf("Updated earnings for %d golfers in tournament %s", updated, t.Name)
	return nil
}

func (i *Ingester) findLiveGolfDataTournamentID(ctx context.Context, t *ent.Tournament) (string, error) {
	if t.ScratchgolfID != nil && *t.ScratchgolfID != "" {
		return *t.ScratchgolfID, nil
	}

	year := t.SeasonYear
	if year == 0 {
		year = t.EndDate.Year()
	}

	schedule, err := i.liveGolfData.GetSchedule(ctx, year)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Live Golf Data schedule: %w", err)
	}

	target := normalizeTournamentName(t.Name)
	var fallback string
	for idx := range schedule {
		s := &schedule[idx]
		name := normalizeTournamentName(s.Name)
		if name == target {
			if err := i.saveLiveGolfDataTournamentID(ctx, t, s.ID); err != nil {
				log.Printf("failed to save Live Golf Data tournament ID: %v", err)
				i.captureJobError("sync_earnings", err)
			}
			return s.ID, nil
		}
		if fallback == "" && (strings.Contains(name, target) || strings.Contains(target, name)) {
			fallback = s.ID
		}
	}

	if fallback != "" {
		if err := i.saveLiveGolfDataTournamentID(ctx, t, fallback); err != nil {
			log.Printf("failed to save Live Golf Data tournament ID: %v", err)
			i.captureJobError("sync_earnings", err)
		}
		return fallback, nil
	}

	return "", fmt.Errorf("no Live Golf Data schedule match for tournament %s", t.Name)
}

func (i *Ingester) saveLiveGolfDataTournamentID(ctx context.Context, t *ent.Tournament, lgdID string) error {
	_, err := t.Update().SetScratchgolfID(lgdID).Save(ctx)
	return err
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

var _ = uuid.Nil
