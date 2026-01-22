package sync

import (
	"context"
	"log"
	"time"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/ent/tournamententry"
)

func (i *Ingester) syncEarnings(ctx context.Context) {
	log.Println("Checking for tournaments needing earnings sync...")

	now := time.Now().UTC()
	currentYear := now.Year()

	tournaments, err := i.db.Tournament.Query().
		Where(
			tournament.HasChampion(),
			tournament.SeasonYearGTE(currentYear-1),
		).
		All(ctx)
	if err != nil {
		log.Printf("failed to query completed tournaments: %v", err)
		i.captureJobError("sync_earnings", err)
		return
	}

	if len(tournaments) == 0 {
		log.Println("No completed tournaments found")
		return
	}

	for _, t := range tournaments {
		if t.BdlID == nil {
			log.Printf("Tournament %s has no BallDontLie ID, skipping earnings sync", t.Name)
			continue
		}

		hasEarnings, err := i.tournamentHasEarnings(ctx, t)
		if err != nil {
			log.Printf("failed to check earnings for tournament %s: %v", t.Name, err)
			i.captureJobError("sync_earnings", err)
			continue
		}

		daysSinceEnd := int(now.Sub(t.EndDate).Hours() / 24)
		shouldSync := !hasEarnings || daysSinceEnd == 1 || daysSinceEnd == 2 || daysSinceEnd == 7

		if shouldSync {
			if !hasEarnings {
				log.Printf("Tournament %s missing earnings, syncing...", t.Name)
			} else {
				log.Printf("Refreshing earnings for recently completed tournament %s (day %d)", t.Name, daysSinceEnd)
			}

			if err := i.syncTournamentLeaderboard(ctx, t); err != nil {
				log.Printf("failed to sync earnings for tournament %s: %v", t.Name, err)
				i.captureJobError("sync_earnings", err)
			}
		}
	}

	i.recordSync(ctx, "earnings")
	log.Println("Earnings sync completed")
}

func (i *Ingester) tournamentHasEarnings(ctx context.Context, t *ent.Tournament) (bool, error) {
	return i.db.TournamentEntry.Query().
		Where(
			tournamententry.HasTournamentWith(tournament.IDEQ(t.ID)),
			tournamententry.EarningsGT(0),
		).
		Exist(ctx)
}
