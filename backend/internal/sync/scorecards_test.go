package sync

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/ent/holescore"
	"github.com/rj-davidson/greenrats/ent/round"
	"github.com/rj-davidson/greenrats/internal/external/balldontlie"
	"github.com/rj-davidson/greenrats/internal/testutil"
)

func newTestIngester(t *testing.T) *Ingester {
	t.Helper()
	db := testutil.NewTestDB(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	syncService := NewService(db, nil, logger)
	return &Ingester{
		db:          db,
		syncService: syncService,
		logger:      logger,
	}
}

func TestProcessScorecards_CreatesRoundForInProgressRound(t *testing.T) {
	ctx := context.Background()
	ing := newTestIngester(t)

	tournament := testutil.CreateTournament(t, ing.db, "Masters", 2026)
	golfer := testutil.CreateGolfer(t, ing.db, "Scottie Scheffler", 1)
	entry := testutil.CreateLeaderboardEntry(t, ing.db, tournament.ID, golfer.ID)

	scorecards := []balldontlie.PlayerScorecard{
		{
			Tournament:  balldontlie.TournamentRef{ID: 1},
			Player:      balldontlie.Player{ID: 1, DisplayName: "Scottie Scheffler"},
			RoundNumber: 3,
			HoleNumber:  1,
			Par:         4,
			Score:       intPtr(4),
		},
	}

	holesProcessed, processedEntries := ing.processScorecards(ctx, tournament, scorecards)

	require.Len(t, processedEntries, 1)
	assert.True(t, processedEntries[entry.ID])
	assert.Equal(t, 1, holesProcessed)

	rounds, err := ing.db.Round.Query().
		Where(round.HasLeaderboardEntryWith()).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, rounds, 1)
	assert.Equal(t, 3, rounds[0].RoundNumber)
	assert.Nil(t, rounds[0].Score)
	assert.Nil(t, rounds[0].ParRelativeScore)

	holeScores, err := ing.db.HoleScore.Query().
		Where(holescore.HasRoundWith(round.IDEQ(rounds[0].ID))).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, holeScores, 1)
	assert.Equal(t, 1, holeScores[0].HoleNumber)
	assert.Equal(t, 4, holeScores[0].Par)
	assert.Equal(t, 4, *holeScores[0].Score)
}

func TestProcessScorecards_ReusesCreatedRoundForMultipleHoles(t *testing.T) {
	ctx := context.Background()
	ing := newTestIngester(t)

	tournament := testutil.CreateTournament(t, ing.db, "Masters", 2026)
	golfer := testutil.CreateGolfer(t, ing.db, "Scottie Scheffler", 1)
	testutil.CreateLeaderboardEntry(t, ing.db, tournament.ID, golfer.ID)

	scorecards := []balldontlie.PlayerScorecard{
		{
			Player:      balldontlie.Player{ID: 1, DisplayName: "Scottie Scheffler"},
			RoundNumber: 3,
			HoleNumber:  1,
			Par:         4,
			Score:       intPtr(4),
		},
		{
			Player:      balldontlie.Player{ID: 1, DisplayName: "Scottie Scheffler"},
			RoundNumber: 3,
			HoleNumber:  2,
			Par:         5,
			Score:       intPtr(4),
		},
		{
			Player:      balldontlie.Player{ID: 1, DisplayName: "Scottie Scheffler"},
			RoundNumber: 3,
			HoleNumber:  3,
			Par:         3,
			Score:       intPtr(3),
		},
	}

	holesProcessed, _ := ing.processScorecards(ctx, tournament, scorecards)

	assert.Equal(t, 3, holesProcessed)

	rounds, err := ing.db.Round.Query().All(ctx)
	require.NoError(t, err)
	require.Len(t, rounds, 1, "should create only one round for all holes in the same round")

	holeScores, err := ing.db.HoleScore.Query().
		Where(holescore.HasRoundWith(round.IDEQ(rounds[0].ID))).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, holeScores, 3)
}

func TestProcessScorecards_UsesExistingRound(t *testing.T) {
	ctx := context.Background()
	ing := newTestIngester(t)

	tournament := testutil.CreateTournament(t, ing.db, "Masters", 2026)
	golfer := testutil.CreateGolfer(t, ing.db, "Scottie Scheffler", 1)
	entry := testutil.CreateLeaderboardEntry(t, ing.db, tournament.ID, golfer.ID)

	bdlRound := &balldontlie.PlayerRoundResult{
		RoundNumber:      2,
		Score:            intPtr(68),
		ParRelativeScore: intPtr(-4),
	}
	existingRound, err := ing.syncService.UpsertRound(ctx, entry.ID, bdlRound)
	require.NoError(t, err)

	scorecards := []balldontlie.PlayerScorecard{
		{
			Player:      balldontlie.Player{ID: 1, DisplayName: "Scottie Scheffler"},
			RoundNumber: 2,
			HoleNumber:  18,
			Par:         4,
			Score:       intPtr(3),
		},
	}

	holesProcessed, _ := ing.processScorecards(ctx, tournament, scorecards)

	assert.Equal(t, 1, holesProcessed)

	rounds, err := ing.db.Round.Query().All(ctx)
	require.NoError(t, err)
	require.Len(t, rounds, 1, "should not create a new round")
	assert.Equal(t, existingRound.ID, rounds[0].ID)
	assert.Equal(t, 68, *rounds[0].Score)

	holeScores, err := ing.db.HoleScore.Query().
		Where(holescore.HasRoundWith(round.IDEQ(existingRound.ID))).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, holeScores, 1)
}

func TestProcessScorecards_HandlesMultipleRounds(t *testing.T) {
	ctx := context.Background()
	ing := newTestIngester(t)

	tournament := testutil.CreateTournament(t, ing.db, "Masters", 2026)
	golfer := testutil.CreateGolfer(t, ing.db, "Scottie Scheffler", 1)
	entry := testutil.CreateLeaderboardEntry(t, ing.db, tournament.ID, golfer.ID)

	bdlRound := &balldontlie.PlayerRoundResult{
		RoundNumber:      1,
		Score:            intPtr(70),
		ParRelativeScore: intPtr(-2),
	}
	_, err := ing.syncService.UpsertRound(ctx, entry.ID, bdlRound)
	require.NoError(t, err)

	scorecards := []balldontlie.PlayerScorecard{
		{
			Player:      balldontlie.Player{ID: 1, DisplayName: "Scottie Scheffler"},
			RoundNumber: 1,
			HoleNumber:  18,
			Par:         4,
			Score:       intPtr(4),
		},
		{
			Player:      balldontlie.Player{ID: 1, DisplayName: "Scottie Scheffler"},
			RoundNumber: 2,
			HoleNumber:  1,
			Par:         4,
			Score:       intPtr(3),
		},
	}

	holesProcessed, _ := ing.processScorecards(ctx, tournament, scorecards)

	assert.Equal(t, 2, holesProcessed)

	rounds, err := ing.db.Round.Query().All(ctx)
	require.NoError(t, err)
	assert.Len(t, rounds, 2, "should have round 1 (existing) and round 2 (created)")
}
