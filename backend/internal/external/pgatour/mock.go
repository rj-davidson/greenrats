package pgatour

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (m *MockClient) GetTournamentField(ctx context.Context, tournamentID string) ([]FieldEntry, error) {
	args := m.Called(ctx, tournamentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	entries, _ := args.Get(0).([]FieldEntry)
	return entries, args.Error(1)
}

var _ ClientInterface = (*MockClient)(nil)
