package leagues

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/internal/testutil"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestHandler_Create(t *testing.T) {
	t.Run("creates league successfully", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db, 2026, discardLogger())
		handler := NewHandler(service, nil, discardLogger())

		factory.EnsureSeason(2026)
		user := factory.CreateUser()

		app := testutil.NewTestApp(t).WithDBUser(user)
		app.App.Post("/leagues", handler.Create)

		resp := app.Post("/leagues", CreateLeagueRequest{
			Name: "Test League",
		})

		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

		var result CreateLeagueResponse
		require.NoError(t, resp.JSON(&result))
		assert.Equal(t, "Test League", result.League.Name)
	})

	t.Run("returns 401 when not authenticated", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		service := NewService(db, 2026, discardLogger())
		handler := NewHandler(service, nil, discardLogger())

		app := testutil.NewTestApp(t)
		app.App.Post("/leagues", handler.Create)

		resp := app.Post("/leagues", CreateLeagueRequest{Name: "Test"})

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("returns 400 when name missing", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db, 2026, discardLogger())
		handler := NewHandler(service, nil, discardLogger())

		user := factory.CreateUser()

		app := testutil.NewTestApp(t).WithDBUser(user)
		app.App.Post("/leagues", handler.Create)

		resp := app.Post("/leagues", CreateLeagueRequest{})

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

func TestHandler_ListUserLeagues(t *testing.T) {
	t.Run("returns user leagues", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db, 2026, discardLogger())
		handler := NewHandler(service, nil, discardLogger())

		user := factory.CreateUser()
		factory.CreateLeague(user, time.Now().Year())

		app := testutil.NewTestApp(t).WithDBUser(user)
		app.App.Get("/leagues", handler.ListUserLeagues)

		resp := app.Get("/leagues")

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result ListUserLeaguesResponse
		require.NoError(t, resp.JSON(&result))
		assert.Equal(t, 1, result.Total)
	})

	t.Run("returns 401 when not authenticated", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		service := NewService(db, 2026, discardLogger())
		handler := NewHandler(service, nil, discardLogger())

		app := testutil.NewTestApp(t)
		app.App.Get("/leagues", handler.ListUserLeagues)

		resp := app.Get("/leagues")

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})
}

func TestHandler_GetByID(t *testing.T) {
	t.Run("returns league when found", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db, 2026, discardLogger())
		handler := NewHandler(service, nil, discardLogger())

		user := factory.CreateUser()
		league := factory.CreateLeague(user, time.Now().Year(), testutil.WithLeagueName("Test League"))

		app := testutil.NewTestApp(t).WithDBUser(user)
		app.App.Get("/leagues/:id", handler.GetByID)

		resp := app.Get("/leagues/" + league.ID.String())

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result GetLeagueResponse
		require.NoError(t, resp.JSON(&result))
		assert.Equal(t, "Test League", result.League.Name)
	})

	t.Run("returns 404 when not found", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db, 2026, discardLogger())
		handler := NewHandler(service, nil, discardLogger())

		user := factory.CreateUser()

		app := testutil.NewTestApp(t).WithDBUser(user)
		app.App.Get("/leagues/:id", handler.GetByID)

		resp := app.Get("/leagues/" + factory.RandomUUID().String())

		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

func TestHandler_JoinLeague(t *testing.T) {
	t.Run("joins league successfully", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db, 2026, discardLogger())
		handler := NewHandler(service, nil, discardLogger())

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()

		app := testutil.NewTestApp(t).WithDBUser(user)
		app.App.Post("/leagues/join", handler.JoinLeague)

		resp := app.Post("/leagues/join", JoinLeagueRequest{Code: league.Code})

		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

		var result JoinLeagueResponse
		require.NoError(t, resp.JSON(&result))
		assert.Equal(t, league.ID, result.League.ID)
	})

	t.Run("returns 400 for invalid code", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db, 2026, discardLogger())
		handler := NewHandler(service, nil, discardLogger())

		user := factory.CreateUser()

		app := testutil.NewTestApp(t).WithDBUser(user)
		app.App.Post("/leagues/join", handler.JoinLeague)

		resp := app.Post("/leagues/join", JoinLeagueRequest{Code: "INVALID"})

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 409 when already member", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db, 2026, discardLogger())
		handler := NewHandler(service, nil, discardLogger())

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		app := testutil.NewTestApp(t).WithDBUser(user)
		app.App.Post("/leagues/join", handler.JoinLeague)

		resp := app.Post("/leagues/join", JoinLeagueRequest{Code: league.Code})

		assert.Equal(t, fiber.StatusConflict, resp.StatusCode)
	})
}

