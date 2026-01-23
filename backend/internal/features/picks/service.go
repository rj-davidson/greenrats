package picks

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/commissioneraction"
	"github.com/rj-davidson/greenrats/ent/fieldentry"
	"github.com/rj-davidson/greenrats/ent/golfer"
	"github.com/rj-davidson/greenrats/ent/leaderboardentry"
	"github.com/rj-davidson/greenrats/ent/league"
	"github.com/rj-davidson/greenrats/ent/leaguemembership"
	"github.com/rj-davidson/greenrats/ent/pick"
	"github.com/rj-davidson/greenrats/ent/season"
	"github.com/rj-davidson/greenrats/ent/tournament"
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
	ErrPickNotFound          = errors.New("pick not found")
	ErrNotCommissioner       = errors.New("only the commissioner can perform this action")
	ErrTournamentCompleted   = errors.New("tournament has already completed")
	ErrSeasonNotFound        = errors.New("season not found")
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

	inField, err := s.db.FieldEntry.Query().
		Where(
			fieldentry.HasTournamentWith(tournament.IDEQ(params.TournamentID)),
			fieldentry.HasGolferWith(golfer.IDEQ(params.GolferID)),
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

	seasonEnt, err := s.db.Season.Query().
		Where(season.YearEQ(tournamentEnt.SeasonYear)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrSeasonNotFound
		}
		return nil, fmt.Errorf("failed to query season: %w", err)
	}

	pickEnt, err := s.db.Pick.Create().
		SetUserID(params.UserID).
		SetTournamentID(params.TournamentID).
		SetGolferID(params.GolferID).
		SetLeagueID(params.LeagueID).
		SetSeasonYear(tournamentEnt.SeasonYear).
		SetSeasonID(seasonEnt.ID).
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

type CreatePickForUserParams struct {
	CommissionerID uuid.UUID
	TargetUserID   uuid.UUID
	LeagueID       uuid.UUID
	TournamentID   uuid.UUID
	GolferID       uuid.UUID
}

