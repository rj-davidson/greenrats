package tournaments

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/tournament"
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
	query := s.db.Tournament.Query()

	if req.Season > 0 {
		query = query.Where(tournament.SeasonYear(req.Season))
	}

	if req.Status != "" {
		query = query.Where(tournament.StatusEQ(tournament.Status(req.Status)))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count tournaments: %w", err)
	}

	if req.Status == "upcoming" {
		query = query.Order(ent.Asc(tournament.FieldStartDate))
	} else {
		query = query.Order(ent.Desc(tournament.FieldStartDate))
	}

	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}
	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	results, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tournaments: %w", err)
	}

	tournaments := make([]Tournament, len(results))
	for i, t := range results {
		tournaments[i] = toTournament(t)
	}

	return &ListTournamentsResponse{
		Tournaments: tournaments,
		Total:       total,
	}, nil
}

// GetByID returns a tournament by its ID.
func (s *Service) GetByID(ctx context.Context, id string) (*Tournament, error) {
	uid, err := uuid.FromString(id)
	if err != nil {
		return nil, fmt.Errorf("invalid tournament ID: %w", err)
	}

	t, err := s.db.Tournament.Get(ctx, uid)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get tournament: %w", err)
	}

	result := toTournament(t)
	return &result, nil
}

// GetActive returns the currently active tournament.
func (s *Service) GetActive(ctx context.Context) (*Tournament, error) {
	t, err := s.db.Tournament.Query().
		Where(tournament.StatusEQ(tournament.StatusActive)).
		Order(ent.Asc(tournament.FieldStartDate)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get active tournament: %w", err)
	}

	result := toTournament(t)
	return &result, nil
}

func toTournament(t *ent.Tournament) Tournament {
	result := Tournament{
		ID:        t.ID.String(),
		Name:      t.Name,
		StartDate: t.StartDate,
		EndDate:   t.EndDate,
		Status:    string(t.Status),
	}

	if t.Location != nil {
		result.Venue = *t.Location
	}
	if t.Course != nil {
		result.Course = *t.Course
	}
	if t.Purse != nil {
		result.Purse = float64(*t.Purse)
	}

	return result
}
