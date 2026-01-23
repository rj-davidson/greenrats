package fields

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/ent"
	entfieldentry "github.com/rj-davidson/greenrats/ent/fieldentry"
	"github.com/rj-davidson/greenrats/ent/golfer"
	"github.com/rj-davidson/greenrats/ent/tournament"
)

type Service struct {
	db     *ent.Client
	logger *slog.Logger
}

func NewService(db *ent.Client, logger *slog.Logger) *Service {
	return &Service{
		db:     db,
		logger: logger,
	}
}

func (s *Service) GetTournamentField(ctx context.Context, tournamentID uuid.UUID) ([]FieldEntry, error) {
	entries, err := s.db.FieldEntry.Query().
		Where(entfieldentry.HasTournamentWith(tournament.ID(tournamentID))).
		WithGolfer().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query field entries: %w", err)
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

func (s *Service) AddFieldEntry(ctx context.Context, tournamentID, golferID uuid.UUID, entryStatus string, qualifier *string) (*ent.FieldEntry, error) {
	t, err := s.db.Tournament.Get(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tournament: %w", err)
	}

	g, err := s.db.Golfer.Get(ctx, golferID)
	if err != nil {
		return nil, fmt.Errorf("failed to get golfer: %w", err)
	}

	existing, err := s.db.FieldEntry.Query().
		Where(
			entfieldentry.HasTournamentWith(tournament.ID(t.ID)),
			entfieldentry.HasGolferWith(golfer.ID(g.ID)),
		).
		Only(ctx)

	if err == nil {
		return existing, nil
	}
	if !ent.IsNotFound(err) {
		return nil, fmt.Errorf("failed to query existing entry: %w", err)
	}

	status := mapEntryStatus(entryStatus)

	entry, err := s.db.FieldEntry.Create().
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

func (s *Service) UpdateFieldEntry(ctx context.Context, entryID uuid.UUID, entryStatus, qualifier *string, owgrAtEntry *int, isAmateur *bool) (*ent.FieldEntry, error) {
	entry, err := s.db.FieldEntry.Get(ctx, entryID)
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
	err := s.db.FieldEntry.DeleteOneID(entryID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}
	return nil
}

func mapEntryStatus(status string) entfieldentry.EntryStatus {
	switch strings.ToLower(status) {
	case "confirmed":
		return entfieldentry.EntryStatusConfirmed
	case "alternate":
		return entfieldentry.EntryStatusAlternate
	case "withdrawn":
		return entfieldentry.EntryStatusWithdrawn
	case "pending":
		return entfieldentry.EntryStatusPending
	default:
		return entfieldentry.EntryStatusConfirmed
	}
}
