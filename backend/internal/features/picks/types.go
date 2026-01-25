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

type GolferSeasonStats struct {
	ScoringAvg      *float64 `json:"scoring_avg,omitempty"`
	DrivingDistance *float64 `json:"driving_distance,omitempty"`
	DrivingAccuracy *float64 `json:"driving_accuracy,omitempty"`
	GIRPct          *float64 `json:"gir_pct,omitempty"`
	PuttingAvg      *float64 `json:"putting_avg,omitempty"`
	ScramblingPct   *float64 `json:"scrambling_pct,omitempty"`
	Top10s          *int     `json:"top_10s,omitempty"`
	CutsMade        *int     `json:"cuts_made,omitempty"`
	EventsPlayed    *int     `json:"events_played,omitempty"`
	Wins            *int     `json:"wins,omitempty"`
	Earnings        *int     `json:"earnings,omitempty"`
}

type GolferBio struct {
	Height            string     `json:"height,omitempty"`
	Weight            string     `json:"weight,omitempty"`
	BirthDate         *time.Time `json:"birth_date,omitempty"`
	BirthplaceCity    string     `json:"birthplace_city,omitempty"`
	BirthplaceState   string     `json:"birthplace_state,omitempty"`
	BirthplaceCountry string     `json:"birthplace_country,omitempty"`
	TurnedPro         *int       `json:"turned_pro,omitempty"`
	School            string     `json:"school,omitempty"`
	ResidenceCity     string     `json:"residence_city,omitempty"`
	ResidenceState    string     `json:"residence_state,omitempty"`
	ResidenceCountry  string     `json:"residence_country,omitempty"`
}

type PickFieldEntry struct {
	GolferID              uuid.UUID          `json:"golfer_id"`
	GolferName            string             `json:"golfer_name"`
	CountryCode           string             `json:"country_code"`
	Country               string             `json:"country,omitempty"`
	ImageURL              string             `json:"image_url,omitempty"`
	EntryStatus           string             `json:"entry_status"`
	Qualifier             string             `json:"qualifier,omitempty"`
	OWGR                  *int               `json:"owgr,omitempty"`
	OWGRAtEntry           *int               `json:"owgr_at_entry,omitempty"`
	SeasonEarnings        *int               `json:"season_earnings,omitempty"`
	IsAmateur             bool               `json:"is_amateur"`
	IsUsed                bool               `json:"is_used"`
	UsedForTournamentID   *uuid.UUID         `json:"used_for_tournament_id,omitempty"`
	UsedForTournamentName string             `json:"used_for_tournament_name,omitempty"`
	SeasonStats           *GolferSeasonStats `json:"season_stats,omitempty"`
	Bio                   *GolferBio         `json:"bio,omitempty"`
}

type GetPickFieldResponse struct {
	TournamentID        uuid.UUID        `json:"tournament_id"`
	TournamentName      string           `json:"tournament_name"`
	Course              string           `json:"course,omitempty"`
	City                string           `json:"city,omitempty"`
	State               string           `json:"state,omitempty"`
	Country             string           `json:"country,omitempty"`
	Purse               *int             `json:"purse,omitempty"`
	StartDate           time.Time        `json:"start_date"`
	EndDate             time.Time        `json:"end_date"`
	PickWindowState     string           `json:"pick_window_state"`
	PickWindowOpensAt   *time.Time       `json:"pick_window_opens_at,omitempty"`
	PickWindowClosesAt  *time.Time       `json:"pick_window_closes_at,omitempty"`
	CurrentPickID       *uuid.UUID       `json:"current_pick_id,omitempty"`
	CurrentPickGolferID *uuid.UUID       `json:"current_pick_golfer_id,omitempty"`
	Entries             []PickFieldEntry `json:"entries"`
	Total               int              `json:"total"`
	AvailableCount      int              `json:"available_count"`
}
