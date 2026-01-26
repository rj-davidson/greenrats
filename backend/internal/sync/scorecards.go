package sync

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/golfer"
	"github.com/rj-davidson/greenrats/ent/round"
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

		g, err := i.db.Golfer.Query().
			Where(golfer.BdlID(result.Player.ID)).
			Only(ctx)
		if err != nil {
			continue
		}

		_, err = i.syncService.UpsertRound(ctx, t.ID, g.ID, result)
		if err != nil {
			i.logger.Error("failed to upsert round", "player", result.Player.DisplayName, "round", result.RoundNumber, "error", err)
			continue
		}
		roundsProcessed++
	}

	holesProcessed := i.fetchAndProcessScorecardsParallel(ctx, t, playerIDs)

	i.logger.Debug("synced scorecards", "tournament", t.Name, "rounds", roundsProcessed, "holes", holesProcessed)
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
	for result := range resultsCh {
		if result.err != nil {
			i.logger.Error("failed to fetch scorecards", "player_id", result.playerID, "error", result.err)
			continue
		}
		holes := i.processPlayerScorecards(ctx, t, result.scorecards)
		holesProcessed += holes
	}

	return holesProcessed
}

func (i *Ingester) processPlayerScorecards(ctx context.Context, t *ent.Tournament, scorecards []balldontlie.PlayerScorecard) int {
	holesProcessed := 0
	for idx := range scorecards {
		sc := &scorecards[idx]

		g, err := i.db.Golfer.Query().
			Where(golfer.BdlID(sc.Player.ID)).
			Only(ctx)
		if err != nil {
			continue
		}

		existingRound, err := i.db.Round.Query().
			Where(
				round.HasTournamentWith(tournament.IDEQ(t.ID)),
				round.HasGolferWith(golfer.IDEQ(g.ID)),
				round.RoundNumberEQ(sc.RoundNumber),
			).
			Only(ctx)

		var roundID uuid.UUID
		switch {
		case ent.IsNotFound(err):
			roundResult := &balldontlie.PlayerRoundResult{
				RoundNumber: sc.RoundNumber,
				Player:      sc.Player,
			}
			newRound, err := i.syncService.UpsertRound(ctx, t.ID, g.ID, roundResult)
			if err != nil {
				i.logger.Error("failed to create round for scorecard", "player", sc.Player.DisplayName, "round", sc.RoundNumber, "error", err)
				continue
			}
			roundID = newRound.ID
		case err != nil:
			continue
		default:
			roundID = existingRound.ID
		}

		if err := i.syncService.UpsertHoleScore(ctx, roundID, sc); err != nil {
			i.logger.Error("failed to upsert hole score", "player", sc.Player.DisplayName, "round", sc.RoundNumber, "hole", sc.HoleNumber, "error", err)
			continue
		}
		holesProcessed++
	}
	return holesProcessed
}
