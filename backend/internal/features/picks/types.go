package picks

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type Pick struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	TournamentID   uuid.UUID `json:"tournament_id"`
	GolferID       uuid.UUID `json:"golfer_id"`
	LeagueID       uuid.UUID `json:"league_id"`
	SeasonYear     int       `json:"season_year"`
	CreatedAt      time.Time `json:"created_at"`
	UserName       string    `json:"user_name,omitempty"`
	TournamentName string    `json:"tournament_name,omitempty"`
	GolferName     string    `json:"golfer_name,omitempty"`
	GolferPosition int       `json:"golfer_position,omitempty"`
	GolferEarnings int       `json:"golfer_earnings,omitempty"`
}

type CreatePickRequest struct {
	TournamentID uuid.UUID `json:"tournament_id"`
	GolferID     uuid.UUID `json:"golfer_id"`
	LeagueID     uuid.UUID `json:"league_id"`
}

type CreatePickResponse struct {
	Pick Pick `json:"pick"`
}

type ListPicksResponse struct {
	Picks []Pick `json:"picks"`
	Total int    `json:"total"`
}

type PickWindowStatus struct {
	TournamentID   uuid.UUID `json:"tournament_id"`
	TournamentName string    `json:"tournament_name"`
	IsOpen         bool      `json:"is_open"`
	OpensAt        time.Time `json:"opens_at"`
	ClosesAt       time.Time `json:"closes_at"`
	Reason         string    `json:"reason,omitempty"`
}

type AvailableGolfer struct {
	ID                  uuid.UUID  `json:"id"`
	Name                string     `json:"name"`
	CountryCode         string     `json:"country_code"`
	Country             string     `json:"country,omitempty"`
	OWGR                int        `json:"owgr,omitempty"`
	ImageURL            string     `json:"image_url,omitempty"`
	IsUsed              bool       `json:"is_used"`
	UsedForTournamentID *uuid.UUID `json:"used_for_tournament_id,omitempty"`
	UsedForTournament   string     `json:"used_for_tournament,omitempty"`
}

type AvailableGolfersResponse struct {
	Golfers []AvailableGolfer `json:"golfers"`
	Total   int               `json:"total"`
}

type OverridePickRequest struct {
	GolferID uuid.UUID `json:"golfer_id"`
}

type OverridePickResponse struct {
	Pick Pick `json:"pick"`
}

type UpdatePickRequest struct {
	GolferID uuid.UUID `json:"golfer_id"`
}

type UpdatePickResponse struct {
	Pick Pick `json:"pick"`
}
