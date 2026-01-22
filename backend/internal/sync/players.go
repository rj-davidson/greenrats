package sync

import (
	"context"
	"log"
)

func (i *Ingester) syncPlayers(ctx context.Context) {
	log.Println("Starting player sync...")

	players, err := i.ballDontLie.GetPlayers(ctx)
	if err != nil {
		log.Printf("failed to fetch players from BallDontLie: %v", err)
		i.captureJobError("sync_players", err)
		return
	}

	log.Printf("Fetched %d players", len(players))

	for idx := range players {
		if err := i.syncService.UpsertPlayer(ctx, &players[idx]); err != nil {
			log.Printf("failed to upsert player %s: %v", players[idx].DisplayName, err)
			i.captureJobError("sync_players", err)
			continue
		}
	}

	i.recordSync(ctx, "players")
	log.Println("Player sync completed")
}
