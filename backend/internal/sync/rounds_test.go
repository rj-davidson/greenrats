package sync

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/ent/golfer"
	"github.com/rj-davidson/greenrats/ent/holescore"
	"github.com/rj-davidson/greenrats/ent/round"
	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/internal/external/balldontlie"
	"github.com/rj-davidson/greenrats/internal/testutil"
)

func TestUpsertRound_Create(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	tournamentEntity := testutil.CreateTournament(t, svc.db, "Masters", 2026)
	golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 1)

	bdlRound := &balldontlie.PlayerRoundResult{
		Tournament:       balldontlie.TournamentRef{ID: 1},
		Player:           balldontlie.Player{ID: 1, DisplayName: "Scottie Scheffler"},
		RoundNumber:      1,
		Score:            68,
		ParRelativeScore: -4,
	}

	created, err := svc.UpsertRound(ctx, tournamentEntity.ID, golferEntity.ID, bdlRound)

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, 1, created.RoundNumber)
	assert.Equal(t, 68, *created.Score)
	assert.Equal(t, -4, *created.ParRelativeScore)
}

func TestUpsertRound_Update(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	tournamentEntity := testutil.CreateTournament(t, svc.db, "Masters", 2026)
	golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 1)

	bdlRound := &balldontlie.PlayerRoundResult{
		RoundNumber:      1,
		Score:            68,
		ParRelativeScore: -4,
	}

	_, err := svc.UpsertRound(ctx, tournamentEntity.ID, golferEntity.ID, bdlRound)
	require.NoError(t, err)

	bdlRound.Score = 70
	bdlRound.ParRelativeScore = -2

	updated, err := svc.UpsertRound(ctx, tournamentEntity.ID, golferEntity.ID, bdlRound)

	require.NoError(t, err)
	assert.Equal(t, 70, *updated.Score)
	assert.Equal(t, -2, *updated.ParRelativeScore)
}

func TestUpsertHoleScore_Create(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	tournamentEntity := testutil.CreateTournament(t, svc.db, "Masters", 2026)
	golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 1)

	bdlRoundResult := &balldontlie.PlayerRoundResult{RoundNumber: 1}
	roundRecord, err := svc.UpsertRound(ctx, tournamentEntity.ID, golferEntity.ID, bdlRoundResult)
	require.NoError(t, err)

	bdlScorecard := &balldontlie.PlayerScorecard{
		RoundNumber: 1,
		HoleNumber:  12,
		Par:         3,
		Score:       intPtr(2),
	}

	err = svc.UpsertHoleScore(ctx, roundRecord.ID, bdlScorecard)
	require.NoError(t, err)

	scores, err := svc.db.HoleScore.Query().
		Where(holescore.HasRoundWith(round.IDEQ(roundRecord.ID))).
		All(ctx)

	require.NoError(t, err)
	require.Len(t, scores, 1)
	assert.Equal(t, 12, scores[0].HoleNumber)
	assert.Equal(t, 3, scores[0].Par)
	assert.Equal(t, 2, *scores[0].Score)
}

func TestUpsertHoleScore_Update(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	tournamentEntity := testutil.CreateTournament(t, svc.db, "Masters", 2026)
	golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 1)

	bdlRoundResult := &balldontlie.PlayerRoundResult{RoundNumber: 1}
	roundRecord, err := svc.UpsertRound(ctx, tournamentEntity.ID, golferEntity.ID, bdlRoundResult)
	require.NoError(t, err)

	bdlScorecard := &balldontlie.PlayerScorecard{
		HoleNumber: 1,
		Par:        4,
		Score:      intPtr(4),
	}

	err = svc.UpsertHoleScore(ctx, roundRecord.ID, bdlScorecard)
	require.NoError(t, err)

	bdlScorecard.Score = intPtr(3)

	err = svc.UpsertHoleScore(ctx, roundRecord.ID, bdlScorecard)
	require.NoError(t, err)

	scores, err := svc.db.HoleScore.Query().
		Where(holescore.HasRoundWith(round.IDEQ(roundRecord.ID))).
		All(ctx)

	require.NoError(t, err)
	require.Len(t, scores, 1)
	assert.Equal(t, 3, *scores[0].Score)
}

