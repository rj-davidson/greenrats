package balldontlie

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-resty/resty/v2"
	"golang.org/x/time/rate"
)

type Client struct {
	client  *resty.Client
	limiter *rate.Limiter
	logger  *slog.Logger
}

func New(apiKey, baseURL string, logger *slog.Logger) *Client {
	client := resty.New().
		SetBaseURL(baseURL).
		SetHeader("Authorization", apiKey).
		SetHeader("Content-Type", "application/json")

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

func (c *Client) GetPlayers(ctx context.Context) ([]Player, error) {
	c.logger.Info("fetching players from BallDontLie")

	var allPlayers []Player
	cursor := 0

	for {
		if err := c.wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}

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
		c.logger.Debug("fetched player page", "cursor", cursor, "count", len(response.Data))

		if response.Meta.NextCursor == 0 {
			break
		}
		cursor = response.Meta.NextCursor
	}

	c.logger.Info("players fetch complete", "total", len(allPlayers))
	return allPlayers, nil
}

func (c *Client) GetTournaments(ctx context.Context, season int) ([]Tournament, error) {
	c.logger.Info("fetching tournaments", "season", season)

	var allTournaments []Tournament
	cursor := 0

	for {
		if err := c.wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}

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

		for _, tournament := range response.Data {
			if tournament.ID > 42 {
				continue
			}
			allTournaments = append(allTournaments, tournament)
		}
		c.logger.Debug("fetched tournament page", "cursor", cursor, "count", len(response.Data))

		if response.Meta.NextCursor == 0 {
			break
		}
		cursor = response.Meta.NextCursor
	}

	c.logger.Info("tournaments fetch complete", "total", len(allTournaments))
	return allTournaments, nil
}

func (c *Client) GetTournamentResults(ctx context.Context, tournamentID int) ([]TournamentResult, error) {
	c.logger.Info("fetching tournament results", "tournament_id", tournamentID)

	var allResults []TournamentResult
	cursor := 0

	for {
		if err := c.wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}

		var response TournamentResultsResponse

		req := c.client.R().
			SetContext(ctx).
			SetResult(&response).
			SetQueryParam("tournament_ids[]", fmt.Sprintf("%d", tournamentID)).
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
		c.logger.Debug("fetched results page", "cursor", cursor, "count", len(response.Data))

		if response.Meta.NextCursor == 0 {
			break
		}
		cursor = response.Meta.NextCursor
	}

	c.logger.Info("tournament results fetch complete", "total", len(allResults))
	return allResults, nil
}
