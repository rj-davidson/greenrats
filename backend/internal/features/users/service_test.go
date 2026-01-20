package users

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/internal/testutil"
)

func TestService_GetOrCreate(t *testing.T) {
	t.Run("creates new user when not exists", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		service := NewService(db)
		ctx := context.Background()

		result, err := service.GetOrCreate(ctx, GetOrCreateParams{
			WorkOSID: "user_new123",
			Email:    "new@example.com",
		})

		require.NoError(t, err)
		assert.True(t, result.Created)
		assert.Equal(t, "user_new123", result.User.WorkosID)
		assert.Equal(t, "new@example.com", result.User.Email)
	})

	t.Run("returns existing user when found", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		existing := factory.CreateUser(
			testutil.WithWorkOSID("user_existing123"),
			testutil.WithEmail("existing@example.com"),
		)

		result, err := service.GetOrCreate(ctx, GetOrCreateParams{
			WorkOSID: "user_existing123",
			Email:    "existing@example.com",
		})

		require.NoError(t, err)
		assert.False(t, result.Created)
		assert.Equal(t, existing.ID, result.User.ID)
	})
}

func TestService_SetDisplayName(t *testing.T) {
	t.Run("sets display name when not already set", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		user := factory.CreateUserWithoutDisplayName()

		updated, err := service.SetDisplayName(ctx, user.ID.String(), "New Name")

		require.NoError(t, err)
		require.NotNil(t, updated.DisplayName)
		assert.Equal(t, "New Name", *updated.DisplayName)
	})

	t.Run("returns error when display name already set", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		user := factory.CreateUser(testutil.WithDisplayName("Existing Name"))

		_, err := service.SetDisplayName(ctx, user.ID.String(), "New Name")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "already set")
	})

	t.Run("returns error when display name already taken", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		factory.CreateUser(testutil.WithDisplayName("Taken Name"))
		user := factory.CreateUserWithoutDisplayName()

		_, err := service.SetDisplayName(ctx, user.ID.String(), "Taken Name")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "already taken")
	})

	t.Run("returns error for invalid user ID", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		service := NewService(db)
		ctx := context.Background()

		_, err := service.SetDisplayName(ctx, "invalid-uuid", "Name")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid user ID")
	})
}

func TestService_IsDisplayNameAvailable(t *testing.T) {
	t.Run("returns true when name is available", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		service := NewService(db)
		ctx := context.Background()

		available, err := service.IsDisplayNameAvailable(ctx, "Available Name")

		require.NoError(t, err)
		assert.True(t, available)
	})

	t.Run("returns false when name is taken", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		factory.CreateUser(testutil.WithDisplayName("Taken Name"))

		available, err := service.IsDisplayNameAvailable(ctx, "Taken Name")

		require.NoError(t, err)
		assert.False(t, available)
	})

	t.Run("case insensitive check", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		factory.CreateUser(testutil.WithDisplayName("Test Name"))

		available, err := service.IsDisplayNameAvailable(ctx, "TEST NAME")

		require.NoError(t, err)
		assert.False(t, available)
	})
}

func TestService_GetByID(t *testing.T) {
	t.Run("returns user when found", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		user := factory.CreateUser()

		found, err := service.GetByID(ctx, user.ID.String())

		require.NoError(t, err)
		assert.Equal(t, user.ID, found.ID)
	})

	t.Run("returns error for invalid ID", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		service := NewService(db)
		ctx := context.Background()

		_, err := service.GetByID(ctx, "invalid")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid user ID")
	})

	t.Run("returns error when user not found", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		_, err := service.GetByID(ctx, factory.RandomUUID().String())

		require.Error(t, err)
	})
}

func TestService_GetByWorkOSID(t *testing.T) {
	t.Run("returns user when found", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		factory := testutil.NewFactory(t, db)
		service := NewService(db)
		ctx := context.Background()

		user := factory.CreateUser(testutil.WithWorkOSID("user_abc123"))

		found, err := service.GetByWorkOSID(ctx, "user_abc123")

		require.NoError(t, err)
		assert.Equal(t, user.ID, found.ID)
	})

	t.Run("returns error when user not found", func(t *testing.T) {
		db := testutil.NewTestDB(t)
		service := NewService(db)
		ctx := context.Background()

		_, err := service.GetByWorkOSID(ctx, "nonexistent")

		require.Error(t, err)
	})
}
