package server

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/internal/config"
	"github.com/rj-davidson/greenrats/internal/sse"
)

// Server holds the HTTP server and its dependencies.
type Server struct {
	app       *fiber.App
	config    *config.Config
	db        *ent.Client
	sseBroker *sse.Broker
	sseHandler *sse.Handler
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

	s := &Server{
		app:        app,
		config:     cfg,
		db:         db,
		sseBroker:  broker,
		sseHandler: sseHandler,
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

// errorHandler is the custom error handler for Fiber.
func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}
