package users

import (
	"log/slog"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/internal/email"
)

var displayNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

type Handler struct {
	service *Service
	email   *email.Client
	logger  *slog.Logger
}

func NewHandler(service *Service, emailClient *email.Client, logger *slog.Logger) *Handler {
	return &Handler{service: service, email: emailClient, logger: logger}
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
	users := router.Group("/users")
	h.RegisterRoutesWithGroup(users)
}

func (h *Handler) RegisterRoutesWithGroup(group fiber.Router) {
	group.Get("/me", h.GetMe)
	group.Get("/me/pending-actions", h.GetPendingActions)
	group.Post("/me/display-name", h.SetDisplayName)
	group.Get("/check-display-name", h.CheckDisplayName)
}

const DBUserKey = "db_user"

func (h *Handler) GetMe(c *fiber.Ctx) error {
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

func (h *Handler) SetDisplayName(c *fiber.Ctx) error {
	user, ok := c.Locals(DBUserKey).(*ent.User)
	if !ok || user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
	}

	var req SetDisplayNameRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

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
		h.logger.Error("failed to set display name", "user_id", user.ID, "error", err)
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

	h.logger.Info("display name set", "user_id", user.ID, "display_name", displayName)

	if h.email != nil {
		go func() {
			if err := h.email.SendWelcome(updated.Email, email.WelcomeData{
				DisplayName: displayName,
			}); err != nil {
				h.logger.Error("failed to send welcome email", "user_id", user.ID, "error", err)
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

func (h *Handler) CheckDisplayName(c *fiber.Ctx) error {
	name := strings.TrimSpace(c.Query("name"))
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "name query parameter is required",
		})
	}

	available, err := h.service.IsDisplayNameAvailable(c.UserContext(), name)
	if err != nil {
		h.logger.Error("failed to check display name availability", "name", name, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to check display name availability",
		})
	}

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
		h.logger.Error("failed to get pending actions", "user_id", user.ID, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get pending actions",
		})
	}

	return c.JSON(resp)
}
