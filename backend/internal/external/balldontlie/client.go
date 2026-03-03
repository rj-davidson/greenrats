package balldontlie

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

type Client struct {
	client  *resty.Client
	limiter *rate.Limiter
	breaker *gobreaker.CircuitBreaker
	logger  *slog.Logger
}

func New(apiKey, baseURL string, isDevelopment bool, logger *slog.Logger) *Client {
	client := resty.New().
		SetBaseURL(baseURL).
		SetHeader("Authorization", apiKey).
		SetHeader("Content-Type", "application/json")

	rateLimit := APIRateLimitPerSecond
	rateBurst := APIRateBurst
	if isDevelopment {
		rateLimit = APIRateLimitPerSecondDev
		rateBurst = APIRateBurstDev
	}
	limiter := rate.NewLimiter(rate.Limit(rateLimit), rateBurst)

	breaker := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "balldontlie",
		MaxRequests: 3,
		Interval:    60 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logger.Warn("circuit breaker state change",
				"name", name,
				"from", from.String(),
				"to", to.String())
			CircuitBreakerState.WithLabelValues(name).Set(float64(to))
		},
	})

	return &Client{
		client:  client,
		limiter: limiter,
		breaker: breaker,
		logger:  logger,
	}
}

func (c *Client) wait(ctx context.Context) error {
	return c.limiter.Wait(ctx)
}

func (c *Client) recordRequest(endpoint string, start time.Time, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}
	APIRequestsTotal.WithLabelValues(endpoint, status).Inc()
	APIRequestDuration.WithLabelValues(endpoint).Observe(time.Since(start).Seconds())
}

func (c *Client) GetPlayers(ctx context.Context) ([]Player, error) {
	c.logger.Info("fetching players from BallDontLie")
	start := time.Now()

	var allPlayers []Player
	cursor := 0

	for {
		if err := c.wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}

		var response PlayersResponse

		result, err := c.breaker.Execute(func() (interface{}, error) {
			req := c.client.R().
				SetContext(ctx).
				SetResult(&response).
				SetQueryParam("per_page", "100")

			if cursor > 0 {
				req.SetQueryParam("cursor", fmt.Sprintf("%d", cursor))
			}

			resp, err := req.Get("/pga/v1/players")
			if err != nil {
				return nil, err
			}

			if resp.IsError() {
				return nil, fmt.Errorf("API error: %s", resp.Status())
			}

			return &response, nil
		})
		if err != nil {
			c.recordRequest("players", start, err)
			if errors.Is(err, gobreaker.ErrOpenState) {
				return nil, fmt.Errorf("failed to fetch players: %w", ErrCircuitOpen)
			}
			return nil, fmt.Errorf("failed to fetch players: %w", err)
		}

		_ = result
		allPlayers = append(allPlayers, response.Data...)

		if response.Meta.NextCursor == 0 {
			break
		}
		cursor = response.Meta.NextCursor
	}

	c.recordRequest("players", start, nil)
	c.logger.Info("players fetch complete", "total", len(allPlayers), "duration", time.Since(start))
	return allPlayers, nil
}

func (c *Client) GetTournaments(ctx context.Context, season int) ([]Tournament, error) {
	c.logger.Info("fetching tournaments", "season", season)
	start := time.Now()

	var allTournaments []Tournament
	cursor := 0

	for {
		if err := c.wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}

		var response TournamentsResponse

		result, err := c.breaker.Execute(func() (interface{}, error) {
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
				return nil, err
			}

			if resp.IsError() {
				return nil, fmt.Errorf("API error: %s", resp.Status())
			}

			return &response, nil
		})
		if err != nil {
			c.recordRequest("tournaments", start, err)
			if errors.Is(err, gobreaker.ErrOpenState) {
				return nil, fmt.Errorf("failed to fetch tournaments: %w", ErrCircuitOpen)
			}
			return nil, fmt.Errorf("failed to fetch tournaments: %w", err)
		}

		_ = result
		for i := range response.Data {
			if response.Data[i].ID > 42 {
				continue
			}
			allTournaments = append(allTournaments, response.Data[i])
		}

		if response.Meta.NextCursor == 0 {
			break
		}
		cursor = response.Meta.NextCursor
	}

	c.recordRequest("tournaments", start, nil)
	c.logger.Info("tournaments fetch complete", "total", len(allTournaments), "duration", time.Since(start))
	return allTournaments, nil
}

