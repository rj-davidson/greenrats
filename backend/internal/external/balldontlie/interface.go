package balldontlie

import "context"

type ClientInterface interface {
	GetPlayers(ctx context.Context) ([]Player, error)
	GetTournaments(ctx context.Context, season int) ([]Tournament, error)
	GetTournamentResults(ctx context.Context, tournamentID int) ([]TournamentResult, error)

	// ALL-STAR tier endpoints
	GetTournamentCourseStats(ctx context.Context, tournamentID int) ([]TournamentCourseStats, error)

	// GOAT tier endpoints
	GetCourses(ctx context.Context) ([]Course, error)
	GetCourseHoles(ctx context.Context, courseID int) ([]CourseHole, error)
	GetPlayerRoundResults(ctx context.Context, tournamentID int) ([]PlayerRoundResult, error)
	GetPlayerRoundStats(ctx context.Context, tournamentID int) ([]PlayerRoundStats, error)
	GetPlayerScorecards(ctx context.Context, tournamentID, playerID int) ([]PlayerScorecard, error)
	GetPlayerSeasonStats(ctx context.Context, season int, statIDs []int) ([]PlayerSeasonStat, error)
	GetTournamentField(ctx context.Context, tournamentID int) ([]TournamentField, error)
}

var _ ClientInterface = (*Client)(nil)