func TestUpsertRound_VerifiesDirectEdges(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	tournamentEntity := testutil.CreateTournament(t, svc.db, "Masters", 2026)
	golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 1)

	bdlRound := &balldontlie.PlayerRoundResult{
		RoundNumber:      1,
		Score:            68,
		ParRelativeScore: -4,
	}

	created, err := svc.UpsertRound(ctx, tournamentEntity.ID, golferEntity.ID, bdlRound)
	require.NoError(t, err)

	roundWithEdges, err := svc.db.Round.Query().
		Where(round.IDEQ(created.ID)).
		WithTournament().
		WithGolfer().
		Only(ctx)
	require.NoError(t, err)

	assert.NotNil(t, roundWithEdges.Edges.Tournament)
	assert.Equal(t, tournamentEntity.ID, roundWithEdges.Edges.Tournament.ID)
	assert.NotNil(t, roundWithEdges.Edges.Golfer)
	assert.Equal(t, golferEntity.ID, roundWithEdges.Edges.Golfer.ID)
}

func TestUpsertRound_UniqueConstraint(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	tournamentEntity := testutil.CreateTournament(t, svc.db, "Masters", 2026)
	golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 1)

	bdlRound := &balldontlie.PlayerRoundResult{
		RoundNumber: 1,
	}

	_, err := svc.UpsertRound(ctx, tournamentEntity.ID, golferEntity.ID, bdlRound)
	require.NoError(t, err)

	_, err = svc.UpsertRound(ctx, tournamentEntity.ID, golferEntity.ID, bdlRound)
	require.NoError(t, err)

	count, err := svc.db.Round.Query().
		Where(
			round.HasTournamentWith(tournament.IDEQ(tournamentEntity.ID)),
			round.HasGolferWith(golfer.IDEQ(golferEntity.ID)),
			round.RoundNumberEQ(1),
		).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "should only create one round for same tournament/golfer/round_number")
}

func TestUpsertHoleScore_UpdatesThru(t *testing.T) {
	t.Run("updates thru after adding hole score", func(t *testing.T) {
		ctx := context.Background()
		svc := newTestService(t)

		tournamentEntity := testutil.CreateTournament(t, svc.db, "Masters", 2026)
		golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 1)

		bdlRound := &balldontlie.PlayerRoundResult{RoundNumber: 1}
		roundRecord, err := svc.UpsertRound(ctx, tournamentEntity.ID, golferEntity.ID, bdlRound)
		require.NoError(t, err)

		err = svc.UpsertHoleScore(ctx, roundRecord.ID, &balldontlie.PlayerScorecard{
			HoleNumber: 1, Par: 4, Score: intPtr(4),
		})
		require.NoError(t, err)

		roundEntity, err := svc.db.Round.Get(ctx, roundRecord.ID)
		require.NoError(t, err)
		require.NotNil(t, roundEntity.Thru)
		assert.Equal(t, 1, *roundEntity.Thru)

		err = svc.UpsertHoleScore(ctx, roundRecord.ID, &balldontlie.PlayerScorecard{
			HoleNumber: 2, Par: 5, Score: intPtr(5),
		})
		require.NoError(t, err)

		roundEntity, err = svc.db.Round.Get(ctx, roundRecord.ID)
		require.NoError(t, err)
		assert.Equal(t, 2, *roundEntity.Thru)
	})

	t.Run("thru only counts holes with scores", func(t *testing.T) {
		ctx := context.Background()
		svc := newTestService(t)

		tournamentEntity := testutil.CreateTournament(t, svc.db, "Masters", 2026)
		golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 1)

		bdlRound := &balldontlie.PlayerRoundResult{RoundNumber: 1}
		roundRecord, err := svc.UpsertRound(ctx, tournamentEntity.ID, golferEntity.ID, bdlRound)
		require.NoError(t, err)

		err = svc.UpsertHoleScore(ctx, roundRecord.ID, &balldontlie.PlayerScorecard{
			HoleNumber: 1, Par: 4, Score: intPtr(4),
		})
		require.NoError(t, err)

		err = svc.UpsertHoleScore(ctx, roundRecord.ID, &balldontlie.PlayerScorecard{
			HoleNumber: 2, Par: 5, Score: nil,
		})
		require.NoError(t, err)

		roundEntity, err := svc.db.Round.Get(ctx, roundRecord.ID)
		require.NoError(t, err)
		require.NotNil(t, roundEntity.Thru)
		assert.Equal(t, 1, *roundEntity.Thru, "thru should only count holes with non-nil scores")
	})

	t.Run("thru counts all scored holes regardless of order", func(t *testing.T) {
		ctx := context.Background()
		svc := newTestService(t)

		tournamentEntity := testutil.CreateTournament(t, svc.db, "Masters", 2026)
		golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 1)

		bdlRound := &balldontlie.PlayerRoundResult{RoundNumber: 1}
		roundRecord, err := svc.UpsertRound(ctx, tournamentEntity.ID, golferEntity.ID, bdlRound)
		require.NoError(t, err)

		err = svc.UpsertHoleScore(ctx, roundRecord.ID, &balldontlie.PlayerScorecard{
			HoleNumber: 5, Par: 4, Score: intPtr(4),
		})
		require.NoError(t, err)

		err = svc.UpsertHoleScore(ctx, roundRecord.ID, &balldontlie.PlayerScorecard{
			HoleNumber: 1, Par: 4, Score: intPtr(3),
		})
		require.NoError(t, err)

		err = svc.UpsertHoleScore(ctx, roundRecord.ID, &balldontlie.PlayerScorecard{
			HoleNumber: 3, Par: 3, Score: intPtr(2),
		})
		require.NoError(t, err)

		roundEntity, err := svc.db.Round.Get(ctx, roundRecord.ID)
		require.NoError(t, err)
		require.NotNil(t, roundEntity.Thru)
		assert.Equal(t, 3, *roundEntity.Thru, "thru should count all holes with scores")
	})
}

