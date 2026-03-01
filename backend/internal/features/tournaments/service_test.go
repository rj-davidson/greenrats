package tournaments

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/ent/tournament"
	"github.com/rj-davidson/greenrats/internal/testutil"
)

func TestService_List(t *testing.T) {
	t.Run("returns all tournaments", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		factory.CreateTournament(testutil.WithTournamentName("Tournament A"))
		factory.CreateTournament(testutil.WithTournamentName("Tournament B"))

		resp, err := service.List(ctx, ListTournamentsRequest{})

		require.NoError(t, err)
		assert.Equal(t, 2, resp.Total)
	})

	t.Run("filters by season", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		factory.CreateTournament(testutil.WithSeasonYear(2024))
		factory.CreateTournament(testutil.WithSeasonYear(2025))

		resp, err := service.List(ctx, ListTournamentsRequest{Season: 2024})

		require.NoError(t, err)
		assert.Equal(t, 1, resp.Total)
	})

	t.Run("filters by status", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		factory.CreateUpcomingTournament(7)
		factory.CreateActiveTournament()
		factory.CreateCompletedTournament()

		resp, err := service.List(ctx, ListTournamentsRequest{Status: "active"})

		require.NoError(t, err)
		assert.Equal(t, 1, resp.Total)
	})

	t.Run("filters completed tournaments with champion", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		factory.CreateUpcomingTournament(7)
		factory.CreateActiveTournament()
		completed := factory.CreateCompletedTournament(testutil.WithTournamentName("Completed With Champion"))

		resp, err := service.List(ctx, ListTournamentsRequest{Status: "completed"})

		require.NoError(t, err)
		assert.Equal(t, 1, resp.Total)
		assert.Equal(t, completed.Name, resp.Tournaments[0].Name)
		assert.Equal(t, "completed", resp.Tournaments[0].Status)
	})

	t.Run("applies limit and offset", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		for i := range 5 {
			factory.CreateCompletedTournament(testutil.WithSeasonYear(2024 + i))
		}

		resp, err := service.List(ctx, ListTournamentsRequest{Limit: 2, Offset: 1})

		require.NoError(t, err)
		assert.Equal(t, 5, resp.Total)
		assert.Len(t, resp.Tournaments, 2)
	})
}

func TestService_GetByID(t *testing.T) {
	t.Run("returns tournament when found", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		tourn := factory.CreateTournament(testutil.WithTournamentName("Test Tournament"))

		found, err := service.GetByID(ctx, tourn.ID.String())

		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, "Test Tournament", found.Name)
	})

	t.Run("returns ErrTournamentNotFound when not found", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		found, err := service.GetByID(ctx, factory.RandomUUID().String())

		assert.Nil(t, found)
		assert.True(t, errors.Is(err, ErrTournamentNotFound))
	})

	t.Run("returns ErrInvalidTournamentID for invalid ID", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		service := NewService(db)
		ctx := context.Background()

		_, err := service.GetByID(ctx, "invalid")

		assert.True(t, errors.Is(err, ErrInvalidTournamentID))
	})

	t.Run("returns completed status when tournament has champion", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		tourn := factory.CreateCompletedTournament(testutil.WithTournamentName("Finished Tournament"))

		found, err := service.GetByID(ctx, tourn.ID.String())

		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, "completed", found.Status)
		assert.NotEmpty(t, found.ChampionID)
		assert.NotEmpty(t, found.ChampionName)
	})
}

func TestService_GetActive(t *testing.T) {
	t.Run("returns active tournament", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		factory.CreateUpcomingTournament(7)
		active := factory.CreateActiveTournament(testutil.WithTournamentName("Active Tournament"))
		factory.CreateCompletedTournament()

		found, err := service.GetActive(ctx)

		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, active.ID.String(), found.ID)
	})

	t.Run("returns nil when no active tournament", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		factory.CreateUpcomingTournament(7)
		factory.CreateCompletedTournament()

		found, err := service.GetActive(ctx)

		require.NoError(t, err)
		assert.Nil(t, found)
	})
}

