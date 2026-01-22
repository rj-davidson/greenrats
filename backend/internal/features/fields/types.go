package fields

import "github.com/gofrs/uuid/v5"

type SyncResult struct {
	TournamentID   uuid.UUID
	TournamentName string
	TotalPlayers   int
	MatchedPlayers int
	NewEntries     int
	UpdatedEntries int
	Errors         []string
}

type FieldEntry struct {
	GolferID    uuid.UUID
	GolferName  string
	CountryCode string
	EntryStatus string
	Qualifier   *string
	OWGRAtEntry *int
	IsAmateur   bool
}
