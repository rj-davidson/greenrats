package livegolfdata

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// Client is the Live Golf Data API client (via RapidAPI).
// FREE tier: 250 requests/month (hard limit), 60 requests/minute.
// Used primarily for tournament field data before tournaments start.
type Client struct {
	client  *resty.Client
	apiKey  string
	baseURL string
}

// New creates a new Live Golf Data API client configured for RapidAPI.
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

// Tournament represents a tournament from the API.
type Tournament struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Status    string `json:"status"`
	Season    int    `json:"season"`
	Course    string `json:"course,omitempty"`
	Location  string `json:"location,omitempty"`
	Purse     int    `json:"purse,omitempty"`
}

// Golfer represents a golfer from the API.
type Golfer struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	FirstName    string `json:"first_name,omitempty"`
	LastName     string `json:"last_name,omitempty"`
	Country      string `json:"country"`
	WorldRanking int    `json:"world_ranking"`
	ImageURL     string `json:"image_url,omitempty"`
}

// LeaderboardEntry represents a golfer's position on the live leaderboard.
type LeaderboardEntry struct {
	GolferID     string `json:"golfer_id"`
	GolferName   string `json:"golfer_name"`
	Position     int    `json:"position"`
	Score        int    `json:"score"`         // Score relative to par
	TotalStrokes int    `json:"total_strokes"` // Total strokes taken
	Thru         int    `json:"thru"`          // Holes completed in current round
	Round        int    `json:"round"`         // Current round (1-4)
	Status       string `json:"status"`        // "active", "cut", "withdrawn"
}

// GetTournaments fetches tournaments for a given season.
// Endpoint: GET /tournaments?season={year}
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

// GetTournamentField fetches golfers in a tournament field.
// This endpoint is critical for getting the player field BEFORE a tournament starts.
// Endpoint: GET /tournament-field/{tournament_id}
// Note: This uses ~1 request from your 250/month quota.
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

// GetLiveLeaderboard fetches the live leaderboard for an active tournament.
// Endpoint: GET /live-leaderboard/{tournament_id}
// Note: Uses requests from your 250/month quota. Consider using BallDontLie for frequent polling.
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

// GetGolfer fetches a single golfer by ID.
// Endpoint: GET /golfers/{golfer_id}
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
