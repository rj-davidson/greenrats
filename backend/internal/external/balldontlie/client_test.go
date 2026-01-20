package balldontlie

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestNewClientCreatesRateLimiter(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := New("test-api-key", "https://api.example.com", logger)

	assert.NotNil(t, client.limiter)
	assert.Equal(t, rate.Limit(APIRateLimitPerSecond), client.limiter.Limit())
	assert.Equal(t, APIRateBurst, client.limiter.Burst())
}

func TestRateLimiterThrottlesRequests(t *testing.T) {
	limiter := rate.NewLimiter(rate.Limit(10), 1)

	start := time.Now()

	for i := 0; i < 3; i++ {
		_ = limiter.Wait(t.Context())
	}

	elapsed := time.Since(start)

	assert.GreaterOrEqual(t, elapsed, 150*time.Millisecond, "expected rate limiting to cause delay")
}

func TestConfigConstants(t *testing.T) {
	assert.Equal(t, 2.0, APIRateLimitPerSecond)
	assert.Equal(t, 5, APIRateBurst)
}
