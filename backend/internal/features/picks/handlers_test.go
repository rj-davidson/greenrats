package picks

import (
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/internal/testutil"
)

func TestHandler_Create(t *testing.T) {
	t.Run("creates pick successfully", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		handler := NewHandler(service)

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn := factory.CreateUpcomingTournament(2)
		golfer := factory.CreateGolfer()
		factory.CreateFieldEntry(tourn, golfer)

		app := testutil.NewTestApp(t).WithDBUser(user)
		app.App.Post("/picks", handler.Create)

		resp := app.Post("/picks", CreatePickRequest{
			TournamentID: tourn.ID,
			GolferID:     golfer.ID,
			LeagueID:     league.ID,
		})

		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
	})

	t.Run("returns 401 when not authenticated", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		service := NewService(db)
		handler := NewHandler(service)

		app := testutil.NewTestApp(t)
		app.App.Post("/picks", handler.Create)

		resp := app.Post("/picks", CreatePickRequest{})

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("returns 400 when missing tournament_id", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		handler := NewHandler(service)

		user := factory.CreateUser()
		golfer := factory.CreateGolfer()
		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		factory.AddUserToLeague(user, league)

		app := testutil.NewTestApp(t).WithDBUser(user)
		app.App.Post("/picks", handler.Create)

		resp := app.Post("/picks", CreatePickRequest{
			GolferID: golfer.ID,
			LeagueID: league.ID,
		})

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 404 when tournament not found", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		handler := NewHandler(service)

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)
		golfer := factory.CreateGolfer()

		app := testutil.NewTestApp(t).WithDBUser(user)
		app.App.Post("/picks", handler.Create)

		resp := app.Post("/picks", CreatePickRequest{
			TournamentID: factory.RandomUUID(),
			GolferID:     golfer.ID,
			LeagueID:     league.ID,
		})

		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("returns 409 when golfer already used", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		handler := NewHandler(service)

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn1 := factory.CreateUpcomingTournament(2)
		tourn2 := factory.CreateUpcomingTournament(2, testutil.WithTournamentName("Second"))
		golfer := factory.CreateGolfer()
		factory.CreateFieldEntry(tourn1, golfer)
		factory.CreateFieldEntry(tourn2, golfer)
		factory.CreatePick(user, tourn1, golfer, league)

		app := testutil.NewTestApp(t).WithDBUser(user)
		app.App.Post("/picks", handler.Create)

		resp := app.Post("/picks", CreatePickRequest{
			TournamentID: tourn2.ID,
			GolferID:     golfer.ID,
			LeagueID:     league.ID,
		})

		assert.Equal(t, fiber.StatusConflict, resp.StatusCode)
	})
}

func TestHandler_GetUserPicks(t *testing.T) {
	t.Run("returns user picks", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		handler := NewHandler(service)

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn := factory.CreateCompletedTournament()
		golfer := factory.CreateGolfer()
		factory.CreatePick(user, tourn, golfer, league)

		app := testutil.NewTestApp(t).WithDBUser(user)
		app.App.Get("/picks", handler.GetUserPicks)

		resp := app.Get("/picks")

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result ListPicksResponse
		require.NoError(t, resp.JSON(&result))
		assert.Equal(t, 1, result.Total)
	})

	t.Run("returns 401 when not authenticated", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		service := NewService(db)
		handler := NewHandler(service)

		app := testutil.NewTestApp(t)
		app.App.Get("/picks", handler.GetUserPicks)

		resp := app.Get("/picks")

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})
}

func TestHandler_GetPickWindow(t *testing.T) {
	t.Run("returns pick window status", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		handler := NewHandler(service)

		tourn := factory.CreateUpcomingTournament(2)

		app := testutil.NewTestApp(t)
		app.App.Get("/tournaments/:id/pick-window", handler.GetPickWindow)

		resp := app.Get("/tournaments/" + tourn.ID.String() + "/pick-window")

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result PickWindowStatus
		require.NoError(t, resp.JSON(&result))
		assert.True(t, result.IsOpen)
	})

	t.Run("returns 404 when tournament not found", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		handler := NewHandler(service)

		app := testutil.NewTestApp(t)
		app.App.Get("/tournaments/:id/pick-window", handler.GetPickWindow)

		resp := app.Get("/tournaments/" + factory.RandomUUID().String() + "/pick-window")

		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

func TestHandler_UpdateUserPick(t *testing.T) {
	t.Run("updates pick successfully", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		handler := NewHandler(service)

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

		app := testutil.NewTestApp(t).WithDBUser(user)
		app.App.Put("/picks/:id", handler.UpdateUserPick)

		resp := app.Put("/picks/"+pick.ID.String(), UpdatePickRequest{
			GolferID: golfer2.ID,
		})

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result UpdatePickResponse
		require.NoError(t, resp.JSON(&result))
		assert.Equal(t, golfer2.ID, result.Pick.GolferID)
	})

	t.Run("returns 403 when not pick owner", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		handler := NewHandler(service)

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

		app := testutil.NewTestApp(t).WithDBUser(otherUser)
		app.App.Put("/picks/:id", handler.UpdateUserPick)

		resp := app.Put("/picks/"+pick.ID.String(), UpdatePickRequest{
			GolferID: golfer2.ID,
		})

		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	})
}
