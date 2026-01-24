package sync

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/ent/golfer"
	"github.com/rj-davidson/greenrats/ent/leaderboardentry"
	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/internal/external/balldontlie"
	"github.com/rj-davidson/greenrats/internal/testutil"
)

func TestUpsertLeaderboardEntry_SetsChampionPosition(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 100)
	tournamentEntity := testutil.CreateTournament(t, svc.db, "The Masters", 2026)

	completedStatus := "COMPLETED"
	position := "1"
	positionNumeric := 1
	totalScore := 268
	parRelative := -20
	earnings := 3600000.0

	result := &balldontlie.TournamentResult{
		Tournament: balldontlie.TournamentRef{
			ID:     1,
			Status: &completedStatus,
		},
		Player: balldontlie.Player{
			ID:          100,
			DisplayName: "Scottie Scheffler",
		},
		Position:         &position,
		PositionNumeric:  &positionNumeric,
		TotalScore:       &totalScore,
		ParRelativeScore: &parRelative,
		Earnings:         &earnings,
	}

	err := svc.UpsertLeaderboardEntry(ctx, tournamentEntity, result)
	require.NoError(t, err)

	entry, err := svc.db.LeaderboardEntry.Query().
		Where(
			leaderboardentry.HasTournamentWith(tournament.IDEQ(tournamentEntity.ID)),
			leaderboardentry.HasGolferWith(golfer.IDEQ(golferEntity.ID)),
		).
		Only(ctx)
	require.NoError(t, err)

	assert.Equal(t, 1, entry.Position, "Champion should have position 1")
	assert.Equal(t, 268, entry.Score, "Score should match")
	assert.Equal(t, 3600000, entry.Earnings, "Earnings should match")
	assert.Equal(t, leaderboardentry.StatusFinished, entry.Status, "Status should be finished for completed tournament")
	assert.False(t, entry.Cut, "Champion should not have cut flag")
}

func TestUpsertLeaderboardEntry_SetsMultiplePositions(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	champion := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 100)
	runnerUp := testutil.CreateGolfer(t, svc.db, "Collin Morikawa", 101)
	third := testutil.CreateGolfer(t, svc.db, "Rory McIlroy", 102)
	tournamentEntity := testutil.CreateTournament(t, svc.db, "The Masters", 2026)

	completedStatus := "COMPLETED"

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
		pos := r.position
		posStr := string(rune('0' + pos))
		earn := r.earnings
		result := &balldontlie.TournamentResult{
			Tournament: balldontlie.TournamentRef{
				ID:     1,
				Status: &completedStatus,
			},
			Player:          *r.golfer,
			Position:        &posStr,
			PositionNumeric: &pos,
			Earnings:        &earn,
		}
		err := svc.UpsertLeaderboardEntry(ctx, tournamentEntity, result)
		require.NoError(t, err)
	}

	championEntry, err := svc.db.LeaderboardEntry.Query().
		Where(
			leaderboardentry.HasTournamentWith(tournament.IDEQ(tournamentEntity.ID)),
			leaderboardentry.HasGolferWith(golfer.IDEQ(champion.ID)),
		).
		Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, championEntry.Position, "Champion should be position 1")
	assert.Equal(t, 3600000, championEntry.Earnings)

	runnerUpEntry, err := svc.db.LeaderboardEntry.Query().
		Where(
			leaderboardentry.HasTournamentWith(tournament.IDEQ(tournamentEntity.ID)),
			leaderboardentry.HasGolferWith(golfer.IDEQ(runnerUp.ID)),
		).
		Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, runnerUpEntry.Position, "Runner-up should be position 2")
	assert.Equal(t, 2160000, runnerUpEntry.Earnings)

	thirdEntry, err := svc.db.LeaderboardEntry.Query().
		Where(
			leaderboardentry.HasTournamentWith(tournament.IDEQ(tournamentEntity.ID)),
			leaderboardentry.HasGolferWith(golfer.IDEQ(third.ID)),
		).
		Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, thirdEntry.Position, "Third place should be position 3")
	assert.Equal(t, 1360000, thirdEntry.Earnings)
}

