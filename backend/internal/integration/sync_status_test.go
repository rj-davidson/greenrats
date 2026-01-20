package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/syncstatus"
	"github.com/rj-davidson/greenrats/internal/testutil"
)

func TestSyncStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db := testutil.NewPostgresTestDB(ctx, t)

	t.Run("shouldSync returns true when no record exists", func(t *testing.T) {
		should := shouldSync(ctx, db, "new_sync_type", time.Hour)
		assert.True(t, should)
	})

	t.Run("shouldSync returns false when recently synced", func(t *testing.T) {
		syncType := "recent_sync"
		now := time.Now()

		_, err := db.SyncStatus.Create().
			SetSyncType(syncType).
			SetLastSyncAt(now).
			Save(ctx)
		require.NoError(t, err)

		should := shouldSync(ctx, db, syncType, time.Hour)
		assert.False(t, should, "should not sync when last sync was recent")
	})

	t.Run("shouldSync returns true when sync is stale", func(t *testing.T) {
		syncType := "stale_sync"
		staleTime := time.Now().Add(-2 * time.Hour)

		_, err := db.SyncStatus.Create().
			SetSyncType(syncType).
			SetLastSyncAt(staleTime).
			Save(ctx)
		require.NoError(t, err)

		should := shouldSync(ctx, db, syncType, time.Hour)
		assert.True(t, should, "should sync when last sync is older than interval")
	})

	t.Run("recordSync creates new record", func(t *testing.T) {
		syncType := "new_record_sync"

		recordSync(ctx, t, db, syncType)

		status, err := db.SyncStatus.Query().
			Where(syncstatus.SyncTypeEQ(syncType)).
			Only(ctx)
		require.NoError(t, err)
		assert.Equal(t, syncType, status.SyncType)
		assert.WithinDuration(t, time.Now(), status.LastSyncAt, time.Second)
	})

	t.Run("recordSync updates existing record", func(t *testing.T) {
		syncType := "update_record_sync"
		oldTime := time.Now().Add(-24 * time.Hour)

		_, err := db.SyncStatus.Create().
			SetSyncType(syncType).
			SetLastSyncAt(oldTime).
			Save(ctx)
		require.NoError(t, err)

		recordSync(ctx, t, db, syncType)

		status, err := db.SyncStatus.Query().
			Where(syncstatus.SyncTypeEQ(syncType)).
			Only(ctx)
		require.NoError(t, err)
		assert.WithinDuration(t, time.Now(), status.LastSyncAt, time.Second)
	})

	t.Run("full sync flow", func(t *testing.T) {
		syncType := "full_flow_sync"
		interval := time.Hour

		assert.True(t, shouldSync(ctx, db, syncType, interval), "first sync should run")

		recordSync(ctx, t, db, syncType)

		assert.False(t, shouldSync(ctx, db, syncType, interval), "immediate re-sync should be skipped")

		status, err := db.SyncStatus.Query().
			Where(syncstatus.SyncTypeEQ(syncType)).
			Only(ctx)
		require.NoError(t, err)

		_, err = status.Update().
			SetLastSyncAt(time.Now().Add(-2 * interval)).
			Save(ctx)
		require.NoError(t, err)

		assert.True(t, shouldSync(ctx, db, syncType, interval), "stale sync should run")
	})
}

func shouldSync(ctx context.Context, db *ent.Client, syncType string, interval time.Duration) bool {
	status, err := db.SyncStatus.Query().
		Where(syncstatus.SyncTypeEQ(syncType)).
		Only(ctx)

	if ent.IsNotFound(err) {
		return true
	}
	if err != nil {
		return true
	}

	return time.Now().After(status.LastSyncAt.Add(interval))
}

func recordSync(ctx context.Context, t *testing.T, db *ent.Client, syncType string) {
	t.Helper()
	now := time.Now()

	existing, err := db.SyncStatus.Query().
		Where(syncstatus.SyncTypeEQ(syncType)).
		Only(ctx)

	if ent.IsNotFound(err) {
		_, err = db.SyncStatus.Create().
			SetSyncType(syncType).
			SetLastSyncAt(now).
			Save(ctx)
	} else if err == nil {
		_, err = existing.Update().
			SetLastSyncAt(now).
			Save(ctx)
	}

	require.NoError(t, err)
}
