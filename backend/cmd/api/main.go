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
	"github.com/rj-davidson/greenrats/internal/server"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Connect to database
	db, err := ent.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	// Run auto-migration in development (use Atlas in production)
	if cfg.IsDevelopment() {
		if err := db.Schema.Create(context.Background()); err != nil {
			return err
		}
	}

	// Create and start server
	srv := server.New(cfg, db)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil {
			errChan <- err
		}
	}()

	log.Printf("GreenRats API started on port %d", cfg.Port)

	// Wait for interrupt signal or server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		log.Println("Shutting down server...")
	case err := <-errChan:
		return err
	}

	// Give outstanding requests 10 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return err
	}

	log.Println("Server exited gracefully")
	return nil
}
