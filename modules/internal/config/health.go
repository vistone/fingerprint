package config

import (
	"fmt"
	"time"
)

// HealthStatus represents a health check status
type HealthStatus string

const (
	HealthOK       HealthStatus = "ok"
	HealthWarning  HealthStatus = "warning"
	HealthCritical HealthStatus = "critical"
)

// HealthCheckResult represents a health check result
type HealthCheckResult struct {
	Status    HealthStatus
	Message   string
	Timestamp time.Time
	Checks    []*HealthCheckItem
	Overall   bool
}

// HealthCheckItem represents a single health check item
type HealthCheckItem struct {
	Name        string
	Status      HealthStatus
	Message     string
	LastChecked time.Time
}

// HealthChecker is a configuration health checker
type HealthChecker struct {
	center *ConfigCenter
	checks []HealthCheck
}

// HealthCheck is the health check function interface
type HealthCheck interface {
	// Check performs a health check
	Check() *HealthCheckItem
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(center *ConfigCenter) *HealthChecker {
	hc := &HealthChecker{
		center: center,
		checks: make([]HealthCheck, 0),
	}

	// Register default checks
	hc.registerDefaultChecks()

	return hc
}

// registerDefaultChecks registers default checks
func (hc *HealthChecker) registerDefaultChecks() {
	// Check whether the configuration has been loaded
	hc.AddCheck(&loadedCheck{center: hc.center})

	// Check configuration validity
	hc.AddCheck(&validityCheck{center: hc.center})

	// Check version history
	hc.AddCheck(&historyCheck{center: hc.center})

	// Check configuration file accessibility
	hc.AddCheck(&fileAccessCheck{center: hc.center})
}

// AddCheck adds a health check
func (hc *HealthChecker) AddCheck(check HealthCheck) {
	hc.checks = append(hc.checks, check)
}

// CheckHealth performs all health checks
func (hc *HealthChecker) CheckHealth() *HealthCheckResult {
	result := &HealthCheckResult{
		Status:    HealthOK,
		Timestamp: time.Now(),
		Checks:    make([]*HealthCheckItem, 0),
		Overall:   true,
	}

	for _, check := range hc.checks {
		item := check.Check()
		result.Checks = append(result.Checks, item)

		// Update overall status
		if item.Status == HealthCritical {
			result.Status = HealthCritical
			result.Overall = false
		} else if item.Status == HealthWarning && result.Status != HealthCritical {
			result.Status = HealthWarning
		}
	}

	// Generate message
	if result.Overall {
		result.Message = "Configuration health is good"
	} else {
		failCount := 0
		for _, check := range result.Checks {
			if check.Status != HealthOK {
				failCount++
			}
		}
		result.Message = fmt.Sprintf("%d checks failed or have warnings", failCount)
	}

	return result
}

// loadedCheck checks whether the configuration has been loaded
type loadedCheck struct {
	center *ConfigCenter
}

func (lc *loadedCheck) Check() *HealthCheckItem {
	item := &HealthCheckItem{
		Name:        "Config Loaded",
		LastChecked: time.Now(),
	}

	if lc.center.IsLoaded() {
		item.Status = HealthOK
		item.Message = "Configuration has been successfully loaded"
	} else {
		item.Status = HealthCritical
		item.Message = "Configuration has not been loaded"
	}

	return item
}

// validityCheck checks configuration validity
type validityCheck struct {
	center *ConfigCenter
}

func (vc *validityCheck) Check() *HealthCheckItem {
	item := &HealthCheckItem{
		Name:        "Config Validity",
		LastChecked: time.Now(),
	}

	config := vc.center.Get()
	if config == nil {
		item.Status = HealthCritical
		item.Message = "Configuration is nil"
		return item
	}

	// Validate configuration
	validator := NewConfigValidator()
	errs := validator.Validate(config)

	if len(errs) == 0 {
		item.Status = HealthOK
		item.Message = "Configuration validation passed"
	} else {
		item.Status = HealthCritical
		item.Message = fmt.Sprintf("Configuration validation failed: %d errors", len(errs))
	}

	return item
}

// historyCheck checks version history
type historyCheck struct {
	center *ConfigCenter
}

func (hc *historyCheck) Check() *HealthCheckItem {
	item := &HealthCheckItem{
		Name:        "Version History",
		LastChecked: time.Now(),
	}

	history := hc.center.GetHistory(100)

	if len(history) == 0 {
		item.Status = HealthWarning
		item.Message = "No version history"
	} else {
		item.Status = HealthOK
		item.Message = fmt.Sprintf("%d versions in history", len(history))
	}

	return item
}

// fileAccessCheck checks configuration file accessibility
type fileAccessCheck struct {
	center *ConfigCenter
}

func (fac *fileAccessCheck) Check() *HealthCheckItem {
	item := &HealthCheckItem{
		Name:        "File Access",
		LastChecked: time.Now(),
	}

	// Try to read the configuration file
	if err := fac.center.Load(); err == nil {
		item.Status = HealthOK
		item.Message = "Configuration file is accessible"
	} else {
		item.Status = HealthWarning
		item.Message = fmt.Sprintf("Error accessing configuration file: %v", err)
	}

	return item
}
