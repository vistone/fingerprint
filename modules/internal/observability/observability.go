//go:build instrumentation
// +build instrumentation

// internal/observability/observability.go
// translated comment

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
// translated comment
// ========================================================================

// translated comment
type FingerprintMetrics struct {
	// translated comment
	GenerationDuration prometheus.HistogramVec

	// translated comment
	CacheHitRate prometheus.GaugeVec

	// translated comment
	ErrorCount prometheus.CounterVec

	// translated comment
	GenerationRate prometheus.GaugeVec
}

// translated comment
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

// translated comment
type BehaviorAnalysisMetrics struct {
	// translated comment
	AnalysisDuration prometheus.HistogramVec

	// translated comment
	AnomalyCount prometheus.CounterVec

	// translated comment
	RiskScoreHistogram prometheus.HistogramVec

	// translated comment
	HighRiskClientCount prometheus.GaugeVec
}

// translated comment
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

// translated comment
type PipelineMetrics struct {
	// translated comment
	StageDuration prometheus.HistogramVec

	// translated comment
	StageErrorCount prometheus.CounterVec

	// translated comment
	TotalDuration prometheus.HistogramVec
}

// translated comment
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
// translated comment
// ========================================================================

// translated comment
type TracingContext struct {
	Tracer   trace.Tracer
	Logger   *zap.SugaredLogger
	Metadata map[string]string // translated comment
}

// translated comment
func (tc *TracingContext) StartSpan(ctx context.Context, spanName string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	ctx, span := tc.Tracer.Start(ctx, spanName)

	// translated comment
	for k, v := range tc.Metadata {
		span.SetAttributes(attribute.String(k, v))
	}

	// translated comment
	span.SetAttributes(attrs...)

	return ctx, span
}

// translated comment
func (tc *TracingContext) RecordEvent(span trace.Span, eventName string, attrs ...attribute.KeyValue) {
	span.AddEvent(eventName, trace.WithAttributes(attrs...))
}

// translated comment
func (tc *TracingContext) RecordError(span trace.Span, err error) {
	span.RecordError(err)
}

// ========================================================================
// translated comment
// ========================================================================

// translated comment
type Logger struct {
	inner *zap.SugaredLogger
}

// translated comment
func NewLogger() *Logger {
	cfg := zap.NewProductionConfig()
	l, _ := cfg.Build()
	return &Logger{inner: l.Sugar()}
}

// translated comment
func (l *Logger) WithContext(ctx context.Context) *Logger {
	fields := make([]interface{}, 0)

	// translated comment
	if requestID := ctx.Value("request_id"); requestID != nil {
		fields = append(fields, "request_id", requestID)
	}
	if traceID := ctx.Value("trace_id"); traceID != nil {
		fields = append(fields, "trace_id", traceID)
	}

	return &Logger{inner: l.inner.With(fields...)}
}

// translated comment
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
// translated comment
// ========================================================================

// translated comment
type ObservabilityMiddleware struct {
	tracer  trace.Tracer
	logger  *Logger
	metrics *PipelineMetrics
}

// translated comment
func NewObservabilityMiddleware(tracer trace.Tracer, logger *Logger, metrics *PipelineMetrics) *ObservabilityMiddleware {
	return &ObservabilityMiddleware{
		tracer:  tracer,
		logger:  logger,
		metrics: metrics,
	}
}

// translated comment
// translated comment
//
//	WrapFunctionWithMetrics("ja3.parse", func() {
// translated comment
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

	// translated comment
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

// ========================================================================
// translated comment
// ========================================================================

/*
在实际项目中的使用：

package main

import (
	"context"
	"github.com/vistone/fingerprint/modules/internal/observability"
)

func init() {
	// translated comment
	tracer := otel.Tracer("fingerprint")
	logger := observability.NewLogger()
	metrics := observability.NewPipelineMetrics()

	middleware := observability.NewObservabilityMiddleware(tracer, logger, metrics)
}

// translated comment
func (ba *BehaviorAnalyzer) Analyze(ctx context.Context, clientIP string) error {
	return middleware.WrapFunctionWithMetrics(
		ctx,
		"behavior.analyze",
		func(ctx context.Context) error {
			// translated comment
			return ba.analyzeImpl(ctx, clientIP)
		},
	)
}

// translated comment
func (pe *ProcessingEngine) Process(ctx context.Context, req *ProcessingRequest) error {
	return middleware.WrapFunctionWithMetrics(
		ctx,
		"processing.process",
		func(ctx context.Context) error {
			// translated comment
			return pe.processImpl(ctx, req)
		},
	)
}

// translated comment
import "net/http"
import "github.com/prometheus/client_golang/prometheus/promhttp"

func init() {
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":8080", nil)
}
*/
