//go:build instrumentation
// +build instrumentation

package extension

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// InstrumentedProcessingEngine 为 ProcessingEngine 添加可观测能力
type InstrumentedProcessingEngine struct {
	inner   *ProcessingEngine
	tracer  trace.Tracer
	logger  *zap.SugaredLogger
	metrics *ProcessingEngineMetrics
	mu      sync.RWMutex
}

// ProcessingEngineMetrics 处理引擎指标
type ProcessingEngineMetrics struct {
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	TotalDuration      time.Duration
	LastDuration       time.Duration
	LastRequestTime    time.Time

	// 按步骤的耗时
	ParseDuration     time.Duration
	AnalyzeDuration   time.Duration
	TransformDuration time.Duration
	HandleDuration    time.Duration
}

// NewInstrumentedProcessingEngine 创建可观测的处理引擎
func NewInstrumentedProcessingEngine(
	engine *ProcessingEngine,
	tracer trace.Tracer,
	logger *zap.SugaredLogger,
) *InstrumentedProcessingEngine {
	if tracer == nil {
		tracer = otel.Tracer("processing-engine")
	}

	return &InstrumentedProcessingEngine{
		inner:   engine,
		tracer:  tracer,
		logger:  logger,
		metrics: &ProcessingEngineMetrics{},
	}
}

// Process 处理扩展请求（添加追踪和指标）
func (ipe *InstrumentedProcessingEngine) Process(request *ProcessingRequest) *ProcessingResult {
	ctx := request.Context
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span := ipe.tracer.Start(ctx, "ProcessingEngine.Process")
	defer span.End()

	startTime := time.Now()
	atomic.AddInt64(&ipe.metrics.TotalRequests, 1)

	span.SetAttributes(
		attribute.String("extension_type", fmt.Sprintf("%v", request.ExtensionType)),
		attribute.StringSlice("steps", request.Steps),
		attribute.Int("raw_data_size", len(request.RawData)),
	)

	// 处理请求
	result := ipe.inner.Process(request)

	duration := time.Since(startTime)
	ipe.recordRequestMetrics(result, duration)

	span.SetAttributes(
		attribute.Bool("success", result.Success),
		attribute.String("error", result.Error),
		attribute.Int64("elapsed_ms", result.ElapsedMs),
	)

	if ipe.logger != nil {
		if result.Success {
			ipe.logger.Infow("request processed successfully",
				"steps", request.Steps,
				"duration_ms", duration.Milliseconds(),
				"results_count", len(result.AnalysisResults),
			)
		} else {
			ipe.logger.Warnw("request processing failed",
				"steps", request.Steps,
				"error", result.Error,
				"duration_ms", duration.Milliseconds(),
			)
		}
	}

	return result
}

// RegisterInterceptor 注册拦截器（代理）
func (ipe *InstrumentedProcessingEngine) RegisterInterceptor(phase string, interceptor Interceptor) error {
	return ipe.inner.RegisterInterceptor(phase, interceptor)
}

// GetConfig 获取配置（代理）
func (ipe *InstrumentedProcessingEngine) GetConfig() *EngineConfig {
	return ipe.inner.GetConfig()
}

// SetConfig 设置配置（代理）
func (ipe *InstrumentedProcessingEngine) SetConfig(config *EngineConfig) error {
	return ipe.inner.SetConfig(config)
}

// PassthroughToInner 获取底层引擎以访问原始接口
func (ipe *InstrumentedProcessingEngine) PassthroughToInner() *ProcessingEngine {
	return ipe.inner
}

// GetMetrics 获取指标快照
func (ipe *InstrumentedProcessingEngine) GetMetrics() *ProcessingEngineMetrics {
	ipe.mu.RLock()
	defer ipe.mu.RUnlock()

	// 返回快照（避免并发修改）
	snapshot := &ProcessingEngineMetrics{
		TotalRequests:      atomic.LoadInt64(&ipe.metrics.TotalRequests),
		SuccessfulRequests: atomic.LoadInt64(&ipe.metrics.SuccessfulRequests),
		FailedRequests:     atomic.LoadInt64(&ipe.metrics.FailedRequests),
		TotalDuration:      ipe.metrics.TotalDuration,
		LastDuration:       ipe.metrics.LastDuration,
		LastRequestTime:    ipe.metrics.LastRequestTime,
		ParseDuration:      ipe.metrics.ParseDuration,
		AnalyzeDuration:    ipe.metrics.AnalyzeDuration,
		TransformDuration:  ipe.metrics.TransformDuration,
		HandleDuration:     ipe.metrics.HandleDuration,
	}
	return snapshot
}

// GetMetricsSnapshot 获取 JSON 格式的指标快照
func (ipe *InstrumentedProcessingEngine) GetMetricsSnapshot() map[string]interface{} {
	metrics := ipe.GetMetrics()

	var successRate float64
	if metrics.TotalRequests > 0 {
		successRate = float64(metrics.SuccessfulRequests) / float64(metrics.TotalRequests)
	}

	return map[string]interface{}{
		"total_requests":        metrics.TotalRequests,
		"successful_requests":   metrics.SuccessfulRequests,
		"failed_requests":       metrics.FailedRequests,
		"success_rate":          successRate,
		"total_duration_ms":     metrics.TotalDuration.Milliseconds(),
		"last_duration_ms":      metrics.LastDuration.Milliseconds(),
		"parse_duration_ms":     metrics.ParseDuration.Milliseconds(),
		"analyze_duration_ms":   metrics.AnalyzeDuration.Milliseconds(),
		"transform_duration_ms": metrics.TransformDuration.Milliseconds(),
		"handle_duration_ms":    metrics.HandleDuration.Milliseconds(),
	}
}

// SetLogger 设置日志记录器
func (ipe *InstrumentedProcessingEngine) SetLogger(logger *zap.SugaredLogger) {
	ipe.mu.Lock()
	defer ipe.mu.Unlock()
	ipe.logger = logger
}

// SetTracer 设置追踪器
func (ipe *InstrumentedProcessingEngine) SetTracer(tracer trace.Tracer) {
	ipe.mu.Lock()
	defer ipe.mu.Unlock()
	ipe.tracer = tracer
}

// 私有方法

func (ipe *InstrumentedProcessingEngine) recordRequestMetrics(result *ProcessingResult, duration time.Duration) {
	ipe.mu.Lock()
	defer ipe.mu.Unlock()

	ipe.metrics.TotalDuration += duration
	ipe.metrics.LastDuration = duration
	ipe.metrics.LastRequestTime = time.Now()

	if result.Success {
		atomic.AddInt64(&ipe.metrics.SuccessfulRequests, 1)
	} else {
		atomic.AddInt64(&ipe.metrics.FailedRequests, 1)
	}
}
