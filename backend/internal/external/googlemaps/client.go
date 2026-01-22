package googlemaps

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"golang.org/x/time/rate"
	"googlemaps.github.io/maps"
)

const (
	requestsPerSecond = 2
	burstSize         = 2
)

type Client struct {
	client  *maps.Client
	limiter *rate.Limiter
	logger  *slog.Logger
}

func New(apiKey string, logger *slog.Logger) (*Client, error) {
	c, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Google Maps client: %w", err)
	}
	return &Client{
		client:  c,
		limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), burstSize),
		logger:  logger,
	}, nil
}

func (c *Client) GetTimezone(ctx context.Context, city, state, country string, timestamp time.Time) (string, error) {
	address := formatAddress(city, state, country)
	c.logger.Info("looking up timezone", "address", address)

	if err := c.limiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limiter wait failed: %w", err)
	}

	geocodeReq := &maps.GeocodingRequest{
		Address: address,
	}
	geocodeResp, err := c.client.Geocode(ctx, geocodeReq)
	if err != nil {
		return "", fmt.Errorf("geocoding failed for %q: %w", address, err)
	}
	if len(geocodeResp) == 0 {
		return "", fmt.Errorf("no geocoding results for %q", address)
	}

	loc := geocodeResp[0].Geometry.Location
	c.logger.Debug("geocoded address", "address", address, "lat", loc.Lat, "lng", loc.Lng)

	if err := c.limiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limiter wait failed: %w", err)
	}

	tzReq := &maps.TimezoneRequest{
		Location:  &maps.LatLng{Lat: loc.Lat, Lng: loc.Lng},
		Timestamp: timestamp,
	}
	tzResp, err := c.client.Timezone(ctx, tzReq)
	if err != nil {
		return "", fmt.Errorf("timezone lookup failed: %w", err)
	}

	c.logger.Info("timezone lookup complete", "address", address, "timezone", tzResp.TimeZoneID)
	return tzResp.TimeZoneID, nil
}

func formatAddress(city, state, country string) string {
	var parts []string
	if city != "" {
		parts = append(parts, city)
	}
	if state != "" {
		parts = append(parts, state)
	}
	if country != "" {
		parts = append(parts, country)
	}
	return strings.Join(parts, ", ")
}
