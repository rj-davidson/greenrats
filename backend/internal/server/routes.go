package server

import (
	"github.com/gofiber/fiber/v2"

	"github.com/rj-davidson/greenrats/internal/auth"
	"github.com/rj-davidson/greenrats/internal/features/admin"
	"github.com/rj-davidson/greenrats/internal/features/leaderboards"
	"github.com/rj-davidson/greenrats/internal/features/leagues"
	"github.com/rj-davidson/greenrats/internal/features/picks"
	"github.com/rj-davidson/greenrats/internal/features/tournaments"
	"github.com/rj-davidson/greenrats/internal/features/users"
	"github.com/rj-davidson/greenrats/internal/sync"
)

func (s *Server) setupRoutes() {
	s.app.Get("/health", s.healthCheck)
	s.app.Get("/health/ingest", s.ingestHealthCheck)

	v1 := s.app.Group("/api/v1")

	v1.Get("/", s.apiInfo)

	v1.Get("/sse/:topic", s.sseHandler.HandleSSE)
	v1.Get("/tournaments/:id/live", s.sseHandler.HandleTournamentSSE)

	ensureUserCfg := auth.EnsureUserConfig{UserService: s.userService}

	tournamentGroup := v1.Group("/tournaments",
		auth.OptionalMiddleware(*s.authConfig),
		auth.OptionalEnsureUserMiddleware(ensureUserCfg),
	)
	tournamentService := tournaments.NewService(s.db)
	tournamentHandler := tournaments.NewHandler(tournamentService)
	tournamentHandler.RegisterRoutesWithGroup(tournamentGroup)

	leagueGroup := v1.Group("/leagues",
		auth.Middleware(*s.authConfig),
		auth.EnsureUserMiddleware(ensureUserCfg),
	)
	leagueService := leagues.NewService(s.db, s.config.CurrentSeason)
	leagueHandler := leagues.NewHandler(leagueService, s.emailClient)
	leagueHandler.RegisterRoutesWithGroup(leagueGroup)

	leaderboardService := leaderboards.NewService(s.db)
	leaderboardHandler := leaderboards.NewHandler(leaderboardService)
	leaderboardHandler.RegisterLeagueRoutes(leagueGroup)

	pickService := picks.NewService(s.db)
	pickHandler := picks.NewHandler(pickService)
	pickHandler.RegisterRoutes(v1.Group("",
		auth.Middleware(*s.authConfig),
		auth.EnsureUserMiddleware(ensureUserCfg),
	))
	pickHandler.RegisterLeagueRoutes(leagueGroup)
	pickHandler.RegisterTournamentRoutes(tournamentGroup)

	userGroup := v1.Group("/users",
		auth.Middleware(*s.authConfig),
		auth.EnsureUserMiddleware(ensureUserCfg),
	)
	userHandler := users.NewHandler(s.userService, s.emailClient)
	userHandler.RegisterRoutesWithGroup(userGroup)

	adminGroup := v1.Group("/admin",
		auth.Middleware(*s.authConfig),
		auth.EnsureUserMiddleware(ensureUserCfg),
		auth.RequireAdminMiddleware(),
	)
	adminService := admin.NewService(s.db)
	adminHandler := admin.NewHandler(adminService, s.adminIngestService, s.logger)
	adminHandler.RegisterRoutesWithGroup(adminGroup)
}

func (s *Server) healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":      "healthy",
		"service":     "greenrats-api",
		"sse_clients": s.sseBroker.ClientCount(),
	})
}

func (s *Server) apiInfo(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"name":    "GreenRats API",
		"version": "1.0.0",
	})
}

func (s *Server) ingestHealthCheck(c *fiber.Ctx) error {
	status, err := sync.GetHealthStatus(c.Context(), s.db)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get sync health status",
		})
	}

	if !status.Healthy {
		return c.Status(fiber.StatusServiceUnavailable).JSON(status)
	}

	return c.JSON(status)
}
