package picks

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/internal/testutil"
)

func TestService_Create(t *testing.T) {
	t.Run("creates pick successfully when validations pass", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn := factory.CreateUpcomingTournament(2)
		golfer := factory.CreateGolfer()
		factory.CreateFieldEntry(tourn, golfer)

		pick, err := service.Create(ctx, CreateParams{
			UserID:       user.ID,
			TournamentID: tourn.ID,
			GolferID:     golfer.ID,
			LeagueID:     league.ID,
		})

		require.NoError(t, err)
		assert.Equal(t, user.ID, pick.UserID)
		assert.Equal(t, tourn.ID, pick.TournamentID)
		assert.Equal(t, golfer.ID, pick.GolferID)
		assert.Equal(t, league.ID, pick.LeagueID)
	})

	t.Run("returns error when tournament not found", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		user := factory.CreateUser()
		golfer := factory.CreateGolfer()
		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		factory.AddUserToLeague(user, league)

		_, err := service.Create(ctx, CreateParams{
			UserID:       user.ID,
			TournamentID: factory.RandomUUID(),
			GolferID:     golfer.ID,
			LeagueID:     league.ID,
		})

		require.ErrorIs(t, err, ErrTournamentNotFound)
	})

	t.Run("returns error when tournament not upcoming", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn := factory.CreateActiveTournament()
		golfer := factory.CreateGolfer()
		factory.CreateFieldEntry(tourn, golfer)

		_, err := service.Create(ctx, CreateParams{
			UserID:       user.ID,
			TournamentID: tourn.ID,
			GolferID:     golfer.ID,
			LeagueID:     league.ID,
		})

		require.ErrorIs(t, err, ErrPickWindowClosed)
	})

	t.Run("returns error when pick window closed", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn := factory.CreateUpcomingTournament(10)
		golfer := factory.CreateGolfer()
		factory.CreateFieldEntry(tourn, golfer)

		_, err := service.Create(ctx, CreateParams{
			UserID:       user.ID,
			TournamentID: tourn.ID,
			GolferID:     golfer.ID,
			LeagueID:     league.ID,
		})

		require.ErrorIs(t, err, ErrPickWindowClosed)
	})

	t.Run("returns error when golfer not found", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn := factory.CreateUpcomingTournament(2)

		_, err := service.Create(ctx, CreateParams{
			UserID:       user.ID,
			TournamentID: tourn.ID,
			GolferID:     factory.RandomUUID(),
			LeagueID:     league.ID,
		})

		require.ErrorIs(t, err, ErrGolferNotFound)
	})

	t.Run("returns error when not league member", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()

		tourn := factory.CreateUpcomingTournament(2)
		golfer := factory.CreateGolfer()
		factory.CreateFieldEntry(tourn, golfer)

		_, err := service.Create(ctx, CreateParams{
			UserID:       user.ID,
			TournamentID: tourn.ID,
			GolferID:     golfer.ID,
			LeagueID:     league.ID,
		})

		require.ErrorIs(t, err, ErrNotLeagueMember)
	})

	t.Run("returns error when golfer not in field", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn := factory.CreateUpcomingTournament(2)
		golfer := factory.CreateGolfer()

		_, err := service.Create(ctx, CreateParams{
			UserID:       user.ID,
			TournamentID: tourn.ID,
			GolferID:     golfer.ID,
			LeagueID:     league.ID,
		})

		require.ErrorIs(t, err, ErrGolferNotInField)
	})

	t.Run("returns error when golfer already used", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn1 := factory.CreateUpcomingTournament(2)
		tourn2 := factory.CreateUpcomingTournament(2, testutil.WithTournamentName("Second Tournament"))
		golfer := factory.CreateGolfer()
		factory.CreateFieldEntry(tourn1, golfer)
		factory.CreateFieldEntry(tourn2, golfer)

		factory.CreatePick(user, tourn1, golfer, league)

		_, err := service.Create(ctx, CreateParams{
			UserID:       user.ID,
			TournamentID: tourn2.ID,
			GolferID:     golfer.ID,
			LeagueID:     league.ID,
		})

		require.ErrorIs(t, err, ErrGolferAlreadyUsed)
	})

	t.Run("returns error when pick already exists", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn := factory.CreateUpcomingTournament(2)
		golfer1 := factory.CreateGolfer()
		golfer2 := factory.CreateGolfer()
		factory.CreateFieldEntry(tourn, golfer1)
		factory.CreateFieldEntry(tourn, golfer2)

		factory.CreatePick(user, tourn, golfer1, league)

		_, err := service.Create(ctx, CreateParams{
			UserID:       user.ID,
			TournamentID: tourn.ID,
			GolferID:     golfer2.ID,
			LeagueID:     league.ID,
		})

		require.ErrorIs(t, err, ErrPickAlreadyExists)
	})
}

