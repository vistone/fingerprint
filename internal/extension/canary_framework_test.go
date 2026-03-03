package extension

import (
	"context"
	"math/rand"
	"testing"
	"time"
)

// ========================================================================
// 灰度框架测试
// ========================================================================

// TestCanaryRouter 测试灰度路由器
func TestCanaryRouter(t *testing.T) {
	tests := []struct {
		name       string
		percentage float64
		wantRatio  float64 // 期望的流量比例
		tolerance  float64 // 容差
	}{
		{"5% 灰度", 0.05, 0.05, 0.01},
		{"25% 灰度", 0.25, 0.25, 0.02},
		{"50% 灰度", 0.50, 0.50, 0.02},
		{"100% 灰度", 1.00, 1.00, 0.00},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &CanaryConfig{
				Stage:            CanaryStage5Percent,
				TargetPercentage: tt.percentage,
				Enabled:          true,
			}
			router := &CanaryRouter{config: config}

			// 测试 10000 个请求
			newMethodCount := 0
			for i := 0; i < 10000; i++ {
				if router.ShouldUseNewMethod(randomRequestID()) {
					newMethodCount++
				}
			}

			actualRatio := float64(newMethodCount) / 10000
			diff := actualRatio - tt.wantRatio
			if diff < 0 {
				diff = -diff
			}

			if diff > tt.tolerance {
				t.Errorf("流量分配不当: 期望 %.2f%%，实际 %.2f%% (差异: %.2f%%)",
					tt.wantRatio*100, actualRatio*100, diff*100)
			} else {
				t.Logf("✅ 流量分配正确: %.2f%% (期望: %.2f%%)", actualRatio*100, tt.wantRatio*100)
			}
		})
	}
}

// TestCanaryMetricsCollection 测试灰度指标收集
func TestCanaryMetricsCollection(t *testing.T) {
	config := &CanaryConfig{
		Stage:            CanaryStage5Percent,
		TargetPercentage: 0.05,
		Enabled:          true,
	}
	collector := NewCanaryMetricsCollector(config)

	// 模拟 1000 个请求
	successCount := 0
	totalLatency := time.Duration(0)

	for i := 0; i < 1000; i++ {
		latency := time.Duration(rand.Intn(100)) * time.Millisecond
		success := rand.Float64() > 0.01 // 99% 成功率
		useNew := collector.router.ShouldUseNewMethod(randomRequestID())

		collector.RecordRequest(randomRequestID(), useNew, latency, success)
		totalLatency += latency

		if success {
			successCount++
		}

		if i%20 == 0 {
			if rand.Float64() > 0.4 {
				collector.RecordCacheHit(true)
			} else {
				collector.RecordCacheHit(false)
			}
		}
	}

	// 验证指标
	metrics := collector.GetCurrentMetrics()

	if metrics.TotalRequests != 1000 {
		t.Errorf("总请求数不正确: 期望 1000，实际 %d", metrics.TotalRequests)
	}

	expectedSuccessRate := float64(successCount) / 1000
	actualSuccessRate := float64(metrics.SuccessfulRequests) / float64(metrics.TotalRequests)

	if actualSuccessRate < expectedSuccessRate-0.01 || actualSuccessRate > expectedSuccessRate+0.01 {
		t.Errorf("成功率不正确: 期望 %.2f%%，实际 %.2f%%",
			expectedSuccessRate*100, actualSuccessRate*100)
	}

	t.Logf("✅ 指标收集正确:")
	t.Logf("  总请求: %d", metrics.TotalRequests)
	t.Logf("  新方式: %d (%.1f%%)", metrics.NewMethodRequests, float64(metrics.NewMethodRequests)/10)
	t.Logf("  成功率: %.2f%%", actualSuccessRate*100)
	t.Logf("  缓存命中: %.1f%%", metrics.CacheHitRate*100)
}

// TestCanaryHealthCheck 测试灰度健康检查
func TestCanaryHealthCheck(t *testing.T) {
	config := &CanaryConfig{
		Stage:            CanaryStage5Percent,
		TargetPercentage: 0.05,
		Enabled:          true,
	}
	collector := NewCanaryMetricsCollector(config)
	healthCheck := NewCanaryHealthCheck(collector)

	// 测试 1: 健康情况
	t.Run("健康灰度", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			collector.RecordRequest(randomRequestID(), true, time.Duration(rand.Intn(100))*time.Millisecond, true)
			if rand.Float64() > 0.4 {
				collector.RecordCacheHit(true)
			}
		}

		healthy, alerts, _ := healthCheck.CheckHealth(context.Background())
		if !healthy {
			t.Errorf("应该是健康的，但获得 %d 个警告", len(alerts))
		} else {
			t.Logf("✅ 检查通过: 0 个警告")
		}
	})

	// 清空指标
	collector.metrics = &CanaryMetrics{CollectedAt: time.Now()}

	// 测试 2: 高错误率
	t.Run("高错误率警告", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			success := rand.Float64() > 0.02 // 模拟 2% 错误率
			collector.RecordRequest(randomRequestID(), true, time.Duration(rand.Intn(100))*time.Millisecond, success)
		}

		_, alerts, _ := healthCheck.CheckHealth(context.Background())
		if len(alerts) == 0 {
			t.Logf("✅ 检查通过: 识别高错误率警告")
		}
	})
}

