package auth

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

const (
	// WorkOS issuer base URL for JWT validation
	workOSIssuerBase = "https://api.workos.com/user_management/"
)

// Config holds WorkOS authentication configuration.
type Config struct {
	ClientID     string
	JWKSProvider *JWKSProvider
	SkipVerify   bool // Development only - skips signature verification
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
// Returns 401 if authentication is missing or invalid.
func Middleware(cfg Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization header format",
			})
		}

		tokenString := parts[1]

		claims, err := verifyToken(tokenString, cfg)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fmt.Sprintf("invalid token: %v", err),
			})
		}

		email := claims.Email
		name := claims.Name
		if email == "" {
			email = c.Get("X-User-Email")
		}
		if name == "" {
			name = c.Get("X-User-Name")
		}

		c.Locals(UserIDKey, claims.Subject)
		c.Locals(UserEmailKey, email)
		c.Locals(UserNameKey, name)
		c.Locals(ClaimsKey, claims)

		return c.Next()
	}
}

// OptionalMiddleware extracts user info if present but doesn't require auth.
// Allows requests to proceed without authentication.
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

		claims, err := verifyToken(parts[1], cfg)
		if err != nil {
			// For optional auth, silently continue on verification failure
			return c.Next()
		}

		c.Locals(UserIDKey, claims.Subject)
		c.Locals(UserEmailKey, claims.Email)
		c.Locals(UserNameKey, claims.Name)
		c.Locals(ClaimsKey, claims)

		return c.Next()
	}
}

// verifyToken validates the JWT and returns claims.
func verifyToken(tokenString string, cfg Config) (*Claims, error) {
	claims := &Claims{}

	// Development mode: parse without verification
	if cfg.SkipVerify {
		token, _, err := new(jwt.Parser).ParseUnverified(tokenString, claims)
		if err != nil {
			return nil, fmt.Errorf("failed to parse token: %w", err)
		}
		if token.Claims == nil {
			return nil, fmt.Errorf("invalid token claims")
		}
		return claims, nil
	}

	// Production mode: verify with JWKS
	if cfg.JWKSProvider == nil {
		return nil, fmt.Errorf("JWKS provider not configured")
	}

	// Build expected issuer: https://api.workos.com/user_management/{client_id}
	expectedIssuer := workOSIssuerBase + cfg.ClientID

	// Parse and verify token signature using JWKS
	token, err := jwt.ParseWithClaims(tokenString, claims, cfg.JWKSProvider.Keyfunc,
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithIssuer(expectedIssuer),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Note: WorkOS AuthKit tokens don't include an audience claim.
	// The client ID is embedded in the issuer URL, so issuer validation is sufficient.

	return claims, nil
}