func TestService_CanMakePick(t *testing.T) {
	t.Run("returns open when in pick window", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		tourn := factory.CreateUpcomingTournament(2)

		status, err := service.CanMakePick(ctx, tourn.ID)

		require.NoError(t, err)
		assert.True(t, status.IsOpen)
	})

	t.Run("returns closed when before pick window", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		tourn := factory.CreateUpcomingTournament(10)

		status, err := service.CanMakePick(ctx, tourn.ID)

		require.NoError(t, err)
		assert.False(t, status.IsOpen)
		assert.Equal(t, "pick window not yet open", status.Reason)
	})

	t.Run("returns closed when tournament active", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		tourn := factory.CreateActiveTournament()

		status, err := service.CanMakePick(ctx, tourn.ID)

		require.NoError(t, err)
		assert.False(t, status.IsOpen)
		assert.Equal(t, "pick window has closed", status.Reason)
	})

	t.Run("returns error when tournament not found", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		_, err := service.CanMakePick(ctx, factory.RandomUUID())

		require.ErrorIs(t, err, ErrTournamentNotFound)
	})
}

func TestService_GetUserPicks(t *testing.T) {
	t.Run("returns all user picks", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn1 := factory.CreateCompletedTournament()
		tourn2 := factory.CreateCompletedTournament(testutil.WithTournamentName("Second"))
		golfer1 := factory.CreateGolfer()
		golfer2 := factory.CreateGolfer()

		factory.CreatePick(user, tourn1, golfer1, league)
		factory.CreatePick(user, tourn2, golfer2, league)

		resp, err := service.GetUserPicks(ctx, user.ID, league.ID, 0)

		require.NoError(t, err)
		assert.Equal(t, 2, resp.Total)
	})

	t.Run("filters by season year", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn := factory.CreateCompletedTournament(testutil.WithSeasonYear(2024))
		golfer := factory.CreateGolfer()

		factory.CreatePick(user, tourn, golfer, league)

		resp, err := service.GetUserPicks(ctx, user.ID, league.ID, 2025)

		require.NoError(t, err)
		assert.Equal(t, 0, resp.Total)
	})
}

func TestService_GetAvailableGolfers(t *testing.T) {
	t.Run("returns all golfers in field", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn := factory.CreateUpcomingTournament(2)
		golfers := factory.CreateGolfers(5)
		factory.CreateTournamentField(tourn, golfers)

		resp, err := service.GetAvailableGolfers(ctx, user.ID, league.ID, tourn.ID)

		require.NoError(t, err)
		assert.Equal(t, 5, resp.Total)
	})

	t.Run("marks used golfers", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn1 := factory.CreateCompletedTournament()
		tourn2 := factory.CreateUpcomingTournament(2)

		golfer := factory.CreateGolfer()
		factory.CreateFieldEntry(tourn1, golfer)
		factory.CreateFieldEntry(tourn2, golfer)

		factory.CreatePick(user, tourn1, golfer, league)

		resp, err := service.GetAvailableGolfers(ctx, user.ID, league.ID, tourn2.ID)

		require.NoError(t, err)
		assert.Equal(t, 1, resp.Total)
		assert.True(t, resp.Golfers[0].IsUsed)
	})

	t.Run("returns error when tournament not found", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		_, err := service.GetAvailableGolfers(ctx, user.ID, league.ID, factory.RandomUUID())

		require.ErrorIs(t, err, ErrTournamentNotFound)
	})
}

