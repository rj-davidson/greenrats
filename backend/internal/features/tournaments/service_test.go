package tournaments

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func TestService_GetLeaderboard(t *testing.T) {
	t.Run("returns leaderboard entries sorted by position", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		tourn := factory.CreateActiveTournament()
		golfer1 := factory.CreateGolfer(testutil.WithGolferName("Player One"))
		golfer2 := factory.CreateGolfer(testutil.WithGolferName("Player Two"))
		golfer3 := factory.CreateGolfer(testutil.WithGolferName("Player Three"))

		factory.CreateLeaderboardEntry(tourn, golfer1, testutil.WithPosition(3), testutil.WithScore(-5))
		factory.CreateLeaderboardEntry(tourn, golfer2, testutil.WithPosition(1), testutil.WithScore(-10))
		factory.CreateLeaderboardEntry(tourn, golfer3, testutil.WithPosition(2), testutil.WithScore(-7))

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

		tourn := factory.CreateActiveTournament()
		golfer1 := factory.CreateGolfer(testutil.WithGolferName("A Player"))
		golfer2 := factory.CreateGolfer(testutil.WithGolferName("B Player"))

		factory.CreateLeaderboardEntry(tourn, golfer1, testutil.WithPosition(2), testutil.WithScore(-5))
		factory.CreateLeaderboardEntry(tourn, golfer2, testutil.WithPosition(2), testutil.WithScore(-5))

		resp, err := service.GetLeaderboard(ctx, tourn.ID.String(), false, "")

		require.NoError(t, err)
		assert.Equal(t, "T2", resp.Entries[0].PositionDisplay)
		assert.Equal(t, "T2", resp.Entries[1].PositionDisplay)
	})

	t.Run("sorts cut players separately", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		tourn := factory.CreateCompletedTournament()
		golfer1 := factory.CreateGolfer(testutil.WithGolferName("Finisher"))
		golfer2 := factory.CreateGolfer(testutil.WithGolferName("Cut Player"))

		factory.CreateLeaderboardEntry(tourn, golfer1, testutil.WithPosition(10), testutil.WithScore(-5))
		factory.CreateLeaderboardEntry(tourn, golfer2, testutil.WithCut(true), testutil.WithScore(5))

		resp, err := service.GetLeaderboard(ctx, tourn.ID.String(), false, "")

		require.NoError(t, err)
		assert.Equal(t, 2, resp.Total)
		assert.Equal(t, "10", resp.Entries[0].PositionDisplay)
		assert.Equal(t, "CUT", resp.Entries[1].PositionDisplay)
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
}
