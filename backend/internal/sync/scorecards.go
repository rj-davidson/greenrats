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

const maxConcurrentFetches = 6

type playerScorecardResult struct {
	playerID   int
	scorecards []balldontlie.PlayerScorecard
	err        error
}

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

	roundResults, err := i.ballDontLie.GetPlayerRoundResults(ctx, *t.BdlID)
	if err != nil {
		return fmt.Errorf("failed to fetch round results: %w", err)
	}

	playerIDs := make(map[int]bool)
	roundsProcessed := 0
	for idx := range roundResults {
		result := &roundResults[idx]
		playerIDs[result.Player.ID] = true

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

	holesProcessed := i.fetchAndProcessScorecardsParallel(ctx, t, playerIDs)

	i.logger.Debug("fetched scorecards", "tournament", t.Name, "rounds", roundsProcessed, "holes", holesProcessed)
	SyncRecordsProcessed.WithLabelValues("scorecards", "rounds").Add(float64(roundsProcessed))
	SyncRecordsProcessed.WithLabelValues("scorecards", "holes").Add(float64(holesProcessed))
	return nil
}

func (i *Ingester) fetchAndProcessScorecardsParallel(ctx context.Context, t *ent.Tournament, playerIDs map[int]bool) int {
	resultsCh := make(chan playerScorecardResult)
	sem := make(chan struct{}, maxConcurrentFetches)

	var wg sync.WaitGroup

	for playerID := range playerIDs {
		wg.Add(1)
		go func(pid int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			scorecards, err := i.ballDontLie.GetPlayerScorecards(ctx, *t.BdlID, pid)
			resultsCh <- playerScorecardResult{playerID: pid, scorecards: scorecards, err: err}
		}(playerID)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	holesProcessed := 0
	processedEntries := make(map[uuid.UUID]bool)
	for result := range resultsCh {
		if result.err != nil {
			i.logger.Error("failed to fetch scorecards", "player_id", result.playerID, "error", result.err)
			continue
		}
		holes, entryID := i.processPlayerScorecards(ctx, t, result.scorecards)
		holesProcessed += holes
		if entryID != nil {
			processedEntries[*entryID] = true
		}
	}

	for entryID := range processedEntries {
		if err := i.syncService.UpdateLeaderboardProgress(ctx, entryID); err != nil {
			i.logger.Error("failed to update progress", "entry_id", entryID, "error", err)
		}
	}

	return holesProcessed
}

func (i *Ingester) processPlayerScorecards(ctx context.Context, t *ent.Tournament, scorecards []balldontlie.PlayerScorecard) (int, *uuid.UUID) {
	holesProcessed := 0
	var entryID *uuid.UUID
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

		if entryID == nil {
			entryID = &entry.ID
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
	return holesProcessed, entryID
}
