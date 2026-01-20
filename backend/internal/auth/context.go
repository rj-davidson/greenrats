package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/ent"
)

// Context keys for storing user information.
const (
	UserIDKey    = "user_id"
	UserEmailKey = "user_email"
	UserNameKey  = "user_name"
	ClaimsKey    = "claims"
	DBUserKey    = "db_user"
	DBUserIDKey  = "db_user_id"
)

// GetUserID retrieves the user ID from the request context.
func GetUserID(c *fiber.Ctx) string {
	if id, ok := c.Locals(UserIDKey).(string); ok {
		return id
	}
	return ""
}

// GetUserEmail retrieves the user email from the request context.
func GetUserEmail(c *fiber.Ctx) string {
	if email, ok := c.Locals(UserEmailKey).(string); ok {
		return email
	}
	return ""
}

// GetUserName retrieves the user name from the request context.
func GetUserName(c *fiber.Ctx) string {
	if name, ok := c.Locals(UserNameKey).(string); ok {
		return name
	}
	return ""
}

// GetClaims retrieves the full JWT claims from the request context.
func GetClaims(c *fiber.Ctx) *Claims {
	if claims, ok := c.Locals(ClaimsKey).(*Claims); ok {
		return claims
	}
	return nil
}

// IsAuthenticated returns true if the request has valid authentication.
func IsAuthenticated(c *fiber.Ctx) bool {
	return GetUserID(c) != ""
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

// IsAdmin returns true if the authenticated user has admin privileges.
func IsAdmin(c *fiber.Ctx) bool {
	user := GetDBUser(c)
	return user != nil && user.IsAdmin
}
