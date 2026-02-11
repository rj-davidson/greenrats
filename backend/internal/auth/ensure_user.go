package auth

import (
	"context"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/rj-davidson/greenrats/internal/features/users"
)

type EnsureUserConfig struct {
	UserService *users.Service
	Logger      *slog.Logger
}

const ensureUserTimeout = 8 * time.Second

func EnsureUserMiddleware(cfg EnsureUserConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		workosID := GetUserID(c)
		if workosID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "authentication required",
			})
		}

		email := GetUserEmail(c)
		if email == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "email claim missing from token",
			})
		}

		ctx, cancel := context.WithTimeout(c.UserContext(), ensureUserTimeout)
		defer cancel()

		// TEMP BACKLOAD FEATURE - Remove this block after Prothero League users have signed up.
		claimedUser, claimed, err := cfg.UserService.ClaimTempUser(ctx, workosID, email)
		if err != nil {
			cfg.Logger.Error("failed to claim temp user", "workos_id", workosID, "error", err)
		}
		if claimed && claimedUser != nil {
			c.Locals(DBUserKey, claimedUser)
			c.Locals(DBUserIDKey, claimedUser.ID)
			return c.Next()
		}
		// END TEMP BACKLOAD FEATURE

		result, err := cfg.UserService.GetOrCreate(ctx, users.GetOrCreateParams{
			WorkOSID: workosID,
			Email:    email,
		})
		if err != nil {
			cfg.Logger.Error("failed to get/create user", "workos_id", workosID, "error", err)
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

		ctx, cancel := context.WithTimeout(c.UserContext(), ensureUserTimeout)
		defer cancel()

		// Get or create the database user
		result, err := cfg.UserService.GetOrCreate(ctx, users.GetOrCreateParams{
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
