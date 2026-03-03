// internal/observability/observability.go
// 完整的可观测性整合（现在就能用）

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
// 1. Prometheus 指标注册
// ========================================================================

// FingerprintMetrics 指纹生成相关指标
type FingerprintMetrics struct {
	// JA3/JA4/JA4S 生成耗时
	GenerationDuration prometheus.HistogramVec

	// 缓存命中率
	CacheHitRate prometheus.GaugeVec

	// 错误计数
	ErrorCount prometheus.CounterVec

	// 生成速率（每秒）
	GenerationRate prometheus.GaugeVec
}

// NewFingerprintMetrics 创建指纹相关指标
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

// BehaviorAnalysisMetrics 行为分析相关指标
type BehaviorAnalysisMetrics struct {
	// 异常检测延迟
	AnalysisDuration prometheus.HistogramVec

	// 检测到的异常数
	AnomalyCount prometheus.CounterVec

	// 风险评分分布
	RiskScoreHistogram prometheus.HistogramVec

	// 高风险客户端数
	HighRiskClientCount prometheus.GaugeVec
}

// NewBehaviorAnalysisMetrics 创建行为分析指标
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

// PipelineMetrics 管道执行相关指标
type PipelineMetrics struct {
	// 各阶段耗时
	StageDuration prometheus.HistogramVec

	// 阶段失败计数
	StageErrorCount prometheus.CounterVec

	// 总管道耗时
	TotalDuration prometheus.HistogramVec
}

// NewPipelineMetrics 创建管道指标
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
// 2. OpenTelemetry 追踪集成
// ========================================================================

// TracingContext 追踪上下文
type TracingContext struct {
	Tracer   trace.Tracer
	Logger   *zap.SugaredLogger
	Metadata map[string]string // request_id, user_id 等
}

// StartSpan 开启追踪 span
func (tc *TracingContext) StartSpan(ctx context.Context, spanName string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	ctx, span := tc.Tracer.Start(ctx, spanName)

	// 添加元数据作为 span 属性
	for k, v := range tc.Metadata {
		span.SetAttributes(attribute.String(k, v))
	}

	// 添加自定义属性
	span.SetAttributes(attrs...)

	return ctx, span
}

// RecordEvent 在 span 中记录事件
func (tc *TracingContext) RecordEvent(span trace.Span, eventName string, attrs ...attribute.KeyValue) {
	span.AddEvent(eventName, trace.WithAttributes(attrs...))
}

// RecordError 记录错误
func (tc *TracingContext) RecordError(span trace.Span, err error) {
	span.RecordError(err)
}

// ========================================================================
// 3. 结构化日志集成
// ========================================================================

// Logger 结构化日志包装
type Logger struct {
	inner *zap.SugaredLogger
}

// NewLogger 创建日志器
func NewLogger() *Logger {
	cfg := zap.NewProductionConfig()
	l, _ := cfg.Build()
	return &Logger{inner: l.Sugar()}
}

// WithContext 创建带上下文的日志器
func (l *Logger) WithContext(ctx context.Context) *Logger {
	fields := make([]interface{}, 0)

	// 从上下文提取标准字段
	if requestID := ctx.Value("request_id"); requestID != nil {
		fields = append(fields, "request_id", requestID)
	}
	if traceID := ctx.Value("trace_id"); traceID != nil {
		fields = append(fields, "trace_id", traceID)
	}

	return &Logger{inner: l.inner.With(fields...)}
}

// WithFields 添加字段
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
// 4. 可观测性中间件
// ========================================================================

// ObservabilityMiddleware 为业务逻辑自动添加指标和追踪
type ObservabilityMiddleware struct {
	tracer  trace.Tracer
	logger  *Logger
	metrics *PipelineMetrics
}

// NewObservabilityMiddleware 创建中间件
func NewObservabilityMiddleware(tracer trace.Tracer, logger *Logger, metrics *PipelineMetrics) *ObservabilityMiddleware {
	return &ObservabilityMiddleware{
		tracer:  tracer,
		logger:  logger,
		metrics: metrics,
	}
}

// WrapFunctionWithMetrics 为函数添加自动指标收集
// 示例：
//
//	WrapFunctionWithMetrics("ja3.parse", func() {
//	    // ya3 解析逻辑
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

	// 执行函数
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
// 5. 使用示例（立即可集成到项目）
// ========================================================================

/*
在实际项目中的使用：

package main

import (
	"context"
	"github.com/vistone/fingerprint/internal/observability"
)

func init() {
	// 初始化 observability
	tracer := otel.Tracer("fingerprint")
	logger := observability.NewLogger()
	metrics := observability.NewPipelineMetrics()

	middleware := observability.NewObservabilityMiddleware(tracer, logger, metrics)
}

// 在 BehaviorAnalyzer 中
func (ba *BehaviorAnalyzer) Analyze(ctx context.Context, clientIP string) error {
	return middleware.WrapFunctionWithMetrics(
		ctx,
		"behavior.analyze",
		func(ctx context.Context) error {
			// 原有的分析逻辑
			return ba.analyzeImpl(ctx, clientIP)
		},
	)
}

// 在 ProcessingEngine 中
func (pe *ProcessingEngine) Process(ctx context.Context, req *ProcessingRequest) error {
	return middleware.WrapFunctionWithMetrics(
		ctx,
		"processing.process",
		func(ctx context.Context) error {
			// 原有的处理逻辑
			return pe.processImpl(ctx, req)
		},
	)
}

// 使用 Prometheus /metrics 端口
import "net/http"
import "github.com/prometheus/client_golang/prometheus/promhttp"

func init() {
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":8080", nil)
}
*/
