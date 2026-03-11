package extension

import (
	"fmt"
	"sync"
	"time"
)

// RateLimiter is a rate limiter that prevents request flooding
//
// Usage example:
//
//	limiter := extension.NewRateLimiter(100, time.Minute)
//	if err := limiter.Allow(); err != nil {
//	    return err  // rate limit exceeded
//	}
//	// process the request
//
// How it works: token bucket algorithm
//   - Allows maxRequests requests per time window
//   - Returns ErrCodeResourceExhausted when exceeded
//   - Uses mutex to ensure thread safety
type RateLimiter struct {
	mu          sync.Mutex
	maxRequests int
	timeWindow  time.Duration
	requests    []time.Time
}

// NewRateLimiter creates a rate limiter
func NewRateLimiter(maxRequests int, timeWindow time.Duration) *RateLimiter {
	return &RateLimiter{
		maxRequests: maxRequests,
		timeWindow:  timeWindow,
		requests:    make([]time.Time, 0, maxRequests),
	}
}

// Allow checks whether a request is allowed
func (rl *RateLimiter) Allow() error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.timeWindow)

	// Remove old request records
	validRequests := []time.Time{}
	for _, t := range rl.requests {
		if t.After(cutoff) {
			validRequests = append(validRequests, t)
		}
	}
	rl.requests = validRequests

	// Check if limit is exceeded
	if len(rl.requests) >= rl.maxRequests {
		return NewError(ErrCodeResourceExhausted,
			fmt.Sprintf("rate limit exceeded: %d requests in %v", rl.maxRequests, rl.timeWindow)).
			WithContext("limit", rl.maxRequests).
			WithContext("window", rl.timeWindow.String())
	}

	rl.requests = append(rl.requests, now)
	return nil
}

// ResourceMonitor monitors resource usage
type ResourceMonitor struct {
	mu                sync.RWMutex
	maxMemoryBytes    int64
	maxGoroutines     int
	maxTimeoutSeconds int

	startTime   time.Time
	allocations map[string]int64 // tracks allocations by type
}

// NewResourceMonitor creates a resource monitor
func NewResourceMonitor(maxMemoryMB int, maxGoroutines int, maxTimeoutSec int) *ResourceMonitor {
	return &ResourceMonitor{
		maxMemoryBytes:    int64(maxMemoryMB) * 1024 * 1024,
		maxGoroutines:     maxGoroutines,
		maxTimeoutSeconds: maxTimeoutSec,
		startTime:         time.Now(),
		allocations:       make(map[string]int64),
	}
}

// CheckMemory checks memory usage
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

// ReleaseMemory releases memory
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

// GetMemoryUsage returns memory usage statistics
func (rm *ResourceMonitor) GetMemoryUsage() map[string]int64 {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	result := make(map[string]int64)
	for k, v := range rm.allocations {
		result[k] = v
	}
	return result
}

// CheckTimeout checks for timeout
func (rm *ResourceMonitor) CheckTimeout() error {
	elapsed := time.Since(rm.startTime).Seconds()
	if int(elapsed) > rm.maxTimeoutSeconds {
		return NewError(ErrCodeTimeout,
			fmt.Sprintf("operation timeout exceeded: %.2fs > %ds", elapsed, rm.maxTimeoutSeconds)).
			WithContext("elapsed", fmt.Sprintf("%.2fs", elapsed))
	}
	return nil
}

// DefensePolicy holds defense policy configuration
//
// Usage example (default policy):
//
//	policy := extension.DefaultDefensePolicy()
//	guard := extension.NewRequestGuard(policy)
//	if err := guard.ValidateRequest(request); err != nil {
//	    return err
//	}
//
// Custom strict policy:
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
// Default values (recommended):
//
//	MaxInputSize: 65536 (64KB)
//	MaxMemoryMB: 256
//	TimeoutSec: 30
//	RateLimit: 1000 (1000 req/min)
type DefensePolicy struct {
	// Input checks
	ValidateInput bool
	MaxInputSize  int

	// Resource limits
	LimitMemory bool
	MaxMemoryMB int

	// Timeout settings
	EnableTimeout bool
	TimeoutSec    int

	// Rate limit
	EnableRateLimit bool
	RateLimit       int // requests per minute

	// Strict mode
	StrictMode bool
}

// DefaultDefensePolicy returns the default defense policy
func DefaultDefensePolicy() *DefensePolicy {
	return &DefensePolicy{
		ValidateInput:   true,
		MaxInputSize:    65536, // 64KB
		LimitMemory:     true,
		MaxMemoryMB:     256, // 256MB
		EnableTimeout:   true,
		TimeoutSec:      30, // 30 seconds
		EnableRateLimit: true,
		RateLimit:       1000, // 1000 requests/minute
		StrictMode:      true,
	}
}

