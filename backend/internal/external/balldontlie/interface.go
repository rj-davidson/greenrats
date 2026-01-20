package balldontlie

import "context"

type ClientInterface interface {
	GetPlayers(ctx context.Context) ([]Player, error)
	GetTournaments(ctx context.Context, season int) ([]Tournament, error)
	GetTournamentResults(ctx context.Context, tournamentID int) ([]TournamentResult, error)
}

var _ ClientInterface = (*Client)(nil)
