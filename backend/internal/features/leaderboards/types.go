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
	TournamentID   uuid.UUID `json:"tournament_id"`
	TournamentName string    `json:"tournament_name"`
	GolferID       uuid.UUID `json:"golfer_id"`
	GolferName     string    `json:"golfer_name"`
	Position       int       `json:"position"`
	Status         string    `json:"status"`
	Earnings       int       `json:"earnings"`
}

type ActivePickEntry struct {
	TournamentID       uuid.UUID `json:"tournament_id"`
	TournamentName     string    `json:"tournament_name"`
	HasPick            bool      `json:"has_pick"`
	GolferID           *string   `json:"golfer_id,omitempty"`
	GolferName         *string   `json:"golfer_name,omitempty"`
	IsPickWindowClosed bool      `json:"is_pick_window_closed"`
}

type StandingsEntry struct {
	UserID          uuid.UUID         `json:"user_id"`
	UserDisplayName string            `json:"user_display_name"`
	TotalEarnings   int               `json:"total_earnings"`
	PickCount       int               `json:"pick_count"`
	HasCurrentPick  bool              `json:"has_current_pick"`
	CurrentPick     *CurrentPick      `json:"current_pick,omitempty"`
	ActivePicks     []ActivePickEntry `json:"active_picks,omitempty"`
	Picks           []PickHistory     `json:"picks,omitempty"`
}

type GetStandingsRequest struct {
	Include string `query:"include"` // "picks" for pick history
}

type LeagueStandingsResponse struct {
	Entries           []StandingsEntry   `json:"entries"`
	Total             int                `json:"total"`
	SeasonYear        int                `json:"season_year"`
	ActiveTournament  *ActiveTournament  `json:"active_tournament,omitempty"`
	ActiveTournaments []ActiveTournament `json:"active_tournaments,omitempty"`
}