func TestHandler_RegenerateJoinCode(t *testing.T) {
	t.Run("regenerates code successfully", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db, 2026, discardLogger())
		handler := NewHandler(service, nil, discardLogger())

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		oldCode := league.Code

		app := testutil.NewTestApp(t).WithDBUser(owner)
		app.App.Post("/leagues/:id/regenerate-code", handler.RegenerateJoinCode)

		resp := app.Post("/leagues/"+league.ID.String()+"/regenerate-code", nil)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result RegenerateCodeResponse
		require.NoError(t, resp.JSON(&result))
		assert.NotEqual(t, oldCode, result.League.Code)
	})

	t.Run("returns 403 when not commissioner", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db, 2026, discardLogger())
		handler := NewHandler(service, nil, discardLogger())

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		member := factory.CreateUser()
		factory.AddUserToLeague(member, league)

		app := testutil.NewTestApp(t).WithDBUser(member)
		app.App.Post("/leagues/:id/regenerate-code", handler.RegenerateJoinCode)

		resp := app.Post("/leagues/"+league.ID.String()+"/regenerate-code", nil)

		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	})
}

func TestHandler_SetJoiningEnabled(t *testing.T) {
	t.Run("updates joining enabled successfully", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db, 2026, discardLogger())
		handler := NewHandler(service, nil, discardLogger())

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())

		app := testutil.NewTestApp(t).WithDBUser(owner)
		app.App.Put("/leagues/:id/joining-enabled", handler.SetJoiningEnabled)

		resp := app.Put("/leagues/"+league.ID.String()+"/joining-enabled", SetJoiningEnabledRequest{Enabled: false})

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result SetJoiningEnabledResponse
		require.NoError(t, resp.JSON(&result))
		assert.False(t, result.League.JoiningEnabled)
	})

	t.Run("returns 403 when not commissioner", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db, 2026, discardLogger())
		handler := NewHandler(service, nil, discardLogger())

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		member := factory.CreateUser()
		factory.AddUserToLeague(member, league)

		app := testutil.NewTestApp(t).WithDBUser(member)
		app.App.Put("/leagues/:id/joining-enabled", handler.SetJoiningEnabled)

		resp := app.Put("/leagues/"+league.ID.String()+"/joining-enabled", SetJoiningEnabledRequest{Enabled: false})

		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	})
}

func TestHandler_RemoveMember(t *testing.T) {
	t.Run("removes member successfully", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db, 2026, discardLogger())
		handler := NewHandler(service, nil, discardLogger())

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		member := factory.CreateUser()
		factory.AddUserToLeague(member, league)

		app := testutil.NewTestApp(t).WithDBUser(owner)
		app.App.Delete("/leagues/:id/members/:userId", handler.RemoveMember)

		resp := app.Delete("/leagues/" + league.ID.String() + "/members/" + member.ID.String())

		assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
	})

	t.Run("returns 401 when not authenticated", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db, 2026, discardLogger())
		handler := NewHandler(service, nil, discardLogger())

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		member := factory.CreateUser()
		factory.AddUserToLeague(member, league)

		app := testutil.NewTestApp(t)
		app.App.Delete("/leagues/:id/members/:userId", handler.RemoveMember)

		resp := app.Delete("/leagues/" + league.ID.String() + "/members/" + member.ID.String())

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("returns 403 when not commissioner", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db, 2026, discardLogger())
		handler := NewHandler(service, nil, discardLogger())

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		member := factory.CreateUser()
		factory.AddUserToLeague(member, league)
		otherMember := factory.CreateUser()
		factory.AddUserToLeague(otherMember, league)

		app := testutil.NewTestApp(t).WithDBUser(member)
		app.App.Delete("/leagues/:id/members/:userId", handler.RemoveMember)

		resp := app.Delete("/leagues/" + league.ID.String() + "/members/" + otherMember.ID.String())

		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	})

	t.Run("returns 403 when trying to remove owner", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db, 2026, discardLogger())
		handler := NewHandler(service, nil, discardLogger())

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())

		app := testutil.NewTestApp(t).WithDBUser(owner)
		app.App.Delete("/leagues/:id/members/:userId", handler.RemoveMember)

		resp := app.Delete("/leagues/" + league.ID.String() + "/members/" + owner.ID.String())

		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	})

	t.Run("returns 404 when member not found", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db, 2026, discardLogger())
		handler := NewHandler(service, nil, discardLogger())

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		nonMember := factory.CreateUser()

		app := testutil.NewTestApp(t).WithDBUser(owner)
		app.App.Delete("/leagues/:id/members/:userId", handler.RemoveMember)

		resp := app.Delete("/leagues/" + league.ID.String() + "/members/" + nonMember.ID.String())

		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})
}
