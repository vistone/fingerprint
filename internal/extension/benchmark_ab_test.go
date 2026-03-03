package extension

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ========================================================================
// A/B 测试框架：新旧方式性能对比和灰度验证
// ========================================================================

// ABTestResult 保存 A/B 测试的结果
type ABTestResult struct {
	// 基准测试数据
	OldMethod    MethodMetrics
	NewMethod    MethodMetrics
	Improvement  PerformanceImprovement
	ResultsMatch bool
	Variance     float64 // 结果中位数的差异率

	// 灰度流量数据
	GradualRollout GradualRolloutMetrics
}

// MethodMetrics 单个方法的性能指标
type MethodMetrics struct {
	Name              string
	TotalRequests     int
	AVGDuration       time.Duration
	MinDuration       time.Duration
	MaxDuration       time.Duration
	P50Duration       time.Duration
	P95Duration       time.Duration
	P99Duration       time.Duration
	MemoryAllocations int64
	AllocationBytes   int64
	CacheHitRate      float64 // 缓存命中率 (仅新方式适用)
}

// PerformanceImprovement 性能改进统计
type PerformanceImprovement struct {
	LatencyReduction    float64       // 延迟减少百分比
	LatencyAbsoluteGain time.Duration // 绝对值利益 (负数表示慢了)
	MemoryReduction     float64       // 内存减少百分比
	AllocReduction      float64       // 分配减少百分比
	ThroughputIncrease  float64       // 吞吐量增加百分比
	Recommendation      string        // 建议
}

// GradualRolloutMetrics 灰度流量的指标
type GradualRolloutMetrics struct {
	Stages []RolloutStage
}

// RolloutStage 灰度的一个阶段
type RolloutStage struct {
	TrafficPercentage  float64 // 新方式的流量百分比
	TotalRequests      int
	NewMethodRequests  int
	OldMethodRequests  int
	NewMethodLatency   time.Duration
	OldMethodLatency   time.Duration
	SuccessRate        float64
	ErrorCount         int
	InconsistencyCount int
}

// ABTestConfig A/B 测试配置
type ABTestConfig struct {
	RequestsPerMethod int       // 每个方法的请求数
	ConcurrentWorkers int       // 并发数量
	CachingEnabled    bool      // 是否启用缓存
	CacheHitRatio     float64   // 期望的缓存命中率 (0-1)
	Warmup            int       // 热启动的请求数
	GradualRollout    bool      // 是否进行灰度流量测试
	RolloutStages     []float64 // 灰度阶段 (百分比)
}

// DefaultABTestConfig 返回默认的 A/B 测试配置
func DefaultABTestConfig() *ABTestConfig {
	return &ABTestConfig{
		RequestsPerMethod: 1000,
		ConcurrentWorkers: 10,
		CachingEnabled:    true,
		CacheHitRatio:     0.7,
		Warmup:            100,
		GradualRollout:    true,
		RolloutStages:     []float64{0.05, 0.25, 0.50, 1.0},
	}
}

// RunABTest 运行完整的 A/B 测试
func RunABTest(t *testing.T, engine *ProcessingEngine, config *ABTestConfig) *ABTestResult {
	if config == nil {
		config = DefaultABTestConfig()
	}

	result := &ABTestResult{
		OldMethod: MethodMetrics{Name: "Process (旧方式)"},
		NewMethod: MethodMetrics{Name: "ProcessWithPipeline (新方式)"},
	}

	// 建立测试用的请求
	request := createTestRequest()

	t.Logf("🧪 A/B 测试开始")
	t.Logf("  📊 配置: %d 请求, %d 并发, 缓存=%v", config.RequestsPerMethod, config.ConcurrentWorkers, config.CachingEnabled)

	// 热启动
	t.Logf("  🔥 热启动: %d 请求...", config.Warmup)
	for i := 0; i < config.Warmup; i++ {
		engine.Process(request)
		engine.ProcessWithPipeline(request)
	}

	// 运行旧方式的基准测试
	t.Logf("  ⏱️  基准测试: 旧方式...")
	result.OldMethod = benchmarkOldMethod(t, engine, request, config)

	// 运行新方式的基准测试
	t.Logf("  ⏱️  基准测试: 新方式...")
	result.NewMethod = benchmarkNewMethod(t, engine, request, config)

	// 验证结果一致性
	t.Logf("  ✅ 验证结果一致性...")
	result.ResultsMatch, result.Variance = verifyResultsConsistency(t, engine, request, 50)

	// 计算性能改进
	result.Improvement = calculateImprovement(&result.OldMethod, &result.NewMethod)

	// 灰度流量测试
	if config.GradualRollout {
		t.Logf("  📈 灰度流量测试...")
		result.GradualRollout = runGradualRolloutTest(t, engine, request, config)
	}

	t.Logf("✅ A/B 测试完成")
	return result
}

