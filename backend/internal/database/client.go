package database

import (
	"context"
	stdsql "database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"github.com/rj-davidson/greenrats/ent"
)

const (
	defaultConnectTimeoutSeconds = 5
	defaultPingTimeout           = 8 * time.Second
	maxOpenConns                 = 25
	maxIdleConns                 = 25
	connMaxIdleTime              = 5 * time.Minute
	connMaxLifetime              = 30 * time.Minute
)

func OpenClient(databaseURL string, logger *slog.Logger) (*ent.Client, error) {
	if strings.TrimSpace(databaseURL) == "" {
		return nil, fmt.Errorf("database url is empty")
	}

	dsn, timeoutApplied := withDefaultConnectTimeout(databaseURL, defaultConnectTimeoutSeconds)
	if timeoutApplied && logger != nil {
		logger.Info("database connect timeout not set, applying default",
			"connect_timeout_seconds", defaultConnectTimeoutSeconds,
		)
	}

	sqlDB, err := stdsql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	pingCtx, cancel := context.WithTimeout(context.Background(), defaultPingTimeout)
	defer cancel()

	if err := sqlDB.PingContext(pingCtx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("failed to ping database host %s: %w", databaseHost(databaseURL), err)
	}

	if logger != nil {
		logger.Info("database connectivity verified", "host", databaseHost(databaseURL))
	}

	drv := entsql.OpenDB(dialect.Postgres, sqlDB)
	return ent.NewClient(ent.Driver(drv)), nil
}

func withDefaultConnectTimeout(databaseURL string, timeoutSeconds int) (string, bool) {
	if timeoutSeconds <= 0 {
		return databaseURL, false
	}

	trimmed := strings.TrimSpace(databaseURL)
	if trimmed == "" {
		return databaseURL, false
	}

	if strings.Contains(trimmed, "://") {
		parsed, err := url.Parse(trimmed)
		if err != nil {
			return databaseURL, false
		}

		query := parsed.Query()
		if query.Get("connect_timeout") != "" {
			return trimmed, false
		}

		query.Set("connect_timeout", strconv.Itoa(timeoutSeconds))
		parsed.RawQuery = query.Encode()
		return parsed.String(), true
	}

	if strings.Contains(trimmed, "connect_timeout=") {
		return trimmed, false
	}

	return trimmed + " connect_timeout=" + strconv.Itoa(timeoutSeconds), true
}

func databaseHost(databaseURL string) string {
	trimmed := strings.TrimSpace(databaseURL)
	if trimmed == "" {
		return "unknown"
	}

	if strings.Contains(trimmed, "://") {
		parsed, err := url.Parse(trimmed)
		if err == nil && parsed.Host != "" {
			return parsed.Host
		}
	}

	for _, field := range strings.Fields(trimmed) {
		if strings.HasPrefix(field, "host=") {
			return strings.TrimPrefix(field, "host=")
		}
	}

	return "unknown"
}
