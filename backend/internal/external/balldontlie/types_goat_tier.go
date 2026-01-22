package balldontlie

// GOAT tier ($39.99/mo) - NOT CURRENTLY SUBSCRIBED
// These types are defined for future use if we upgrade.
// Endpoints: player_round_results, player_round_stats, player_season_stats, player_scorecards

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
	Player      Player        `json:"player"`
	RoundNumber int           `json:"round_number"`
	HoleNumber  int           `json:"hole_number"`
	Par         int           `json:"par"`
	Score       *int          `json:"score"`
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

type TournamentField struct {
	ID          int        `json:"id"`
	Tournament  Tournament `json:"tournament"`
	Player      Player     `json:"player"`
	EntryStatus string     `json:"entry_status"`
	Qualifier   string     `json:"qualifier"`
	OWGR        *int       `json:"owgr"`
	IsAmateur   bool       `json:"is_amateur"`
}

type TournamentFieldResponse struct {
	Data []TournamentField `json:"data"`
	Meta Meta              `json:"meta"`
}
