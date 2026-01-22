package fields

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/golfer"
	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/ent/tournamententry"
	"github.com/rj-davidson/greenrats/internal/external/pgatour"
	"github.com/rj-davidson/greenrats/internal/features/golfers"
)

type Service struct {
	db            *ent.Client
	pgatourClient pgatour.ClientInterface
	golferService *golfers.Service
	logger        *slog.Logger
}

func NewService(db *ent.Client, pgatourClient pgatour.ClientInterface, logger *slog.Logger) *Service {
	return &Service{
		db:            db,
		pgatourClient: pgatourClient,
		golferService: golfers.NewService(db),
		logger:        logger,
	}
}

func (s *Service) SyncTournamentField(ctx context.Context, tournamentID uuid.UUID) (*SyncResult, error) {
	t, err := s.db.Tournament.Get(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tournament: %w", err)
	}

	if t.PgaTourID == nil || *t.PgaTourID == "" {
		return nil, fmt.Errorf("tournament %s has no PGA Tour ID", t.Name)
	}

	result := &SyncResult{
		TournamentID:   t.ID,
		TournamentName: t.Name,
	}

	entries, err := s.pgatourClient.GetTournamentField(ctx, *t.PgaTourID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch field from PGA Tour: %w", err)
	}

	if len(entries) == 0 {
		s.logger.Info("no field data available", "tournament", t.Name, "pga_tour_id", *t.PgaTourID)
		return result, nil
	}

	result.TotalPlayers = len(entries)

	for _, entry := range entries {
		name := entry.DisplayName
		if name == "" {
			name = strings.TrimSpace(entry.FirstName + " " + entry.LastName)
		}

		g, err := s.golferService.LookupByName(ctx, name)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("lookup error for %s: %v", name, err))
			continue
		}

		if g == nil {
			s.logger.Warn("golfer not found in database", "name", name, "tournament", t.Name)
			continue
		}

		result.MatchedPlayers++

		existingEntry, err := s.db.TournamentEntry.Query().
			Where(
				tournamententry.HasTournamentWith(tournament.ID(t.ID)),
				tournamententry.HasGolferWith(golfer.ID(g.ID)),
			).
			Only(ctx)

		switch {
		case ent.IsNotFound(err):
			_, err = s.db.TournamentEntry.Create().
				SetTournament(t).
				SetGolfer(g).
				SetEntryStatus(tournamententry.EntryStatusConfirmed).
				SetIsAmateur(entry.IsAmateur).
				Save(ctx)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to create entry for %s: %v", name, err))
				continue
			}
			result.NewEntries++
		case err != nil:
			result.Errors = append(result.Errors, fmt.Sprintf("failed to query entry for %s: %v", name, err))
			continue
		default:
			_, err = existingEntry.Update().
				SetIsAmateur(entry.IsAmateur).
				Save(ctx)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to update entry for %s: %v", name, err))
				continue
			}
			result.UpdatedEntries++
		}
	}

	s.logger.Info("field sync complete",
		"tournament", t.Name,
		"total", result.TotalPlayers,
		"matched", result.MatchedPlayers,
		"new", result.NewEntries,
		"updated", result.UpdatedEntries,
		"errors", len(result.Errors),
	)

	return result, nil
}

func (s *Service) GetTournamentField(ctx context.Context, tournamentID uuid.UUID) ([]FieldEntry, error) {
	entries, err := s.db.TournamentEntry.Query().
		Where(tournamententry.HasTournamentWith(tournament.ID(tournamentID))).
		WithGolfer().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query tournament entries: %w", err)
	}

	result := make([]FieldEntry, 0, len(entries))
	for _, e := range entries {
		if e.Edges.Golfer == nil {
			continue
		}
		g := e.Edges.Golfer

		result = append(result, FieldEntry{
			GolferID:    g.ID,
			GolferName:  g.Name,
			CountryCode: g.CountryCode,
			EntryStatus: string(e.EntryStatus),
			Qualifier:   e.Qualifier,
			OWGRAtEntry: e.OwgrAtEntry,
			IsAmateur:   e.IsAmateur,
		})
	}

	return result, nil
}

func (s *Service) AddFieldEntry(ctx context.Context, tournamentID, golferID uuid.UUID, entryStatus string, qualifier *string) (*ent.TournamentEntry, error) {
	t, err := s.db.Tournament.Get(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tournament: %w", err)
	}

	g, err := s.db.Golfer.Get(ctx, golferID)
	if err != nil {
		return nil, fmt.Errorf("failed to get golfer: %w", err)
	}

	existing, err := s.db.TournamentEntry.Query().
		Where(
			tournamententry.HasTournamentWith(tournament.ID(t.ID)),
			tournamententry.HasGolferWith(golfer.ID(g.ID)),
		).
		Only(ctx)

	if err == nil {
		return existing, nil
	}
	if !ent.IsNotFound(err) {
		return nil, fmt.Errorf("failed to query existing entry: %w", err)
	}

	status := mapEntryStatus(entryStatus)

	entry, err := s.db.TournamentEntry.Create().
		SetTournament(t).
		SetGolfer(g).
		SetEntryStatus(status).
		SetNillableQualifier(qualifier).
		SetNillableOwgrAtEntry(g.Owgr).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create entry: %w", err)
	}

	return entry, nil
}

func (s *Service) UpdateFieldEntry(ctx context.Context, entryID uuid.UUID, entryStatus, qualifier *string, owgrAtEntry *int, isAmateur *bool) (*ent.TournamentEntry, error) {
	entry, err := s.db.TournamentEntry.Get(ctx, entryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entry: %w", err)
	}

	update := entry.Update()

	if entryStatus != nil {
		update.SetEntryStatus(mapEntryStatus(*entryStatus))
	}
	if qualifier != nil {
		update.SetQualifier(*qualifier)
	}
	if owgrAtEntry != nil {
		update.SetOwgrAtEntry(*owgrAtEntry)
	}
	if isAmateur != nil {
		update.SetIsAmateur(*isAmateur)
	}

	updated, err := update.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update entry: %w", err)
	}

	return updated, nil
}

func (s *Service) DeleteFieldEntry(ctx context.Context, entryID uuid.UUID) error {
	err := s.db.TournamentEntry.DeleteOneID(entryID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}
	return nil
}

func mapEntryStatus(status string) tournamententry.EntryStatus {
	switch strings.ToLower(status) {
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
