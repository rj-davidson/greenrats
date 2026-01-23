package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/golfer"
	"github.com/rj-davidson/greenrats/ent/leaderboardentry"
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
			tournament.BdlIDNotNil(),
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

	roundResults, err := i.ballDontLie.GetPlayerRoundResults(ctx, *t.BdlID)
	if err != nil {
		return fmt.Errorf("failed to fetch round results: %w", err)
	}

	roundsProcessed := 0
	for idx := range roundResults {
		result := &roundResults[idx]

		entry, err := i.db.LeaderboardEntry.Query().
			Where(
				leaderboardentry.HasTournamentWith(tournament.IDEQ(t.ID)),
				leaderboardentry.HasGolferWith(golfer.BdlID(result.Player.ID)),
			).
			Only(ctx)
		if err != nil {
			continue
		}

		_, err = i.syncService.UpsertRound(ctx, entry.ID, result)
		if err != nil {
			i.logger.Error("failed to upsert round", "player", result.Player.DisplayName, "round", result.RoundNumber, "error", err)
			continue
		}
		roundsProcessed++
	}

	scorecards, err := i.ballDontLie.GetPlayerScorecards(ctx, *t.BdlID)
	if err != nil {
		return fmt.Errorf("failed to fetch scorecards: %w", err)
	}

	holesProcessed := 0
	for idx := range scorecards {
		sc := &scorecards[idx]

		entry, err := i.db.LeaderboardEntry.Query().
			Where(
				leaderboardentry.HasTournamentWith(tournament.IDEQ(t.ID)),
				leaderboardentry.HasGolferWith(golfer.BdlID(sc.Player.ID)),
			).
			WithRounds().
			Only(ctx)
		if err != nil {
			continue
		}

		var roundID *uuid.UUID
		for _, r := range entry.Edges.Rounds {
			if r.RoundNumber == sc.RoundNumber {
				roundID = &r.ID
				break
			}
		}
		if roundID == nil {
			continue
		}

		if err := i.syncService.UpsertHoleScore(ctx, *roundID, sc); err != nil {
			i.logger.Error("failed to upsert hole score", "player", sc.Player.DisplayName, "round", sc.RoundNumber, "hole", sc.HoleNumber, "error", err)
			continue
		}
		holesProcessed++
	}

	i.logger.Debug("fetched scorecards", "tournament", t.Name, "rounds", roundsProcessed, "holes", holesProcessed)
	SyncRecordsProcessed.WithLabelValues("scorecards", "rounds").Add(float64(roundsProcessed))
	SyncRecordsProcessed.WithLabelValues("scorecards", "holes").Add(float64(holesProcessed))
	return nil
}
