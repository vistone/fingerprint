package extension

import (
	"fmt"
	"sync"
	"time"
)

// translated comment
//
// translated comment
//
//	limiter := extension.NewRateLimiter(100, time.Minute)
//	if err := limiter.Allow(); err != nil {
// translated comment
//	}
// translated comment
//
// translated comment
// translated comment
// translated comment
// translated comment
type RateLimiter struct {
	mu          sync.Mutex
	maxRequests int
	timeWindow  time.Duration
	requests    []time.Time
}

// translated comment
func NewRateLimiter(maxRequests int, timeWindow time.Duration) *RateLimiter {
	return &RateLimiter{
		maxRequests: maxRequests,
		timeWindow:  timeWindow,
		requests:    make([]time.Time, 0, maxRequests),
	}
}

// translated comment
func (rl *RateLimiter) Allow() error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.timeWindow)

	// translated comment
	validRequests := []time.Time{}
	for _, t := range rl.requests {
		if t.After(cutoff) {
			validRequests = append(validRequests, t)
		}
	}
	rl.requests = validRequests

	// translated comment
	if len(rl.requests) >= rl.maxRequests {
		return NewError(ErrCodeResourceExhausted,
			fmt.Sprintf("rate limit exceeded: %d requests in %v", rl.maxRequests, rl.timeWindow)).
			WithContext("limit", rl.maxRequests).
			WithContext("window", rl.timeWindow.String())
	}

	rl.requests = append(rl.requests, now)
	return nil
}

// translated comment
type ResourceMonitor struct {
	mu                sync.RWMutex
	maxMemoryBytes    int64
	maxGoroutines     int
	maxTimeoutSeconds int

	startTime   time.Time
	allocations map[string]int64 // translated comment
}

// translated comment
func NewResourceMonitor(maxMemoryMB int, maxGoroutines int, maxTimeoutSec int) *ResourceMonitor {
	return &ResourceMonitor{
		maxMemoryBytes:    int64(maxMemoryMB) * 1024 * 1024,
		maxGoroutines:     maxGoroutines,
		maxTimeoutSeconds: maxTimeoutSec,
		startTime:         time.Now(),
		allocations:       make(map[string]int64),
	}
}

// translated comment
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

// translated comment
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

// translated comment
func (rm *ResourceMonitor) GetMemoryUsage() map[string]int64 {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	result := make(map[string]int64)
	for k, v := range rm.allocations {
		result[k] = v
	}
	return result
}

// translated comment
func (rm *ResourceMonitor) CheckTimeout() error {
	elapsed := time.Since(rm.startTime).Seconds()
	if int(elapsed) > rm.maxTimeoutSeconds {
		return NewError(ErrCodeTimeout,
			fmt.Sprintf("operation timeout exceeded: %.2fs > %ds", elapsed, rm.maxTimeoutSeconds)).
			WithContext("elapsed", fmt.Sprintf("%.2fs", elapsed))
	}
	return nil
}

// translated comment
//
// translated comment
//
//	policy := extension.DefaultDefensePolicy()
//	guard := extension.NewRequestGuard(policy)
//	if err := guard.ValidateRequest(request); err != nil {
//	    return err
//	}
//
// translated comment
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
// translated comment
//
//	MaxInputSize: 65536 (64KB)
//	MaxMemoryMB: 256
//	TimeoutSec: 30
//	RateLimit: 1000 (1000 req/min)
type DefensePolicy struct {
	// translated comment
	ValidateInput bool
	MaxInputSize  int

	// translated comment
	LimitMemory bool
	MaxMemoryMB int

	// translated comment
	EnableTimeout bool
	TimeoutSec    int

	// translated comment
	EnableRateLimit bool
	RateLimit       int // translated comment

	// translated comment
	StrictMode bool
}

