package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// FingerprintGenerationTotal 指纹生成总次数
	FingerprintGenerationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fingerprint_generation_total",
			Help: "Total number of fingerprint generations",
		},
		[]string{"browser", "os"},
	)

	// FingerprintGenerationDuration 指纹生成耗时（毫秒）
	FingerprintGenerationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "fingerprint_generation_duration_ms",
			Help:    "Fingerprint generation duration in milliseconds",
			Buckets: prometheus.ExponentialBuckets(1, 2, 10), // 1, 2, 4, 8, ..., 512
		},
		[]string{"browser"},
	)

	// ProfileCacheHit 配置文件缓存命中次数
	ProfileCacheHit = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "fingerprint_profile_cache_hit_total",
			Help: "Total number of profile cache hits",
		},
	)

	// ProfileCacheMiss 配置文件缓存未命中次数
	ProfileCacheMiss = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "fingerprint_profile_cache_miss_total",
			Help: "Total number of profile cache misses",
		},
	)

	// ActiveConnections 当前活跃连接数
	ActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "fingerprint_active_connections",
			Help: "Number of currently active connections",
		},
	)

	// BehaviorAnalysisSignals 行为分析检测到的信号数
	BehaviorAnalysisSignals = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fingerprint_behavior_signals_total",
			Help: "Total number of behavior signals detected",
		},
		[]string{"risk_level"},
	)

	// HTTP2SignatureAnalysisDuration HTTP/2 签名分析耗时
	HTTP2SignatureAnalysisDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "fingerprint_http2_analysis_duration_ms",
			Help:    "HTTP/2 signature analysis duration in milliseconds",
			Buckets: prometheus.ExponentialBuckets(0.1, 2, 10),
		},
	)
)

// RecordFingerprintGeneration 记录指纹生成指标
func RecordFingerprintGeneration(browser, os string, durationMs float64) {
	FingerprintGenerationTotal.WithLabelValues(browser, os).Inc()
	FingerprintGenerationDuration.WithLabelValues(browser).Observe(durationMs)
}

// RecordProfileCacheAccess 记录配置文件缓存访问
func RecordProfileCacheAccess(hit bool) {
	if hit {
		ProfileCacheHit.Inc()
	} else {
		ProfileCacheMiss.Inc()
	}
}

// RecordBehaviorSignal 记录行为分析信号
func RecordBehaviorSignal(riskLevel string) {
	BehaviorAnalysisSignals.WithLabelValues(riskLevel).Inc()
}
