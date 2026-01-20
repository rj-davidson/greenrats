package scrapedo

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

func (m *MockClient) Scrape(ctx context.Context, targetURL string, opts ScrapeOptions) (*ScrapeResponse, error) {
	args := m.Called(ctx, targetURL, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result, _ := args.Get(0).(*ScrapeResponse)
	return result, args.Error(1)
}