func TestService_UpdateUserPick(t *testing.T) {
	t.Run("updates pick successfully", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn := factory.CreateUpcomingTournament(2)
		golfer1 := factory.CreateGolfer()
		golfer2 := factory.CreateGolfer()
		factory.CreateFieldEntry(tourn, golfer1)
		factory.CreateFieldEntry(tourn, golfer2)

		pick := factory.CreatePick(user, tourn, golfer1, league)

		updated, err := service.UpdateUserPick(ctx, UpdatePickParams{
			UserID:      user.ID,
			PickID:      pick.ID,
			NewGolferID: golfer2.ID,
		})

		require.NoError(t, err)
		assert.Equal(t, golfer2.ID, updated.GolferID)
	})

	t.Run("returns error when not pick owner", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		otherUser := factory.CreateUser()
		factory.AddUserToLeague(user, league)
		factory.AddUserToLeague(otherUser, league)

		tourn := factory.CreateUpcomingTournament(2)
		golfer1 := factory.CreateGolfer()
		golfer2 := factory.CreateGolfer()
		factory.CreateFieldEntry(tourn, golfer1)
		factory.CreateFieldEntry(tourn, golfer2)

		pick := factory.CreatePick(user, tourn, golfer1, league)

		_, err := service.UpdateUserPick(ctx, UpdatePickParams{
			UserID:      otherUser.ID,
			PickID:      pick.ID,
			NewGolferID: golfer2.ID,
		})

		require.ErrorIs(t, err, ErrNotPickOwner)
	})

	t.Run("returns error when tournament not upcoming", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn := factory.CreateActiveTournament()
		golfer1 := factory.CreateGolfer()
		golfer2 := factory.CreateGolfer()
		factory.CreateFieldEntry(tourn, golfer1)
		factory.CreateFieldEntry(tourn, golfer2)

		pick := factory.CreatePick(user, tourn, golfer1, league)

		_, err := service.UpdateUserPick(ctx, UpdatePickParams{
			UserID:      user.ID,
			PickID:      pick.ID,
			NewGolferID: golfer2.ID,
		})

		require.ErrorIs(t, err, ErrPickWindowClosed)
	})

	t.Run("returns error when golfer already used", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn1 := factory.CreateCompletedTournament()
		tourn2 := factory.CreateUpcomingTournament(2)
		golfer1 := factory.CreateGolfer()
		golfer2 := factory.CreateGolfer()
		factory.CreateFieldEntry(tourn1, golfer2)
		factory.CreateFieldEntry(tourn2, golfer1)
		factory.CreateFieldEntry(tourn2, golfer2)

		factory.CreatePick(user, tourn1, golfer2, league)
		pick := factory.CreatePick(user, tourn2, golfer1, league)

		_, err := service.UpdateUserPick(ctx, UpdatePickParams{
			UserID:      user.ID,
			PickID:      pick.ID,
			NewGolferID: golfer2.ID,
		})

		require.ErrorIs(t, err, ErrGolferAlreadyUsed)
	})
}

func TestService_OverridePick(t *testing.T) {
	t.Run("commissioner overrides pick successfully", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn := factory.CreateUpcomingTournament(2)
		golfer1 := factory.CreateGolfer()
		golfer2 := factory.CreateGolfer()
		factory.CreateFieldEntry(tourn, golfer1)
		factory.CreateFieldEntry(tourn, golfer2)

		pick := factory.CreatePick(user, tourn, golfer1, league)

		updated, err := service.OverridePick(ctx, OverridePickParams{
			LeagueID:       league.ID,
			PickID:         pick.ID,
			NewGolferID:    golfer2.ID,
			CommissionerID: owner.ID,
		})

		require.NoError(t, err)
		assert.Equal(t, golfer2.ID, updated.GolferID)
	})

	t.Run("returns error when not commissioner", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn := factory.CreateUpcomingTournament(2)
		golfer1 := factory.CreateGolfer()
		golfer2 := factory.CreateGolfer()
		factory.CreateFieldEntry(tourn, golfer1)
		factory.CreateFieldEntry(tourn, golfer2)

		pick := factory.CreatePick(user, tourn, golfer1, league)

		_, err := service.OverridePick(ctx, OverridePickParams{
			LeagueID:       league.ID,
			PickID:         pick.ID,
			NewGolferID:    golfer2.ID,
			CommissionerID: user.ID,
		})

		require.ErrorIs(t, err, ErrNotCommissioner)
	})

	t.Run("commissioner overrides pick for completed tournament", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn := factory.CreateCompletedTournament()
		golfer1 := factory.CreateGolfer()
		golfer2 := factory.CreateGolfer()
		factory.CreateFieldEntry(tourn, golfer1)
		factory.CreateFieldEntry(tourn, golfer2)

		pick := factory.CreatePick(user, tourn, golfer1, league)

		updated, err := service.OverridePick(ctx, OverridePickParams{
			LeagueID:       league.ID,
			PickID:         pick.ID,
			NewGolferID:    golfer2.ID,
			CommissionerID: owner.ID,
		})

		require.NoError(t, err)
		assert.Equal(t, golfer2.ID, updated.GolferID)
	})
}

