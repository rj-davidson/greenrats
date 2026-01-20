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
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Status    string    `json:"status"` // upcoming, active, completed
	Venue     string    `json:"venue,omitempty"`
	Course    string    `json:"course,omitempty"`
	Purse     float64   `json:"purse,omitempty"`
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

// LeaderboardEntry represents a golfer's position on the tournament leaderboard.
type LeaderboardEntry struct {
	Position        int    `json:"position"`
	PositionDisplay string `json:"position_display"`
	GolferID        string `json:"golfer_id"`
	GolferName      string `json:"golfer_name"`
	CountryCode     string `json:"country_code"`
	Score           int    `json:"score"`
	TotalStrokes    int    `json:"total_strokes"`
	Thru            int    `json:"thru"`
	CurrentRound    int    `json:"current_round"`
	Cut             bool   `json:"cut"`
	Status          string `json:"status"`
	Earnings        int    `json:"earnings"`
}

// GetLeaderboardResponse represents the response for getting a tournament leaderboard.
type GetLeaderboardResponse struct {
	Entries []LeaderboardEntry `json:"entries"`
	Total   int                `json:"total"`
}