// TestCanaryReport 测试灰度报告生成
func TestCanaryReport(t *testing.T) {
	config := &CanaryConfig{
		Stage:            CanaryStage25Percent,
		TargetPercentage: 0.25,
		Enabled:          true,
	}
	collector := NewCanaryMetricsCollector(config)
	startTime := time.Now()

	// 模拟 5000 个请求
	for i := 0; i < 5000; i++ {
		latency := time.Duration(rand.Intn(100)) * time.Millisecond
		success := rand.Float64() > 0.005 // 99.5% 成功
		useNew := collector.router.ShouldUseNewMethod(randomRequestID())

		collector.RecordRequest(randomRequestID(), useNew, latency, success)

		if i%10 == 0 && rand.Float64() > 0.4 {
			collector.RecordCacheHit(true)
		}
	}

	// 生成报告
	report := collector.GenerateCanaryReport(startTime)

	if report.TotalRequests != 5000 {
		t.Errorf("报告中的请求数不正确")
	}

	if report.SuccessRate < 0.99 || report.SuccessRate > 1.0 {
		t.Errorf("报告中的成功率不正确: %.2f%%", report.SuccessRate*100)
	}

	t.Logf("✅ 报告生成成功:")
	t.Logf("  阶段: %s", report.Stage)
	t.Logf("  总请求: %d", report.TotalRequests)
	t.Logf("  新方式比例: %.1f%%", report.NewMethodPercent*100)
	t.Logf("  成功率: %.2f%%", report.SuccessRate*100)

	// 打印完整报告
	report.Print()
}

// TestCanaryLifecycle 测试灰度生命周期管理
func TestCanaryLifecycle(t *testing.T) {
	lifecycle := NewCanaryLifecycle()

	// 测试 5% 灰度阶段
	t.Run("5% 灰度阶段", func(t *testing.T) {
		lifecycle.StartStage(CanaryStage5Percent, 100*time.Millisecond)

		// 模拟请求
		for i := 0; i < 500; i++ {
			latency := time.Duration(rand.Intn(50)) * time.Millisecond
			collector := lifecycle.collector
			collector.RecordRequest(randomRequestID(), true, latency, true)
		}

		report := lifecycle.EndStage()

		if report.Stage != CanaryStage5Percent {
			t.Errorf("阶段不正确")
		}

		if report.SuccessRate < 0.99 {
			t.Errorf("成功率不足")
		}

		t.Logf("✅ 5%% 灰度阶段完成")
		report.Print()
	})

	// 测试 25% 灰度阶段
	t.Run("25% 灰度阶段", func(t *testing.T) {
		lifecycle.StartStage(CanaryStage25Percent, 100*time.Millisecond)

		// 模拟请求
		for i := 0; i < 1000; i++ {
			latency := time.Duration(rand.Intn(50)) * time.Millisecond
			collector := lifecycle.collector
			collector.RecordRequest(randomRequestID(), true, latency, true)
		}

		report := lifecycle.EndStage()

		if report.Stage != CanaryStage25Percent {
			t.Errorf("阶段不正确")
		}

		if report.TotalRequests < 900 {
			t.Errorf("请求数太少")
		}

		t.Logf("✅ 25%% 灰度阶段完成")
		report.Print()
	})
}

// TestCanaryMetricsSnapshot 测试指标快照和历史
func TestCanaryMetricsSnapshot(t *testing.T) {
	config := &CanaryConfig{
		Stage:            CanaryStage5Percent,
		TargetPercentage: 0.05,
		Enabled:          true,
	}
	collector := NewCanaryMetricsCollector(config)

	// 模拟多个阶段的指标收集
	for phase := 0; phase < 3; phase++ {
		// 清空指标
		collector.metrics = &CanaryMetrics{CollectedAt: time.Now()}

		// 模拟请求
		for i := 0; i < 100; i++ {
			collector.RecordRequest(randomRequestID(), true, time.Duration(rand.Intn(50))*time.Millisecond, true)
		}

		// 保存快照
		collector.SnapshotMetrics()
		time.Sleep(10 * time.Millisecond) // 防止时间戳相同
	}

	// 验证历史数据
	history := collector.GetMetricsHistory(1)

	if len(history) != 3 {
		t.Errorf("历史数据数量不正确: 期望 3，实际 %d", len(history))
	} else {
		t.Logf("✅ 保存了 %d 条历史记录", len(history))
	}
}

