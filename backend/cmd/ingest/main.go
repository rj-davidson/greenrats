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
	_ "github.com/lib/pq"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/internal/config"
	"github.com/rj-davidson/greenrats/internal/email"
	"github.com/rj-davidson/greenrats/internal/external/balldontlie"
	"github.com/rj-davidson/greenrats/internal/external/googlemaps"
	"github.com/rj-davidson/greenrats/internal/external/pgatour"
	"github.com/rj-davidson/greenrats/internal/features/fields"
	"github.com/rj-davidson/greenrats/internal/sync"
)

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

	bdlClient := balldontlie.New(cfg.BallDontLieAPIKey, cfg.BallDontLieBaseURL, cfg.IsDevelopment(), logger)
	emailClient := email.New(cfg)
	pgatourClient := pgatour.New(cfg.PGATourAPIKey, logger)
	fieldsService := fields.NewService(db, pgatourClient, logger)

	var gmapsClient *googlemaps.Client
	if cfg.GoogleMapsAPIKey != "" {
		var err error
		gmapsClient, err = googlemaps.New(cfg.GoogleMapsAPIKey, logger)
		if err != nil {
			log.Printf("Warning: failed to create Google Maps client: %v", err)
		}
	}

	syncService := sync.NewService(db, gmapsClient, logger)

	ingester := sync.NewIngester(
		db,
		cfg,
		bdlClient,
		pgatourClient,
		syncService,
		fieldsService,
		emailClient,
		logger,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go ingester.Run(ctx)

	log.Println("GreenRats Ingest Service started")
	log.Printf("Sync intervals: tournaments=%v, leaderboards(play)=%v, leaderboards(idle)=%v, scorecards=%v",
		sync.TournamentSyncInterval, sync.LeaderboardPlayInterval, sync.LeaderboardIdleInterval, sync.ScorecardSyncInterval)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down ingest service...")
	cancel()

	time.Sleep(2 * time.Second)
	log.Println("Ingest service exited gracefully")
	return nil
}