func TestUpsertLeaderboardEntry_SetsCutStatus(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	golferEntity := testutil.CreateGolfer(t, svc.db, "Tiger Woods", 200)
	tournamentEntity := testutil.CreateTournament(t, svc.db, "The Open", 2026)

	completedStatus := "COMPLETED"
	cutPosition := "CUT"

	result := &balldontlie.TournamentResult{
		Tournament: balldontlie.TournamentRef{
			ID:     2,
			Status: &completedStatus,
		},
		Player: balldontlie.Player{
			ID:          200,
			DisplayName: "Tiger Woods",
		},
		Position: &cutPosition,
	}

	err := svc.UpsertLeaderboardEntry(ctx, tournamentEntity, result)
	require.NoError(t, err)

	entry, err := svc.db.LeaderboardEntry.Query().
		Where(
			leaderboardentry.HasTournamentWith(tournament.IDEQ(tournamentEntity.ID)),
			leaderboardentry.HasGolferWith(golfer.IDEQ(golferEntity.ID)),
		).
		Only(ctx)
	require.NoError(t, err)

	assert.True(t, entry.Cut, "Player who missed cut should have cut flag set")
	assert.Equal(t, 0, entry.Position, "Cut players should have position 0")
	assert.Equal(t, leaderboardentry.StatusFinished, entry.Status, "Cut players should have finished status")
}

func TestUpsertLeaderboardEntry_UpdatesExistingEntry(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	golferEntity := testutil.CreateGolfer(t, svc.db, "Jon Rahm", 300)
	tournamentEntity := testutil.CreateTournament(t, svc.db, "US Open", 2026)

	inProgressStatus := "IN_PROGRESS"
	initialPos := 5
	initialPosStr := "5"

	initialResult := &balldontlie.TournamentResult{
		Tournament: balldontlie.TournamentRef{
			ID:     3,
			Status: &inProgressStatus,
		},
		Player: balldontlie.Player{
			ID:          300,
			DisplayName: "Jon Rahm",
		},
		Position:        &initialPosStr,
		PositionNumeric: &initialPos,
	}

	err := svc.UpsertLeaderboardEntry(ctx, tournamentEntity, initialResult)
	require.NoError(t, err)

	entry, err := svc.db.LeaderboardEntry.Query().
		Where(
			leaderboardentry.HasTournamentWith(tournament.IDEQ(tournamentEntity.ID)),
			leaderboardentry.HasGolferWith(golfer.IDEQ(golferEntity.ID)),
		).
		Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, 5, entry.Position)
	assert.Equal(t, leaderboardentry.StatusActive, entry.Status)

	completedStatus := "COMPLETED"
	finalPos := 1
	finalPosStr := "1"
	earnings := 4300000.0

	finalResult := &balldontlie.TournamentResult{
		Tournament: balldontlie.TournamentRef{
			ID:     3,
			Status: &completedStatus,
		},
		Player: balldontlie.Player{
			ID:          300,
			DisplayName: "Jon Rahm",
		},
		Position:        &finalPosStr,
		PositionNumeric: &finalPos,
		Earnings:        &earnings,
	}

	err = svc.UpsertLeaderboardEntry(ctx, tournamentEntity, finalResult)
	require.NoError(t, err)

	updatedEntry, err := svc.db.LeaderboardEntry.Query().
		Where(
			leaderboardentry.HasTournamentWith(tournament.IDEQ(tournamentEntity.ID)),
			leaderboardentry.HasGolferWith(golfer.IDEQ(golferEntity.ID)),
		).
		Only(ctx)
	require.NoError(t, err)

	assert.Equal(t, 1, updatedEntry.Position, "Position should be updated to 1 (champion)")
	assert.Equal(t, 4300000, updatedEntry.Earnings, "Earnings should be updated")
	assert.Equal(t, leaderboardentry.StatusFinished, updatedEntry.Status, "Status should be finished")
}

