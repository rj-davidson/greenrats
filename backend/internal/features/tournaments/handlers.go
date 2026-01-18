package tournaments

import (
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
}

// List handles GET /tournaments
func (h *Handler) List(c *fiber.Ctx) error {
	var req ListTournamentsRequest
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid query parameters",
		})
	}

	resp, err := h.service.List(c.Context(), req)
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

	tournament, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get tournament",
		})
	}

	if tournament == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "tournament not found",
		})
	}

	return c.JSON(GetTournamentResponse{Tournament: *tournament})
}

// GetActive handles GET /tournaments/active
func (h *Handler) GetActive(c *fiber.Ctx) error {
	tournament, err := h.service.GetActive(c.Context())
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

	resp, err := h.service.GetLeaderboard(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get leaderboard",
		})
	}

	if resp == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "tournament not found",
		})
	}

	return c.JSON(resp)
}
