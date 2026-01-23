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
	ID                 uuid.UUID  `json:"id"`
	Name               string     `json:"name"`
	StartDate          time.Time  `json:"start_date"`
	EndDate            time.Time  `json:"end_date"`
	Status             string     `json:"status"`
	HasUserPick        bool       `json:"has_user_pick"`
	UserPickID         uuid.UUID  `json:"user_pick_id,omitempty"`
	GolferName         string     `json:"golfer_name,omitempty"`
	GolferEarnings     int        `json:"golfer_earnings,omitempty"`
	PickCount          int        `json:"pick_count"`
	PickWindowOpensAt  *time.Time `json:"pick_window_opens_at,omitempty"`
	PickWindowClosesAt *time.Time `json:"pick_window_closes_at,omitempty"`
}

type ListLeagueTournamentsResponse struct {
	Tournaments []LeagueTournament `json:"tournaments"`
	Total       int                `json:"total"`
}

type CommissionerAction struct {
	ID               uuid.UUID      `json:"id"`
	ActionType       string         `json:"action_type"`
	Description      string         `json:"description"`
	Metadata         map[string]any `json:"metadata,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	CommissionerID   uuid.UUID      `json:"commissioner_id,omitempty"`
	CommissionerName string         `json:"commissioner_name,omitempty"`
	AffectedUserID   uuid.UUID      `json:"affected_user_id,omitempty"`
	AffectedUserName string         `json:"affected_user_name,omitempty"`
}

type CommissionerActionsResponse struct {
	Actions []CommissionerAction `json:"actions"`
	Total   int                  `json:"total"`
}

type MemberPick struct {
	ID         uuid.UUID `json:"id"`
	GolferID   uuid.UUID `json:"golfer_id"`
	GolferName string    `json:"golfer_name"`
}

type LeagueMember struct {
	ID          uuid.UUID   `json:"id"`
	DisplayName string      `json:"display_name"`
	Role        string      `json:"role"`
	JoinedAt    time.Time   `json:"joined_at"`
	Pick        *MemberPick `json:"pick,omitempty"`
}

type LeagueMembersResponse struct {
	Members []LeagueMember `json:"members"`
	Total   int            `json:"total"`
}
