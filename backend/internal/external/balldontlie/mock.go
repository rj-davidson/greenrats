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

func (m *MockClient) GetCourses(ctx context.Context) ([]Course, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, _ := args.Get(0).([]Course)
	return result, args.Error(1)
}

func (m *MockClient) GetCourseHoles(ctx context.Context, courseID int) ([]CourseHole, error) {
	args := m.Called(ctx, courseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, _ := args.Get(0).([]CourseHole)
	return result, args.Error(1)
}

func (m *MockClient) GetPlayerRoundResults(ctx context.Context, tournamentID int) ([]PlayerRoundResult, error) {
	args := m.Called(ctx, tournamentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, _ := args.Get(0).([]PlayerRoundResult)
	return result, args.Error(1)
}

func (m *MockClient) GetPlayerScorecards(ctx context.Context, tournamentID, playerID int) ([]PlayerScorecard, error) {
	args := m.Called(ctx, tournamentID, playerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, _ := args.Get(0).([]PlayerScorecard)
	return result, args.Error(1)
}

func (m *MockClient) GetPlayerSeasonStats(ctx context.Context, season int, statIDs []int) ([]PlayerSeasonStat, error) {
	args := m.Called(ctx, season, statIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, _ := args.Get(0).([]PlayerSeasonStat)
	return result, args.Error(1)
}

func (m *MockClient) GetTournamentCourseStats(ctx context.Context, tournamentID int) ([]TournamentCourseStats, error) {
	args := m.Called(ctx, tournamentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, _ := args.Get(0).([]TournamentCourseStats)
	return result, args.Error(1)
}

func (m *MockClient) GetTournamentField(ctx context.Context, tournamentID int) ([]TournamentField, error) {
	args := m.Called(ctx, tournamentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, _ := args.Get(0).([]TournamentField)
	return result, args.Error(1)
}

func (m *MockClient) GetPlayerRoundStats(ctx context.Context, tournamentID int) ([]PlayerRoundStats, error) {
	args := m.Called(ctx, tournamentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, _ := args.Get(0).([]PlayerRoundStats)
	return result, args.Error(1)
}
