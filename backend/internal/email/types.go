package email

type WelcomeData struct {
	DisplayName string
}

type LeagueJoinData struct {
	DisplayName   string
	LeagueName    string
	NewMemberName string
	IsNewMember   bool
}

type PickReminderData struct {
	DisplayName    string
	TournamentName string
	LeagueName     string
	Deadline       string
}

type TournamentResultsData struct {
	DisplayName      string
	TournamentName   string
	TournamentWinner string
	LeagueName       string
	GolferName       string
	GolferPosition   string
	GolferEarnings   string
	UserRank         int
	TotalEarnings    string
}

type CommissionerActionData struct {
	DisplayName       string
	LeagueName        string
	CommissionerName  string
	ActionDescription string
}