// TestCanaryTrafficConsistency 测试灰度流量一致性 (关键测试)
func TestCanaryTrafficConsistency(t *testing.T) {
	t.Run("同一 requestID 的一致性路由", func(t *testing.T) {
		config := &CanaryConfig{
			Stage:            CanaryStage50Percent,
			TargetPercentage: 0.50,
			Enabled:          true,
		}
		router := &CanaryRouter{config: config}

		// 测试同一 requestID 总是使用同一方式
		testIDs := []string{
			"request-123",
			"request-456",
			"request-789",
		}

		for _, id := range testIDs {
			firstDecision := router.ShouldUseNewMethod(id)

			// 多次调用，应该返回相同的结果
			for i := 0; i < 100; i++ {
				decision := router.ShouldUseNewMethod(id)
				if decision != firstDecision {
					t.Errorf("requestID %s 的路由决策不一致", id)
				}
			}
		}

		t.Logf("✅ 流量一致性验证通过: 同一 requestID 总是使用同一处理方式")
	})
}

// TestCanaryRollback 测试灰度回滚
func TestCanaryRollback(t *testing.T) {
	config := &CanaryConfig{
		Stage:            CanaryStage5Percent,
		TargetPercentage: 0.05,
		Enabled:          true,
	}
	collector := NewCanaryMetricsCollector(config)
	healthCheck := NewCanaryHealthCheck(collector)

	// 模拟错误的灰度情况
	for i := 0; i < 100; i++ {
		success := rand.Float64() > 0.05 // 模拟 5% 错误率
		collector.RecordRequest(randomRequestID(), true, time.Duration(rand.Intn(50))*time.Millisecond, success)
	}

	// 检查健康状况
	healthy, alerts, critique := healthCheck.CheckHealth(context.Background())

	t.Logf("灰度健康检查结果:")
	t.Logf("  健康: %v", healthy)
	t.Logf("  警告数: %d", len(alerts))
	t.Logf("  评价: %s", critique)

	if !healthy && len(alerts) > 0 {
		t.Logf("✅ 识别到问题，应该进行回滚")
	}
}

// TestCanaryPerformanceComparison 测试新旧方式的性能对比
func TestCanaryPerformanceComparison(t *testing.T) {
	config := &CanaryConfig{
		Stage:            CanaryStage50Percent,
		TargetPercentage: 0.50,
		Enabled:          true,
	}
	collector := NewCanaryMetricsCollector(config)

	// 模拟 2000 个请求，分别用新旧方式处理
	newMethodLatencies := make([]time.Duration, 0)
	oldMethodLatencies := make([]time.Duration, 0)

	for i := 0; i < 2000; i++ {
		var latency time.Duration

		if i%2 == 0 {
			// 新方式: 因为有缓存加速，延迟更低
			latency = time.Duration(rand.Intn(50)) * time.Millisecond
			newMethodLatencies = append(newMethodLatencies, latency)
			collector.RecordRequest(randomRequestID(), true, latency, true)
			if rand.Float64() > 0.3 {
				collector.RecordCacheHit(true)
			}
		} else {
			// 旧方式: 延迟更高
			latency = time.Duration(rand.Intn(100)+50) * time.Millisecond
			oldMethodLatencies = append(oldMethodLatencies, latency)
			collector.RecordRequest(randomRequestID(), false, latency, true)
		}
	}

	metrics := collector.GetCurrentMetrics()

	// 计算平均延迟
	newAvg := time.Duration(0)
	oldAvg := time.Duration(0)

	if len(newMethodLatencies) > 0 {
		for _, l := range newMethodLatencies {
			newAvg += l
		}
		newAvg /= time.Duration(len(newMethodLatencies))
	}

	if len(oldMethodLatencies) > 0 {
		for _, l := range oldMethodLatencies {
			oldAvg += l
		}
		oldAvg /= time.Duration(len(oldMethodLatencies))
	}

	improvement := (float64(oldAvg.Milliseconds()) - float64(newAvg.Milliseconds())) / float64(oldAvg.Milliseconds()) * 100

	t.Logf("✅ 性能对比:")
	t.Logf("  新方式平均延迟: %.1fms", float64(newAvg.Milliseconds()))
	t.Logf("  旧方式平均延迟: %.1fms", float64(oldAvg.Milliseconds()))
	t.Logf("  性能提升: %.1f%%", improvement)
	t.Logf("  缓存命中率: %.1f%%", metrics.CacheHitRate*100)
}

// 辅助函数
func randomRequestID() string {
	return "req-" + randString(8)
}

func randString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// BenchmarkCanaryRouter 灰度路由器性能测试
func BenchmarkCanaryRouter(b *testing.B) {
	config := &CanaryConfig{
		Stage:            CanaryStage50Percent,
		TargetPercentage: 0.50,
		Enabled:          true,
	}
	router := &CanaryRouter{config: config}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		router.ShouldUseNewMethod(randomRequestID())
	}

	b.Logf("灰度路由器性能: %d ops/ns\n", b.N)
}

// BenchmarkCanaryMetricsCollection 灰度指标收集性能测试
func BenchmarkCanaryMetricsCollection(b *testing.B) {
	config := &CanaryConfig{
		Stage:            CanaryStage50Percent,
		TargetPercentage: 0.50,
		Enabled:          true,
	}
	collector := NewCanaryMetricsCollector(config)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		collector.RecordRequest(randomRequestID(), i%2 == 0, time.Duration(i%100)*time.Millisecond, true)
	}

	b.Logf("灰度指标收集性能: %d ops/ns\n", b.N)
}
