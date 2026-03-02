package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/vistone/fingerprint/internal/metrics"
	"github.com/vistone/fingerprint/internal/monitor"
	"github.com/vistone/fingerprint/internal/tracing"
)

// FingerprintService 模拟指纹识别服务
type FingerprintService struct {
	perfMetrics *metrics.PerformanceMetrics
	tracer      *tracing.Tracer
	monitor     *monitor.Monitor
}

// NewFingerprintService 创建服务实例
func NewFingerprintService() *FingerprintService {
	svc := &FingerprintService{
		perfMetrics: metrics.NewPerformanceMetrics(),
		tracer:      tracing.NewTracer(),
		monitor:     monitor.NewMonitor(),
	}

	// 注册健康检查
	svc.monitor.RegisterChecker(&ServiceHealthChecker{
		service: svc,
	})

	// 注册指标到全局注册表
	registry := metrics.GetGlobalRegistry()
	registry.Register(metrics.NewCounter("fingerprint_requests", "Total fingerprint requests"))
	registry.Register(metrics.NewGauge("active_sessions", "Active analysis sessions"))

	return svc
}

// AnalyzeFingerprint 分析指纹
func (fs *FingerprintService) AnalyzeFingerprint(data map[string]interface{}) (result map[string]interface{}, err error) {
	// 生成 TraceID
	traceID := tracing.GenerateTraceID()

	// 记录开始时间
	start := time.Now()

	// 启动整体追踪
	mainSpan := fs.tracer.StartSpan(traceID, "AnalyzeFingerprint")
	mainSpan.AddTag("service", "fingerprint-analysis")
	defer mainSpan.End()

	// 阶段 1: TCP/IP 分析
	tcpSpan := fs.tracer.StartSpan(traceID, "TCP/IP-Analysis")
	tcpSpan.AddTag("component", "tcp")
	time.Sleep(time.Duration(5+rand.Intn(15)) * time.Millisecond) // 模拟处理
	tcpResult := map[string]interface{}{
		"ttl":       64,
		"mss":       1460,
		"window":    65535,
		"os_family": "Linux",
	}
	tcpSpan.End()

	// 阶段 2: TLS 分析
	tlsSpan := fs.tracer.StartSpan(traceID, "TLS-Analysis")
	tlsSpan.AddTag("component", "tls")
	time.Sleep(time.Duration(8+rand.Intn(12)) * time.Millisecond) // 模拟处理
	tlsResult := map[string]interface{}{
		"version":        "TLS1.3",
		"cipher_suites":  []string{"TLS_AES_256_GCM_SHA384"},
		"extensions":     []string{"server_name", "supported_groups"},
		"browser_family": "Chrome",
	}
	tlsSpan.End()

	// 阶段 3: 行为分析
	behaviorSpan := fs.tracer.StartSpan(traceID, "Behavior-Analysis")
	behaviorSpan.AddTag("component", "behavior")
	time.Sleep(time.Duration(10+rand.Intn(20)) * time.Millisecond) // 模拟处理
	behaviorResult := map[string]interface{}{
		"user_agent": data["user_agent"],
		"is_bot":     rand.Float64() < 0.1, // 10% 概率是机器人
		"risk_score": rand.Float64() * 100,
	}
	behaviorSpan.End()

	// 阶段 4: 风险评估
	riskSpan := fs.tracer.StartSpan(traceID, "Risk-Assessment")
	riskSpan.AddTag("component", "risk")
	time.Sleep(time.Duration(3+rand.Intn(7)) * time.Millisecond) // 模拟处理
	riskLevel := "low"
	if behaviorResult["is_bot"].(bool) {
		riskLevel = "high"
	}
	riskSpan.End()

	// 计算总耗时
	duration := time.Since(start).Seconds() * 1000 // 毫秒

	// 记录指标
	isError := riskLevel == "high" && rand.Float64() < 0.3 // 高风险有30%概率认为是错误
	fs.perfMetrics.RecordRequest(duration, isError)

	// 更新资源指标
	fs.perfMetrics.SetActiveConnections(rand.Intn(100) + 50)
	fs.perfMetrics.SetMemoryUsage(int64(rand.Intn(100000000) + 50000000))
	fs.perfMetrics.SetCPUUsage(float64(rand.Intn(60) + 10))

	// 组装结果
	result = map[string]interface{}{
		"trace_id":   string(traceID),
		"tcp_info":   tcpResult,
		"tls_info":   tlsResult,
		"behavior":   behaviorResult,
		"risk_level": riskLevel,
		"duration":   duration,
	}

	if isError {
		return nil, fmt.Errorf("high risk detected: %s", traceID)
	}

	return result, nil
}

// GetMetrics 获取性能指标
func (fs *FingerprintService) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"qps":                fs.perfMetrics.GetQPS(),
		"errors":             fs.perfMetrics.GetErrors(),
		"error_rate":         fs.perfMetrics.GetErrorRate(),
		"mean_latency":       fs.perfMetrics.GetMeanLatency(),
		"p50_latency":        fs.perfMetrics.GetP50Latency(),
		"p95_latency":        fs.perfMetrics.GetP95Latency(),
		"p99_latency":        fs.perfMetrics.GetP99Latency(),
		"active_connections": fs.perfMetrics.GetActiveConnections(),
		"memory_usage":       fs.perfMetrics.GetMemoryUsage(),
		"cpu_usage":          fs.perfMetrics.GetCPUUsage(),
	}
}