func (c *Client) GetTournamentResults(ctx context.Context, tournamentID int) ([]TournamentResult, error) {
	c.logger.Info("fetching tournament results", "tournament_id", tournamentID)
	start := time.Now()

	var allResults []TournamentResult
	cursor := 0

	for {
		if err := c.wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}

		var response TournamentResultsResponse

		result, err := c.breaker.Execute(func() (interface{}, error) {
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
				return nil, err
			}

			if resp.IsError() {
				return nil, fmt.Errorf("API error: %s", resp.Status())
			}

			return &response, nil
		})
		if err != nil {
			c.recordRequest("tournament_results", start, err)
			if errors.Is(err, gobreaker.ErrOpenState) {
				return nil, fmt.Errorf("failed to fetch tournament results: %w", ErrCircuitOpen)
			}
			return nil, fmt.Errorf("failed to fetch tournament results: %w", err)
		}

		_ = result
		allResults = append(allResults, response.Data...)

		if response.Meta.NextCursor == 0 {
			break
		}
		cursor = response.Meta.NextCursor
	}

	c.recordRequest("tournament_results", start, nil)
	c.logger.Info("tournament results fetch complete", "total", len(allResults), "duration", time.Since(start))
	return allResults, nil
}

func (c *Client) GetCourses(ctx context.Context) ([]Course, error) {
	c.logger.Info("fetching courses from BallDontLie")
	start := time.Now()

	var allCourses []Course
	cursor := 0

	for {
		if err := c.wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}

		var response CoursesResponse

		result, err := c.breaker.Execute(func() (interface{}, error) {
			req := c.client.R().
				SetContext(ctx).
				SetResult(&response).
				SetQueryParam("per_page", "100")

			if cursor > 0 {
				req.SetQueryParam("cursor", fmt.Sprintf("%d", cursor))
			}

			resp, err := req.Get("/pga/v1/courses")
			if err != nil {
				return nil, err
			}

			if resp.IsError() {
				return nil, fmt.Errorf("API error: %s", resp.Status())
			}

			return &response, nil
		})
		if err != nil {
			c.recordRequest("courses", start, err)
			if errors.Is(err, gobreaker.ErrOpenState) {
				return nil, fmt.Errorf("failed to fetch courses: %w", ErrCircuitOpen)
			}
			return nil, fmt.Errorf("failed to fetch courses: %w", err)
		}

		_ = result
		allCourses = append(allCourses, response.Data...)

		if response.Meta.NextCursor == 0 {
			break
		}
		cursor = response.Meta.NextCursor
	}

	c.recordRequest("courses", start, nil)
	c.logger.Info("courses fetch complete", "total", len(allCourses), "duration", time.Since(start))
	return allCourses, nil
}

func (c *Client) GetCourseHoles(ctx context.Context, courseID int) ([]CourseHole, error) {
	c.logger.Info("fetching course holes", "course_id", courseID)
	start := time.Now()

	var allHoles []CourseHole
	cursor := 0

	for {
		if err := c.wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}

		var response CourseHolesResponse

		result, err := c.breaker.Execute(func() (interface{}, error) {
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
				return nil, err
			}

			if resp.IsError() {
				return nil, fmt.Errorf("API error: %s", resp.Status())
			}

			return &response, nil
		})
		if err != nil {
			c.recordRequest("course_holes", start, err)
			if errors.Is(err, gobreaker.ErrOpenState) {
				return nil, fmt.Errorf("failed to fetch course holes: %w", ErrCircuitOpen)
			}
			return nil, fmt.Errorf("failed to fetch course holes: %w", err)
		}

		_ = result
		allHoles = append(allHoles, response.Data...)

		if response.Meta.NextCursor == 0 {
			break
		}
		cursor = response.Meta.NextCursor
	}

	c.recordRequest("course_holes", start, nil)
	c.logger.Info("course holes fetch complete", "total", len(allHoles), "duration", time.Since(start))
	return allHoles, nil
}

func (c *Client) GetPlayerRoundResults(ctx context.Context, tournamentID int) ([]PlayerRoundResult, error) {
	c.logger.Info("fetching player round results", "tournament_id", tournamentID)
	start := time.Now()

	var allResults []PlayerRoundResult
	cursor := 0

	for {
		if err := c.wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}

		var response PlayerRoundResultsResponse

		result, err := c.breaker.Execute(func() (interface{}, error) {
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
				return nil, err
			}

			if resp.IsError() {
				return nil, fmt.Errorf("API error: %s", resp.Status())
			}

			return &response, nil
		})
		if err != nil {
			c.recordRequest("player_round_results", start, err)
			if errors.Is(err, gobreaker.ErrOpenState) {
				return nil, fmt.Errorf("failed to fetch player round results: %w", ErrCircuitOpen)
			}
			return nil, fmt.Errorf("failed to fetch player round results: %w", err)
		}

		_ = result
		allResults = append(allResults, response.Data...)

		if response.Meta.NextCursor == 0 {
			break
		}
		cursor = response.Meta.NextCursor
	}

	c.recordRequest("player_round_results", start, nil)
	c.logger.Info("player round results fetch complete", "total", len(allResults), "duration", time.Since(start))
	return allResults, nil
}

