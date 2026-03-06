package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/ent"
)

const (
	UserIDKey    = "user_id"
	UserEmailKey = "user_email"
	UserNameKey  = "user_name"
	ClaimsKey    = "claims"
	DBUserKey    = "db_user"
	DBUserIDKey  = "db_user_id"
)

func GetUserID(c *fiber.Ctx) string {
	if id, ok := c.Locals(UserIDKey).(string); ok {
		return id
	}
	return ""
}

func GetUserEmail(c *fiber.Ctx) string {
	if email, ok := c.Locals(UserEmailKey).(string); ok {
		return email
	}
	return ""
}

func GetUserName(c *fiber.Ctx) string {
	if name, ok := c.Locals(UserNameKey).(string); ok {
		return name
	}
	return ""
}

func GetClaims(c *fiber.Ctx) *Claims {
	if claims, ok := c.Locals(ClaimsKey).(*Claims); ok {
		return claims
	}
	return nil
}

func IsAuthenticated(c *fiber.Ctx) bool {
	return GetUserID(c) != ""
}

func GetDBUser(c *fiber.Ctx) *ent.User {
	if user, ok := c.Locals(DBUserKey).(*ent.User); ok {
		return user
	}
	return nil
}

func GetDBUserID(c *fiber.Ctx) uuid.UUID {
	if id, ok := c.Locals(DBUserIDKey).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

func IsAdmin(c *fiber.Ctx) bool {
	user := GetDBUser(c)
	return user != nil && user.IsAdmin
}
