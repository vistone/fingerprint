package extension

import (
	"fmt"
	"sync"
	"time"
)

// RateLimiter 速率限制器，防止请求溅射
//
// 使用示例：
//
//	limiter := extension.NewRateLimiter(100, time.Minute)
//	if err := limiter.Allow(); err != nil {
//	    return err  // 超过限制
//	}
//	// 处理请求
//
// 工作原理：令牌桶算法
//   - 每个时间窗口允许 maxRequests 个请求
//   - 超过则返回 ErrCodeResourceExhausted
//   - 使用互斥锁保证线程安全
type RateLimiter struct {
	mu          sync.Mutex
	maxRequests int
	timeWindow  time.Duration
	requests    []time.Time
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(maxRequests int, timeWindow time.Duration) *RateLimiter {
	return &RateLimiter{
		maxRequests: maxRequests,
		timeWindow:  timeWindow,
		requests:    make([]time.Time, 0, maxRequests),
	}
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow() error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.timeWindow)

	// 移除旧的请求记录
	validRequests := []time.Time{}
	for _, t := range rl.requests {
		if t.After(cutoff) {
			validRequests = append(validRequests, t)
		}
	}
	rl.requests = validRequests

	// 检查是否超过限制
	if len(rl.requests) >= rl.maxRequests {
		return NewError(ErrCodeResourceExhausted,
			fmt.Sprintf("rate limit exceeded: %d requests in %v", rl.maxRequests, rl.timeWindow)).
			WithContext("limit", rl.maxRequests).
			WithContext("window", rl.timeWindow.String())
	}

	rl.requests = append(rl.requests, now)
	return nil
}

// ResourceMonitor 资源监测器
type ResourceMonitor struct {
	mu                sync.RWMutex
	maxMemoryBytes    int64
	maxGoroutines     int
	maxTimeoutSeconds int

	startTime   time.Time
	allocations map[string]int64 // 追踪按类型分配
}

// NewResourceMonitor 创建资源监测器
func NewResourceMonitor(maxMemoryMB int, maxGoroutines int, maxTimeoutSec int) *ResourceMonitor {
	return &ResourceMonitor{
		maxMemoryBytes:    int64(maxMemoryMB) * 1024 * 1024,
		maxGoroutines:     maxGoroutines,
		maxTimeoutSeconds: maxTimeoutSec,
		startTime:         time.Now(),
		allocations:       make(map[string]int64),
	}
}

// CheckMemory 检查内存使用
func (rm *ResourceMonitor) CheckMemory(size int64, label string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	current := rm.allocations[label]
	if current+size > rm.maxMemoryBytes {
		return NewError(ErrCodeMemoryExhausted,
			fmt.Sprintf("memory allocation would exceed limit: %d + %d > %d",
				current, size, rm.maxMemoryBytes)).
			WithContext("label", label).
			WithContext("current", current).
			WithContext("requested", size)
	}

	rm.allocations[label] = current + size
	return nil
}

// ReleaseMemory 释放内存
func (rm *ResourceMonitor) ReleaseMemory(size int64, label string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	current := rm.allocations[label]
	if current >= size {
		rm.allocations[label] = current - size
	} else {
		rm.allocations[label] = 0
	}
}

// GetMemoryUsage 获取内存使用统计
func (rm *ResourceMonitor) GetMemoryUsage() map[string]int64 {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	result := make(map[string]int64)
	for k, v := range rm.allocations {
		result[k] = v
	}
	return result
}

// CheckTimeout 检查超时
func (rm *ResourceMonitor) CheckTimeout() error {
	elapsed := time.Since(rm.startTime).Seconds()
	if int(elapsed) > rm.maxTimeoutSeconds {
		return NewError(ErrCodeTimeout,
			fmt.Sprintf("operation timeout exceeded: %.2fs > %ds", elapsed, rm.maxTimeoutSeconds)).
			WithContext("elapsed", fmt.Sprintf("%.2fs", elapsed))
	}
	return nil
}

// DefensePolicy 防御策略配置
//
// 使用示例（默认策略）：
//
//	policy := extension.DefaultDefensePolicy()
//	guard := extension.NewRequestGuard(policy)
//	if err := guard.ValidateRequest(request); err != nil {
//	    return err
//	}
//
// 自定义严格策略：
//
//	policy := &extension.DefensePolicy{
//	    ValidateInput:    true,
//	    MaxInputSize:     4096,      // 4KB
//	    LimitMemory:      true,
//	    MaxMemoryMB:      256,
//	    EnableTimeout:    true,
//	    TimeoutSec:       10,
//	    EnableRateLimit:  true,
//	    RateLimit:        500,       // 500 req/min
//	    StrictMode:       true,
//	}
//
// 默认值（推荐）：
//
//	MaxInputSize: 65536 (64KB)
//	MaxMemoryMB: 256
//	TimeoutSec: 30
//	RateLimit: 1000 (1000 req/min)
type DefensePolicy struct {
	// 输入检查
	ValidateInput bool
	MaxInputSize  int

	// 资源限制
	LimitMemory bool
	MaxMemoryMB int

	// 超时设置
	EnableTimeout bool
	TimeoutSec    int

	// 速率限制
	EnableRateLimit bool
	RateLimit       int // 每分钟请求数

	// 严格模式
	StrictMode bool
}

