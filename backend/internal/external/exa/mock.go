package exa

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

func (m *MockClient) SearchEarnings(ctx context.Context, tournamentName string, year int) (*SearchResponse, error) {
	args := m.Called(ctx, tournamentName, year)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, _ := args.Get(0).(*SearchResponse)
	return result, args.Error(1)
}
