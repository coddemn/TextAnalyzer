package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	FilesProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "analyzer_files_processed_total",
			Help: "Total files processed",
		},
		[]string{"status"}, // ok / error
	)

	ProcessDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "analyzer_process_duration_seconds",
			Help:    "Time to analyze a single file",
			Buckets: prometheus.DefBuckets,
		},
	)

	RetryAttempts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "analyzer_retry_attempts_total",
			Help: "Total retry attempts",
		},
		[]string{"status"}, // success / failed
	)

	// (gauge — растёт и падает)
	ActiveWorkers = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "analyzer_active_workers",
			Help: "Currently active workers",
		},
	)
)

func Init() {
	prometheus.MustRegister(
		FilesProcessed,
		ProcessDuration,
		RetryAttempts,
		ActiveWorkers,
	)
}
