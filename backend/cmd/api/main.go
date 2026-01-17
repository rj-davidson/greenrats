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

	// Run auto-migration in development (use Atlas in production)
	if cfg.IsDevelopment() {
		if err := db.Schema.Create(context.Background()); err != nil {
			log.Fatalf("failed to run migrations: %v", err)
		}
	}

	// Create and start server
	srv := server.New(cfg, db)

	// Graceful shutdown
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	log.Printf("GreenRats API started on port %d", cfg.Port)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Give outstanding requests 10 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
