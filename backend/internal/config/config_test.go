package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear any existing env vars that might interfere
	os.Unsetenv("PORT")
	os.Unsetenv("ENV")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, 8000, cfg.Port)
	assert.Equal(t, "development", cfg.Env)
	assert.Equal(t, "https://api.balldontlie.io", cfg.BallDontLieBaseURL)
}

func TestLoad_FromEnv(t *testing.T) {
	t.Setenv("PORT", "9000")
	t.Setenv("ENV", "production")
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, 9000, cfg.Port)
	assert.Equal(t, "production", cfg.Env)
	assert.Equal(t, "postgres://test:test@localhost:5432/test", cfg.DatabaseURL)
}

func TestConfig_IsDevelopment(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want bool
	}{
		{"development", "development", true},
		{"production", "production", false},
		{"staging", "staging", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Env: tt.env}
			assert.Equal(t, tt.want, cfg.IsDevelopment())
		})
	}
}

func TestConfig_IsProduction(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want bool
	}{
		{"development", "development", false},
		{"production", "production", true},
		{"staging", "staging", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Env: tt.env}
			assert.Equal(t, tt.want, cfg.IsProduction())
		})
	}
}
