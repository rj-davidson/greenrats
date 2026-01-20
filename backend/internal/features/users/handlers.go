package users

import (
	"log"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/internal/email"
)

var displayNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// Handler handles HTTP requests for users.
type Handler struct {
	service *Service
	email   *email.Client
}

// NewHandler creates a new user handler.
func NewHandler(service *Service, emailClient *email.Client) *Handler {
	return &Handler{service: service, email: emailClient}
}

// RegisterRoutes registers user routes on the given router.
func (h *Handler) RegisterRoutes(router fiber.Router) {
	users := router.Group("/users")
	h.RegisterRoutesWithGroup(users)
}

// RegisterRoutesWithGroup registers user routes on an existing group.
func (h *Handler) RegisterRoutesWithGroup(group fiber.Router) {
	group.Get("/me", h.GetMe)
	group.Get("/me/pending-actions", h.GetPendingActions)
	group.Post("/me/display-name", h.SetDisplayName)
	group.Get("/check-display-name", h.CheckDisplayName)
}

// DBUserKey is the context key for the database user (set by auth middleware).
const DBUserKey = "db_user"

// GetMe handles GET /users/me - returns the current authenticated user.
func (h *Handler) GetMe(c *fiber.Ctx) error {
	// Get the authenticated user from context (already provisioned by middleware)
	user, ok := c.Locals(DBUserKey).(*ent.User)
	if !ok || user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
	}

	return c.JSON(UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		IsAdmin:     user.IsAdmin,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	})
}

// SetDisplayName handles POST /users/me/display-name - sets the user's display name.
func (h *Handler) SetDisplayName(c *fiber.Ctx) error {
	log.Printf("[USERS] SetDisplayName request received")

	user, ok := c.Locals(DBUserKey).(*ent.User)
	if !ok || user == nil {
		log.Printf("[USERS] SetDisplayName: no authenticated user in context")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
	}

	log.Printf("[USERS] SetDisplayName: user=%s, current_display_name=%v", user.ID, user.DisplayName)

	var req SetDisplayNameRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("[USERS] SetDisplayName: failed to parse body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	log.Printf("[USERS] SetDisplayName: requested display_name=%q", req.DisplayName)

	// Validate display name
	displayName := strings.TrimSpace(req.DisplayName)
	if displayName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "display name is required",
		})
	}

	if len(displayName) < 3 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "display name must be at least 3 characters",
		})
	}

	if len(displayName) > 20 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "display name must be at most 20 characters",
		})
	}

	if !displayNameRegex.MatchString(displayName) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "display name can only contain letters, numbers, and underscores",
		})
	}

	updated, err := h.service.SetDisplayName(c.UserContext(), user.ID.String(), displayName)
	if err != nil {
		log.Printf("[USERS] SetDisplayName: service error: %v", err)
		if strings.Contains(err.Error(), "already set") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "display name is already set and cannot be changed",
			})
		}
		if strings.Contains(err.Error(), "already taken") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "display name is already taken",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to set display name",
		})
	}

	log.Printf("[USERS] SetDisplayName: success user=%s display_name=%s", user.ID, displayName)

	if h.email != nil {
		go func() {
			if err := h.email.SendWelcome(updated.Email, email.WelcomeData{
				DisplayName: displayName,
			}); err != nil {
				log.Printf("[USERS] Failed to send welcome email: %v", err)
			}
		}()
	}

	return c.JSON(UserResponse{
		ID:          updated.ID,
		Email:       updated.Email,
		DisplayName: updated.DisplayName,
		IsAdmin:     updated.IsAdmin,
		CreatedAt:   updated.CreatedAt,
		UpdatedAt:   updated.UpdatedAt,
	})
}

// CheckDisplayName handles GET /users/check-display-name - checks if a display name is available.
func (h *Handler) CheckDisplayName(c *fiber.Ctx) error {
	log.Printf("[USERS] CheckDisplayName request received")

	name := strings.TrimSpace(c.Query("name"))
	if name == "" {
		log.Printf("[USERS] CheckDisplayName: missing name parameter")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "name query parameter is required",
		})
	}

	log.Printf("[USERS] CheckDisplayName: checking name=%q", name)

	available, err := h.service.IsDisplayNameAvailable(c.UserContext(), name)
	if err != nil {
		log.Printf("[USERS] CheckDisplayName: error checking availability: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to check display name availability",
		})
	}

	log.Printf("[USERS] CheckDisplayName: name=%q available=%v", name, available)

	return c.JSON(CheckDisplayNameResponse{
		Available: available,
		Name:      name,
	})
}

func (h *Handler) GetPendingActions(c *fiber.Ctx) error {
	user, ok := c.Locals(DBUserKey).(*ent.User)
	if !ok || user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
	}

	resp, err := h.service.GetPendingActions(c.UserContext(), user.ID)
	if err != nil {
		log.Printf("[USERS] GetPendingActions: error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get pending actions",
		})
	}

	return c.JSON(resp)
}
