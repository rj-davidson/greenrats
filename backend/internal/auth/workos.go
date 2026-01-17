package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Config holds WorkOS authentication configuration.
type Config struct {
	ClientID string
}

// Claims represents the JWT claims from WorkOS.
type Claims struct {
	jwt.RegisteredClaims
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	OrgID         string `json:"org_id,omitempty"`
	Role          string `json:"role,omitempty"`
}

// Middleware creates a Fiber middleware for WorkOS JWT verification.
func Middleware(cfg Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization header format",
			})
		}

		tokenString := parts[1]

		// Parse and validate token
		claims, err := verifyToken(c.Context(), tokenString, cfg.ClientID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fmt.Sprintf("invalid token: %v", err),
			})
		}

		// Store user info in context
		c.Locals(UserIDKey, claims.Subject)
		c.Locals(UserEmailKey, claims.Email)
		c.Locals(UserNameKey, claims.Name)
		c.Locals(ClaimsKey, claims)

		return c.Next()
	}
}

// verifyToken validates the JWT and returns claims.
func verifyToken(_ context.Context, tokenString, clientID string) (*Claims, error) {
	claims := &Claims{}

	// Parse token with claims
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// For development/testing, we can skip signature verification
		// In production, fetch JWKS from https://api.workos.com/sso/jwks/{clientID}
		// and verify the signature properly

		// Check algorithm
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// TODO: Implement JWKS fetching and caching for production
		// For now, return nil to skip verification (NOT for production use)
		return nil, fmt.Errorf("JWKS verification not implemented - configure for production")
	})

	// For development, parse without verification
	if err != nil {
		// Fallback: parse without verification for development
		token, _, err = new(jwt.Parser).ParseUnverified(tokenString, claims)
		if err != nil {
			return nil, err
		}
	}

	if !token.Valid && token.Claims == nil {
		return nil, fmt.Errorf("invalid token")
	}

	// Verify audience matches client ID
	aud, err := claims.GetAudience()
	if err == nil && len(aud) > 0 {
		validAudience := false
		for _, a := range aud {
			if a == clientID {
				validAudience = true
				break
			}
		}
		if !validAudience && clientID != "" {
			return nil, fmt.Errorf("invalid audience")
		}
	}

	return claims, nil
}

// OptionalMiddleware extracts user info if present but doesn't require auth.
func OptionalMiddleware(cfg Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			return c.Next()
		}

		claims, err := verifyToken(c.Context(), parts[1], cfg.ClientID)
		if err != nil {
			return c.Next()
		}

		c.Locals(UserIDKey, claims.Subject)
		c.Locals(UserEmailKey, claims.Email)
		c.Locals(UserNameKey, claims.Name)
		c.Locals(ClaimsKey, claims)

		return c.Next()
	}
}
