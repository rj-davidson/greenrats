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
			logger.Warn("sentry initialization failed", "error", err)
		} else {
			defer sentry.Flush(2 * time.Second)
			logger.Info("sentry initialized")
		}
	}

	db, err := ent.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	bdlClient := balldontlie.New(cfg.BallDontLieAPIKey, cfg.BallDontLieBaseURL, cfg.IsDevelopment(), logger)
	emailClient := email.New(cfg, logger)

	var gmapsClient *googlemaps.Client
	if cfg.GoogleMapsAPIKey != "" {
		var err error
		gmapsClient, err = googlemaps.New(cfg.GoogleMapsAPIKey, logger)
		if err != nil {
			logger.Warn("failed to create Google Maps client", "error", err)
		}
	}

	syncService := sync.NewService(db, gmapsClient, logger)

	ingester := sync.NewIngester(
		db,
		cfg,
		bdlClient,
		syncService,
		emailClient,
		logger,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go ingester.Run(ctx)

	logger.Info("GreenRats Ingest Service started",
		"tournament_interval", sync.TournamentSyncInterval,
		"placement_interval", sync.PlacementSyncInterval,
		"scorecard_interval", sync.ScorecardSyncInterval,
	)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down ingest service")
	cancel()

	time.Sleep(2 * time.Second)
	logger.Info("ingest service exited gracefully")
	return nil
}
