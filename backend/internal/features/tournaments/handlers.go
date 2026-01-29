package tournaments

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

// Handler handles HTTP requests for tournaments.
type Handler struct {
	service *Service
}

// NewHandler creates a new tournament handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers tournament routes on the given router.
func (h *Handler) RegisterRoutes(router fiber.Router) {
	tournaments := router.Group("/tournaments")
	h.RegisterRoutesWithGroup(tournaments)
}

// RegisterRoutesWithGroup registers tournament routes on an existing group.
func (h *Handler) RegisterRoutesWithGroup(group fiber.Router) {
	group.Get("/", h.List)
	group.Get("/active", h.GetActive)
	group.Get("/:id", h.GetByID)
	group.Get("/:id/leaderboard", h.GetLeaderboard)
	group.Get("/:id/field", h.GetField)
	group.Get("/:id/scorecard/:golferId", h.GetScorecard)
}

// List handles GET /tournaments
func (h *Handler) List(c *fiber.Ctx) error {
	var req ListTournamentsRequest
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid query parameters",
		})
	}

	resp, err := h.service.List(c.UserContext(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list tournaments",
		})
	}

	return c.JSON(resp)
}

// GetByID handles GET /tournaments/:id
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "tournament id is required",
		})
	}

	tournament, err := h.service.GetByID(c.UserContext(), id)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidTournamentID):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid tournament id",
			})
		case errors.Is(err, ErrTournamentNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "tournament not found",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to get tournament",
			})
		}
	}

	return c.JSON(GetTournamentResponse{Tournament: *tournament})
}

// GetActive handles GET /tournaments/active
func (h *Handler) GetActive(c *fiber.Ctx) error {
	tournament, err := h.service.GetActive(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get active tournament",
		})
	}

	if tournament == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "no active tournament",
		})
	}

	return c.JSON(GetTournamentResponse{Tournament: *tournament})
}

// GetLeaderboard handles GET /tournaments/:id/leaderboard
func (h *Handler) GetLeaderboard(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "tournament id is required",
		})
	}

	includeHoles := c.Query("include") == "holes"
	leagueID := c.Query("league_id")

	resp, err := h.service.GetLeaderboard(c.UserContext(), id, includeHoles, leagueID)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidTournamentID):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid tournament id",
			})
		case errors.Is(err, ErrTournamentNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "tournament not found",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to get leaderboard",
			})
		}
	}

	return c.JSON(resp)
}

// GetField handles GET /tournaments/:id/field
func (h *Handler) GetField(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "tournament id is required",
		})
	}

	resp, err := h.service.GetField(c.UserContext(), id)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidTournamentID):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid tournament id",
			})
		case errors.Is(err, ErrTournamentNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "tournament not found",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to get field",
			})
		}
	}

	return c.JSON(resp)
}

// GetScorecard handles GET /tournaments/:id/scorecard/:golferId
func (h *Handler) GetScorecard(c *fiber.Ctx) error {
	tournamentID := c.Params("id")
	golferID := c.Params("golferId")

	if tournamentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "tournament id is required",
		})
	}
	if golferID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "golfer id is required",
		})
	}

	resp, err := h.service.GetScorecard(c.UserContext(), tournamentID, golferID)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidTournamentID):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid tournament id",
			})
		case errors.Is(err, ErrInvalidGolferID):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid golfer id",
			})
		case errors.Is(err, ErrTournamentNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "tournament not found",
			})
		case errors.Is(err, ErrGolferNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "golfer not found",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to get scorecard",
			})
		}
	}

	return c.JSON(resp)
}
