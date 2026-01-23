package sync

import (
	"context"
	"time"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/leaderboardentry"
	"github.com/rj-davidson/greenrats/ent/tournament"
)

func (i *Ingester) syncEarnings(ctx context.Context) {
	start := time.Now()
	i.logger.Info("sync started", "type", "earnings")

	now := time.Now().UTC()
	currentYear := now.Year()

	tournaments, err := i.db.Tournament.Query().
		Where(
			tournament.HasChampion(),
			tournament.SeasonYearGTE(currentYear-1),
		).
		All(ctx)
	if err != nil {
		i.logger.Error("failed to query completed tournaments", "error", err)
		i.captureJobError("sync_earnings", err)
		SyncErrors.WithLabelValues("earnings").Inc()
		SyncRunsTotal.WithLabelValues("earnings", "error").Inc()
		return
	}

	if len(tournaments) == 0 {
		i.logger.Debug("no completed tournaments found")
		SyncRunsTotal.WithLabelValues("earnings", "skipped").Inc()
		return
	}

	synced := 0
	for _, t := range tournaments {
		if t.BdlID == nil {
			i.logger.Debug("tournament has no BallDontLie ID, skipping", "tournament", t.Name)
			continue
		}

		hasEarnings, err := i.tournamentHasEarnings(ctx, t)
		if err != nil {
			i.logger.Error("failed to check earnings", "tournament", t.Name, "error", err)
			i.captureJobError("sync_earnings", err)
			SyncErrors.WithLabelValues("earnings").Inc()
			continue
		}

		daysSinceEnd := int(now.Sub(t.EndDate).Hours() / 24)
		shouldSync := !hasEarnings ||
			daysSinceEnd == 1 || daysSinceEnd == 2 || daysSinceEnd == 3 ||
			daysSinceEnd == 7 || daysSinceEnd == 14

		if shouldSync {
			if !hasEarnings {
				i.logger.Debug("tournament missing earnings, syncing", "tournament", t.Name)
			} else {
				i.logger.Debug("refreshing earnings for recently completed tournament", "tournament", t.Name, "days_since_end", daysSinceEnd)
			}

			if err := i.syncTournamentLeaderboard(ctx, t); err != nil {
				i.logger.Error("failed to sync earnings", "tournament", t.Name, "error", err)
				i.captureJobError("sync_earnings", err)
				SyncErrors.WithLabelValues("earnings").Inc()
				continue
			}
			synced++
		}
	}

	i.recordSync(ctx, "earnings")
	duration := time.Since(start)
	SyncDuration.WithLabelValues("earnings").Observe(duration.Seconds())
	SyncRecordsProcessed.WithLabelValues("earnings", "tournaments").Add(float64(synced))
	SyncRunsTotal.WithLabelValues("earnings", "success").Inc()
	LastSyncTimestamp.WithLabelValues("earnings").Set(float64(time.Now().Unix()))
	i.logger.Info("sync completed", "type", "earnings", "duration", duration, "tournaments_synced", synced)
}

func (i *Ingester) tournamentHasEarnings(ctx context.Context, t *ent.Tournament) (bool, error) {
	return i.db.LeaderboardEntry.Query().
		Where(
			leaderboardentry.HasTournamentWith(tournament.IDEQ(t.ID)),
			leaderboardentry.EarningsGT(0),
		).
		Exist(ctx)
}
