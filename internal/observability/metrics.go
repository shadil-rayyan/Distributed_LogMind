package observability

import (
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	LogsIngested = promauto.NewCounter(prometheus.CounterOpts{
		Name: "logmind_logs_ingested_total",
		Help: "Total logs received",
	})

	QueueDepth = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "logmind_queue_depth",
		Help: "Current log channel buffer depth",
	})

	BatchFlushDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "logmind_batch_flush_seconds",
		Help:    "Time to flush a batch to SQLite",
		Buckets: prometheus.DefBuckets,
	})

	ActiveIncidents = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "logmind_active_incidents",
		Help: "Current number of active high-severity incidents",
	})

	_ = promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "logmind_goroutines",
		Help: "Current number of goroutines",
	}, func() float64 { return float64(runtime.NumGoroutine()) })
)
