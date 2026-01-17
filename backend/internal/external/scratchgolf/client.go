package scratchgolf

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// Client is the Scratch Golf API client.
type Client struct {
	client  *resty.Client
	apiKey  string
	baseURL string
}

// New creates a new Scratch Golf API client.
func New(apiKey, baseURL string) *Client {
	client := resty.New().
		SetBaseURL(baseURL).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", apiKey)).
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
}

// Golfer represents a golfer from the API.
type Golfer struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Country      string `json:"country"`
	WorldRanking int    `json:"world_ranking"`
	ImageURL     string `json:"image_url"`
}

// GetTournaments fetches tournaments for a given season.
func (c *Client) GetTournaments(ctx context.Context, season int) ([]Tournament, error) {
	var result struct {
		Tournaments []Tournament `json:"tournaments"`
	}

	_, err := c.client.R().
		SetContext(ctx).
		SetQueryParam("season", fmt.Sprintf("%d", season)).
		SetResult(&result).
		Get("/tournaments")

	if err != nil {
		return nil, fmt.Errorf("failed to fetch tournaments: %w", err)
	}

	return result.Tournaments, nil
}

// GetTournamentField fetches golfers in a tournament field.
func (c *Client) GetTournamentField(ctx context.Context, tournamentID string) ([]Golfer, error) {
	var result struct {
		Golfers []Golfer `json:"golfers"`
	}

	_, err := c.client.R().
		SetContext(ctx).
		SetResult(&result).
		Get(fmt.Sprintf("/tournaments/%s/field", tournamentID))

	if err != nil {
		return nil, fmt.Errorf("failed to fetch tournament field: %w", err)
	}

	return result.Golfers, nil
}

// GetGolfer fetches a single golfer by ID.
func (c *Client) GetGolfer(ctx context.Context, golferID string) (*Golfer, error) {
	var golfer Golfer

	_, err := c.client.R().
		SetContext(ctx).
		SetResult(&golfer).
		Get(fmt.Sprintf("/golfers/%s", golferID))

	if err != nil {
		return nil, fmt.Errorf("failed to fetch golfer: %w", err)
	}

	return &golfer, nil
}
