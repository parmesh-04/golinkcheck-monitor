// metrics/metrics.go

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// ChecksTotal is a Counter for the total number of health checks performed.
	// We can add labels to distinguish between success and failure.
	ChecksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "golinkcheck_checks_total",
			Help: "The total number of health checks performed.",
		},
		[]string{"status"}, // Labels: "success" or "failure"
	)

	// CheckDuration is a Histogram to observe the duration of health checks.
	// Histograms are powerful for calculating quantiles (e.g., 95th percentile latency).
	CheckDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "golinkcheck_check_duration_seconds",
		Help:    "The duration of health checks in seconds.",
		Buckets: prometheus.LinearBuckets(0.1, 0.1, 10), // 10 buckets, starting at 0.1s, 0.1s wide
	})

	// ActiveJobs is a Gauge to track the current number of active jobs in the scheduler.
	ActiveJobs = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "golinkcheck_scheduler_active_jobs",
		Help: "The current number of active jobs in the scheduler.",
	})
)