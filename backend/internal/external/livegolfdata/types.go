package livegolfdata

type ScheduleEntry struct {
	TournID   string  `json:"tournId"`
	Name      string  `json:"name"`
	StartDate *string `json:"startDate"`
	EndDate   *string `json:"endDate"`
	Course    *string `json:"course"`
	City      *string `json:"city"`
	State     *string `json:"state"`
	Country   *string `json:"country"`
	Purse     *int    `json:"purse"`
	Status    *string `json:"status"`
}

type Tournament struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Status    string `json:"status"`
	Season    int    `json:"season"`
	Course    string `json:"course,omitempty"`
	Location  string `json:"location,omitempty"`
	Purse     int    `json:"purse,omitempty"`
}

type TournamentInfo struct {
	TournID   string        `json:"tournId"`
	Name      string        `json:"name"`
	Year      int           `json:"year"`
	StartDate *string       `json:"startDate"`
	EndDate   *string       `json:"endDate"`
	Courses   []CourseInfo  `json:"courses"`
	Players   []PlayerEntry `json:"players"`
}

type CourseInfo struct {
	CourseID   string  `json:"courseId"`
	CourseName string  `json:"courseName"`
	Par        *int    `json:"par"`
	Yardage    *int    `json:"yardage"`
	City       *string `json:"city"`
	State      *string `json:"state"`
	Country    *string `json:"country"`
}

type Player struct {
	PlayerID    string  `json:"playerId"`
	FirstName   string  `json:"firstName"`
	LastName    string  `json:"lastName"`
	Country     *string `json:"country"`
	CountryCode *string `json:"countryCode"`
	Amateur     *bool   `json:"amateur"`
	Birthdate   *string `json:"birthdate"`
	TurnedPro   *int    `json:"turnedPro"`
}

type PlayerEntry struct {
	PlayerID    string  `json:"playerId"`
	FirstName   string  `json:"firstName"`
	LastName    string  `json:"lastName"`
	Status      *string `json:"status"`
	IsAlternate *bool   `json:"isAlternate"`
}

type Golfer struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	FirstName    string `json:"first_name,omitempty"`
	LastName     string `json:"last_name,omitempty"`
	Country      string `json:"country"`
	WorldRanking int    `json:"world_ranking"`
	ImageURL     string `json:"image_url,omitempty"`
}

type LeaderboardEntry struct {
	GolferID     string `json:"golfer_id"`
	GolferName   string `json:"golfer_name"`
	Position     int    `json:"position"`
	Score        int    `json:"score"`
	TotalStrokes int    `json:"total_strokes"`
	Thru         int    `json:"thru"`
	Round        int    `json:"round"`
	Status       string `json:"status"`
}

type LeaderboardRow struct {
	PlayerID     string  `json:"playerId"`
	FirstName    string  `json:"firstName"`
	LastName     string  `json:"lastName"`
	Position     *int    `json:"position"`
	Total        *int    `json:"total"`
	TotalStrokes *int    `json:"totalStrokes"`
	Thru         *int    `json:"thru"`
	Round        *int    `json:"round"`
	Status       *string `json:"status"`
	Rounds       []int   `json:"rounds"`
}

type EarningsEntry struct {
	PlayerID  string `json:"playerId"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Earnings  int    `json:"earnings"`
}

type EarningsRow struct {
	PlayerID  string `json:"playerId"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Earnings  int    `json:"earnings"`
	Position  *int   `json:"position"`
}

type WorldRankingEntry struct {
	PlayerID    string   `json:"playerId"`
	FirstName   string   `json:"firstName"`
	LastName    string   `json:"lastName"`
	Rank        int      `json:"rank"`
	PrevRank    *int     `json:"prevRank"`
	AvgPoints   *float64 `json:"avgPoints"`
	TotalPoints *float64 `json:"totalPoints"`
	Events      *int     `json:"events"`
}

type FedExCupEntry struct {
	PlayerID  string   `json:"playerId"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	Rank      int      `json:"rank"`
	PrevRank  *int     `json:"prevRank"`
	Points    *float64 `json:"points"`
	Events    *int     `json:"events"`
}

type PointsEntry struct {
	PlayerID  string   `json:"playerId"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	Points    *float64 `json:"points"`
	Position  *int     `json:"position"`
}

type ScorecardHole struct {
	HoleNumber int  `json:"holeNumber"`
	Par        int  `json:"par"`
	Yardage    *int `json:"yardage"`
	Score      *int `json:"score"`
	ToPar      *int `json:"toPar"`
}

type Scorecard struct {
	PlayerID  string          `json:"playerId"`
	FirstName string          `json:"firstName"`
	LastName  string          `json:"lastName"`
	Round     int             `json:"round"`
	Holes     []ScorecardHole `json:"holes"`
	Total     *int            `json:"total"`
	ToPar     *int            `json:"toPar"`
}

type Organization struct {
	OrgID   string `json:"orgId"`
	OrgName string `json:"orgName"`
}

type ScheduleResponse struct {
	Schedule []ScheduleEntry `json:"schedule"`
}

type TournamentResponse struct {
	Tournament TournamentInfo `json:"tournament"`
}

type PlayersResponse struct {
	Players []Player `json:"players"`
}

type LeaderboardResponse struct {
	Leaderboard []LeaderboardRow `json:"leaderboard"`
}

type EarningsResponse struct {
	Leaderboard []EarningsEntry `json:"leaderboard"`
}

type WorldRankingResponse struct {
	Rankings []WorldRankingEntry `json:"rankings"`
}

type FedExCupResponse struct {
	Rankings []FedExCupEntry `json:"rankings"`
}

type PointsResponse struct {
	Points []PointsEntry `json:"points"`
}

type ScorecardResponse struct {
	Scorecard Scorecard `json:"scorecard"`
}

type OrganizationsResponse struct {
	Organizations []Organization `json:"organizations"`
}
