package tournaments

import (
	"context"

	"github.com/rj-davidson/greenrats/ent"
)

// Service handles tournament business logic.
type Service struct {
	db *ent.Client
}

// NewService creates a new tournament service.
func NewService(db *ent.Client) *Service {
	return &Service{db: db}
}

// List returns a list of tournaments with optional filtering.
func (s *Service) List(ctx context.Context, req ListTournamentsRequest) (*ListTournamentsResponse, error) {
	// TODO: Implement tournament listing with Ent queries
	return &ListTournamentsResponse{
		Tournaments: []Tournament{},
		Total:       0,
	}, nil
}

// GetByID returns a tournament by its ID.
func (s *Service) GetByID(ctx context.Context, id string) (*Tournament, error) {
	// TODO: Implement tournament retrieval with Ent
	return nil, nil
}

// GetActive returns the currently active tournament.
func (s *Service) GetActive(ctx context.Context) (*Tournament, error) {
	// TODO: Implement active tournament retrieval
	return nil, nil
}
