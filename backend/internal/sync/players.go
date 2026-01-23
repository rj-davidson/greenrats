package sync

import (
	"context"
	"time"
)

func (i *Ingester) syncPlayers(ctx context.Context) {
	start := time.Now()
	i.logger.Info("sync started", "type", "players")

	players, err := i.ballDontLie.GetPlayers(ctx)
	if err != nil {
		i.logger.Error("failed to fetch players from BallDontLie", "error", err)
		i.captureJobError("sync_players", err)
		SyncErrors.WithLabelValues("players").Inc()
		SyncRunsTotal.WithLabelValues("players", "error").Inc()
		return
	}

	i.logger.Debug("fetched players", "count", len(players))

	processed := 0
	for idx := range players {
		if err := i.syncService.UpsertPlayer(ctx, &players[idx]); err != nil {
			i.logger.Error("failed to upsert player", "player", players[idx].DisplayName, "error", err)
			i.captureJobError("sync_players", err)
			SyncErrors.WithLabelValues("players").Inc()
			continue
		}
		processed++
	}

	i.recordSync(ctx, "players")
	duration := time.Since(start)
	SyncDuration.WithLabelValues("players").Observe(duration.Seconds())
	SyncRecordsProcessed.WithLabelValues("players", "upserted").Add(float64(processed))
	SyncRunsTotal.WithLabelValues("players", "success").Inc()
	LastSyncTimestamp.WithLabelValues("players").Set(float64(time.Now().Unix()))
	i.logger.Info("sync completed", "type", "players", "duration", duration, "processed", processed)
}
