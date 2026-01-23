package sync

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/ent/holescore"
	"github.com/rj-davidson/greenrats/ent/round"
	"github.com/rj-davidson/greenrats/internal/external/balldontlie"
	"github.com/rj-davidson/greenrats/internal/testutil"
)

func TestUpsertRound_Create(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	tournament := testutil.CreateTournament(t, svc.db, "Masters", 2026)
	golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 1)
	entry := testutil.CreateLeaderboardEntry(t, svc.db, tournament.ID, golferEntity.ID)

	bdlRound := &balldontlie.PlayerRoundResult{
		Tournament:       balldontlie.TournamentRef{ID: 1},
		Player:           balldontlie.Player{ID: 1, DisplayName: "Scottie Scheffler"},
		RoundNumber:      1,
		Score:            intPtr(68),
		ParRelativeScore: intPtr(-4),
	}

	created, err := svc.UpsertRound(ctx, entry.ID, bdlRound)

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, 1, created.RoundNumber)
	assert.Equal(t, 68, *created.Score)
	assert.Equal(t, -4, *created.ParRelativeScore)
}

func TestUpsertRound_Update(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	tournament := testutil.CreateTournament(t, svc.db, "Masters", 2026)
	golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 1)
	entry := testutil.CreateLeaderboardEntry(t, svc.db, tournament.ID, golferEntity.ID)

	bdlRound := &balldontlie.PlayerRoundResult{
		RoundNumber:      1,
		Score:            intPtr(68),
		ParRelativeScore: intPtr(-4),
	}

	_, err := svc.UpsertRound(ctx, entry.ID, bdlRound)
	require.NoError(t, err)

	bdlRound.Score = intPtr(70)
	bdlRound.ParRelativeScore = intPtr(-2)

	updated, err := svc.UpsertRound(ctx, entry.ID, bdlRound)

	require.NoError(t, err)
	assert.Equal(t, 70, *updated.Score)
	assert.Equal(t, -2, *updated.ParRelativeScore)
}

func TestUpsertHoleScore_Create(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	tournament := testutil.CreateTournament(t, svc.db, "Masters", 2026)
	golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 1)
	entry := testutil.CreateLeaderboardEntry(t, svc.db, tournament.ID, golferEntity.ID)

	bdlRoundResult := &balldontlie.PlayerRoundResult{RoundNumber: 1}
	roundRecord, err := svc.UpsertRound(ctx, entry.ID, bdlRoundResult)
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

	tournament := testutil.CreateTournament(t, svc.db, "Masters", 2026)
	golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 1)
	entry := testutil.CreateLeaderboardEntry(t, svc.db, tournament.ID, golferEntity.ID)

	bdlRoundResult := &balldontlie.PlayerRoundResult{RoundNumber: 1}
	roundRecord, err := svc.UpsertRound(ctx, entry.ID, bdlRoundResult)
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
