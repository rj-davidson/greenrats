package admin

import (
	"context"
	"errors"
	"fmt"

	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/league"
	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/ent/user"
)

var ErrLeagueNotFound = errors.New("league not found")

type Service struct {
	db *ent.Client
}

func NewService(db *ent.Client) *Service {
	return &Service{db: db}
}

func (s *Service) ListUsers(ctx context.Context) (*ListUsersResponse, error) {
	users, err := s.db.User.Query().
		Order(ent.Desc(user.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}

	result := make([]AdminUser, len(users))
	for i, u := range users {
		result[i] = AdminUser{
			ID:          u.ID,
			Email:       u.Email,
			DisplayName: u.DisplayName,
			IsAdmin:     u.IsAdmin,
			CreatedAt:   u.CreatedAt,
		}
	}

	return &ListUsersResponse{
		Users: result,
		Total: len(result),
	}, nil
}

func (s *Service) ListLeagues(ctx context.Context) (*ListLeaguesResponse, error) {
	leagues, err := s.db.League.Query().
		Order(ent.Desc(league.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query leagues: %w", err)
	}

	result := make([]AdminLeague, len(leagues))
	for i, l := range leagues {
		memberCount, err := s.db.League.QueryMemberships(l).Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to count members for league %s: %w", l.ID, err)
		}

		result[i] = AdminLeague{
			ID:          l.ID,
			Name:        l.Name,
			SeasonYear:  l.SeasonYear,
			MemberCount: memberCount,
			CreatedAt:   l.CreatedAt,
		}
	}

	return &ListLeaguesResponse{
		Leagues: result,
		Total:   len(result),
	}, nil
}

func (s *Service) DeleteLeague(ctx context.Context, id uuid.UUID) error {
	err := s.db.League.DeleteOneID(id).Exec(ctx)
	if ent.IsNotFound(err) {
		return ErrLeagueNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to delete league: %w", err)
	}
	return nil
}

func (s *Service) ListTournaments(ctx context.Context) (*ListTournamentsResponse, error) {
	tournaments, err := s.db.Tournament.Query().
		Order(ent.Desc(tournament.FieldStartDate)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query tournaments: %w", err)
	}

	result := make([]AdminTournament, len(tournaments))
	for i, t := range tournaments {
		result[i] = AdminTournament{
			ID:        t.ID,
			Name:      t.Name,
			Status:    string(t.Status),
			StartDate: t.StartDate,
			EndDate:   t.EndDate,
		}
	}

	return &ListTournamentsResponse{
		Tournaments: result,
		Total:       len(result),
	}, nil
}
