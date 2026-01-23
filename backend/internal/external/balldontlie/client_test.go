package balldontlie

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

func newTestClient(serverURL string) *Client {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := resty.New().
		SetBaseURL(serverURL).
		SetHeader("Content-Type", "application/json")

	return &Client{
		client:  client,
		limiter: rate.NewLimiter(rate.Limit(100), 100),
		logger:  logger,
	}
}

func TestNewClientCreatesRateLimiter(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := New("test-api-key", "https://api.example.com", logger)

	assert.NotNil(t, client.limiter)
	assert.Equal(t, rate.Limit(APIRateLimitPerSecond), client.limiter.Limit())
	assert.Equal(t, APIRateBurst, client.limiter.Burst())
}

func TestRateLimiterThrottlesRequests(t *testing.T) {
	limiter := rate.NewLimiter(rate.Limit(10), 1)

	start := time.Now()

	for i := 0; i < 3; i++ {
		_ = limiter.Wait(t.Context())
	}

	elapsed := time.Since(start)

	assert.GreaterOrEqual(t, elapsed, 150*time.Millisecond, "expected rate limiting to cause delay")
}

func TestConfigConstants(t *testing.T) {
	assert.Equal(t, 2.0, APIRateLimitPerSecond)
	assert.Equal(t, 5, APIRateBurst)
}

func TestGetCourses_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/pga/v1/courses", r.URL.Path)
		assert.Equal(t, "100", r.URL.Query().Get("per_page"))

		response := CoursesResponse{
			Data: []Course{
				{
					ID:      1,
					Name:    "Augusta National",
					City:    strPtr("Augusta"),
					State:   strPtr("GA"),
					Country: strPtr("USA"),
					Par:     intPtr(72),
					Yardage: strPtr("7475"),
				},
				{
					ID:   2,
					Name: "Pebble Beach",
					Par:  intPtr(72),
				},
			},
			Meta: Meta{NextCursor: 0, PerPage: 100},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	courses, err := client.GetCourses(context.Background())

	require.NoError(t, err)
	require.Len(t, courses, 2)
	assert.Equal(t, "Augusta National", courses[0].Name)
	assert.Equal(t, 72, *courses[0].Par)
	assert.Equal(t, "Pebble Beach", courses[1].Name)
}

func TestGetCourses_Pagination(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var response CoursesResponse

		if callCount == 1 {
			response = CoursesResponse{
				Data: []Course{{ID: 1, Name: "Course 1"}},
				Meta: Meta{NextCursor: 2, PerPage: 100},
			}
		} else {
			response = CoursesResponse{
				Data: []Course{{ID: 2, Name: "Course 2"}},
				Meta: Meta{NextCursor: 0, PerPage: 100},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	courses, err := client.GetCourses(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 2, callCount)
	require.Len(t, courses, 2)
}

func TestGetCourses_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	courses, err := client.GetCourses(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "API error")
	assert.Nil(t, courses)
}

func TestGetCourseHoles_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/pga/v1/course_holes", r.URL.Path)
		assert.Equal(t, "123", r.URL.Query().Get("course_ids[]"))

		response := CourseHolesResponse{
			Data: []CourseHole{
				{Course: CourseRef{ID: 123}, HoleNumber: 1, Par: 4, Yardage: intPtr(445)},
				{Course: CourseRef{ID: 123}, HoleNumber: 2, Par: 5, Yardage: intPtr(575)},
				{Course: CourseRef{ID: 123}, HoleNumber: 3, Par: 3, Yardage: intPtr(170)},
			},
			Meta: Meta{NextCursor: 0, PerPage: 100},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	holes, err := client.GetCourseHoles(context.Background(), 123)

	require.NoError(t, err)
	require.Len(t, holes, 3)
	assert.Equal(t, 1, holes[0].HoleNumber)
	assert.Equal(t, 4, holes[0].Par)
	assert.Equal(t, 445, *holes[0].Yardage)
}

func TestGetPlayerRoundResults_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/pga/v1/player_round_results", r.URL.Path)
		assert.Equal(t, "456", r.URL.Query().Get("tournament_ids[]"))

		response := PlayerRoundResultsResponse{
			Data: []PlayerRoundResult{
				{
					Tournament:       TournamentRef{ID: 456, Name: "Masters"},
					Player:           Player{ID: 1, DisplayName: "Scottie Scheffler"},
					RoundNumber:      1,
					Score:            intPtr(68),
					ParRelativeScore: intPtr(-4),
				},
				{
					Tournament:       TournamentRef{ID: 456, Name: "Masters"},
					Player:           Player{ID: 1, DisplayName: "Scottie Scheffler"},
					RoundNumber:      2,
					Score:            intPtr(70),
					ParRelativeScore: intPtr(-2),
				},
			},
			Meta: Meta{NextCursor: 0, PerPage: 100},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	results, err := client.GetPlayerRoundResults(context.Background(), 456)

	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, 1, results[0].RoundNumber)
	assert.Equal(t, 68, *results[0].Score)
	assert.Equal(t, -4, *results[0].ParRelativeScore)
}

