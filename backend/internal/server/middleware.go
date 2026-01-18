package server

import (
	"github.com/getsentry/sentry-go"
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

	s.app.Use(sentryMiddleware())

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

func sentryMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		hub := sentry.CurrentHub().Clone()
		hub.Scope().SetTag("method", c.Method())
		hub.Scope().SetTag("path", c.Path())
		hub.Scope().SetExtra("request_id", c.Locals("requestid"))

		defer func() {
			if r := recover(); r != nil {
				hub.RecoverWithContext(c.Context(), r)
				panic(r)
			}
		}()

		err := c.Next()
		if err != nil {
			hub.CaptureException(err)
		}
		return err
	}
}
