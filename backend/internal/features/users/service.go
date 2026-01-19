package users

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/league"
	"github.com/rj-davidson/greenrats/ent/leaguemembership"
	"github.com/rj-davidson/greenrats/ent/pick"
	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/ent/user"
)

// Service handles user business logic.
type Service struct {
	db *ent.Client
}

// NewService creates a new user service.
func NewService(db *ent.Client) *Service {
	return &Service{db: db}
}

// GetOrCreateParams contains the parameters for GetOrCreate.
type GetOrCreateParams struct {
	WorkOSID string
	Email    string
}

// GetOrCreateResult contains the result of GetOrCreate.
type GetOrCreateResult struct {
	User    *ent.User
	Created bool
}

// GetOrCreate finds an existing user by WorkOS ID or creates a new one.
// Handles race conditions by retrying fetch on constraint error.
func (s *Service) GetOrCreate(ctx context.Context, params GetOrCreateParams) (*GetOrCreateResult, error) {
	// First, try to find existing user by WorkOS ID
	existingUser, err := s.db.User.
		Query().
		Where(user.WorkosID(params.WorkOSID)).
		Only(ctx)

	if err == nil {
		return &GetOrCreateResult{User: existingUser, Created: false}, nil
	}

	// If error is not "not found", return it
	if !ent.IsNotFound(err) {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	// User not found, create a new one (display_name will be null until set during onboarding)
	newUser, err := s.db.User.
		Create().
		SetWorkosID(params.WorkOSID).
		SetEmail(params.Email).
		Save(ctx)

	if err == nil {
		log.Printf("Created new user: workos_id=%s, email=%s", params.WorkOSID, params.Email)
		return &GetOrCreateResult{User: newUser, Created: true}, nil
	}

	// Check if this is a constraint error (race condition - user was created by another request)
	if ent.IsConstraintError(err) {
		// Retry fetch
		existingUser, retryErr := s.db.User.
			Query().
			Where(user.WorkosID(params.WorkOSID)).
			Only(ctx)

		if retryErr != nil {
			return nil, fmt.Errorf("failed to fetch user after constraint error: %w", retryErr)
		}

		return &GetOrCreateResult{User: existingUser, Created: false}, nil
	}

	return nil, fmt.Errorf("failed to create user: %w", err)
}

// GetByID returns a user by their database ID.
func (s *Service) GetByID(ctx context.Context, id string) (*ent.User, error) {
	uid, err := uuid.FromString(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	u, err := s.db.User.Get(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return u, nil
}

// GetByWorkOSID returns a user by their WorkOS ID.
func (s *Service) GetByWorkOSID(ctx context.Context, workosID string) (*ent.User, error) {
	return s.db.User.
		Query().
		Where(user.WorkosID(workosID)).
		Only(ctx)
}

// SetDisplayName sets the display name for a user.
// Returns an error if the user already has a display name set.
func (s *Service) SetDisplayName(ctx context.Context, userID, displayName string) (*ent.User, error) {
	// Parse the user ID
	id, err := uuid.FromString(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Get the user first to check if display_name is already set
	u, err := s.db.User.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if display name is already set
	if u.DisplayName != nil {
		return nil, fmt.Errorf("display name is already set and cannot be changed")
	}

	// Update the display name
	updated, err := u.Update().
		SetDisplayName(displayName).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, fmt.Errorf("display name is already taken")
		}
		return nil, fmt.Errorf("failed to set display name: %w", err)
	}

	return updated, nil
}

func (s *Service) IsDisplayNameAvailable(ctx context.Context, displayName string) (bool, error) {
	exists, err := s.db.User.
		Query().
		Where(user.DisplayNameEqualFold(displayName)).
		Exist(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check display name: %w", err)
	}
	return !exists, nil
}

func (s *Service) GetPendingActions(ctx context.Context, userID uuid.UUID) (*PendingActionsResponse, error) {
	now := time.Now().UTC()
	pickWindowDays := 3

	memberships, err := s.db.LeagueMembership.
		Query().
		Where(leaguemembership.HasUserWith(user.IDEQ(userID))).
		WithLeague().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get memberships: %w", err)
	}

	upcomingTournaments, err := s.db.Tournament.
		Query().
		Where(tournament.StatusEQ(tournament.StatusUpcoming)).
		Order(tournament.ByStartDate()).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming tournaments: %w", err)
	}

	pendingPicks := make([]PendingPickAction, 0)

	for _, m := range memberships {
		if m.Edges.League == nil {
			continue
		}
		l := m.Edges.League

		for _, t := range upcomingTournaments {
			if t.SeasonYear != l.SeasonYear {
				continue
			}

			windowOpens := t.StartDate.AddDate(0, 0, -pickWindowDays)
			if now.Before(windowOpens) || now.After(t.StartDate) {
				continue
			}

			hasPick, err := s.db.Pick.
				Query().
				Where(
					pick.HasUserWith(user.IDEQ(userID)),
					pick.HasLeagueWith(league.IDEQ(l.ID)),
					pick.HasTournamentWith(tournament.IDEQ(t.ID)),
				).
				Exist(ctx)
			if err != nil {
				continue
			}
			if hasPick {
				continue
			}

			pendingPicks = append(pendingPicks, PendingPickAction{
				LeagueID:       l.ID,
				LeagueName:     l.Name,
				TournamentID:   t.ID,
				TournamentName: t.Name,
				PickDeadline:   t.StartDate,
			})
		}
	}

	upcoming := make([]UpcomingTournament, 0, len(upcomingTournaments))
	for _, t := range upcomingTournaments {
		daysUntil := t.StartDate.Sub(now).Hours() / 24
		if daysUntil <= 7 {
			upcoming = append(upcoming, UpcomingTournament{
				ID:        t.ID,
				Name:      t.Name,
				StartDate: t.StartDate,
				EndDate:   t.EndDate,
				Status:    string(t.Status),
			})
		}
	}

	return &PendingActionsResponse{
		PendingPicks:        pendingPicks,
		UpcomingTournaments: upcoming,
	}, nil
}
