package testutil

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/ent"
)

const (
	UserIDKey    = "user_id"
	UserEmailKey = "user_email"
	UserNameKey  = "user_name"
	DBUserKey    = "db_user"
	DBUserIDKey  = "db_user_id"
)

type AuthContext struct {
	UserID      string
	WorkOSID    string
	Email       string
	DisplayName string
	DBUser      *ent.User
}

func DefaultAuthContext() *AuthContext {
	return &AuthContext{
		UserID:      "user_01HXYZ123456789ABCDEFGH",
		WorkOSID:    "user_01HXYZ123456789ABCDEFGH",
		Email:       "test@example.com",
		DisplayName: "Test User",
	}
}

func (ac *AuthContext) WithUserID(id string) *AuthContext {
	ac.UserID = id
	ac.WorkOSID = id
	return ac
}

func (ac *AuthContext) WithEmail(email string) *AuthContext {
	ac.Email = email
	return ac
}

func (ac *AuthContext) WithDisplayName(name string) *AuthContext {
	ac.DisplayName = name
	return ac
}

func (ac *AuthContext) WithDBUser(user *ent.User) *AuthContext {
	ac.DBUser = user
	return ac
}

func InjectAuth(ac *AuthContext) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if ac == nil {
			return c.Next()
		}

		c.Locals(UserIDKey, ac.UserID)
		c.Locals(UserEmailKey, ac.Email)
		c.Locals(UserNameKey, ac.DisplayName)

		if ac.DBUser != nil {
			c.Locals(DBUserKey, ac.DBUser)
			c.Locals(DBUserIDKey, ac.DBUser.ID)
		}

		return c.Next()
	}
}

func NoAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Next()
	}
}

func InjectDBUser(user *ent.User) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if user != nil {
			c.Locals(DBUserKey, user)
			c.Locals(DBUserIDKey, user.ID)
		}
		return c.Next()
	}
}

func InjectDBUserID(id uuid.UUID) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals(DBUserIDKey, id)
		return c.Next()
	}
}
