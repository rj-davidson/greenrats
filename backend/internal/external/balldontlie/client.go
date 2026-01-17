package balldontlie

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// Client is the BallDontLie API client for PGA golf data.
// Requires ALL-STAR tier ($9.99/mo) for tournament_results endpoint.
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

// Player represents a golfer from the BallDontLie PGA API.
type Player struct {
	ID          int    `json:"id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DisplayName string `json:"display_name"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
	OWGR        int    `json:"owgr"`
	Active      bool   `json:"active"`
}

// Tournament represents a tournament from the BallDontLie PGA API.
type Tournament struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Course    string `json:"course"`
	Location  string `json:"location"`
	Purse     string `json:"purse"` // Total prize money as string (API returns string)
	Season    int    `json:"season"`
}

// TournamentResult represents a golfer's result in a tournament.
// Available from the ALL-STAR tier endpoint.
// Note: API returns many numeric fields as strings, so we use string types and parse in ingest.
type TournamentResult struct {
	ID           int     `json:"id"`
	TournamentID int     `json:"tournament_id"`
	PlayerID     int     `json:"player_id"`
	Player       *Player `json:"player,omitempty"`
	Position     string  `json:"position"`
	Score        string  `json:"score"` // Score relative to par
	TotalStrokes string  `json:"total_strokes"`
	Earnings     string  `json:"earnings"` // Prize money in dollars
	Status       string  `json:"status"`   // "active", "cut", "withdrawn", "finished"
}

// Meta contains pagination metadata from the API.
type Meta struct {
	NextCursor int `json:"next_cursor,omitempty"`
	PerPage    int `json:"per_page"`
}

// PlayersResponse is the API response for players endpoint.
type PlayersResponse struct {
	Data []Player `json:"data"`
	Meta Meta     `json:"meta"`
}

// TournamentsResponse is the API response for tournaments endpoint.
type TournamentsResponse struct {
	Data []Tournament `json:"data"`
	Meta Meta         `json:"meta"`
}

// TournamentResultsResponse is the API response for tournament results endpoint.
type TournamentResultsResponse struct {
	Data []TournamentResult `json:"data"`
	Meta Meta               `json:"meta"`
}

// GetPlayers fetches all PGA players.
// FREE tier endpoint: GET /pga/v1/players
func (c *Client) GetPlayers(ctx context.Context) ([]Player, error) {
	var allPlayers []Player
	cursor := 0

	for {
		var response PlayersResponse

		req := c.client.R().
			SetContext(ctx).
			SetResult(&response).
			SetQueryParam("per_page", "100")

		if cursor > 0 {
			req.SetQueryParam("cursor", fmt.Sprintf("%d", cursor))
		}

		resp, err := req.Get("/pga/v1/players")
		if err != nil {
			return nil, fmt.Errorf("failed to fetch players: %w", err)
		}

		if resp.IsError() {
			return nil, fmt.Errorf("API error fetching players: %s", resp.Status())
		}

		allPlayers = append(allPlayers, response.Data...)

		if response.Meta.NextCursor == 0 {
			break
		}
		cursor = response.Meta.NextCursor
	}

	return allPlayers, nil
}

// GetTournaments fetches tournaments for a given season.
// FREE tier endpoint: GET /pga/v1/tournaments
func (c *Client) GetTournaments(ctx context.Context, season int) ([]Tournament, error) {
	var allTournaments []Tournament
	cursor := 0

	for {
		var response TournamentsResponse

		req := c.client.R().
			SetContext(ctx).
			SetResult(&response).
			SetQueryParam("season", fmt.Sprintf("%d", season)).
			SetQueryParam("per_page", "100")

		if cursor > 0 {
			req.SetQueryParam("cursor", fmt.Sprintf("%d", cursor))
		}

		resp, err := req.Get("/pga/v1/tournaments")
		if err != nil {
			return nil, fmt.Errorf("failed to fetch tournaments: %w", err)
		}

		if resp.IsError() {
			return nil, fmt.Errorf("API error fetching tournaments: %s", resp.Status())
		}

		allTournaments = append(allTournaments, response.Data...)

		if response.Meta.NextCursor == 0 {
			break
		}
		cursor = response.Meta.NextCursor
	}

	return allTournaments, nil
}

// GetTournamentResults fetches results/leaderboard for a tournament.
// ALL-STAR tier endpoint ($9.99/mo): GET /pga/v1/tournament_results
// Returns positions, scores, and earnings for all players in the tournament.
func (c *Client) GetTournamentResults(ctx context.Context, tournamentID int) ([]TournamentResult, error) {
	var allResults []TournamentResult
	cursor := 0

	for {
		var response TournamentResultsResponse

		req := c.client.R().
			SetContext(ctx).
			SetResult(&response).
			SetQueryParam("tournament_id", fmt.Sprintf("%d", tournamentID)).
			SetQueryParam("per_page", "100")

		if cursor > 0 {
			req.SetQueryParam("cursor", fmt.Sprintf("%d", cursor))
		}

		resp, err := req.Get("/pga/v1/tournament_results")
		if err != nil {
			return nil, fmt.Errorf("failed to fetch tournament results: %w", err)
		}

		if resp.IsError() {
			return nil, fmt.Errorf("API error fetching tournament results: %s", resp.Status())
		}

		allResults = append(allResults, response.Data...)

		if response.Meta.NextCursor == 0 {
			break
		}
		cursor = response.Meta.NextCursor
	}

	return allResults, nil
}
