package balldontlie

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	APIRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "bdl_api_requests_total",
		Help: "Total BallDontLie API requests",
	}, []string{"endpoint", "status"})

	APIRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "bdl_api_request_duration_seconds",
		Help:    "BallDontLie API request duration",
		Buckets: prometheus.DefBuckets,
	}, []string{"endpoint"})

	CircuitBreakerState = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "bdl_circuit_breaker_state",
		Help: "BallDontLie circuit breaker state (0=closed, 1=half-open, 2=open)",
	}, []string{"name"})
)
