package admin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/golfer"
	"github.com/rj-davidson/greenrats/ent/league"
	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/ent/tournamententry"
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
		WithChampion().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query tournaments: %w", err)
	}

	now := time.Now().UTC()
	result := make([]AdminTournament, len(tournaments))
	for i, t := range tournaments {
		status := deriveAdminTournamentStatus(t, now)
		result[i] = AdminTournament{
			ID:        t.ID,
			Name:      t.Name,
			Status:    status,
			StartDate: t.StartDate,
			EndDate:   t.EndDate,
		}
	}

	return &ListTournamentsResponse{
		Tournaments: result,
		Total:       len(result),
	}, nil
}

func deriveAdminTournamentStatus(t *ent.Tournament, now time.Time) string {
	if t.Edges.Champion != nil {
		return "completed"
	}

	if t.PickWindowClosesAt != nil && now.After(*t.PickWindowClosesAt) {
		return "active"
	}
	return "upcoming"
}

var ErrTournamentNotFound = errors.New("tournament not found")
var ErrGolferNotFound = errors.New("golfer not found")
var ErrEntryNotFound = errors.New("entry not found")

func (s *Service) ListTournamentField(ctx context.Context, tournamentID uuid.UUID) (*ListFieldResponse, error) {
	entries, err := s.db.TournamentEntry.Query().
		Where(tournamententry.HasTournamentWith(tournament.ID(tournamentID))).
		WithGolfer().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query tournament entries: %w", err)
	}

	result := make([]FieldEntryResponse, 0, len(entries))
	for _, e := range entries {
		if e.Edges.Golfer == nil {
			continue
		}
		g := e.Edges.Golfer

		result = append(result, FieldEntryResponse{
			ID:          e.ID,
			GolferID:    g.ID,
			GolferName:  g.Name,
			CountryCode: g.CountryCode,
			EntryStatus: string(e.EntryStatus),
			Qualifier:   e.Qualifier,
			OWGRAtEntry: e.OwgrAtEntry,
			IsAmateur:   e.IsAmateur,
		})
	}

	return &ListFieldResponse{
		Entries: result,
		Total:   len(result),
	}, nil
}

func (s *Service) AddFieldEntry(ctx context.Context, tournamentID uuid.UUID, req *AddFieldEntryRequest) (*FieldEntryResponse, error) {
	t, err := s.db.Tournament.Get(ctx, tournamentID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrTournamentNotFound
		}
		return nil, fmt.Errorf("failed to get tournament: %w", err)
	}

	g, err := s.db.Golfer.Get(ctx, req.GolferID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrGolferNotFound
		}
		return nil, fmt.Errorf("failed to get golfer: %w", err)
	}

	existing, err := s.db.TournamentEntry.Query().
		Where(
			tournamententry.HasTournamentWith(tournament.ID(t.ID)),
			tournamententry.HasGolferWith(golfer.ID(g.ID)),
		).
		Only(ctx)

	if err == nil {
		return &FieldEntryResponse{
			ID:          existing.ID,
			GolferID:    g.ID,
			GolferName:  g.Name,
			CountryCode: g.CountryCode,
			EntryStatus: string(existing.EntryStatus),
			Qualifier:   existing.Qualifier,
			OWGRAtEntry: existing.OwgrAtEntry,
			IsAmateur:   existing.IsAmateur,
		}, nil
	}
	if !ent.IsNotFound(err) {
		return nil, fmt.Errorf("failed to check existing entry: %w", err)
	}

	entryStatus := tournamententry.EntryStatusConfirmed
	if req.EntryStatus != "" {
		entryStatus = mapEntryStatus(req.EntryStatus)
	}

	entry, err := s.db.TournamentEntry.Create().
		SetTournament(t).
		SetGolfer(g).
		SetEntryStatus(entryStatus).
		SetNillableQualifier(req.Qualifier).
		SetNillableOwgrAtEntry(g.Owgr).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create entry: %w", err)
	}

	return &FieldEntryResponse{
		ID:          entry.ID,
		GolferID:    g.ID,
		GolferName:  g.Name,
		CountryCode: g.CountryCode,
		EntryStatus: string(entry.EntryStatus),
		Qualifier:   entry.Qualifier,
		OWGRAtEntry: entry.OwgrAtEntry,
		IsAmateur:   entry.IsAmateur,
	}, nil
}

func (s *Service) UpdateFieldEntry(ctx context.Context, entryID uuid.UUID, req *UpdateFieldEntryRequest) (*FieldEntryResponse, error) {
	entry, err := s.db.TournamentEntry.Query().
		Where(tournamententry.ID(entryID)).
		WithGolfer().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrEntryNotFound
		}
		return nil, fmt.Errorf("failed to get entry: %w", err)
	}

	update := entry.Update()

	if req.EntryStatus != nil {
		update.SetEntryStatus(mapEntryStatus(*req.EntryStatus))
	}
	if req.Qualifier != nil {
		update.SetQualifier(*req.Qualifier)
	}
	if req.OWGRAtEntry != nil {
		update.SetOwgrAtEntry(*req.OWGRAtEntry)
	}
	if req.IsAmateur != nil {
		update.SetIsAmateur(*req.IsAmateur)
	}

	updated, err := update.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update entry: %w", err)
	}

	g := entry.Edges.Golfer
	if g == nil {
		g, _ = s.db.Golfer.Get(ctx, entry.QueryGolfer().OnlyIDX(ctx))
	}

	golferName := ""
	countryCode := ""
	golferID := uuid.UUID{}
	if g != nil {
		golferID = g.ID
		golferName = g.Name
		countryCode = g.CountryCode
	}

	return &FieldEntryResponse{
		ID:          updated.ID,
		GolferID:    golferID,
		GolferName:  golferName,
		CountryCode: countryCode,
		EntryStatus: string(updated.EntryStatus),
		Qualifier:   updated.Qualifier,
		OWGRAtEntry: updated.OwgrAtEntry,
		IsAmateur:   updated.IsAmateur,
	}, nil
}

func (s *Service) DeleteFieldEntry(ctx context.Context, entryID uuid.UUID) error {
	err := s.db.TournamentEntry.DeleteOneID(entryID).Exec(ctx)
	if ent.IsNotFound(err) {
		return ErrEntryNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}
	return nil
}

func mapEntryStatus(status string) tournamententry.EntryStatus {
	switch status {
	case "confirmed":
		return tournamententry.EntryStatusConfirmed
	case "alternate":
		return tournamententry.EntryStatusAlternate
	case "withdrawn":
		return tournamententry.EntryStatusWithdrawn
	case "pending":
		return tournamententry.EntryStatusPending
	default:
		return tournamententry.EntryStatusConfirmed
	}
}
