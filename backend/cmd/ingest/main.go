package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/golfer"
	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/ent/tournamententry"
	"github.com/rj-davidson/greenrats/internal/config"
	"github.com/rj-davidson/greenrats/internal/external/balldontlie"
	"github.com/rj-davidson/greenrats/internal/external/scratchgolf"
)

// Sync intervals based on cost optimization plan:
// - Tournaments: Daily (24 hours) via BallDontLie
// - Leaderboards: Hourly during active tournaments via BallDontLie
// - Players: Weekly via BallDontLie
// - Tournament Field: On-demand (3 days before) via SlashGolf (250 req/mo limit)
const (
	tournamentSyncInterval  = 24 * time.Hour
	leaderboardSyncInterval = 1 * time.Hour
	playerSyncInterval      = 7 * 24 * time.Hour // Weekly
	fieldCheckInterval      = 6 * time.Hour      // Check for upcoming tournaments needing field sync
	daysBeforeFieldSync     = 5                  // Sync field 5 days before tournament starts
)

type Ingester struct {
	db          *ent.Client
	config      *config.Config
	scratchGolf *scratchgolf.Client
	ballDontLie *balldontlie.Client
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Connect to database
	db, err := ent.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize external API clients
	sgClient := scratchgolf.New(cfg.ScratchGolfAPIKey, cfg.ScratchGolfBaseURL)
	bdlClient := balldontlie.New(cfg.BallDontLieAPIKey, cfg.BallDontLieBaseURL)

	ingester := &Ingester{
		db:          db,
		config:      cfg,
		scratchGolf: sgClient,
		ballDontLie: bdlClient,
	}

	// Create context that cancels on shutdown signal
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start scheduled jobs
	go ingester.runScheduledJobs(ctx)

	log.Println("GreenRats Ingest Service started")
	log.Printf("Sync intervals: tournaments=%v, leaderboards=%v, players=%v",
		tournamentSyncInterval, leaderboardSyncInterval, playerSyncInterval)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down ingest service...")
	cancel()

	// Give jobs time to complete
	time.Sleep(2 * time.Second)
	log.Println("Ingest service exited gracefully")
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

	// Run initial syncs on startup
	log.Println("Running initial data syncs...")
	i.syncTournaments(ctx)
	i.syncPlayers(ctx)
	i.syncUpcomingTournamentFields(ctx)
	i.syncLeaderboards(ctx)

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
		}
	}
}

