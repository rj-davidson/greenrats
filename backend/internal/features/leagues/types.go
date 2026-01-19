package leagues

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type League struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Code           string    `json:"code"`
	SeasonYear     int       `json:"season_year"`
	JoiningEnabled bool      `json:"joining_enabled"`
	CreatedAt      time.Time `json:"created_at"`
	Role           string    `json:"role,omitempty"`
	MemberCount    int       `json:"member_count,omitempty"`
}

type CreateLeagueRequest struct {
	Name string `json:"name"`
}

type CreateLeagueResponse struct {
	League League `json:"league"`
}

type GetLeagueResponse struct {
	League League `json:"league"`
}

type ListUserLeaguesResponse struct {
	Leagues []League `json:"leagues"`
	Total   int      `json:"total"`
}

type JoinLeagueRequest struct {
	Code string `json:"code"`
}

type JoinLeagueResponse struct {
	League League `json:"league"`
}

type SetJoiningEnabledRequest struct {
	Enabled bool `json:"enabled"`
}

type SetJoiningEnabledResponse struct {
	League League `json:"league"`
}

type RegenerateCodeResponse struct {
	League League `json:"league"`
}
