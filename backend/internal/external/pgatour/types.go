package pgatour

type GraphQLRequest struct {
	Query         string         `json:"query"`
	Variables     map[string]any `json:"variables,omitempty"`
	OperationName string         `json:"operationName,omitempty"`
}

type GraphQLResponse struct {
	Data   *FieldData     `json:"data,omitempty"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

type GraphQLError struct {
	Message string `json:"message"`
}

type FieldData struct {
	Field *Field `json:"field,omitempty"`
}

type Field struct {
	TournamentName string        `json:"tournamentName"`
	ID             string        `json:"id"`
	Players        []FieldPlayer `json:"players"`
}

type FieldPlayer struct {
	ID          string `json:"id"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	DisplayName string `json:"displayName"`
	Country     string `json:"country"`
	CountryCode string `json:"countryFlag"`
	IsAmateur   bool   `json:"amateur"`
}

type FieldEntry struct {
	PGATourID   string
	FirstName   string
	LastName    string
	DisplayName string
	CountryCode string
	IsAmateur   bool
}
