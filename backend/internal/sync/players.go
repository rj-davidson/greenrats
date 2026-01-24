package sync

import (
	"context"
	"fmt"
	"time"
)

func (i *Ingester) syncPlayers(ctx context.Context) error {
	start := time.Now()
	i.logger.Info("sync started", "type", "players")

	players, err := i.ballDontLie.GetPlayers(ctx)
	if err != nil {
		SyncErrors.WithLabelValues("players").Inc()
		SyncRunsTotal.WithLabelValues("players", "error").Inc()
		return fmt.Errorf("fetch players: %w", err)
	}

	i.logger.Debug("fetched players", "count", len(players))

	processed := 0
	for idx := range players {
		if err := i.syncService.UpsertPlayer(ctx, &players[idx]); err != nil {
			if isContextError(err) {
				return fmt.Errorf("upsert player %s: %w", players[idx].DisplayName, err)
			}
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
	return nil
}
