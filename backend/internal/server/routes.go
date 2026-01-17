package server

import (
	"github.com/gofiber/fiber/v2"

	"github.com/rj-davidson/greenrats/internal/auth"
	"github.com/rj-davidson/greenrats/internal/features/tournaments"
)

// setupRoutes configures all routes for the server.
func (s *Server) setupRoutes() {
	// Health check endpoint (no prefix) - public
	s.app.Get("/health", s.healthCheck)

	// API v1 routes
	v1 := s.app.Group("/api/v1")

	// Public routes
	v1.Get("/", s.apiInfo)

	// SSE routes for live updates - public
	v1.Get("/sse/:topic", s.sseHandler.HandleSSE)
	v1.Get("/tournaments/:id/live", s.sseHandler.HandleTournamentSSE)

	// Configure user provisioning middleware
	ensureUserCfg := auth.EnsureUserConfig{UserService: s.userService}

	// Tournament routes - optional auth (personalized data when auth present)
	tournamentGroup := v1.Group("/tournaments",
		auth.OptionalMiddleware(*s.authConfig),
		auth.OptionalEnsureUserMiddleware(ensureUserCfg),
	)
	tournamentService := tournaments.NewService(s.db)
	tournamentHandler := tournaments.NewHandler(tournamentService)
	tournamentHandler.RegisterRoutesWithGroup(tournamentGroup)

	// Golfer routes - public
	// golferGroup := v1.Group("/golfers")

	// League routes - requires auth and user provisioning
	// leagueGroup := v1.Group("/leagues",
	//     auth.Middleware(*s.authConfig),
	//     auth.EnsureUserMiddleware(ensureUserCfg),
	// )

	// Pick routes - requires auth and user provisioning
	// pickGroup := v1.Group("/picks",
	//     auth.Middleware(*s.authConfig),
	//     auth.EnsureUserMiddleware(ensureUserCfg),
	// )

	// User routes - requires auth and user provisioning
	// userGroup := v1.Group("/users",
	//     auth.Middleware(*s.authConfig),
	//     auth.EnsureUserMiddleware(ensureUserCfg),
	// )
}

// healthCheck returns the health status of the API.
func (s *Server) healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":      "healthy",
		"service":     "greenrats-api",
		"sse_clients": s.sseBroker.ClientCount(),
	})
}

// apiInfo returns basic API information.
func (s *Server) apiInfo(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"name":    "GreenRats API",
		"version": "1.0.0",
	})
}