// SecurityAuditor is the security auditor
type SecurityAuditor struct {
	mu             sync.Mutex
	events         []AuditEvent
	maxEvents      int
	alertThreshold int // alert threshold
	blockThreshold int // block threshold
}

// AuditEvent represents an audit event
type AuditEvent struct {
	Timestamp time.Time
	EventType string // "validation_failed", "resource_exceeded", "security_violation"
	Severity  string // "info", "warning", "critical"
	Message   string
	Details   map[string]interface{}
}

// NewSecurityAuditor creates a security auditor
func NewSecurityAuditor(maxEvents int) *SecurityAuditor {
	return &SecurityAuditor{
		events:         make([]AuditEvent, 0, maxEvents),
		maxEvents:      maxEvents,
		alertThreshold: 10, // send alert after 10 warnings
		blockThreshold: 20, // block after 20 critical events
	}
}

// RecordEvent records an event
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

	// Check if should block
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

	// Add event
	sa.events = append(sa.events, event)

	// Maintain max event count limit
	if len(sa.events) > sa.maxEvents {
		sa.events = sa.events[len(sa.events)-sa.maxEvents:]
	}

	return nil
}

// GetAuditLog returns the audit log
func (sa *SecurityAuditor) GetAuditLog() []AuditEvent {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	result := make([]AuditEvent, len(sa.events))
	copy(result, sa.events)
	return result
}

// ClearAuditLog clears the audit log
func (sa *SecurityAuditor) ClearAuditLog() {
	sa.mu.Lock()
	defer sa.mu.Unlock()
	sa.events = make([]AuditEvent, 0, sa.maxEvents)
}

// RequestGuard is the request guard that integrates all defense mechanisms
//
// Assembles all defense components: RateLimiter, ResourceMonitor, SecurityAuditor, Validator
//
// Usage example:
//
//	policy := extension.DefaultDefensePolicy()
//	guard := extension.NewRequestGuard(policy)
//
//	// validate request
//	if err := guard.ValidateRequest(request); err != nil {
//	    // request rejected (size exceeded, rate limited, timeout, etc.)
//	    return err
//	}
//
//	// request validated, safe to process
//	return processRequest(request)
//
// Defense layers (in execution order):
//  1. Rate limiting - prevents traffic flooding
//  2. Size check - prevents oversized requests
//  3. Timeout check - prevents indefinite waiting
//  4. Data validation - prevents malformed and malicious input
//
// All check failures are recorded in the audit log
type RequestGuard struct {
	policy    *DefensePolicy
	monitor   *ResourceMonitor
	auditor   *SecurityAuditor
	validator *DefaultValidator
	limiter   *RateLimiter
}

// NewRequestGuard creates a request guard
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

// ValidateRequest validates a request
//
// Performs four-layer defense checks (in order):
// 1. Rate limiting - prevents traffic flooding
// 2. Size check - prevents oversized requests
// 3. Timeout check - prevents indefinite waiting
// 4. Data validation - prevents malformed and malicious input
//
// All failures are recorded in the audit log
func (rg *RequestGuard) ValidateRequest(data []byte) error {
	// Layer 1: Rate limiting
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

	// Layer 2: Size check
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

	// Layer 3: Timeout check
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

	// Layer 4: Data validation
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

// logAndReturnError records an audit event and returns an error
// Extracts repeated error handling and audit log recording logic
func (rg *RequestGuard) logAndReturnError(
	eventType, severity, message string,
	details map[string]interface{},
	err error,
) error {
	// Ignore errors from RecordEvent, since audit failures should not block the main flow
	// Priority checks are handled automatically through the blockThreshold mechanism
	_ = rg.auditor.RecordEvent(eventType, severity, message, details)
	return err
}

// DefenseConfig is the defense configuration helper
type DefenseConfig struct {
	guards map[string]*RequestGuard
	mu     sync.RWMutex
}

// NewDefenseConfig creates a defense configuration
func NewDefenseConfig() *DefenseConfig {
	return &DefenseConfig{
		guards: make(map[string]*RequestGuard),
	}
}

// RegisterGuard registers a guard
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

// GetGuard retrieves a guard
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

// ValidationResult holds validation results
type ValidationResult struct {
	Valid     bool
	Error     error
	Warnings  []string
	Details   map[string]interface{}
	CheckedAt time.Time
}

// ComprehensiveValidation performs comprehensive validation
func ComprehensiveValidation(data []byte, metadata *ExtensionMetadata, policy *DefensePolicy) *ValidationResult {
	result := &ValidationResult{
		Valid:     true,
		Warnings:  make([]string, 0),
		Details:   make(map[string]interface{}),
		CheckedAt: time.Now(),
	}

	// Create validator
	validator := NewDefaultValidator()

	// Validate data
	if err := validator.ValidateData(data); err != nil {
		result.Valid = false
		result.Error = err
		return result
	}

	// Validate metadata
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
