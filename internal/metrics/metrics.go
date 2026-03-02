package metrics

import (
	"sync"
)

// Metric 指标基础接口
type Metric interface {
	Name() string
	Type() string
	Help() string
}

// Counter 计数器指标
type Counter struct {
	name  string
	help  string
	value int64
	mu    sync.RWMutex
}

// NewCounter 创建计数器
func NewCounter(name, help string) *Counter {
	return &Counter{
		name: name,
		help: help,
	}
}

// Name 返回指标名
func (c *Counter) Name() string {
	return c.name
}

// Type 返回指标类型
func (c *Counter) Type() string {
	return "counter"
}

// Help 返回帮助信息
func (c *Counter) Help() string {
	return c.help
}

// Inc 增加计数器值
func (c *Counter) Inc(delta int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value += delta
}

// Get 获取计数器值
func (c *Counter) Get() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.value
}

// Gauge 仪表指标
type Gauge struct {
	name  string
	help  string
	value float64
	mu    sync.RWMutex
}

// NewGauge 创建仪表指标
func NewGauge(name, help string) *Gauge {
	return &Gauge{
		name: name,
		help: help,
	}
}

// Name 返回指标名
func (g *Gauge) Name() string {
	return g.name
}

// Type 返回指标类型
func (g *Gauge) Type() string {
	return "gauge"
}

// Help 返回帮助信息
func (g *Gauge) Help() string {
	return g.help
}

// Set 设置仪表值
func (g *Gauge) Set(value float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value = value
}

// Get 获取仪表值
func (g *Gauge) Get() float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.value
}

// Histogram 直方图指标
type Histogram struct {
	name    string
	help    string
	buckets []float64
	mu      sync.RWMutex
}

// NewHistogram 创建直方图指标
func NewHistogram(name, help string, buckets []float64) *Histogram {
	if buckets == nil {
		buckets = []float64{0.1, 0.5, 1, 2, 5, 10, 50, 100}
	}
	return &Histogram{
		name:    name,
		help:    help,
		buckets: buckets,
	}
}

// Name 返回指标名
func (h *Histogram) Name() string {
	return h.name
}

// Type 返回指标类型
func (h *Histogram) Type() string {
	return "histogram"
}

// Help 返回帮助信息
func (h *Histogram) Help() string {
	return h.help
}

// Observe 观察值
func (h *Histogram) Observe(value float64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.buckets = append(h.buckets, value)
}

// Mean 获取平均值
func (h *Histogram) Mean() float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.buckets) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range h.buckets {
		sum += v
	}
	return sum / float64(len(h.buckets))
}

// GetPercentile 获取百分位数
func (h *Histogram) GetPercentile(p float64) float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.buckets) == 0 {
		return 0
	}

	// 排序查找百分位数
	index := int(float64(len(h.buckets)) * p / 100)
	if index >= len(h.buckets) {
		index = len(h.buckets) - 1
	}

	return h.buckets[index]
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	qps             *Counter
	errors          *Counter
	requestDuration *Histogram
	connections     *Gauge
	memory          *Gauge
	cpu             *Gauge
	mu              sync.RWMutex
}

// NewPerformanceMetrics 创建性能指标
func NewPerformanceMetrics() *PerformanceMetrics {
	return &PerformanceMetrics{
		qps:             NewCounter("qps", "Requests per second"),
		errors:          NewCounter("errors", "Total errors"),
		requestDuration: NewHistogram("request_duration", "Request duration", nil),
		connections:     NewGauge("connections", "Active connections"),
		memory:          NewGauge("memory", "Memory usage"),
		cpu:             NewGauge("cpu", "CPU usage"),
	}
}

// RecordRequest 记录请求
func (pm *PerformanceMetrics) RecordRequest(duration float64, isError bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.qps.Inc(1)
	pm.requestDuration.Observe(duration)
	if isError {
		pm.errors.Inc(1)
	}
}

// GetQPS 获取 QPS
func (pm *PerformanceMetrics) GetQPS() int64 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.qps.Get()
}

// GetErrors 获取错误数
func (pm *PerformanceMetrics) GetErrors() int64 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.errors.Get()
}

// GetMeanLatency 获取平均延迟
func (pm *PerformanceMetrics) GetMeanLatency() float64 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.requestDuration.Mean()
}

// GetP50Latency 获取 P50 延迟
func (pm *PerformanceMetrics) GetP50Latency() float64 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.requestDuration.GetPercentile(50)
}

// GetP95Latency 获取 P95 延迟
func (pm *PerformanceMetrics) GetP95Latency() float64 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.requestDuration.GetPercentile(95)
}

// GetP99Latency 获取 P99 延迟
func (pm *PerformanceMetrics) GetP99Latency() float64 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.requestDuration.GetPercentile(99)
}

// Export 导出指标
func (pm *PerformanceMetrics) Export() map[string]interface{} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return map[string]interface{}{
		"qps":      pm.qps.Get(),
		"errors":   pm.errors.Get(),
		"duration": pm.requestDuration.Mean(),
	}
}

// RecordCacheHit 记录缓存命中
func (pm *PerformanceMetrics) RecordCacheHit() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.qps.Inc(1)
}

// RecordPacketsProcessed 记录处理的数据包
func (pm *PerformanceMetrics) RecordPacketsProcessed(count int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.qps.Inc(int64(count))
}

// SetActiveConnections 设置活跃连接数
func (pm *PerformanceMetrics) SetActiveConnections(count int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.connections.Set(float64(count))
}

// SetMemoryUsage 设置内存使用量
func (pm *PerformanceMetrics) SetMemoryUsage(bytes int64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.memory.Set(float64(bytes))
}

// SetCPUUsage 设置 CPU 使用率
func (pm *PerformanceMetrics) SetCPUUsage(percentage float64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.cpu.Set(percentage)
}

// GetErrorRate 获取错误率
func (pm *PerformanceMetrics) GetErrorRate() float64 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	qps := pm.qps.Get()
	if qps == 0 {
		return 0
	}
	return float64(pm.errors.Get()) / float64(qps) * 100
}

// GetActiveConnections 获取活跃连接数
func (pm *PerformanceMetrics) GetActiveConnections() float64 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.connections.Get()
}

// GetMemoryUsage 获取内存使用量
func (pm *PerformanceMetrics) GetMemoryUsage() float64 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.memory.Get()
}

// GetCPUUsage 获取 CPU 使用率
func (pm *PerformanceMetrics) GetCPUUsage() float64 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.cpu.Get()
}
