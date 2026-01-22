package sync

import (
	"context"
	"log"
)

func (i *Ingester) syncTournaments(ctx context.Context) {
	log.Println("Starting tournament sync...")

	tournaments, err := i.ballDontLie.GetTournaments(ctx, i.config.CurrentSeason)
	if err != nil {
		log.Printf("failed to fetch tournaments from BallDontLie: %v", err)
		i.captureJobError("sync_tournaments", err)
		return
	}

	log.Printf("Fetched %d tournaments for %d season", len(tournaments), i.config.CurrentSeason)

	for idx := range tournaments {
		result, err := i.syncService.UpsertTournament(ctx, &tournaments[idx])
		if err != nil {
			log.Printf("failed to upsert tournament %s: %v", tournaments[idx].Name, err)
			i.captureJobError("sync_tournaments", err)
			continue
		}

		if result.Created {
			log.Printf("Created tournament: %s", tournaments[idx].Name)
		} else {
			log.Printf("Updated tournament: %s", tournaments[idx].Name)
			if result.BecameCompleted {
				i.sendTournamentResultsEmails(ctx, result.Tournament)
			}
		}
	}

	i.recordSync(ctx, "tournaments")
	log.Println("Tournament sync completed")
}