func TestService_GetLeaguePicksEnhanced(t *testing.T) {
	t.Run("returns picks for completed tournament with leaderboard data", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn := factory.CreateCompletedTournament()
		golfer := factory.CreateGolfer()
		factory.CreateFieldEntry(tourn, golfer)
		factory.CreatePick(user, tourn, golfer, league)
		factory.CreatePlacement(tourn, golfer, testutil.WithPosition(5), testutil.WithEarnings(100000))

		resp, err := service.GetLeaguePicksEnhanced(ctx, league.ID, tourn.ID, false)

		require.NoError(t, err)
		assert.Equal(t, 1, resp.Total)
		require.Len(t, resp.Entries, 1)
		assert.NotNil(t, resp.Entries[0].Leaderboard)
		assert.Equal(t, 5, resp.Entries[0].Leaderboard.Position)
		assert.Equal(t, 100000, resp.Entries[0].Leaderboard.Earnings)
	})

	t.Run("returns picks for completed tournament with champion", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn := factory.CreateCompletedTournament()
		golfer := factory.CreateGolfer()
		factory.CreateFieldEntry(tourn, golfer)
		factory.CreatePick(user, tourn, golfer, league)
		factory.CreatePlacement(tourn, golfer, testutil.WithPosition(10), testutil.WithEarnings(50000))

		resp, err := service.GetLeaguePicksEnhanced(ctx, league.ID, tourn.ID, false)

		require.NoError(t, err)
		assert.Equal(t, 1, resp.Total)
		require.Len(t, resp.Entries, 1)
		assert.NotNil(t, resp.Entries[0].Leaderboard)
		assert.Equal(t, 10, resp.Entries[0].Leaderboard.Position)
	})

	t.Run("returns picks for upcoming tournament without leaderboard lookup", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn := factory.CreateUpcomingTournament(7)
		golfer := factory.CreateGolfer()
		factory.CreateFieldEntry(tourn, golfer)
		factory.CreatePick(user, tourn, golfer, league)

		resp, err := service.GetLeaguePicksEnhanced(ctx, league.ID, tourn.ID, false)

		require.NoError(t, err)
		assert.Equal(t, 1, resp.Total)
		require.Len(t, resp.Entries, 1)
		assert.Nil(t, resp.Entries[0].Leaderboard)
	})

	t.Run("returns empty entries when no picks exist", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())

		tourn := factory.CreateActiveTournament()

		resp, err := service.GetLeaguePicksEnhanced(ctx, league.ID, tourn.ID, false)

		require.NoError(t, err)
		assert.Equal(t, 0, resp.Total)
		assert.Empty(t, resp.Entries)
		assert.Equal(t, 1, resp.MembersWithoutPicks)
	})

	t.Run("includes round data when requested", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn := factory.CreateCompletedTournament()
		golfer := factory.CreateGolfer()
		factory.CreateFieldEntry(tourn, golfer)
		factory.CreatePick(user, tourn, golfer, league)
		factory.CreatePlacement(tourn, golfer, testutil.WithPosition(1))

		db.Round.Create().
			SetTournament(tourn).
			SetGolfer(golfer).
			SetRoundNumber(1).
			SetNillableScore(intPtr(72)).
			SetNillableParRelativeScore(intPtr(0)).
			SaveX(ctx)

		resp, err := service.GetLeaguePicksEnhanced(ctx, league.ID, tourn.ID, true)

		require.NoError(t, err)
		require.Len(t, resp.Entries, 1)
		require.NotNil(t, resp.Entries[0].Leaderboard)
		assert.Len(t, resp.Entries[0].Leaderboard.Rounds, 1)
		assert.Equal(t, 1, resp.Entries[0].Leaderboard.Rounds[0].RoundNumber)
	})
}

func TestService_GetAvailableGolfersForUserOverride(t *testing.T) {
	t.Run("returns available golfers for target user", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn := factory.CreateUpcomingTournament(2)
		golfers := factory.CreateGolfers(5)
		factory.CreateTournamentField(tourn, golfers)

		resp, err := service.GetAvailableGolfersForUserOverride(ctx, AvailableGolfersForUserParams{
			CommissionerID: owner.ID,
			TargetUserID:   user.ID,
			LeagueID:       league.ID,
			TournamentID:   tourn.ID,
		})

		require.NoError(t, err)
		assert.Equal(t, 5, resp.Total)
	})

	t.Run("marks golfers used by target user", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn1 := factory.CreateCompletedTournament()
		tourn2 := factory.CreateUpcomingTournament(2)

		golfer := factory.CreateGolfer()
		factory.CreateFieldEntry(tourn1, golfer)
		factory.CreateFieldEntry(tourn2, golfer)

		factory.CreatePick(user, tourn1, golfer, league)

		resp, err := service.GetAvailableGolfersForUserOverride(ctx, AvailableGolfersForUserParams{
			CommissionerID: owner.ID,
			TargetUserID:   user.ID,
			LeagueID:       league.ID,
			TournamentID:   tourn2.ID,
		})

		require.NoError(t, err)
		assert.Equal(t, 1, resp.Total)
		assert.True(t, resp.Golfers[0].IsUsed)
	})

	t.Run("returns error when not commissioner", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn := factory.CreateUpcomingTournament(2)

		_, err := service.GetAvailableGolfersForUserOverride(ctx, AvailableGolfersForUserParams{
			CommissionerID: user.ID,
			TargetUserID:   user.ID,
			LeagueID:       league.ID,
			TournamentID:   tourn.ID,
		})

		require.ErrorIs(t, err, ErrNotCommissioner)
	})
}

func intPtr(i int) *int {
	return &i
}