// translated comment
func DefaultDefensePolicy() *DefensePolicy {
	return &DefensePolicy{
		ValidateInput:   true,
		MaxInputSize:    65536, // 64KB
		LimitMemory:     true,
		MaxMemoryMB:     256, // 256MB
		EnableTimeout:   true,
		TimeoutSec:      30, // translated comment
		EnableRateLimit: true,
		RateLimit:       1000, // translated comment
		StrictMode:      true,
	}
}

// translated comment
type SecurityAuditor struct {
	mu             sync.Mutex
	events         []AuditEvent
	maxEvents      int
	alertThreshold int // translated comment
	blockThreshold int // translated comment
}

// translated comment
type AuditEvent struct {
	Timestamp time.Time
	EventType string // "validation_failed", "resource_exceeded", "security_violation"
	Severity  string // "info", "warning", "critical"
	Message   string
	Details   map[string]interface{}
}

// translated comment
func NewSecurityAuditor(maxEvents int) *SecurityAuditor {
	return &SecurityAuditor{
		events:         make([]AuditEvent, 0, maxEvents),
		maxEvents:      maxEvents,
		alertThreshold: 10, // translated comment
		blockThreshold: 20, // translated comment
	}
}

// translated comment
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

	// translated comment
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

	// translated comment
	sa.events = append(sa.events, event)

	// translated comment
	if len(sa.events) > sa.maxEvents {
		sa.events = sa.events[len(sa.events)-sa.maxEvents:]
	}

	return nil
}

// translated comment
func (sa *SecurityAuditor) GetAuditLog() []AuditEvent {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	result := make([]AuditEvent, len(sa.events))
	copy(result, sa.events)
	return result
}

// translated comment
func (sa *SecurityAuditor) ClearAuditLog() {
	sa.mu.Lock()
	defer sa.mu.Unlock()
	sa.events = make([]AuditEvent, 0, sa.maxEvents)
}

// translated comment
//
// translated comment
//
// translated comment
//
//	policy := extension.DefaultDefensePolicy()
//	guard := extension.NewRequestGuard(policy)
//
// translated comment
//	if err := guard.ValidateRequest(request); err != nil {
// translated comment
//	    return err
//	}
//
// translated comment
//	return processRequest(request)
//
// translated comment
// translated comment
// translated comment
// translated comment
// translated comment
//
// translated comment
type RequestGuard struct {
	policy    *DefensePolicy
	monitor   *ResourceMonitor
	auditor   *SecurityAuditor
	validator *DefaultValidator
	limiter   *RateLimiter
}

// translated comment
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

// translated comment
//
// translated comment
// translated comment
// translated comment
// translated comment
// translated comment
//
// translated comment
func (rg *RequestGuard) ValidateRequest(data []byte) error {
	// translated comment
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

	// translated comment
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

	// translated comment
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

	// translated comment
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

// translated comment
// translated comment
func (rg *RequestGuard) logAndReturnError(
	eventType, severity, message string,
	details map[string]interface{},
	err error,
) error {
	// translated comment
	// translated comment
	_ = rg.auditor.RecordEvent(eventType, severity, message, details)
	return err
}

// translated comment
type DefenseConfig struct {
	guards map[string]*RequestGuard
	mu     sync.RWMutex
}

// translated comment
func NewDefenseConfig() *DefenseConfig {
	return &DefenseConfig{
		guards: make(map[string]*RequestGuard),
	}
}

// translated comment
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

// translated comment
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

// translated comment
type ValidationResult struct {
	Valid     bool
	Error     error
	Warnings  []string
	Details   map[string]interface{}
	CheckedAt time.Time
}

// translated comment
func ComprehensiveValidation(data []byte, metadata *ExtensionMetadata, policy *DefensePolicy) *ValidationResult {
	result := &ValidationResult{
		Valid:     true,
		Warnings:  make([]string, 0),
		Details:   make(map[string]interface{}),
		CheckedAt: time.Now(),
	}

	// translated comment
	validator := NewDefaultValidator()

	// translated comment
	if err := validator.ValidateData(data); err != nil {
		result.Valid = false
		result.Error = err
		return result
	}

	// translated comment
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
