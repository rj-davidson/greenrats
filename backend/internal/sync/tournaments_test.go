package sync

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/internal/external/balldontlie"
	"github.com/rj-davidson/greenrats/internal/testutil"
)

func TestUpsertTournament_SetsChampion(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	golferEntity := testutil.CreateGolfer(t, svc.db, "Chris Gotterup", 63)
	seasonEntity := testutil.CreateSeason(t, svc.db, 2026)

	bdlTournament := &balldontlie.Tournament{
		ID:        7,
		Season:    2026,
		Name:      "Sony Open in Hawaii",
		StartDate: "2026-01-15T00:00:00.000Z",
		EndDate:   strPtr("2026-01-18T00:00:00.000Z"),
		Status:    strPtr("COMPLETED"),
		Champion: &balldontlie.Player{
			ID:          63,
			DisplayName: "Chris Gotterup",
		},
	}

	result, err := svc.UpsertTournament(ctx, bdlTournament, seasonEntity)
	require.NoError(t, err)
	assert.True(t, result.Created)

	tournamentWithChampion, err := svc.db.Tournament.Query().
		Where(tournament.BdlID(7)).
		WithChampion().
		Only(ctx)
	require.NoError(t, err)
	require.NotNil(t, tournamentWithChampion.Edges.Champion, "Champion should be set on create")
	assert.Equal(t, golferEntity.ID, tournamentWithChampion.Edges.Champion.ID)
}

func TestUpsertTournament_UpdatesChampion(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	golferEntity := testutil.CreateGolfer(t, svc.db, "Chris Gotterup", 63)
	seasonEntity := testutil.CreateSeason(t, svc.db, 2026)

	bdlTournament := &balldontlie.Tournament{
		ID:        7,
		Season:    2026,
		Name:      "Sony Open in Hawaii",
		StartDate: "2026-01-15T00:00:00.000Z",
		EndDate:   strPtr("2026-01-18T00:00:00.000Z"),
		Status:    strPtr("IN_PROGRESS"),
		Champion:  nil,
	}

	result, err := svc.UpsertTournament(ctx, bdlTournament, seasonEntity)
	require.NoError(t, err)
	assert.True(t, result.Created)

	tournamentNoChampion, err := svc.db.Tournament.Query().
		Where(tournament.BdlID(7)).
		WithChampion().
		Only(ctx)
	require.NoError(t, err)
	assert.Nil(t, tournamentNoChampion.Edges.Champion, "Champion should be nil initially")

	bdlTournament.Status = strPtr("COMPLETED")
	bdlTournament.Champion = &balldontlie.Player{
		ID:          63,
		DisplayName: "Chris Gotterup",
	}

	result, err = svc.UpsertTournament(ctx, bdlTournament, seasonEntity)
	require.NoError(t, err)
	assert.False(t, result.Created)
	assert.True(t, result.BecameCompleted, "BecameCompleted should be true when champion is set")

	tournamentWithChampion, err := svc.db.Tournament.Query().
		Where(tournament.BdlID(7)).
		WithChampion().
		Only(ctx)
	require.NoError(t, err)
	require.NotNil(t, tournamentWithChampion.Edges.Champion, "Champion should be set on update")
	assert.Equal(t, golferEntity.ID, tournamentWithChampion.Edges.Champion.ID)
}

func TestUpsertTournament_ChampionGolferNotFound(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	seasonEntity := testutil.CreateSeason(t, svc.db, 2026)

	bdlTournament := &balldontlie.Tournament{
		ID:        7,
		Season:    2026,
		Name:      "Sony Open in Hawaii",
		StartDate: "2026-01-15T00:00:00.000Z",
		EndDate:   strPtr("2026-01-18T00:00:00.000Z"),
		Status:    strPtr("COMPLETED"),
		Champion: &balldontlie.Player{
			ID:          9999, // Golfer doesn't exist
			DisplayName: "Unknown Player",
		},
	}

	result, err := svc.UpsertTournament(ctx, bdlTournament, seasonEntity)
	require.NoError(t, err)
	assert.True(t, result.Created)

	tournamentNoChampion, err := svc.db.Tournament.Query().
		Where(tournament.BdlID(7)).
		WithChampion().
		Only(ctx)
	require.NoError(t, err)
	assert.Nil(t, tournamentNoChampion.Edges.Champion, "Champion should be nil when golfer not found")
}