func TestGetPlayerScorecards_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/pga/v1/player_scorecards", r.URL.Path)
		assert.Equal(t, "456", r.URL.Query().Get("tournament_ids[]"))

		response := PlayerScorecardsResponse{
			Data: []PlayerScorecard{
				{
					Tournament:  TournamentRef{ID: 456, Name: "Masters"},
					Player:      Player{ID: 1, DisplayName: "Scottie Scheffler"},
					RoundNumber: 1,
					HoleNumber:  1,
					Par:         4,
					Score:       intPtr(3),
				},
				{
					Tournament:  TournamentRef{ID: 456, Name: "Masters"},
					Player:      Player{ID: 1, DisplayName: "Scottie Scheffler"},
					RoundNumber: 1,
					HoleNumber:  2,
					Par:         5,
					Score:       intPtr(4),
				},
			},
			Meta: Meta{NextCursor: 0, PerPage: 100},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	scorecards, err := client.GetPlayerScorecards(context.Background(), 456)

	require.NoError(t, err)
	require.Len(t, scorecards, 2)
	assert.Equal(t, 1, scorecards[0].HoleNumber)
	assert.Equal(t, 4, scorecards[0].Par)
	assert.Equal(t, 3, *scorecards[0].Score)
}

func TestGetPlayerSeasonStats_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/pga/v1/player_season_stats", r.URL.Path)
		assert.Equal(t, "2026", r.URL.Query().Get("season"))

		response := PlayerSeasonStatsResponse{
			Data: []PlayerSeasonStat{
				{
					Player:    Player{ID: 1, DisplayName: "Scottie Scheffler"},
					StatID:    120,
					StatName:  "Scoring Average",
					Season:    2026,
					Rank:      intPtr(1),
					StatValue: []StatValueItem{{StatValue: "68.5"}},
				},
				{
					Player:    Player{ID: 1, DisplayName: "Scottie Scheffler"},
					StatID:    102,
					StatName:  "Driving Distance",
					Season:    2026,
					Rank:      intPtr(15),
					StatValue: []StatValueItem{{StatValue: "310.5"}},
				},
			},
			Meta: Meta{NextCursor: 0, PerPage: 100},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	stats, err := client.GetPlayerSeasonStats(context.Background(), 2026, []int{120, 102})

	require.NoError(t, err)
	require.Len(t, stats, 2)
	assert.Equal(t, "Scoring Average", stats[0].StatName)
	assert.Equal(t, "68.5", stats[0].StatValue[0].StatValue)
	assert.Equal(t, "Driving Distance", stats[1].StatName)
}

