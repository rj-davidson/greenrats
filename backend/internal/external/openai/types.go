package openai

type GolferInput struct {
	GolferID  string `json:"golfer_id"`
	Name      string `json:"name"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

type EarningsResult struct {
	GolferID string `json:"golfer_id" jsonschema_description:"The golfer_id from the input list that matches this result"`
	Earnings int    `json:"earnings" jsonschema_description:"Prize money earned in USD"`
}

type EarningsResponse struct {
	Results []EarningsResult `json:"results" jsonschema_description:"List of tournament results matched to golfer_id values from the input"`
}

type LeaderboardEntry struct {
	Name     string `json:"name"`
	Earnings int    `json:"earnings"`
}

type LeaderboardResponse struct {
	TournamentName string             `json:"tournament_name"`
	Entries        []LeaderboardEntry `json:"entries"`
}

type DuplicateEarningsCandidate struct {
	Earnings int    `json:"earnings"`
	Context  string `json:"context"`
}

type DuplicateEarningsInput struct {
	GolferName string                       `json:"golfer_name"`
	Candidates []DuplicateEarningsCandidate `json:"candidates"`
}

type DuplicateEarningsResult struct {
	GolferName string `json:"golfer_name" jsonschema_description:"The golfer name from the input"`
	Earnings   int    `json:"earnings" jsonschema_description:"The correct prize money earned in USD"`
}

type DuplicateEarningsResponse struct {
	Results []DuplicateEarningsResult `json:"results" jsonschema_description:"Resolved earnings for each golfer"`
}
