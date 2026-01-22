package sync

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/tournament"
)

func (i *Ingester) syncLeaderboards(ctx context.Context) {
	log.Println("Checking for active tournament leaderboards...")

	now := time.Now().UTC()
	tournaments, err := i.db.Tournament.Query().
		Where(
			tournament.Not(tournament.HasChampion()),
			tournament.PickWindowClosesAtLT(now),
		).
		All(ctx)
	if err != nil {
		log.Printf("failed to query active tournaments: %v", err)
		i.captureJobError("sync_leaderboards", err)
		return
	}

	if len(tournaments) == 0 {
		log.Println("No active tournaments found")
		return
	}

	for _, t := range tournaments {
		if t.BdlID == nil {
			log.Printf("Tournament %s has no BallDontLie ID, skipping leaderboard sync", t.Name)
			continue
		}

		if err := i.syncTournamentLeaderboard(ctx, t); err != nil {
			log.Printf("failed to sync leaderboard for tournament %s: %v", t.Name, err)
			i.captureJobError("sync_leaderboards", err)
		}
	}

	i.recordSync(ctx, "leaderboards")
	log.Println("Leaderboard sync completed")
}

func (i *Ingester) syncTournamentLeaderboard(ctx context.Context, t *ent.Tournament) error {
	log.Printf("Syncing leaderboard for tournament: %s", t.Name)

	results, err := i.ballDontLie.GetTournamentResults(ctx, *t.BdlID)
	if err != nil {
		return fmt.Errorf("failed to fetch tournament results: %w", err)
	}

	log.Printf("Fetched %d results for tournament %s", len(results), t.Name)

	for idx := range results {
		if err := i.syncService.UpsertTournamentEntry(ctx, t, &results[idx]); err != nil {
			log.Printf("failed to upsert result for player %s: %v", results[idx].Player.DisplayName, err)
			i.captureJobError("sync_tournament_leaderboard", err)
			continue
		}
	}

	return nil
}
