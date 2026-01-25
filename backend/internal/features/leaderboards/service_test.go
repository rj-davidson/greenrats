package leaderboards

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/internal/testutil"
)

func TestService_GetLeagueLeaderboard(t *testing.T) {
	t.Run("returns leaderboard with cumulative earnings", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db, nil)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user1 := factory.CreateUser(testutil.WithDisplayName("Player One"))
		user2 := factory.CreateUser(testutil.WithDisplayName("Player Two"))
		factory.AddUserToLeague(user1, league)
		factory.AddUserToLeague(user2, league)

		tourn1 := factory.CreateCompletedTournament(testutil.WithSeasonYear(time.Now().Year()))
		tourn2 := factory.CreateCompletedTournament(testutil.WithSeasonYear(time.Now().Year()), testutil.WithTournamentName("Second"))

		golfer1 := factory.CreateGolfer()
		golfer2 := factory.CreateGolfer()
		golfer3 := factory.CreateGolfer()
		golfer4 := factory.CreateGolfer()

		factory.CreatePlacement(tourn1, golfer1, testutil.WithEarnings(100000))
		factory.CreatePlacement(tourn1, golfer2, testutil.WithEarnings(50000))
		factory.CreatePlacement(tourn2, golfer3, testutil.WithEarnings(200000))
		factory.CreatePlacement(tourn2, golfer4, testutil.WithEarnings(75000))

		factory.CreatePick(user1, tourn1, golfer1, league)
		factory.CreatePick(user1, tourn2, golfer3, league)
		factory.CreatePick(user2, tourn1, golfer2, league)
		factory.CreatePick(user2, tourn2, golfer4, league)

		resp, err := service.GetLeagueLeaderboard(ctx, league.ID, 0)

		require.NoError(t, err)
		assert.Equal(t, 2, resp.Total)
		assert.Equal(t, 300000, resp.Entries[0].Earnings)
		assert.Equal(t, "Player One", resp.Entries[0].DisplayName)
		assert.Equal(t, 125000, resp.Entries[1].Earnings)
		assert.Equal(t, "Player Two", resp.Entries[1].DisplayName)
	})

	t.Run("assigns ranks correctly", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db, nil)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user1 := factory.CreateUser(testutil.WithDisplayName("AAA"))
		user2 := factory.CreateUser(testutil.WithDisplayName("BBB"))
		user3 := factory.CreateUser(testutil.WithDisplayName("CCC"))
		factory.AddUserToLeague(user1, league)
		factory.AddUserToLeague(user2, league)
		factory.AddUserToLeague(user3, league)

		tourn := factory.CreateCompletedTournament(testutil.WithSeasonYear(time.Now().Year()))
		golfer1 := factory.CreateGolfer()
		golfer2 := factory.CreateGolfer()
		golfer3 := factory.CreateGolfer()
		factory.CreatePlacement(tourn, golfer1, testutil.WithEarnings(100000))
		factory.CreatePlacement(tourn, golfer2, testutil.WithEarnings(100000))
		factory.CreatePlacement(tourn, golfer3, testutil.WithEarnings(50000))

		factory.CreatePick(user1, tourn, golfer1, league)
		factory.CreatePick(user2, tourn, golfer2, league)
		factory.CreatePick(user3, tourn, golfer3, league)

		resp, err := service.GetLeagueLeaderboard(ctx, league.ID, 0)

		require.NoError(t, err)
		assert.Equal(t, 3, resp.Total)
		assert.Equal(t, 1, resp.Entries[0].Rank)
		assert.Equal(t, 1, resp.Entries[1].Rank)
		assert.Equal(t, 3, resp.Entries[2].Rank)
	})

	t.Run("sorts ties alphabetically", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db, nil)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user1 := factory.CreateUser(testutil.WithDisplayName("Zara"))
		user2 := factory.CreateUser(testutil.WithDisplayName("Aaron"))
		factory.AddUserToLeague(user1, league)
		factory.AddUserToLeague(user2, league)

		tourn := factory.CreateCompletedTournament(testutil.WithSeasonYear(time.Now().Year()))
		golfer1 := factory.CreateGolfer()
		golfer2 := factory.CreateGolfer()
		factory.CreatePlacement(tourn, golfer1, testutil.WithEarnings(100000))
		factory.CreatePlacement(tourn, golfer2, testutil.WithEarnings(100000))

		factory.CreatePick(user1, tourn, golfer1, league)
		factory.CreatePick(user2, tourn, golfer2, league)

		resp, err := service.GetLeagueLeaderboard(ctx, league.ID, 0)

		require.NoError(t, err)
		assert.Equal(t, 2, resp.Total)
		assert.Equal(t, "Aaron", resp.Entries[0].DisplayName)
		assert.Equal(t, "Zara", resp.Entries[1].DisplayName)
	})

	t.Run("filters by season year", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db, nil)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, 2025)
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn2024 := factory.CreateCompletedTournament(testutil.WithSeasonYear(2024))
		tourn2025 := factory.CreateCompletedTournament(testutil.WithSeasonYear(2025))

		golfer1 := factory.CreateGolfer()
		golfer2 := factory.CreateGolfer()
		factory.CreatePlacement(tourn2024, golfer1, testutil.WithEarnings(100000))
		factory.CreatePlacement(tourn2025, golfer2, testutil.WithEarnings(200000))

		factory.CreatePick(user, tourn2024, golfer1, league)
		factory.CreatePick(user, tourn2025, golfer2, league)

		resp2024, err := service.GetLeagueLeaderboard(ctx, league.ID, 2024)
		require.NoError(t, err)
		assert.Equal(t, 100000, resp2024.Entries[0].Earnings)

		resp2025, err := service.GetLeagueLeaderboard(ctx, league.ID, 2025)
		require.NoError(t, err)
		assert.Equal(t, 200000, resp2025.Entries[0].Earnings)
	})

	t.Run("returns error for invalid league", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db, nil)
		ctx := context.Background()

		_, err := service.GetLeagueLeaderboard(ctx, factory.RandomUUID(), 0)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "league not found")
	})

	t.Run("handles picks with no earnings", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db, nil)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn := factory.CreateCompletedTournament(testutil.WithSeasonYear(time.Now().Year()))
		golfer := factory.CreateGolfer()
		factory.CreatePlacement(tourn, golfer, testutil.WithCut(true), testutil.WithEarnings(0))

		factory.CreatePick(user, tourn, golfer, league)

		resp, err := service.GetLeagueLeaderboard(ctx, league.ID, 0)

		require.NoError(t, err)
		assert.Equal(t, 1, resp.Total)
		assert.Equal(t, 0, resp.Entries[0].Earnings)
	})

	t.Run("counts picks correctly", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db, nil)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn1 := factory.CreateCompletedTournament(testutil.WithSeasonYear(time.Now().Year()))
		tourn2 := factory.CreateCompletedTournament(testutil.WithSeasonYear(time.Now().Year()), testutil.WithTournamentName("Second"))

		golfer1 := factory.CreateGolfer()
		golfer2 := factory.CreateGolfer()
		factory.CreatePlacement(tourn1, golfer1, testutil.WithEarnings(50000))
		factory.CreatePlacement(tourn2, golfer2, testutil.WithEarnings(50000))

		factory.CreatePick(user, tourn1, golfer1, league)
		factory.CreatePick(user, tourn2, golfer2, league)

		resp, err := service.GetLeagueLeaderboard(ctx, league.ID, 0)

		require.NoError(t, err)
		assert.Equal(t, 2, resp.Entries[0].PickCount)
	})
}
