package auth

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/ent"
)

func TestRequireAdminMiddleware(t *testing.T) {
	t.Run("returns 401 when no user in context", func(t *testing.T) {
		app := fiber.New()
		app.Use(RequireAdminMiddleware())
		app.Get("/admin/test", func(c *fiber.Ctx) error {
			return c.SendString("OK")
		})

		req := httptest.NewRequest("GET", "/admin/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var errResp map[string]string
		err = json.Unmarshal(body, &errResp)
		require.NoError(t, err)

		assert.Equal(t, "authentication required", errResp["error"])
	})

	t.Run("returns 403 when user is not admin", func(t *testing.T) {
		app := fiber.New()

		userID := uuid.Must(uuid.NewV4())
		mockUser := &ent.User{
			ID:        userID,
			Email:     "user@example.com",
			IsAdmin:   false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		app.Use(func(c *fiber.Ctx) error {
			c.Locals(DBUserKey, mockUser)
			return c.Next()
		})
		app.Use(RequireAdminMiddleware())
		app.Get("/admin/test", func(c *fiber.Ctx) error {
			return c.SendString("OK")
		})

		req := httptest.NewRequest("GET", "/admin/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var errResp map[string]string
		err = json.Unmarshal(body, &errResp)
		require.NoError(t, err)

		assert.Equal(t, "admin access required", errResp["error"])
	})

	t.Run("allows request when user is admin", func(t *testing.T) {
		app := fiber.New()

		userID := uuid.Must(uuid.NewV4())
		mockUser := &ent.User{
			ID:        userID,
			Email:     "admin@example.com",
			IsAdmin:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		app.Use(func(c *fiber.Ctx) error {
			c.Locals(DBUserKey, mockUser)
			return c.Next()
		})
		app.Use(RequireAdminMiddleware())
		app.Get("/admin/test", func(c *fiber.Ctx) error {
			return c.SendString("OK")
		})

		req := httptest.NewRequest("GET", "/admin/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, "OK", string(body))
	})
}

func TestIsAdmin(t *testing.T) {
	t.Run("returns false when no user in context", func(t *testing.T) {
		app := fiber.New()
		var result bool

		app.Get("/test", func(c *fiber.Ctx) error {
			result = IsAdmin(c)
			return c.SendString("OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		_, err := app.Test(req)
		require.NoError(t, err)

		assert.False(t, result)
	})

	t.Run("returns false when user is not admin", func(t *testing.T) {
		app := fiber.New()
		var result bool

		mockUser := &ent.User{
			ID:      uuid.Must(uuid.NewV4()),
			IsAdmin: false,
		}

		app.Get("/test", func(c *fiber.Ctx) error {
			c.Locals(DBUserKey, mockUser)
			result = IsAdmin(c)
			return c.SendString("OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		_, err := app.Test(req)
		require.NoError(t, err)

		assert.False(t, result)
	})

	t.Run("returns true when user is admin", func(t *testing.T) {
		app := fiber.New()
		var result bool

		mockUser := &ent.User{
			ID:      uuid.Must(uuid.NewV4()),
			IsAdmin: true,
		}

		app.Get("/test", func(c *fiber.Ctx) error {
			c.Locals(DBUserKey, mockUser)
			result = IsAdmin(c)
			return c.SendString("OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		_, err := app.Test(req)
		require.NoError(t, err)

		assert.True(t, result)
	})
}
