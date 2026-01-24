package leaderboards

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type CurrentPick struct {
	TournamentID   uuid.UUID `json:"tournament_id"`
	TournamentName string    `json:"tournament_name"`
	GolferID       uuid.UUID `json:"golfer_id"`
	GolferName     string    `json:"golfer_name"`
}

type ActiveTournament struct {
	ID                 uuid.UUID `json:"id"`
	Name               string    `json:"name"`
	IsPickWindowClosed bool      `json:"is_pick_window_closed"`
	StartDate          time.Time `json:"start_date"`
}

type LeaderboardEntry struct {
	Rank        int       `json:"rank"`
	RankDisplay string    `json:"rank_display"`
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

type PickHistory struct {
	TournamentID    uuid.UUID `json:"tournament_id"`
	TournamentName  string    `json:"tournament_name"`
	GolferID        uuid.UUID `json:"golfer_id"`
	GolferName      string    `json:"golfer_name"`
	PositionDisplay string    `json:"position_display,omitempty"`
	Earnings        int       `json:"earnings"`
}

type StandingsEntry struct {
	Rank            int           `json:"rank"`
	RankDisplay     string        `json:"rank_display"`
	UserID          uuid.UUID     `json:"user_id"`
	UserDisplayName string        `json:"user_display_name"`
	TotalEarnings   int           `json:"total_earnings"`
	PickCount       int           `json:"pick_count"`
	CurrentPick     *CurrentPick  `json:"current_pick,omitempty"`
	Picks           []PickHistory `json:"picks,omitempty"`
}

type GetStandingsRequest struct {
	Include string `query:"include"` // "picks" for pick history
}

type LeagueStandingsResponse struct {
	Entries          []StandingsEntry  `json:"entries"`
	Total            int               `json:"total"`
	SeasonYear       int               `json:"season_year"`
	ActiveTournament *ActiveTournament `json:"active_tournament,omitempty"`
}
