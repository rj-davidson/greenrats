package users

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"github.com/rj-davidson/greenrats/ent"
)

// Handler handles HTTP requests for users.
type Handler struct {
	service *Service
}

// NewHandler creates a new user handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers user routes on the given router.
func (h *Handler) RegisterRoutes(router fiber.Router) {
	users := router.Group("/users")
	h.RegisterRoutesWithGroup(users)
}

// RegisterRoutesWithGroup registers user routes on an existing group.
func (h *Handler) RegisterRoutesWithGroup(group fiber.Router) {
	group.Get("/me", h.GetMe)
}

// DBUserKey is the context key for the database user (set by auth middleware).
const DBUserKey = "db_user"

// GetMe handles GET /users/me - returns the current authenticated user.
func (h *Handler) GetMe(c *fiber.Ctx) error {
	log.Printf("[USER_HANDLER] GetMe called")

	// Get the authenticated user from context (already provisioned by middleware)
	user, ok := c.Locals(DBUserKey).(*ent.User)
	if !ok || user == nil {
		log.Printf("[USER_HANDLER] No user in context - ok=%v, user=%v", ok, user)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
	}

	log.Printf("[USER_HANDLER] Returning user: id=%s, email=%s", user.ID, user.Email)

	return c.JSON(UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	})
}
