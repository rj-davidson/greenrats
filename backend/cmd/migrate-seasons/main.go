package main

import (
	"context"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/league"
	"github.com/rj-davidson/greenrats/ent/pick"
	"github.com/rj-davidson/greenrats/ent/season"
	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/internal/config"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	db, err := ent.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	ctx := context.Background()

	if err := createSeasons(ctx, db); err != nil {
		return fmt.Errorf("failed to create seasons: %w", err)
	}

	if err := linkTournaments(ctx, db); err != nil {
		return fmt.Errorf("failed to link tournaments: %w", err)
	}

	if err := linkLeagues(ctx, db); err != nil {
		return fmt.Errorf("failed to link leagues: %w", err)
	}

	if err := linkPicks(ctx, db); err != nil {
		return fmt.Errorf("failed to link picks: %w", err)
	}

	log.Println("Season migration completed successfully")
	return nil
}

func createSeasons(ctx context.Context, db *ent.Client) error {
	seasonYears := []int{2025, 2026}

	for _, year := range seasonYears {
		exists, err := db.Season.Query().
			Where(season.YearEQ(year)).
			Exist(ctx)
		if err != nil {
			return err
		}
		if exists {
			log.Printf("Season %d already exists, skipping", year)
			continue
		}

		startDate := time.Date(year-1, time.October, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(year, time.September, 30, 23, 59, 59, 0, time.UTC)
		isCurrent := year == 2026

		_, err = db.Season.Create().
			SetYear(year).
			SetStartDate(startDate).
			SetEndDate(endDate).
			SetIsCurrent(isCurrent).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create season %d: %w", year, err)
		}
		log.Printf("Created season %d (current=%v)", year, isCurrent)
	}

	return nil
}

func linkTournaments(ctx context.Context, db *ent.Client) error {
	tournaments, err := db.Tournament.Query().
		Where(tournament.Not(tournament.HasSeason())).
		All(ctx)
	if err != nil {
		return err
	}

	log.Printf("Found %d tournaments without season link", len(tournaments))

	for _, t := range tournaments {
		seasonRecord, err := db.Season.Query().
			Where(season.YearEQ(t.SeasonYear)).
			Only(ctx)
		if err != nil {
			log.Printf("Warning: no season found for tournament %s (year %d): %v", t.Name, t.SeasonYear, err)
			continue
		}

		_, err = t.Update().
			SetSeason(seasonRecord).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to link tournament %s: %w", t.Name, err)
		}
		log.Printf("Linked tournament %s to season %d", t.Name, t.SeasonYear)
	}

	return nil
}

func linkLeagues(ctx context.Context, db *ent.Client) error {
	leagues, err := db.League.Query().
		Where(league.Not(league.HasSeason())).
		All(ctx)
	if err != nil {
		return err
	}

	log.Printf("Found %d leagues without season link", len(leagues))

	for _, l := range leagues {
		seasonRecord, err := db.Season.Query().
			Where(season.YearEQ(l.SeasonYear)).
			Only(ctx)
		if err != nil {
			log.Printf("Warning: no season found for league %s (year %d): %v", l.Name, l.SeasonYear, err)
			continue
		}

		_, err = l.Update().
			SetSeason(seasonRecord).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to link league %s: %w", l.Name, err)
		}
		log.Printf("Linked league %s to season %d", l.Name, l.SeasonYear)
	}

	return nil
}

func linkPicks(ctx context.Context, db *ent.Client) error {
	picks, err := db.Pick.Query().
		Where(pick.Not(pick.HasSeason())).
		All(ctx)
	if err != nil {
		return err
	}

	log.Printf("Found %d picks without season link", len(picks))

	for _, p := range picks {
		seasonRecord, err := db.Season.Query().
			Where(season.YearEQ(p.SeasonYear)).
			Only(ctx)
		if err != nil {
			log.Printf("Warning: no season found for pick (year %d): %v", p.SeasonYear, err)
			continue
		}

		_, err = p.Update().
			SetSeason(seasonRecord).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to link pick: %w", err)
		}
	}

	log.Printf("Linked %d picks to seasons", len(picks))
	return nil
}
