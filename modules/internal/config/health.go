package config

import (
	"fmt"
	"time"
)

// HealthStatus 健康检查状态
type HealthStatus string

const (
	HealthOK       HealthStatus = "ok"
	HealthWarning  HealthStatus = "warning"
	HealthCritical HealthStatus = "critical"
)

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	Status    HealthStatus
	Message   string
	Timestamp time.Time
	Checks    []*HealthCheckItem
	Overall   bool
}

// HealthCheckItem 单个健康检查项
type HealthCheckItem struct {
	Name        string
	Status      HealthStatus
	Message     string
	LastChecked time.Time
}

// HealthChecker 配置健康检查器
type HealthChecker struct {
	center *ConfigCenter
	checks []HealthCheck
}

// HealthCheck 健康检查函数接口
type HealthCheck interface {
	// 执行健康检查
	Check() *HealthCheckItem
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(center *ConfigCenter) *HealthChecker {
	hc := &HealthChecker{
		center: center,
		checks: make([]HealthCheck, 0),
	}

	// 注册默认检查
	hc.registerDefaultChecks()

	return hc
}

// registerDefaultChecks 注册默认检查
func (hc *HealthChecker) registerDefaultChecks() {
	// 检查配置是否已加载
	hc.AddCheck(&loadedCheck{center: hc.center})

	// 检查配置有效性
	hc.AddCheck(&validityCheck{center: hc.center})

	// 检查历史版本
	hc.AddCheck(&historyCheck{center: hc.center})

	// 检查配置文件可访问性
	hc.AddCheck(&fileAccessCheck{center: hc.center})
}

// AddCheck 添加健康检查
func (hc *HealthChecker) AddCheck(check HealthCheck) {
	hc.checks = append(hc.checks, check)
}

// CheckHealth 执行所有健康检查
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

		// 更新整体状态
		if item.Status == HealthCritical {
			result.Status = HealthCritical
			result.Overall = false
		} else if item.Status == HealthWarning && result.Status != HealthCritical {
			result.Status = HealthWarning
		}
	}

	// 生成消息
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

// loadedCheck 检查配置是否已加载
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

// validityCheck 检查配置有效性
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

	// 验证配置
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

// historyCheck 检查历史版本
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

// fileAccessCheck 检查配置文件可访问性
type fileAccessCheck struct {
	center *ConfigCenter
}

func (fac *fileAccessCheck) Check() *HealthCheckItem {
	item := &HealthCheckItem{
		Name:        "File Access",
		LastChecked: time.Now(),
	}

	// 尝试读取配置文件
	if err := fac.center.Load(); err == nil {
		item.Status = HealthOK
		item.Message = "Configuration file is accessible"
	} else {
		item.Status = HealthWarning
		item.Message = fmt.Sprintf("Error accessing configuration file: %v", err)
	}

	return item
}
