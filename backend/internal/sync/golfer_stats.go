package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/rj-davidson/greenrats/ent/golfer"
	"github.com/rj-davidson/greenrats/ent/season"
)

var relevantStatIDs = []int{
	120, // Scoring Average
	110, // Top 10 Finishes
	107, // Cuts Made
	106, // Events Played
	109, // Wins
	109, // Official Money
	102, // Driving Distance
	101, // Driving Accuracy Percentage
	103, // Greens in Regulation Percentage
	104, // Putting Average
	130, // Scrambling
}

func (i *Ingester) syncGolferSeasonStats(ctx context.Context) error {
	start := time.Now()
	i.logger.Info("sync started", "type", "golfer_stats")

	currentSeason, err := i.db.Season.Query().
		Where(season.IsCurrent(true)).
		Only(ctx)
	if err != nil {
		SyncErrors.WithLabelValues("golfer_stats").Inc()
		SyncRunsTotal.WithLabelValues("golfer_stats", "error").Inc()
		return fmt.Errorf("get current season: %w", err)
	}

	stats, err := i.ballDontLie.GetPlayerSeasonStats(ctx, currentSeason.Year, relevantStatIDs)
	if err != nil {
		SyncErrors.WithLabelValues("golfer_stats").Inc()
		SyncRunsTotal.WithLabelValues("golfer_stats", "error").Inc()
		return fmt.Errorf("fetch player season stats: %w", err)
	}

	processed := 0
	for idx := range stats {
		stat := &stats[idx]

		g, err := i.db.Golfer.Query().
			Where(golfer.BdlID(stat.Player.ID)).
			Only(ctx)
		if err != nil {
			if isContextError(err) {
				return fmt.Errorf("query golfer %s: %w", stat.Player.DisplayName, err)
			}
			continue
		}

		if err := i.syncService.UpsertGolferSeasonStat(ctx, g.ID, currentSeason.ID, stat); err != nil {
			if isContextError(err) {
				return fmt.Errorf("upsert stat for %s: %w", stat.Player.DisplayName, err)
			}
			continue
		}
		processed++
	}

	i.recordSync(ctx, "golfer_stats")
	duration := time.Since(start)
	SyncDuration.WithLabelValues("golfer_stats").Observe(duration.Seconds())
	SyncRecordsProcessed.WithLabelValues("golfer_stats", "stats").Add(float64(processed))
	SyncRunsTotal.WithLabelValues("golfer_stats", "success").Inc()
	LastSyncTimestamp.WithLabelValues("golfer_stats").Set(float64(time.Now().Unix()))
	i.logger.Info("sync completed", "type", "golfer_stats", "duration", duration, "stats_processed", processed)
	return nil
}
