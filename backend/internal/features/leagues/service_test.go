package leagues

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/internal/testutil"
)

func TestService_Create(t *testing.T) {
	t.Run("creates league with owner membership", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()

		league, err := service.Create(ctx, CreateParams{
			Name:   "Test League",
			UserID: owner.ID,
		})

		require.NoError(t, err)
		assert.Equal(t, "Test League", league.Name)
		assert.Len(t, league.Code, 6)
		assert.Equal(t, time.Now().Year(), league.SeasonYear)
		assert.Equal(t, "owner", league.Role)
		assert.Equal(t, 1, league.MemberCount)
	})
}

func TestService_JoinLeague(t *testing.T) {
	t.Run("joins league with valid code", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()

		joined, err := service.JoinLeague(ctx, user.ID, league.Code)

		require.NoError(t, err)
		assert.Equal(t, league.ID, joined.ID)
		assert.Equal(t, "member", joined.Role)
		assert.Equal(t, 2, joined.MemberCount)
	})

	t.Run("returns error for invalid code", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		user := factory.CreateUser()

		_, err := service.JoinLeague(ctx, user.ID, "INVALID")

		require.ErrorIs(t, err, ErrInvalidJoinCode)
	})

	t.Run("returns error when already member", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		user := factory.CreateUser()
		factory.AddUserToLeague(user, league)

		_, err := service.JoinLeague(ctx, user.ID, league.Code)

		require.ErrorIs(t, err, ErrAlreadyMember)
	})

	t.Run("returns error when joining disabled", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year(), testutil.WithJoiningEnabled(false))
		user := factory.CreateUser()

		_, err := service.JoinLeague(ctx, user.ID, league.Code)

		require.ErrorIs(t, err, ErrJoiningDisabled)
	})
}

func TestService_GetByID(t *testing.T) {
	t.Run("returns league when found", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		created := factory.CreateLeague(owner, time.Now().Year(), testutil.WithLeagueName("Test League"))

		found, err := service.GetByID(ctx, created.ID)

		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, "Test League", found.Name)
	})

	t.Run("returns nil when not found", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		found, err := service.GetByID(ctx, factory.RandomUUID())

		require.NoError(t, err)
		assert.Nil(t, found)
	})
}

func TestService_GetByIDWithRole(t *testing.T) {
	t.Run("returns league with user role", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())

		found, err := service.GetByIDWithRole(ctx, league.ID, owner.ID)

		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, "owner", found.Role)
	})

	t.Run("returns league with member role", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		member := factory.CreateUser()
		factory.AddUserToLeague(member, league)

		found, err := service.GetByIDWithRole(ctx, league.ID, member.ID)

		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, "member", found.Role)
	})

	t.Run("returns league without role for non-member", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		nonMember := factory.CreateUser()

		found, err := service.GetByIDWithRole(ctx, league.ID, nonMember.ID)

		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Empty(t, found.Role)
	})
}

func TestService_ListUserLeagues(t *testing.T) {
	t.Run("returns all user leagues", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		factory.CreateLeague(owner, time.Now().Year())
		factory.CreateLeague(owner, time.Now().Year())

		resp, err := service.ListUserLeagues(ctx, owner.ID)

		require.NoError(t, err)
		assert.Equal(t, 2, resp.Total)
	})

	t.Run("returns empty for user with no leagues", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		user := factory.CreateUser()

		resp, err := service.ListUserLeagues(ctx, user.ID)

		require.NoError(t, err)
		assert.Equal(t, 0, resp.Total)
	})
}

func TestService_RegenerateJoinCode(t *testing.T) {
	t.Run("owner can regenerate code", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		oldCode := league.Code

		updated, err := service.RegenerateJoinCode(ctx, league.ID, owner.ID)

		require.NoError(t, err)
		assert.NotEqual(t, oldCode, updated.Code)
		assert.Len(t, updated.Code, 6)
	})

	t.Run("returns error when not commissioner", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		member := factory.CreateUser()
		factory.AddUserToLeague(member, league)

		_, err := service.RegenerateJoinCode(ctx, league.ID, member.ID)

		require.ErrorIs(t, err, ErrNotCommissioner)
	})
}

func TestService_SetJoiningEnabled(t *testing.T) {
	t.Run("owner can disable joining", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())

		updated, err := service.SetJoiningEnabled(ctx, league.ID, owner.ID, false)

		require.NoError(t, err)
		assert.False(t, updated.JoiningEnabled)
	})

	t.Run("owner can enable joining", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year(), testutil.WithJoiningEnabled(false))

		updated, err := service.SetJoiningEnabled(ctx, league.ID, owner.ID, true)

		require.NoError(t, err)
		assert.True(t, updated.JoiningEnabled)
	})

	t.Run("returns error when not commissioner", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		member := factory.CreateUser()
		factory.AddUserToLeague(member, league)

		_, err := service.SetJoiningEnabled(ctx, league.ID, member.ID, false)

		require.ErrorIs(t, err, ErrNotCommissioner)
	})
}

func TestService_GetLeagueTournaments(t *testing.T) {
	t.Run("returns tournaments for league season", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		factory.CreateTournament(testutil.WithSeasonYear(time.Now().Year()))
		factory.CreateTournament(testutil.WithSeasonYear(time.Now().Year()))

		resp, err := service.GetLeagueTournaments(ctx, league.ID, owner.ID)

		require.NoError(t, err)
		assert.Equal(t, 2, resp.Total)
	})

	t.Run("marks tournaments with user picks", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		owner := factory.CreateUser()
		league := factory.CreateLeague(owner, time.Now().Year())
		tourn := factory.CreateCompletedTournament(testutil.WithSeasonYear(time.Now().Year()))
		golfer := factory.CreateGolfer()
		factory.CreatePick(owner, tourn, golfer, league)

		resp, err := service.GetLeagueTournaments(ctx, league.ID, owner.ID)

		require.NoError(t, err)
		require.Equal(t, 1, resp.Total)
		assert.True(t, resp.Tournaments[0].HasUserPick)
	})

	t.Run("returns error for invalid league", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		user := factory.CreateUser()

		_, err := service.GetLeagueTournaments(ctx, factory.RandomUUID(), user.ID)

		require.ErrorIs(t, err, ErrLeagueNotFound)
	})
}
