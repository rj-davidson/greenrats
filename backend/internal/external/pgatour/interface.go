package pgatour

import "context"

type ClientInterface interface {
	GetTournamentField(ctx context.Context, tournamentID string) ([]FieldEntry, error)
}

var _ ClientInterface = (*Client)(nil)
