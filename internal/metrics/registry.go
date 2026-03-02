package metrics

import (
	"fmt"
	"strings"
	"sync"
)

// Registry 指标注册表
type Registry struct {
	metrics map[string]Metric
	mu      sync.RWMutex
}

// NewRegistry 创建注册表
func NewRegistry() *Registry {
	return &Registry{
		metrics: make(map[string]Metric),
	}
}

// Register 注册指标
func (r *Registry) Register(metric Metric) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := metric.Name()
	if _, exists := r.metrics[name]; exists {
		return fmt.Errorf("metric %s already registered", name)
	}

	r.metrics[name] = metric
	return nil
}

// Unregister 注销指标
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.metrics, name)
}

// Get 获取指标
func (r *Registry) Get(name string) (Metric, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	metric, exists := r.metrics[name]
	return metric, exists
}

// GetAll 获取所有指标
func (r *Registry) GetAll() []Metric {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metrics := make([]Metric, 0, len(r.metrics))
	for _, m := range r.metrics {
		metrics = append(metrics, m)
	}
	return metrics
}

// PrometheusExporter Prometheus 导出器
type PrometheusExporter struct {
	registry *Registry
}

// NewPrometheusExporter 创建 Prometheus 导出器
func NewPrometheusExporter(registry *Registry) *PrometheusExporter {
	return &PrometheusExporter{
		registry: registry,
	}
}

// Export 导出 Prometheus 格式
func (e *PrometheusExporter) Export() string {
	var sb strings.Builder

	metrics := e.registry.GetAll()
	for _, metric := range metrics {
		// HELP 行
		sb.WriteString(fmt.Sprintf("# HELP %s %s\n", metric.Name(), metric.Help()))

		// TYPE 行
		sb.WriteString(fmt.Sprintf("# TYPE %s %s\n", metric.Name(), metric.Type()))

		// 值行
		switch m := metric.(type) {
		case *Counter:
			sb.WriteString(fmt.Sprintf("%s %d\n", m.Name(), m.Get()))
		case *Gauge:
			sb.WriteString(fmt.Sprintf("%s %.2f\n", m.Name(), m.Get()))
		case *Histogram:
			sb.WriteString(fmt.Sprintf("%s_mean %.2f\n", m.Name(), m.Mean()))
			sb.WriteString(fmt.Sprintf("%s_p50 %.2f\n", m.Name(), m.GetPercentile(50)))
			sb.WriteString(fmt.Sprintf("%s_p95 %.2f\n", m.Name(), m.GetPercentile(95)))
			sb.WriteString(fmt.Sprintf("%s_p99 %.2f\n", m.Name(), m.GetPercentile(99)))
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// 全局注册表
var globalRegistry = NewRegistry()

// GetGlobalRegistry 获取全局注册表
func GetGlobalRegistry() *Registry {
	return globalRegistry
}

// RegisterMetric 注册指标到全局注册表
func RegisterMetric(metric Metric) error {
	return globalRegistry.Register(metric)
}

// GetMetric 从全局注册表获取指标
func GetMetric(name string) (Metric, bool) {
	return globalRegistry.Get(name)
}

// UnregisterMetric 从全局注册表注销指标
func UnregisterMetric(name string) {
	globalRegistry.Unregister(name)
}
