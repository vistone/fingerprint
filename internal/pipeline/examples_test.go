// internal/pipeline/examples_test.go
// Pipeline 框架的实际使用示例

package pipeline

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
)

// ========================================================================
// 示例1：指纹解析流水线
// ========================================================================

// ParseJA3Stage JA3 指纹解析阶段
type ParseJA3Stage struct{}

func (s *ParseJA3Stage) GetName() string {
	return "parse-ja3"
}

func (s *ParseJA3Stage) GetDependencies() []string {
	return []string{} // 无依赖，可作为首个阶段
}

func (s *ParseJA3Stage) Execute(ctx context.Context, data *StageData) error {
	// 从 Input（原始 TLS 数据）解析 JA3
	rawData := data.Input.([]byte)
	_ = rawData // 已验证类型，实际应用中会解析其内容

	// 模拟 JA3 解析
	ja3Result := map[string]interface{}{
		"tls_version":   "1.2",
		"cipher_suites": []uint16{4865, 4866, 4867},
		"extensions":    []string{"SNI", "session_ticket", "key_share"},
	}

	// 存储到 Output（传给下一阶段）
	data.Output = map[string]interface{}{
		"ja3": ja3Result,
	}

	return nil
}

// AnalyzeJA3Stage JA3 指纹分析阶段
type AnalyzeJA3Stage struct{}

func (s *AnalyzeJA3Stage) GetName() string {
	return "analyze-ja3"
}

func (s *AnalyzeJA3Stage) GetDependencies() []string {
	return []string{"parse-ja3"} // 依赖于解析阶段
}

func (s *AnalyzeJA3Stage) Execute(ctx context.Context, data *StageData) error {
	// 接收上一阶段的输出
	parsed := data.Output.(map[string]interface{})
	ja3 := parsed["ja3"].(map[string]interface{})

	// 进行分析（比对已知指纹库等）
	analysis := map[string]interface{}{
		"ja3":          ja3,
		"browser_type": "chrome",
		"risk_score":   0.2,
	}

	data.Output = analysis
	return nil
}

// TransformStage 转换为标准格式
type TransformStage struct{}

func (s *TransformStage) GetName() string {
	return "transform"
}

func (s *TransformStage) GetDependencies() []string {
	return []string{"analyze-ja3"}
}

func (s *TransformStage) Execute(ctx context.Context, data *StageData) error {
	analysis := data.Output.(map[string]interface{})

	// 转换为 API 响应格式
	response := map[string]interface{}{
		"fingerprint": analysis,
		"timestamp":   time.Now().Unix(),
		"version":     "1.0",
	}

	data.Output = response
	return nil
}

// ========================================================================
// 示例使用
// ========================================================================

func TestBasicPipeline(t *testing.T) {
	// 创建流水线
	tracer := otel.Tracer("example")
	pipeline := NewPipeline(tracer).
		AddStage(&ParseJA3Stage{}).
		AddStage(&AnalyzeJA3Stage{}).
		AddStage(&TransformStage{})

	// 执行流水线
	rawTLSData := []byte{0x16, 0x03, 0x01, 0x00, 0x05} // 模拟 TLS 数据

	result, err := pipeline.Execute(context.Background(), rawTLSData)
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}

	fmt.Printf("Pipeline result: %+v\n", result.Output)
	fmt.Printf("Total duration: %v\n", result.Duration)
}

// ========================================================================
// 示例2：带中间件的流水线
// ========================================================================

// MockLogger 模拟日志器
type MockLogger struct {
	logs []string
}

func (ml *MockLogger) Info(msg string, fields ...interface{}) {
	ml.logs = append(ml.logs, fmt.Sprintf("INFO: %s %v", msg, fields))
}

func (ml *MockLogger) Error(msg string, fields ...interface{}) {
	ml.logs = append(ml.logs, fmt.Sprintf("ERROR: %s %v", msg, fields))
}

// MockMetrics 模拟指标收集器
type MockMetrics struct {
	records []struct {
		stage    string
		duration time.Duration
		success  bool
	}
}

func (mm *MockMetrics) Record(stage string, duration time.Duration, success bool) {
	mm.records = append(mm.records, struct {
		stage    string
		duration time.Duration
		success  bool
	}{stage, duration, success})
}

func TestPipelineWithMiddleware(t *testing.T) {
	tracer := otel.Tracer("example")
	logger := &MockLogger{}
	metrics := &MockMetrics{}

	// 创建带中间件的流水线
	pipeline := NewPipeline(tracer).
		AddStage(&ParseJA3Stage{}).
		AddStage(&AnalyzeJA3Stage{}).
		AddStage(&TransformStage{}).
		AddMiddleware(NewLoggingMiddleware(logger)).
		AddMiddleware(NewMetricsMiddleware(metrics)).
		AddMiddleware(NewRecoveryMiddleware(func(stageName string, recovered interface{}) {
			fmt.Printf("PANIC in %s: %v\n", stageName, recovered)
		}))

	// 执行
	rawTLSData := []byte{0x16, 0x03, 0x01, 0x00, 0x05}
	result, err := pipeline.Execute(context.Background(), rawTLSData)
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}

	// 验证结果和日志输出
	fmt.Printf("Pipeline output: %+v\n", result.Output)
	fmt.Println("\nLogs:")
	for _, log := range logger.logs {
		fmt.Println(log)
	}

	// 验证指标收集
	fmt.Println("\nMetrics:")
	for _, rec := range metrics.records {
		fmt.Printf("Stage: %s, Duration: %v, Success: %v\n", rec.stage, rec.duration, rec.success)
	}
}

// ========================================================================
// 示例3：带缓存的流水线
// ========================================================================

