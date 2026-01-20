package admin

import (
	"context"
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid/v5"
)

type Handler struct {
	service       *Service
	ingestService *IngestService
	logger        *slog.Logger
}

func NewHandler(service *Service, ingestService *IngestService, logger *slog.Logger) *Handler {
	return &Handler{
		service:       service,
		ingestService: ingestService,
		logger:        logger,
	}
}

func (h *Handler) RegisterRoutesWithGroup(group fiber.Router) {
	group.Get("/users", h.ListUsers)
	group.Get("/leagues", h.ListLeagues)
	group.Delete("/leagues/:id", h.DeleteLeague)
	group.Get("/tournaments", h.ListTournaments)

	automations := group.Group("/automations")
	automations.Post("/sync-tournaments", h.SyncTournaments)
	automations.Post("/sync-players", h.SyncPlayers)
	automations.Post("/sync-leaderboard/:tournamentId", h.SyncLeaderboard)
	automations.Post("/sync-earnings/:tournamentId", h.SyncEarnings)
}

func (h *Handler) ListUsers(c *fiber.Ctx) error {
	resp, err := h.service.ListUsers(c.UserContext())
	if err != nil {
		h.logger.Error("failed to list users", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list users",
		})
	}
	return c.JSON(resp)
}

func (h *Handler) ListLeagues(c *fiber.Ctx) error {
	resp, err := h.service.ListLeagues(c.UserContext())
	if err != nil {
		h.logger.Error("failed to list leagues", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list leagues",
		})
	}
	return c.JSON(resp)
}

func (h *Handler) DeleteLeague(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.FromString(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid league id",
		})
	}

	err = h.service.DeleteLeague(c.UserContext(), id)
	if errors.Is(err, ErrLeagueNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "league not found",
		})
	}
	if err != nil {
		h.logger.Error("failed to delete league", "error", err, "id", id)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete league",
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

func (h *Handler) ListTournaments(c *fiber.Ctx) error {
	resp, err := h.service.ListTournaments(c.UserContext())
	if err != nil {
		h.logger.Error("failed to list tournaments", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list tournaments",
		})
	}
	return c.JSON(resp)
}

func (h *Handler) SyncTournaments(c *fiber.Ctx) error {
	go func() {
		ctx := context.Background()
		if err := h.ingestService.SyncTournaments(ctx); err != nil {
			h.logger.Error("tournament sync failed", "error", err)
		}
	}()

	return c.Status(fiber.StatusAccepted).JSON(TriggerResponse{
		Message: "Tournament sync started",
	})
}

func (h *Handler) SyncPlayers(c *fiber.Ctx) error {
	go func() {
		ctx := context.Background()
		if err := h.ingestService.SyncPlayers(ctx); err != nil {
			h.logger.Error("player sync failed", "error", err)
		}
	}()

	return c.Status(fiber.StatusAccepted).JSON(TriggerResponse{
		Message: "Player sync started",
	})
}

func (h *Handler) SyncLeaderboard(c *fiber.Ctx) error {
	idParam := c.Params("tournamentId")
	id, err := uuid.FromString(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid tournament id",
		})
	}

	go func() {
		ctx := context.Background()
		if err := h.ingestService.SyncLeaderboard(ctx, id); err != nil {
			h.logger.Error("leaderboard sync failed", "error", err, "tournament_id", id)
		}
	}()

	return c.Status(fiber.StatusAccepted).JSON(TriggerResponse{
		Message: "Leaderboard sync started",
	})
}

func (h *Handler) SyncEarnings(c *fiber.Ctx) error {
	idParam := c.Params("tournamentId")
	id, err := uuid.FromString(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid tournament id",
		})
	}

	go func() {
		ctx := context.Background()
		if err := h.ingestService.SyncEarnings(ctx, id); err != nil {
			h.logger.Error("earnings sync failed", "error", err, "tournament_id", id)
		}
	}()

	return c.Status(fiber.StatusAccepted).JSON(TriggerResponse{
		Message: "Earnings sync started",
	})
}