func (s *Service) CreatePickForUser(ctx context.Context, params *CreatePickForUserParams) (*Pick, error) {
	isOwner, err := s.isLeagueOwner(ctx, params.LeagueID, params.CommissionerID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, ErrNotCommissioner
	}

	tournamentEnt, err := s.db.Tournament.Query().
		Where(tournament.IDEQ(params.TournamentID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrTournamentNotFound
		}
		return nil, fmt.Errorf("failed to get tournament: %w", err)
	}

	targetUserEnt, err := s.db.User.Query().
		Where(user.IDEQ(params.TargetUserID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	isMember, err := s.db.LeagueMembership.Query().
		Where(
			leaguemembership.HasUserWith(user.IDEQ(params.TargetUserID)),
			leaguemembership.HasLeagueWith(league.IDEQ(params.LeagueID)),
		).
		Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if !isMember {
		return nil, ErrNotLeagueMember
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

	inField, err := s.db.FieldEntry.Query().
		Where(
			fieldentry.HasTournamentWith(tournament.IDEQ(params.TournamentID)),
			fieldentry.HasGolferWith(golfer.IDEQ(params.GolferID)),
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
			pick.HasUserWith(user.IDEQ(params.TargetUserID)),
			pick.HasGolferWith(golfer.IDEQ(params.GolferID)),
			pick.HasLeagueWith(league.IDEQ(params.LeagueID)),
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
			pick.HasUserWith(user.IDEQ(params.TargetUserID)),
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

	seasonEnt, err := s.db.Season.Query().
		Where(season.YearEQ(tournamentEnt.SeasonYear)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrSeasonNotFound
		}
		return nil, fmt.Errorf("failed to query season: %w", err)
	}

	pickEnt, err := s.db.Pick.Create().
		SetUserID(params.TargetUserID).
		SetTournamentID(params.TournamentID).
		SetGolferID(params.GolferID).
		SetLeagueID(params.LeagueID).
		SetSeasonYear(tournamentEnt.SeasonYear).
		SetSeasonID(seasonEnt.ID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create pick: %w", err)
	}

	metadata := map[string]any{
		"golfer_id":       params.GolferID.String(),
		"golfer_name":     golferEnt.Name,
		"pick_user_id":    params.TargetUserID.String(),
		"tournament_id":   tournamentEnt.ID.String(),
		"tournament_name": tournamentEnt.Name,
	}

	_, err = s.db.CommissionerAction.
		Create().
		SetActionType(commissioneraction.ActionTypePickChange).
		SetDescription(fmt.Sprintf("Added pick %s", golferEnt.Name)).
		SetMetadata(metadata).
		SetLeagueID(params.LeagueID).
		SetCommissionerID(params.CommissionerID).
		SetAffectedUserID(params.TargetUserID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to log action: %w", err)
	}

	userName := ""
	if targetUserEnt.DisplayName != nil {
		userName = *targetUserEnt.DisplayName
	}

	return &Pick{
		ID:             pickEnt.ID,
		UserID:         params.TargetUserID,
		TournamentID:   params.TournamentID,
		GolferID:       params.GolferID,
		LeagueID:       params.LeagueID,
		SeasonYear:     pickEnt.SeasonYear,
		CreatedAt:      pickEnt.CreatedAt,
		UserName:       userName,
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
		WithChampion().
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

			if tournamentEnt != nil {
				hasChampion := tournamentEnt.Edges.Champion != nil
				status := deriveTournamentStatus(tournamentEnt, hasChampion)
				if status != "upcoming" {
					entry, err := s.db.LeaderboardEntry.Query().
						Where(
							leaderboardentry.HasTournamentWith(tournament.IDEQ(tournamentID)),
							leaderboardentry.HasGolferWith(golfer.IDEQ(p.Edges.Golfer.ID)),
						).
						Only(ctx)
					if err == nil {
						item.GolferPosition = entry.Position
						item.GolferEarnings = entry.Earnings
					}
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

	var opensAt, closesAt time.Time
	if t.PickWindowOpensAt != nil && t.PickWindowClosesAt != nil {
		opensAt = *t.PickWindowOpensAt
		closesAt = *t.PickWindowClosesAt
	} else {
		closesAt = t.StartDate
		opensAt = closesAt.AddDate(0, 0, -pickWindowDays)
	}

	status := PickWindowStatus{
		TournamentID:   t.ID,
		TournamentName: t.Name,
		OpensAt:        opensAt,
		ClosesAt:       closesAt,
	}

	switch {
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

type usedGolferInfo struct {
	TournamentID   uuid.UUID
	TournamentName string
}

func (s *Service) GetAvailableGolfers(ctx context.Context, userID, leagueID, tournamentID uuid.UUID) (*AvailableGolfersResponse, error) {
	tournamentExists, err := s.db.Tournament.Query().
		Where(tournament.IDEQ(tournamentID)).
		Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check tournament: %w", err)
	}
	if !tournamentExists {
		return nil, ErrTournamentNotFound
	}

	entries, err := s.db.FieldEntry.Query().
		Where(fieldentry.HasTournamentWith(tournament.IDEQ(tournamentID))).
		WithGolfer().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get field entries: %w", err)
	}

	usedGolfers := make(map[uuid.UUID]usedGolferInfo)
	usedPicks, err := s.db.Pick.Query().
		Where(
			pick.HasUserWith(user.IDEQ(userID)),
			pick.HasLeagueWith(league.IDEQ(leagueID)),
			pick.HasTournamentWith(tournament.IDNEQ(tournamentID)),
		).
		WithGolfer().
		WithTournament().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get used picks: %w", err)
	}
	for _, p := range usedPicks {
		if p.Edges.Golfer != nil && p.Edges.Tournament != nil {
			usedGolfers[p.Edges.Golfer.ID] = usedGolferInfo{
				TournamentID:   p.Edges.Tournament.ID,
				TournamentName: p.Edges.Tournament.Name,
			}
		}
	}

	golfers := make([]AvailableGolfer, 0)
	for _, entry := range entries {
		if entry.Edges.Golfer == nil {
			continue
		}
		g := entry.Edges.Golfer

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

		if info, used := usedGolfers[g.ID]; used {
			ag.IsUsed = true
			ag.UsedForTournamentID = &info.TournamentID
			ag.UsedForTournament = info.TournamentName
		}

		golfers = append(golfers, ag)
	}

	return &AvailableGolfersResponse{
		Golfers: golfers,
		Total:   len(golfers),
	}, nil
}

type AvailableGolfersForUserParams struct {
	CommissionerID uuid.UUID
	TargetUserID   uuid.UUID
	LeagueID       uuid.UUID
	TournamentID   uuid.UUID
}

func (s *Service) GetAvailableGolfersForUserOverride(ctx context.Context, params AvailableGolfersForUserParams) (*AvailableGolfersResponse, error) {
	isOwner, err := s.isLeagueOwner(ctx, params.LeagueID, params.CommissionerID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, ErrNotCommissioner
	}

	tournamentExists, err := s.db.Tournament.Query().
		Where(tournament.IDEQ(params.TournamentID)).
		Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check tournament: %w", err)
	}
	if !tournamentExists {
		return nil, ErrTournamentNotFound
	}

	entries, err := s.db.FieldEntry.Query().
		Where(fieldentry.HasTournamentWith(tournament.IDEQ(params.TournamentID))).
		WithGolfer().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get field entries: %w", err)
	}

	usedGolfers := make(map[uuid.UUID]usedGolferInfo)
	usedPicks, err := s.db.Pick.Query().
		Where(
			pick.HasUserWith(user.IDEQ(params.TargetUserID)),
			pick.HasLeagueWith(league.IDEQ(params.LeagueID)),
			pick.HasTournamentWith(tournament.IDNEQ(params.TournamentID)),
		).
		WithGolfer().
		WithTournament().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get used picks: %w", err)
	}
	for _, p := range usedPicks {
		if p.Edges.Golfer != nil && p.Edges.Tournament != nil {
			usedGolfers[p.Edges.Golfer.ID] = usedGolferInfo{
				TournamentID:   p.Edges.Tournament.ID,
				TournamentName: p.Edges.Tournament.Name,
			}
		}
	}

	golfers := make([]AvailableGolfer, 0)
	for _, entry := range entries {
		if entry.Edges.Golfer == nil {
			continue
		}
		g := entry.Edges.Golfer

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

		if info, used := usedGolfers[g.ID]; used {
			ag.IsUsed = true
			ag.UsedForTournamentID = &info.TournamentID
			ag.UsedForTournament = info.TournamentName
		}

		golfers = append(golfers, ag)
	}

	return &AvailableGolfersResponse{
		Golfers: golfers,
		Total:   len(golfers),
	}, nil
}

type OverridePickParams struct {
	LeagueID       uuid.UUID
	PickID         uuid.UUID
	NewGolferID    uuid.UUID
	CommissionerID uuid.UUID
}

func (s *Service) OverridePick(ctx context.Context, params OverridePickParams) (*Pick, error) {
	isOwner, err := s.isLeagueOwner(ctx, params.LeagueID, params.CommissionerID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, ErrNotCommissioner
	}

	pickEnt, err := s.db.Pick.Query().
		Where(
			pick.IDEQ(params.PickID),
			pick.HasLeagueWith(league.IDEQ(params.LeagueID)),
		).
		WithTournament().
		WithGolfer().
		WithUser().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrPickNotFound
		}
		return nil, fmt.Errorf("failed to get pick: %w", err)
	}

	if pickEnt.Edges.Tournament == nil {
		return nil, ErrTournamentNotFound
	}
	tournamentEnt := pickEnt.Edges.Tournament

	newGolferEnt, err := s.db.Golfer.Query().
		Where(golfer.IDEQ(params.NewGolferID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrGolferNotFound
		}
		return nil, fmt.Errorf("failed to get golfer: %w", err)
	}

	inField, err := s.db.FieldEntry.Query().
		Where(
			fieldentry.HasTournamentWith(tournament.IDEQ(tournamentEnt.ID)),
			fieldentry.HasGolferWith(golfer.IDEQ(params.NewGolferID)),
		).
		Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check tournament field: %w", err)
	}
	if !inField {
		return nil, ErrGolferNotInField
	}

	if pickEnt.Edges.User == nil {
		return nil, fmt.Errorf("pick has no associated user")
	}
	pickUserID := pickEnt.Edges.User.ID

	golferUsed, err := s.db.Pick.Query().
		Where(
			pick.HasUserWith(user.IDEQ(pickUserID)),
			pick.HasGolferWith(golfer.IDEQ(params.NewGolferID)),
			pick.HasLeagueWith(league.IDEQ(params.LeagueID)),
			pick.IDNEQ(params.PickID),
		).
		Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check golfer usage: %w", err)
	}
	if golferUsed {
		return nil, ErrGolferAlreadyUsed
	}

	oldGolferName := ""
	if pickEnt.Edges.Golfer != nil {
		oldGolferName = pickEnt.Edges.Golfer.Name
	}

	updatedPick, err := s.db.Pick.
		UpdateOneID(params.PickID).
		SetGolferID(params.NewGolferID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update pick: %w", err)
	}

	metadata := map[string]any{
		"new_golfer_id":   params.NewGolferID.String(),
		"new_golfer_name": newGolferEnt.Name,
		"pick_user_id":    pickUserID.String(),
		"tournament_id":   tournamentEnt.ID.String(),
		"tournament_name": tournamentEnt.Name,
	}
	if pickEnt.Edges.Golfer != nil {
		metadata["old_golfer_id"] = pickEnt.Edges.Golfer.ID.String()
		metadata["old_golfer_name"] = oldGolferName
	}

	_, err = s.db.CommissionerAction.
		Create().
		SetActionType(commissioneraction.ActionTypePickChange).
		SetDescription(fmt.Sprintf("Changed pick from %s to %s", oldGolferName, newGolferEnt.Name)).
		SetMetadata(metadata).
		SetLeagueID(params.LeagueID).
		SetCommissionerID(params.CommissionerID).
		SetAffectedUserID(pickUserID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to log action: %w", err)
	}

	userName := ""
	if pickEnt.Edges.User.DisplayName != nil {
		userName = *pickEnt.Edges.User.DisplayName
	}

	return &Pick{
		ID:             updatedPick.ID,
		UserID:         pickUserID,
		TournamentID:   tournamentEnt.ID,
		GolferID:       params.NewGolferID,
		LeagueID:       params.LeagueID,
		SeasonYear:     updatedPick.SeasonYear,
		CreatedAt:      updatedPick.CreatedAt,
		UserName:       userName,
		TournamentName: tournamentEnt.Name,
		GolferName:     newGolferEnt.Name,
	}, nil
}

