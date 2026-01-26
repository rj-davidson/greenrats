package balldontlie

const DateFormat = "2006-01-02T15:04:05.000Z"

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

type TournamentCourse struct {
	Course CourseRef `json:"course"`
	Rounds []int     `json:"rounds"`
}

type Tournament struct {
	ID         int                `json:"id"`
	Season     int                `json:"season"`
	Name       string             `json:"name"`
	StartDate  string             `json:"start_date"`
	EndDate    *string            `json:"end_date"`
	City       *string            `json:"city"`
	State      *string            `json:"state"`
	Country    *string            `json:"country"`
	CourseName *string            `json:"course_name"`
	Courses    []TournamentCourse `json:"courses"`
	Purse      *string            `json:"purse"`
	Status     *string            `json:"status"`
	Champion   *Player            `json:"champion"`
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

type TournamentResult struct {
	Tournament       TournamentRef `json:"tournament"`
	Player           Player        `json:"player"`
	Position         *string       `json:"position"`
	PositionNumeric  *int          `json:"position_numeric"`
	TotalScore       *int          `json:"total_score"`
	ParRelativeScore *int          `json:"par_relative_score"`
	Earnings         *float64      `json:"earnings"`
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

type PlayerRoundResult struct {
	Tournament       TournamentRef `json:"tournament"`
	Player           Player        `json:"player"`
	RoundNumber      int           `json:"round_number"`
	Score            *int          `json:"score"`
	ParRelativeScore *int          `json:"par_relative_score"`
}

type PlayerRoundStats struct {
	Tournament             TournamentRef `json:"tournament"`
	Player                 Player        `json:"player"`
	RoundNumber            int           `json:"round_number"`
	SGOffTee               *float64      `json:"sg_off_tee"`
	SGOffTeeRank           *int          `json:"sg_off_tee_rank"`
	SGApproach             *float64      `json:"sg_approach"`
	SGApproachRank         *int          `json:"sg_approach_rank"`
	SGAroundGreen          *float64      `json:"sg_around_green"`
	SGAroundGreenRank      *int          `json:"sg_around_green_rank"`
	SGPutting              *float64      `json:"sg_putting"`
	SGPuttingRank          *int          `json:"sg_putting_rank"`
	SGTotal                *float64      `json:"sg_total"`
	SGTotalRank            *int          `json:"sg_total_rank"`
	DrivingAccuracy        *float64      `json:"driving_accuracy"`
	DrivingAccuracyRank    *int          `json:"driving_accuracy_rank"`
	DrivingDistance        *float64      `json:"driving_distance"`
	DrivingDistanceRank    *int          `json:"driving_distance_rank"`
	LongestDrive           *int          `json:"longest_drive"`
	LongestDriveRank       *int          `json:"longest_drive_rank"`
	GreensInRegulation     *float64      `json:"greens_in_regulation"`
	GreensInRegulationRank *int          `json:"greens_in_regulation_rank"`
	SandSaves              *float64      `json:"sand_saves"`
	SandSavesRank          *int          `json:"sand_saves_rank"`
	Scrambling             *float64      `json:"scrambling"`
	ScramblingRank         *int          `json:"scrambling_rank"`
	PuttsPerGIR            *float64      `json:"putts_per_gir"`
	PuttsPerGIRRank        *int          `json:"putts_per_gir_rank"`
	Eagles                 *int          `json:"eagles"`
	Birdies                *int          `json:"birdies"`
	Pars                   *int          `json:"pars"`
	Bogeys                 *int          `json:"bogeys"`
	DoubleBogeys           *int          `json:"double_bogeys"`
}

type StatValueItem struct {
	StatName  string `json:"statName"`
	StatValue string `json:"statValue"`
}

type PlayerSeasonStat struct {
	Player       Player          `json:"player"`
	StatID       int             `json:"stat_id"`
	StatName     string          `json:"stat_name"`
	StatCategory *string         `json:"stat_category"`
	Season       int             `json:"season"`
	Rank         *int            `json:"rank"`
	StatValue    []StatValueItem `json:"stat_value"`
}

type PlayerScorecard struct {
	Tournament  TournamentRef `json:"tournament"`
	Course      *CourseRef    `json:"course"`
	Player      Player        `json:"player"`
	RoundNumber int           `json:"round_number"`
	HoleNumber  int           `json:"hole_number"`
	Par         int           `json:"par"`
	Score       *int          `json:"score"`
}

type TournamentField struct {
	ID          int        `json:"id"`
	Tournament  Tournament `json:"tournament"`
	Player      Player     `json:"player"`
	EntryStatus string     `json:"entry_status"`
	Qualifier   string     `json:"qualifier"`
	OWGR        *int       `json:"owgr"`
	IsAmateur   bool       `json:"is_amateur"`
}

type TournamentRef struct {
	ID         int                `json:"id"`
	Season     int                `json:"season"`
	Name       string             `json:"name"`
	StartDate  string             `json:"start_date"`
	EndDate    *string            `json:"end_date"`
	City       *string            `json:"city"`
	State      *string            `json:"state"`
	Country    *string            `json:"country"`
	CourseName *string            `json:"course_name"`
	Courses    []TournamentCourse `json:"courses"`
	Purse      *string            `json:"purse"`
	Status     *string            `json:"status"`
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

type PlayerRoundResultsResponse struct {
	Data []PlayerRoundResult `json:"data"`
	Meta Meta                `json:"meta"`
}

type PlayerRoundStatsResponse struct {
	Data []PlayerRoundStats `json:"data"`
	Meta Meta               `json:"meta"`
}

type PlayerSeasonStatsResponse struct {
	Data []PlayerSeasonStat `json:"data"`
	Meta Meta               `json:"meta"`
}

type PlayerScorecardsResponse struct {
	Data []PlayerScorecard `json:"data"`
	Meta Meta              `json:"meta"`
}

type TournamentFieldResponse struct {
	Data []TournamentField `json:"data"`
	Meta Meta              `json:"meta"`
}
