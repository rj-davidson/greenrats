package config

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Port int    `mapstructure:"PORT"`
	Env  string `mapstructure:"ENV"`

	DatabaseURL string `mapstructure:"DATABASE_URL"`

	WorkOSAPIKey   string `mapstructure:"WORKOS_API_KEY"`
	WorkOSClientID string `mapstructure:"WORKOS_CLIENT_ID"`

	BallDontLieAPIKey  string `mapstructure:"BALL_DONT_LIE_API_KEY"`
	BallDontLieBaseURL string `mapstructure:"BALL_DONT_LIE_BASE_URL"`
	PGATourAPIKey      string `mapstructure:"PGA_TOUR_API_KEY"`
	GoogleMapsAPIKey   string `mapstructure:"GOOGLE_MAPS_API_KEY"`

	SentryDSN string `mapstructure:"SENTRY_DSN"`

	ResendAPIKey string `mapstructure:"RESEND_API_KEY"`
	FromEmail    string `mapstructure:"FROM_EMAIL"`
	SendEmails   bool   `mapstructure:"SEND_EMAILS"`

	CurrentSeason int `mapstructure:"CURRENT_SEASON"`

	AdminEmails []string `mapstructure:"-"`
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetDefault("PORT", 8000)
	v.SetDefault("ENV", "development")
	v.SetDefault("DATABASE_URL", "")
	v.SetDefault("WORKOS_API_KEY", "")
	v.SetDefault("WORKOS_CLIENT_ID", "")
	v.SetDefault("BALL_DONT_LIE_API_KEY", "")
	v.SetDefault("BALL_DONT_LIE_BASE_URL", "https://api.balldontlie.io")
	v.SetDefault("PGA_TOUR_API_KEY", "")
	v.SetDefault("GOOGLE_MAPS_API_KEY", "")
	v.SetDefault("SENTRY_DSN", "")
	v.SetDefault("RESEND_API_KEY", "")
	v.SetDefault("FROM_EMAIL", "noreply@greenrats.com")
	v.SetDefault("SEND_EMAILS", false)
	v.SetDefault("CURRENT_SEASON", 2026)

	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")
	v.AddConfigPath("..")

	_ = v.ReadInConfig()

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	_ = v.BindEnv("PORT")
	_ = v.BindEnv("ENV")
	_ = v.BindEnv("DATABASE_URL")
	_ = v.BindEnv("WORKOS_API_KEY")
	_ = v.BindEnv("WORKOS_CLIENT_ID")
	_ = v.BindEnv("BALL_DONT_LIE_API_KEY")
	_ = v.BindEnv("BALL_DONT_LIE_BASE_URL")
	_ = v.BindEnv("PGA_TOUR_API_KEY")
	_ = v.BindEnv("GOOGLE_MAPS_API_KEY")
	_ = v.BindEnv("SENTRY_DSN")
	_ = v.BindEnv("RESEND_API_KEY")
	_ = v.BindEnv("FROM_EMAIL")
	_ = v.BindEnv("SEND_EMAILS")
	_ = v.BindEnv("CURRENT_SEASON")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	adminEmailsRaw := v.GetString("ADMIN_EMAILS")
	if adminEmailsRaw != "" {
		for _, email := range strings.Split(adminEmailsRaw, ",") {
			trimmed := strings.TrimSpace(strings.ToLower(email))
			if trimmed != "" {
				cfg.AdminEmails = append(cfg.AdminEmails, trimmed)
			}
		}
	}

	return &cfg, nil
}

func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

func (c *Config) IsAdminEmail(email string) bool {
	normalized := strings.TrimSpace(strings.ToLower(email))
	return slices.Contains(c.AdminEmails, normalized)
}
