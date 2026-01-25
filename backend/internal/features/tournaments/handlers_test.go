package tournaments

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/internal/testutil"
)

func TestHandler_List(t *testing.T) {
	t.Run("returns tournaments list", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		handler := NewHandler(service)

		factory.CreateTournament()
		factory.CreateTournament()

		app := testutil.NewTestApp(t)
		app.App.Get("/tournaments", handler.List)

		resp := app.Get("/tournaments")

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result ListTournamentsResponse
		require.NoError(t, resp.JSON(&result))
		assert.Equal(t, 2, result.Total)
	})

	t.Run("filters by query params", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		handler := NewHandler(service)

		factory.CreateUpcomingTournament(7, testutil.WithSeasonYear(2025))
		factory.CreateCompletedTournament(testutil.WithSeasonYear(2024))

		app := testutil.NewTestApp(t)
		app.App.Get("/tournaments", handler.List)

		resp := app.Get("/tournaments?status=upcoming")

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result ListTournamentsResponse
		require.NoError(t, resp.JSON(&result))
		assert.Equal(t, 1, result.Total)
	})

	t.Run("returns correct status fields for each tournament type", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		handler := NewHandler(service)

		factory.CreateUpcomingTournament(7, testutil.WithTournamentName("Upcoming"))
		factory.CreateActiveTournament(testutil.WithTournamentName("Active"))
		factory.CreateCompletedTournament(testutil.WithTournamentName("Completed"))

		app := testutil.NewTestApp(t)
		app.App.Get("/tournaments", handler.List)

		resp := app.Get("/tournaments")

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result ListTournamentsResponse
		require.NoError(t, resp.JSON(&result))
		assert.Equal(t, 3, result.Total)

		statusByName := make(map[string]string)
		for _, t := range result.Tournaments {
			statusByName[t.Name] = t.Status
		}

		assert.Equal(t, "upcoming", statusByName["Upcoming"])
		assert.Equal(t, "active", statusByName["Active"])
		assert.Equal(t, "completed", statusByName["Completed"])
	})
}

func TestHandler_GetByID(t *testing.T) {
	t.Run("returns tournament when found", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		handler := NewHandler(service)

		tourn := factory.CreateTournament(testutil.WithTournamentName("Test Tournament"))

		app := testutil.NewTestApp(t)
		app.App.Get("/tournaments/:id", handler.GetByID)

		resp := app.Get("/tournaments/" + tourn.ID.String())

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result GetTournamentResponse
		require.NoError(t, resp.JSON(&result))
		assert.Equal(t, "Test Tournament", result.Tournament.Name)
		assert.NotEmpty(t, result.Tournament.Status)
	})

	t.Run("returns completed status for tournament with champion", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		handler := NewHandler(service)

		tourn := factory.CreateCompletedTournament(testutil.WithTournamentName("Completed Tournament"))

		app := testutil.NewTestApp(t)
		app.App.Get("/tournaments/:id", handler.GetByID)

		resp := app.Get("/tournaments/" + tourn.ID.String())

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result GetTournamentResponse
		require.NoError(t, resp.JSON(&result))
		assert.Equal(t, "completed", result.Tournament.Status)
		assert.NotEmpty(t, result.Tournament.ChampionID)
	})

	t.Run("returns 404 when not found", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		handler := NewHandler(service)

		app := testutil.NewTestApp(t)
		app.App.Get("/tournaments/:id", handler.GetByID)

		resp := app.Get("/tournaments/" + factory.RandomUUID().String())

		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("returns 400 for invalid ID format", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		service := NewService(db)
		handler := NewHandler(service)

		app := testutil.NewTestApp(t)
		app.App.Get("/tournaments/:id", handler.GetByID)

		resp := app.Get("/tournaments/invalid")

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

func TestHandler_GetActive(t *testing.T) {
	t.Run("returns active tournament", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		handler := NewHandler(service)

		factory.CreateActiveTournament(testutil.WithTournamentName("Active Tournament"))

		app := testutil.NewTestApp(t)
		app.App.Get("/tournaments/active", handler.GetActive)

		resp := app.Get("/tournaments/active")

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result GetTournamentResponse
		require.NoError(t, resp.JSON(&result))
		assert.Equal(t, "Active Tournament", result.Tournament.Name)
		assert.Equal(t, "active", result.Tournament.Status)
	})

	t.Run("returns 404 when no active tournament", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		handler := NewHandler(service)

		factory.CreateUpcomingTournament(7)

		app := testutil.NewTestApp(t)
		app.App.Get("/tournaments/active", handler.GetActive)

		resp := app.Get("/tournaments/active")

		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

func TestHandler_GetLeaderboard(t *testing.T) {
	t.Run("returns leaderboard", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		handler := NewHandler(service)

		tourn := factory.CreateCompletedTournament()
		golfer := factory.CreateGolfer()
		factory.CreatePlacement(tourn, golfer, testutil.WithPosition(1))

		app := testutil.NewTestApp(t)
		app.App.Get("/tournaments/:id/leaderboard", handler.GetLeaderboard)

		resp := app.Get("/tournaments/" + tourn.ID.String() + "/leaderboard")

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result GetLeaderboardResponse
		require.NoError(t, resp.JSON(&result))
		assert.Equal(t, 1, result.Total)
	})

	t.Run("returns 404 when tournament not found", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		handler := NewHandler(service)

		app := testutil.NewTestApp(t)
		app.App.Get("/tournaments/:id/leaderboard", handler.GetLeaderboard)

		resp := app.Get("/tournaments/" + factory.RandomUUID().String() + "/leaderboard")

		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})
}
