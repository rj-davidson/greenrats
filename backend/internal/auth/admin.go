package auth

import (
	"github.com/gofiber/fiber/v2"
)

func RequireAdminMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := GetDBUser(c)
		if user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "authentication required",
			})
		}
		if !user.IsAdmin {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "admin access required",
			})
		}
		return c.Next()
	}
}
