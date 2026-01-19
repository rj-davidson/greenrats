package leaderboards

import "github.com/gofrs/uuid/v5"

type LeaderboardEntry struct {
	Rank        int       `json:"rank"`
	UserID      uuid.UUID `json:"user_id"`
	DisplayName string    `json:"display_name"`
	Earnings    int       `json:"earnings"`
	PickCount   int       `json:"pick_count"`
}

type LeagueLeaderboardResponse struct {
	Entries    []LeaderboardEntry `json:"entries"`
	Total      int                `json:"total"`
	SeasonYear int                `json:"season_year"`
}
