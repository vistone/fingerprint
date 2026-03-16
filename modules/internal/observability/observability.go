//go:build instrumentation
// +build instrumentation

// internal/observability/observability.go
// Complete observability integration (ready to use)

package observability

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// ========================================================================
// 1. Prometheus metrics registration
// ========================================================================

// FingerprintMetrics defines metrics for fingerprint generation
type FingerprintMetrics struct {
	// JA3/JA4/JA4S generation latency
	GenerationDuration prometheus.HistogramVec

	// Cache hit rate
	CacheHitRate prometheus.GaugeVec

	// Error count
	ErrorCount prometheus.CounterVec

	// Generation rate (per second)
	GenerationRate prometheus.GaugeVec
}

// NewFingerprintMetrics creates fingerprint-related metrics
func NewFingerprintMetrics() *FingerprintMetrics {
	return &FingerprintMetrics{
		GenerationDuration: *promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "fingerprint_generation_duration_seconds",
				Help:    "Time spent generating fingerprints",
				Buckets: prometheus.ExponentialBuckets(0.001, 2, 10), // 1ms to 512ms
			},
			[]string{"fingerprint_type"},
		),
		CacheHitRate: *promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "fingerprint_cache_hit_rate",
				Help: "Cache hit percentage (0-1)",
			},
			[]string{"cache_name"},
		),
		ErrorCount: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "fingerprint_errors_total",
				Help: "Total number of errors",
			},
			[]string{"error_type", "component"},
		),
		GenerationRate: *promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "fingerprint_generation_rate",
				Help: "Fingerprints generated per second",
			},
			[]string{"fingerprint_type"},
		),
	}
}

// BehaviorAnalysisMetrics defines behavior-analysis metrics
type BehaviorAnalysisMetrics struct {
	// Anomaly detection latency
	AnalysisDuration prometheus.HistogramVec

	// Number of detected anomalies
	AnomalyCount prometheus.CounterVec

	// Risk score distribution
	RiskScoreHistogram prometheus.HistogramVec

	// Number of high-risk clients
	HighRiskClientCount prometheus.GaugeVec
}

// NewBehaviorAnalysisMetrics creates behavior-analysis metrics
func NewBehaviorAnalysisMetrics() *BehaviorAnalysisMetrics {
	return &BehaviorAnalysisMetrics{
		AnalysisDuration: *promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "behavior_analysis_duration_seconds",
				Help:    "Time spent analyzing behavior",
				Buckets: prometheus.ExponentialBuckets(0.01, 2, 8), // 10ms to 1.28s
			},
			[]string{"analysis_type"},
		),
		AnomalyCount: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "behavior_anomalies_detected_total",
				Help: "Total number of anomalies detected",
			},
			[]string{"anomaly_type"},
		),
		RiskScoreHistogram: *promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "behavior_risk_score",
				Help:    "Risk score distribution (0-1)",
				Buckets: []float64{0, 0.2, 0.4, 0.6, 0.8, 1.0},
			},
			[]string{""},
		),
		HighRiskClientCount: *promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "behavior_high_risk_clients",
				Help: "Number of clients with risk score > 0.8",
			},
			[]string{},
		),
	}
}

// PipelineMetrics defines pipeline execution metrics
type PipelineMetrics struct {
	// Per-stage latency
	StageDuration prometheus.HistogramVec

	// Stage failure count
	StageErrorCount prometheus.CounterVec

	// Total pipeline duration
	TotalDuration prometheus.HistogramVec
}

// NewPipelineMetrics creates pipeline metrics
func NewPipelineMetrics() *PipelineMetrics {
	return &PipelineMetrics{
		StageDuration: *promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "pipeline_stage_duration_seconds",
				Help:    "Time spent in each pipeline stage",
				Buckets: prometheus.ExponentialBuckets(0.001, 2, 10),
			},
			[]string{"stage_name"},
		),
		StageErrorCount: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "pipeline_stage_errors_total",
				Help: "Total errors per stage",
			},
			[]string{"stage_name", "error_type"},
		),
		TotalDuration: *promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "pipeline_total_duration_seconds",
				Help:    "Total pipeline execution time",
				Buckets: prometheus.ExponentialBuckets(0.01, 2, 10),
			},
			[]string{""},
		),
	}
}