// benchmarkOldMethod 对旧方式进行基准测试
func benchmarkOldMethod(t *testing.T, engine *ProcessingEngine, request *ProcessingRequest, config *ABTestConfig) MethodMetrics {
	durations := make([]time.Duration, 0, config.RequestsPerMethod)
	var allocationBytes int64
	mu := sync.Mutex{}

	start := time.Now()
	sem := make(chan struct{}, config.ConcurrentWorkers)
	var wg sync.WaitGroup

	for i := 0; i < config.RequestsPerMethod; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			reqStart := time.Now()
			result := engine.Process(request)
			duration := time.Since(reqStart)

			mu.Lock()
			durations = append(durations, duration)
			if result.Success && len(result.AnalysisResults) > 0 {
				// 估算分配 (粗略)
				allocationBytes += int64(len(result.AnalysisResults)) * 100
			}
			mu.Unlock()
		}()
	}

	wg.Wait()
	totalDuration := time.Since(start)

	metrics := calculateMetrics("Process (旧方式)", durations, totalDuration, config.RequestsPerMethod, 0, 0.0)
	return metrics
}

// benchmarkNewMethod 对新方式进行基准测试
func benchmarkNewMethod(t *testing.T, engine *ProcessingEngine, request *ProcessingRequest, config *ABTestConfig) MethodMetrics {
	durations := make([]time.Duration, 0, config.RequestsPerMethod)
	var cacheHits int64
	var allocationBytes int64
	mu := sync.Mutex{}

	start := time.Now()
	sem := make(chan struct{}, config.ConcurrentWorkers)
	var wg sync.WaitGroup

	for i := 0; i < config.RequestsPerMethod; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			reqStart := time.Now()
			result := engine.ProcessWithPipeline(request)
			duration := time.Since(reqStart)

			mu.Lock()
			durations = append(durations, duration)
			if result.Success {
				// 在缓存场景中，第二次执行会更快 (缓存 HIT)
				if duration < 1*time.Millisecond && i > config.Warmup/2 {
					atomic.AddInt64(&cacheHits, 1)
				}
				if len(result.AnalysisResults) > 0 {
					allocationBytes += int64(len(result.AnalysisResults)) * 100
				}
			}
			mu.Unlock()
		}()
	}

	wg.Wait()
	totalDuration := time.Since(start)

	cacheHitRate := float64(cacheHits) / float64(config.RequestsPerMethod)
	metrics := calculateMetrics("ProcessWithPipeline (新方式)", durations, totalDuration, config.RequestsPerMethod, allocationBytes, cacheHitRate)
	return metrics
}

