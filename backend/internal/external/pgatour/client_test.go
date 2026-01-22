package pgatour

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

func newTestClient(serverURL string, logger *slog.Logger) *Client {
	client := resty.New().
		SetBaseURL(serverURL).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json")

	return &Client{
		client:  client,
		limiter: rate.NewLimiter(rate.Limit(APIRateLimitPerSecond), APIRateBurst),
		logger:  logger,
	}
}

func TestNewClient_DefaultURL(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := New("test-api-key", logger)

	assert.NotNil(t, client)
	assert.NotNil(t, client.client)
	assert.NotNil(t, client.limiter)
}

func TestNewClient_WithAPIKey(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := New("my-api-key", logger)

	assert.NotNil(t, client)
}

func TestNewClient_RateLimiter(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := New("", logger)

	assert.Equal(t, rate.Limit(APIRateLimitPerSecond), client.limiter.Limit())
	assert.Equal(t, APIRateBurst, client.limiter.Burst())
}

func TestConfigConstants(t *testing.T) {
	assert.Equal(t, 1, APIRateLimitPerSecond)
	assert.Equal(t, 3, APIRateBurst)
	assert.Equal(t, "https://orchestrator.pgatour.com/graphql", DefaultGraphQLURL)
}

func TestGetTournamentField_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req GraphQLRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "Field", req.OperationName)
		assert.Equal(t, "R2025001", req.Variables["id"])

		response := GraphQLResponse{
			Data: &ResponseData{
				Field: &Field{
					TournamentName: "The Sentry",
					ID:             "R2025001",
					Players: []FieldPlayer{
						{
							ID:          "12345",
							FirstName:   "Scottie",
							LastName:    "Scheffler",
							DisplayName: "Scottie Scheffler",
							CountryCode: "USA",
							IsAmateur:   false,
						},
						{
							ID:          "67890",
							FirstName:   "Rory",
							LastName:    "McIlroy",
							DisplayName: "Rory McIlroy",
							CountryCode: "NIR",
							IsAmateur:   false,
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := newTestClient(server.URL, logger)

	entries, err := client.GetTournamentField(context.Background(), "R2025001")

	require.NoError(t, err)
	require.Len(t, entries, 2)

	assert.Equal(t, "12345", entries[0].PGATourID)
	assert.Equal(t, "Scottie Scheffler", entries[0].DisplayName)
	assert.Equal(t, "Scottie", entries[0].FirstName)
	assert.Equal(t, "Scheffler", entries[0].LastName)
	assert.Equal(t, "USA", entries[0].CountryCode)
	assert.False(t, entries[0].IsAmateur)

	assert.Equal(t, "67890", entries[1].PGATourID)
	assert.Equal(t, "Rory McIlroy", entries[1].DisplayName)
	assert.Equal(t, "NIR", entries[1].CountryCode)
}

func TestGetTournamentField_EmptyField(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := GraphQLResponse{
			Data: &ResponseData{
				Field: nil,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := newTestClient(server.URL, logger)

	entries, err := client.GetTournamentField(context.Background(), "R2025001")

	require.NoError(t, err)
	assert.Nil(t, entries)
}

func TestGetTournamentField_NoData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := GraphQLResponse{
			Data: nil,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := newTestClient(server.URL, logger)

	entries, err := client.GetTournamentField(context.Background(), "R2025001")

	require.NoError(t, err)
	assert.Nil(t, entries)
}

func TestGetTournamentField_GraphQLError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := GraphQLResponse{
			Errors: []GraphQLError{
				{Message: "Tournament not found"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := newTestClient(server.URL, logger)

	entries, err := client.GetTournamentField(context.Background(), "INVALID")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "GraphQL error: Tournament not found")
	assert.Nil(t, entries)
}

func TestGetTournamentField_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := newTestClient(server.URL, logger)

	entries, err := client.GetTournamentField(context.Background(), "R2025001")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "API error")
	assert.Nil(t, entries)
}

func TestGetTournamentField_Amateur(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := GraphQLResponse{
			Data: &ResponseData{
				Field: &Field{
					TournamentName: "Masters",
					ID:             "R2025001",
					Players: []FieldPlayer{
						{
							ID:          "12345",
							FirstName:   "John",
							LastName:    "Doe",
							DisplayName: "John Doe",
							CountryCode: "USA",
							IsAmateur:   true,
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := newTestClient(server.URL, logger)

	entries, err := client.GetTournamentField(context.Background(), "R2025001")

	require.NoError(t, err)
	require.Len(t, entries, 1)

	assert.Equal(t, "John Doe", entries[0].DisplayName)
	assert.True(t, entries[0].IsAmateur)
}

func TestGetTournamentField_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := newTestClient(server.URL, logger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.GetTournamentField(ctx, "R2025001")

	require.Error(t, err)
}
