package picks

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/golfer"
	"github.com/rj-davidson/greenrats/ent/league"
	"github.com/rj-davidson/greenrats/ent/leaguemembership"
	"github.com/rj-davidson/greenrats/ent/pick"
	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/ent/tournamententry"
	"github.com/rj-davidson/greenrats/ent/user"
)

const pickWindowDays = 3

var (
	ErrTournamentNotFound    = errors.New("tournament not found")
	ErrGolferNotFound        = errors.New("golfer not found")
	ErrLeagueNotFound        = errors.New("league not found")
	ErrNotLeagueMember       = errors.New("user is not a member of this league")
	ErrPickWindowClosed      = errors.New("pick window is closed")
	ErrGolferNotInField      = errors.New("golfer is not in tournament field")
	ErrGolferAlreadyUsed     = errors.New("golfer already used this season")
	ErrPickAlreadyExists     = errors.New("pick already exists for this tournament")
	ErrTournamentNotUpcoming = errors.New("tournament is not upcoming")
)

type Service struct {
	db *ent.Client
}

func NewService(db *ent.Client) *Service {
	return &Service{db: db}
}

type CreateParams struct {
	UserID       uuid.UUID
	TournamentID uuid.UUID
	GolferID     uuid.UUID
	LeagueID     uuid.UUID
}

func (s *Service) Create(ctx context.Context, params CreateParams) (*Pick, error) {
	tournamentEnt, err := s.db.Tournament.Query().
		Where(tournament.IDEQ(params.TournamentID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrTournamentNotFound
		}
		return nil, fmt.Errorf("failed to get tournament: %w", err)
	}

	if tournamentEnt.Status != tournament.StatusUpcoming {
		return nil, ErrTournamentNotUpcoming
	}

	windowStatus := s.getPickWindowStatus(tournamentEnt)
	if !windowStatus.IsOpen {
		return nil, ErrPickWindowClosed
	}

	golferEnt, err := s.db.Golfer.Query().
		Where(golfer.IDEQ(params.GolferID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrGolferNotFound
		}
		return nil, fmt.Errorf("failed to get golfer: %w", err)
	}

	leagueExists, err := s.db.League.Query().
		Where(league.IDEQ(params.LeagueID)).
		Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check league: %w", err)
	}
	if !leagueExists {
		return nil, ErrLeagueNotFound
	}

	isMember, err := s.db.LeagueMembership.Query().
		Where(
			leaguemembership.HasUserWith(user.IDEQ(params.UserID)),
			leaguemembership.HasLeagueWith(league.IDEQ(params.LeagueID)),
		).
		Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if !isMember {
		return nil, ErrNotLeagueMember
	}

	inField, err := s.db.TournamentEntry.Query().
		Where(
			tournamententry.HasTournamentWith(tournament.IDEQ(params.TournamentID)),
			tournamententry.HasGolferWith(golfer.IDEQ(params.GolferID)),
		).
		Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check tournament field: %w", err)
	}
	if !inField {
		return nil, ErrGolferNotInField
	}

	golferUsed, err := s.db.Pick.Query().
		Where(
			pick.HasUserWith(user.IDEQ(params.UserID)),
			pick.HasGolferWith(golfer.IDEQ(params.GolferID)),
			pick.HasLeagueWith(league.IDEQ(params.LeagueID)),
			pick.SeasonYearEQ(tournamentEnt.SeasonYear),
		).
		Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check golfer usage: %w", err)
	}
	if golferUsed {
		return nil, ErrGolferAlreadyUsed
	}

	pickExists, err := s.db.Pick.Query().
		Where(
			pick.HasUserWith(user.IDEQ(params.UserID)),
			pick.HasTournamentWith(tournament.IDEQ(params.TournamentID)),
			pick.HasLeagueWith(league.IDEQ(params.LeagueID)),
		).
		Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing pick: %w", err)
	}
	if pickExists {
		return nil, ErrPickAlreadyExists
	}

	pickEnt, err := s.db.Pick.Create().
		SetUserID(params.UserID).
		SetTournamentID(params.TournamentID).
		SetGolferID(params.GolferID).
		SetLeagueID(params.LeagueID).
		SetSeasonYear(tournamentEnt.SeasonYear).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create pick: %w", err)
	}

	return &Pick{
		ID:             pickEnt.ID,
		UserID:         params.UserID,
		TournamentID:   params.TournamentID,
		GolferID:       params.GolferID,
		LeagueID:       params.LeagueID,
		SeasonYear:     pickEnt.SeasonYear,
		CreatedAt:      pickEnt.CreatedAt,
		TournamentName: tournamentEnt.Name,
		GolferName:     golferEnt.Name,
	}, nil
}

func (s *Service) GetUserPicks(ctx context.Context, userID, leagueID uuid.UUID, seasonYear int) (*ListPicksResponse, error) {
	query := s.db.Pick.Query().
		Where(pick.HasUserWith(user.IDEQ(userID))).
		WithTournament().
		WithGolfer().
		Order(ent.Desc(pick.FieldCreatedAt))

	if leagueID != uuid.Nil {
		query = query.Where(pick.HasLeagueWith(league.IDEQ(leagueID)))
	}
	if seasonYear > 0 {
		query = query.Where(pick.SeasonYearEQ(seasonYear))
	}

	picks, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user picks: %w", err)
	}

	result := make([]Pick, 0, len(picks))
	for _, p := range picks {
		item := Pick{
			ID:         p.ID,
			UserID:     userID,
			SeasonYear: p.SeasonYear,
			CreatedAt:  p.CreatedAt,
		}

		if p.Edges.Tournament != nil {
			item.TournamentID = p.Edges.Tournament.ID
			item.TournamentName = p.Edges.Tournament.Name
		}
		if p.Edges.Golfer != nil {
			item.GolferID = p.Edges.Golfer.ID
			item.GolferName = p.Edges.Golfer.Name
		}

		result = append(result, item)
	}

	return &ListPicksResponse{
		Picks: result,
		Total: len(result),
	}, nil
}

