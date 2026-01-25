package sync

import (
	"context"
	"time"

	"github.com/rj-davidson/greenrats/ent"
	"github.com/rj-davidson/greenrats/ent/syncstatus"
)

type SyncHealthStatus struct {
	Healthy      bool                   `json:"healthy"`
	SyncStatuses map[string]*SyncStatus `json:"sync_statuses"`
}

type SyncStatus struct {
	LastSyncAt     *time.Time `json:"last_sync_at"`
	ExpectedMaxAge string     `json:"expected_max_age"`
	IsStale        bool       `json:"is_stale"`
}

var SyncExpectedMaxAges = map[string]time.Duration{
	"tournaments": TournamentSyncInterval + (30 * time.Minute),
	"players":     7 * 24 * time.Hour,
	"placements":  PlacementSyncInterval + (30 * time.Minute),
	"earnings":    EarningsCheckInterval + (1 * time.Hour),
	"fields":      24*time.Hour + (1 * time.Hour),
	"scorecards":  ScorecardSyncInterval + (30 * time.Minute),
}

func GetHealthStatus(ctx context.Context, db *ent.Client) (*SyncHealthStatus, error) {
	statuses := make(map[string]*SyncStatus)
	allHealthy := true
	now := time.Now()

	for syncType, maxAge := range SyncExpectedMaxAges {
		status := &SyncStatus{
			ExpectedMaxAge: maxAge.String(),
		}

		dbStatus, err := db.SyncStatus.Query().
			Where(syncstatus.SyncTypeEQ(syncType)).
			Only(ctx)

		if err == nil {
			status.LastSyncAt = &dbStatus.LastSyncAt
			status.IsStale = now.After(dbStatus.LastSyncAt.Add(maxAge))
		} else {
			status.IsStale = true
		}

		if status.IsStale {
			allHealthy = false
		}

		statuses[syncType] = status
	}

	return &SyncHealthStatus{
		Healthy:      allHealthy,
		SyncStatuses: statuses,
	}, nil
}

func (i *Ingester) GetHealthStatus(ctx context.Context) (*SyncHealthStatus, error) {
	return GetHealthStatus(ctx, i.db)
}
