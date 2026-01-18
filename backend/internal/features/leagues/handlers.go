package leagues

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/internal/auth"
)

// Handler handles HTTP requests for leagues.
type Handler struct {
	service *Service
}

// NewHandler creates a new league handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers league routes on the given router.
func (h *Handler) RegisterRoutes(router fiber.Router) {
	leagues := router.Group("/leagues")
	h.RegisterRoutesWithGroup(leagues)
}

// RegisterRoutesWithGroup registers league routes on an existing group.
func (h *Handler) RegisterRoutesWithGroup(group fiber.Router) {
	group.Post("/", h.Create)
	group.Get("/", h.ListUserLeagues)
	group.Get("/:id", h.GetByID)
}

// Create handles POST /leagues
func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateLeagueRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate name
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "league name is required",
		})
	}

	// Get the authenticated user's ID
	userID := auth.GetDBUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
	}

	league, err := h.service.Create(c.Context(), CreateParams{
		Name:   name,
		UserID: userID,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create league",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(CreateLeagueResponse{
		League: *league,
	})
}

// ListUserLeagues handles GET /leagues
func (h *Handler) ListUserLeagues(c *fiber.Ctx) error {
	// Get the authenticated user's ID
	userID := auth.GetDBUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
	}

	resp, err := h.service.ListUserLeagues(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list leagues",
		})
	}

	return c.JSON(resp)
}

// GetByID handles GET /leagues/:id
func (h *Handler) GetByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "league id is required",
		})
	}

	id, err := uuid.FromString(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid league id",
		})
	}

	// Get the authenticated user's ID for role info
	userID := auth.GetDBUserID(c)

	var league *League
	if userID != uuid.Nil {
		league, err = h.service.GetByIDWithRole(c.Context(), id, userID)
	} else {
		league, err = h.service.GetByID(c.Context(), id)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get league",
		})
	}

	if league == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "league not found",
		})
	}

	return c.JSON(GetLeagueResponse{League: *league})
}
