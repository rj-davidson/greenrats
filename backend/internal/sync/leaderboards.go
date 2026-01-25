package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/tournament"
)

func (i *Ingester) syncPlacements(ctx context.Context) error {
	start := time.Now()
	i.logger.Info("sync started", "type", "placements")

	now := time.Now().UTC()
	tournaments, err := i.db.Tournament.Query().
		Where(
			tournament.Not(tournament.HasChampion()),
			tournament.PickWindowClosesAtLT(now),
		).
		All(ctx)
	if err != nil {
		SyncErrors.WithLabelValues("placements").Inc()
		SyncRunsTotal.WithLabelValues("placements", "error").Inc()
		return fmt.Errorf("query active tournaments: %w", err)
	}

	if len(tournaments) == 0 {
		i.logger.Debug("no active tournaments found")
		SyncRunsTotal.WithLabelValues("placements", "skipped").Inc()
		return nil
	}

	synced := 0
	for _, t := range tournaments {
		if t.BdlID == nil {
			i.logger.Debug("tournament has no BallDontLie ID, skipping", "tournament", t.Name)
			continue
		}

		if err := i.syncTournamentPlacements(ctx, t); err != nil {
			if isContextError(err) {
				return fmt.Errorf("sync placements for %s: %w", t.Name, err)
			}
			SyncErrors.WithLabelValues("placements").Inc()
			continue
		}
		synced++
	}

	i.recordSync(ctx, "placements")
	duration := time.Since(start)
	SyncDuration.WithLabelValues("placements").Observe(duration.Seconds())
	SyncRecordsProcessed.WithLabelValues("placements", "tournaments").Add(float64(synced))
	SyncRunsTotal.WithLabelValues("placements", "success").Inc()
	LastSyncTimestamp.WithLabelValues("placements").Set(float64(time.Now().Unix()))
	i.logger.Info("sync completed", "type", "placements", "duration", duration, "tournaments_synced", synced)
	return nil
}

func (i *Ingester) syncTournamentPlacements(ctx context.Context, t *ent.Tournament) error {
	i.logger.Debug("syncing placements", "tournament", t.Name)

	results, err := i.ballDontLie.GetTournamentResults(ctx, *t.BdlID)
	if err != nil {
		return fmt.Errorf("fetch tournament results: %w", err)
	}

	i.logger.Debug("fetched results", "tournament", t.Name, "count", len(results))

	processed := 0
	for idx := range results {
		if err := i.syncService.UpsertPlacement(ctx, t, &results[idx]); err != nil {
			if isContextError(err) {
				return fmt.Errorf("upsert placement for %s: %w", results[idx].Player.DisplayName, err)
			}
			continue
		}
		processed++
	}

	SyncRecordsProcessed.WithLabelValues("placements", "entries").Add(float64(processed))
	return nil
}
