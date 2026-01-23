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

type PickRoundScore struct {
	RoundNumber      int  `json:"round_number"`
	Score            *int `json:"score"`
	ParRelativeScore *int `json:"par_relative_score"`
}

type PickLeaderboardData struct {
	Position        int              `json:"position"`
	PositionDisplay string           `json:"position_display"`
	Score           int              `json:"score"`
	Thru            int              `json:"thru"`
	CurrentRound    int              `json:"current_round"`
	Cut             bool             `json:"cut"`
	Status          string           `json:"status"`
	Earnings        int              `json:"earnings"`
	Rounds          []PickRoundScore `json:"rounds,omitempty"`
}

type LeaguePickEntry struct {
	PickID            uuid.UUID            `json:"pick_id"`
	UserID            uuid.UUID            `json:"user_id"`
	UserDisplayName   string               `json:"user_display_name"`
	GolferID          uuid.UUID            `json:"golfer_id"`
	GolferName        string               `json:"golfer_name"`
	GolferCountryCode string               `json:"golfer_country_code"`
	GolferImageURL    string               `json:"golfer_image_url,omitempty"`
	CreatedAt         time.Time            `json:"created_at"`
	Leaderboard       *PickLeaderboardData `json:"leaderboard,omitempty"`
}

type GetLeaguePicksRequest struct {
	Include string `query:"include"` // "rounds" for round-by-round data
}

type GetLeaguePicksResponse struct {
	Entries             []LeaguePickEntry `json:"entries"`
	Total               int               `json:"total"`
	MembersWithoutPicks int               `json:"members_without_picks"`
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

type CreatePickForUserRequest struct {
	UserID       uuid.UUID `json:"user_id"`
	TournamentID uuid.UUID `json:"tournament_id"`
	GolferID     uuid.UUID `json:"golfer_id"`
}

type CreatePickForUserResponse struct {
	Pick Pick `json:"pick"`
}

type UserPublicPick struct {
	TournamentID        uuid.UUID `json:"tournament_id"`
	TournamentName      string    `json:"tournament_name"`
	TournamentStartDate time.Time `json:"tournament_start_date"`
	GolferID            uuid.UUID `json:"golfer_id"`
	GolferName          string    `json:"golfer_name"`
	PositionDisplay     string    `json:"position_display,omitempty"`
	Earnings            int       `json:"earnings"`
}

type UserPublicPicksResponse struct {
	Picks []UserPublicPick `json:"picks"`
}