func (c *Client) GetPlayerScorecards(ctx context.Context, tournamentID, playerID int) ([]PlayerScorecard, error) {
	c.logger.Info("fetching player scorecards", "tournament_id", tournamentID, "player_id", playerID)
	start := time.Now()

	var allScorecards []PlayerScorecard
	cursor := 0

	for {
		if err := c.wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}

		var response PlayerScorecardsResponse

		result, err := c.breaker.Execute(func() (interface{}, error) {
			req := c.client.R().
				SetContext(ctx).
				SetResult(&response).
				SetQueryParam("tournament_ids[]", fmt.Sprintf("%d", tournamentID)).
				SetQueryParam("player_ids[]", fmt.Sprintf("%d", playerID)).
				SetQueryParam("per_page", "100")

			if cursor > 0 {
				req.SetQueryParam("cursor", fmt.Sprintf("%d", cursor))
			}

			resp, err := req.Get("/pga/v1/player_scorecards")
			if err != nil {
				return nil, err
			}

			if resp.IsError() {
				return nil, fmt.Errorf("API error: %s", resp.Status())
			}

			return &response, nil
		})
		if err != nil {
			c.recordRequest("player_scorecards", start, err)
			if errors.Is(err, gobreaker.ErrOpenState) {
				return nil, fmt.Errorf("failed to fetch player scorecards: %w", ErrCircuitOpen)
			}
			return nil, fmt.Errorf("failed to fetch player scorecards: %w", err)
		}

		_ = result
		allScorecards = append(allScorecards, response.Data...)

		if response.Meta.NextCursor == 0 {
			break
		}
		cursor = response.Meta.NextCursor
	}

	c.recordRequest("player_scorecards", start, nil)
	c.logger.Info("player scorecards fetch complete", "total", len(allScorecards), "duration", time.Since(start))
	return allScorecards, nil
}

func (c *Client) GetPlayerSeasonStats(ctx context.Context, season int, statIDs []int) ([]PlayerSeasonStat, error) {
	c.logger.Info("fetching player season stats", "season", season, "stat_ids", statIDs)
	start := time.Now()

	var allStats []PlayerSeasonStat
	cursor := 0

	for {
		if err := c.wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}

		var response PlayerSeasonStatsResponse

		result, err := c.breaker.Execute(func() (interface{}, error) {
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
				return nil, err
			}

			if resp.IsError() {
				return nil, fmt.Errorf("API error: %s", resp.Status())
			}

			return &response, nil
		})
		if err != nil {
			c.recordRequest("player_season_stats", start, err)
			if errors.Is(err, gobreaker.ErrOpenState) {
				return nil, fmt.Errorf("failed to fetch player season stats: %w", ErrCircuitOpen)
			}
			return nil, fmt.Errorf("failed to fetch player season stats: %w", err)
		}

		_ = result
		allStats = append(allStats, response.Data...)

		if response.Meta.NextCursor == 0 {
			break
		}
		cursor = response.Meta.NextCursor
	}

	c.recordRequest("player_season_stats", start, nil)
	c.logger.Info("player season stats fetch complete", "total", len(allStats), "duration", time.Since(start))
	return allStats, nil
}

func (c *Client) GetTournamentCourseStats(ctx context.Context, tournamentID int) ([]TournamentCourseStats, error) {
	c.logger.Info("fetching tournament course stats", "tournament_id", tournamentID)
	start := time.Now()

	var allStats []TournamentCourseStats
	cursor := 0

	for {
		if err := c.wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}

		var response TournamentCourseStatsResponse

		result, err := c.breaker.Execute(func() (interface{}, error) {
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
				return nil, err
			}

			if resp.IsError() {
				return nil, fmt.Errorf("API error: %s", resp.Status())
			}

			return &response, nil
		})
		if err != nil {
			c.recordRequest("tournament_course_stats", start, err)
			if errors.Is(err, gobreaker.ErrOpenState) {
				return nil, fmt.Errorf("failed to fetch tournament course stats: %w", ErrCircuitOpen)
			}
			return nil, fmt.Errorf("failed to fetch tournament course stats: %w", err)
		}

		_ = result
		allStats = append(allStats, response.Data...)

		if response.Meta.NextCursor == 0 {
			break
		}
		cursor = response.Meta.NextCursor
	}

	c.recordRequest("tournament_course_stats", start, nil)
	c.logger.Info("tournament course stats fetch complete", "total", len(allStats), "duration", time.Since(start))
	return allStats, nil
}

