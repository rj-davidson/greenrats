package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/tournament"
)

func (i *Ingester) syncTournaments(ctx context.Context) error {
	start := time.Now()
	i.logger.Info("sync started", "type", "tournaments", "season", i.config.CurrentSeason)

	seasonEnt, err := i.syncService.UpsertSeason(ctx, i.config.CurrentSeason)
	if err != nil {
		SyncErrors.WithLabelValues("tournaments").Inc()
		SyncRunsTotal.WithLabelValues("tournaments", "error").Inc()
		return fmt.Errorf("upsert season: %w", err)
	}

	tournaments, err := i.ballDontLie.GetTournaments(ctx, i.config.CurrentSeason)
	if err != nil {
		SyncErrors.WithLabelValues("tournaments").Inc()
		SyncRunsTotal.WithLabelValues("tournaments", "error").Inc()
		return fmt.Errorf("fetch tournaments: %w", err)
	}

	i.logger.Debug("fetched tournaments", "count", len(tournaments), "season", i.config.CurrentSeason)

	created, updated := 0, 0
	for idx := range tournaments {
		result, err := i.syncService.UpsertTournament(ctx, &tournaments[idx], seasonEnt)
		if err != nil {
			if isContextError(err) {
				return fmt.Errorf("upsert tournament %s: %w", tournaments[idx].Name, err)
			}
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
	return nil
}

func (i *Ingester) checkTournamentCompletion(ctx context.Context) error {
	start := time.Now()
	i.logger.Info("sync started", "type", "completion_check")

	now := time.Now().UTC()
	activeTournaments, err := i.db.Tournament.Query().
		Where(
			tournament.Not(tournament.HasChampion()),
			tournament.PickWindowClosesAtLT(now),
			tournament.BdlIDNotNil(),
		).
		All(ctx)
	if err != nil {
		return fmt.Errorf("query active tournaments: %w", err)
	}

	if len(activeTournaments) == 0 {
		i.logger.Debug("no active tournaments to check for completion")
		return nil
	}

	bdlToLocal := make(map[int]*ent.Tournament)
	for _, t := range activeTournaments {
		if t.BdlID != nil {
			bdlToLocal[*t.BdlID] = t
		}
	}

	tournaments, err := i.ballDontLie.GetTournaments(ctx, i.config.CurrentSeason)
	if err != nil {
		return fmt.Errorf("fetch tournaments: %w", err)
	}

	seasonEnt, err := i.syncService.UpsertSeason(ctx, i.config.CurrentSeason)
	if err != nil {
		return fmt.Errorf("upsert season: %w", err)
	}

	completed := 0
	for idx := range tournaments {
		apiTournament := &tournaments[idx]
		localTournament, exists := bdlToLocal[apiTournament.ID]
		if !exists {
			continue
		}

		if apiTournament.Champion == nil {
			continue
		}

		i.logger.Info("tournament completion detected",
			"tournament", localTournament.Name,
			"champion", apiTournament.Champion.DisplayName)

		result, err := i.syncService.UpsertTournament(ctx, apiTournament, seasonEnt)
		if err != nil {
			i.logger.Error("failed to update completed tournament",
				"tournament", localTournament.Name,
				"error", err)
			continue
		}

		if result.BecameCompleted {
			i.sendTournamentResultsEmails(ctx, result.Tournament)
		}
		completed++
	}

	i.logger.Info("sync completed", "type", "completion_check",
		"duration", time.Since(start),
		"active_checked", len(activeTournaments),
		"newly_completed", completed)
	return nil
}
