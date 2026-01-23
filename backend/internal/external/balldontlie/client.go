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

func (c *Client) GetCourses(ctx context.Context) ([]Course, error) {
	c.logger.Info("fetching courses from BallDontLie")

	var allCourses []Course
	cursor := 0

	for {
		if err := c.wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}

		var response CoursesResponse

		req := c.client.R().
			SetContext(ctx).
			SetResult(&response).
			SetQueryParam("per_page", "100")

		if cursor > 0 {
			req.SetQueryParam("cursor", fmt.Sprintf("%d", cursor))
		}

		resp, err := req.Get("/pga/v1/courses")
		if err != nil {
			return nil, fmt.Errorf("failed to fetch courses: %w", err)
		}

		if resp.IsError() {
			return nil, fmt.Errorf("API error fetching courses: %s", resp.Status())
		}

		allCourses = append(allCourses, response.Data...)
		c.logger.Debug("fetched courses page", "cursor", cursor, "count", len(response.Data))

		if response.Meta.NextCursor == 0 {
			break
		}
		cursor = response.Meta.NextCursor
	}

	c.logger.Info("courses fetch complete", "total", len(allCourses))
	return allCourses, nil
}

func (c *Client) GetCourseHoles(ctx context.Context, courseID int) ([]CourseHole, error) {
	c.logger.Info("fetching course holes", "course_id", courseID)

	var allHoles []CourseHole
	cursor := 0

	for {
		if err := c.wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}

		var response CourseHolesResponse

		req := c.client.R().
			SetContext(ctx).
			SetResult(&response).
			SetQueryParam("course_ids[]", fmt.Sprintf("%d", courseID)).
			SetQueryParam("per_page", "100")

		if cursor > 0 {
			req.SetQueryParam("cursor", fmt.Sprintf("%d", cursor))
		}

		resp, err := req.Get("/pga/v1/course_holes")
		if err != nil {
			return nil, fmt.Errorf("failed to fetch course holes: %w", err)
		}

		if resp.IsError() {
			return nil, fmt.Errorf("API error fetching course holes: %s", resp.Status())
		}

		allHoles = append(allHoles, response.Data...)
		c.logger.Debug("fetched holes page", "cursor", cursor, "count", len(response.Data))

		if response.Meta.NextCursor == 0 {
			break
		}
		cursor = response.Meta.NextCursor
	}

	c.logger.Info("course holes fetch complete", "total", len(allHoles))
	return allHoles, nil
}

func (c *Client) GetPlayerRoundResults(ctx context.Context, tournamentID int) ([]PlayerRoundResult, error) {
	c.logger.Info("fetching player round results", "tournament_id", tournamentID)

	var allResults []PlayerRoundResult
	cursor := 0

	for {
		if err := c.wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}

		var response PlayerRoundResultsResponse

		req := c.client.R().
			SetContext(ctx).
			SetResult(&response).
			SetQueryParam("tournament_ids[]", fmt.Sprintf("%d", tournamentID)).
			SetQueryParam("per_page", "100")

		if cursor > 0 {
			req.SetQueryParam("cursor", fmt.Sprintf("%d", cursor))
		}

		resp, err := req.Get("/pga/v1/player_round_results")
		if err != nil {
			return nil, fmt.Errorf("failed to fetch player round results: %w", err)
		}

		if resp.IsError() {
			return nil, fmt.Errorf("API error fetching player round results: %s", resp.Status())
		}

		allResults = append(allResults, response.Data...)
		c.logger.Debug("fetched round results page", "cursor", cursor, "count", len(response.Data))

		if response.Meta.NextCursor == 0 {
			break
		}
		cursor = response.Meta.NextCursor
	}

	c.logger.Info("player round results fetch complete", "total", len(allResults))
	return allResults, nil
}

func (c *Client) GetPlayerScorecards(ctx context.Context, tournamentID int) ([]PlayerScorecard, error) {
	c.logger.Info("fetching player scorecards", "tournament_id", tournamentID)

	var allScorecards []PlayerScorecard
	cursor := 0

	for {
		if err := c.wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}

		var response PlayerScorecardsResponse

		req := c.client.R().
			SetContext(ctx).
			SetResult(&response).
			SetQueryParam("tournament_ids[]", fmt.Sprintf("%d", tournamentID)).
			SetQueryParam("per_page", "100")

		if cursor > 0 {
			req.SetQueryParam("cursor", fmt.Sprintf("%d", cursor))
		}

		resp, err := req.Get("/pga/v1/player_scorecards")
		if err != nil {
			return nil, fmt.Errorf("failed to fetch player scorecards: %w", err)
		}

		if resp.IsError() {
			return nil, fmt.Errorf("API error fetching player scorecards: %s", resp.Status())
		}

		allScorecards = append(allScorecards, response.Data...)
		c.logger.Debug("fetched scorecards page", "cursor", cursor, "count", len(response.Data))

		if response.Meta.NextCursor == 0 {
			break
		}
		cursor = response.Meta.NextCursor
	}

	c.logger.Info("player scorecards fetch complete", "total", len(allScorecards))
	return allScorecards, nil
}