func (c *Client) GetTournamentField(ctx context.Context, tournamentID int) ([]TournamentField, error) {
	c.logger.Info("fetching tournament field", "tournament_id", tournamentID)
	start := time.Now()

	var allEntries []TournamentField
	cursor := 0

	for {
		if err := c.wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}

		var response TournamentFieldResponse

		result, err := c.breaker.Execute(func() (interface{}, error) {
			req := c.client.R().
				SetContext(ctx).
				SetResult(&response).
				SetQueryParam("tournament_id", fmt.Sprintf("%d", tournamentID)).
				SetQueryParam("per_page", "100")

			if cursor > 0 {
				req.SetQueryParam("cursor", fmt.Sprintf("%d", cursor))
			}

			resp, err := req.Get("/pga/v1/tournament_field")
			if err != nil {
				return nil, err
			}

			if resp.IsError() {
				return nil, fmt.Errorf("API error: %s", resp.Status())
			}

			return &response, nil
		})
		if err != nil {
			c.recordRequest("tournament_field", start, err)
			if errors.Is(err, gobreaker.ErrOpenState) {
				return nil, fmt.Errorf("failed to fetch tournament field: %w", ErrCircuitOpen)
			}
			return nil, fmt.Errorf("failed to fetch tournament field: %w", err)
		}

		_ = result
		allEntries = append(allEntries, response.Data...)

		if response.Meta.NextCursor == 0 {
			break
		}
		cursor = response.Meta.NextCursor
	}

	c.recordRequest("tournament_field", start, nil)
	c.logger.Info("tournament field fetch complete", "total", len(allEntries), "duration", time.Since(start))
	return allEntries, nil
}

func (c *Client) GetFutures(ctx context.Context, tournamentID int) ([]Future, error) {
	c.logger.Info("fetching futures", "tournament_id", tournamentID)
	start := time.Now()

	var allFutures []Future
	cursor := 0

	for {
		if err := c.wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}

		var response FuturesResponse

		result, err := c.breaker.Execute(func() (interface{}, error) {
			req := c.client.R().
				SetContext(ctx).
				SetResult(&response).
				SetQueryParam("tournament_ids[]", fmt.Sprintf("%d", tournamentID)).
				SetQueryParam("per_page", "100")

			if cursor > 0 {
				req.SetQueryParam("cursor", fmt.Sprintf("%d", cursor))
			}

			resp, err := req.Get("/pga/v1/futures")
			if err != nil {
				return nil, err
			}

			if resp.IsError() {
				return nil, fmt.Errorf("API error: %s", resp.Status())
			}

			return &response, nil
		})
		if err != nil {
			c.recordRequest("futures", start, err)
			if errors.Is(err, gobreaker.ErrOpenState) {
				return nil, fmt.Errorf("failed to fetch futures: %w", ErrCircuitOpen)
			}
			return nil, fmt.Errorf("failed to fetch futures: %w", err)
		}

		_ = result
		allFutures = append(allFutures, response.Data...)

		if response.Meta.NextCursor == 0 {
			break
		}
		cursor = response.Meta.NextCursor
	}

	c.recordRequest("futures", start, nil)
	c.logger.Info("futures fetch complete", "total", len(allFutures), "duration", time.Since(start))
	return allFutures, nil
}

func (c *Client) GetPlayerRoundStats(ctx context.Context, tournamentID int) ([]PlayerRoundStats, error) {
	c.logger.Info("fetching player round stats", "tournament_id", tournamentID)
	start := time.Now()

	var allStats []PlayerRoundStats
	cursor := 0

	for {
		if err := c.wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}

		var response PlayerRoundStatsResponse

		result, err := c.breaker.Execute(func() (interface{}, error) {
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
				return nil, err
			}

			if resp.IsError() {
				return nil, fmt.Errorf("API error: %s", resp.Status())
			}

			return &response, nil
		})
		if err != nil {
			c.recordRequest("player_round_stats", start, err)
			if errors.Is(err, gobreaker.ErrOpenState) {
				return nil, fmt.Errorf("failed to fetch player round stats: %w", ErrCircuitOpen)
			}
			return nil, fmt.Errorf("failed to fetch player round stats: %w", err)
		}

		_ = result
		allStats = append(allStats, response.Data...)

		if response.Meta.NextCursor == 0 {
			break
		}
		cursor = response.Meta.NextCursor
	}

	c.recordRequest("player_round_stats", start, nil)
	c.logger.Info("player round stats fetch complete", "total", len(allStats), "duration", time.Since(start))
	return allStats, nil
}
