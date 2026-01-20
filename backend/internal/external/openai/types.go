package openai

type GolferInput struct {
	GolferID  string `json:"golfer_id"`
	Name      string `json:"name"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

type EarningsResult struct {
	GolferID string `json:"golfer_id" jsonschema_description:"The golfer_id from the input list that matches this result"`
	Position int    `json:"position" jsonschema_description:"Final finishing position in the tournament (1 for winner, 2 for second, etc)"`
	Earnings int    `json:"earnings" jsonschema_description:"Prize money earned in USD"`
}

type EarningsResponse struct {
	Results []EarningsResult `json:"results" jsonschema_description:"List of tournament results matched to golfer_id values from the input"`
}
