package leagues

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

// League represents a league in the response.
type League struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Code       string    `json:"code"`
	SeasonYear int       `json:"season_year"`
	CreatedAt  time.Time `json:"created_at"`
	Role       string    `json:"role,omitempty"`
}

// CreateLeagueRequest represents the request body for creating a league.
type CreateLeagueRequest struct {
	Name string `json:"name"`
}

// CreateLeagueResponse represents the response for creating a league.
type CreateLeagueResponse struct {
	League League `json:"league"`
}

// GetLeagueResponse represents the response for getting a single league.
type GetLeagueResponse struct {
	League League `json:"league"`
}

// ListUserLeaguesResponse represents the response for listing user's leagues.
type ListUserLeaguesResponse struct {
	Leagues []League `json:"leagues"`
	Total   int      `json:"total"`
}
