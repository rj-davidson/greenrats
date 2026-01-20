package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v2"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/internal/auth"
	"github.com/rj-davidson/greenrats/internal/config"
	"github.com/rj-davidson/greenrats/internal/email"
	"github.com/rj-davidson/greenrats/internal/external/balldontlie"
	"github.com/rj-davidson/greenrats/internal/external/exa"
	"github.com/rj-davidson/greenrats/internal/external/openai"
	"github.com/rj-davidson/greenrats/internal/external/scrapedo"
	"github.com/rj-davidson/greenrats/internal/features/admin"
	"github.com/rj-davidson/greenrats/internal/features/golfers"
	"github.com/rj-davidson/greenrats/internal/features/users"
	"github.com/rj-davidson/greenrats/internal/sse"
)

// Server holds the HTTP server and its dependencies.
type Server struct {
	app                *fiber.App
	config             *config.Config
	db                 *ent.Client
	sseBroker          *sse.Broker
	sseHandler         *sse.Handler
	authConfig         *auth.Config
	jwksProvider       *auth.JWKSProvider
	userService        *users.Service
	emailClient        *email.Client
	adminIngestService *admin.IngestService
	logger             *slog.Logger
}

// New creates a new Server instance.
func New(cfg *config.Config, db *ent.Client) *Server {
	app := fiber.New(fiber.Config{
		AppName:      "GreenRats API",
		ErrorHandler: errorHandler,
	})

	// Initialize SSE broker
	broker := sse.NewBroker()
	sseHandler := sse.NewHandler(broker)

	// Initialize auth configuration
	authCfg := &auth.Config{
		ClientID: cfg.WorkOSClientID,
	}

	var jwksProvider *auth.JWKSProvider

	// Initialize JWKS provider if client ID is configured
	if cfg.WorkOSClientID != "" {
		var err error
		jwksProvider, err = auth.NewJWKSProvider(cfg.WorkOSClientID)
		if err != nil {
			if cfg.IsProduction() {
				log.Fatalf("Failed to initialize JWKS provider in production: %v", err)
			}
			log.Printf("Warning: failed to initialize JWKS provider: %v", err)
			log.Printf("Auth will use SkipVerify mode (development only)")
			authCfg.SkipVerify = true
		} else {
			authCfg.JWKSProvider = jwksProvider
			log.Printf("JWKS provider initialized for client ID: %s", cfg.WorkOSClientID)
		}
	} else {
		// No client ID configured - enable dev mode
		log.Printf("Warning: WORKOS_CLIENT_ID not set - auth verification disabled (development only)")
		authCfg.SkipVerify = true
	}

	// Initialize user service
	userService := users.New(db, cfg)

	// Initialize email client
	emailClient := email.New(cfg)

	// Initialize logger
	logLevel := slog.LevelInfo
	if cfg.IsDevelopment() {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	// Initialize external clients for admin ingest service
	bdlClient := balldontlie.New(cfg.BallDontLieAPIKey, cfg.BallDontLieBaseURL, logger)
	exaClient := exa.New(cfg.ExaAPIKey, logger)
	openaiClient := openai.New(cfg.OpenAIAPIKey, cfg.OpenAIModel, logger)
	scrapeDoClient := scrapedo.New(cfg.ScrapeDoAPIKey, logger)
	golferSvc := golfers.NewService(db)

	// Initialize admin ingest service
	adminIngestSvc := admin.NewIngestService(
		db, cfg, bdlClient, exaClient, openaiClient, scrapeDoClient, golferSvc, logger,
	)

	s := &Server{
		app:                app,
		config:             cfg,
		db:                 db,
		sseBroker:          broker,
		sseHandler:         sseHandler,
		authConfig:         authCfg,
		jwksProvider:       jwksProvider,
		userService:        userService,
		emailClient:        emailClient,
		adminIngestService: adminIngestSvc,
		logger:             logger,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// Start begins listening for HTTP requests.
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	return s.app.Listen(addr)
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	// Close JWKS provider to stop background refresh
	if s.jwksProvider != nil {
		s.jwksProvider.Close()
	}

	return s.app.ShutdownWithContext(ctx)
}

// App returns the underlying Fiber app for testing.
func (s *Server) App() *fiber.App {
	return s.app
}

// SSEBroker returns the SSE broker for external broadcasting.
func (s *Server) SSEBroker() *sse.Broker {
	return s.sseBroker
}

// AuthConfig returns the auth configuration for use in routes.
func (s *Server) AuthConfig() *auth.Config {
	return s.authConfig
}

// errorHandler is the custom error handler for Fiber.
func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}
