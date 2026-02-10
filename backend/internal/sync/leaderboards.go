package sync

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/placement"
	"github.com/rj-davidson/greenrats/ent/tournament"
)

var placementSyncTimes = sync.Map{}

func (i *Ingester) shouldSyncPlacement(t *ent.Tournament) bool {
	if t.Edges.Champion == nil {
		return false
	}

	elapsed := time.Since(t.EndDate)

	if elapsed >= 365*24*time.Hour {
		return false
	}

	lastSyncVal, ok := placementSyncTimes.Load(t.ID)
	if !ok {
		return true
	}
	lastSync, ok := lastSyncVal.(time.Time)
	if !ok {
		return true
	}
	hoursSinceSync := time.Since(lastSync).Hours()

	switch {
	case elapsed < 6*time.Hour:
		return hoursSinceSync >= 0.5
	case elapsed < 24*time.Hour:
		return hoursSinceSync >= 2
	case elapsed < 3*24*time.Hour:
		return hoursSinceSync >= 6
	case elapsed < 7*24*time.Hour:
		return hoursSinceSync >= 24
	case elapsed < 28*24*time.Hour:
		return hoursSinceSync >= 72
	default:
		return hoursSinceSync >= 168
	}
}

func (i *Ingester) recordPlacementSync(tournamentID uuid.UUID) {
	placementSyncTimes.Store(tournamentID, time.Now())
}

func (i *Ingester) syncPlacements(ctx context.Context) error {
	start := time.Now()
	i.logger.Info("sync started", "type", "placements")

	tournaments, err := i.db.Tournament.Query().
		Where(
			tournament.HasChampion(),
			tournament.BdlIDNotNil(),
		).
		WithChampion().
		All(ctx)
	if err != nil {
		SyncErrors.WithLabelValues("placements").Inc()
		SyncRunsTotal.WithLabelValues("placements", "error").Inc()
		return fmt.Errorf("query completed tournaments: %w", err)
	}

	if len(tournaments) == 0 {
		i.logger.Debug("no completed tournaments found")
		SyncRunsTotal.WithLabelValues("placements", "skipped").Inc()
		return nil
	}

	synced := 0
	skipped := 0
	for _, t := range tournaments {
		if !i.shouldSyncPlacement(t) {
			skipped++
			continue
		}

		if err := i.syncTournamentPlacements(ctx, t); err != nil {
			if isContextError(err) {
				return fmt.Errorf("sync placements for %s: %w", t.Name, err)
			}
			SyncErrors.WithLabelValues("placements").Inc()
			continue
		}
		i.recordPlacementSync(t.ID)
		synced++
	}

	i.recordSync(ctx, "placements")
	duration := time.Since(start)
	SyncDuration.WithLabelValues("placements").Observe(duration.Seconds())
	SyncRecordsProcessed.WithLabelValues("placements", "tournaments").Add(float64(synced))
	SyncRunsTotal.WithLabelValues("placements", "success").Inc()
	LastSyncTimestamp.WithLabelValues("placements").Set(float64(time.Now().Unix()))
	i.logger.Info("sync completed", "type", "placements", "duration", duration, "tournaments_synced", synced, "tournaments_skipped", skipped)
	return nil
}

func (i *Ingester) syncActivePlacements(ctx context.Context) error {
	start := time.Now()
	i.logger.Info("sync started", "type", "active_placements")

	now := time.Now().UTC()
	tournaments, err := i.db.Tournament.Query().
		Where(
			tournament.Not(tournament.HasChampion()),
			tournament.PickWindowClosesAtLT(now),
			tournament.BdlIDNotNil(),
		).
		All(ctx)
	if err != nil {
		SyncErrors.WithLabelValues("active_placements").Inc()
		SyncRunsTotal.WithLabelValues("active_placements", "error").Inc()
		return fmt.Errorf("query active tournaments: %w", err)
	}

	if len(tournaments) == 0 {
		i.logger.Debug("no active tournaments found for placement sync")
		SyncRunsTotal.WithLabelValues("active_placements", "skipped").Inc()
		return nil
	}

	synced := 0
	for _, t := range tournaments {
		if err := i.syncTournamentPlacements(ctx, t); err != nil {
			if isContextError(err) {
				return fmt.Errorf("sync placements for %s: %w", t.Name, err)
			}
			SyncErrors.WithLabelValues("active_placements").Inc()
			continue
		}
		synced++
	}

	i.recordSync(ctx, "active_placements")
	duration := time.Since(start)
	SyncDuration.WithLabelValues("active_placements").Observe(duration.Seconds())
	SyncRecordsProcessed.WithLabelValues("active_placements", "tournaments").Add(float64(synced))
	SyncRunsTotal.WithLabelValues("active_placements", "success").Inc()
	LastSyncTimestamp.WithLabelValues("active_placements").Set(float64(time.Now().Unix()))
	i.logger.Info("sync completed", "type", "active_placements", "duration", duration, "tournaments_synced", synced)
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
	seenPlayerIDs := make(map[int]struct{}, len(results))
	for idx := range results {
		seenPlayerIDs[results[idx].Player.ID] = struct{}{}
		if err := i.syncService.UpsertPlacement(ctx, t, &results[idx]); err != nil {
			if isContextError(err) {
				return fmt.Errorf("upsert placement for %s: %w", results[idx].Player.DisplayName, err)
			}
			continue
		}
		processed++
	}

	completed, err := i.tournamentIsCompleted(ctx, t)
	if err != nil {
		return fmt.Errorf("check tournament completion for %s: %w", t.Name, err)
	}

	if completed {
		removed, err := i.pruneStalePlacements(ctx, t, seenPlayerIDs)
		if err != nil {
			return fmt.Errorf("prune stale placements for %s: %w", t.Name, err)
		}
		if removed > 0 {
			i.logger.Info("removed stale placements", "tournament", t.Name, "removed", removed)
		}
	}

	SyncRecordsProcessed.WithLabelValues("placements", "entries").Add(float64(processed))
	return nil
}

func (i *Ingester) tournamentIsCompleted(ctx context.Context, t *ent.Tournament) (bool, error) {
	if t.Edges.Champion != nil {
		return true, nil
	}

	return i.db.Tournament.Query().
		Where(
			tournament.IDEQ(t.ID),
			tournament.HasChampion(),
		).
		Exist(ctx)
}

func (i *Ingester) pruneStalePlacements(ctx context.Context, t *ent.Tournament, seenPlayerIDs map[int]struct{}) (int, error) {
	if len(seenPlayerIDs) == 0 {
		return 0, nil
	}

	existing, err := i.db.Placement.Query().
		Where(placement.HasTournamentWith(tournament.IDEQ(t.ID))).
		WithGolfer().
		All(ctx)
	if err != nil {
		return 0, err
	}

	stalePlacementIDs := make([]uuid.UUID, 0)
	for _, p := range existing {
		if p.Edges.Golfer == nil || p.Edges.Golfer.BdlID == nil {
			continue
		}

		if _, ok := seenPlayerIDs[*p.Edges.Golfer.BdlID]; ok {
			continue
		}

		stalePlacementIDs = append(stalePlacementIDs, p.ID)
	}

	if len(stalePlacementIDs) == 0 {
		return 0, nil
	}

	deleted, err := i.db.Placement.Delete().
		Where(placement.IDIn(stalePlacementIDs...)).
		Exec(ctx)
	if err != nil {
		return 0, err
	}

	return deleted, nil
}
