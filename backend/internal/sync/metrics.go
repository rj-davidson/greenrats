package sync

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	SyncDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "sync_duration_seconds",
		Help:    "Duration of sync operations",
		Buckets: []float64{1, 5, 15, 30, 60, 120, 300, 600},
	}, []string{"sync_type"})

	SyncRecordsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sync_records_processed_total",
		Help: "Records processed per sync",
	}, []string{"sync_type", "operation"})

	SyncErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sync_errors_total",
		Help: "Total sync errors by type",
	}, []string{"sync_type"})

	SyncRunsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sync_runs_total",
		Help: "Total sync runs by type and status",
	}, []string{"sync_type", "status"})

	LastSyncTimestamp = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sync_last_timestamp_seconds",
		Help: "Unix timestamp of last successful sync",
	}, []string{"sync_type"})
)
