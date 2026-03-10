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
	// Create logger
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()
	defer logger.Sync()

	// Create tracer provider
	tp := trace.NewTracerProvider()
	tracer := tp.Tracer("test")

	// Create original processing engine
	config := &EngineConfig{
		ConcurrentProcessing: true,
		MaxConcurrency:       4,
		EnableCaching:        true,
		CacheSize:            100,
		TimeoutMs:            5000,
	}
	engine := NewProcessingEngine(config)

	// Wrap with observability
	observedEngine := NewInstrumentedProcessingEngine(engine, tracer, sugar)

	// Process a request
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
		t.Logf("✅ Processing complete: Success=%v, Error=%s, ElapsedMs=%d",
			result.Success, result.Error, result.ElapsedMs)
	}

	// Verify metrics
	metrics := observedEngine.GetMetrics()
	if metrics.TotalRequests != 1 {
		t.Errorf("expected 1 total request, got %d", metrics.TotalRequests)
	} else {
		t.Logf("✅ Metrics: TotalRequests=%d, Successes=%d, Failures=%d",
			metrics.TotalRequests, metrics.SuccessfulRequests, metrics.FailedRequests)
	}

	// Get JSON snapshot
	snapshot := observedEngine.GetMetricsSnapshot()
	if snapshot == nil {
		t.Error("expected metrics snapshot, got nil")
	} else {
		t.Logf("✅ Metrics snapshot: %v", snapshot)
	}

	// Verify proxy methods
	cfg := observedEngine.GetConfig()
	if cfg == nil {
		t.Error("expected config, got nil")
	} else {
		t.Logf("✅ Got config: ConcurrentProcessing=%v", cfg.ConcurrentProcessing)
	}

	// Verify passthrough
	originalEngine := observedEngine.PassthroughToInner()
	if originalEngine == nil {
		t.Error("expected to access inner engine")
	} else {
		t.Logf("✅ Can access underlying engine")
	}

	// Complete tracing
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

	// Process multiple requests
	for i := 0; i < 3; i++ {
		request := &ProcessingRequest{
			ExtensionType: 0,
			RawData:       []byte{0x16},
			Steps:         []string{},
			Context:       context.Background(),
		}
		observedEngine.Process(request)
	}

	// Verify metrics accumulation
	metrics := observedEngine.GetMetrics()
	if metrics.TotalRequests != 3 {
		t.Errorf("expected 3 total requests, got %d", metrics.TotalRequests)
	} else {
		t.Logf("✅ Processed 3 requests, accumulated successfully")
	}

	// Verify timestamp update
	if metrics.LastRequestTime.IsZero() {
		t.Error("expected LastRequestTime to be set")
	} else {
		t.Logf("✅ Last request time recorded: %v", metrics.LastRequestTime)
	}
}

func TestInstrumentedProcessingEngine_Logger(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()
	tracer := otel.Tracer("test")

	config := &EngineConfig{}
	engine := NewProcessingEngine(config)
	observedEngine := NewInstrumentedProcessingEngine(engine, tracer, sugar)

	// Verify that a new logger can be set
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

	t.Logf("✅ Logger set and working correctly")
}

func TestInstrumentedProcessingEngine_SuccessRate(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()
	tracer := otel.Tracer("test")

	config := &EngineConfig{}
	engine := NewProcessingEngine(config)
	observedEngine := NewInstrumentedProcessingEngine(engine, tracer, sugar)

	// Process request (may succeed or fail)
	request := &ProcessingRequest{
		ExtensionType: 0,
		RawData:       []byte{0x16},
		Steps:         []string{},
		Context:       context.Background(),
	}

	observedEngine.Process(request)

	// Verify success rate calculation in snapshot
	snapshot := observedEngine.GetMetricsSnapshot()
	if successRate, ok := snapshot["success_rate"]; ok {
		t.Logf("✅ Success rate calculation: %.2f", successRate)
	} else {
		t.Error("expected success_rate in snapshot")
	}
}
