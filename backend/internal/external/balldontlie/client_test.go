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
					StatValue: 68.5,
				},
				{
					Player:    Player{ID: 1, DisplayName: "Scottie Scheffler"},
					StatID:    102,
					StatName:  "Driving Distance",
					Season:    2026,
					Rank:      intPtr(15),
					StatValue: 310.5,
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
	assert.Equal(t, 68.5, stats[0].StatValue)
	assert.Equal(t, "Driving Distance", stats[1].StatName)
}

func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
