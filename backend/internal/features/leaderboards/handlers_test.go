package leaderboards

import (
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/internal/testutil"
)

func TestHandler_GetLeagueLeaderboard(t *testing.T) {
	t.Run("returns leaderboard", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db, nil)
		handler := NewHandler(service)

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn := factory.CreateCompletedTournament(testutil.WithSeasonYear(time.Now().Year()))
		golfer := factory.CreateGolfer()
		factory.CreatePlacement(tourn, golfer, testutil.WithEarnings(100000))
		factory.CreatePick(user, tourn, golfer, league)

		app := testutil.NewTestApp(t).WithDBUser(user)
		app.App.Get("/leagues/:id/leaderboard", handler.GetLeagueLeaderboard)

		resp := app.Get("/leagues/" + league.ID.String() + "/leaderboard")

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result LeagueLeaderboardResponse
		require.NoError(t, resp.JSON(&result))
		assert.Equal(t, 1, result.Total)
		assert.Equal(t, 100000, result.Entries[0].Earnings)
	})

	t.Run("filters by season year query param", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db, nil)
		handler := NewHandler(service)

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, 2025)
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn := factory.CreateCompletedTournament(testutil.WithSeasonYear(2024))
		golfer := factory.CreateGolfer()
		factory.CreatePlacement(tourn, golfer, testutil.WithEarnings(100000))
		factory.CreatePick(user, tourn, golfer, league)

		app := testutil.NewTestApp(t).WithDBUser(user)
		app.App.Get("/leagues/:id/leaderboard", handler.GetLeagueLeaderboard)

		resp := app.Get("/leagues/" + league.ID.String() + "/leaderboard?season_year=2024")

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result LeagueLeaderboardResponse
		require.NoError(t, resp.JSON(&result))
		assert.Equal(t, 2024, result.SeasonYear)
		assert.Equal(t, 1, result.Total)
	})

	t.Run("returns empty leaderboard when no picks", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db, nil)
		handler := NewHandler(service)

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())

		app := testutil.NewTestApp(t)
		app.App.Get("/leagues/:id/leaderboard", handler.GetLeagueLeaderboard)

		resp := app.Get("/leagues/" + league.ID.String() + "/leaderboard")

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result LeagueLeaderboardResponse
		require.NoError(t, resp.JSON(&result))
		assert.Equal(t, 0, result.Total)
	})

	t.Run("returns 404 when league not found", func(t *testing.T) {
		t.Skip("TODO: investigate SQLite context issue with Fiber handlers")
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db, nil)
		handler := NewHandler(service)

		user := factory.CreateUser()

		app := testutil.NewTestApp(t).WithDBUser(user)
		app.App.Get("/leagues/:id/leaderboard", handler.GetLeagueLeaderboard)

		resp := app.Get("/leagues/" + factory.RandomUUID().String() + "/leaderboard")

		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("returns 400 for invalid league ID", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db, nil)
		handler := NewHandler(service)

		user := factory.CreateUser()

		app := testutil.NewTestApp(t).WithDBUser(user)
		app.App.Get("/leagues/:id/leaderboard", handler.GetLeagueLeaderboard)

		resp := app.Get("/leagues/invalid-uuid/leaderboard")

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}
