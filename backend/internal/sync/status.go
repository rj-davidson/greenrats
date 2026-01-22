package sync

import (
	"context"
	"log"
	"time"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/syncstatus"
)

func (i *Ingester) shouldSync(ctx context.Context, syncType string, interval time.Duration) bool {
	status, err := i.db.SyncStatus.Query().
		Where(syncstatus.SyncTypeEQ(syncType)).
		Only(ctx)

	if ent.IsNotFound(err) {
		return true
	}
	if err != nil {
		log.Printf("failed to check sync status for %s: %v", syncType, err)
		return true
	}

	return time.Now().After(status.LastSyncAt.Add(interval))
}

func (i *Ingester) recordSync(ctx context.Context, syncType string) {
	now := time.Now()

	existing, err := i.db.SyncStatus.Query().
		Where(syncstatus.SyncTypeEQ(syncType)).
		Only(ctx)

	if ent.IsNotFound(err) {
		_, err = i.db.SyncStatus.Create().
			SetSyncType(syncType).
			SetLastSyncAt(now).
			Save(ctx)
	} else if err == nil {
		_, err = existing.Update().
			SetLastSyncAt(now).
			Save(ctx)
	}

	if err != nil {
		log.Printf("failed to record sync for %s: %v", syncType, err)
	}
}
