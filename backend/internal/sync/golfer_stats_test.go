package sync

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/ent/golfer"
	"github.com/rj-davidson/greenrats/ent/golferseason"
	"github.com/rj-davidson/greenrats/internal/external/balldontlie"
	"github.com/rj-davidson/greenrats/internal/testutil"
)

func statValue(v any) []balldontlie.StatValueItem {
	return []balldontlie.StatValueItem{{StatValue: fmt.Sprintf("%v", v)}}
}

func TestUpsertGolferSeasonStat_Create(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 1)
	season := testutil.CreateSeason(t, svc.db, 2026)

	stat := &balldontlie.PlayerSeasonStat{
		StatName:  "Scoring Average",
		StatValue: statValue(68.5),
	}

	err := svc.UpsertGolferSeasonStat(ctx, golferEntity.ID, season.ID, stat)
	require.NoError(t, err)

	gs, err := svc.db.GolferSeason.Query().
		Where(golferseason.HasGolferWith(golfer.IDEQ(golferEntity.ID))).
		Only(ctx)

	require.NoError(t, err)
	assert.Equal(t, 68.5, *gs.ScoringAvg)
}

func TestUpsertGolferSeasonStat_Update(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 1)
	season := testutil.CreateSeason(t, svc.db, 2026)

	stat := &balldontlie.PlayerSeasonStat{
		StatName:  "Scoring Average",
		StatValue: statValue(68.5),
	}

	err := svc.UpsertGolferSeasonStat(ctx, golferEntity.ID, season.ID, stat)
	require.NoError(t, err)

	stat.StatName = "Driving Distance"
	stat.StatValue = statValue(310.5)

	err = svc.UpsertGolferSeasonStat(ctx, golferEntity.ID, season.ID, stat)
	require.NoError(t, err)

	gs, err := svc.db.GolferSeason.Query().
		Where(golferseason.HasGolferWith(golfer.IDEQ(golferEntity.ID))).
		Only(ctx)

	require.NoError(t, err)
	assert.Equal(t, 68.5, *gs.ScoringAvg)
	assert.Equal(t, 310.5, *gs.DrivingDistance)
}

func TestUpsertGolferSeasonStat_AllStats(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 1)
	season := testutil.CreateSeason(t, svc.db, 2026)

	stats := []balldontlie.PlayerSeasonStat{
		{StatName: "Scoring Average", StatValue: statValue(68.5)},
		{StatName: "Top 10 Finishes", StatValue: statValue(10)},
		{StatName: "Cuts Made", StatValue: statValue(18)},
		{StatName: "Events Played", StatValue: statValue(20)},
		{StatName: "Wins", StatValue: statValue(4)},
		{StatName: "Official Money", StatValue: statValue("$15,000,000")},
		{StatName: "Driving Distance", StatValue: statValue(310.5)},
		{StatName: "Driving Accuracy Percentage", StatValue: statValue(62.5)},
		{StatName: "Greens in Regulation Percentage", StatValue: statValue(70.2)},
		{StatName: "Putting Average", StatValue: statValue(1.72)},
		{StatName: "Scrambling", StatValue: statValue(65.8)},
	}

	for _, stat := range stats {
		err := svc.UpsertGolferSeasonStat(ctx, golferEntity.ID, season.ID, &stat)
		require.NoError(t, err)
	}

	gs, err := svc.db.GolferSeason.Query().
		Where(golferseason.HasGolferWith(golfer.IDEQ(golferEntity.ID))).
		Only(ctx)

	require.NoError(t, err)
	assert.Equal(t, 68.5, *gs.ScoringAvg)
	assert.Equal(t, 10, *gs.Top10s)
	assert.Equal(t, 18, *gs.CutsMade)
	assert.Equal(t, 20, *gs.EventsPlayed)
	assert.Equal(t, 4, *gs.Wins)
	assert.Equal(t, 15000000, *gs.Earnings)
	assert.Equal(t, 310.5, *gs.DrivingDistance)
	assert.Equal(t, 62.5, *gs.DrivingAccuracy)
	assert.Equal(t, 70.2, *gs.GirPct)
	assert.Equal(t, 1.72, *gs.PuttingAvg)
	assert.Equal(t, 65.8, *gs.ScramblingPct)
	assert.NotNil(t, gs.LastSyncedAt)
}
