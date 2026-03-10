package config

import (
	"fmt"
	"time"
)

// translated comment
type HealthStatus string

const (
	HealthOK       HealthStatus = "ok"
	HealthWarning  HealthStatus = "warning"
	HealthCritical HealthStatus = "critical"
)

// translated comment
type HealthCheckResult struct {
	Status    HealthStatus
	Message   string
	Timestamp time.Time
	Checks    []*HealthCheckItem
	Overall   bool
}

// translated comment
type HealthCheckItem struct {
	Name        string
	Status      HealthStatus
	Message     string
	LastChecked time.Time
}

// translated comment
type HealthChecker struct {
	center *ConfigCenter
	checks []HealthCheck
}

// translated comment
type HealthCheck interface {
	// translated comment
	Check() *HealthCheckItem
}

// translated comment
func NewHealthChecker(center *ConfigCenter) *HealthChecker {
	hc := &HealthChecker{
		center: center,
		checks: make([]HealthCheck, 0),
	}

	// translated comment
	hc.registerDefaultChecks()

	return hc
}

// translated comment
func (hc *HealthChecker) registerDefaultChecks() {
	// translated comment
	hc.AddCheck(&loadedCheck{center: hc.center})

	// translated comment
	hc.AddCheck(&validityCheck{center: hc.center})

	// translated comment
	hc.AddCheck(&historyCheck{center: hc.center})

	// translated comment
	hc.AddCheck(&fileAccessCheck{center: hc.center})
}

// translated comment
func (hc *HealthChecker) AddCheck(check HealthCheck) {
	hc.checks = append(hc.checks, check)
}

// translated comment
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

		// translated comment
		if item.Status == HealthCritical {
			result.Status = HealthCritical
			result.Overall = false
		} else if item.Status == HealthWarning && result.Status != HealthCritical {
			result.Status = HealthWarning
		}
	}

	// translated comment
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

// translated comment
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

// translated comment
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

	// translated comment
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

// translated comment
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

// translated comment
type fileAccessCheck struct {
	center *ConfigCenter
}

func (fac *fileAccessCheck) Check() *HealthCheckItem {
	item := &HealthCheckItem{
		Name:        "File Access",
		LastChecked: time.Now(),
	}

	// translated comment
	if err := fac.center.Load(); err == nil {
		item.Status = HealthOK
		item.Message = "Configuration file is accessible"
	} else {
		item.Status = HealthWarning
		item.Message = fmt.Sprintf("Error accessing configuration file: %v", err)
	}

	return item
}
