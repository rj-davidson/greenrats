package integration

import (
	"context"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/internal/features/leaderboards"
	"github.com/rj-davidson/greenrats/internal/features/leagues"
	"github.com/rj-davidson/greenrats/internal/testutil"
)

func TestLeaguesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := testutil.NewPostgresTestDB(ctx, t)
	factory := testutil.NewFactory(t, db)

	leagueService := leagues.NewService(db, 2026)
	leagueHandler := leagues.NewHandler(leagueService, nil)
	leaderboardService := leaderboards.NewService(db)
	leaderboardHandler := leaderboards.NewHandler(leaderboardService)

	t.Run("full league flow", func(t *testing.T) {
		owner := factory.CreateUser(testutil.WithDisplayName("League Owner"))

		app := testutil.NewTestApp(t).WithDBUser(owner)
		app.App.Post("/leagues", leagueHandler.Create)
		app.App.Get("/leagues", leagueHandler.ListUserLeagues)
		app.App.Get("/leagues/:id", leagueHandler.GetByID)

		resp := app.Post("/leagues", leagues.CreateLeagueRequest{
			Name: "Test Integration League",
		})
		require.Equal(t, fiber.StatusCreated, resp.StatusCode)

		var createResult leagues.CreateLeagueResponse
		require.NoError(t, resp.JSON(&createResult))
		assert.Equal(t, "Test Integration League", createResult.League.Name)
		assert.Equal(t, "owner", createResult.League.Role)
		leagueID := createResult.League.ID
		joinCode := createResult.League.Code

		resp = app.Get("/leagues")
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var listResult leagues.ListUserLeaguesResponse
		require.NoError(t, resp.JSON(&listResult))
		assert.Equal(t, 1, listResult.Total)

		member := factory.CreateUser(testutil.WithDisplayName("League Member"))
		memberApp := testutil.NewTestApp(t).WithDBUser(member)
		memberApp.App.Post("/leagues/join", leagueHandler.JoinLeague)
		memberApp.App.Get("/leagues/:id", leagueHandler.GetByID)

		resp = memberApp.Post("/leagues/join", leagues.JoinLeagueRequest{
			Code: joinCode,
		})
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var joinResult leagues.JoinLeagueResponse
		require.NoError(t, resp.JSON(&joinResult))
		assert.Equal(t, leagueID, joinResult.League.ID)
		assert.Equal(t, "member", joinResult.League.Role)
		assert.Equal(t, 2, joinResult.League.MemberCount)

		resp = memberApp.Get("/leagues/" + leagueID.String())
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var getResult leagues.League
		require.NoError(t, resp.JSON(&getResult))
		assert.Equal(t, 2, getResult.MemberCount)
	})

	t.Run("commissioner controls", func(t *testing.T) {
		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		member := factory.CreateUser()
		factory.AddUserToLeague(member, league)

		ownerApp := testutil.NewTestApp(t).WithDBUser(owner)
		ownerApp.App.Post("/leagues/:id/regenerate-code", leagueHandler.RegenerateJoinCode)
		ownerApp.App.Put("/leagues/:id/joining-enabled", leagueHandler.SetJoiningEnabled)

		memberApp := testutil.NewTestApp(t).WithDBUser(member)
		memberApp.App.Post("/leagues/:id/regenerate-code", leagueHandler.RegenerateJoinCode)
		memberApp.App.Put("/leagues/:id/joining-enabled", leagueHandler.SetJoiningEnabled)

		oldCode := league.Code
		resp := ownerApp.Post("/leagues/"+league.ID.String()+"/regenerate-code", nil)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var codeResult leagues.RegenerateCodeResponse
		require.NoError(t, resp.JSON(&codeResult))
		assert.NotEqual(t, oldCode, codeResult.League.Code)

		resp = memberApp.Post("/leagues/"+league.ID.String()+"/regenerate-code", nil)
		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

		resp = ownerApp.Put("/leagues/"+league.ID.String()+"/joining-enabled", leagues.SetJoiningEnabledRequest{
			Enabled: false,
		})
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var enabledResult leagues.SetJoiningEnabledResponse
		require.NoError(t, resp.JSON(&enabledResult))
		assert.False(t, enabledResult.League.JoiningEnabled)

		newUser := factory.CreateUser()
		newUserApp := testutil.NewTestApp(t).WithDBUser(newUser)
		newUserApp.App.Post("/leagues/join", leagueHandler.JoinLeague)

		resp = newUserApp.Post("/leagues/join", leagues.JoinLeagueRequest{
			Code: codeResult.League.Code,
		})
		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	})

	t.Run("leaderboard integration", func(t *testing.T) {
		owner := factory.CreateUser(testutil.WithDisplayName("First Place"))
		league := factory.CreateLeague(owner, time.Now().Year())
		user2 := factory.CreateUser(testutil.WithDisplayName("Second Place"))
		factory.AddUserToLeague(user2, league)

		tourn := factory.CreateCompletedTournament(testutil.WithSeasonYear(time.Now().Year()))
		golfer1 := factory.CreateGolfer()
		golfer2 := factory.CreateGolfer()
		factory.CreateTournamentEntry(tourn, golfer1, testutil.WithEarnings(500000))
		factory.CreateTournamentEntry(tourn, golfer2, testutil.WithEarnings(250000))

		factory.CreatePick(owner, tourn, golfer1, league)
		factory.CreatePick(user2, tourn, golfer2, league)

		app := testutil.NewTestApp(t).WithDBUser(owner)
		app.App.Get("/leagues/:id/leaderboard", leaderboardHandler.GetLeagueLeaderboard)

		resp := app.Get("/leagues/" + league.ID.String() + "/leaderboard")
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result leaderboards.LeagueLeaderboardResponse
		require.NoError(t, resp.JSON(&result))
		assert.Equal(t, 2, result.Total)
		assert.Equal(t, "First Place", result.Entries[0].DisplayName)
		assert.Equal(t, 500000, result.Entries[0].Earnings)
		assert.Equal(t, 1, result.Entries[0].Rank)
		assert.Equal(t, "Second Place", result.Entries[1].DisplayName)
		assert.Equal(t, 250000, result.Entries[1].Earnings)
		assert.Equal(t, 2, result.Entries[1].Rank)
	})

	t.Run("league tournaments view", func(t *testing.T) {
		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn1 := factory.CreateCompletedTournament(
			testutil.WithTournamentName("Completed"),
			testutil.WithSeasonYear(time.Now().Year()),
		)
		factory.CreateUpcomingTournament(
			7,
			testutil.WithTournamentName("Upcoming"),
			testutil.WithSeasonYear(time.Now().Year()),
		)

		golfer := factory.CreateGolfer()
		factory.CreateTournamentEntry(tourn1, golfer, testutil.WithEarnings(100000))
		factory.CreatePick(user, tourn1, golfer, league)

		app := testutil.NewTestApp(t).WithDBUser(user)
		app.App.Get("/leagues/:id/tournaments", leagueHandler.GetLeagueTournaments)

		resp := app.Get("/leagues/" + league.ID.String() + "/tournaments")
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result leagues.ListLeagueTournamentsResponse
		require.NoError(t, resp.JSON(&result))
		assert.Equal(t, 2, result.Total)

		var completedTourn, upcomingTourn *leagues.LeagueTournament
		for i := range result.Tournaments {
			if result.Tournaments[i].Name == "Completed" {
				completedTourn = &result.Tournaments[i]
			} else if result.Tournaments[i].Name == "Upcoming" {
				upcomingTourn = &result.Tournaments[i]
			}
		}

		require.NotNil(t, completedTourn)
		require.NotNil(t, upcomingTourn)
		assert.True(t, completedTourn.HasUserPick)
		assert.Equal(t, 1, completedTourn.PickCount)
		assert.False(t, upcomingTourn.HasUserPick)
		assert.Equal(t, 0, upcomingTourn.PickCount)
	})
}
