package admin

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/internal/auth"
)

func TestHandler_SyncCourses(t *testing.T) {
	t.Run("returns 401 when no user in context", func(t *testing.T) {
		app := fiber.New()
		app.Use(auth.RequireAdminMiddleware())
		app.Post("/admin/automations/sync-courses", func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusAccepted).JSON(TriggerResponse{Message: "ok"})
		})

		req := httptest.NewRequest("POST", "/admin/automations/sync-courses", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("returns 403 when user is not admin", func(t *testing.T) {
		app := fiber.New()
		mockUser := &ent.User{
			ID:        uuid.Must(uuid.NewV4()),
			Email:     "user@example.com",
			IsAdmin:   false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		app.Use(func(c *fiber.Ctx) error {
			c.Locals(auth.DBUserKey, mockUser)
			return c.Next()
		})
		app.Use(auth.RequireAdminMiddleware())
		app.Post("/admin/automations/sync-courses", func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusAccepted).JSON(TriggerResponse{Message: "ok"})
		})

		req := httptest.NewRequest("POST", "/admin/automations/sync-courses", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	})

	t.Run("returns 202 when user is admin", func(t *testing.T) {
		app := fiber.New()
		mockUser := &ent.User{
			ID:        uuid.Must(uuid.NewV4()),
			Email:     "admin@example.com",
			IsAdmin:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		app.Use(func(c *fiber.Ctx) error {
			c.Locals(auth.DBUserKey, mockUser)
			return c.Next()
		})
		app.Use(auth.RequireAdminMiddleware())
		app.Post("/admin/automations/sync-courses", func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusAccepted).JSON(TriggerResponse{Message: "ok"})
		})

		req := httptest.NewRequest("POST", "/admin/automations/sync-courses", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusAccepted, resp.StatusCode)
	})
}

func TestHandler_SyncGolferSeasonStats(t *testing.T) {
	t.Run("returns 401 when no user in context", func(t *testing.T) {
		app := fiber.New()
		app.Use(auth.RequireAdminMiddleware())
		app.Post("/admin/automations/sync-golfer-stats", func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusAccepted).JSON(TriggerResponse{Message: "ok"})
		})

		req := httptest.NewRequest("POST", "/admin/automations/sync-golfer-stats", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("returns 403 when user is not admin", func(t *testing.T) {
		app := fiber.New()
		mockUser := &ent.User{
			ID:        uuid.Must(uuid.NewV4()),
			Email:     "user@example.com",
			IsAdmin:   false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		app.Use(func(c *fiber.Ctx) error {
			c.Locals(auth.DBUserKey, mockUser)
			return c.Next()
		})
		app.Use(auth.RequireAdminMiddleware())
		app.Post("/admin/automations/sync-golfer-stats", func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusAccepted).JSON(TriggerResponse{Message: "ok"})
		})

		req := httptest.NewRequest("POST", "/admin/automations/sync-golfer-stats", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	})

	t.Run("returns 202 when user is admin", func(t *testing.T) {
		app := fiber.New()
		mockUser := &ent.User{
			ID:        uuid.Must(uuid.NewV4()),
			Email:     "admin@example.com",
			IsAdmin:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		app.Use(func(c *fiber.Ctx) error {
			c.Locals(auth.DBUserKey, mockUser)
			return c.Next()
		})
		app.Use(auth.RequireAdminMiddleware())
		app.Post("/admin/automations/sync-golfer-stats", func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusAccepted).JSON(TriggerResponse{Message: "ok"})
		})

		req := httptest.NewRequest("POST", "/admin/automations/sync-golfer-stats", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusAccepted, resp.StatusCode)
	})
}
