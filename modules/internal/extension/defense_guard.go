package extension

import (
	"fmt"
	"sync"
	"time"
)

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
