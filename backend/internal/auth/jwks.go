package auth

import (
	"context"
	"fmt"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

// JWKSProvider manages JWKS fetching and caching for JWT verification.
type JWKSProvider struct {
	jwks     keyfunc.Keyfunc
	clientID string
	cancel   context.CancelFunc
}

// NewJWKSProvider creates a new JWKS provider that fetches keys from WorkOS.
// The provider handles background refresh automatically.
func NewJWKSProvider(clientID string) (*JWKSProvider, error) {
	jwksURL := fmt.Sprintf("https://api.workos.com/sso/jwks/%s", clientID)

	// Create a cancellable context for the JWKS background refresh
	ctx, cancel := context.WithCancel(context.Background())

	jwks, err := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create JWKS keyfunc: %w", err)
	}

	return &JWKSProvider{
		jwks:     jwks,
		clientID: clientID,
		cancel:   cancel,
	}, nil
}

// Keyfunc returns a jwt.Keyfunc for use with jwt.ParseWithClaims.
func (p *JWKSProvider) Keyfunc(token *jwt.Token) (interface{}, error) {
	return p.jwks.Keyfunc(token)
}

// Close stops the background JWKS refresh goroutine.
func (p *JWKSProvider) Close() {
	if p.cancel != nil {
		p.cancel()
	}
}