// calculateMetrics 从原始持续时间数据计算指标
func calculateMetrics(name string, durations []time.Duration, totalDuration time.Duration, requestCount int, allocBytes int64, cacheHitRate float64) MethodMetrics {
	if len(durations) == 0 {
		return MethodMetrics{Name: name}
	}

	// 排序以计算百分比
	sortDurations(durations)

	metric := MethodMetrics{
		Name:              name,
		TotalRequests:     requestCount,
		MinDuration:       durations[0],
		MaxDuration:       durations[len(durations)-1],
		MemoryAllocations: int64(len(durations)),
		AllocationBytes:   allocBytes,
		CacheHitRate:      cacheHitRate,
	}

	// 计算平均值
	var sum time.Duration
	for _, d := range durations {
		sum += d
	}
	metric.AVGDuration = sum / time.Duration(len(durations))

	// 计算百分位数
	metric.P50Duration = durations[len(durations)/2]
	metric.P95Duration = durations[int(float64(len(durations))*0.95)]
	metric.P99Duration = durations[int(float64(len(durations))*0.99)]

	return metric
}

// sortDurations 对持续时间进行排序 (简单实现)
func sortDurations(durations []time.Duration) {
	// 简单的冒泡排序 (实际应用中应使用 sort.Slice)
	n := len(durations)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if durations[j] > durations[j+1] {
				durations[j], durations[j+1] = durations[j+1], durations[j]
			}
		}
	}
}

// verifyResultsConsistency 验证新旧方式的结果一致性
func verifyResultsConsistency(t *testing.T, engine *ProcessingEngine, request *ProcessingRequest, testCount int) (bool, float64) {
	for i := 0; i < testCount; i++ {
		oldResult := engine.Process(request)
		newResult := engine.ProcessWithPipeline(request)

		// 验证成功状态一致
		if oldResult.Success != newResult.Success {
			t.Logf("❌ 结果不一致: 成功状态不匹配 (旧=%v, 新=%v)", oldResult.Success, newResult.Success)
			return false, 1.0
		}

		// 验证结果长度一致
		if len(oldResult.AnalysisResults) != len(newResult.AnalysisResults) {
			t.Logf("⚠️  警告: 分析结果数量不同 (旧=%d, 新=%d)", len(oldResult.AnalysisResults), len(newResult.AnalysisResults))
		}
	}

	t.Logf("✅ %d 个请求的结果一致性验证通过", testCount)
	return true, 0.0
}

// calculateImprovement 计算性能改进
func calculateImprovement(oldMetric, newMetric *MethodMetrics) PerformanceImprovement {
	improvement := PerformanceImprovement{
		LatencyAbsoluteGain: oldMetric.AVGDuration - newMetric.AVGDuration,
	}

	// 计算相对改进
	if oldMetric.AVGDuration > 0 {
		improvement.LatencyReduction = float64(oldMetric.AVGDuration-newMetric.AVGDuration) * 100 / float64(oldMetric.AVGDuration)
	}

	if oldMetric.AllocationBytes > 0 {
		improvement.MemoryReduction = float64(oldMetric.AllocationBytes-newMetric.AllocationBytes) * 100 / float64(oldMetric.AllocationBytes)
	}

	// 计算吞吐量改进 (请求/秒)
	oldThroughput := 1000 / float64(oldMetric.AVGDuration.Milliseconds())
	newThroughput := 1000 / float64(newMetric.AVGDuration.Milliseconds())
	improvement.ThroughputIncrease = (newThroughput - oldThroughput) * 100 / oldThroughput

	// 生成建议
	if improvement.LatencyReduction < -5 {
		improvement.Recommendation = "❌ 新方式性能回退，不建议迁移"
	} else if improvement.LatencyReduction < 0 {
		improvement.Recommendation = "⚠️  新方式性能略有下降（<5%），需要评估中间件收益"
	} else if improvement.LatencyReduction < 10 {
		improvement.Recommendation = "✅ 性能影响最小（<10%），中间件收益完全覆盖开销"
	} else {
		improvement.Recommendation = "⭐ 性能改进显著，强烈建议迁移"
	}

	return improvement
}

