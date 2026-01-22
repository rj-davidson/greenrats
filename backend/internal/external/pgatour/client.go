package pgatour

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-resty/resty/v2"
	"golang.org/x/time/rate"
)

const fieldQuery = `
query Field($id: ID!) {
  field(id: $id) {
    tournamentName
    id
    players {
      id
      firstName
      lastName
      displayName
      country
      countryFlag
      amateur
    }
  }
}
`

const scheduleQuery = `
query Schedule($tourCode: String!, $year: String!) {
  schedule(tourCode: $tourCode, year: $year) {
    completed {
      month
      year
      tournaments {
        id
        tournamentName
        startDate
      }
    }
    upcoming {
      month
      year
      tournaments {
        id
        tournamentName
        startDate
      }
    }
  }
}
`

type Client struct {
	client  *resty.Client
	limiter *rate.Limiter
	logger  *slog.Logger
}

func New(apiKey string, logger *slog.Logger) *Client {
	client := resty.New().
		SetBaseURL(DefaultGraphQLURL).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json")

	if apiKey != "" {
		client.SetHeader("x-api-key", apiKey)
	}

	limiter := rate.NewLimiter(rate.Limit(APIRateLimitPerSecond), APIRateBurst)

	return &Client{
		client:  client,
		limiter: limiter,
		logger:  logger,
	}
}

func (c *Client) wait(ctx context.Context) error {
	return c.limiter.Wait(ctx)
}

func (c *Client) GetTournamentField(ctx context.Context, tournamentID string) ([]FieldEntry, error) {
	c.logger.Info("fetching tournament field from PGA Tour", "tournament_id", tournamentID)

	if err := c.wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter: %w", err)
	}

	req := GraphQLRequest{
		Query:         fieldQuery,
		OperationName: "Field",
		Variables: map[string]any{
			"id": tournamentID,
		},
	}

	var response GraphQLResponse

	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&response).
		Post("")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tournament field: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error fetching field: %s", resp.Status())
	}

	if len(response.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL error: %s", response.Errors[0].Message)
	}

	if response.Data == nil || response.Data.Field == nil {
		c.logger.Debug("no field data returned", "tournament_id", tournamentID)
		return nil, nil
	}

	players := response.Data.Field.Players
	entries := make([]FieldEntry, len(players))

	for i, p := range players {
		entries[i] = FieldEntry{
			PGATourID:   p.ID,
			FirstName:   p.FirstName,
			LastName:    p.LastName,
			DisplayName: p.DisplayName,
			CountryCode: p.CountryCode,
			IsAmateur:   p.IsAmateur,
		}
	}

	c.logger.Info("tournament field fetch complete", "tournament_id", tournamentID, "player_count", len(entries))
	return entries, nil
}

func (c *Client) GetSchedule(ctx context.Context, year string) ([]ScheduleTournament, error) {
	c.logger.Info("fetching PGA Tour schedule", "year", year)

	if err := c.wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter: %w", err)
	}

	req := GraphQLRequest{
		Query:         scheduleQuery,
		OperationName: "Schedule",
		Variables: map[string]any{
			"tourCode": "R",
			"year":     year,
		},
	}

	var response GraphQLResponse

	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&response).
		Post("")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schedule: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error fetching schedule: %s", resp.Status())
	}

	if len(response.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL error: %s", response.Errors[0].Message)
	}

	if response.Data == nil || response.Data.Schedule == nil {
		c.logger.Debug("no schedule data returned", "year", year)
		return nil, nil
	}

	var all []ScheduleTournament
	for _, month := range response.Data.Schedule.Completed {
		all = append(all, month.Tournaments...)
	}
	for _, month := range response.Data.Schedule.Upcoming {
		all = append(all, month.Tournaments...)
	}

	c.logger.Info("schedule fetch complete", "year", year, "total_tournaments", len(all))
	return all, nil
}
