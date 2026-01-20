package server

import (
	"time"

	"github.com/getsentry/sentry-go"
	sentryfiber "github.com/getsentry/sentry-go/fiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func (s *Server) setupMiddleware() {
	s.app.Use(requestid.New())

	s.app.Use(logger.New(logger.Config{
		Format:     "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path} | ${error}\n",
		TimeFormat: "2006-01-02 15:04:05",
	}))

	s.app.Use(sentryfiber.New(sentryfiber.Options{
		Repanic:         true,
		WaitForDelivery: false,
		Timeout:         2 * time.Second,
	}))
	s.app.Use(func(c *fiber.Ctx) error {
		if hub := sentry.GetHubFromContext(c.UserContext()); hub != nil {
			if id, ok := c.Locals("requestid").(string); ok && id != "" {
				hub.Scope().SetTag("request_id", id)
			}
		}
		return c.Next()
	})

	s.app.Use(fiberrecover.New(fiberrecover.Config{
		EnableStackTrace: s.config.IsDevelopment(),
	}))

	// CORS middleware
	s.app.Use(cors.New(cors.Config{
		AllowOrigins:     s.corsOrigins(),
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-User-Email,X-User-Name",
		AllowCredentials: true,
		MaxAge:           86400,
	}))
}

func (s *Server) corsOrigins() string {
	if s.config.IsDevelopment() {
		return "http://localhost:3000,http://127.0.0.1:3000"
	}
	return "https://greenrats.com"
}
