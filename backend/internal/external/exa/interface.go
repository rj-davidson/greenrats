package exa

import "context"

type ClientInterface interface {
	SearchEarnings(ctx context.Context, tournamentName string, year int) (*SearchResponse, error)
}

var _ ClientInterface = (*Client)(nil)