func TestService_GetCurrentOrUpcoming(t *testing.T) {
	t.Run("returns upcoming tournament when pick window is open", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		upcoming := factory.CreateUpcomingTournament(7, testutil.WithTournamentName("Upcoming Tournament"))
		factory.CreateCompletedTournament()

		found, err := service.GetCurrentOrUpcoming(ctx)

		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, upcoming.ID.String(), found.ID)
		assert.Equal(t, "Upcoming Tournament", found.Name)
	})

	t.Run("returns active tournament when pick window is closed", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		active := factory.CreateActiveTournament(testutil.WithTournamentName("Active Tournament"))
		factory.CreateCompletedTournament()

		found, err := service.GetCurrentOrUpcoming(ctx)

		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, active.ID.String(), found.ID)
	})

	t.Run("returns earliest tournament without champion", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		factory.CreateUpcomingTournament(14, testutil.WithTournamentName("Later Tournament"))
		earlier := factory.CreateUpcomingTournament(7, testutil.WithTournamentName("Earlier Tournament"))
		factory.CreateCompletedTournament()

		found, err := service.GetCurrentOrUpcoming(ctx)

		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, earlier.ID.String(), found.ID)
	})

	t.Run("returns nil when all tournaments completed", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		factory.CreateCompletedTournament()
		factory.CreateCompletedTournament(testutil.WithTournamentName("Another Completed"))

		found, err := service.GetCurrentOrUpcoming(ctx)

		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("returns nil when no tournaments exist", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		service := NewService(db)
		ctx := context.Background()

		found, err := service.GetCurrentOrUpcoming(ctx)

		require.NoError(t, err)
		assert.Nil(t, found)
	})
}

func TestService_GetLeaderboard(t *testing.T) {
	t.Run("returns leaderboard entries sorted by position", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		tourn := factory.CreateCompletedTournament()
		golfer1 := factory.CreateGolfer(testutil.WithGolferName("Player One"))
		golfer2 := factory.CreateGolfer(testutil.WithGolferName("Player Two"))
		golfer3 := factory.CreateGolfer(testutil.WithGolferName("Player Three"))

		factory.CreatePlacement(tourn, golfer1, testutil.WithPosition(3), testutil.WithParRelativeScore(-5))
		factory.CreatePlacement(tourn, golfer2, testutil.WithPosition(1), testutil.WithParRelativeScore(-10))
		factory.CreatePlacement(tourn, golfer3, testutil.WithPosition(2), testutil.WithParRelativeScore(-7))

		resp, err := service.GetLeaderboard(ctx, tourn.ID.String(), false, "")

		require.NoError(t, err)
		assert.Equal(t, 3, resp.Total)
		assert.Equal(t, 1, resp.Entries[0].Position)
		assert.Equal(t, 2, resp.Entries[1].Position)
		assert.Equal(t, 3, resp.Entries[2].Position)
	})

	t.Run("marks ties correctly", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		tourn := factory.CreateCompletedTournament()
		golfer1 := factory.CreateGolfer(testutil.WithGolferName("A Player"))
		golfer2 := factory.CreateGolfer(testutil.WithGolferName("B Player"))

		factory.CreatePlacement(tourn, golfer1, testutil.WithPosition(2), testutil.WithParRelativeScore(-5))
		factory.CreatePlacement(tourn, golfer2, testutil.WithPosition(2), testutil.WithParRelativeScore(-5))

		resp, err := service.GetLeaderboard(ctx, tourn.ID.String(), false, "")

		require.NoError(t, err)
		assert.Equal(t, 2, resp.Entries[0].Position)
		assert.Equal(t, 2, resp.Entries[1].Position)
	})

	t.Run("sorts cut players separately", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		tourn := factory.CreateCompletedTournament()
		golfer1 := factory.CreateGolfer(testutil.WithGolferName("Finisher"))
		golfer2 := factory.CreateGolfer(testutil.WithGolferName("Cut Player"))

		factory.CreatePlacement(tourn, golfer1, testutil.WithPosition(10), testutil.WithParRelativeScore(-5))
		factory.CreatePlacement(tourn, golfer2, testutil.WithCut(true), testutil.WithParRelativeScore(5))

		resp, err := service.GetLeaderboard(ctx, tourn.ID.String(), false, "")

		require.NoError(t, err)
		assert.Equal(t, 2, resp.Total)
		assert.Equal(t, 10, resp.Entries[0].Position)
		assert.Equal(t, "finished", resp.Entries[0].Status)
		assert.Equal(t, 0, resp.Entries[1].Position)
		assert.Equal(t, "cut", resp.Entries[1].Status)
	})

	t.Run("returns ErrTournamentNotFound when tournament not found", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		resp, err := service.GetLeaderboard(ctx, factory.RandomUUID().String(), false, "")

		assert.Nil(t, resp)
		assert.True(t, errors.Is(err, ErrTournamentNotFound))
	})
}

