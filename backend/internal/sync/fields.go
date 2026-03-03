package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/tournament"
)

func (i *Ingester) syncFields(ctx context.Context) error {
	start := time.Now()
	i.logger.Info("sync started", "type", "fields")

	now := time.Now().UTC()
	windowEnd := now.Add(7 * 24 * time.Hour)

	tournaments, err := i.db.Tournament.Query().
		Where(
			tournament.Not(tournament.HasChampion()),
			tournament.EndDateGTE(now),
			tournament.StartDateLTE(windowEnd),
			tournament.BdlIDNotNil(),
		).
		All(ctx)
	if err != nil {
		SyncErrors.WithLabelValues("fields").Inc()
		SyncRunsTotal.WithLabelValues("fields", "error").Inc()
		return fmt.Errorf("query upcoming tournaments: %w", err)
	}

	if len(tournaments) == 0 {
		i.logger.Debug("no upcoming tournaments in sync window")
		i.recordSync(ctx, "fields")
		SyncRunsTotal.WithLabelValues("fields", "skipped").Inc()
		return nil
	}

	synced := 0
	for _, t := range tournaments {
		i.logger.Debug("syncing field", "tournament", t.Name)
		if err := i.syncTournamentField(ctx, t); err != nil {
			if isContextError(err) {
				return fmt.Errorf("sync field for %s: %w", t.Name, err)
			}
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
	return nil
}

func (i *Ingester) syncTournamentField(ctx context.Context, t *ent.Tournament) error {
	fields, err := i.ballDontLie.GetTournamentField(ctx, *t.BdlID)
	if err != nil {
		return fmt.Errorf("fetch tournament field: %w", err)
	}

	i.logger.Debug("fetched field", "tournament", t.Name, "count", len(fields))

	processed := 0
	for idx := range fields {
		if err := i.syncService.UpsertFieldEntry(ctx, t, &fields[idx]); err != nil {
			if isContextError(err) {
				return fmt.Errorf("upsert field entry for %s: %w", fields[idx].Player.DisplayName, err)
			}
			continue
		}
		processed++
	}

	SyncRecordsProcessed.WithLabelValues("fields", "entries").Add(float64(processed))

	if err := i.syncTournamentFutures(ctx, t); err != nil {
		if isContextError(err) {
			return fmt.Errorf("sync futures for %s: %w", t.Name, err)
		}
		i.logger.Warn("failed to sync futures", "tournament", t.Name, "error", err)
	}

	return nil
}

func (i *Ingester) syncTournamentFutures(ctx context.Context, t *ent.Tournament) error {
	futures, err := i.ballDontLie.GetFutures(ctx, *t.BdlID)
	if err != nil {
		return fmt.Errorf("fetch futures: %w", err)
	}

	i.logger.Debug("fetched futures", "tournament", t.Name, "count", len(futures))

	processed := 0
	for idx := range futures {
		if err := i.syncService.UpsertTournamentOdds(ctx, t, &futures[idx]); err != nil {
			if isContextError(err) {
				return fmt.Errorf("upsert tournament odds for %s: %w", futures[idx].Player.DisplayName, err)
			}
			continue
		}
		processed++
	}

	SyncRecordsProcessed.WithLabelValues("fields", "odds").Add(float64(processed))
	return nil
}
