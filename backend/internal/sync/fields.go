package sync

import (
	"context"
	"log"
	"time"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/ent/tournamententry"
)

func (i *Ingester) syncFields(ctx context.Context) {
	log.Println("Checking for tournaments needing field sync...")

	now := time.Now().UTC()
	windowEnd := now.Add(7 * 24 * time.Hour)

	tournaments, err := i.db.Tournament.Query().
		Where(
			tournament.Not(tournament.HasChampion()),
			tournament.PickWindowClosesAtGTE(now),
			tournament.StartDateLTE(windowEnd),
		).
		All(ctx)
	if err != nil {
		log.Printf("failed to query upcoming tournaments: %v", err)
		i.captureJobError("sync_fields", err)
		return
	}

	if len(tournaments) == 0 {
		log.Println("No upcoming tournaments in sync window")
		i.recordSync(ctx, "fields")
		return
	}

	for _, t := range tournaments {
		if t.PgaTourID == nil || *t.PgaTourID == "" {
			log.Printf("Tournament %s has no PGA Tour ID, skipping field sync", t.Name)
			continue
		}

		hasField, err := i.tournamentHasField(ctx, t)
		if err != nil {
			log.Printf("failed to check field for tournament %s: %v", t.Name, err)
			i.captureJobError("sync_fields", err)
			continue
		}

		if hasField {
			log.Printf("Tournament %s already has field data, skipping", t.Name)
			continue
		}

		log.Printf("Syncing field for tournament: %s", t.Name)
		result, err := i.fieldsService.SyncTournamentField(ctx, t.ID)
		if err != nil {
			log.Printf("failed to sync field for tournament %s: %v", t.Name, err)
			i.captureJobError("sync_fields", err)
			continue
		}

		log.Printf("Field sync for %s: total=%d, matched=%d, new=%d, updated=%d",
			t.Name, result.TotalPlayers, result.MatchedPlayers, result.NewEntries, result.UpdatedEntries)
	}

	i.recordSync(ctx, "fields")
	log.Println("Field sync completed")
}

func (i *Ingester) tournamentHasField(ctx context.Context, t *ent.Tournament) (bool, error) {
	count, err := i.db.TournamentEntry.Query().
		Where(tournamententry.HasTournamentWith(tournament.IDEQ(t.ID))).
		Count(ctx)
	if err != nil {
		return false, err
	}
	return count >= 50, nil
}
