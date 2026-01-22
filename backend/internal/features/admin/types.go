package admin

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type AdminUser struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	DisplayName *string   `json:"display_name"`
	IsAdmin     bool      `json:"is_admin"`
	CreatedAt   time.Time `json:"created_at"`
}

type ListUsersResponse struct {
	Users []AdminUser `json:"users"`
	Total int         `json:"total"`
}

type AdminLeague struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	SeasonYear  int       `json:"season_year"`
	MemberCount int       `json:"member_count"`
	CreatedAt   time.Time `json:"created_at"`
}

type ListLeaguesResponse struct {
	Leagues []AdminLeague `json:"leagues"`
	Total   int           `json:"total"`
}

type AdminTournament struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

type ListTournamentsResponse struct {
	Tournaments []AdminTournament `json:"tournaments"`
	Total       int               `json:"total"`
}

type TriggerResponse struct {
	Message string `json:"message"`
}

type FieldEntryResponse struct {
	ID          uuid.UUID `json:"id"`
	GolferID    uuid.UUID `json:"golfer_id"`
	GolferName  string    `json:"golfer_name"`
	CountryCode string    `json:"country_code"`
	EntryStatus string    `json:"entry_status"`
	Qualifier   *string   `json:"qualifier"`
	OWGRAtEntry *int      `json:"owgr_at_entry"`
	IsAmateur   bool      `json:"is_amateur"`
}

type ListFieldResponse struct {
	Entries []FieldEntryResponse `json:"entries"`
	Total   int                  `json:"total"`
}

type AddFieldEntryRequest struct {
	GolferID    uuid.UUID `json:"golfer_id" validate:"required"`
	EntryStatus string    `json:"entry_status"`
	Qualifier   *string   `json:"qualifier"`
}

type UpdateFieldEntryRequest struct {
	EntryStatus *string `json:"entry_status"`
	Qualifier   *string `json:"qualifier"`
	OWGRAtEntry *int    `json:"owgr_at_entry"`
	IsAmateur   *bool   `json:"is_amateur"`
}
