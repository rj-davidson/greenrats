package balldontlie

// FREE tier

type Player struct {
	ID                int     `json:"id"`
	FirstName         *string `json:"first_name"`
	LastName          *string `json:"last_name"`
	DisplayName       string  `json:"display_name"`
	Country           *string `json:"country"`
	CountryCode       *string `json:"country_code"`
	Height            *string `json:"height"`
	Weight            *string `json:"weight"`
	BirthDate         *string `json:"birth_date"`
	BirthplaceCity    *string `json:"birthplace_city"`
	BirthplaceState   *string `json:"birthplace_state"`
	BirthplaceCountry *string `json:"birthplace_country"`
	TurnedPro         *string `json:"turned_pro"`
	School            *string `json:"school"`
	ResidenceCity     *string `json:"residence_city"`
	ResidenceState    *string `json:"residence_state"`
	ResidenceCountry  *string `json:"residence_country"`
	OWGR              *int    `json:"owgr"`
	Active            bool    `json:"active"`
}

type Tournament struct {
	ID         int     `json:"id"`
	Season     int     `json:"season"`
	Name       string  `json:"name"`
	StartDate  string  `json:"start_date"`
	EndDate    *string `json:"end_date"`
	City       *string `json:"city"`
	State      *string `json:"state"`
	Country    *string `json:"country"`
	CourseName *string `json:"course_name"`
	Purse      *string `json:"purse"`
	Status     *string `json:"status"`
	Champion   *Player `json:"champion"`
}

type Course struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	City         *string `json:"city"`
	State        *string `json:"state"`
	Country      *string `json:"country"`
	Par          *int    `json:"par"`
	Yardage      *string `json:"yardage"`
	Established  *string `json:"established"`
	Architect    *string `json:"architect"`
	FairwayGrass *string `json:"fairway_grass"`
	RoughGrass   *string `json:"rough_grass"`
	GreenGrass   *string `json:"green_grass"`
}

// ALL-STAR tier ($9.99/mo)

type TournamentResult struct {
	Tournament       TournamentRef `json:"tournament"`
	Player           Player        `json:"player"`
	Position         *string       `json:"position"`
	PositionNumeric  *int          `json:"position_numeric"`
	TotalScore       *int          `json:"total_score"`
	ParRelativeScore *int          `json:"par_relative_score"`
}

type TournamentCourseStats struct {
	Tournament     TournamentRef `json:"tournament"`
	Course         CourseRef     `json:"course"`
	HoleNumber     int           `json:"hole_number"`
	RoundNumber    *int          `json:"round_number"`
	ScoringAverage *float64      `json:"scoring_average"`
	ScoringDiff    *float64      `json:"scoring_diff"`
	DifficultyRank *int          `json:"difficulty_rank"`
	Eagles         *int          `json:"eagles"`
	Birdies        *int          `json:"birdies"`
	Pars           *int          `json:"pars"`
	Bogeys         *int          `json:"bogeys"`
	DoubleBogeys   *int          `json:"double_bogeys"`
}

type CourseHole struct {
	Course     CourseRef `json:"course"`
	HoleNumber int       `json:"hole_number"`
	Par        int       `json:"par"`
	Yardage    *int      `json:"yardage"`
}

// Shared reference types

type TournamentRef struct {
	ID         int     `json:"id"`
	Season     int     `json:"season"`
	Name       string  `json:"name"`
	StartDate  string  `json:"start_date"`
	EndDate    *string `json:"end_date"`
	City       *string `json:"city"`
	State      *string `json:"state"`
	Country    *string `json:"country"`
	CourseName *string `json:"course_name"`
	Purse      *string `json:"purse"`
	Status     *string `json:"status"`
}

type CourseRef struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	City    *string `json:"city"`
	State   *string `json:"state"`
	Country *string `json:"country"`
	Par     *int    `json:"par"`
	Yardage *string `json:"yardage"`
}

type Meta struct {
	NextCursor int `json:"next_cursor,omitempty"`
	PerPage    int `json:"per_page"`
}

// Response wrappers

type PlayersResponse struct {
	Data []Player `json:"data"`
	Meta Meta     `json:"meta"`
}

type TournamentsResponse struct {
	Data []Tournament `json:"data"`
	Meta Meta         `json:"meta"`
}

type CoursesResponse struct {
	Data []Course `json:"data"`
	Meta Meta     `json:"meta"`
}

type TournamentResultsResponse struct {
	Data []TournamentResult `json:"data"`
	Meta Meta               `json:"meta"`
}

type TournamentCourseStatsResponse struct {
	Data []TournamentCourseStats `json:"data"`
	Meta Meta                    `json:"meta"`
}

type CourseHolesResponse struct {
	Data []CourseHole `json:"data"`
	Meta Meta         `json:"meta"`
}
