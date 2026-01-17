package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/internal/config"
	"github.com/rj-davidson/greenrats/internal/external/balldontlie"
	"github.com/rj-davidson/greenrats/internal/external/scratchgolf"
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
	// Tournament sync - every 6 hours
	tournamentTicker := time.NewTicker(6 * time.Hour)
	defer tournamentTicker.Stop()

	// Live leaderboard sync - every 2 minutes during active tournaments
	leaderboardTicker := time.NewTicker(2 * time.Minute)
	defer leaderboardTicker.Stop()

	// Golfer stats sync - every 24 hours
	statsTicker := time.NewTicker(24 * time.Hour)
	defer statsTicker.Stop()

	// Run initial sync on startup
	i.syncTournaments(ctx)
	i.syncGolferStats(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-tournamentTicker.C:
			i.syncTournaments(ctx)
		case <-leaderboardTicker.C:
			i.syncLeaderboards(ctx)
		case <-statsTicker.C:
			i.syncGolferStats(ctx)
		}
	}
}

// syncTournaments fetches and stores tournament data.
func (i *Ingester) syncTournaments(ctx context.Context) {
	log.Println("Starting tournament sync...")

	currentYear := time.Now().Year()
	tournaments, err := i.scratchGolf.GetTournaments(ctx, currentYear)
	if err != nil {
		log.Printf("failed to fetch tournaments: %v", err)
		return
	}

	log.Printf("Fetched %d tournaments for %d season", len(tournaments), currentYear)

	// TODO: Upsert tournaments to database
	for _, t := range tournaments {
		log.Printf("Tournament: %s (%s)", t.Name, t.ID)
	}

	log.Println("Tournament sync completed")
}

// syncLeaderboards fetches live leaderboard data for active tournaments.
func (i *Ingester) syncLeaderboards(ctx context.Context) {
	// TODO: Query for active tournaments
	// For each active tournament, fetch and broadcast leaderboard updates
	log.Println("Checking for active tournament leaderboards...")
}

// syncGolferStats fetches and stores golfer statistics.
func (i *Ingester) syncGolferStats(ctx context.Context) {
	log.Println("Starting golfer stats sync...")

	// TODO: Query all golfers and update their stats
	log.Println("Golfer stats sync completed")
}