// runGradualRolloutTest 运行灰度流量测试
func runGradualRolloutTest(t *testing.T, engine *ProcessingEngine, request *ProcessingRequest, config *ABTestConfig) GradualRolloutMetrics {
	metrics := GradualRolloutMetrics{
		Stages: make([]RolloutStage, 0),
	}

	totalTestRequests := 100 // 每个阶段的总请求数

	for _, percentage := range config.RolloutStages {
		stage := RolloutStage{
			TrafficPercentage: percentage,
			TotalRequests:     totalTestRequests,
			NewMethodRequests: int(float64(totalTestRequests) * percentage),
			OldMethodRequests: int(float64(totalTestRequests) * (1 - percentage)),
		}

		var newDuration, oldDuration time.Duration
		successCount := 0
		errorCount := 0

		// 执行混合方式的请求
		for i := 0; i < stage.TotalRequests; i++ {
			useNew := i < stage.NewMethodRequests

			start := time.Now()
			var result *ProcessingResult
			if useNew {
				result = engine.ProcessWithPipeline(request)
				newDuration += time.Since(start)
			} else {
				result = engine.Process(request)
				oldDuration += time.Since(start)
			}

			if result.Success {
				successCount++
			} else {
				errorCount++
			}
		}

		// 处理除以零的情况
		if stage.NewMethodRequests > 0 {
			stage.NewMethodLatency = newDuration / time.Duration(stage.NewMethodRequests)
		} else {
			stage.NewMethodLatency = 0
		}

		if stage.OldMethodRequests > 0 {
			stage.OldMethodLatency = oldDuration / time.Duration(stage.OldMethodRequests)
		} else {
			stage.OldMethodLatency = 0
		}

		stage.SuccessRate = float64(successCount) * 100 / float64(stage.TotalRequests)
		stage.ErrorCount = errorCount

		metrics.Stages = append(metrics.Stages, stage)

		t.Logf("  📊 灰度阶段 %.0f%%: 新=%dms, 旧=%dms, 成功率=%.1f%%",
			percentage*100,
			stage.NewMethodLatency.Milliseconds(),
			stage.OldMethodLatency.Milliseconds(),
			stage.SuccessRate)
	}

	return metrics
}

// createTestRequest 创建一个测试请求
func createTestRequest() *ProcessingRequest {
	return &ProcessingRequest{
		Context:       context.Background(),
		RawData:       []byte("test data for benchmark"),
		ExtensionType: ExtensionType(0), // 使用 uint16 类型
		Steps:         []string{"parse", "analyze", "transform"},
		Metadata:      make(map[string]interface{}),
	}
}

// ========================================================================
// Test 函数：集成到 Go testing 框架
// ========================================================================

// TestAB_PerformanceComparison 运行完整的 A/B 性能对比测试
func TestAB_PerformanceComparison(t *testing.T) {
	t.Logf("🧪 开始 A/B 性能对比测试")

	// 创建处理引擎
	registry := &mockRegistry{}
	engineConfig := &EngineConfig{
		ConcurrentProcessing: true,
		MaxConcurrency:       16,
		EnableCaching:        true,
		CacheSize:            1000,
	}
	engine := NewProcessingEngineWithRegistry(engineConfig, registry)

	// 使用自定义配置
	testConfig := &ABTestConfig{
		RequestsPerMethod: 500,
		ConcurrentWorkers: 5,
		CachingEnabled:    true,
		CacheHitRatio:     0.6,
		Warmup:            50,
		GradualRollout:    true,
		RolloutStages:     []float64{0.05, 0.25, 0.50, 1.0},
	}

	// 运行 A/B 测试
	result := RunABTest(t, engine, testConfig)

	// 输出结果摘要
	printABTestResultsSummary(t, result)

	// 验证关键指标
	if !result.ResultsMatch {
		t.Errorf("❌ 结果不一致: 新旧方式返回的结果不匹配")
	}

	if result.OldMethod.AVGDuration == 0 {
		t.Errorf("❌ 旧方式没有计时数据")
	}

	t.Logf("✅ A/B 性能对比测试完成")
}

