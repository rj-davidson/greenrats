package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/fieldentry"
	"github.com/rj-davidson/greenrats/ent/tournament"
)

func (i *Ingester) syncFields(ctx context.Context) {
	start := time.Now()
	i.logger.Info("sync started", "type", "fields")

	now := time.Now().UTC()
	windowEnd := now.Add(7 * 24 * time.Hour)

	tournaments, err := i.db.Tournament.Query().
		Where(
			tournament.Not(tournament.HasChampion()),
			tournament.PickWindowClosesAtGTE(now),
			tournament.StartDateLTE(windowEnd),
			tournament.BdlIDNotNil(),
		).
		All(ctx)
	if err != nil {
		i.logger.Error("failed to query upcoming tournaments", "error", err)
		i.captureJobError("sync_fields", err)
		SyncErrors.WithLabelValues("fields").Inc()
		SyncRunsTotal.WithLabelValues("fields", "error").Inc()
		return
	}

	if len(tournaments) == 0 {
		i.logger.Debug("no upcoming tournaments in sync window")
		i.recordSync(ctx, "fields")
		SyncRunsTotal.WithLabelValues("fields", "skipped").Inc()
		return
	}

	synced := 0
	for _, t := range tournaments {
		hasField, err := i.tournamentHasField(ctx, t)
		if err != nil {
			i.logger.Error("failed to check field", "tournament", t.Name, "error", err)
			i.captureJobError("sync_fields", err)
			SyncErrors.WithLabelValues("fields").Inc()
			continue
		}

		if hasField {
			i.logger.Debug("tournament already has field data, skipping", "tournament", t.Name)
			continue
		}

		i.logger.Debug("syncing field", "tournament", t.Name)
		if err := i.syncTournamentField(ctx, t); err != nil {
			i.logger.Error("failed to sync field", "tournament", t.Name, "error", err)
			i.captureJobError("sync_fields", err)
			SyncErrors.WithLabelValues("fields").Inc()
			continue
		}
		synced++
	}

	i.recordSync(ctx, "fields")
	duration := time.Since(start)
	SyncDuration.WithLabelValues("fields").Observe(duration.Seconds())
	SyncRecordsProcessed.WithLabelValues("fields", "tournaments").Add(float64(synced))
	SyncRunsTotal.WithLabelValues("fields", "success").Inc()
	LastSyncTimestamp.WithLabelValues("fields").Set(float64(time.Now().Unix()))
	i.logger.Info("sync completed", "type", "fields", "duration", duration, "tournaments_synced", synced)
}

func (i *Ingester) syncTournamentField(ctx context.Context, t *ent.Tournament) error {
	fields, err := i.ballDontLie.GetTournamentField(ctx, *t.BdlID)
	if err != nil {
		return fmt.Errorf("failed to fetch tournament field: %w", err)
	}

	i.logger.Debug("fetched field", "tournament", t.Name, "count", len(fields))

	processed := 0
	for idx := range fields {
		if err := i.syncService.UpsertFieldEntry(ctx, t, &fields[idx]); err != nil {
			i.logger.Error("failed to upsert field entry", "player", fields[idx].Player.DisplayName, "error", err)
			i.captureJobError("sync_tournament_field", err)
			continue
		}
		processed++
	}

	SyncRecordsProcessed.WithLabelValues("fields", "entries").Add(float64(processed))
	return nil
}

func (i *Ingester) tournamentHasField(ctx context.Context, t *ent.Tournament) (bool, error) {
	count, err := i.db.FieldEntry.Query().
		Where(fieldentry.HasTournamentWith(tournament.IDEQ(t.ID))).
		Count(ctx)
	if err != nil {
		return false, err
	}
	return count >= 50, nil
}
