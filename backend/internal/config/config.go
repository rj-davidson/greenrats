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
	LiveGolfDataAPIKey  string `mapstructure:"LIVE_GOLF_DATA_API_KEY"`
	LiveGolfDataBaseURL string `mapstructure:"LIVE_GOLF_DATA_BASE_URL"`
	BallDontLieAPIKey   string `mapstructure:"BALL_DONT_LIE_API_KEY"`
	BallDontLieBaseURL  string `mapstructure:"BALL_DONT_LIE_BASE_URL"`
	PGATourAPIKey       string `mapstructure:"PGA_TOUR_API_KEY"`
	PGATourBaseURL      string `mapstructure:"PGA_TOUR_BASE_URL"`

	// Monitoring
	SentryDSN string `mapstructure:"SENTRY_DSN"`

	// Email
	ResendAPIKey string `mapstructure:"RESEND_API_KEY"`
	FromEmail    string `mapstructure:"FROM_EMAIL"`
	SendEmails   bool   `mapstructure:"SEND_EMAILS"`
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
	v.SetDefault("LIVE_GOLF_DATA_API_KEY", "")
	v.SetDefault("LIVE_GOLF_DATA_BASE_URL", "https://live-golf-data.p.rapidapi.com")
	v.SetDefault("BALL_DONT_LIE_API_KEY", "")
	v.SetDefault("BALL_DONT_LIE_BASE_URL", "https://api.balldontlie.io")
	v.SetDefault("PGA_TOUR_API_KEY", "")
	v.SetDefault("PGA_TOUR_BASE_URL", "https://orchestrator.pgatour.com/graphql")
	v.SetDefault("SENTRY_DSN", "")
	v.SetDefault("RESEND_API_KEY", "")
	v.SetDefault("FROM_EMAIL", "noreply@greenrats.com")
	v.SetDefault("SEND_EMAILS", false)

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
	_ = v.BindEnv("LIVE_GOLF_DATA_API_KEY")
	_ = v.BindEnv("LIVE_GOLF_DATA_BASE_URL")
	_ = v.BindEnv("BALL_DONT_LIE_API_KEY")
	_ = v.BindEnv("BALL_DONT_LIE_BASE_URL")
	_ = v.BindEnv("PGA_TOUR_API_KEY")
	_ = v.BindEnv("PGA_TOUR_BASE_URL")
	_ = v.BindEnv("SENTRY_DSN")
	_ = v.BindEnv("RESEND_API_KEY")
	_ = v.BindEnv("FROM_EMAIL")
	_ = v.BindEnv("SEND_EMAILS")

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