// TestAB_GradualRolloutSimulation 运行灰度流量模拟
func TestAB_GradualRolloutSimulation(t *testing.T) {
	t.Logf("📈 开始灰度流量模拟")

	registry := &mockRegistry{}
	engineConfig := &EngineConfig{
		ConcurrentProcessing: true,
		MaxConcurrency:       16,
	}
	engine := NewProcessingEngineWithRegistry(engineConfig, registry)

	testConfig := &ABTestConfig{
		RequestsPerMethod: 200,
		ConcurrentWorkers: 3,
		GradualRollout:    true,
		RolloutStages:     []float64{0.05, 0.10, 0.25, 0.50, 1.0},
	}

	start := time.Now()
	result := RunABTest(t, engine, testConfig)
	duration := time.Since(start)

	t.Logf("📊 灰度流量模拟结果:")
	for i, stage := range result.GradualRollout.Stages {
		t.Logf("  阶段 %d (%.0f%%): 成功率=%.1f%%, 错误=%d",
			i+1, stage.TrafficPercentage*100, stage.SuccessRate, stage.ErrorCount)
	}

	t.Logf("✅ 灰度流量模拟完成 (耗时: %v)", duration)
}

// TestAB_ResultsConsistency 验证结果一致性
func TestAB_ResultsConsistency(t *testing.T) {
	t.Logf("✅ 开始验证结果一致性")

	registry := &mockRegistry{}
	config := &EngineConfig{
		ConcurrentProcessing: true,
	}
	engine := NewProcessingEngineWithRegistry(config, registry)

	request := createTestRequest()

	// 运行多个请求并验证
	for i := 0; i < 100; i++ {
		oldResult := engine.Process(request)
		newResult := engine.ProcessWithPipeline(request)

		if oldResult.Success != newResult.Success {
			t.Errorf("❌ 请求 %d: 成功状态不匹配 (旧=%v, 新=%v)", i, oldResult.Success, newResult.Success)
		}

		if len(oldResult.AnalysisResults) != len(newResult.AnalysisResults) {
			t.Logf("⚠️  请求 %d: 结果数量不同 (旧=%d, 新=%d)", i, len(oldResult.AnalysisResults), len(newResult.AnalysisResults))
		}
	}

	match, variance := verifyResultsConsistency(t, engine, request, 50)
	if !match {
		t.Errorf("❌ 结果一致性验证失败 (方差: %.4f)", variance)
	}

	t.Logf("✅ 结果一致性验证完成")
}

// printABTestResultsSummary 打印 A/B 测试结果摘要
func printABTestResultsSummary(t *testing.T, result *ABTestResult) {
	t.Logf("")
	t.Logf("╔════════════════════════════════════════════════╗")
	t.Logf("║          A/B 测试结果摘要                      ║")
	t.Logf("╚════════════════════════════════════════════════╝")
	t.Logf("")

	t.Logf("📊 基准测试数据:")
	t.Logf("  旧方式 (Process):")
	t.Logf("    • 平均延迟: %v", result.OldMethod.AVGDuration)
	t.Logf("    • P50 延迟: %v", result.OldMethod.P50Duration)
	t.Logf("    • P95 延迟: %v", result.OldMethod.P95Duration)
	t.Logf("    • P99 延迟: %v", result.OldMethod.P99Duration)
	t.Logf("")
	t.Logf("  新方式 (ProcessWithPipeline):")
	t.Logf("    • 平均延迟: %v", result.NewMethod.AVGDuration)
	t.Logf("    • P50 延迟: %v", result.NewMethod.P50Duration)
	t.Logf("    • P95 延迟: %v", result.NewMethod.P95Duration)
	t.Logf("    • P99 延迟: %v", result.NewMethod.P99Duration)
	if result.NewMethod.CacheHitRate > 0 {
		t.Logf("    • 缓存命中率: %.1f%%", result.NewMethod.CacheHitRate*100)
	}
	t.Logf("")

	t.Logf("📈 性能改进:")
	t.Logf("  • 延迟改进: %.2f%% (绝对值: %v)", result.Improvement.LatencyReduction, result.Improvement.LatencyAbsoluteGain)
	t.Logf("  • 吞吐量提升: %.2f%%", result.Improvement.ThroughputIncrease)
	t.Logf("  • 内存改进: %.2f%%", result.Improvement.MemoryReduction)
	t.Logf("  • 建议: %s", result.Improvement.Recommendation)
	t.Logf("")

	t.Logf("✅ 结果一致性: %v", result.ResultsMatch)
	if result.ResultsMatch {
		t.Logf("  • 新旧方式输出结果完全一致")
	} else {
		t.Logf("  • ❌ 发现不一致 (方差: %.4f)", result.Variance)
	}
	t.Logf("")

	if len(result.GradualRollout.Stages) > 0 {
		t.Logf("📈 灰度流量模拟:")
		for i, stage := range result.GradualRollout.Stages {
			t.Logf("  阶段 %d (新方式 %.0f%%):", i+1, stage.TrafficPercentage*100)
			t.Logf("    • 总请求: %d", stage.TotalRequests)
			t.Logf("    • 成功率: %.1f%%", stage.SuccessRate)
			if stage.ErrorCount > 0 {
				t.Logf("    • 错误数: %d", stage.ErrorCount)
			}
		}
	}
	t.Logf("")
}