// ========================================================================
// 2. OpenTelemetry tracing integration
// ========================================================================

// TracingContext stores tracing context
type TracingContext struct {
	Tracer   trace.Tracer
	Logger   *zap.SugaredLogger
	Metadata map[string]string // request_id, user_id, etc.
}

// StartSpan starts a tracing span
func (tc *TracingContext) StartSpan(ctx context.Context, spanName string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	ctx, span := tc.Tracer.Start(ctx, spanName)

	// Attach metadata as span attributes
	for k, v := range tc.Metadata {
		span.SetAttributes(attribute.String(k, v))
	}

	// Attach custom attributes
	span.SetAttributes(attrs...)

	return ctx, span
}

// RecordEvent logs an event on span
func (tc *TracingContext) RecordEvent(span trace.Span, eventName string, attrs ...attribute.KeyValue) {
	span.AddEvent(eventName, trace.WithAttributes(attrs...))
}

// RecordError records an error
func (tc *TracingContext) RecordError(span trace.Span, err error) {
	span.RecordError(err)
}

// ========================================================================
// 3. Structured logging integration
// ========================================================================

// Logger wraps structured logging
type Logger struct {
	inner *zap.SugaredLogger
}

// NewLogger creates a logger
func NewLogger() *Logger {
	cfg := zap.NewProductionConfig()
	l, _ := cfg.Build()
	return &Logger{inner: l.Sugar()}
}

// WithContext creates a context-aware logger
func (l *Logger) WithContext(ctx context.Context) *Logger {
	fields := make([]interface{}, 0)

	// Extract standard fields from context
	if requestID := ctx.Value("request_id"); requestID != nil {
		fields = append(fields, "request_id", requestID)
	}
	if traceID := ctx.Value("trace_id"); traceID != nil {
		fields = append(fields, "trace_id", traceID)
	}

	return &Logger{inner: l.inner.With(fields...)}
}

// WithFields appends fields
func (l *Logger) WithFields(fields ...interface{}) *Logger {
	return &Logger{inner: l.inner.With(fields...)}
}

func (l *Logger) Debug(msg string, fields ...interface{}) {
	l.inner.Debugw(msg, fields...)
}

func (l *Logger) Info(msg string, fields ...interface{}) {
	l.inner.Infow(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...interface{}) {
	l.inner.Warnw(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...interface{}) {
	l.inner.Errorw(msg, fields...)
}

// ========================================================================
// 4. Observability middleware
// ========================================================================

// ObservabilityMiddleware automatically adds metrics and tracing around business logic
type ObservabilityMiddleware struct {
	tracer  trace.Tracer
	logger  *Logger
	metrics *PipelineMetrics
}

// NewObservabilityMiddleware creates middleware
func NewObservabilityMiddleware(tracer trace.Tracer, logger *Logger, metrics *PipelineMetrics) *ObservabilityMiddleware {
	return &ObservabilityMiddleware{
		tracer:  tracer,
		logger:  logger,
		metrics: metrics,
	}
}

// WrapFunctionWithMetrics adds automatic metrics collection around a function
// Example:
//
//	WrapFunctionWithMetrics("ja3.parse", func() {
//	    // JA3 parsing logic
//	})
func (om *ObservabilityMiddleware) WrapFunctionWithMetrics(
	ctx context.Context,
	functionName string,
	fn func(ctx context.Context) error,
) error {
	ctx, span := om.tracer.Start(ctx, functionName)
	defer span.End()

	startTime := time.Now()
	log := om.logger.WithContext(ctx)

	log.Info("function started", "function", functionName)

	// Execute function
	err := fn(ctx)

	duration := time.Since(startTime)

	if err != nil {
		log.Error("function failed",
			"function", functionName,
			"error", err,
			"duration_ms", duration.Milliseconds(),
		)
		om.metrics.StageErrorCount.WithLabelValues(functionName, "unknown").Inc()
		span.RecordError(err)
	} else {
		log.Info("function completed",
			"function", functionName,
			"duration_ms", duration.Milliseconds(),
		)
		om.metrics.StageDuration.WithLabelValues(functionName).Observe(duration.Seconds())
	}

	return err
}
