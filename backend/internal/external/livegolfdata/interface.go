package livegolfdata

import "context"

type ClientInterface interface {
	GetTournaments(ctx context.Context, season int) ([]Tournament, error)
	GetTournamentField(ctx context.Context, tournamentID string) ([]Golfer, error)
	GetLiveLeaderboard(ctx context.Context, tournamentID string) ([]LeaderboardEntry, error)
	GetGolfer(ctx context.Context, golferID string) (*Golfer, error)
	GetEarnings(ctx context.Context, tournamentID string, year int) ([]EarningsEntry, error)
	GetSchedule(ctx context.Context, year int) ([]Tournament, error)
}

var _ ClientInterface = (*Client)(nil)
