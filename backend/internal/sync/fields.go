package sync

import (
	"context"
	"time"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/ent/tournamententry"
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
		if t.PgaTourID == nil || *t.PgaTourID == "" {
			i.logger.Debug("tournament has no PGA Tour ID, skipping", "tournament", t.Name)
			continue
		}

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
		result, err := i.fieldsService.SyncTournamentField(ctx, t.ID)
		if err != nil {
			i.logger.Error("failed to sync field", "tournament", t.Name, "error", err)
			i.captureJobError("sync_fields", err)
			SyncErrors.WithLabelValues("fields").Inc()
			continue
		}

		i.logger.Debug("field sync completed",
			"tournament", t.Name,
			"total", result.TotalPlayers,
			"matched", result.MatchedPlayers,
			"new", result.NewEntries,
			"updated", result.UpdatedEntries)
		synced++
		SyncRecordsProcessed.WithLabelValues("fields", "entries").Add(float64(result.NewEntries + result.UpdatedEntries))
	}

	i.recordSync(ctx, "fields")
	duration := time.Since(start)
	SyncDuration.WithLabelValues("fields").Observe(duration.Seconds())
	SyncRecordsProcessed.WithLabelValues("fields", "tournaments").Add(float64(synced))
	SyncRunsTotal.WithLabelValues("fields", "success").Inc()
	LastSyncTimestamp.WithLabelValues("fields").Set(float64(time.Now().Unix()))
	i.logger.Info("sync completed", "type", "fields", "duration", duration, "tournaments_synced", synced)
}

func (i *Ingester) tournamentHasField(ctx context.Context, t *ent.Tournament) (bool, error) {
	count, err := i.db.TournamentEntry.Query().
		Where(tournamententry.HasTournamentWith(tournament.IDEQ(t.ID))).
		Count(ctx)
	if err != nil {
		return false, err
	}
	return count >= 50, nil
}
