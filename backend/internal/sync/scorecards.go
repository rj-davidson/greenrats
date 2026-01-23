package sync

import (
	"context"
	"time"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/tournament"
)

func (i *Ingester) syncScorecards(ctx context.Context) {
	start := time.Now()
	i.logger.Info("sync started", "type", "scorecards")

	if !i.isAnyTournamentInPlayHours(ctx) {
		i.logger.Debug("not during play hours, skipping scorecard sync")
		SyncRunsTotal.WithLabelValues("scorecards", "skipped").Inc()
		return
	}

	now := time.Now().UTC()
	tournaments, err := i.db.Tournament.Query().
		Where(
			tournament.Not(tournament.HasChampion()),
			tournament.PickWindowClosesAtLT(now),
		).
		All(ctx)
	if err != nil {
		i.logger.Error("failed to query active tournaments", "error", err)
		i.captureJobError("sync_scorecards", err)
		SyncErrors.WithLabelValues("scorecards").Inc()
		SyncRunsTotal.WithLabelValues("scorecards", "error").Inc()
		return
	}

	if len(tournaments) == 0 {
		i.logger.Debug("no active tournaments found")
		SyncRunsTotal.WithLabelValues("scorecards", "skipped").Inc()
		return
	}

	synced := 0
	for _, t := range tournaments {
		if t.BdlID == nil {
			i.logger.Debug("tournament has no BallDontLie ID, skipping", "tournament", t.Name)
			continue
		}

		if !i.isTournamentInPlayHours(t) {
			continue
		}

		if err := i.syncTournamentScorecards(ctx, t); err != nil {
			i.logger.Error("failed to sync scorecards", "tournament", t.Name, "error", err)
			i.captureJobError("sync_scorecards", err)
			SyncErrors.WithLabelValues("scorecards").Inc()
			continue
		}
		synced++
	}

	i.recordSync(ctx, "scorecards")
	duration := time.Since(start)
	SyncDuration.WithLabelValues("scorecards").Observe(duration.Seconds())
	SyncRecordsProcessed.WithLabelValues("scorecards", "tournaments").Add(float64(synced))
	SyncRunsTotal.WithLabelValues("scorecards", "success").Inc()
	LastSyncTimestamp.WithLabelValues("scorecards").Set(float64(time.Now().Unix()))
	i.logger.Info("sync completed", "type", "scorecards", "duration", duration, "tournaments_synced", synced)
}

func (i *Ingester) syncTournamentScorecards(ctx context.Context, t *ent.Tournament) error {
	i.logger.Debug("syncing scorecards", "tournament", t.Name)

	scorecards, err := i.ballDontLie.GetPlayerScorecards(ctx, *t.BdlID)
	if err != nil {
		return err
	}

	i.logger.Debug("fetched scorecards", "tournament", t.Name, "count", len(scorecards))
	SyncRecordsProcessed.WithLabelValues("scorecards", "entries").Add(float64(len(scorecards)))
	return nil
}
