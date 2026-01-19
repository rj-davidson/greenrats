package users

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

func TestHandler_GetMe(t *testing.T) {
	t.Run("returns user when authenticated", func(t *testing.T) {
		app := fiber.New()
		handler := NewHandler(nil, nil) // service not needed for this test

		// Create a mock user
		userID := uuid.Must(uuid.NewV4())
		displayName := "Test User"
		mockUser := &ent.User{
			ID:          userID,
			Email:       "test@example.com",
			DisplayName: &displayName,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Set up route with middleware that injects user into context
		app.Get("/users/me", func(c *fiber.Ctx) error {
			c.Locals(DBUserKey, mockUser)
			return c.Next()
		}, handler.GetMe)

		req := httptest.NewRequest("GET", "/users/me", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var userResp UserResponse
		err = json.Unmarshal(body, &userResp)
		require.NoError(t, err)

		assert.Equal(t, userID, userResp.ID)
		assert.Equal(t, "test@example.com", userResp.Email)
		require.NotNil(t, userResp.DisplayName)
		assert.Equal(t, "Test User", *userResp.DisplayName)
	})

	t.Run("returns 401 when not authenticated", func(t *testing.T) {
		app := fiber.New()
		handler := NewHandler(nil, nil)

		app.Get("/users/me", handler.GetMe)

		req := httptest.NewRequest("GET", "/users/me", nil)
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
}

func TestHandler_RegisterRoutesWithGroup(t *testing.T) {
	app := fiber.New()
	handler := NewHandler(nil, nil)

	group := app.Group("/users")
	handler.RegisterRoutesWithGroup(group)

	// Verify route is registered by checking the stack
	routes := app.GetRoutes()
	var found bool
	for _, route := range routes {
		if route.Path == "/users/me" && route.Method == "GET" {
			found = true
			break
		}
	}
	assert.True(t, found, "GET /users/me route should be registered")
}
