package sync

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/golfer"
	"github.com/rj-davidson/greenrats/ent/leaderboardentry"
	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/internal/external/balldontlie"
)

const playerBatchSize = 5

func (i *Ingester) syncScorecards(ctx context.Context) error {
	start := time.Now()
	i.logger.Info("sync started", "type", "scorecards")

	if !i.isAnyTournamentInPlayHours(ctx) {
		i.logger.Debug("not during play hours, skipping scorecard sync")
		SyncRunsTotal.WithLabelValues("scorecards", "skipped").Inc()
		return nil
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
		SyncErrors.WithLabelValues("scorecards").Inc()
		SyncRunsTotal.WithLabelValues("scorecards", "error").Inc()
		return fmt.Errorf("query active tournaments: %w", err)
	}

	if len(tournaments) == 0 {
		i.logger.Debug("no active tournaments found")
		SyncRunsTotal.WithLabelValues("scorecards", "skipped").Inc()
		return nil
	}

	synced := 0
	for _, t := range tournaments {
		if !i.isTournamentInPlayHours(t) {
			continue
		}

		if err := i.syncTournamentScorecards(ctx, t); err != nil {
			if isContextError(err) {
				return fmt.Errorf("sync scorecards for %s: %w", t.Name, err)
			}
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
	return nil
}

func (i *Ingester) syncTournamentScorecards(ctx context.Context, t *ent.Tournament) error {
	i.logger.Debug("syncing scorecards", "tournament", t.Name)

	entries, err := t.QueryLeaderboardEntries().WithGolfer().All(ctx)
	if err != nil {
		return fmt.Errorf("failed to query leaderboard entries: %w", err)
	}

	playerIDs := make([]int, 0, len(entries))
	for _, e := range entries {
		if e.Edges.Golfer != nil && e.Edges.Golfer.BdlID != nil && *e.Edges.Golfer.BdlID != 0 {
			playerIDs = append(playerIDs, *e.Edges.Golfer.BdlID)
		}
	}

	roundsProcessed := 0
	holesProcessed := 0

	for batchStart := 0; batchStart < len(playerIDs); batchStart += playerBatchSize {
		batchEnd := batchStart + playerBatchSize
		if batchEnd > len(playerIDs) {
			batchEnd = len(playerIDs)
		}
		batch := playerIDs[batchStart:batchEnd]

		var wg sync.WaitGroup
		var scorecards []balldontlie.PlayerScorecard
		var roundResults []balldontlie.PlayerRoundResult
		var scorecardsErr, roundsErr error

		wg.Add(2)
		go func() {
			defer wg.Done()
			scorecards, scorecardsErr = i.ballDontLie.GetPlayerScorecardsBatch(ctx, *t.BdlID, batch)
		}()
		go func() {
			defer wg.Done()
			roundResults, roundsErr = i.ballDontLie.GetPlayerRoundResultsBatch(ctx, *t.BdlID, batch)
		}()
		wg.Wait()

		if scorecardsErr != nil {
			i.logger.Error("failed to fetch scorecard batch", "error", scorecardsErr)
		}
		if roundsErr != nil {
			i.logger.Error("failed to fetch round results batch", "error", roundsErr)
		}

		processedEntries := make(map[uuid.UUID]bool)
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
				continue
			}
			processedEntries[entry.ID] = true
			roundsProcessed++
		}

		holes, entryIDs := i.processScorecards(ctx, t, scorecards)
		holesProcessed += holes
		for entryID := range entryIDs {
			processedEntries[entryID] = true
		}

		for entryID := range processedEntries {
			if err := i.syncService.UpdateLeaderboardProgress(ctx, entryID); err != nil {
				i.logger.Error("failed to update progress", "entry_id", entryID, "error", err)
			}
		}
	}

	i.logger.Debug("synced scorecards", "tournament", t.Name, "rounds", roundsProcessed, "holes", holesProcessed)
	SyncRecordsProcessed.WithLabelValues("scorecards", "rounds").Add(float64(roundsProcessed))
	SyncRecordsProcessed.WithLabelValues("scorecards", "holes").Add(float64(holesProcessed))
	return nil
}

func (i *Ingester) processScorecards(ctx context.Context, t *ent.Tournament, scorecards []balldontlie.PlayerScorecard) (int, map[uuid.UUID]bool) {
	holesProcessed := 0
	processedEntries := make(map[uuid.UUID]bool)
	entryCache := make(map[int]*ent.LeaderboardEntry)

	for idx := range scorecards {
		sc := &scorecards[idx]

		entry, ok := entryCache[sc.Player.ID]
		if !ok {
			var err error
			entry, err = i.db.LeaderboardEntry.Query().
				Where(
					leaderboardentry.HasTournamentWith(tournament.IDEQ(t.ID)),
					leaderboardentry.HasGolferWith(golfer.BdlID(sc.Player.ID)),
				).
				WithRounds().
				Only(ctx)
			if err != nil {
				continue
			}
			entryCache[sc.Player.ID] = entry
		}

		var roundID *uuid.UUID
		for _, r := range entry.Edges.Rounds {
			if r.RoundNumber == sc.RoundNumber {
				roundID = &r.ID
				break
			}
		}
		if roundID == nil {
			roundResult := &balldontlie.PlayerRoundResult{
				RoundNumber: sc.RoundNumber,
				Player:      sc.Player,
			}
			newRound, err := i.syncService.UpsertRound(ctx, entry.ID, roundResult)
			if err != nil {
				i.logger.Error("failed to create round for scorecard", "player", sc.Player.DisplayName, "round", sc.RoundNumber, "error", err)
				continue
			}
			roundID = &newRound.ID
			entry.Edges.Rounds = append(entry.Edges.Rounds, newRound)
		}

		if err := i.syncService.UpsertHoleScore(ctx, *roundID, sc); err != nil {
			i.logger.Error("failed to upsert hole score", "player", sc.Player.DisplayName, "round", sc.RoundNumber, "hole", sc.HoleNumber, "error", err)
			continue
		}
		holesProcessed++
		processedEntries[entry.ID] = true
	}
	return holesProcessed, processedEntries
}
