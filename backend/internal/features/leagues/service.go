package leagues

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/commissioneraction"
	"github.com/rj-davidson/greenrats/ent/league"
	"github.com/rj-davidson/greenrats/ent/leaguemembership"
	"github.com/rj-davidson/greenrats/ent/pick"
	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/ent/user"
)

const (
	joinCodeLength  = 6
	joinCodeCharset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	maxMembers      = 200
)

var (
	ErrLeagueNotFound  = errors.New("league not found")
	ErrInvalidJoinCode = errors.New("invalid join code")
	ErrAlreadyMember   = errors.New("already a member of this league")
	ErrJoiningDisabled = errors.New("joining is disabled for this league")
	ErrLeagueFull      = errors.New("league has reached maximum members")
	ErrNotCommissioner = errors.New("only the commissioner can perform this action")
	ErrUserNotFound    = errors.New("user not found")
)

type Service struct {
	db *ent.Client
}

func NewService(db *ent.Client) *Service {
	return &Service{db: db}
}

type CreateParams struct {
	Name   string
	UserID uuid.UUID
}

func (s *Service) Create(ctx context.Context, params CreateParams) (*League, error) {
	code, err := s.generateUniqueJoinCode(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate join code: %w", err)
	}

	seasonYear := time.Now().Year()

	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

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

	_, err = tx.LeagueMembership.
		Create().
		SetUserID(params.UserID).
		SetLeagueID(entLeague.ID).
		SetRole(leaguemembership.RoleOwner).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create membership: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &League{
		ID:             entLeague.ID,
		Name:           entLeague.Name,
		Code:           entLeague.Code,
		SeasonYear:     entLeague.SeasonYear,
		JoiningEnabled: entLeague.JoiningEnabled,
		CreatedAt:      entLeague.CreatedAt,
		Role:           string(leaguemembership.RoleOwner),
		MemberCount:    1,
	}, nil
}

func (s *Service) ListUserLeagues(ctx context.Context, userID uuid.UUID) (*ListUserLeaguesResponse, error) {
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
			ID:             l.ID,
			Name:           l.Name,
			Code:           l.Code,
			SeasonYear:     l.SeasonYear,
			JoiningEnabled: l.JoiningEnabled,
			CreatedAt:      l.CreatedAt,
			Role:           string(m.Role),
		})
	}

	return &ListUserLeaguesResponse{
		Leagues: leagues,
		Total:   len(leagues),
	}, nil
}

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

	memberCount, _ := s.db.LeagueMembership.
		Query().
		Where(leaguemembership.HasLeagueWith(league.IDEQ(id))).
		Count(ctx)

	return &League{
		ID:             entLeague.ID,
		Name:           entLeague.Name,
		Code:           entLeague.Code,
		SeasonYear:     entLeague.SeasonYear,
		JoiningEnabled: entLeague.JoiningEnabled,
		CreatedAt:      entLeague.CreatedAt,
		MemberCount:    memberCount,
	}, nil
}

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

	memberCount, _ := s.db.LeagueMembership.
		Query().
		Where(leaguemembership.HasLeagueWith(league.IDEQ(id))).
		Count(ctx)

	l := &League{
		ID:             entLeague.ID,
		Name:           entLeague.Name,
		Code:           entLeague.Code,
		SeasonYear:     entLeague.SeasonYear,
		JoiningEnabled: entLeague.JoiningEnabled,
		CreatedAt:      entLeague.CreatedAt,
		MemberCount:    memberCount,
	}

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

func (s *Service) JoinLeague(ctx context.Context, userID uuid.UUID, code string) (*League, error) {
	entLeague, err := s.db.League.
		Query().
		Where(league.CodeEQ(code)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrInvalidJoinCode
		}
		return nil, fmt.Errorf("failed to find league: %w", err)
	}

	if !entLeague.JoiningEnabled {
		return nil, ErrJoiningDisabled
	}

	alreadyMember, err := s.db.LeagueMembership.
		Query().
		Where(
			leaguemembership.HasLeagueWith(league.IDEQ(entLeague.ID)),
			leaguemembership.HasUserWith(user.IDEQ(userID)),
		).
		Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if alreadyMember {
		return nil, ErrAlreadyMember
	}

	memberCount, err := s.db.LeagueMembership.
		Query().
		Where(leaguemembership.HasLeagueWith(league.IDEQ(entLeague.ID))).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count members: %w", err)
	}
	if memberCount >= maxMembers {
		return nil, ErrLeagueFull
	}

	_, err = s.db.LeagueMembership.
		Create().
		SetUserID(userID).
		SetLeagueID(entLeague.ID).
		SetRole(leaguemembership.RoleMember).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create membership: %w", err)
	}

	return &League{
		ID:             entLeague.ID,
		Name:           entLeague.Name,
		Code:           entLeague.Code,
		SeasonYear:     entLeague.SeasonYear,
		JoiningEnabled: entLeague.JoiningEnabled,
		CreatedAt:      entLeague.CreatedAt,
		Role:           string(leaguemembership.RoleMember),
		MemberCount:    memberCount + 1,
	}, nil
}