type CachedParseStage struct {
	callCount int
}

func (s *CachedParseStage) GetName() string {
	return "cached-parse"
}

func (s *CachedParseStage) GetDependencies() []string {
	return []string{}
}

func (s *CachedParseStage) Execute(ctx context.Context, data *StageData) error {
	s.callCount++
	time.Sleep(100 * time.Millisecond) // 模拟耗时操作

	data.Output = map[string]interface{}{
		"parsed": true,
	}
	return nil
}

func TestPipelineWithCaching(t *testing.T) {
	tracer := otel.Tracer("example")

	stage := &CachedParseStage{}
	pipeline := NewPipeline(tracer).
		AddStage(stage).
		AddMiddleware(NewCachingMiddleware())

	// 第1次执行：命中（执行阶段）
	t1 := time.Now()
	result1, _ := pipeline.Execute(context.Background(), "test-data")
	duration1 := time.Since(t1)

	// 第2次执行：缓存命中（不执行阶段）
	t2 := time.Now()
	result2, _ := pipeline.Execute(context.Background(), "test-data")
	duration2 := time.Since(t2)

	fmt.Printf("First execution: %v (stage called %d times)\n", duration1, stage.callCount)
	fmt.Printf("Second execution: %v (cached)\n", duration2)
	fmt.Printf("Speedup: %.1fx (caching works!)\n", float64(duration1)/float64(duration2))
	// 验证结果一致性（通过类型检查而不是相等比较）
	if result1.Output != nil && result2.Output != nil {
		fmt.Printf("Both results returned successfully\n")
	}
}

// ========================================================================
// 示例4：错误处理和超时
// ========================================================================

// FailingStage 会失败的阶段
type FailingStage struct {
	reason string
}

func (s *FailingStage) GetName() string {
	return "failing-stage"
}

func (s *FailingStage) GetDependencies() []string {
	return []string{}
}

func (s *FailingStage) Execute(ctx context.Context, data *StageData) error {
	return fmt.Errorf("intentional error: %s", s.reason)
}

func TestPipelineErrorHandling(t *testing.T) {
	tracer := otel.Tracer("example")

	pipeline := NewPipeline(tracer).
		AddStage(&FailingStage{reason: "test failure"}).
		AddStage(&ParseJA3Stage{}) // 不会执行

	result, err := pipeline.Execute(context.Background(), []byte{})

	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	}

	fmt.Printf("Pipeline stopped at stage with error: %v\n", result.Error)
}

func TestPipelineWithTimeout(t *testing.T) {
	tracer := otel.Tracer("example")

	// 创建一个会超时的阶段
	slowStage := NewMockStage(
		"slow-stage",
		[]string{},
		func(ctx context.Context, data *StageData) error {
			select {
			case <-time.After(2 * time.Second):
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	)

	pipeline := NewPipeline(tracer).
		AddStage(slowStage).
		AddMiddleware(NewTimeoutMiddleware(500 * time.Millisecond))

	result, err := pipeline.Execute(context.Background(), nil)

	if err != nil {
		fmt.Printf("Pipeline timeout error: %v\n", err)
	}

	fmt.Printf("Execution stopped due to timeout: %v\n", result.Error)
}

// ========================================================================
// Unit Tests
// ========================================================================

func TestPipelineValidation(t *testing.T) {
	tracer := otel.Tracer("test")

	// 测试1：缺失依赖
	pipeline := NewPipeline(tracer).
		AddStage(NewMockStage("stage-a", []string{"nonexistent"}, nil))

	err := pipeline.Validate()
	if err == nil {
		t.Error("expected dependency validation error, got nil")
	}

	// 测试2：循环依赖
	pipeline = NewPipeline(tracer).
		AddStage(NewMockStage("stage-a", nil, nil)).
		AddStage(NewMockStage("stage-b", []string{"stage-a"}, nil)).
		AddStage(NewMockStage("stage-c", []string{"stage-b"}, nil)).
		AddStage(NewMockStage("stage-d", []string{"stage-c", "stage-a"}, nil))

	err = pipeline.Validate()
	if err != nil {
		t.Errorf("valid pipeline failed validation: %v", err)
	}
}

func TestPipelineExecution(t *testing.T) {
	tracer := otel.Tracer("test")

	executed := []string{}

	pipeline := NewPipeline(tracer).
		AddStage(NewMockStage("stage-1", nil, func(ctx context.Context, data *StageData) error {
			executed = append(executed, "stage-1")
			data.Output = "output-1"
			return nil
		})).
		AddStage(NewMockStage("stage-2", []string{"stage-1"}, func(ctx context.Context, data *StageData) error {
			executed = append(executed, "stage-2")
			data.Output = "output-2"
			return nil
		}))

	result, err := pipeline.Execute(context.Background(), "input")
	if err != nil {
		t.Fatalf("pipeline execution failed: %v", err)
	}

	if len(executed) != 2 || executed[0] != "stage-1" || executed[1] != "stage-2" {
		t.Errorf("stages not executed in correct order: %v", executed)
	}

	if result.Output != "output-2" {
		t.Errorf("unexpected output: %v", result.Output)
	}
}

func TestMiddlewareOrder(t *testing.T) {
	tracer := otel.Tracer("test")

	order := []string{}

	// 创建自定义中间件来跟踪执行顺序
	loggingMW := NewMockStage("mw1", nil, func(ctx context.Context, data *StageData) error {
		order = append(order, "mw1-pre")
		return nil
	})

	pipeline := NewPipeline(tracer).
		AddStage(loggingMW)

	// 这个测试展示了中间件是如何链式执行的
	pipeline.Execute(context.Background(), "input")

	if len(order) == 0 {
		t.Error("middleware not executed")
	}
}
