package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
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
		log.Printf("[AUTH] Processing request: %s %s", c.Method(), c.Path())

		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			log.Printf("[AUTH] Missing authorization header for %s %s", c.Method(), c.Path())
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		log.Printf("[AUTH] Got auth header (length=%d)", len(authHeader))

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			log.Printf("[AUTH] Invalid auth header format: parts=%d", len(parts))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization header format",
			})
		}

		tokenString := parts[1]
		log.Printf("[AUTH] Token extracted (length=%d)", len(tokenString))

		// Debug: decode raw JWT payload to see all claims
		jwtParts := strings.Split(tokenString, ".")
		if len(jwtParts) == 3 {
			payload, err := base64.RawURLEncoding.DecodeString(jwtParts[1])
			if err == nil {
				var rawClaims map[string]interface{}
				if json.Unmarshal(payload, &rawClaims) == nil {
					log.Printf("[AUTH] Raw JWT claims: %v", rawClaims)
				}
			}
		}

		// Parse and validate token
		claims, err := verifyToken(tokenString, cfg)
		if err != nil {
			log.Printf("[AUTH] Token verification failed: %v", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fmt.Sprintf("invalid token: %v", err),
			})
		}

		log.Printf("[AUTH] Token verified - subject=%s, email=%s, name=%s, emailVerified=%v, orgID=%s, role=%s",
			claims.Subject, claims.Email, claims.Name, claims.EmailVerified, claims.OrgID, claims.Role)

		// Get email/name from headers if not in token (WorkOS access tokens don't include these)
		email := claims.Email
		name := claims.Name
		if email == "" {
			email = c.Get("X-User-Email")
		}
		if name == "" {
			name = c.Get("X-User-Name")
		}
		log.Printf("[AUTH] Final user info: subject=%s, email=%s, name=%s", claims.Subject, email, name)

		// Store user info in context
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

	// First, parse without validation to inspect the issuer
	_, _, err := new(jwt.Parser).ParseUnverified(tokenString, claims)
	if err == nil {
		iss, _ := claims.GetIssuer()
		log.Printf("[AUTH] Token issuer: %q (expected: %q)", iss, expectedIssuer)
	}

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
