package leaderboards

import (
	"context"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
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

func TestService_GetLeagueStandings_CurrentPickVisibility(t *testing.T) {
	t.Run("sets HasCurrentPick true when user has pick for active tournament", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		tournamentService := tournaments.NewService(db)
		service := NewService(db, tournamentService)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user1 := factory.CreateUser(testutil.WithDisplayName("Has Pick"))
		user2 := factory.CreateUser(testutil.WithDisplayName("No Pick"))
		factory.AddUserToLeague(user1, league)
		factory.AddUserToLeague(user2, league)

		activeTourn := factory.CreateActiveTournament(testutil.WithSeasonYear(time.Now().Year()))
		completedTourn := factory.CreateCompletedTournament(testutil.WithSeasonYear(time.Now().Year()), testutil.WithTournamentName("Completed"))

		golfer1 := factory.CreateGolfer()
		golfer2 := factory.CreateGolfer()
		factory.CreatePlacement(completedTourn, golfer1, testutil.WithEarnings(50000))
		factory.CreatePlacement(completedTourn, golfer2, testutil.WithEarnings(50000))

		factory.CreatePick(user1, completedTourn, golfer1, league)
		factory.CreatePick(user2, completedTourn, golfer2, league)

		golfer3 := factory.CreateGolfer()
		factory.CreatePick(user1, activeTourn, golfer3, league)

		resp, err := service.GetLeagueStandings(ctx, league.ID, 0, false, uuid.Nil)

		require.NoError(t, err)
		require.Equal(t, 2, len(resp.Entries))

		var user1Entry, user2Entry *StandingsEntry
		for i := range resp.Entries {
			if resp.Entries[i].UserDisplayName == "Has Pick" {
				user1Entry = &resp.Entries[i]
			} else if resp.Entries[i].UserDisplayName == "No Pick" {
				user2Entry = &resp.Entries[i]
			}
		}

		require.NotNil(t, user1Entry)
		require.NotNil(t, user2Entry)
		assert.True(t, user1Entry.HasCurrentPick)
		assert.False(t, user2Entry.HasCurrentPick)
	})

	t.Run("hides current_pick details when pick window is open for other users", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		tournamentService := tournaments.NewService(db)
		service := NewService(db, tournamentService)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user1 := factory.CreateUser(testutil.WithDisplayName("User One"))
		user2 := factory.CreateUser(testutil.WithDisplayName("User Two"))
		factory.AddUserToLeague(user1, league)
		factory.AddUserToLeague(user2, league)

		startDate := testutil.TournamentInPickWindow()
		activeTourn := factory.CreateTournament(
			testutil.WithSeasonYear(time.Now().Year()),
			testutil.WithStartDate(startDate),
			testutil.WithPickWindow(startDate.AddDate(0, 0, -3), startDate.Add(time.Hour)),
		)

		golfer1 := factory.CreateGolfer()
		golfer2 := factory.CreateGolfer()
		factory.CreatePick(user1, activeTourn, golfer1, league)
		factory.CreatePick(user2, activeTourn, golfer2, league)

		resp, err := service.GetLeagueStandings(ctx, league.ID, 0, false, user1.ID)

		require.NoError(t, err)
		require.Equal(t, 2, len(resp.Entries))

		var user1Entry, user2Entry *StandingsEntry
		for i := range resp.Entries {
			if resp.Entries[i].UserDisplayName == "User One" {
				user1Entry = &resp.Entries[i]
			} else if resp.Entries[i].UserDisplayName == "User Two" {
				user2Entry = &resp.Entries[i]
			}
		}

		require.NotNil(t, user1Entry)
		require.NotNil(t, user2Entry)

		assert.True(t, user1Entry.HasCurrentPick)
		assert.True(t, user2Entry.HasCurrentPick)

		assert.NotNil(t, user1Entry.CurrentPick, "requesting user should see their own pick")
		assert.Equal(t, golfer1.Name, user1Entry.CurrentPick.GolferName)

		assert.Nil(t, user2Entry.CurrentPick, "other user's pick should be hidden when window is open")
	})

	t.Run("exposes current_pick details when pick window is closed", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		tournamentService := tournaments.NewService(db)
		service := NewService(db, tournamentService)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user1 := factory.CreateUser(testutil.WithDisplayName("User One"))
		user2 := factory.CreateUser(testutil.WithDisplayName("User Two"))
		factory.AddUserToLeague(user1, league)
		factory.AddUserToLeague(user2, league)

		activeTourn := factory.CreateActiveTournament(testutil.WithSeasonYear(time.Now().Year()))

		golfer1 := factory.CreateGolfer()
		golfer2 := factory.CreateGolfer()
		factory.CreatePick(user1, activeTourn, golfer1, league)
		factory.CreatePick(user2, activeTourn, golfer2, league)

		resp, err := service.GetLeagueStandings(ctx, league.ID, 0, false, user1.ID)

		require.NoError(t, err)
		require.Equal(t, 2, len(resp.Entries))

		var user1Entry, user2Entry *StandingsEntry
		for i := range resp.Entries {
			if resp.Entries[i].UserDisplayName == "User One" {
				user1Entry = &resp.Entries[i]
			} else if resp.Entries[i].UserDisplayName == "User Two" {
				user2Entry = &resp.Entries[i]
			}
		}

		require.NotNil(t, user1Entry)
		require.NotNil(t, user2Entry)

		assert.True(t, user1Entry.HasCurrentPick)
		assert.True(t, user2Entry.HasCurrentPick)

		assert.NotNil(t, user1Entry.CurrentPick, "all picks should be visible when window is closed")
		assert.NotNil(t, user2Entry.CurrentPick, "all picks should be visible when window is closed")
		assert.Equal(t, golfer1.Name, user1Entry.CurrentPick.GolferName)
		assert.Equal(t, golfer2.Name, user2Entry.CurrentPick.GolferName)
	})

	t.Run("hides all current_picks when requesting user is not authenticated", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		tournamentService := tournaments.NewService(db)
		service := NewService(db, tournamentService)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		startDate := testutil.TournamentInPickWindow()
		activeTourn := factory.CreateTournament(
			testutil.WithSeasonYear(time.Now().Year()),
			testutil.WithStartDate(startDate),
			testutil.WithPickWindow(startDate.AddDate(0, 0, -3), startDate.Add(time.Hour)),
		)

		golfer := factory.CreateGolfer()
		factory.CreatePick(user, activeTourn, golfer, league)

		resp, err := service.GetLeagueStandings(ctx, league.ID, 0, false, uuid.Nil)

		require.NoError(t, err)
		require.Equal(t, 1, len(resp.Entries))

		assert.True(t, resp.Entries[0].HasCurrentPick)
		assert.Nil(t, resp.Entries[0].CurrentPick, "pick should be hidden for unauthenticated requests")
	})
}
