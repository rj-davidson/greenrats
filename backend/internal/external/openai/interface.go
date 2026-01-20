package openai

import "context"

type ClientInterface interface {
	SearchTournamentEarnings(ctx context.Context, tournamentName string, year int, golfers []GolferInput) ([]EarningsResult, error)
	SearchTournamentLeaderboard(ctx context.Context, tournamentName string, year int) (*LeaderboardResponse, error)
	MatchPlayersToLeaderboard(ctx context.Context, leaderboard *LeaderboardResponse, golfers []GolferInput) ([]EarningsResult, error)
	ParseLeaderboardContent(ctx context.Context, content, tournamentName string) (*LeaderboardResponse, error)
}

var _ ClientInterface = (*Client)(nil)
