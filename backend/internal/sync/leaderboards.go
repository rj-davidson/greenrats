package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/tournament"
)

func (i *Ingester) syncLeaderboards(ctx context.Context) {
	start := time.Now()
	i.logger.Info("sync started", "type", "leaderboards")

	now := time.Now().UTC()
	tournaments, err := i.db.Tournament.Query().
		Where(
			tournament.Not(tournament.HasChampion()),
			tournament.PickWindowClosesAtLT(now),
		).
		All(ctx)
	if err != nil {
		i.logger.Error("failed to query active tournaments", "error", err)
		i.captureJobError("sync_leaderboards", err)
		SyncErrors.WithLabelValues("leaderboards").Inc()
		SyncRunsTotal.WithLabelValues("leaderboards", "error").Inc()
		return
	}

	if len(tournaments) == 0 {
		i.logger.Debug("no active tournaments found")
		SyncRunsTotal.WithLabelValues("leaderboards", "skipped").Inc()
		return
	}

	synced := 0
	for _, t := range tournaments {
		if t.BdlID == nil {
			i.logger.Debug("tournament has no BallDontLie ID, skipping", "tournament", t.Name)
			continue
		}

		if err := i.syncTournamentLeaderboard(ctx, t); err != nil {
			i.logger.Error("failed to sync leaderboard", "tournament", t.Name, "error", err)
			i.captureJobError("sync_leaderboards", err)
			SyncErrors.WithLabelValues("leaderboards").Inc()
			continue
		}
		synced++
	}

	i.recordSync(ctx, "leaderboards")
	duration := time.Since(start)
	SyncDuration.WithLabelValues("leaderboards").Observe(duration.Seconds())
	SyncRecordsProcessed.WithLabelValues("leaderboards", "tournaments").Add(float64(synced))
	SyncRunsTotal.WithLabelValues("leaderboards", "success").Inc()
	LastSyncTimestamp.WithLabelValues("leaderboards").Set(float64(time.Now().Unix()))
	i.logger.Info("sync completed", "type", "leaderboards", "duration", duration, "tournaments_synced", synced)
}

func (i *Ingester) syncTournamentLeaderboard(ctx context.Context, t *ent.Tournament) error {
	i.logger.Debug("syncing leaderboard", "tournament", t.Name)

	results, err := i.ballDontLie.GetTournamentResults(ctx, *t.BdlID)
	if err != nil {
		return fmt.Errorf("failed to fetch tournament results: %w", err)
	}

	i.logger.Debug("fetched results", "tournament", t.Name, "count", len(results))

	processed := 0
	for idx := range results {
		if err := i.syncService.UpsertLeaderboardEntry(ctx, t, &results[idx]); err != nil {
			i.logger.Error("failed to upsert result", "player", results[idx].Player.DisplayName, "error", err)
			i.captureJobError("sync_tournament_leaderboard", err)
			continue
		}
		processed++
	}

	SyncRecordsProcessed.WithLabelValues("leaderboards", "entries").Add(float64(processed))
	return nil
}