func TestService_GetScorecard(t *testing.T) {
	t.Run("returns scorecard with rounds and hole scores", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		tourn := factory.CreateActiveTournament()
		golfer := factory.CreateGolfer()
		round := factory.CreateRound(tourn, golfer, 1, testutil.WithRoundScore(68))
		factory.CreateHoleScore(round, 1, 4, intPtr(4))
		factory.CreateHoleScore(round, 2, 4, intPtr(3))

		resp, err := service.GetScorecard(ctx, tourn.ID.String(), golfer.ID.String())

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, tourn.ID.String(), resp.TournamentID)
		assert.Equal(t, golfer.ID.String(), resp.GolferID)
		require.Len(t, resp.Rounds, 1)
		assert.Len(t, resp.Rounds[0].Holes, 2)
	})

	t.Run("returns empty rounds when golfer has no rounds", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		tourn := factory.CreateActiveTournament()
		golfer := factory.CreateGolfer()

		resp, err := service.GetScorecard(ctx, tourn.ID.String(), golfer.ID.String())

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.Rounds)
	})

	t.Run("returns ErrInvalidTournamentID for bad UUID", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		golfer := factory.CreateGolfer()
		_, err := service.GetScorecard(ctx, "invalid", golfer.ID.String())

		assert.True(t, errors.Is(err, ErrInvalidTournamentID))
	})

	t.Run("returns ErrInvalidGolferID for bad UUID", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		tourn := factory.CreateActiveTournament()
		_, err := service.GetScorecard(ctx, tourn.ID.String(), "invalid")

		assert.True(t, errors.Is(err, ErrInvalidGolferID))
	})

	t.Run("returns ErrTournamentNotFound for missing tournament", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		golfer := factory.CreateGolfer()
		_, err := service.GetScorecard(ctx, factory.RandomUUID().String(), golfer.ID.String())

		assert.True(t, errors.Is(err, ErrTournamentNotFound))
	})

	t.Run("returns ErrGolferNotFound for missing golfer", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		tourn := factory.CreateActiveTournament()
		_, err := service.GetScorecard(ctx, tourn.ID.String(), factory.RandomUUID().String())

		assert.True(t, errors.Is(err, ErrGolferNotFound))
	})

	t.Run("holes are sorted by hole_number", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		tourn := factory.CreateActiveTournament()
		golfer := factory.CreateGolfer()
		round := factory.CreateRound(tourn, golfer, 1)
		factory.CreateHoleScore(round, 5, 4, intPtr(4))
		factory.CreateHoleScore(round, 1, 4, intPtr(3))
		factory.CreateHoleScore(round, 3, 3, intPtr(2))

		resp, err := service.GetScorecard(ctx, tourn.ID.String(), golfer.ID.String())

		require.NoError(t, err)
		require.Len(t, resp.Rounds[0].Holes, 3)
		assert.Equal(t, 1, resp.Rounds[0].Holes[0].HoleNumber)
		assert.Equal(t, 3, resp.Rounds[0].Holes[1].HoleNumber)
		assert.Equal(t, 5, resp.Rounds[0].Holes[2].HoleNumber)
	})

	t.Run("rounds are sorted by round_number", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		tourn := factory.CreateActiveTournament()
		golfer := factory.CreateGolfer()
		factory.CreateRound(tourn, golfer, 3, testutil.WithRoundScore(70))
		factory.CreateRound(tourn, golfer, 1, testutil.WithRoundScore(68))
		factory.CreateRound(tourn, golfer, 2, testutil.WithRoundScore(72))

		resp, err := service.GetScorecard(ctx, tourn.ID.String(), golfer.ID.String())

		require.NoError(t, err)
		require.Len(t, resp.Rounds, 3)
		assert.Equal(t, 1, resp.Rounds[0].RoundNumber)
		assert.Equal(t, 2, resp.Rounds[1].RoundNumber)
		assert.Equal(t, 3, resp.Rounds[2].RoundNumber)
	})
}

func TestService_GetAllActive(t *testing.T) {
	t.Run("returns multiple active tournaments", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		factory.CreateUpcomingTournament(7)
		active1 := factory.CreateActiveTournament(testutil.WithTournamentName("Active One"))
		active2 := factory.CreateActiveTournament(testutil.WithTournamentName("Active Two"))
		factory.CreateCompletedTournament()

		found, err := service.GetAllActive(ctx)

		require.NoError(t, err)
		assert.Len(t, found, 2)
		ids := map[string]bool{found[0].ID: true, found[1].ID: true}
		assert.True(t, ids[active1.ID.String()])
		assert.True(t, ids[active2.ID.String()])
	})

	t.Run("returns empty slice when no active tournaments", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		factory.CreateUpcomingTournament(7)
		factory.CreateCompletedTournament()

		found, err := service.GetAllActive(ctx)

		require.NoError(t, err)
		assert.Empty(t, found)
	})
}

