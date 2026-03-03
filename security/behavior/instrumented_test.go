package behavior

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

func TestInstrumentedBehaviorAnalyzer(t *testing.T) {
	// 创建日志记录器
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()
	defer logger.Sync()

	// 创建 tracer provider
	tp := trace.NewTracerProvider()
	tracer := tp.Tracer("test")

	// 创建原始分析器
	config := &BehaviorAnalysisConfig{
		TimeWindowSize:         60,
		MinRequestsForAnalysis: 2,
		RegularityThreshold:    0.3,
	}
	analyzer := NewBehaviorAnalyzer(config)

	// 包装成可观测版本
	observedAnalyzer := NewInstrumentedBehaviorAnalyzer(analyzer, tracer, sugar)

	// 添加一些请求行为
	for i := 0; i < 5; i++ {
		req := RequestBehavior{
			Timestamp:         time.Now().Add(time.Duration(i*100) * time.Millisecond),
			TLSVersion:        "1.3",
			CipherSuite:       "TLS_AES_256_GCM_SHA384",
			HTTPVersion:       "2",
			SourceIP:          "192.168.1.100",
			DestinationIP:     "10.0.0.1",
			SNI:               "example.com",
			ReusingConnection: i > 0,
		}
		observedAnalyzer.AddRequest(context.Background(), req)
	}

	// 验证指标
	metrics := observedAnalyzer.GetMetrics()
	if metrics.AddRequestCount != 5 {
		t.Errorf("expected 5 requests, got %d", metrics.AddRequestCount)
	}

	t.Logf("✅ AddRequest 指标: %d", metrics.AddRequestCount)

	// 分析时序模式
	pattern := observedAnalyzer.AnalyzeTemporalPattern(context.Background(), "example.com")
	if pattern == nil {
		t.Error("expected pattern, got nil")
	} else {
		t.Logf("✅ 时序模式分析完成，间隔数: %d，规律性: %.2f", len(pattern.Intervals), pattern.RegularityIndex)
	}

	// 验证分析耗时被记录
	metrics = observedAnalyzer.GetMetrics()
	if metrics.AnalysisCount == 0 {
		t.Error("expected analysis metrics to be recorded")
	} else {
		t.Logf("✅ 分析耗时记录: %d 次，最后耗时: %v", metrics.AnalysisCount, metrics.LastAnalysisDuration)
	}

	// 生成行为信号
	signals := observedAnalyzer.GenerateBehaviorSignals(context.Background(), "example.com")
	t.Logf("✅ 生成行为信号: %d 个", len(signals))

	// 获取所有信号
	allSignals := observedAnalyzer.GetAllSignals(context.Background())
	t.Logf("✅ 获取所有信号: %d 个", len(allSignals))

	// 获取风险评分
	if riskScore := observedAnalyzer.GetRiskScore(context.Background()); riskScore >= 0 {
		t.Logf("✅ 风险评分: %.2f", riskScore)
	}

	// 获取分析摘要
	summary := observedAnalyzer.GetAnalysisSummary(context.Background())
	if len(summary) > 0 {
		t.Logf("✅ 分析摘要: %s", summary[:50])
	}

	// 验证 passthrough 访问原始分析器
	originalAnalyzer := observedAnalyzer.PassthroughToInner()
	if originalAnalyzer == nil {
		t.Error("expected to access inner analyzer")
	} else {
		t.Logf("✅ 可以访问底层分析器")
	}

	// 完成追踪
	tp.ForceFlush(context.Background())
}

func TestInstrumentedBehaviorAnalyzer_Logger(t *testing.T) {
	// 测试日志功能
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()

	tracer := otel.Tracer("test")

	config := &BehaviorAnalysisConfig{
		TimeWindowSize:         60,
		MinRequestsForAnalysis: 1,
	}
	analyzer := NewBehaviorAnalyzer(config)
	observedAnalyzer := NewInstrumentedBehaviorAnalyzer(analyzer, tracer, sugar)

	// 添加请求
	req := RequestBehavior{
		Timestamp:   time.Now(),
		TLSVersion:  "1.3",
		HTTPVersion: "2",
		SourceIP:    "192.168.1.1",
		SNI:         "test.com",
	}

	observedAnalyzer.AddRequest(context.Background(), req)
	t.Logf("✅ 日志功能正常")
}

func TestGetMetrics(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()
	tracer := otel.Tracer("test")

	config := &BehaviorAnalysisConfig{TimeWindowSize: 60}
	analyzer := NewBehaviorAnalyzer(config)
	observedAnalyzer := NewInstrumentedBehaviorAnalyzer(analyzer, tracer, sugar)

	// 验证初始指标
	metrics := observedAnalyzer.GetMetrics()
	if metrics.AddRequestCount != 0 {
		t.Errorf("expected 0 initial requests, got %d", metrics.AddRequestCount)
	}

	// 添加请求并验证计数增加
	observedAnalyzer.AddRequest(context.Background(), RequestBehavior{
		Timestamp:  time.Now(),
		TLSVersion: "1.3",
	})

	metrics = observedAnalyzer.GetMetrics()
	if metrics.AddRequestCount != 1 {
		t.Errorf("expected 1 request after AddRequest, got %d", metrics.AddRequestCount)
	}

	t.Logf("✅ 指标收集正常")
}
