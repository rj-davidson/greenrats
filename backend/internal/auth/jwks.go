package auth

import (
	"context"
	"fmt"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

type JWKSProvider struct {
	jwks     keyfunc.Keyfunc
	clientID string
	cancel   context.CancelFunc
}

func NewJWKSProvider(clientID string) (*JWKSProvider, error) {
	jwksURL := fmt.Sprintf("https://api.workos.com/sso/jwks/%s", clientID)

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

func (p *JWKSProvider) Keyfunc(token *jwt.Token) (interface{}, error) {
	return p.jwks.Keyfunc(token)
}

func (p *JWKSProvider) Close() {
	if p.cancel != nil {
		p.cancel()
	}
}
