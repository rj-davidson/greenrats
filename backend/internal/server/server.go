package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v2"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/internal/auth"
	"github.com/rj-davidson/greenrats/internal/config"
	"github.com/rj-davidson/greenrats/internal/email"
	"github.com/rj-davidson/greenrats/internal/external/balldontlie"
	"github.com/rj-davidson/greenrats/internal/external/exa"
	"github.com/rj-davidson/greenrats/internal/external/googlemaps"
	"github.com/rj-davidson/greenrats/internal/external/openai"
	"github.com/rj-davidson/greenrats/internal/external/pgatour"
	"github.com/rj-davidson/greenrats/internal/features/admin"
	"github.com/rj-davidson/greenrats/internal/features/golfers"
	"github.com/rj-davidson/greenrats/internal/features/users"
	"github.com/rj-davidson/greenrats/internal/sse"
)

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

func New(cfg *config.Config, db *ent.Client) *Server {
	logLevel := slog.LevelInfo
	if cfg.IsDevelopment() {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	app := fiber.New(fiber.Config{
		AppName:      "GreenRats API",
		ErrorHandler: errorHandler,
	})

	broker := sse.NewBroker()
	sseHandler := sse.NewHandler(broker)

	authCfg := &auth.Config{
		ClientID: cfg.WorkOSClientID,
	}

	var jwksProvider *auth.JWKSProvider

	if cfg.WorkOSClientID != "" {
		var err error
		jwksProvider, err = auth.NewJWKSProvider(cfg.WorkOSClientID)
		if err != nil {
			if cfg.IsProduction() {
				logger.Error("failed to initialize JWKS provider in production", "error", err)
				os.Exit(1)
			}
			logger.Warn("failed to initialize JWKS provider, using SkipVerify mode", "error", err)
			authCfg.SkipVerify = true
		} else {
			authCfg.JWKSProvider = jwksProvider
			logger.Info("JWKS provider initialized", "client_id", cfg.WorkOSClientID)
		}
	} else {
		logger.Warn("WORKOS_CLIENT_ID not set, auth verification disabled")
		authCfg.SkipVerify = true
	}

	userService := users.New(db, cfg, logger)

	emailClient := email.New(cfg, logger)

	bdlClient := balldontlie.New(cfg.BallDontLieAPIKey, cfg.BallDontLieBaseURL, cfg.IsDevelopment(), logger)
	pgatourClient := pgatour.New(cfg.PGATourAPIKey, logger)
	exaClient := exa.New(cfg.ExaAPIKey, logger)
	openaiClient := openai.New(cfg.OpenAIAPIKey, cfg.OpenAIModel, logger)
	golferSvc := golfers.NewService(db)

	var gmapsClient *googlemaps.Client
	if cfg.GoogleMapsAPIKey != "" {
		var err error
		gmapsClient, err = googlemaps.New(cfg.GoogleMapsAPIKey, logger)
		if err != nil {
			logger.Warn("failed to create Google Maps client", "error", err)
		}
	}

	adminIngestSvc := admin.NewIngestService(
		db, cfg, bdlClient, pgatourClient, gmapsClient, exaClient, openaiClient, golferSvc, logger,
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

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	return s.app.Listen(addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.jwksProvider != nil {
		s.jwksProvider.Close()
	}

	return s.app.ShutdownWithContext(ctx)
}

func (s *Server) App() *fiber.App {
	return s.app
}

func (s *Server) SSEBroker() *sse.Broker {
	return s.sseBroker
}

func (s *Server) AuthConfig() *auth.Config {
	return s.authConfig
}

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
