package integration

import (
	"context"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/internal/features/picks"
	"github.com/rj-davidson/greenrats/internal/testutil"
)

func TestPicksIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := testutil.NewPostgresTestDB(ctx, t)
	factory := testutil.NewFactory(t, db)

	service := picks.NewService(db)
	handler := picks.NewHandler(service)

	t.Run("full pick flow", func(t *testing.T) {
		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn := factory.CreateUpcomingTournament(2)
		golfer1 := factory.CreateGolfer()
		golfer2 := factory.CreateGolfer()
		factory.CreateFieldEntry(tourn, golfer1)
		factory.CreateFieldEntry(tourn, golfer2)

		app := testutil.NewTestApp(t).WithDBUser(user)
		app.App.Post("/picks", handler.Create)
		app.App.Get("/picks", handler.GetUserPicks)
		app.App.Put("/picks/:id", handler.UpdateUserPick)

		resp := app.Post("/picks", picks.CreatePickRequest{
			TournamentID: tourn.ID,
			GolferID:     golfer1.ID,
			LeagueID:     league.ID,
		})
		require.Equal(t, fiber.StatusCreated, resp.StatusCode)

		var createResult picks.CreatePickResponse
		require.NoError(t, resp.JSON(&createResult))
		pickID := createResult.Pick.ID

		resp = app.Get("/picks")
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var listResult picks.ListPicksResponse
		require.NoError(t, resp.JSON(&listResult))
		assert.Equal(t, 1, listResult.Total)
		assert.Equal(t, golfer1.ID, listResult.Picks[0].GolferID)

		resp = app.Put("/picks/"+pickID.String(), picks.UpdatePickRequest{
			GolferID: golfer2.ID,
		})
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var updateResult picks.UpdatePickResponse
		require.NoError(t, resp.JSON(&updateResult))
		assert.Equal(t, golfer2.ID, updateResult.Pick.GolferID)

		resp = app.Get("/picks")
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var finalListResult picks.ListPicksResponse
		require.NoError(t, resp.JSON(&finalListResult))
		assert.Equal(t, 1, finalListResult.Total)
		assert.Equal(t, golfer2.ID, finalListResult.Picks[0].GolferID)
	})

	t.Run("golfer reuse prevention", func(t *testing.T) {
		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn1 := factory.CreateUpcomingTournament(2, testutil.WithTournamentName("Tournament 1"))
		tourn2 := factory.CreateUpcomingTournament(2, testutil.WithTournamentName("Tournament 2"))
		golfer := factory.CreateGolfer()
		factory.CreateFieldEntry(tourn1, golfer)
		factory.CreateFieldEntry(tourn2, golfer)

		app := testutil.NewTestApp(t).WithDBUser(user)
		app.App.Post("/picks", handler.Create)

		resp := app.Post("/picks", picks.CreatePickRequest{
			TournamentID: tourn1.ID,
			GolferID:     golfer.ID,
			LeagueID:     league.ID,
		})
		require.Equal(t, fiber.StatusCreated, resp.StatusCode)

		resp = app.Post("/picks", picks.CreatePickRequest{
			TournamentID: tourn2.ID,
			GolferID:     golfer.ID,
			LeagueID:     league.ID,
		})
		assert.Equal(t, fiber.StatusConflict, resp.StatusCode)
	})

	t.Run("commissioner override", func(t *testing.T) {
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

		ownerApp := testutil.NewTestApp(t).WithDBUser(owner)
		ownerApp.App.Put("/leagues/:id/picks/:pickId", handler.OverridePick)

		resp := ownerApp.Put("/leagues/"+league.ID.String()+"/picks/"+pick.ID.String(), picks.OverridePickRequest{
			GolferID: golfer2.ID,
		})
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result picks.OverridePickResponse
		require.NoError(t, resp.JSON(&result))
		assert.Equal(t, golfer2.ID, result.Pick.GolferID)
	})

	t.Run("available golfers marks used", func(t *testing.T) {
		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		tourn1 := factory.CreateCompletedTournament()
		tourn2 := factory.CreateUpcomingTournament(2)

		golfer1 := factory.CreateGolfer(testutil.WithGolferName("Used Golfer"))
		golfer2 := factory.CreateGolfer(testutil.WithGolferName("Available Golfer"))
		factory.CreateFieldEntry(tourn1, golfer1)
		factory.CreateFieldEntry(tourn2, golfer1)
		factory.CreateFieldEntry(tourn2, golfer2)

		factory.CreatePick(user, tourn1, golfer1, league)

		app := testutil.NewTestApp(t).WithDBUser(user)
		app.App.Get("/leagues/:id/available-golfers", handler.GetAvailableGolfers)

		resp := app.Get("/leagues/" + league.ID.String() + "/available-golfers?tournament_id=" + tourn2.ID.String())
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		var result picks.AvailableGolfersResponse
		require.NoError(t, resp.JSON(&result))
		assert.Equal(t, 2, result.Total)

		var usedGolfer, availableGolfer *picks.AvailableGolfer
		for i := range result.Golfers {
			switch result.Golfers[i].Name {
			case "Used Golfer":
				usedGolfer = &result.Golfers[i]
			case "Available Golfer":
				availableGolfer = &result.Golfers[i]
			}
		}

		require.NotNil(t, usedGolfer)
		require.NotNil(t, availableGolfer)
		assert.True(t, usedGolfer.IsUsed)
		assert.False(t, availableGolfer.IsUsed)
	})
}
