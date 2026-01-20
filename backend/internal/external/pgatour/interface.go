package pgatour

import "context"

type ClientInterface interface {
	GetField(ctx context.Context, fieldID string, includeWithdrawn, changesOnly bool) (*Field, error)
	GetUpcomingSchedule(ctx context.Context, tourCode string, year int) ([]ScheduleTournament, error)
}

var _ ClientInterface = (*Client)(nil)
