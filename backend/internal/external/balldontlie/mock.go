package balldontlie

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

func (m *MockClient) GetPlayers(ctx context.Context) ([]Player, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, _ := args.Get(0).([]Player)
	return result, args.Error(1)
}

func (m *MockClient) GetTournaments(ctx context.Context, season int) ([]Tournament, error) {
	args := m.Called(ctx, season)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, _ := args.Get(0).([]Tournament)
	return result, args.Error(1)
}

func (m *MockClient) GetTournamentResults(ctx context.Context, tournamentID int) ([]TournamentResult, error) {
	args := m.Called(ctx, tournamentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, _ := args.Get(0).([]TournamentResult)
	return result, args.Error(1)
}
