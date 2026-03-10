package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// translated comment
	FingerprintGenerationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fingerprint_generation_total",
			Help: "Total number of fingerprint generations",
		},
		[]string{"browser", "os"},
	)

	// translated comment
	FingerprintGenerationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "fingerprint_generation_duration_ms",
			Help:    "Fingerprint generation duration in milliseconds",
			Buckets: prometheus.ExponentialBuckets(1, 2, 10), // 1, 2, 4, 8, ..., 512
		},
		[]string{"browser"},
	)

	// translated comment
	ProfileCacheHit = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "fingerprint_profile_cache_hit_total",
			Help: "Total number of profile cache hits",
		},
	)

	// translated comment
	ProfileCacheMiss = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "fingerprint_profile_cache_miss_total",
			Help: "Total number of profile cache misses",
		},
	)

	// translated comment
	ActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "fingerprint_active_connections",
			Help: "Number of currently active connections",
		},
	)

	// translated comment
	BehaviorAnalysisSignals = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fingerprint_behavior_signals_total",
			Help: "Total number of behavior signals detected",
		},
		[]string{"risk_level"},
	)

	// translated comment
	HTTP2SignatureAnalysisDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "fingerprint_http2_analysis_duration_ms",
			Help:    "HTTP/2 signature analysis duration in milliseconds",
			Buckets: prometheus.ExponentialBuckets(0.1, 2, 10),
		},
	)
)

// translated comment
func RecordFingerprintGeneration(browser, os string, durationMs float64) {
	FingerprintGenerationTotal.WithLabelValues(browser, os).Inc()
	FingerprintGenerationDuration.WithLabelValues(browser).Observe(durationMs)
}

// translated comment
func RecordProfileCacheAccess(hit bool) {
	if hit {
		ProfileCacheHit.Inc()
	} else {
		ProfileCacheMiss.Inc()
	}
}

// translated comment
func RecordBehaviorSignal(riskLevel string) {
	BehaviorAnalysisSignals.WithLabelValues(riskLevel).Inc()
}
