package picks

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/internal/auth"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
	picks := router.Group("/picks")
	picks.Post("/", h.Create)
	picks.Get("/", h.GetUserPicks)
}

func (h *Handler) RegisterLeagueRoutes(group fiber.Router) {
	group.Get("/:id/picks", h.GetLeaguePicks)
	group.Get("/:id/available-golfers", h.GetAvailableGolfers)
	group.Put("/:id/picks/:pickId", h.OverridePick)
}

func (h *Handler) RegisterTournamentRoutes(group fiber.Router) {
	group.Get("/:id/pick-window", h.GetPickWindow)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	userID := auth.GetDBUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
	}

	var req CreatePickRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.TournamentID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "tournament_id is required",
		})
	}
	if req.GolferID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "golfer_id is required",
		})
	}
	if req.LeagueID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "league_id is required",
		})
	}

	pick, err := h.service.Create(c.Context(), CreateParams{
		UserID:       userID,
		TournamentID: req.TournamentID,
		GolferID:     req.GolferID,
		LeagueID:     req.LeagueID,
	})
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(CreatePickResponse{Pick: *pick})
}

func (h *Handler) GetUserPicks(c *fiber.Ctx) error {
	userID := auth.GetDBUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
	}

	var leagueID uuid.UUID
	if leagueIDStr := c.Query("league_id"); leagueIDStr != "" {
		id, err := uuid.FromString(leagueIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid league_id",
			})
		}
		leagueID = id
	}

	var seasonYear int
	if seasonYearStr := c.Query("season_year"); seasonYearStr != "" {
		year, err := strconv.Atoi(seasonYearStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid season_year",
			})
		}
		seasonYear = year
	}

	resp, err := h.service.GetUserPicks(c.Context(), userID, leagueID, seasonYear)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get picks",
		})
	}

	return c.JSON(resp)
}

func (h *Handler) GetLeaguePicks(c *fiber.Ctx) error {
	userID := auth.GetDBUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
	}

	leagueID, err := uuid.FromString(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid league id",
		})
	}

	tournamentIDStr := c.Query("tournament_id")
	if tournamentIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "tournament_id query parameter is required",
		})
	}

	tournamentID, err := uuid.FromString(tournamentIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid tournament_id",
		})
	}

	resp, err := h.service.GetLeaguePicks(c.Context(), leagueID, tournamentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get league picks",
		})
	}

	return c.JSON(resp)
}

func (h *Handler) GetAvailableGolfers(c *fiber.Ctx) error {
	userID := auth.GetDBUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
	}

	leagueID, err := uuid.FromString(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid league id",
		})
	}

	tournamentIDStr := c.Query("tournament_id")
	if tournamentIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "tournament_id query parameter is required",
		})
	}

	tournamentID, err := uuid.FromString(tournamentIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid tournament_id",
		})
	}

	resp, err := h.service.GetAvailableGolfers(c.Context(), userID, leagueID, tournamentID)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return c.JSON(resp)
}

func (h *Handler) GetPickWindow(c *fiber.Ctx) error {
	tournamentID, err := uuid.FromString(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid tournament id",
		})
	}

	status, err := h.service.CanMakePick(c.Context(), tournamentID)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return c.JSON(status)
}

func (h *Handler) OverridePick(c *fiber.Ctx) error {
	userID := auth.GetDBUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
	}

	leagueID, err := uuid.FromString(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid league id",
		})
	}

	pickID, err := uuid.FromString(c.Params("pickId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid pick id",
		})
	}

	var req OverridePickRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.GolferID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "golfer_id is required",
		})
	}

	pick, err := h.service.OverridePick(c.Context(), OverridePickParams{
		LeagueID:       leagueID,
		PickID:         pickID,
		NewGolferID:    req.GolferID,
		CommissionerID: userID,
	})
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return c.JSON(OverridePickResponse{Pick: *pick})
}

func (h *Handler) handleServiceError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrTournamentNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "tournament not found",
		})
	case errors.Is(err, ErrGolferNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "golfer not found",
		})
	case errors.Is(err, ErrLeagueNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "league not found",
		})
	case errors.Is(err, ErrNotLeagueMember):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "you are not a member of this league",
		})
	case errors.Is(err, ErrPickWindowClosed):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "pick window is closed",
		})
	case errors.Is(err, ErrGolferNotInField):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "golfer is not in the tournament field",
		})
	case errors.Is(err, ErrGolferAlreadyUsed):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "golfer has already been used this season",
		})
	case errors.Is(err, ErrPickAlreadyExists):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "a pick already exists for this tournament",
		})
	case errors.Is(err, ErrTournamentNotUpcoming):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "tournament is not upcoming",
		})
	case errors.Is(err, ErrPickNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "pick not found",
		})
	case errors.Is(err, ErrNotCommissioner):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "only the commissioner can perform this action",
		})
	case errors.Is(err, ErrTournamentCompleted):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "tournament has already completed",
		})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error",
		})
	}
}