func TestUpsertHoleScore_UpdatesRoundScores(t *testing.T) {
	t.Run("computes score and par_relative_score from holes", func(t *testing.T) {
		ctx := context.Background()
		svc := newTestService(t)

		tournamentEntity := testutil.CreateTournament(t, svc.db, "Masters", 2026)
		golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 1)

		bdlRound := &balldontlie.PlayerRoundResult{RoundNumber: 1}
		roundRecord, err := svc.UpsertRound(ctx, tournamentEntity.ID, golferEntity.ID, bdlRound)
		require.NoError(t, err)

		err = svc.UpsertHoleScore(ctx, roundRecord.ID, &balldontlie.PlayerScorecard{
			HoleNumber: 1, Par: 4, Score: intPtr(3),
		})
		require.NoError(t, err)

		err = svc.UpsertHoleScore(ctx, roundRecord.ID, &balldontlie.PlayerScorecard{
			HoleNumber: 2, Par: 5, Score: intPtr(6),
		})
		require.NoError(t, err)

		roundEntity, err := svc.db.Round.Get(ctx, roundRecord.ID)
		require.NoError(t, err)

		require.NotNil(t, roundEntity.Score)
		assert.Equal(t, 9, *roundEntity.Score)

		require.NotNil(t, roundEntity.ParRelativeScore)
		assert.Equal(t, 0, *roundEntity.ParRelativeScore)
	})

	t.Run("ignores holes without scores in calculation", func(t *testing.T) {
		ctx := context.Background()
		svc := newTestService(t)

		tournamentEntity := testutil.CreateTournament(t, svc.db, "Masters", 2026)
		golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 1)

		bdlRound := &balldontlie.PlayerRoundResult{RoundNumber: 1}
		roundRecord, err := svc.UpsertRound(ctx, tournamentEntity.ID, golferEntity.ID, bdlRound)
		require.NoError(t, err)

		err = svc.UpsertHoleScore(ctx, roundRecord.ID, &balldontlie.PlayerScorecard{
			HoleNumber: 1, Par: 4, Score: intPtr(4),
		})
		require.NoError(t, err)

		err = svc.UpsertHoleScore(ctx, roundRecord.ID, &balldontlie.PlayerScorecard{
			HoleNumber: 2, Par: 5, Score: nil,
		})
		require.NoError(t, err)

		roundEntity, err := svc.db.Round.Get(ctx, roundRecord.ID)
		require.NoError(t, err)

		require.NotNil(t, roundEntity.Score)
		assert.Equal(t, 4, *roundEntity.Score)

		require.NotNil(t, roundEntity.ParRelativeScore)
		assert.Equal(t, 0, *roundEntity.ParRelativeScore)
	})
}
