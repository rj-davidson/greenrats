package leaderboards

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/internal/features/tournaments"
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
		assert.Equal(t, 100000, resp.Entries[0].Earnings)
		assert.Equal(t, 100000, resp.Entries[1].Earnings)
		assert.Equal(t, 50000, resp.Entries[2].Earnings)
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

func TestService_GetLeagueStandings_HasCurrentPick(t *testing.T) {
	t.Run("HasCurrentPick is true and CurrentPick populated when pick window closed", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		tournamentService := tournaments.NewService(db)
		service := NewService(db, tournamentService)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser(testutil.WithDisplayName("Test User"))
		factory.AddUserToLeague(user, league)

		activeTournament := factory.CreateActiveTournament(testutil.WithSeasonYear(time.Now().Year()))
		golfer := factory.CreateGolfer(testutil.WithGolferName("Tiger Woods"))
		factory.CreatePlacement(activeTournament, golfer, testutil.WithEarnings(0))
		factory.CreatePick(user, activeTournament, golfer, league)

		resp, err := service.GetLeagueStandings(ctx, league.ID, 0, false)

		require.NoError(t, err)
		require.Len(t, resp.Entries, 1)
		assert.True(t, resp.Entries[0].HasCurrentPick)
		assert.NotNil(t, resp.Entries[0].CurrentPick)
		assert.Equal(t, "Tiger Woods", resp.Entries[0].CurrentPick.GolferName)
	})

	t.Run("HasCurrentPick is true but CurrentPick nil when pick window open for privacy", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		tournamentService := tournaments.NewService(db)
		service := NewService(db, tournamentService)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser(testutil.WithDisplayName("Test User"))
		factory.AddUserToLeague(user, league)

		pickWindowOpenTournament := factory.CreateTournament(
			testutil.WithSeasonYear(time.Now().Year()),
			testutil.WithStartDate(time.Now().AddDate(0, 0, 2)),
			testutil.WithEndDate(time.Now().AddDate(0, 0, 6)),
			testutil.WithPickWindow(time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 1)),
		)
		golfer := factory.CreateGolfer(testutil.WithGolferName("Tiger Woods"))
		factory.CreatePlacement(pickWindowOpenTournament, golfer, testutil.WithEarnings(0))
		factory.CreatePick(user, pickWindowOpenTournament, golfer, league)

		resp, err := service.GetLeagueStandings(ctx, league.ID, 0, false)

		require.NoError(t, err)
		require.Len(t, resp.Entries, 1)
		assert.True(t, resp.Entries[0].HasCurrentPick, "HasCurrentPick should be true even when pick window open")
		assert.Nil(t, resp.Entries[0].CurrentPick, "CurrentPick should be nil for privacy when pick window open")
	})

	t.Run("active_tournament returned even when pick window open", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		tournamentService := tournaments.NewService(db)
		service := NewService(db, tournamentService)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		pickWindowOpenTournament := factory.CreateTournament(
			testutil.WithSeasonYear(time.Now().Year()),
			testutil.WithTournamentName("Upcoming Open"),
			testutil.WithStartDate(time.Now().AddDate(0, 0, 2)),
			testutil.WithEndDate(time.Now().AddDate(0, 0, 6)),
			testutil.WithPickWindow(time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 1)),
		)
		golfer := factory.CreateGolfer()
		factory.CreatePlacement(pickWindowOpenTournament, golfer, testutil.WithEarnings(0))
		factory.CreatePick(user, pickWindowOpenTournament, golfer, league)

		resp, err := service.GetLeagueStandings(ctx, league.ID, 0, false)

		require.NoError(t, err)
		require.NotNil(t, resp.ActiveTournament, "ActiveTournament should be returned even when pick window open")
		assert.Equal(t, "Upcoming Open", resp.ActiveTournament.Name)
		assert.False(t, resp.ActiveTournament.IsPickWindowClosed)
	})

	t.Run("active_tournament shows IsPickWindowClosed true when window closed", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		tournamentService := tournaments.NewService(db)
		service := NewService(db, tournamentService)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		activeTournament := factory.CreateActiveTournament(testutil.WithSeasonYear(time.Now().Year()), testutil.WithTournamentName("Active Tournament"))
		golfer := factory.CreateGolfer()
		factory.CreatePlacement(activeTournament, golfer, testutil.WithEarnings(0))
		factory.CreatePick(user, activeTournament, golfer, league)

		resp, err := service.GetLeagueStandings(ctx, league.ID, 0, false)

		require.NoError(t, err)
		require.NotNil(t, resp.ActiveTournament)
		assert.Equal(t, "Active Tournament", resp.ActiveTournament.Name)
		assert.True(t, resp.ActiveTournament.IsPickWindowClosed)
	})

	t.Run("HasCurrentPick is false when user has no pick for current tournament", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		tournamentService := tournaments.NewService(db)
		service := NewService(db, tournamentService)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		completedTournament := factory.CreateCompletedTournament(testutil.WithSeasonYear(time.Now().Year()))
		_ = factory.CreateActiveTournament(testutil.WithSeasonYear(time.Now().Year()))

		golfer := factory.CreateGolfer()
		factory.CreatePlacement(completedTournament, golfer, testutil.WithEarnings(100000))
		factory.CreatePick(user, completedTournament, golfer, league)

		resp, err := service.GetLeagueStandings(ctx, league.ID, 0, false)

		require.NoError(t, err)
		require.Len(t, resp.Entries, 1)
		assert.False(t, resp.Entries[0].HasCurrentPick)
		assert.Nil(t, resp.Entries[0].CurrentPick)
	})

	t.Run("HasCurrentPick is false when no current tournament", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		tournamentService := tournaments.NewService(db)
		service := NewService(db, tournamentService)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		completedTournament := factory.CreateCompletedTournament(testutil.WithSeasonYear(time.Now().Year()))
		golfer := factory.CreateGolfer()
		factory.CreatePlacement(completedTournament, golfer, testutil.WithEarnings(100000))
		factory.CreatePick(user, completedTournament, golfer, league)

		resp, err := service.GetLeagueStandings(ctx, league.ID, 0, false)

		require.NoError(t, err)
		require.Len(t, resp.Entries, 1)
		assert.False(t, resp.Entries[0].HasCurrentPick)
	})

	t.Run("mixed users with and without current picks when window closed", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		tournamentService := tournaments.NewService(db)
		service := NewService(db, tournamentService)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		userWithPick := factory.CreateUser(testutil.WithDisplayName("Has Pick"))
		userWithoutPick := factory.CreateUser(testutil.WithDisplayName("No Pick"))
		factory.AddUserToLeague(userWithPick, league)
		factory.AddUserToLeague(userWithoutPick, league)

		completedTournament := factory.CreateCompletedTournament(testutil.WithSeasonYear(time.Now().Year()))
		activeTournament := factory.CreateActiveTournament(testutil.WithSeasonYear(time.Now().Year()))

		golfer1 := factory.CreateGolfer()
		golfer2 := factory.CreateGolfer()
		golfer3 := factory.CreateGolfer()
		factory.CreatePlacement(completedTournament, golfer1, testutil.WithEarnings(100000))
		factory.CreatePlacement(completedTournament, golfer2, testutil.WithEarnings(50000))
		factory.CreatePlacement(activeTournament, golfer3, testutil.WithEarnings(0))

		factory.CreatePick(userWithPick, completedTournament, golfer1, league)
		factory.CreatePick(userWithPick, activeTournament, golfer3, league)
		factory.CreatePick(userWithoutPick, completedTournament, golfer2, league)

		resp, err := service.GetLeagueStandings(ctx, league.ID, 0, false)

		require.NoError(t, err)
		require.Len(t, resp.Entries, 2)

		var entryWithPick, entryWithoutPick StandingsEntry
		for _, entry := range resp.Entries {
			if entry.UserDisplayName == "Has Pick" {
				entryWithPick = entry
			} else {
				entryWithoutPick = entry
			}
		}

		assert.True(t, entryWithPick.HasCurrentPick)
		assert.NotNil(t, entryWithPick.CurrentPick)
		assert.False(t, entryWithoutPick.HasCurrentPick)
		assert.Nil(t, entryWithoutPick.CurrentPick)
	})

	t.Run("active_tournaments returns multiple overlapping tournaments", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		tournamentService := tournaments.NewService(db)
		service := NewService(db, tournamentService)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser(testutil.WithDisplayName("Test User"))
		factory.AddUserToLeague(user, league)

		active1 := factory.CreateActiveTournament(testutil.WithSeasonYear(time.Now().Year()), testutil.WithTournamentName("Tournament A"))
		active2 := factory.CreateActiveTournament(testutil.WithSeasonYear(time.Now().Year()), testutil.WithTournamentName("Tournament B"))

		golfer1 := factory.CreateGolfer(testutil.WithGolferName("Golfer One"))
		golfer2 := factory.CreateGolfer(testutil.WithGolferName("Golfer Two"))
		factory.CreatePlacement(active1, golfer1, testutil.WithEarnings(0))
		factory.CreatePlacement(active2, golfer2, testutil.WithEarnings(0))
		factory.CreatePick(user, active1, golfer1, league)
		factory.CreatePick(user, active2, golfer2, league)

		resp, err := service.GetLeagueStandings(ctx, league.ID, 0, false)

		require.NoError(t, err)
		require.NotNil(t, resp.ActiveTournaments)
		assert.Len(t, resp.ActiveTournaments, 2)

		require.Len(t, resp.Entries, 1)
		assert.Len(t, resp.Entries[0].ActivePicks, 2)

		for _, ap := range resp.Entries[0].ActivePicks {
			assert.True(t, ap.HasPick)
			assert.True(t, ap.IsPickWindowClosed)
			assert.NotNil(t, ap.GolferName)
		}

		require.NotNil(t, resp.ActiveTournament, "backward-compat singular field should still be set")
	})

	t.Run("active_picks shows no pick for tournament user skipped", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		tournamentService := tournaments.NewService(db)
		service := NewService(db, tournamentService)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser(testutil.WithDisplayName("Test User"))
		factory.AddUserToLeague(user, league)

		active1 := factory.CreateActiveTournament(testutil.WithSeasonYear(time.Now().Year()), testutil.WithTournamentName("Tournament A"))
		_ = factory.CreateActiveTournament(testutil.WithSeasonYear(time.Now().Year()), testutil.WithTournamentName("Tournament B"))

		golfer := factory.CreateGolfer()
		factory.CreatePlacement(active1, golfer, testutil.WithEarnings(0))
		factory.CreatePick(user, active1, golfer, league)

		resp, err := service.GetLeagueStandings(ctx, league.ID, 0, false)

		require.NoError(t, err)
		require.Len(t, resp.Entries, 1)
		require.Len(t, resp.Entries[0].ActivePicks, 2)

		var pickedCount, notPickedCount int
		for _, ap := range resp.Entries[0].ActivePicks {
			if ap.HasPick {
				pickedCount++
			} else {
				notPickedCount++
			}
		}
		assert.Equal(t, 1, pickedCount)
		assert.Equal(t, 1, notPickedCount)
	})

	t.Run("mixed users with pick window open hides golfer details", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		tournamentService := tournaments.NewService(db)
		service := NewService(db, tournamentService)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		userWithPick := factory.CreateUser(testutil.WithDisplayName("Has Pick"))
		userWithoutPick := factory.CreateUser(testutil.WithDisplayName("No Pick"))
		factory.AddUserToLeague(userWithPick, league)
		factory.AddUserToLeague(userWithoutPick, league)

		completedTournament := factory.CreateCompletedTournament(testutil.WithSeasonYear(time.Now().Year()))
		pickWindowOpenTournament := factory.CreateTournament(
			testutil.WithSeasonYear(time.Now().Year()),
			testutil.WithStartDate(time.Now().AddDate(0, 0, 2)),
			testutil.WithEndDate(time.Now().AddDate(0, 0, 6)),
			testutil.WithPickWindow(time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 1)),
		)

		golfer1 := factory.CreateGolfer()
		golfer2 := factory.CreateGolfer()
		golfer3 := factory.CreateGolfer(testutil.WithGolferName("Secret Pick"))
		factory.CreatePlacement(completedTournament, golfer1, testutil.WithEarnings(100000))
		factory.CreatePlacement(completedTournament, golfer2, testutil.WithEarnings(50000))
		factory.CreatePlacement(pickWindowOpenTournament, golfer3, testutil.WithEarnings(0))

		factory.CreatePick(userWithPick, completedTournament, golfer1, league)
		factory.CreatePick(userWithPick, pickWindowOpenTournament, golfer3, league)
		factory.CreatePick(userWithoutPick, completedTournament, golfer2, league)

		resp, err := service.GetLeagueStandings(ctx, league.ID, 0, false)

		require.NoError(t, err)
		require.Len(t, resp.Entries, 2)

		var entryWithPick, entryWithoutPick StandingsEntry
		for _, entry := range resp.Entries {
			if entry.UserDisplayName == "Has Pick" {
				entryWithPick = entry
			} else {
				entryWithoutPick = entry
			}
		}

		assert.True(t, entryWithPick.HasCurrentPick, "Should show user has a pick")
		assert.Nil(t, entryWithPick.CurrentPick, "Should NOT reveal golfer details while pick window open")
		assert.False(t, entryWithoutPick.HasCurrentPick)
		assert.Nil(t, entryWithoutPick.CurrentPick)
	})
}