func (s *Service) isLeagueOwner(ctx context.Context, leagueID, userID uuid.UUID) (bool, error) {
	exists, err := s.db.LeagueMembership.
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
	return exists, nil
}

type UpdatePickParams struct {
	UserID      uuid.UUID
	PickID      uuid.UUID
	NewGolferID uuid.UUID
}

var ErrNotPickOwner = errors.New("user does not own this pick")

func (s *Service) UpdateUserPick(ctx context.Context, params UpdatePickParams) (*Pick, error) {
	pickEnt, err := s.db.Pick.Query().
		Where(pick.IDEQ(params.PickID)).
		WithTournament().
		WithGolfer().
		WithUser().
		WithLeague().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrPickNotFound
		}
		return nil, fmt.Errorf("failed to get pick: %w", err)
	}

	if pickEnt.Edges.User == nil || pickEnt.Edges.User.ID != params.UserID {
		return nil, ErrNotPickOwner
	}

	if pickEnt.Edges.Tournament == nil {
		return nil, ErrTournamentNotFound
	}
	tournamentEnt := pickEnt.Edges.Tournament

	windowStatus := s.getPickWindowStatus(tournamentEnt)
	if !windowStatus.IsOpen {
		return nil, ErrPickWindowClosed
	}

	newGolferEnt, err := s.db.Golfer.Query().
		Where(golfer.IDEQ(params.NewGolferID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrGolferNotFound
		}
		return nil, fmt.Errorf("failed to get golfer: %w", err)
	}

	inField, err := s.db.FieldEntry.Query().
		Where(
			fieldentry.HasTournamentWith(tournament.IDEQ(tournamentEnt.ID)),
			fieldentry.HasGolferWith(golfer.IDEQ(params.NewGolferID)),
		).
		Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check tournament field: %w", err)
	}
	if !inField {
		return nil, ErrGolferNotInField
	}

	if pickEnt.Edges.League == nil {
		return nil, ErrLeagueNotFound
	}
	leagueID := pickEnt.Edges.League.ID

	golferUsed, err := s.db.Pick.Query().
		Where(
			pick.HasUserWith(user.IDEQ(params.UserID)),
			pick.HasGolferWith(golfer.IDEQ(params.NewGolferID)),
			pick.HasLeagueWith(league.IDEQ(leagueID)),
			pick.IDNEQ(params.PickID),
		).
		Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check golfer usage: %w", err)
	}
	if golferUsed {
		return nil, ErrGolferAlreadyUsed
	}

	updatedPick, err := s.db.Pick.
		UpdateOneID(params.PickID).
		SetGolferID(params.NewGolferID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update pick: %w", err)
	}

	return &Pick{
		ID:             updatedPick.ID,
		UserID:         params.UserID,
		TournamentID:   tournamentEnt.ID,
		GolferID:       params.NewGolferID,
		LeagueID:       leagueID,
		SeasonYear:     updatedPick.SeasonYear,
		CreatedAt:      updatedPick.CreatedAt,
		TournamentName: tournamentEnt.Name,
		GolferName:     newGolferEnt.Name,
	}, nil
}

func deriveTournamentStatus(t *ent.Tournament, hasChampion bool) string {
	if hasChampion {
		return "completed"
	}

	now := time.Now().UTC()
	if t.PickWindowClosesAt != nil && now.After(*t.PickWindowClosesAt) {
		return "active"
	}
	return "upcoming"
}
