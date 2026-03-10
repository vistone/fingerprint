package config

import (
	"fmt"
)

// ConfigValidator is a configuration validator
type ConfigValidator struct {
	rules []ValidationRule
}

// ValidationRule represents a validation rule
type ValidationRule struct {
	// Rule name
	Name string

	// Validation function
	Validate func(*ManagedConfig) error
}

// ValidationError represents a validation error
type ValidationError struct {
	Field  string
	Reason string
	Value  interface{}
}

func (ve ValidationError) Error() string {
	return fmt.Sprintf("validation error on %s: %s (value: %v)", ve.Field, ve.Reason, ve.Value)
}

// NewConfigValidator creates a new configuration validator
func NewConfigValidator() *ConfigValidator {
	validator := &ConfigValidator{
		rules: make([]ValidationRule, 0),
	}

	// Register default validation rules
	validator.registerDefaultRules()

	return validator
}

// registerDefaultRules registers default validation rules
func (cv *ConfigValidator) registerDefaultRules() {
	// Behavior analysis validation
	cv.AddRule("behavior_analysis_min_requests", func(cfg *ManagedConfig) error {
		if cfg.BehaviorAnalysis == nil {
			return nil
		}
		if cfg.BehaviorAnalysis.MinRequestsForAnalysis <= 0 {
			return ValidationError{
				Field:  "behavior_analysis.min_requests_for_analysis",
				Reason: "must be greater than 0",
				Value:  cfg.BehaviorAnalysis.MinRequestsForAnalysis,
			}
		}
		return nil
	})

	cv.AddRule("behavior_analysis_thresholds", func(cfg *ManagedConfig) error {
		if cfg.BehaviorAnalysis == nil {
			return nil
		}

		if cfg.BehaviorAnalysis.RegularityThreshold < 0 || cfg.BehaviorAnalysis.RegularityThreshold > 1 {
			return ValidationError{
				Field:  "behavior_analysis.regularity_threshold",
				Reason: "must be between 0 and 1",
				Value:  cfg.BehaviorAnalysis.RegularityThreshold,
			}
		}

		if cfg.BehaviorAnalysis.EntropyThreshold < 0 || cfg.BehaviorAnalysis.EntropyThreshold > 1 {
			return ValidationError{
				Field:  "behavior_analysis.entropy_threshold",
				Reason: "must be between 0 and 1",
				Value:  cfg.BehaviorAnalysis.EntropyThreshold,
			}
		}

		if cfg.BehaviorAnalysis.AnomalousIntervalRateThreshold < 0 || cfg.BehaviorAnalysis.AnomalousIntervalRateThreshold > 1 {
			return ValidationError{
				Field:  "behavior_analysis.anomalous_interval_rate_threshold",
				Reason: "must be between 0 and 1",
				Value:  cfg.BehaviorAnalysis.AnomalousIntervalRateThreshold,
			}
		}

		return nil
	})

	// Risk scoring validation
	cv.AddRule("risk_scoring_thresholds", func(cfg *ManagedConfig) error {
		if cfg.RiskScoring == nil {
			return nil
		}

		// Validate threshold order
		if cfg.RiskScoring.CriticalThreshold <= cfg.RiskScoring.HighThreshold {
			return ValidationError{
				Field:  "risk_scoring.thresholds",
				Reason: "critical_threshold must be > high_threshold",
				Value:  fmt.Sprintf("critical=%f, high=%f", cfg.RiskScoring.CriticalThreshold, cfg.RiskScoring.HighThreshold),
			}
		}

		// Validate effective range
		for threshold := cfg.RiskScoring.CriticalThreshold; threshold <= cfg.RiskScoring.LowThreshold; {
			if threshold < 0 || threshold > 1 {
				return ValidationError{
					Field:  "risk_scoring.thresholds",
					Reason: "all thresholds must be between 0 and 1",
					Value:  threshold,
				}
			}
			break
		}

		return nil
	})

	// Feature extraction validation
	cv.AddRule("features_thresholds", func(cfg *ManagedConfig) error {
		if cfg.Features == nil {
			return nil
		}

		if cfg.Features.EntropyHighThreshold < 0 {
			return ValidationError{
				Field:  "features.entropy_high_threshold",
				Reason: "must be >= 0",
				Value:  cfg.Features.EntropyHighThreshold,
			}
		}

		if cfg.Features.EntropyLowThreshold < 0 {
			return ValidationError{
				Field:  "features.entropy_low_threshold",
				Reason: "must be >= 0",
				Value:  cfg.Features.EntropyLowThreshold,
			}
		}

		if cfg.Features.MobileScreenWidthMax <= 0 {
			return ValidationError{
				Field:  "features.mobile_screen_width_max",
				Reason: "must be > 0",
				Value:  cfg.Features.MobileScreenWidthMax,
			}
		}

		if cfg.Features.DesktopScreenWidthMin <= 0 {
			return ValidationError{
				Field:  "features.desktop_screen_width_min",
				Reason: "must be > 0",
				Value:  cfg.Features.DesktopScreenWidthMin,
			}
		}

		return nil
	})

	// QUIC validation
	cv.AddRule("quic_parameters", func(cfg *ManagedConfig) error {
		if cfg.QUIC == nil {
			return nil
		}

		if cfg.QUIC.MinInitialMaxData < 0 {
			return ValidationError{
				Field:  "quic.min_initial_max_data",
				Reason: "must be >= 0",
				Value:  cfg.QUIC.MinInitialMaxData,
			}
		}

		if cfg.QUIC.MinStreamData < 0 {
			return ValidationError{
				Field:  "quic.min_stream_data",
				Reason: "must be >= 0",
				Value:  cfg.QUIC.MinStreamData,
			}
		}

		return nil
	})

	// Global configuration validation
	cv.AddRule("global_config", func(cfg *ManagedConfig) error {
		if cfg.Global == nil {
			return nil
		}

		if cfg.Global.MaxConcurrency <= 0 {
			return ValidationError{
				Field:  "global.max_concurrency",
				Reason: "must be > 0",
				Value:  cfg.Global.MaxConcurrency,
			}
		}

		if cfg.Global.RequestTimeout < 0 {
			return ValidationError{
				Field:  "global.request_timeout",
				Reason: "must be >= 0",
				Value:  cfg.Global.RequestTimeout,
			}
		}

		return nil
	})
}

// AddRule adds a custom validation rule
func (cv *ConfigValidator) AddRule(name string, validate func(*ManagedConfig) error) {
	cv.rules = append(cv.rules, ValidationRule{
		Name:     name,
		Validate: validate,
	})
}

// Validate validates the configuration
func (cv *ConfigValidator) Validate(config *ManagedConfig) []error {
	errors := make([]error, 0)

	for _, rule := range cv.rules {
		if err := rule.Validate(config); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// ValidateField validates the specified field
func (cv *ConfigValidator) ValidateField(fieldPath string, value interface{}) error {
	// Simplified implementation - in practice, more complex field validation logic should be used
	return nil
}