// DefaultDefensePolicy 默认防御策略
func DefaultDefensePolicy() *DefensePolicy {
	return &DefensePolicy{
		ValidateInput:   true,
		MaxInputSize:    65536, // 64KB
		LimitMemory:     true,
		MaxMemoryMB:     256, // 256MB
		EnableTimeout:   true,
		TimeoutSec:      30, // 30 秒
		EnableRateLimit: true,
		RateLimit:       1000, // 1000 请求/分钟
		StrictMode:      true,
	}
}

// SecurityAuditor 安全审计员
type SecurityAuditor struct {
	mu             sync.Mutex
	events         []AuditEvent
	maxEvents      int
	alertThreshold int // 警告阈值
	blockThreshold int // 阻止阈值
}

// AuditEvent 审计事件
type AuditEvent struct {
	Timestamp time.Time
	EventType string // "validation_failed", "resource_exceeded", "security_violation"
	Severity  string // "info", "warning", "critical"
	Message   string
	Details   map[string]interface{}
}

// NewSecurityAuditor 创建安全审计员
func NewSecurityAuditor(maxEvents int) *SecurityAuditor {
	return &SecurityAuditor{
		events:         make([]AuditEvent, 0, maxEvents),
		maxEvents:      maxEvents,
		alertThreshold: 10, // 10 个警告后发送警报
		blockThreshold: 20, // 20 个严重事件后阻止
	}
}

// RecordEvent 记录事件
func (sa *SecurityAuditor) RecordEvent(eventType, severity, message string, details map[string]interface{}) error {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	event := AuditEvent{
		Timestamp: time.Now(),
		EventType: eventType,
		Severity:  severity,
		Message:   message,
		Details:   details,
	}

	// 检查是否应该阻止
	criticalCount := 0
	for _, e := range sa.events {
		if e.Severity == "critical" {
			criticalCount++
		}
	}

	if criticalCount >= sa.blockThreshold {
		return NewError(ErrCodeSecurityViolation,
			"security policy threshold exceeded").
			WithContext("critical_events", criticalCount)
	}

	// 添加事件
	sa.events = append(sa.events, event)

	// 保持最大事件数限制
	if len(sa.events) > sa.maxEvents {
		sa.events = sa.events[len(sa.events)-sa.maxEvents:]
	}

	return nil
}

// GetAuditLog 获取审计日志
func (sa *SecurityAuditor) GetAuditLog() []AuditEvent {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	result := make([]AuditEvent, len(sa.events))
	copy(result, sa.events)
	return result
}

// ClearAuditLog 清除审计日志
func (sa *SecurityAuditor) ClearAuditLog() {
	sa.mu.Lock()
	defer sa.mu.Unlock()
	sa.events = make([]AuditEvent, 0, sa.maxEvents)
}

// RequestGuard 请求守卫，集成所有防御机制
//
// 装配所有防御组件：RateLimiter、ResourceMonitor、SecurityAuditor、Validator
//
// 使用示例：
//
//	policy := extension.DefaultDefensePolicy()
//	guard := extension.NewRequestGuard(policy)
//
//	// 验证请求
//	if err := guard.ValidateRequest(request); err != nil {
//	    // 请求被拒绝（大小超限、速率限制、超时等）
//	    return err
//	}
//
//	// 请求已验证，可以安全处理
//	return processRequest(request)
//
// 防御层次（按执行顺序）：
//  1. 速率限制 - 防止流量溅射
//  2. 大小检查 - 防止超大请求
//  3. 超时检查 - 防止无限等待
//  4. 数据验证 - 防止格式错误和恶意输入
//
// 所有检查失败都会被记录到审计日志
type RequestGuard struct {
	policy    *DefensePolicy
	monitor   *ResourceMonitor
	auditor   *SecurityAuditor
	validator *DefaultValidator
	limiter   *RateLimiter
}

// NewRequestGuard 创建请求守卫
func NewRequestGuard(policy *DefensePolicy) *RequestGuard {
	rg := &RequestGuard{
		policy:    policy,
		monitor:   NewResourceMonitor(policy.MaxMemoryMB, 1000, policy.TimeoutSec),
		auditor:   NewSecurityAuditor(1000),
		validator: NewDefaultValidator(),
		limiter:   NewRateLimiter(policy.RateLimit, time.Minute),
	}

	return rg
}