// GetHealth 获取健康状态
func (fs *FingerprintService) GetHealth() map[string]interface{} {
	results := fs.monitor.Check()
	overallStatus := fs.monitor.GetStatus()

	checks := make(map[string]string)
	for name, result := range results {
		checks[name] = string(result.Status)
	}

	return map[string]interface{}{
		"status": string(overallStatus),
		"checks": checks,
	}
}

// GetTrace 获取追踪信息
func (fs *FingerprintService) GetTrace(traceID string) []map[string]interface{} {
	spans := fs.tracer.GetTrace(tracing.TraceID(traceID))

	result := make([]map[string]interface{}, len(spans))
	for i, span := range spans {
		result[i] = map[string]interface{}{
			"name":     span.Name,
			"duration": span.Duration.String(),
			"tags":     span.Tags,
		}
	}

	return result
}

// ExportPrometheusMetrics 导出 Prometheus 格式指标
func (fs *FingerprintService) ExportPrometheusMetrics() string {
	registry := metrics.GetGlobalRegistry()
	exporter := metrics.NewPrometheusExporter(registry)
	return exporter.Export()
}

// ServiceHealthChecker 服务健康检查器
type ServiceHealthChecker struct {
	service *FingerprintService
}

func (shc *ServiceHealthChecker) Name() string {
	return "fingerprint-service"
}

func (shc *ServiceHealthChecker) Check() monitor.HealthCheckResult {
	// 检查错误率
	errorRate := shc.service.perfMetrics.GetErrorRate()
	cpu := shc.service.perfMetrics.GetCPUUsage()

	status := monitor.Healthy
	message := "Service is healthy"

	if errorRate > 5 {
		status = monitor.Degraded
		message = fmt.Sprintf("High error rate: %.2f%%", errorRate)
	} else if cpu > 90 {
		status = monitor.Degraded
		message = fmt.Sprintf("High CPU usage: %.2f%%", cpu)
	}

	if errorRate > 20 {
		status = monitor.Unhealthy
		message = fmt.Sprintf("Critical error rate: %.2f%%", errorRate)
	}

	return monitor.HealthCheckResult{
		Name:      shc.Name(),
		Status:    status,
		Message:   message,
		Timestamp: time.Now(),
	}
}

// ========== 主程序 ==========

func main() {
	fmt.Println("=== 指纹识别服务 - 监控集成演示 ===\n")

	// 创建服务
	service := NewFingerprintService()

	// 模拟处理多个请求
	fmt.Println("🔄 处理指纹识别请求...\n")

	successCount := 0
	errorCount := 0

	for i := 0; i < 30; i++ {
		requestData := map[string]interface{}{
			"user_agent": fmt.Sprintf("Mozilla/5.0 (Request %d)", i),
			"ip":         fmt.Sprintf("192.168.1.%d", i+1),
		}

		result, err := service.AnalyzeFingerprint(requestData)
		if err != nil {
			errorCount++
			fmt.Printf("   [✗] 请求 %d 失败: %v\n", i+1, err)
		} else {
			successCount++
			if i < 3 { // 只显示前几个成功的请求
				fmt.Printf("   [✓] 请求 %d 成功 - TraceID: %s, 耗时: %.2fms, 风险: %s\n",
					i+1, result["trace_id"], result["duration"], result["risk_level"])
			}
		}

		// 模拟请求间隔
		time.Sleep(50 * time.Millisecond)
	}

	fmt.Printf("\n✅ 处理完成: %d 成功, %d 失败\n\n", successCount, errorCount)

	// 显示性能指标
	fmt.Println("📊 性能指标:")
	fmt.Println(strings.Repeat("-", 60))
	metricsData := service.GetMetrics()
	fmt.Printf("   QPS:             %d 请求/秒\n", metricsData["qps"])
	fmt.Printf("   错误数:          %d\n", metricsData["errors"])
	fmt.Printf("   错误率:          %.2f%%\n", metricsData["error_rate"])
	fmt.Printf("   平均延迟:        %.2f ms\n", metricsData["mean_latency"])
	fmt.Printf("   P50 延迟:        %.2f ms\n", metricsData["p50_latency"])
	fmt.Printf("   P95 延迟:        %.2f ms\n", metricsData["p95_latency"])
	fmt.Printf("   P99 延迟:        %.2f ms\n", metricsData["p99_latency"])
	fmt.Printf("   活跃连接:        %.0f\n", metricsData["active_connections"])
	fmt.Printf("   内存使用:        %.2f MB\n", metricsData["memory_usage"].(float64)/1024/1024)
	fmt.Printf("   CPU 使用率:      %.2f%%\n\n", metricsData["cpu_usage"])

	// 显示健康状态
	fmt.Println("🏥 健康检查:")
	fmt.Println(strings.Repeat("-", 60))
	healthData := service.GetHealth()
	fmt.Printf("   总体状态:        %s\n", healthData["status"])
	checks := healthData["checks"].(map[string]string)
	for name, status := range checks {
		icon := "✓"
		if status == "DEGRADED" {
			icon = "⚠️"
		} else if status == "UNHEALTHY" {
			icon = "✗"
		}
		fmt.Printf("   %s %s: %s\n", icon, name, status)
	}

	// 导出 Prometheus 指标
	fmt.Println("\n📈 Prometheus 指标导出:\n")
	fmt.Println(strings.Repeat("-", 60))
	prometheusData := service.ExportPrometheusMetrics()
	fmt.Println(prometheusData)

	fmt.Println("✅ 监控集成演示完成!")
}
