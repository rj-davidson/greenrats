package leagues

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/league"
	"github.com/rj-davidson/greenrats/ent/leaguemembership"
	"github.com/rj-davidson/greenrats/ent/user"
)

const (
	joinCodeLength  = 6
	joinCodeCharset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Excludes I, O, 0, 1 for readability
)

// Service handles league business logic.
type Service struct {
	db *ent.Client
}

// NewService creates a new league service.
func NewService(db *ent.Client) *Service {
	return &Service{db: db}
}

// CreateParams contains the parameters for creating a league.
type CreateParams struct {
	Name   string
	UserID uuid.UUID
}

// Create creates a new league with the given name and adds the creator as owner.
func (s *Service) Create(ctx context.Context, params CreateParams) (*League, error) {
	// Generate a unique join code
	code, err := s.generateUniqueJoinCode(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate join code: %w", err)
	}

	// Use current year as season year
	seasonYear := time.Now().Year()

	// Start a transaction
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Create the league
	entLeague, err := tx.League.
		Create().
		SetName(params.Name).
		SetCode(code).
		SetSeasonYear(seasonYear).
		SetCreatedByID(params.UserID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create league: %w", err)
	}

	// Create the membership with owner role
	_, err = tx.LeagueMembership.
		Create().
		SetUserID(params.UserID).
		SetLeagueID(entLeague.ID).
		SetRole(leaguemembership.RoleOwner).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create membership: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &League{
		ID:         entLeague.ID,
		Name:       entLeague.Name,
		Code:       entLeague.Code,
		SeasonYear: entLeague.SeasonYear,
		CreatedAt:  entLeague.CreatedAt,
		Role:       string(leaguemembership.RoleOwner),
	}, nil
}

// ListUserLeagues returns all leagues a user belongs to with their role.
func (s *Service) ListUserLeagues(ctx context.Context, userID uuid.UUID) (*ListUserLeaguesResponse, error) {
	// Get all memberships for this user with their leagues
	memberships, err := s.db.LeagueMembership.
		Query().
		Where(leaguemembership.HasUserWith(user.IDEQ(userID))).
		WithLeague().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query leagues: %w", err)
	}

	leagues := make([]League, 0, len(memberships))
	for _, m := range memberships {
		if m.Edges.League == nil {
			continue
		}
		l := m.Edges.League
		leagues = append(leagues, League{
			ID:         l.ID,
			Name:       l.Name,
			Code:       l.Code,
			SeasonYear: l.SeasonYear,
			CreatedAt:  l.CreatedAt,
			Role:       string(m.Role),
		})
	}

	return &ListUserLeaguesResponse{
		Leagues: leagues,
		Total:   len(leagues),
	}, nil
}

// GetByID returns a league by its ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*League, error) {
	entLeague, err := s.db.League.
		Query().
		Where(league.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get league: %w", err)
	}

	return &League{
		ID:         entLeague.ID,
		Name:       entLeague.Name,
		Code:       entLeague.Code,
		SeasonYear: entLeague.SeasonYear,
		CreatedAt:  entLeague.CreatedAt,
	}, nil
}

// GetByIDWithRole returns a league by its ID with the user's role if they are a member.
func (s *Service) GetByIDWithRole(ctx context.Context, id, userID uuid.UUID) (*League, error) {
	entLeague, err := s.db.League.
		Query().
		Where(league.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get league: %w", err)
	}

	l := &League{
		ID:         entLeague.ID,
		Name:       entLeague.Name,
		Code:       entLeague.Code,
		SeasonYear: entLeague.SeasonYear,
		CreatedAt:  entLeague.CreatedAt,
	}

	// Get the user's membership to determine their role
	membership, err := s.db.LeagueMembership.
		Query().
		Where(
			leaguemembership.HasLeagueWith(league.IDEQ(id)),
			leaguemembership.HasUserWith(user.IDEQ(userID)),
		).
		Only(ctx)
	if err == nil {
		l.Role = string(membership.Role)
	}

	return l, nil
}

// generateUniqueJoinCode generates a unique 6-character join code.
func (s *Service) generateUniqueJoinCode(ctx context.Context) (string, error) {
	const maxAttempts = 10
	for i := range maxAttempts {
		code, err := generateJoinCode()
		if err != nil {
			return "", err
		}

		// Check if code already exists
		exists, err := s.db.League.
			Query().
			Where(league.CodeEQ(code)).
			Exist(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to check code uniqueness: %w", err)
		}

		if !exists {
			return code, nil
		}

		// If we're on the last attempt, log a warning
		if i == maxAttempts-1 {
			return "", fmt.Errorf("failed to generate unique code after %d attempts", maxAttempts)
		}
	}

	return "", fmt.Errorf("failed to generate unique code")
}

// generateJoinCode generates a random 6-character alphanumeric code.
func generateJoinCode() (string, error) {
	code := make([]byte, joinCodeLength)
	charsetLen := big.NewInt(int64(len(joinCodeCharset)))

	for i := range code {
		idx, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate random index: %w", err)
		}
		code[i] = joinCodeCharset[idx.Int64()]
	}

	return string(code), nil
}