func (s *Service) GetLeaguePicks(ctx context.Context, leagueID, tournamentID uuid.UUID) (*ListPicksResponse, error) {
	picks, err := s.db.Pick.Query().
		Where(
			pick.HasLeagueWith(league.IDEQ(leagueID)),
			pick.HasTournamentWith(tournament.IDEQ(tournamentID)),
		).
		WithUser().
		WithGolfer().
		WithTournament().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get league picks: %w", err)
	}

	tournamentEnt, err := s.db.Tournament.Query().
		Where(tournament.IDEQ(tournamentID)).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, fmt.Errorf("failed to get tournament: %w", err)
	}

	result := make([]Pick, 0, len(picks))
	for _, p := range picks {
		item := Pick{
			ID:         p.ID,
			LeagueID:   leagueID,
			SeasonYear: p.SeasonYear,
			CreatedAt:  p.CreatedAt,
		}

		if p.Edges.User != nil {
			item.UserID = p.Edges.User.ID
			if p.Edges.User.DisplayName != nil {
				item.UserName = *p.Edges.User.DisplayName
			}
		}
		if p.Edges.Tournament != nil {
			item.TournamentID = p.Edges.Tournament.ID
			item.TournamentName = p.Edges.Tournament.Name
		}
		if p.Edges.Golfer != nil {
			item.GolferID = p.Edges.Golfer.ID
			item.GolferName = p.Edges.Golfer.Name

			if tournamentEnt != nil && tournamentEnt.Status != tournament.StatusUpcoming {
				entry, err := s.db.TournamentEntry.Query().
					Where(
						tournamententry.HasTournamentWith(tournament.IDEQ(tournamentID)),
						tournamententry.HasGolferWith(golfer.IDEQ(p.Edges.Golfer.ID)),
					).
					Only(ctx)
				if err == nil {
					item.GolferPosition = entry.Position
					item.GolferEarnings = entry.Earnings
				}
			}
		}

		result = append(result, item)
	}

	return &ListPicksResponse{
		Picks: result,
		Total: len(result),
	}, nil
}

func (s *Service) CanMakePick(ctx context.Context, tournamentID uuid.UUID) (*PickWindowStatus, error) {
	tournamentEnt, err := s.db.Tournament.Query().
		Where(tournament.IDEQ(tournamentID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrTournamentNotFound
		}
		return nil, fmt.Errorf("failed to get tournament: %w", err)
	}

	status := s.getPickWindowStatus(tournamentEnt)
	return &status, nil
}

func (s *Service) getPickWindowStatus(t *ent.Tournament) PickWindowStatus {
	now := time.Now().UTC()
	opensAt := t.StartDate.AddDate(0, 0, -pickWindowDays)
	closesAt := t.StartDate

	status := PickWindowStatus{
		TournamentID:   t.ID,
		TournamentName: t.Name,
		OpensAt:        opensAt,
		ClosesAt:       closesAt,
	}

	switch {
	case t.Status != tournament.StatusUpcoming:
		status.IsOpen = false
		status.Reason = "tournament has already started"
	case now.Before(opensAt):
		status.IsOpen = false
		status.Reason = "pick window not yet open"
	case now.After(closesAt):
		status.IsOpen = false
		status.Reason = "pick window has closed"
	default:
		status.IsOpen = true
	}

	return status
}

func (s *Service) GetAvailableGolfers(ctx context.Context, userID, leagueID, tournamentID uuid.UUID) (*AvailableGolfersResponse, error) {
	tournamentEnt, err := s.db.Tournament.Query().
		Where(tournament.IDEQ(tournamentID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrTournamentNotFound
		}
		return nil, fmt.Errorf("failed to get tournament: %w", err)
	}

	entries, err := s.db.TournamentEntry.Query().
		Where(tournamententry.HasTournamentWith(tournament.IDEQ(tournamentID))).
		WithGolfer().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tournament entries: %w", err)
	}

	usedGolferIDs := make(map[uuid.UUID]bool)
	usedPicks, err := s.db.Pick.Query().
		Where(
			pick.HasUserWith(user.IDEQ(userID)),
			pick.HasLeagueWith(league.IDEQ(leagueID)),
			pick.SeasonYearEQ(tournamentEnt.SeasonYear),
		).
		WithGolfer().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get used picks: %w", err)
	}
	for _, p := range usedPicks {
		if p.Edges.Golfer != nil {
			usedGolferIDs[p.Edges.Golfer.ID] = true
		}
	}

	golfers := make([]AvailableGolfer, 0)
	for _, entry := range entries {
		if entry.Edges.Golfer == nil {
			continue
		}
		g := entry.Edges.Golfer

		if usedGolferIDs[g.ID] {
			continue
		}

		ag := AvailableGolfer{
			ID:          g.ID,
			Name:        g.Name,
			CountryCode: g.CountryCode,
		}
		if g.Owgr != nil {
			ag.OWGR = *g.Owgr
		}
		if g.Country != nil {
			ag.Country = *g.Country
		}
		if g.ImageURL != nil {
			ag.ImageURL = *g.ImageURL
		}

		golfers = append(golfers, ag)
	}

	return &AvailableGolfersResponse{
		Golfers: golfers,
		Total:   len(golfers),
	}, nil
}