func TestUpsertTournament_UpdatesExistingTournamentWithChampion(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	golferEntity := testutil.CreateGolfer(t, svc.db, "Chris Gotterup", 63)
	seasonEntity := testutil.CreateSeason(t, svc.db, 2026)

	startDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 19, 6, 0, 0, 0, time.UTC)
	existingTournament, err := svc.db.Tournament.Create().
		SetBdlID(7).
		SetName("Sony Open in Hawaii").
		SetStartDate(startDate).
		SetEndDate(endDate).
		SetSeasonYear(2026).
		SetSeason(seasonEntity).
		Save(ctx)
	require.NoError(t, err)
	t.Logf("Created tournament ID: %s, BdlID: %v", existingTournament.ID, existingTournament.BdlID)

	tournamentBefore, err := svc.db.Tournament.Query().
		Where(tournament.BdlID(7)).
		WithChampion().
		Only(ctx)
	require.NoError(t, err)
	assert.Nil(t, tournamentBefore.Edges.Champion, "Champion should be nil before upsert")

	bdlTournament := &balldontlie.Tournament{
		ID:        7,
		Season:    2026,
		Name:      "Sony Open in Hawaii",
		StartDate: "2026-01-15T00:00:00.000Z",
		EndDate:   strPtr("2026-01-18T00:00:00.000Z"),
		Status:    strPtr("COMPLETED"),
		Champion: &balldontlie.Player{
			ID:          63,
			DisplayName: "Chris Gotterup",
		},
	}

	result, err := svc.UpsertTournament(ctx, bdlTournament, seasonEntity)
	require.NoError(t, err)
	assert.False(t, result.Created, "Should be update, not create")
	assert.True(t, result.BecameCompleted, "BecameCompleted should be true")

	tournamentAfter, err := svc.db.Tournament.Query().
		Where(tournament.BdlID(7)).
		WithChampion().
		Only(ctx)
	require.NoError(t, err)
	require.NotNil(t, tournamentAfter.Edges.Champion, "Champion should be set after upsert")
	assert.Equal(t, golferEntity.ID, tournamentAfter.Edges.Champion.ID)
}

func TestUpsertTournament_PreservesExistingChampion(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	golferEntity := testutil.CreateGolfer(t, svc.db, "Chris Gotterup", 63)
	seasonEntity := testutil.CreateSeason(t, svc.db, 2026)

	bdlTournament := &balldontlie.Tournament{
		ID:        7,
		Season:    2026,
		Name:      "Sony Open in Hawaii",
		StartDate: "2026-01-15T00:00:00.000Z",
		EndDate:   strPtr("2026-01-18T00:00:00.000Z"),
		Status:    strPtr("COMPLETED"),
		Champion: &balldontlie.Player{
			ID:          63,
			DisplayName: "Chris Gotterup",
		},
	}

	_, err := svc.UpsertTournament(ctx, bdlTournament, seasonEntity)
	require.NoError(t, err)

	result, err := svc.UpsertTournament(ctx, bdlTournament, seasonEntity)
	require.NoError(t, err)
	assert.False(t, result.BecameCompleted, "BecameCompleted should be false when already had champion")

	tournamentWithChampion, err := svc.db.Tournament.Query().
		Where(tournament.BdlID(7)).
		WithChampion().
		Only(ctx)
	require.NoError(t, err)
	require.NotNil(t, tournamentWithChampion.Edges.Champion)
	assert.Equal(t, golferEntity.ID, tournamentWithChampion.Edges.Champion.ID)
}

func float64Ptr(f float64) *float64 {
	return &f
}

func timeMustParse(layout, value string) time.Time {
	t, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}
	return t
}
