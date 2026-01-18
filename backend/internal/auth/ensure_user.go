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
		log.Printf("[ENSURE_USER] Processing request: %s %s", c.Method(), c.Path())

		// Check if auth middleware has set user info
		workosID := GetUserID(c)
		if workosID == "" {
			log.Printf("[ENSURE_USER] No workos ID in context - auth middleware may not have run")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "authentication required",
			})
		}

		email := GetUserEmail(c)
		name := GetUserName(c)

		log.Printf("[ENSURE_USER] Found user in context: workosID=%s, email=%s, name=%s", workosID, email, name)

		// WorkOS access tokens don't contain email/name - only the user ID.
		// Use placeholder values if missing (user can update later, or we fetch from WorkOS API)
		if email == "" {
			email = workosID + "@placeholder.greenrats.app"
			log.Printf("[ENSURE_USER] Email empty, using placeholder: %s", email)
		}

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
			log.Printf("[ENSURE_USER] Failed to get/create user: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to provision user",
			})
		}

		log.Printf("[ENSURE_USER] User provisioned: id=%s, created=%v", result.User.ID, result.Created)

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
