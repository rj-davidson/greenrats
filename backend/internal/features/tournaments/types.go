package tournaments

import (
	"errors"
	"time"
)

var (
	ErrInvalidTournamentID = errors.New("invalid tournament ID")
	ErrTournamentNotFound  = errors.New("tournament not found")
	ErrInvalidGolferID     = errors.New("invalid golfer ID")
	ErrGolferNotFound      = errors.New("golfer not found")
)

type CourseInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Par     int    `json:"par,omitempty"`
	Yardage int    `json:"yardage,omitempty"`
	City    string `json:"city,omitempty"`
	State   string `json:"state,omitempty"`
	Country string `json:"country,omitempty"`
}

type TournamentCourseInfo struct {
	Course CourseInfo `json:"course"`
	Rounds []int      `json:"rounds"`
}

type Tournament struct {
	ID                 string                 `json:"id"`
	Name               string                 `json:"name"`
	StartDate          time.Time              `json:"start_date"`
	EndDate            time.Time              `json:"end_date"`
	Status             string                 `json:"status"` // upcoming, active, completed
	Course             string                 `json:"course,omitempty"`
	Courses            []TournamentCourseInfo `json:"courses,omitempty"`
	Purse              string                 `json:"purse,omitempty"`
	City               string                 `json:"city,omitempty"`
	State              string                 `json:"state,omitempty"`
	Country            string                 `json:"country,omitempty"`
	Timezone           string                 `json:"timezone,omitempty"`
	PickWindowOpensAt  *time.Time             `json:"pick_window_opens_at,omitempty"`
	PickWindowClosesAt *time.Time             `json:"pick_window_closes_at,omitempty"`
	ChampionID         string                 `json:"champion_id,omitempty"`
	ChampionName       string                 `json:"champion_name,omitempty"`
}

type ListTournamentsRequest struct {
	Season int    `query:"season"`
	Status string `query:"status"`
	Limit  int    `query:"limit"`
	Offset int    `query:"offset"`
}

type ListTournamentsResponse struct {
	Tournaments []Tournament `json:"tournaments"`
	Total       int          `json:"total"`
}

type GetTournamentResponse struct {
	Tournament Tournament `json:"tournament"`
}

type RoundScore struct {
	RoundNumber      int         `json:"round_number"`
	Score            *int        `json:"score"`
	ParRelativeScore *int        `json:"par_relative_score"`
	TeeTime          *time.Time  `json:"tee_time,omitempty"`
	Holes            []HoleScore `json:"holes,omitempty"`
	Course           *CourseInfo `json:"course,omitempty"`
}

type HoleScore struct {
	HoleNumber int  `json:"hole_number"`
	Par        int  `json:"par"`
	Score      *int `json:"score"`
}

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

type GetLeaderboardRequest struct {
	Include  string `query:"include"`   // "holes" for hole-by-hole data
	LeagueID string `query:"league_id"` // optional league context for picks
}

type GetLeaderboardResponse struct {
	TournamentID   string             `json:"tournament_id"`
	TournamentName string             `json:"tournament_name"`
	CurrentRound   int                `json:"current_round"`
	Entries        []LeaderboardEntry `json:"entries"`
	Total          int                `json:"total"`
}

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

type GetFieldResponse struct {
	TournamentID   string       `json:"tournament_id"`
	TournamentName string       `json:"tournament_name"`
	Entries        []FieldEntry `json:"entries"`
	Total          int          `json:"total"`
}

type GetScorecardResponse struct {
	TournamentID   string       `json:"tournament_id"`
	TournamentName string       `json:"tournament_name"`
	GolferID       string       `json:"golfer_id"`
	GolferName     string       `json:"golfer_name"`
	Rounds         []RoundScore `json:"rounds"`
}
