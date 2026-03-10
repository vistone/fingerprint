package monitor

import (
	"sync"
	"time"
)

// HealthStatus represents health status
type HealthStatus string

const (
	Healthy   HealthStatus = "HEALTHY"
	Degraded  HealthStatus = "DEGRADED"
	Unhealthy HealthStatus = "UNHEALTHY"
)

// HealthCheckResult represents check result
type HealthCheckResult struct {
	Name      string
	Status    HealthStatus
	Message   string
	Timestamp time.Time
}

// HealthChecker is the health checker interface
type HealthChecker interface {
	Check() HealthCheckResult
	Name() string
}

// Monitor is the monitor
type Monitor struct {
	checkers map[string]HealthChecker
	results  map[string]HealthCheckResult
	mu       sync.RWMutex
}

// NewMonitor creates a monitor
func NewMonitor() *Monitor {
	return &Monitor{
		checkers: make(map[string]HealthChecker),
		results:  make(map[string]HealthCheckResult),
	}
}

// RegisterChecker registers a checker
func (m *Monitor) RegisterChecker(checker HealthChecker) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checkers[checker.Name()] = checker
	return nil
}

// Check executes checks
func (m *Monitor) Check() map[string]HealthCheckResult {
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, checker := range m.checkers {
		m.results[name] = checker.Check()
	}
	return m.results
}

// GetStatus gets overall status
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