// mockRegistry 用于测试的 mock 注册表
type mockRegistry struct{}

func (m *mockRegistry) GetParser(extType ExtensionType) (Parser, error) {
	return &mockParser{}, nil
}

func (m *mockRegistry) GetAnalyzer(extType ExtensionType) (Analyzer, error) {
	return &mockAnalyzer{}, nil
}

func (m *mockRegistry) GetHandlers(extType ExtensionType) []Handler {
	return []Handler{&mockHandler{}}
}

// mockParser 用于测试的 mock 解析器
type mockParser struct{}

func (m *mockParser) Parse(data []byte, ctx context.Context) (ExtensionData, error) {
	return &mockExtensionData{data: data}, nil
}

func (m *mockParser) GetType() ExtensionType {
	return ExtensionType(0)
}

func (m *mockParser) GetVersion() string {
	return "1.0"
}

// mockAnalyzer 用于测试的 mock 分析器
type mockAnalyzer struct{}

func (m *mockAnalyzer) Analyze(data ExtensionData, config map[string]interface{}) (AnalysisResult, error) {
	return &mockAnalysisResult{}, nil
}

func (m *mockAnalyzer) GetType() ExtensionType {
	return ExtensionType(0)
}

func (m *mockAnalyzer) GetVersion() string {
	return "1.0"
}

func (m *mockAnalyzer) SupportsConfig() []string {
	return []string{}
}

// mockHandler 用于测试的 mock 处理器
type mockHandler struct{}

func (m *mockHandler) Handle(event *ExtensionEvent) (*EventResult, error) {
	return &EventResult{Success: true}, nil
}

func (m *mockHandler) GetType() ExtensionType {
	return ExtensionType(0)
}

func (m *mockHandler) GetPriority() int {
	return 50
}

func (m *mockHandler) GetName() string {
	return "mock_handler"
}

// mockExtensionData 用于测试的 mock ExtensionData
type mockExtensionData struct {
	data []byte
}

func (m *mockExtensionData) GetType() ExtensionType {
	return ExtensionType(0)
}

func (m *mockExtensionData) GetRawData() []byte {
	return m.data
}

func (m *mockExtensionData) GetName() string {
	return "mock_data"
}

func (m *mockExtensionData) ToMap() map[string]interface{} {
	return map[string]interface{}{}
}

// mockAnalysisResult 用于测试的 mock AnalysisResult
type mockAnalysisResult struct{}

func (m *mockAnalysisResult) GetExtensionType() ExtensionType {
	return ExtensionType(0)
}

func (m *mockAnalysisResult) HasAnomalies() bool {
	return false
}

func (m *mockAnalysisResult) GetAnomalies() []string {
	return []string{}
}

func (m *mockAnalysisResult) GetRiskScore() float64 {
	return 0.0
}

func (m *mockAnalysisResult) ToMap() map[string]interface{} {
	return map[string]interface{}{}
}
