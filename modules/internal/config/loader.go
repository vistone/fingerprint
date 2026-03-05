package config

import "github.com/vistone/fingerprint/modules/internal/extension"

// 兼容层说明：
// 规则配置主入口已收敛到 extension 包。
// 本包保留同名类型与函数，便于历史调用方平滑迁移。

// RulesConfig 兼容别名（迁移到 extension.RulesConfig）
type RulesConfig = extension.RulesConfig

// EntropyConfig 兼容别名
type EntropyConfig = extension.EntropyConfig

// ToolMarkersConfig 兼容别名
type ToolMarkersConfig = extension.ToolMarkersConfig

// MarkerInfo 兼容别名
type MarkerInfo = extension.MarkerInfo

// HeadlessBrowserConfig 兼容别名
type HeadlessBrowserConfig = extension.HeadlessBrowserConfig

// OSPlatformConfig 兼容别名
type OSPlatformConfig = extension.OSPlatformConfig

// OSRule 兼容别名
type OSRule = extension.OSRule

// UAOSConfig 兼容别名
type UAOSConfig = extension.UAOSConfig

// UARuleOS 兼容别名
type UARuleOS = extension.UARuleOS

// MobileScreenConfig 兼容别名
type MobileScreenConfig = extension.MobileScreenConfig

// MobileScreenRule 兼容别名
type MobileScreenRule = extension.MobileScreenRule

// UAFeatureConfig 兼容别名
type UAFeatureConfig = extension.UAFeatureConfig

// UAFeatureRule 兼容别名
type UAFeatureRule = extension.UAFeatureRule

// ScoringConfig 兼容别名（对应 extension.RulesScoringConfig）
type ScoringConfig = extension.RulesScoringConfig

// RiskLevel 兼容别名
type RiskLevel = extension.RiskLevel

// LoadRulesConfig 兼容转发
func LoadRulesConfig(path string) (*RulesConfig, error) {
	return extension.LoadRulesConfig(path)
}

// LoadRulesConfigByFilename 兼容转发
func LoadRulesConfigByFilename(filename string) (*RulesConfig, error) {
	return extension.LoadRulesConfigByFilename(filename)
}

// DefaultRulesConfig 兼容转发
func DefaultRulesConfig() *RulesConfig {
	return extension.DefaultRulesConfig()
}
