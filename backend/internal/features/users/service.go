package users

import (
	"context"
	"fmt"
	"log"

	"github.com/rj-davidson/greenrats/ent"
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
	WorkOSID    string
	Email       string
	DisplayName string
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

	// User not found, create a new one
	newUser, err := s.db.User.
		Create().
		SetWorkosID(params.WorkOSID).
		SetEmail(params.Email).
		SetDisplayName(params.DisplayName).
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
	// TODO: Implement
	return nil, nil
}

// GetByWorkOSID returns a user by their WorkOS ID.
func (s *Service) GetByWorkOSID(ctx context.Context, workosID string) (*ent.User, error) {
	return s.db.User.
		Query().
		Where(user.WorkosID(workosID)).
		Only(ctx)
}
