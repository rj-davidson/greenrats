package livegolfdata

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	client  *resty.Client
	apiKey  string
	baseURL string
}

func New(apiKey, baseURL string) *Client {
	client := resty.New().
		SetBaseURL(baseURL).
		SetHeader("X-RapidAPI-Key", apiKey).
		SetHeader("Content-Type", "application/json")

	return &Client{
		client:  client,
		apiKey:  apiKey,
		baseURL: baseURL,
	}
}

func (c *Client) GetTournaments(ctx context.Context, season int) ([]Tournament, error) {
	var result struct {
		Tournaments []Tournament `json:"tournaments"`
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetQueryParam("season", fmt.Sprintf("%d", season)).
		SetResult(&result).
		Get("/tournaments")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tournaments: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error fetching tournaments: %s", resp.Status())
	}

	return result.Tournaments, nil
}

func (c *Client) GetTournamentField(ctx context.Context, tournamentID string) ([]Golfer, error) {
	var result struct {
		Golfers []Golfer `json:"golfers"`
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetResult(&result).
		Get(fmt.Sprintf("/tournament-field/%s", tournamentID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tournament field: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error fetching tournament field: %s", resp.Status())
	}

	return result.Golfers, nil
}

func (c *Client) GetLiveLeaderboard(ctx context.Context, tournamentID string) ([]LeaderboardEntry, error) {
	var result struct {
		Leaderboard []LeaderboardEntry `json:"leaderboard"`
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetResult(&result).
		Get(fmt.Sprintf("/live-leaderboard/%s", tournamentID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch live leaderboard: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error fetching live leaderboard: %s", resp.Status())
	}

	return result.Leaderboard, nil
}

func (c *Client) GetGolfer(ctx context.Context, golferID string) (*Golfer, error) {
	var golfer Golfer

	resp, err := c.client.R().
		SetContext(ctx).
		SetResult(&golfer).
		Get(fmt.Sprintf("/golfers/%s", golferID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch golfer: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error fetching golfer: %s", resp.Status())
	}

	return &golfer, nil
}

func (c *Client) GetEarnings(ctx context.Context, tournamentID string, year int) ([]EarningsEntry, error) {
	var result struct {
		Leaderboard []EarningsEntry `json:"leaderboard"`
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetQueryParam("tournId", tournamentID).
		SetQueryParam("year", fmt.Sprintf("%d", year)).
		SetResult(&result).
		Get("/earnings")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch earnings: %w", err)
	}

	if resp.IsError() {
		if resp.StatusCode() == 400 {
			return nil, nil
		}
		return nil, fmt.Errorf("API error fetching earnings: %s", resp.Status())
	}

	return result.Leaderboard, nil
}

func (c *Client) GetSchedule(ctx context.Context, year int) ([]Tournament, error) {
	resp, err := c.client.R().
		SetContext(ctx).
		SetQueryParam("orgId", "1").
		SetQueryParam("year", fmt.Sprintf("%d", year)).
		Get("/schedule")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schedule: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error fetching schedule: %s", resp.Status())
	}

	var tournaments []Tournament
	if err := parseScheduleResponse(resp.Body(), &tournaments); err != nil {
		return nil, err
	}

	return tournaments, nil
}

func parseScheduleResponse(data []byte, tournaments *[]Tournament) error {
	var payload any
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("failed to parse schedule: %w", err)
	}

	collectTournamentsFromPayload(payload, tournaments)
	return nil
}

func collectTournamentsFromPayload(value any, tournaments *[]Tournament) {
	switch v := value.(type) {
	case map[string]any:
		if t, ok := extractTournamentFromMap(v); ok {
			*tournaments = append(*tournaments, t)
		}
		for _, child := range v {
			collectTournamentsFromPayload(child, tournaments)
		}
	case []any:
		for _, child := range v {
			collectTournamentsFromPayload(child, tournaments)
		}
	}
}

func extractTournamentFromMap(m map[string]any) (Tournament, bool) {
	name := stringFromMap(m, "name", "tournamentName", "tournament_name")
	if name == "" {
		return Tournament{}, false
	}

	id := stringFromMap(m, "tournId", "tourn_id", "tournamentId", "tournament_id", "id")
	if id == "" {
		return Tournament{}, false
	}

	return Tournament{
		ID:   id,
		Name: name,
	}, true
}

func stringFromMap(m map[string]any, keys ...string) string {
	for _, key := range keys {
		if val, ok := m[key]; ok {
			switch v := val.(type) {
			case string:
				return v
			case float64:
				if v == float64(int64(v)) {
					return fmt.Sprintf("%d", int64(v))
				}
				return fmt.Sprintf("%v", v)
			}
		}
	}
	return ""
}