// syncTournaments fetches and stores tournament data from BallDontLie.
// Runs daily to minimize API calls.
func (i *Ingester) syncTournaments(ctx context.Context) {
	log.Println("Starting tournament sync...")

	currentYear := time.Now().Year()
	tournaments, err := i.ballDontLie.GetTournaments(ctx, currentYear)
	if err != nil {
		log.Printf("failed to fetch tournaments from BallDontLie: %v", err)
		return
	}

	log.Printf("Fetched %d tournaments for %d season", len(tournaments), currentYear)

	for idx := range tournaments {
		if err := i.upsertTournament(ctx, &tournaments[idx]); err != nil {
			log.Printf("failed to upsert tournament %s: %v", tournaments[idx].Name, err)
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

	if ent.IsNotFound(err) {
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
	} else if err != nil {
		return fmt.Errorf("failed to query tournament: %w", err)
	} else {
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
		return
	}

	log.Printf("Fetched %d players", len(players))

	for idx := range players {
		if err := i.upsertPlayer(ctx, &players[idx]); err != nil {
			log.Printf("failed to upsert player %s %s: %v", players[idx].FirstName, players[idx].LastName, err)
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
		// Try to find by name match (for linking with ScratchGolf data)
		existing, err = i.db.Golfer.Query().
			Where(golfer.Name(name)).
			Only(ctx)
	}

	if ent.IsNotFound(err) {
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
	} else if err != nil {
		return fmt.Errorf("failed to query golfer: %w", err)
	} else {
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
// and syncs their field from SlashGolf if not already synced.
// Uses SlashGolf's tournament-field endpoint (counts toward 250/month limit).
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
		return
	}

	for _, t := range tournaments {
		count, err := t.QueryEntries().Count(ctx)
		if err != nil {
			log.Printf("failed to count entries for tournament %s: %v", t.Name, err)
			continue
		}

		if count > 0 {
			log.Printf("Tournament %s already has %d entries in field", t.Name, count)
			continue
		}

		if err := i.syncTournamentField(ctx, t); err != nil {
			log.Printf("failed to sync field for tournament %s: %v", t.Name, err)
		}
	}

	log.Println("Field sync check completed")
}

func (i *Ingester) syncTournamentField(ctx context.Context, t *ent.Tournament) error {
	if t.ScratchgolfID == nil {
		log.Printf("Tournament %s has no ScratchGolf ID, skipping field sync", t.Name)
		return nil
	}

	log.Printf("Syncing field for tournament: %s (ID: %s)", t.Name, *t.ScratchgolfID)

	golfers, err := i.scratchGolf.GetTournamentField(ctx, *t.ScratchgolfID)
	if err != nil {
		return fmt.Errorf("failed to fetch tournament field: %w", err)
	}

	log.Printf("Fetched %d golfers for tournament %s", len(golfers), t.Name)

	var entryCount int
	for idx := range golfers {
		golferEnt, err := i.upsertGolferFromSlashGolf(ctx, &golfers[idx])
		if err != nil {
			log.Printf("failed to upsert golfer %s: %v", golfers[idx].Name, err)
			continue
		}

		_, err = i.db.TournamentEntry.Create().
			SetTournament(t).
			SetGolfer(golferEnt).
			SetStatus(tournamententry.StatusPending).
			Save(ctx)
		if err != nil {
			log.Printf("failed to create entry for golfer %s: %v", golfers[idx].Name, err)
			continue
		}
		entryCount++
	}

	log.Printf("Created %d entries for tournament %s", entryCount, t.Name)
	return nil
}

// upsertGolferFromSlashGolf creates or updates a golfer from SlashGolf data.
func (i *Ingester) upsertGolferFromSlashGolf(ctx context.Context, g *scratchgolf.Golfer) (*ent.Golfer, error) {
	// Try to find existing golfer by SlashGolf ID
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

		// Found by name, update with SlashGolf ID
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
			log.Printf("failed to upsert result for player %d: %v", results[idx].PlayerID, err)
			continue
		}
	}

	return nil
}

func (i *Ingester) upsertTournamentEntry(ctx context.Context, t *ent.Tournament, r *balldontlie.TournamentResult) error {
	g, err := i.db.Golfer.Query().
		Where(golfer.BdlID(r.PlayerID)).
		Only(ctx)

	if ent.IsNotFound(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to query golfer: %w", err)
	}

	status := tournamententry.StatusActive
	cut := false
	switch r.Status {
	case "cut":
		status = tournamententry.StatusFinished
		cut = true
	case "withdrawn":
		status = tournamententry.StatusWithdrawn
	case "finished":
		status = tournamententry.StatusFinished
	}

	position := parsePosition(r.Position)
	if r.Position == "CUT" {
		cut = true
	}
	score, _ := strconv.Atoi(r.Score)
	totalStrokes, _ := strconv.Atoi(r.TotalStrokes)
	earnings, _ := strconv.Atoi(r.Earnings)

	existing, err := i.db.TournamentEntry.Query().
		Where(
			tournamententry.HasTournamentWith(tournament.ID(t.ID)),
			tournamententry.HasGolferWith(golfer.ID(g.ID)),
		).
		Only(ctx)

	if ent.IsNotFound(err) {
		_, err = i.db.TournamentEntry.Create().
			SetTournament(t).
			SetGolfer(g).
			SetPosition(position).
			SetCut(cut).
			SetScore(score).
			SetTotalStrokes(totalStrokes).
			SetEarnings(earnings).
			SetStatus(status).
			Save(ctx)

		if err != nil {
			return fmt.Errorf("failed to create tournament entry: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to query tournament entry: %w", err)
	} else {
		_, err = existing.Update().
			SetPosition(position).
			SetCut(cut).
			SetScore(score).
			SetTotalStrokes(totalStrokes).
			SetEarnings(earnings).
			SetStatus(status).
			Save(ctx)

		if err != nil {
			return fmt.Errorf("failed to update tournament entry: %w", err)
		}
	}

	return nil
}

func parsePosition(pos string) int {
	if pos == "" || pos == "CUT" {
		return 0
	}
	pos = strings.TrimPrefix(pos, "T")
	position, _ := strconv.Atoi(pos)
	return position
}