func TestGetTournamentCourseStats_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/pga/v1/tournament_course_stats", r.URL.Path)
		assert.Equal(t, "456", r.URL.Query().Get("tournament_ids[]"))

		response := TournamentCourseStatsResponse{
			Data: []TournamentCourseStats{
				{
					Tournament:     TournamentRef{ID: 456, Name: "Masters"},
					Course:         CourseRef{ID: 1, Name: "Augusta National"},
					HoleNumber:     1,
					RoundNumber:    intPtr(1),
					ScoringAverage: floatPtr(4.25),
					DifficultyRank: intPtr(3),
					Birdies:        intPtr(12),
					Pars:           intPtr(45),
				},
				{
					Tournament:     TournamentRef{ID: 456, Name: "Masters"},
					Course:         CourseRef{ID: 1, Name: "Augusta National"},
					HoleNumber:     2,
					RoundNumber:    intPtr(1),
					ScoringAverage: floatPtr(4.85),
					DifficultyRank: intPtr(1),
					Eagles:         intPtr(2),
				},
			},
			Meta: Meta{NextCursor: 0, PerPage: 100},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	stats, err := client.GetTournamentCourseStats(context.Background(), 456)

	require.NoError(t, err)
	require.Len(t, stats, 2)
	assert.Equal(t, 1, stats[0].HoleNumber)
	assert.Equal(t, 4.25, *stats[0].ScoringAverage)
	assert.Equal(t, 3, *stats[0].DifficultyRank)
	assert.Equal(t, 2, stats[1].HoleNumber)
}

func TestGetTournamentField_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/pga/v1/tournament_field", r.URL.Path)
		assert.Equal(t, "456", r.URL.Query().Get("tournament_ids[]"))

		response := TournamentFieldResponse{
			Data: []TournamentField{
				{
					ID:          1,
					Tournament:  Tournament{ID: 456, Name: "Masters", Season: 2025},
					Player:      Player{ID: 1, DisplayName: "Scottie Scheffler"},
					EntryStatus: "Committed",
					Qualifier:   "Exempt",
					OWGR:        intPtr(1),
					IsAmateur:   false,
				},
				{
					ID:          2,
					Tournament:  Tournament{ID: 456, Name: "Masters", Season: 2025},
					Player:      Player{ID: 2, DisplayName: "Rory McIlroy"},
					EntryStatus: "Committed",
					Qualifier:   "Exempt",
					OWGR:        intPtr(3),
					IsAmateur:   false,
				},
			},
			Meta: Meta{NextCursor: 0, PerPage: 100},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	field, err := client.GetTournamentField(context.Background(), 456)

	require.NoError(t, err)
	require.Len(t, field, 2)
	assert.Equal(t, "Scottie Scheffler", field[0].Player.DisplayName)
	assert.Equal(t, "Committed", field[0].EntryStatus)
	assert.Equal(t, 1, *field[0].OWGR)
	assert.Equal(t, "Rory McIlroy", field[1].Player.DisplayName)
}

func TestGetPlayerRoundStats_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/pga/v1/player_round_stats", r.URL.Path)
		assert.Equal(t, "456", r.URL.Query().Get("tournament_ids[]"))

		response := PlayerRoundStatsResponse{
			Data: []PlayerRoundStats{
				{
					Tournament:         TournamentRef{ID: 456, Name: "Masters"},
					Player:             Player{ID: 1, DisplayName: "Scottie Scheffler"},
					RoundNumber:        1,
					SGTotal:            floatPtr(2.5),
					SGTotalRank:        intPtr(1),
					SGOffTee:           floatPtr(0.8),
					SGApproach:         floatPtr(1.2),
					SGAroundGreen:      floatPtr(0.3),
					SGPutting:          floatPtr(0.2),
					DrivingDistance:    floatPtr(310.5),
					GreensInRegulation: floatPtr(77.8),
					Birdies:            intPtr(5),
					Pars:               intPtr(11),
					Bogeys:             intPtr(2),
				},
				{
					Tournament:  TournamentRef{ID: 456, Name: "Masters"},
					Player:      Player{ID: 1, DisplayName: "Scottie Scheffler"},
					RoundNumber: 2,
					SGTotal:     floatPtr(1.8),
					SGTotalRank: intPtr(5),
					Birdies:     intPtr(4),
					Pars:        intPtr(12),
					Bogeys:      intPtr(2),
				},
			},
			Meta: Meta{NextCursor: 0, PerPage: 100},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	stats, err := client.GetPlayerRoundStats(context.Background(), 456)

	require.NoError(t, err)
	require.Len(t, stats, 2)
	assert.Equal(t, 1, stats[0].RoundNumber)
	assert.Equal(t, 2.5, *stats[0].SGTotal)
	assert.Equal(t, 1, *stats[0].SGTotalRank)
	assert.Equal(t, 5, *stats[0].Birdies)
	assert.Equal(t, 2, stats[1].RoundNumber)
}

func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}
