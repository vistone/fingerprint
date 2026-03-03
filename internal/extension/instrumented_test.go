//go:build instrumentation
// +build instrumentation

package extension

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

func TestInstrumentedProcessingEngine(t *testing.T) {
	// 创建日志记录器
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()
	defer logger.Sync()

	// 创建 tracer provider
	tp := trace.NewTracerProvider()
	tracer := tp.Tracer("test")

	// 创建原始处理引擎
	config := &EngineConfig{
		ConcurrentProcessing: true,
		MaxConcurrency:       4,
		EnableCaching:        true,
		CacheSize:            100,
		TimeoutMs:            5000,
	}
	engine := NewProcessingEngine(config)

	// 包装成可观测版本
	observedEngine := NewInstrumentedProcessingEngine(engine, tracer, sugar)

	// 处理一个请求
	request := &ProcessingRequest{
		ExtensionType: 0,
		RawData:       []byte{0x16, 0x03, 0x01},
		Steps:         []string{"parse"},
		Context:       context.Background(),
	}

	result := observedEngine.Process(request)

	if result == nil {
		t.Error("expected result, got nil")
	} else {
		t.Logf("✅ 处理完成: Success=%v, Error=%s, ElapsedMs=%d",
			result.Success, result.Error, result.ElapsedMs)
	}

	// 验证指标
	metrics := observedEngine.GetMetrics()
	if metrics.TotalRequests != 1 {
		t.Errorf("expected 1 total request, got %d", metrics.TotalRequests)
	} else {
		t.Logf("✅ 指标: 总请求=%d, 成功=%d, 失败=%d",
			metrics.TotalRequests, metrics.SuccessfulRequests, metrics.FailedRequests)
	}

	// 获取 JSON 快照
	snapshot := observedEngine.GetMetricsSnapshot()
	if snapshot == nil {
		t.Error("expected metrics snapshot, got nil")
	} else {
		t.Logf("✅ 指标快照: %v", snapshot)
	}

	// 验证代理方法
	cfg := observedEngine.GetConfig()
	if cfg == nil {
		t.Error("expected config, got nil")
	} else {
		t.Logf("✅ 获取配置: ConcurrentProcessing=%v", cfg.ConcurrentProcessing)
	}

	// 验证 passthrough
	originalEngine := observedEngine.PassthroughToInner()
	if originalEngine == nil {
		t.Error("expected to access inner engine")
	} else {
		t.Logf("✅ 可以访问底层引擎")
	}

	// 完成追踪
	tp.ForceFlush(context.Background())
}

func TestInstrumentedProcessingEngine_Metrics(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()
	tracer := otel.Tracer("test")

	config := &EngineConfig{
		TimeoutMs: 5000,
	}
	engine := NewProcessingEngine(config)
	observedEngine := NewInstrumentedProcessingEngine(engine, tracer, sugar)

	// 处理多个请求
	for i := 0; i < 3; i++ {
		request := &ProcessingRequest{
			ExtensionType: 0,
			RawData:       []byte{0x16},
			Steps:         []string{},
			Context:       context.Background(),
		}
		observedEngine.Process(request)
	}

	// 验证指标累积
	metrics := observedEngine.GetMetrics()
	if metrics.TotalRequests != 3 {
		t.Errorf("expected 3 total requests, got %d", metrics.TotalRequests)
	} else {
		t.Logf("✅ 处理了 3 个请求，累积成功")
	}

	// 验证时间戳更新
	if metrics.LastRequestTime.IsZero() {
		t.Error("expected LastRequestTime to be set")
	} else {
		t.Logf("✅ 最后请求时间被记录: %v", metrics.LastRequestTime)
	}
}

func TestInstrumentedProcessingEngine_Logger(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()
	tracer := otel.Tracer("test")

	config := &EngineConfig{}
	engine := NewProcessingEngine(config)
	observedEngine := NewInstrumentedProcessingEngine(engine, tracer, sugar)

	// 验证可以设置新的日志记录器
	newLogger, _ := zap.NewProduction()
	newSugar := newLogger.Sugar()
	observedEngine.SetLogger(newSugar)

	request := &ProcessingRequest{
		ExtensionType: 0,
		RawData:       []byte{0x16},
		Steps:         []string{"parse"},
		Context:       context.Background(),
	}
	observedEngine.Process(request)

	t.Logf("✅ 日志记录器设置和使用正常")
}

func TestInstrumentedProcessingEngine_SuccessRate(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()
	tracer := otel.Tracer("test")

	config := &EngineConfig{}
	engine := NewProcessingEngine(config)
	observedEngine := NewInstrumentedProcessingEngine(engine, tracer, sugar)

	// 处理请求（可能成功或失败）
	request := &ProcessingRequest{
		ExtensionType: 0,
		RawData:       []byte{0x16},
		Steps:         []string{},
		Context:       context.Background(),
	}

	observedEngine.Process(request)

	// 验证快照中的成功率计算
	snapshot := observedEngine.GetMetricsSnapshot()
	if successRate, ok := snapshot["success_rate"]; ok {
		t.Logf("✅ 成功率计算: %.2f", successRate)
	} else {
		t.Error("expected success_rate in snapshot")
	}
}
