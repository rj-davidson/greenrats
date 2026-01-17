package balldontlie

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// Client is the BallDontLie API client for golf statistics.
type Client struct {
	client  *resty.Client
	apiKey  string
	baseURL string
}

// New creates a new BallDontLie API client.
func New(apiKey, baseURL string) *Client {
	client := resty.New().
		SetBaseURL(baseURL).
		SetHeader("Authorization", apiKey).
		SetHeader("Content-Type", "application/json")

	return &Client{
		client:  client,
		apiKey:  apiKey,
		baseURL: baseURL,
	}
}

// GolferStats represents golfer statistics from the API.
type GolferStats struct {
	GolferID     string  `json:"golfer_id"`
	Earnings     float64 `json:"earnings"`
	Wins         int     `json:"wins"`
	TopTens      int     `json:"top_tens"`
	CutsMade     int     `json:"cuts_made"`
	StrokesAvg   float64 `json:"strokes_avg"`
	Season       int     `json:"season"`
}

// TournamentResult represents a golfer's result in a tournament.
type TournamentResult struct {
	GolferID     string  `json:"golfer_id"`
	TournamentID string  `json:"tournament_id"`
	Position     int     `json:"position"`
	Score        int     `json:"score"`
	Earnings     float64 `json:"earnings"`
}

// GetGolferStats fetches statistics for a golfer.
func (c *Client) GetGolferStats(ctx context.Context, golferID string, season int) (*GolferStats, error) {
	var stats GolferStats

	_, err := c.client.R().
		SetContext(ctx).
		SetQueryParam("season", fmt.Sprintf("%d", season)).
		SetResult(&stats).
		Get(fmt.Sprintf("/golfers/%s/stats", golferID))

	if err != nil {
		return nil, fmt.Errorf("failed to fetch golfer stats: %w", err)
	}

	return &stats, nil
}

// GetTournamentResults fetches results for a tournament.
func (c *Client) GetTournamentResults(ctx context.Context, tournamentID string) ([]TournamentResult, error) {
	var result struct {
		Results []TournamentResult `json:"results"`
	}

	_, err := c.client.R().
		SetContext(ctx).
		SetResult(&result).
		Get(fmt.Sprintf("/tournaments/%s/results", tournamentID))

	if err != nil {
		return nil, fmt.Errorf("failed to fetch tournament results: %w", err)
	}

	return result.Results, nil
}

// GetLiveLeaderboard fetches live leaderboard for an active tournament.
func (c *Client) GetLiveLeaderboard(ctx context.Context, tournamentID string) ([]TournamentResult, error) {
	var result struct {
		Leaderboard []TournamentResult `json:"leaderboard"`
	}

	_, err := c.client.R().
		SetContext(ctx).
		SetResult(&result).
		Get(fmt.Sprintf("/tournaments/%s/leaderboard", tournamentID))

	if err != nil {
		return nil, fmt.Errorf("failed to fetch live leaderboard: %w", err)
	}

	return result.Leaderboard, nil
}