func TestService_GetAllCurrentOrUpcoming(t *testing.T) {
	t.Run("returns active tournament and tournament with open pick window", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		active := factory.CreateActiveTournament(testutil.WithTournamentName("Active"))

		pickWindowOpen := factory.CreateTournament(
			testutil.WithTournamentName("Pick Window Open"),
			testutil.WithStartDate(time.Now().AddDate(0, 0, 2)),
			testutil.WithEndDate(time.Now().AddDate(0, 0, 6)),
			testutil.WithPickWindow(time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 1)),
		)

		factory.CreateUpcomingTournament(14, testutil.WithTournamentName("Far Future"))
		factory.CreateCompletedTournament()

		found, err := service.GetAllCurrentOrUpcoming(ctx)

		require.NoError(t, err)
		assert.Len(t, found, 2)
		ids := map[string]bool{found[0].ID: true, found[1].ID: true}
		assert.True(t, ids[active.ID.String()])
		assert.True(t, ids[pickWindowOpen.ID.String()])
	})

	t.Run("excludes far-future tournaments with closed pick windows", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		active := factory.CreateActiveTournament(testutil.WithTournamentName("Active"))
		factory.CreateUpcomingTournament(14, testutil.WithTournamentName("Far Future"))

		found, err := service.GetAllCurrentOrUpcoming(ctx)

		require.NoError(t, err)
		assert.Len(t, found, 1)
		assert.Equal(t, active.ID.String(), found[0].ID)
	})

	t.Run("returns empty slice when all completed", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		factory.CreateCompletedTournament()

		found, err := service.GetAllCurrentOrUpcoming(ctx)

		require.NoError(t, err)
		assert.Empty(t, found)
	})

	t.Run("returns overlapping active tournaments sorted by start date", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		startDate1 := time.Now().AddDate(0, 0, -1)
		startDate2 := time.Now().AddDate(0, 0, -2)
		pickWindowCloses := time.Now().AddDate(0, 0, -3)

		factory.CreateTournament(
			testutil.WithTournamentName("Later Active"),
			testutil.WithStartDate(startDate1),
			testutil.WithEndDate(startDate1.AddDate(0, 0, 4)),
			testutil.WithPickWindow(startDate1.AddDate(0, 0, -5), pickWindowCloses),
		)
		factory.CreateTournament(
			testutil.WithTournamentName("Earlier Active"),
			testutil.WithStartDate(startDate2),
			testutil.WithEndDate(startDate2.AddDate(0, 0, 4)),
			testutil.WithPickWindow(startDate2.AddDate(0, 0, -5), pickWindowCloses),
		)

		found, err := service.GetAllCurrentOrUpcoming(ctx)

		require.NoError(t, err)
		assert.Len(t, found, 2)
		assert.Equal(t, "Earlier Active", found[0].Name)
		assert.Equal(t, "Later Active", found[1].Name)
	})
}

func TestTournamentToDTO(t *testing.T) {
	t.Run("converts ent tournament to DTO", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		ctx := context.Background()

		tourn := factory.CreateActiveTournament(
			testutil.WithTournamentName("Test Open"),
			testutil.WithCourse("Augusta National"),
			testutil.WithPurse(20000000),
		)

		entTourn, _ := db.Tournament.Get(ctx, tourn.ID)
		dto := toTournament(entTourn)

		assert.Equal(t, "Test Open", dto.Name)
		assert.Equal(t, "Augusta National", dto.Course)
		assert.Equal(t, float64(20000000), dto.Purse)
		assert.Equal(t, "active", dto.Status)
	})

	t.Run("converts completed tournament with champion", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		ctx := context.Background()

		tourn := factory.CreateCompletedTournament(
			testutil.WithTournamentName("Finished Open"),
		)

		entTourn, _ := db.Tournament.Query().
			Where(tournament.IDEQ(tourn.ID)).
			WithChampion().
			Only(ctx)
		dto := toTournament(entTourn)

		assert.Equal(t, "Finished Open", dto.Name)
		assert.Equal(t, "completed", dto.Status)
		assert.NotEmpty(t, dto.ChampionID)
		assert.NotEmpty(t, dto.ChampionName)
	})

	t.Run("converts completed tournament by time without champion", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		ctx := context.Background()

		startDate := time.Now().AddDate(0, 0, -10)
		endDate := startDate.AddDate(0, 0, 4)
		pickWindowCloses := startDate.Add(-1 * time.Hour)

		tourn := factory.CreateTournament(
			testutil.WithTournamentName("Past Tournament"),
			testutil.WithStartDate(startDate),
			testutil.WithEndDate(endDate),
			testutil.WithPickWindow(startDate.AddDate(0, 0, -3), pickWindowCloses),
		)

		entTourn, _ := db.Tournament.Query().
			Where(tournament.IDEQ(tourn.ID)).
			WithChampion().
			Only(ctx)
		dto := toTournament(entTourn)

		assert.Equal(t, "Past Tournament", dto.Name)
		assert.Equal(t, "completed", dto.Status)
		assert.Empty(t, dto.ChampionID)
	})
}
