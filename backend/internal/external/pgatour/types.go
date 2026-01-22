package pgatour

type GraphQLRequest struct {
	Query         string         `json:"query"`
	Variables     map[string]any `json:"variables,omitempty"`
	OperationName string         `json:"operationName,omitempty"`
}

type GraphQLResponse struct {
	Data   *ResponseData  `json:"data,omitempty"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

type ResponseData struct {
	Field    *Field    `json:"field,omitempty"`
	Schedule *Schedule `json:"schedule,omitempty"`
}

type GraphQLError struct {
	Message string `json:"message"`
}

type Schedule struct {
	Completed []ScheduleMonth `json:"completed"`
	Upcoming  []ScheduleMonth `json:"upcoming"`
}

type ScheduleMonth struct {
	Month       string               `json:"month"`
	Year        string               `json:"year"`
	Tournaments []ScheduleTournament `json:"tournaments"`
}

type ScheduleTournament struct {
	ID             string `json:"id"`
	TournamentName string `json:"tournamentName"`
	StartDate      int64  `json:"startDate"`
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
