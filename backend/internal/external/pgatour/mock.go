package pgatour

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

func (m *MockClient) GetField(ctx context.Context, fieldID string, includeWithdrawn, changesOnly bool) (*Field, error) {
	args := m.Called(ctx, fieldID, includeWithdrawn, changesOnly)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, _ := args.Get(0).(*Field)
	return result, args.Error(1)
}

func (m *MockClient) GetUpcomingSchedule(ctx context.Context, tourCode string, year int) ([]ScheduleTournament, error) {
	args := m.Called(ctx, tourCode, year)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, _ := args.Get(0).([]ScheduleTournament)
	return result, args.Error(1)
}
