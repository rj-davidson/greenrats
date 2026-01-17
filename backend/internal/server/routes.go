package server

import (
	"github.com/gofiber/fiber/v2"
)

// setupRoutes configures all routes for the server.
func (s *Server) setupRoutes() {
	// Health check endpoint (no prefix)
	s.app.Get("/health", s.healthCheck)

	// API v1 routes
	v1 := s.app.Group("/api/v1")

	// Public routes
	v1.Get("/", s.apiInfo)

	// SSE routes for live updates
	v1.Get("/sse/:topic", s.sseHandler.HandleSSE)
	v1.Get("/tournaments/:id/live", s.sseHandler.HandleTournamentSSE)

	// Tournament routes (will be expanded in feature scaffolding)
	// tournaments := v1.Group("/tournaments")

	// Golfer routes
	// golfers := v1.Group("/golfers")

	// League routes (requires auth)
	// leagues := v1.Group("/leagues")

	// Pick routes (requires auth)
	// picks := v1.Group("/picks")

	// User routes (requires auth)
	// users := v1.Group("/users")
}

// healthCheck returns the health status of the API.
func (s *Server) healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "healthy",
		"service": "greenrats-api",
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