func TestUpsertLeaderboardEntry_SkipsUnknownGolfer(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	tournamentEntity := testutil.CreateTournament(t, svc.db, "PGA Championship", 2026)

	completedStatus := "COMPLETED"
	pos := 1
	posStr := "1"

	result := &balldontlie.TournamentResult{
		Tournament: balldontlie.TournamentRef{
			ID:     4,
			Status: &completedStatus,
		},
		Player: balldontlie.Player{
			ID:          99999,
			DisplayName: "Unknown Player",
		},
		Position:        &posStr,
		PositionNumeric: &pos,
	}

	err := svc.UpsertLeaderboardEntry(ctx, tournamentEntity, result)
	require.NoError(t, err, "Should not error for unknown golfer")

	count, err := svc.db.LeaderboardEntry.Query().
		Where(leaderboardentry.HasTournamentWith(tournament.IDEQ(tournamentEntity.ID))).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "No entry should be created for unknown golfer")
}

func TestUpdateLeaderboardProgress_PlayerMidRound(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	tournamentEntity := testutil.CreateTournament(t, svc.db, "The Masters", 2026)
	golferEntity := testutil.CreateGolfer(t, svc.db, "Scottie Scheffler", 100)
	entry := testutil.CreateLeaderboardEntry(t, svc.db, tournamentEntity.ID, golferEntity.ID)

	round1, err := svc.db.Round.Create().
		SetLeaderboardEntry(entry).
		SetRoundNumber(1).
		Save(ctx)
	require.NoError(t, err)

	for hole := 1; hole <= 9; hole++ {
		score := 4
		_, err := svc.db.HoleScore.Create().
			SetRound(round1).
			SetHoleNumber(hole).
			SetPar(4).
			SetScore(score).
			Save(ctx)
		require.NoError(t, err)
	}

	err = svc.UpdateLeaderboardProgress(ctx, entry.ID)
	require.NoError(t, err)

	updated, err := svc.db.LeaderboardEntry.Get(ctx, entry.ID)
	require.NoError(t, err)

	assert.Equal(t, 1, updated.CurrentRound, "Current round should be 1")
	assert.Equal(t, 9, updated.Thru, "Thru should be 9 holes")
}

func TestUpdateLeaderboardProgress_PlayerFinishedRound(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	tournamentEntity := testutil.CreateTournament(t, svc.db, "The Masters", 2026)
	golferEntity := testutil.CreateGolfer(t, svc.db, "Rory McIlroy", 101)
	entry := testutil.CreateLeaderboardEntry(t, svc.db, tournamentEntity.ID, golferEntity.ID)

	round1, err := svc.db.Round.Create().
		SetLeaderboardEntry(entry).
		SetRoundNumber(1).
		Save(ctx)
	require.NoError(t, err)

	for hole := 1; hole <= 18; hole++ {
		score := 4
		_, err := svc.db.HoleScore.Create().
			SetRound(round1).
			SetHoleNumber(hole).
			SetPar(4).
			SetScore(score).
			Save(ctx)
		require.NoError(t, err)
	}

	err = svc.UpdateLeaderboardProgress(ctx, entry.ID)
	require.NoError(t, err)

	updated, err := svc.db.LeaderboardEntry.Get(ctx, entry.ID)
	require.NoError(t, err)

	assert.Equal(t, 1, updated.CurrentRound, "Current round should be 1")
	assert.Equal(t, 18, updated.Thru, "Thru should be 18 (finished)")
}

