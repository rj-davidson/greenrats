package leagues

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type RecentPick struct {
	GolferName     string `json:"golfer_name"`
	TournamentName string `json:"tournament_name"`
}

type NextDeadline struct {
	TournamentID   uuid.UUID `json:"tournament_id"`
	TournamentName string    `json:"tournament_name"`
	Deadline       time.Time `json:"deadline"`
}

type League struct {
	ID             uuid.UUID     `json:"id"`
	Name           string        `json:"name"`
	Code           string        `json:"code"`
	SeasonYear     int           `json:"season_year"`
	JoiningEnabled bool          `json:"joining_enabled"`
	CreatedAt      time.Time     `json:"created_at"`
	Role           string        `json:"role,omitempty"`
	MemberCount    int           `json:"member_count,omitempty"`
	RecentPick     *RecentPick   `json:"recent_pick,omitempty"`
	NextDeadline   *NextDeadline `json:"next_deadline,omitempty"`
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

type LeagueTournament struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	Status         string    `json:"status"`
	HasUserPick    bool      `json:"has_user_pick"`
	UserPickID     uuid.UUID `json:"user_pick_id,omitempty"`
	GolferName     string    `json:"golfer_name,omitempty"`
	GolferEarnings int       `json:"golfer_earnings,omitempty"`
	PickCount      int       `json:"pick_count"`
}

type ListLeagueTournamentsResponse struct {
	Tournaments []LeagueTournament `json:"tournaments"`
	Total       int                `json:"total"`
}
