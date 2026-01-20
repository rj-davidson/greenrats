package openai

import "context"

type ClientInterface interface {
	SearchTournamentEarnings(ctx context.Context, tournamentName string, year int, golfers []GolferInput) ([]EarningsResult, error)
}

var _ ClientInterface = (*Client)(nil)
