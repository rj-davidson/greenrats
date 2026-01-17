package server

import (
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

// setupMiddleware configures all middleware for the server.
func (s *Server) setupMiddleware() {
	// Request ID middleware
	s.app.Use(requestid.New())

	// Logger middleware
	s.app.Use(logger.New(logger.Config{
		Format:     "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path} | ${error}\n",
		TimeFormat: "2006-01-02 15:04:05",
	}))

	// Recover middleware - recovers from panics
	s.app.Use(recover.New(recover.Config{
		EnableStackTrace: s.config.IsDevelopment(),
	}))

	// CORS middleware
	s.app.Use(cors.New(cors.Config{
		AllowOrigins:     s.corsOrigins(),
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
		MaxAge:           86400,
	}))
}

// corsOrigins returns allowed origins based on environment.
func (s *Server) corsOrigins() string {
	if s.config.IsDevelopment() {
		return "http://localhost:3000,http://127.0.0.1:3000"
	}
	// In production, this should be configured via environment variable
	return "https://greenrats.com"
}
