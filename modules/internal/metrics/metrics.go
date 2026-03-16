package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// FingerprintGenerationTotal counts fingerprint generations
	FingerprintGenerationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fingerprint_generation_total",
			Help: "Total number of fingerprint generations",
		},
		[]string{"browser", "os"},
	)

	// FingerprintGenerationDuration tracks generation duration (milliseconds)
	FingerprintGenerationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "fingerprint_generation_duration_ms",
			Help:    "Fingerprint generation duration in milliseconds",
			Buckets: prometheus.ExponentialBuckets(1, 2, 10), // 1, 2, 4, 8, ..., 512
		},
		[]string{"browser"},
	)

	// ProfileCacheHit counts profile cache hits
	ProfileCacheHit = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "fingerprint_profile_cache_hit_total",
			Help: "Total number of profile cache hits",
		},
	)

	// ProfileCacheMiss counts profile cache misses
	ProfileCacheMiss = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "fingerprint_profile_cache_miss_total",
			Help: "Total number of profile cache misses",
		},
	)

	// ActiveConnections tracks current active connections
	ActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "fingerprint_active_connections",
			Help: "Number of currently active connections",
		},
	)

	// BehaviorAnalysisSignals counts detected behavior-analysis signals
	BehaviorAnalysisSignals = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fingerprint_behavior_signals_total",
			Help: "Total number of behavior signals detected",
		},
		[]string{"risk_level"},
	)

	// HTTP2SignatureAnalysisDuration tracks HTTP/2 signature analysis duration
	HTTP2SignatureAnalysisDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "fingerprint_http2_analysis_duration_ms",
			Help:    "HTTP/2 signature analysis duration in milliseconds",
			Buckets: prometheus.ExponentialBuckets(0.1, 2, 10),
		},
	)
)

// RecordFingerprintGeneration records fingerprint-generation metrics
func RecordFingerprintGeneration(browser, os string, durationMs float64) {
	FingerprintGenerationTotal.WithLabelValues(browser, os).Inc()
	FingerprintGenerationDuration.WithLabelValues(browser).Observe(durationMs)
}

// RecordProfileCacheAccess records profile cache access
func RecordProfileCacheAccess(hit bool) {
	if hit {
		ProfileCacheHit.Inc()
	} else {
		ProfileCacheMiss.Inc()
	}
}

// RecordBehaviorSignal records behavior-analysis signal
func RecordBehaviorSignal(riskLevel string) {
	BehaviorAnalysisSignals.WithLabelValues(riskLevel).Inc()
}