func (c *Client) GetPlayerSeasonStats(ctx context.Context, season int, statIDs []int) ([]PlayerSeasonStat, error) {
	c.logger.Info("fetching player season stats", "season", season, "stat_ids", statIDs)

	var allStats []PlayerSeasonStat
	cursor := 0

	for {
		if err := c.wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}

		var response PlayerSeasonStatsResponse

		req := c.client.R().
			SetContext(ctx).
			SetResult(&response).
			SetQueryParam("season", fmt.Sprintf("%d", season)).
			SetQueryParam("per_page", "100")

		for _, statID := range statIDs {
			req.SetQueryParam("stat_ids[]", fmt.Sprintf("%d", statID))
		}

		if cursor > 0 {
			req.SetQueryParam("cursor", fmt.Sprintf("%d", cursor))
		}

		resp, err := req.Get("/pga/v1/player_season_stats")
		if err != nil {
			return nil, fmt.Errorf("failed to fetch player season stats: %w", err)
		}

		if resp.IsError() {
			return nil, fmt.Errorf("API error fetching player season stats: %s", resp.Status())
		}

		allStats = append(allStats, response.Data...)
		c.logger.Debug("fetched season stats page", "cursor", cursor, "count", len(response.Data))

		if response.Meta.NextCursor == 0 {
			break
		}
		cursor = response.Meta.NextCursor
	}

	c.logger.Info("player season stats fetch complete", "total", len(allStats))
	return allStats, nil
}

func (c *Client) GetTournamentCourseStats(ctx context.Context, tournamentID int) ([]TournamentCourseStats, error) {
	c.logger.Info("fetching tournament course stats", "tournament_id", tournamentID)

	var allStats []TournamentCourseStats
	cursor := 0

	for {
		if err := c.wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}

		var response TournamentCourseStatsResponse

		req := c.client.R().
			SetContext(ctx).
			SetResult(&response).
			SetQueryParam("tournament_ids[]", fmt.Sprintf("%d", tournamentID)).
			SetQueryParam("per_page", "100")

		if cursor > 0 {
			req.SetQueryParam("cursor", fmt.Sprintf("%d", cursor))
		}

		resp, err := req.Get("/pga/v1/tournament_course_stats")
		if err != nil {
			return nil, fmt.Errorf("failed to fetch tournament course stats: %w", err)
		}

		if resp.IsError() {
			return nil, fmt.Errorf("API error fetching tournament course stats: %s", resp.Status())
		}

		allStats = append(allStats, response.Data...)
		c.logger.Debug("fetched course stats page", "cursor", cursor, "count", len(response.Data))

		if response.Meta.NextCursor == 0 {
			break
		}
		cursor = response.Meta.NextCursor
	}

	c.logger.Info("tournament course stats fetch complete", "total", len(allStats))
	return allStats, nil
}

func (c *Client) GetTournamentField(ctx context.Context, tournamentID int) ([]TournamentField, error) {
	c.logger.Info("fetching tournament field", "tournament_id", tournamentID)

	var allEntries []TournamentField
	cursor := 0

	for {
		if err := c.wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}

		var response TournamentFieldResponse

		req := c.client.R().
			SetContext(ctx).
			SetResult(&response).
			SetQueryParam("tournament_ids[]", fmt.Sprintf("%d", tournamentID)).
			SetQueryParam("per_page", "100")

		if cursor > 0 {
			req.SetQueryParam("cursor", fmt.Sprintf("%d", cursor))
		}

		resp, err := req.Get("/pga/v1/tournament_field")
		if err != nil {
			return nil, fmt.Errorf("failed to fetch tournament field: %w", err)
		}

		if resp.IsError() {
			return nil, fmt.Errorf("API error fetching tournament field: %s", resp.Status())
		}

		allEntries = append(allEntries, response.Data...)
		c.logger.Debug("fetched field page", "cursor", cursor, "count", len(response.Data))

		if response.Meta.NextCursor == 0 {
			break
		}
		cursor = response.Meta.NextCursor
	}

	c.logger.Info("tournament field fetch complete", "total", len(allEntries))
	return allEntries, nil
}

func (c *Client) GetPlayerRoundStats(ctx context.Context, tournamentID int) ([]PlayerRoundStats, error) {
	c.logger.Info("fetching player round stats", "tournament_id", tournamentID)

	var allStats []PlayerRoundStats
	cursor := 0

	for {
		if err := c.wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}

		var response PlayerRoundStatsResponse

		req := c.client.R().
			SetContext(ctx).
			SetResult(&response).
			SetQueryParam("tournament_ids[]", fmt.Sprintf("%d", tournamentID)).
			SetQueryParam("per_page", "100")

		if cursor > 0 {
			req.SetQueryParam("cursor", fmt.Sprintf("%d", cursor))
		}

		resp, err := req.Get("/pga/v1/player_round_stats")
		if err != nil {
			return nil, fmt.Errorf("failed to fetch player round stats: %w", err)
		}

		if resp.IsError() {
			return nil, fmt.Errorf("API error fetching player round stats: %s", resp.Status())
		}

		allStats = append(allStats, response.Data...)
		c.logger.Debug("fetched round stats page", "cursor", cursor, "count", len(response.Data))

		if response.Meta.NextCursor == 0 {
			break
		}
		cursor = response.Meta.NextCursor
	}

	c.logger.Info("player round stats fetch complete", "total", len(allStats))
	return allStats, nil
}
