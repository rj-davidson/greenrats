package livegolfdata

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

func (m *MockClient) GetTournaments(ctx context.Context, season int) ([]Tournament, error) {
	args := m.Called(ctx, season)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, _ := args.Get(0).([]Tournament)
	return result, args.Error(1)
}

func (m *MockClient) GetTournamentField(ctx context.Context, tournamentID string) ([]Golfer, error) {
	args := m.Called(ctx, tournamentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, _ := args.Get(0).([]Golfer)
	return result, args.Error(1)
}

func (m *MockClient) GetLiveLeaderboard(ctx context.Context, tournamentID string) ([]LeaderboardEntry, error) {
	args := m.Called(ctx, tournamentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, _ := args.Get(0).([]LeaderboardEntry)
	return result, args.Error(1)
}

func (m *MockClient) GetGolfer(ctx context.Context, golferID string) (*Golfer, error) {
	args := m.Called(ctx, golferID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, _ := args.Get(0).(*Golfer)
	return result, args.Error(1)
}

func (m *MockClient) GetEarnings(ctx context.Context, tournamentID string, year int) ([]EarningsEntry, error) {
	args := m.Called(ctx, tournamentID, year)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, _ := args.Get(0).([]EarningsEntry)
	return result, args.Error(1)
}

func (m *MockClient) GetSchedule(ctx context.Context, year int) ([]Tournament, error) {
	args := m.Called(ctx, year)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, _ := args.Get(0).([]Tournament)
	return result, args.Error(1)
}
