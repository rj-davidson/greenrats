package balldontlie

import "context"

type ClientInterface interface {
	GetPlayers(ctx context.Context) ([]Player, error)
	GetTournaments(ctx context.Context, season int) ([]Tournament, error)
	GetTournamentResults(ctx context.Context, tournamentID int) ([]TournamentResult, error)

	// GOAT tier endpoints
	GetCourses(ctx context.Context) ([]Course, error)
	GetCourseHoles(ctx context.Context, courseID int) ([]CourseHole, error)
	GetPlayerRoundResults(ctx context.Context, tournamentID int) ([]PlayerRoundResult, error)
	GetPlayerScorecards(ctx context.Context, tournamentID int) ([]PlayerScorecard, error)
	GetPlayerSeasonStats(ctx context.Context, season int, statIDs []int) ([]PlayerSeasonStat, error)
}

var _ ClientInterface = (*Client)(nil)
