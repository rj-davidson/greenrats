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
