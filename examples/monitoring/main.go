package main

import (
	"fmt"
	"time"

	"github.com/vistone/fingerprint/internal/metrics"
	"github.com/vistone/fingerprint/internal/monitor"
	"github.com/vistone/fingerprint/internal/tracing"
)

func main() {
	fmt.Println("=== 监控和可观测性系统演示 ===\n")

	demonstrateMetrics()

	fmt.Println("\n" + string(make([]byte, 50)) + "\n")

	demonstrateTracing()

	fmt.Println("\n" + string(make([]byte, 50)) + "\n")

	demonstrateHealthCheck()
}

// demonstrateMetrics 演示指标收集
func demonstrateMetrics() {
	fmt.Println("📊 演示 1: 指标收集")
	fmt.Println("-" + string(make([]byte, 49)))

	pm := metrics.NewPerformanceMetrics()

	fmt.Println("📈 模拟指纹识别请求处理:")

	// 模拟 50 个请求
	for i := 0; i < 50; i++ {
		duration := 10.0 + float64(i%30)
		isError := i%7 == 0 // 大约 15% 错误率
		pm.RecordRequest(duration, isError)
	}

	fmt.Printf("✓ QPS: %d\n", pm.GetQPS())
	fmt.Printf("✓ 错误数: %d\n", pm.GetErrors())
	fmt.Printf("✓ 平均延迟: %.2f ms\n", pm.GetMeanLatency())

	// 导出完整指标
	exported := pm.Export()
	fmt.Println("\n📋 完整指标导出:")
	for key, value := range exported {
		fmt.Printf("   %s: %v\n", key, value)
	}

	fmt.Println("\n✅ 指标收集演示完成!")
}

// demonstrateTracing 演示分布式追踪
func demonstrateTracing() {
	fmt.Println("🔗 演示 2: 分布式追踪")
	fmt.Println("-" + string(make([]byte, 49)))

	tracer := tracing.NewTracer()
	traceID := tracing.TraceID("trace-001")

	fmt.Println("🔄 创建追踪链:")

	operations := []string{"TCP/IP Analysis", "TLS Handshake", "Behavior Analysis", "Risk Scoring"}

	for _, op := range operations {
		span := tracer.StartSpan(traceID, op)
		span.AddTag("service", "analyzer")
		span.AddTag("version", "1.0")

		// 模拟处理
		time.Sleep(10 * time.Millisecond)

		span.End()

		fmt.Printf("   ✓ %s (%v)\n", op, span.Duration)
	}

	// 获取完整追踪
	fmt.Println("\n📊 追踪汇总:")
	spans := tracer.GetTrace(traceID)
	fmt.Printf("   总 Span 数: %d\n", len(spans))

	var totalDuration time.Duration
	for _, span := range spans {
		totalDuration += span.Duration
	}
	fmt.Printf("   总耗时: %v\n", totalDuration)

	fmt.Println("\n✅ 分布式追踪演示完成!")
}

// demonstrateHealthCheck 演示健康检查
func demonstrateHealthCheck() {
	fmt.Println("🏥 演示 3: 健康检查和监控")
	fmt.Println("-" + string(make([]byte, 49)))

	// 创建简单的健康检查器
	simpleChecker := &SimpleHealthChecker{
		name:   "fingerprint-service",
		status: monitor.Healthy,
	}

	// 创建监控器并注册检查器
	mon := monitor.NewMonitor()
	mon.RegisterChecker(simpleChecker)

	fmt.Println("🔍 执行初始健康检查:")
	results := mon.Check()
	for name, result := range results {
		statusIcon := "✓"
		if result.Status == monitor.Degraded {
			statusIcon = "⚠️"
		} else if result.Status == monitor.Unhealthy {
			statusIcon = "✗"
		}
		fmt.Printf("   %s %s: %s\n", statusIcon, name, result.Status)
	}

	// 模拟服务故障
	fmt.Println("\n⚠️  模拟服务故障:")
	simpleChecker.status = monitor.Unhealthy
	results = mon.Check()
	for name, result := range results {
		statusIcon := "✓"
		if result.Status == monitor.Degraded {
			statusIcon = "⚠️"
		} else if result.Status == monitor.Unhealthy {
			statusIcon = "✗"
		}
		fmt.Printf("   %s %s: %s\n", statusIcon, name, result.Status)
	}

	// 恢复服务
	fmt.Println("\n✅ 服务恢复:")
	simpleChecker.status = monitor.Healthy
	results = mon.Check()
	for name, result := range results {
		fmt.Printf("   ✓ %s: %s\n", name, result.Status)
	}

	fmt.Println("\n📊 总体健康状态: " + string(mon.GetStatus()))
	fmt.Println("\n✅ 健康检查演示完成!")
}

// SimpleHealthChecker 简单的健康检查器实现
type SimpleHealthChecker struct {
	name   string
	status monitor.HealthStatus
}

// Check 执行检查
func (sc *SimpleHealthChecker) Check() monitor.HealthCheckResult {
	return monitor.HealthCheckResult{
		Name:      sc.name,
		Status:    sc.status,
		Message:   "Service check completed",
		Timestamp: time.Now(),
	}
}

// Name 返回名称
func (sc *SimpleHealthChecker) Name() string {
	return sc.name
}
