package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	// Server
	Port int    `mapstructure:"PORT"`
	Env  string `mapstructure:"ENV"`

	// Database
	DatabaseURL string `mapstructure:"DATABASE_URL"`

	// WorkOS Authentication
	WorkOSAPIKey   string `mapstructure:"WORKOS_API_KEY"`
	WorkOSClientID string `mapstructure:"WORKOS_CLIENT_ID"`

	// External APIs
	ScratchGolfAPIKey  string `mapstructure:"SCRATCH_GOLF_API_KEY"`
	ScratchGolfBaseURL string `mapstructure:"SCRATCH_GOLF_BASE_URL"`
	BallDontLieAPIKey  string `mapstructure:"BALL_DONT_LIE_API_KEY"`
	BallDontLieBaseURL string `mapstructure:"BALL_DONT_LIE_BASE_URL"`
}

// Load reads configuration from environment variables and .env file.
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("PORT", 8000)
	v.SetDefault("ENV", "development")
	v.SetDefault("DATABASE_URL", "")
	v.SetDefault("WORKOS_API_KEY", "")
	v.SetDefault("WORKOS_CLIENT_ID", "")
	v.SetDefault("SCRATCH_GOLF_API_KEY", "")
	v.SetDefault("SCRATCH_GOLF_BASE_URL", "https://api.scratchgolf.com")
	v.SetDefault("BALL_DONT_LIE_API_KEY", "")
	v.SetDefault("BALL_DONT_LIE_BASE_URL", "https://api.balldontlie.io")

	// Read from .env file if it exists
	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")
	v.AddConfigPath("..")

	// Ignore error if .env file doesn't exist
	_ = v.ReadInConfig()

	// Read from environment variables (must bind keys for AutomaticEnv to work)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind environment variables explicitly
	_ = v.BindEnv("PORT")
	_ = v.BindEnv("ENV")
	_ = v.BindEnv("DATABASE_URL")
	_ = v.BindEnv("WORKOS_API_KEY")
	_ = v.BindEnv("WORKOS_CLIENT_ID")
	_ = v.BindEnv("SCRATCH_GOLF_API_KEY")
	_ = v.BindEnv("SCRATCH_GOLF_BASE_URL")
	_ = v.BindEnv("BALL_DONT_LIE_API_KEY")
	_ = v.BindEnv("BALL_DONT_LIE_BASE_URL")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// IsDevelopment returns true if the environment is development.
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

// IsProduction returns true if the environment is production.
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}
