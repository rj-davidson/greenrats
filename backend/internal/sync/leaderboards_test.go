package sync

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/ent/golfer"
	"github.com/rj-davidson/greenrats/ent/placement"
	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/internal/external/balldontlie"
	"github.com/rj-davidson/greenrats/internal/testutil"
)

func TestUpsertPlacement_SetsChampionPosition(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 100)
	tournamentEntity := testutil.CreateTournament(t, svc.db, "The Masters", 2026)

	earnings := 3600000.0

	result := &balldontlie.TournamentResult{
		Tournament: balldontlie.TournamentRef{
			ID:     1,
			Status: "COMPLETED",
		},
		Player: balldontlie.Player{
			ID:          100,
			DisplayName: "Scottie Scheffler",
		},
		Position:         "1",
		PositionNumeric:  1,
		TotalScore:       268,
		ParRelativeScore: -20,
		Earnings:         &earnings,
	}

	err := svc.UpsertPlacement(ctx, tournamentEntity, result)
	require.NoError(t, err)

	entry, err := svc.db.Placement.Query().
		Where(
			placement.HasTournamentWith(tournament.IDEQ(tournamentEntity.ID)),
			placement.HasGolferWith(golfer.IDEQ(golferEntity.ID)),
		).
		Only(ctx)
	require.NoError(t, err)

	assert.Equal(t, 1, *entry.PositionNumeric, "Champion should have position 1")
	assert.Equal(t, -20, *entry.ParRelativeScore, "Score should match")
	assert.Equal(t, 3600000, entry.Earnings, "Earnings should match")
	assert.Equal(t, placement.StatusFinished, entry.Status, "Status should be finished for completed tournament")
}

func TestUpsertPlacement_SetsMultiplePositions(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	champion := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 100)
	runnerUp := testutil.CreateGolfer(t, svc.db, "Collin Morikawa", 101)
	third := testutil.CreateGolfer(t, svc.db, "Rory McIlroy", 102)
	tournamentEntity := testutil.CreateTournament(t, svc.db, "The Masters", 2026)

	results := []struct {
		golfer   *balldontlie.Player
		position int
		earnings float64
	}{
		{&balldontlie.Player{ID: 100, DisplayName: "Scottie Scheffler"}, 1, 3600000},
		{&balldontlie.Player{ID: 101, DisplayName: "Collin Morikawa"}, 2, 2160000},
		{&balldontlie.Player{ID: 102, DisplayName: "Rory McIlroy"}, 3, 1360000},
	}

	for _, r := range results {
		earn := r.earnings
		result := &balldontlie.TournamentResult{
			Tournament: balldontlie.TournamentRef{
				ID:     1,
				Status: "COMPLETED",
			},
			Player:          *r.golfer,
			Position:        strconv.Itoa(r.position),
			PositionNumeric: r.position,
			Earnings:        &earn,
		}
		err := svc.UpsertPlacement(ctx, tournamentEntity, result)
		require.NoError(t, err)
	}

	championEntry, err := svc.db.Placement.Query().
		Where(
			placement.HasTournamentWith(tournament.IDEQ(tournamentEntity.ID)),
			placement.HasGolferWith(golfer.IDEQ(champion.ID)),
		).
		Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, *championEntry.PositionNumeric, "Champion should be position 1")
	assert.Equal(t, 3600000, championEntry.Earnings)

	runnerUpEntry, err := svc.db.Placement.Query().
		Where(
			placement.HasTournamentWith(tournament.IDEQ(tournamentEntity.ID)),
			placement.HasGolferWith(golfer.IDEQ(runnerUp.ID)),
		).
		Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, *runnerUpEntry.PositionNumeric, "Runner-up should be position 2")
	assert.Equal(t, 2160000, runnerUpEntry.Earnings)

	thirdEntry, err := svc.db.Placement.Query().
		Where(
			placement.HasTournamentWith(tournament.IDEQ(tournamentEntity.ID)),
			placement.HasGolferWith(golfer.IDEQ(third.ID)),
		).
		Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, *thirdEntry.PositionNumeric, "Third place should be position 3")
	assert.Equal(t, 1360000, thirdEntry.Earnings)
}

