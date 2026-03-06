package users

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	DisplayName *string   `json:"display_name"`
	IsAdmin     bool      `json:"is_admin"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SetDisplayNameRequest struct {
	DisplayName string `json:"display_name"`
}

type CheckDisplayNameResponse struct {
	Available bool   `json:"available"`
	Name      string `json:"name"`
}

type PendingPickAction struct {
	LeagueID       uuid.UUID `json:"league_id"`
	LeagueName     string    `json:"league_name"`
	TournamentID   uuid.UUID `json:"tournament_id"`
	TournamentName string    `json:"tournament_name"`
	PickDeadline   time.Time `json:"pick_deadline"`
}

type UpcomingTournament struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Status    string    `json:"status"`
}

type PendingActionsResponse struct {
	PendingPicks        []PendingPickAction  `json:"pending_picks"`
	UpcomingTournaments []UpcomingTournament `json:"upcoming_tournaments"`
}
