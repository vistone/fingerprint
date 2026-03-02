package monitor

import (
	"sync"
	"time"
)

// HealthStatus 健康状态
type HealthStatus string

const (
	Healthy   HealthStatus = "HEALTHY"
	Degraded  HealthStatus = "DEGRADED"
	Unhealthy HealthStatus = "UNHEALTHY"
)

// HealthCheckResult 检查结果
type HealthCheckResult struct {
	Name      string
	Status    HealthStatus
	Message   string
	Timestamp time.Time
}

// HealthChecker 健康检查器接口
type HealthChecker interface {
	Check() HealthCheckResult
	Name() string
}

// Monitor 监控器
type Monitor struct {
	checkers map[string]HealthChecker
	results  map[string]HealthCheckResult
	mu       sync.RWMutex
}

// NewMonitor 创建监控器
func NewMonitor() *Monitor {
	return &Monitor{
		checkers: make(map[string]HealthChecker),
		results:  make(map[string]HealthCheckResult),
	}
}

// RegisterChecker 注册检查器
func (m *Monitor) RegisterChecker(checker HealthChecker) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checkers[checker.Name()] = checker
	return nil
}

// Check 执行检查
func (m *Monitor) Check() map[string]HealthCheckResult {
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, checker := range m.checkers {
		m.results[name] = checker.Check()
	}
	return m.results
}

// GetStatus 获取总体状态
func (m *Monitor) GetStatus() HealthStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, result := range m.results {
		if result.Status == Unhealthy {
			return Unhealthy
		}
	}
	return Healthy
}