func (s *Service) RegenerateJoinCode(ctx context.Context, leagueID, commissionerID uuid.UUID) (*League, error) {
	isOwner, err := s.isLeagueOwner(ctx, leagueID, commissionerID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, ErrNotCommissioner
	}

	newCode, err := s.generateUniqueJoinCode(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new code: %w", err)
	}

	entLeague, err := s.db.League.
		UpdateOneID(leagueID).
		SetCode(newCode).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update league: %w", err)
	}

	_, err = s.db.CommissionerAction.
		Create().
		SetActionType(commissioneraction.ActionTypeJoinCodeReset).
		SetDescription("Join code regenerated").
		SetLeagueID(leagueID).
		SetCommissionerID(commissionerID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to log action: %w", err)
	}

	memberCount, _ := s.db.LeagueMembership.
		Query().
		Where(leaguemembership.HasLeagueWith(league.IDEQ(leagueID))).
		Count(ctx)

	return &League{
		ID:             entLeague.ID,
		Name:           entLeague.Name,
		Code:           entLeague.Code,
		SeasonYear:     entLeague.SeasonYear,
		JoiningEnabled: entLeague.JoiningEnabled,
		CreatedAt:      entLeague.CreatedAt,
		Role:           string(leaguemembership.RoleOwner),
		MemberCount:    memberCount,
	}, nil
}

func (s *Service) SetJoiningEnabled(ctx context.Context, leagueID, commissionerID uuid.UUID, enabled bool) (*League, error) {
	isOwner, err := s.isLeagueOwner(ctx, leagueID, commissionerID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, ErrNotCommissioner
	}

	entLeague, err := s.db.League.
		UpdateOneID(leagueID).
		SetJoiningEnabled(enabled).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrLeagueNotFound
		}
		return nil, fmt.Errorf("failed to update league: %w", err)
	}

	actionType := commissioneraction.ActionTypeJoiningEnabled
	description := "Joining enabled"
	if !enabled {
		actionType = commissioneraction.ActionTypeJoiningDisabled
		description = "Joining disabled"
	}

	_, err = s.db.CommissionerAction.
		Create().
		SetActionType(actionType).
		SetDescription(description).
		SetLeagueID(leagueID).
		SetCommissionerID(commissionerID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to log action: %w", err)
	}

	memberCount, _ := s.db.LeagueMembership.
		Query().
		Where(leaguemembership.HasLeagueWith(league.IDEQ(leagueID))).
		Count(ctx)

	return &League{
		ID:             entLeague.ID,
		Name:           entLeague.Name,
		Code:           entLeague.Code,
		SeasonYear:     entLeague.SeasonYear,
		JoiningEnabled: entLeague.JoiningEnabled,
		CreatedAt:      entLeague.CreatedAt,
		Role:           string(leaguemembership.RoleOwner),
		MemberCount:    memberCount,
	}, nil
}

func (s *Service) isLeagueOwner(ctx context.Context, leagueID, userID uuid.UUID) (bool, error) {
	membership, err := s.db.LeagueMembership.
		Query().
		Where(
			leaguemembership.HasLeagueWith(league.IDEQ(leagueID)),
			leaguemembership.HasUserWith(user.IDEQ(userID)),
			leaguemembership.RoleEQ(leaguemembership.RoleOwner),
		).
		Exist(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check ownership: %w", err)
	}
	return membership, nil
}

func (s *Service) generateUniqueJoinCode(ctx context.Context) (string, error) {
	const maxAttempts = 10
	for i := range maxAttempts {
		code, err := generateJoinCode()
		if err != nil {
			return "", err
		}

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

		if i == maxAttempts-1 {
			return "", fmt.Errorf("failed to generate unique code after %d attempts", maxAttempts)
		}
	}

	return "", fmt.Errorf("failed to generate unique code")
}

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

func (s *Service) GetLeagueTournaments(ctx context.Context, leagueID, userID uuid.UUID) (*ListLeagueTournamentsResponse, error) {
	entLeague, err := s.db.League.
		Query().
		Where(league.IDEQ(leagueID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrLeagueNotFound
		}
		return nil, fmt.Errorf("failed to get league: %w", err)
	}

	tournaments, err := s.db.Tournament.
		Query().
		Where(tournament.SeasonYearEQ(entLeague.SeasonYear)).
		Order(tournament.ByStartDate()).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query tournaments: %w", err)
	}

	result := make([]LeagueTournament, 0, len(tournaments))
	for _, t := range tournaments {
		lt := LeagueTournament{
			ID:        t.ID,
			Name:      t.Name,
			StartDate: t.StartDate,
			EndDate:   t.EndDate,
			Status:    string(t.Status),
		}

		pickCount, _ := s.db.Pick.
			Query().
			Where(
				pick.HasLeagueWith(league.IDEQ(leagueID)),
				pick.HasTournamentWith(tournament.IDEQ(t.ID)),
			).
			Count(ctx)
		lt.PickCount = pickCount

		userPick, err := s.db.Pick.
			Query().
			Where(
				pick.HasLeagueWith(league.IDEQ(leagueID)),
				pick.HasTournamentWith(tournament.IDEQ(t.ID)),
				pick.HasUserWith(user.IDEQ(userID)),
			).
			WithGolfer().
			Only(ctx)
		if err == nil {
			lt.HasUserPick = true
			lt.UserPickID = userPick.ID
			if userPick.Edges.Golfer != nil {
				lt.GolferName = userPick.Edges.Golfer.Name
			}
		}

		result = append(result, lt)
	}

	return &ListLeagueTournamentsResponse{
		Tournaments: result,
		Total:       len(result),
	}, nil
}
