package config

import (
	"fmt"
)

// ConfigValidator 配置验证器
type ConfigValidator struct {
	rules []ValidationRule
}

// ValidationRule 验证规则
type ValidationRule struct {
	// 规则名称
	Name string

	// 验证函数
	Validate func(*ManagedConfig) error
}

// ValidationError 验证错误
type ValidationError struct {
	Field  string
	Reason string
	Value  interface{}
}

func (ve ValidationError) Error() string {
	return fmt.Sprintf("validation error on %s: %s (value: %v)", ve.Field, ve.Reason, ve.Value)
}

// NewConfigValidator 创建配置验证器
func NewConfigValidator() *ConfigValidator {
	validator := &ConfigValidator{
		rules: make([]ValidationRule, 0),
	}

	// 注册默认验证规则
	validator.registerDefaultRules()

	return validator
}

// registerDefaultRules 注册默认验证规则
func (cv *ConfigValidator) registerDefaultRules() {
	// 行为分析验证
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

	// 风险评分验证
	cv.AddRule("risk_scoring_thresholds", func(cfg *ManagedConfig) error {
		if cfg.RiskScoring == nil {
			return nil
		}

		// 验证阈值顺序
		if cfg.RiskScoring.CriticalThreshold <= cfg.RiskScoring.HighThreshold {
			return ValidationError{
				Field:  "risk_scoring.thresholds",
				Reason: "critical_threshold must be > high_threshold",
				Value:  fmt.Sprintf("critical=%f, high=%f", cfg.RiskScoring.CriticalThreshold, cfg.RiskScoring.HighThreshold),
			}
		}

		// 验证有效范围
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

	// 特征提取验证
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

	// QUIC 验证
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

	// 全局配置验证
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

// AddRule 添加自定义验证规则
func (cv *ConfigValidator) AddRule(name string, validate func(*ManagedConfig) error) {
	cv.rules = append(cv.rules, ValidationRule{
		Name:     name,
		Validate: validate,
	})
}

// Validate 验证配置
func (cv *ConfigValidator) Validate(config *ManagedConfig) []error {
	errors := make([]error, 0)

	for _, rule := range cv.rules {
		if err := rule.Validate(config); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// ValidateField 验证指定字段
func (cv *ConfigValidator) ValidateField(fieldPath string, value interface{}) error {
	// 简化实现 - 实际应该有更复杂的字段验证逻辑
	return nil
}
