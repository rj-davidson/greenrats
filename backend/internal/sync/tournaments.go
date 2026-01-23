package sync

import (
	"context"
	"time"
)

func (i *Ingester) syncTournaments(ctx context.Context) {
	start := time.Now()
	i.logger.Info("sync started", "type", "tournaments", "season", i.config.CurrentSeason)

	tournaments, err := i.ballDontLie.GetTournaments(ctx, i.config.CurrentSeason)
	if err != nil {
		i.logger.Error("failed to fetch tournaments from BallDontLie", "error", err)
		i.captureJobError("sync_tournaments", err)
		SyncErrors.WithLabelValues("tournaments").Inc()
		SyncRunsTotal.WithLabelValues("tournaments", "error").Inc()
		return
	}

	i.logger.Debug("fetched tournaments", "count", len(tournaments), "season", i.config.CurrentSeason)

	created, updated := 0, 0
	for idx := range tournaments {
		result, err := i.syncService.UpsertTournament(ctx, &tournaments[idx])
		if err != nil {
			i.logger.Error("failed to upsert tournament", "tournament", tournaments[idx].Name, "error", err)
			i.captureJobError("sync_tournaments", err)
			SyncErrors.WithLabelValues("tournaments").Inc()
			continue
		}

		if result.Created {
			created++
			i.logger.Debug("created tournament", "name", tournaments[idx].Name)
		} else {
			updated++
			if result.BecameCompleted {
				i.sendTournamentResultsEmails(ctx, result.Tournament)
			}
		}
	}

	i.recordSync(ctx, "tournaments")
	duration := time.Since(start)
	SyncDuration.WithLabelValues("tournaments").Observe(duration.Seconds())
	SyncRecordsProcessed.WithLabelValues("tournaments", "created").Add(float64(created))
	SyncRecordsProcessed.WithLabelValues("tournaments", "updated").Add(float64(updated))
	SyncRunsTotal.WithLabelValues("tournaments", "success").Inc()
	LastSyncTimestamp.WithLabelValues("tournaments").Set(float64(time.Now().Unix()))
	i.logger.Info("sync completed", "type", "tournaments", "duration", duration, "created", created, "updated", updated)
}