func TestUpsertPlacement_SetsCutStatus(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	golferEntity := testutil.CreateGolfer(t, svc.db, "Tiger Woods", 200)
	tournamentEntity := testutil.CreateTournament(t, svc.db, "The Open", 2026)

	result := &balldontlie.TournamentResult{
		Tournament: balldontlie.TournamentRef{
			ID:     2,
			Status: "COMPLETED",
		},
		Player: balldontlie.Player{
			ID:          200,
			DisplayName: "Tiger Woods",
		},
		Position: "CUT",
	}

	err := svc.UpsertPlacement(ctx, tournamentEntity, result)
	require.NoError(t, err)

	entry, err := svc.db.Placement.Query().
		Where(
			placement.HasTournamentWith(tournament.IDEQ(tournamentEntity.ID)),
			placement.HasGolferWith(golfer.IDEQ(golferEntity.ID)),
		).
		Only(ctx)
	require.NoError(t, err)

	assert.Equal(t, placement.StatusCut, entry.Status, "Player who missed cut should have cut status")
	assert.Equal(t, "CUT", entry.Position, "Cut players should have CUT position string")
	assert.Nil(t, entry.PositionNumeric, "Cut players should have nil position_numeric")
}

func TestUpsertPlacement_UpdatesExistingEntry(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	golferEntity := testutil.CreateGolfer(t, svc.db, "Jon Rahm", 300)
	tournamentEntity := testutil.CreateTournament(t, svc.db, "US Open", 2026)

	initialResult := &balldontlie.TournamentResult{
		Tournament: balldontlie.TournamentRef{
			ID:     3,
			Status: "COMPLETED",
		},
		Player: balldontlie.Player{
			ID:          300,
			DisplayName: "Jon Rahm",
		},
		Position:        "5",
		PositionNumeric: 5,
	}

	err := svc.UpsertPlacement(ctx, tournamentEntity, initialResult)
	require.NoError(t, err)

	entry, err := svc.db.Placement.Query().
		Where(
			placement.HasTournamentWith(tournament.IDEQ(tournamentEntity.ID)),
			placement.HasGolferWith(golfer.IDEQ(golferEntity.ID)),
		).
		Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, 5, *entry.PositionNumeric)
	assert.Equal(t, placement.StatusFinished, entry.Status)

	earnings := 4300000.0

	finalResult := &balldontlie.TournamentResult{
		Tournament: balldontlie.TournamentRef{
			ID:     3,
			Status: "COMPLETED",
		},
		Player: balldontlie.Player{
			ID:          300,
			DisplayName: "Jon Rahm",
		},
		Position:        "1",
		PositionNumeric: 1,
		Earnings:        &earnings,
	}

	err = svc.UpsertPlacement(ctx, tournamentEntity, finalResult)
	require.NoError(t, err)

	updatedEntry, err := svc.db.Placement.Query().
		Where(
			placement.HasTournamentWith(tournament.IDEQ(tournamentEntity.ID)),
			placement.HasGolferWith(golfer.IDEQ(golferEntity.ID)),
		).
		Only(ctx)
	require.NoError(t, err)

	assert.Equal(t, 1, *updatedEntry.PositionNumeric, "Position should be updated to 1 (champion)")
	assert.Equal(t, 4300000, updatedEntry.Earnings, "Earnings should be updated")
	assert.Equal(t, placement.StatusFinished, updatedEntry.Status, "Status should be finished")
}

func TestUpsertPlacement_SkipsUnknownGolfer(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	tournamentEntity := testutil.CreateTournament(t, svc.db, "PGA Championship", 2026)

	result := &balldontlie.TournamentResult{
		Tournament: balldontlie.TournamentRef{
			ID:     4,
			Status: "COMPLETED",
		},
		Player: balldontlie.Player{
			ID:          99999,
			DisplayName: "Unknown Player",
		},
		Position:        "1",
		PositionNumeric: 1,
	}

	err := svc.UpsertPlacement(ctx, tournamentEntity, result)
	require.NoError(t, err, "Should not error for unknown golfer")

	count, err := svc.db.Placement.Query().
		Where(placement.HasTournamentWith(tournament.IDEQ(tournamentEntity.ID))).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "No entry should be created for unknown golfer")
}
