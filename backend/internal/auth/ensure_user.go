package auth

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"github.com/rj-davidson/greenrats/internal/features/users"
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
		workosID := GetUserID(c)
		if workosID == "" {
			log.Printf("[ENSURE_USER] No workos ID in context - auth middleware may not have run")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "authentication required",
			})
		}

		email := GetUserEmail(c)
		if email == "" {
			email = workosID + "@placeholder.greenrats.app"
		}

		result, err := cfg.UserService.GetOrCreate(c.Context(), users.GetOrCreateParams{
			WorkOSID: workosID,
			Email:    email,
		})
		if err != nil {
			log.Printf("[ENSURE_USER] Failed to get/create user: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to provision user",
			})
		}

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

		// Get or create the database user
		result, err := cfg.UserService.GetOrCreate(c.Context(), users.GetOrCreateParams{
			WorkOSID: workosID,
			Email:    email,
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