func TestUpdateLeaderboardProgress_PlayerNotStarted(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	tournamentEntity := testutil.CreateTournament(t, svc.db, "The Masters", 2026)
	golferEntity := testutil.CreateGolfer(t, svc.db, "Tiger Woods", 102)
	entry := testutil.CreateLeaderboardEntry(t, svc.db, tournamentEntity.ID, golferEntity.ID)

	err := svc.UpdateLeaderboardProgress(ctx, entry.ID)
	require.NoError(t, err)

	updated, err := svc.db.LeaderboardEntry.Get(ctx, entry.ID)
	require.NoError(t, err)

	assert.Equal(t, 0, updated.CurrentRound, "Current round should be 0 (not started)")
	assert.Equal(t, 0, updated.Thru, "Thru should be 0 (not started)")
}

func TestUpdateLeaderboardProgress_PlayerInRound2(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	tournamentEntity := testutil.CreateTournament(t, svc.db, "The Masters", 2026)
	golferEntity := testutil.CreateGolfer(t, svc.db, "Jon Rahm", 103)
	entry := testutil.CreateLeaderboardEntry(t, svc.db, tournamentEntity.ID, golferEntity.ID)

	round1, err := svc.db.Round.Create().
		SetLeaderboardEntry(entry).
		SetRoundNumber(1).
		Save(ctx)
	require.NoError(t, err)

	for hole := 1; hole <= 18; hole++ {
		_, err := svc.db.HoleScore.Create().
			SetRound(round1).
			SetHoleNumber(hole).
			SetPar(4).
			SetScore(4).
			Save(ctx)
		require.NoError(t, err)
	}

	round2, err := svc.db.Round.Create().
		SetLeaderboardEntry(entry).
		SetRoundNumber(2).
		Save(ctx)
	require.NoError(t, err)

	for hole := 1; hole <= 5; hole++ {
		_, err := svc.db.HoleScore.Create().
			SetRound(round2).
			SetHoleNumber(hole).
			SetPar(4).
			SetScore(3).
			Save(ctx)
		require.NoError(t, err)
	}

	err = svc.UpdateLeaderboardProgress(ctx, entry.ID)
	require.NoError(t, err)

	updated, err := svc.db.LeaderboardEntry.Get(ctx, entry.ID)
	require.NoError(t, err)

	assert.Equal(t, 2, updated.CurrentRound, "Current round should be 2")
	assert.Equal(t, 5, updated.Thru, "Thru should be 5 holes in round 2")
}

func TestUpdateLeaderboardProgress_HolesWithNilScores(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	tournamentEntity := testutil.CreateTournament(t, svc.db, "The Masters", 2026)
	golferEntity := testutil.CreateGolfer(t, svc.db, "Collin Morikawa", 104)
	entry := testutil.CreateLeaderboardEntry(t, svc.db, tournamentEntity.ID, golferEntity.ID)

	round1, err := svc.db.Round.Create().
		SetLeaderboardEntry(entry).
		SetRoundNumber(1).
		Save(ctx)
	require.NoError(t, err)

	for hole := 1; hole <= 6; hole++ {
		_, err := svc.db.HoleScore.Create().
			SetRound(round1).
			SetHoleNumber(hole).
			SetPar(4).
			SetScore(4).
			Save(ctx)
		require.NoError(t, err)
	}

	for hole := 7; hole <= 18; hole++ {
		_, err := svc.db.HoleScore.Create().
			SetRound(round1).
			SetHoleNumber(hole).
			SetPar(4).
			Save(ctx)
		require.NoError(t, err)
	}

	err = svc.UpdateLeaderboardProgress(ctx, entry.ID)
	require.NoError(t, err)

	updated, err := svc.db.LeaderboardEntry.Get(ctx, entry.ID)
	require.NoError(t, err)

	assert.Equal(t, 1, updated.CurrentRound, "Current round should be 1")
	assert.Equal(t, 6, updated.Thru, "Thru should be 6 (only holes with scores)")
}

func TestUpdateLeaderboardProgress_EntryNotFound(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	nonExistentID := testutil.NewFactory(t, svc.db).RandomUUID()

	err := svc.UpdateLeaderboardProgress(ctx, nonExistentID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to query entry")
}