// ValidateRequest 验证请求
//
// 执行四层防御检查（按顺序）：
// 1. 速率限制 - 防止流量溅射
// 2. 大小检查 - 防止超大请求
// 3. 超时检查 - 防止无限等待
// 4. 数据验证 - 防止格式错误和恶意输入
//
// 所有失败都会被记录到审计日志
func (rg *RequestGuard) ValidateRequest(data []byte) error {
	// 第1层：速率限制
	if rg.policy.EnableRateLimit {
		if err := rg.limiter.Allow(); err != nil {
			return rg.logAndReturnError(
				"rate_limit_exceeded",
				"warning",
				"rate limit exceeded",
				map[string]interface{}{"size": len(data)},
				err,
			)
		}
	}

	// 第2层：大小检查
	if rg.policy.ValidateInput {
		if len(data) > rg.policy.MaxInputSize {
			err := NewError(ErrCodeFieldSizeMismatch,
				fmt.Sprintf("input exceeds maximum: %d > %d", len(data), rg.policy.MaxInputSize))
			return rg.logAndReturnError(
				"validation_failed",
				"critical",
				"input size exceeds policy",
				map[string]interface{}{
					"actual":  len(data),
					"maximum": rg.policy.MaxInputSize,
				},
				err,
			)
		}
	}

	// 第3层：超时检查
	if rg.policy.EnableTimeout {
		if err := rg.monitor.CheckTimeout(); err != nil {
			return rg.logAndReturnError(
				"timeout",
				"critical",
				"request timeout exceeded",
				map[string]interface{}{},
				err,
			)
		}
	}

	// 第4层：数据验证
	if err := rg.validator.ValidateData(data); err != nil {
		return rg.logAndReturnError(
			"validation_failed",
			"warning",
			"data validation failed",
			map[string]interface{}{"error": err.Error()},
			err,
		)
	}

	return nil
}

// logAndReturnError 记录审计事件并返回错误
// 提取重复的错误处理和审计日志记录逻辑
func (rg *RequestGuard) logAndReturnError(
	eventType, severity, message string,
	details map[string]interface{},
	err error,
) error {
	// 忽略 RecordEvent 可能返回的错误，因为审计失败不应该阻止主流程
	// 但优先级检查通过 blockThreshold 机制自动处理
	_ = rg.auditor.RecordEvent(eventType, severity, message, details)
	return err
}

// DefenseConfig 防御配置助手
type DefenseConfig struct {
	guards map[string]*RequestGuard
	mu     sync.RWMutex
}

// NewDefenseConfig 创建防御配置
func NewDefenseConfig() *DefenseConfig {
	return &DefenseConfig{
		guards: make(map[string]*RequestGuard),
	}
}

// RegisterGuard 注册守卫
func (dc *DefenseConfig) RegisterGuard(name string, policy *DefensePolicy) error {
	if name == "" {
		return NewError(ErrCodeInvalidConfig, "guard name cannot be empty")
	}

	dc.mu.Lock()
	defer dc.mu.Unlock()

	if _, ok := dc.guards[name]; ok {
		return NewError(ErrCodeAlreadyRegistered,
			fmt.Sprintf("guard already registered: %s", name))
	}

	dc.guards[name] = NewRequestGuard(policy)
	return nil
}

// GetGuard 获取守卫
func (dc *DefenseConfig) GetGuard(name string) (*RequestGuard, error) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	guard, ok := dc.guards[name]
	if !ok {
		return nil, NewError(ErrCodeNotFound,
			fmt.Sprintf("guard not found: %s", name))
	}

	return guard, nil
}

// ValidationResult 验证结果
type ValidationResult struct {
	Valid     bool
	Error     error
	Warnings  []string
	Details   map[string]interface{}
	CheckedAt time.Time
}

// ComprehensiveValidation 全面验证
func ComprehensiveValidation(data []byte, metadata *ExtensionMetadata, policy *DefensePolicy) *ValidationResult {
	result := &ValidationResult{
		Valid:     true,
		Warnings:  make([]string, 0),
		Details:   make(map[string]interface{}),
		CheckedAt: time.Now(),
	}

	// 创建验证器
	validator := NewDefaultValidator()

	// 验证数据
	if err := validator.ValidateData(data); err != nil {
		result.Valid = false
		result.Error = err
		return result
	}

	// 验证元数据
	if metadata != nil {
		if err := validator.ValidateMetadata(metadata); err != nil {
			if extErr, ok := err.(*Error); ok && extErr.Severity == SeverityWarning {
				result.Warnings = append(result.Warnings, err.Error())
			} else {
				result.Valid = false
				result.Error = err
				return result
			}
		}
	}

	result.Details["data_size"] = len(data)
	if metadata != nil {
		result.Details["extension_type"] = metadata.Type
		result.Details["extension_name"] = metadata.Name
	}

	return result
}
