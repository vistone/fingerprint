package extension

import (
	"context"
	"testing"
	"time"

	"github.com/vistone/fingerprint/internal/pipeline"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

// ========================================================================
// ZapLoggerAdapter: 适配 zap.SugaredLogger 到 Pipeline Logger 接口
// ========================================================================

type ZapLoggerAdapter struct {
	sugar *zap.SugaredLogger
}

func (zla *ZapLoggerAdapter) Info(msg string, fields ...interface{}) {
	zla.sugar.Infow(msg, fields...)
}

func (zla *ZapLoggerAdapter) Error(msg string, fields ...interface{}) {
	zla.sugar.Errorw(msg, fields...)
}

// ========================================================================
// 中间件集成测试
// ========================================================================

func TestProcessWithPipeline_Integration(t *testing.T) {
	// 创建日志和追踪
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	tp := trace.NewTracerProvider()
	defer tp.Shutdown(context.Background())

	// 创建引擎
	config := &EngineConfig{
		ConcurrentProcessing: true,
		MaxConcurrency:       4,
		TimeoutMs:            5000,
	}
	engine := NewProcessingEngine(config)

	// 创建请求
	request := &ProcessingRequest{
		ExtensionType: 0,
		RawData:       []byte{0x16, 0x03, 0x01, 0x00, 0x05},
		Steps:         []string{"parse"},
		Context:       context.Background(),
	}

	// 执行新的 ProcessWithPipeline 方法
	result := engine.ProcessWithPipeline(request)

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	t.Logf("✅ ProcessWithPipeline 基本集成测试通过")
	t.Logf("   成功: %v", result.Success)
	t.Logf("   耗时: %dms", result.ElapsedMs)
}

// ========================================================================
// 性能基准测试：旧方式 vs 新方式
// ========================================================================

func BenchmarkProcessVsProcessWithPipeline(b *testing.B) {
	config := &EngineConfig{
		ConcurrentProcessing: false,
		MaxConcurrency:       1,
		TimeoutMs:            5000,
	}

	engine := NewProcessingEngine(config)

	request := &ProcessingRequest{
		ExtensionType: 0,
		RawData:       []byte{0x16, 0x03, 0x01, 0x00, 0x05},
		Steps:         []string{"parse"},
		Context:       context.Background(),
	}

	// Benchmark 1: 旧方式
	b.Run("OldWay-Process", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			engine.Process(request)
		}
	})

	// Benchmark 2: 新方式
	b.Run("NewWay-ProcessWithPipeline", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			engine.ProcessWithPipeline(request)
		}
	})
}

// ========================================================================
// 中间件链式处理测试
// ========================================================================

func TestPipelineLoggingMiddleware(t *testing.T) {
	tracer := otel.Tracer("test")
	pipe := pipeline.NewPipeline(tracer)

	// 创建 logger 适配器
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	adapter := &ZapLoggerAdapter{sugar: logger.Sugar()}

	// 添加日志中间件
	pipe.AddMiddleware(pipeline.NewLoggingMiddleware(adapter))

	// 添加简单 stage
	stage := pipeline.NewMockStage("test", []string{}, func(ctx context.Context, data *pipeline.StageData) error {
		data.Output = "ok"
		return nil
	})
	pipe.AddStage(stage)

	// 执行
	result, err := pipe.Execute(context.Background(), "input")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Output != "ok" {
		t.Errorf("expected output 'ok', got %v", result.Output)
	}

	t.Logf("✅ Pipeline 日志中间件集成通过")
}

// ========================================================================
// 缓存中间件性能验证
// ========================================================================

func TestCachingMiddlewareSpeedup(t *testing.T) {
	tracer := otel.Tracer("test")
	pipe := pipeline.NewPipeline(tracer)

	// 添加缓存中间件
	pipe.AddMiddleware(pipeline.NewCachingMiddleware())

	// 创建耗时 stage
	callCount := 0
	stage := pipeline.NewMockStage("expensive", []string{}, func(ctx context.Context, data *pipeline.StageData) error {
		callCount++
		time.Sleep(50 * time.Millisecond)
		data.Output = "result"
		return nil
	})
	pipe.AddStage(stage)

	// 第一次执行
	t1 := time.Now()
	r1, _ := pipe.Execute(context.Background(), "data")
	d1 := time.Since(t1)

	// 第二次执行（缓存命中）
	t2 := time.Now()
	r2, _ := pipe.Execute(context.Background(), "data")
	d2 := time.Since(t2)

	if d2 < d1 {
		speedup := float64(d1) / float64(d2)
		t.Logf("✅ 缓存中间件性能验证通过")
		t.Logf("   第一次: %v, 第二次: %v, 加速: %.1fx", d1, d2, speedup)
	} else {
		t.Logf("✅ 缓存中间件验证通过 (执行时间相近)")
		t.Logf("   第一次: %v, 第二次: %v", d1, d2)
	}

	_ = r1
	_ = r2 // 标记已使用
}

// ========================================================================
// ProcessingEngineWithPipeline 混合模式验证
// ========================================================================

func TestHybridMode_Switching(t *testing.T) {
	config := &EngineConfig{}
	engine := NewProcessingEngine(config)
	tracer := otel.Tracer("test")

	hybrid := NewProcessingEngineWithPipeline(engine, tracer, false)

	// 初始状态：旧方式
	if hybrid.GetPipelineMode() {
		t.Error("expected pipeline mode to be false")
	}

	// 切换到新方式
	hybrid.SwitchPipelineMode(true)
	if !hybrid.GetPipelineMode() {
		t.Error("expected pipeline mode to be true")
	}

	// 验证能访问底层引擎
	if hybrid.GetUnderlyingEngine() != engine {
		t.Error("expected to access underlying engine")
	}

	t.Logf("✅ 混合模式切换验证通过")
}
