package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/internal/features/users"
)

// Context keys for database user information.
const (
	DBUserKey   = "db_user"
	DBUserIDKey = "db_user_id"
)

// EnsureUserConfig holds configuration for user provisioning middleware.
type EnsureUserConfig struct {
	UserService *users.Service
}

// EnsureUserMiddleware creates a middleware that provisions a database user.
// This middleware REQUIRES that auth middleware has already run and set user context.
// Returns 401 if no authenticated user is present.
func EnsureUserMiddleware(cfg EnsureUserConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if auth middleware has set user info
		workosID := GetUserID(c)
		if workosID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "authentication required",
			})
		}

		email := GetUserEmail(c)
		name := GetUserName(c)

		// Use email as display name if name is empty
		displayName := name
		if displayName == "" {
			displayName = email
		}

		// Get or create the database user
		result, err := cfg.UserService.GetOrCreate(c.Context(), users.GetOrCreateParams{
			WorkOSID:    workosID,
			Email:       email,
			DisplayName: displayName,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to provision user",
			})
		}

		// Store database user in context
		c.Locals(DBUserKey, result.User)
		c.Locals(DBUserIDKey, result.User.ID)

		return c.Next()
	}
}

// OptionalEnsureUserMiddleware provisions a database user if authenticated.
// If no auth is present, continues without setting db_user context.
func OptionalEnsureUserMiddleware(cfg EnsureUserConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if auth middleware has set user info
		workosID := GetUserID(c)
		if workosID == "" {
			// No auth, continue without provisioning
			return c.Next()
		}

		email := GetUserEmail(c)
		name := GetUserName(c)

		// Use email as display name if name is empty
		displayName := name
		if displayName == "" {
			displayName = email
		}

		// Get or create the database user
		result, err := cfg.UserService.GetOrCreate(c.Context(), users.GetOrCreateParams{
			WorkOSID:    workosID,
			Email:       email,
			DisplayName: displayName,
		})
		if err != nil {
			// For optional middleware, log error but continue
			// The handler can check if db_user is nil
			return c.Next()
		}

		// Store database user in context
		c.Locals(DBUserKey, result.User)
		c.Locals(DBUserIDKey, result.User.ID)

		return c.Next()
	}
}

// GetDBUser retrieves the database user from the request context.
func GetDBUser(c *fiber.Ctx) *ent.User {
	if user, ok := c.Locals(DBUserKey).(*ent.User); ok {
		return user
	}
	return nil
}

// GetDBUserID retrieves the database user ID from the request context.
func GetDBUserID(c *fiber.Ctx) uuid.UUID {
	if id, ok := c.Locals(DBUserIDKey).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}
