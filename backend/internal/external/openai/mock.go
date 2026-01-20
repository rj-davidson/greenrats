package openai

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
}

var _ ClientInterface = (*MockClient)(nil)

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (m *MockClient) SearchTournamentEarnings(ctx context.Context, tournamentName string, year int, golfers []GolferInput) ([]EarningsResult, error) {
	args := m.Called(ctx, tournamentName, year, golfers)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, _ := args.Get(0).([]EarningsResult)
	return result, args.Error(1)
}

func (m *MockClient) SearchTournamentLeaderboard(ctx context.Context, tournamentName string, year int) (*LeaderboardResponse, error) {
	args := m.Called(ctx, tournamentName, year)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, _ := args.Get(0).(*LeaderboardResponse)
	return result, args.Error(1)
}

func (m *MockClient) MatchPlayersToLeaderboard(ctx context.Context, leaderboard *LeaderboardResponse, golfers []GolferInput) ([]EarningsResult, error) {
	args := m.Called(ctx, leaderboard, golfers)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, _ := args.Get(0).([]EarningsResult)
	return result, args.Error(1)
}
