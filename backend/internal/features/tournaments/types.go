package tournaments

import (
	"errors"
	"time"
)

var (
	ErrInvalidTournamentID = errors.New("invalid tournament ID")
	ErrTournamentNotFound  = errors.New("tournament not found")
)

// Tournament represents a golf tournament.
type Tournament struct {
	ID                 string     `json:"id"`
	Name               string     `json:"name"`
	StartDate          time.Time  `json:"start_date"`
	EndDate            time.Time  `json:"end_date"`
	Status             string     `json:"status"` // upcoming, active, completed
	Course             string     `json:"course,omitempty"`
	Purse              float64    `json:"purse,omitempty"`
	City               string     `json:"city,omitempty"`
	State              string     `json:"state,omitempty"`
	Country            string     `json:"country,omitempty"`
	Timezone           string     `json:"timezone,omitempty"`
	PickWindowOpensAt  *time.Time `json:"pick_window_opens_at,omitempty"`
	PickWindowClosesAt *time.Time `json:"pick_window_closes_at,omitempty"`
	ChampionID         string     `json:"champion_id,omitempty"`
	ChampionName       string     `json:"champion_name,omitempty"`
}

// ListTournamentsRequest represents the request parameters for listing tournaments.
type ListTournamentsRequest struct {
	Season int    `query:"season"`
	Status string `query:"status"`
	Limit  int    `query:"limit"`
	Offset int    `query:"offset"`
}

// ListTournamentsResponse represents the response for listing tournaments.
type ListTournamentsResponse struct {
	Tournaments []Tournament `json:"tournaments"`
	Total       int          `json:"total"`
}

// GetTournamentResponse represents the response for getting a single tournament.
type GetTournamentResponse struct {
	Tournament Tournament `json:"tournament"`
}

// RoundScore represents a golfer's score for a single round.
type RoundScore struct {
	RoundNumber      int         `json:"round_number"`
	Score            *int        `json:"score"`
	ParRelativeScore *int        `json:"par_relative_score"`
	TeeTime          *time.Time  `json:"tee_time,omitempty"`
	Holes            []HoleScore `json:"holes,omitempty"`
}

// HoleScore represents a golfer's score on a single hole.
type HoleScore struct {
	HoleNumber int  `json:"hole_number"`
	Par        int  `json:"par"`
	Score      *int `json:"score"`
}

// LeaderboardEntry represents a golfer's position on the tournament leaderboard.
type LeaderboardEntry struct {
	Position         int          `json:"position"`
	PreviousPosition *int         `json:"previous_position,omitempty"`
	PositionChange   *int         `json:"position_change,omitempty"`
	GolferID         string       `json:"golfer_id"`
	GolferName       string       `json:"golfer_name"`
	CountryCode      string       `json:"country_code"`
	Country          string       `json:"country,omitempty"`
	ImageURL         string       `json:"image_url,omitempty"`
	Score            int          `json:"score"`
	TotalStrokes     int          `json:"total_strokes"`
	Thru             int          `json:"thru"`
	CurrentRound     int          `json:"current_round"`
	Status           string       `json:"status"`
	Earnings         int          `json:"earnings"`
	Rounds           []RoundScore `json:"rounds"`
	PickedBy         []string     `json:"picked_by,omitempty"`
}

// GetLeaderboardRequest represents optional query parameters for leaderboard.
type GetLeaderboardRequest struct {
	Include  string `query:"include"`   // "holes" for hole-by-hole data
	LeagueID string `query:"league_id"` // optional league context for picks
}

// GetLeaderboardResponse represents the response for getting a tournament leaderboard.
type GetLeaderboardResponse struct {
	TournamentID   string             `json:"tournament_id"`
	TournamentName string             `json:"tournament_name"`
	CurrentRound   int                `json:"current_round"`
	Entries        []LeaderboardEntry `json:"entries"`
	Total          int                `json:"total"`
}

// FieldEntry represents a golfer in the tournament field.
type FieldEntry struct {
	GolferID    string `json:"golfer_id"`
	GolferName  string `json:"golfer_name"`
	CountryCode string `json:"country_code"`
	Country     string `json:"country,omitempty"`
	OWGR        *int   `json:"owgr,omitempty"`
	OWGRAtEntry *int   `json:"owgr_at_entry,omitempty"`
	EntryStatus string `json:"entry_status"`
	Qualifier   string `json:"qualifier,omitempty"`
	IsAmateur   bool   `json:"is_amateur"`
	ImageURL    string `json:"image_url,omitempty"`
}

// GetFieldResponse represents the response for getting a tournament field.
type GetFieldResponse struct {
	TournamentID   string       `json:"tournament_id"`
	TournamentName string       `json:"tournament_name"`
	Entries        []FieldEntry `json:"entries"`
	Total          int          `json:"total"`
}
