package config

import "github.com/vistone/fingerprint/modules/internal/extension"

// Compatibility layer notes:
// The main entry point for rule configuration has been consolidated into the extension package.
// This package retains types and functions with the same names to facilitate smooth migration for existing callers.

// RulesConfig is a compatibility alias (migrated to extension.RulesConfig)
type RulesConfig = extension.RulesConfig

// EntropyConfig is a compatibility alias
type EntropyConfig = extension.EntropyConfig

// ToolMarkersConfig is a compatibility alias
type ToolMarkersConfig = extension.ToolMarkersConfig

// MarkerInfo is a compatibility alias
type MarkerInfo = extension.MarkerInfo

// HeadlessBrowserConfig is a compatibility alias
type HeadlessBrowserConfig = extension.HeadlessBrowserConfig

// OSPlatformConfig is a compatibility alias
type OSPlatformConfig = extension.OSPlatformConfig

// OSRule is a compatibility alias
type OSRule = extension.OSRule

// UAOSConfig is a compatibility alias
type UAOSConfig = extension.UAOSConfig

// UARuleOS is a compatibility alias
type UARuleOS = extension.UARuleOS

// MobileScreenConfig is a compatibility alias
type MobileScreenConfig = extension.MobileScreenConfig

// MobileScreenRule is a compatibility alias
type MobileScreenRule = extension.MobileScreenRule

// UAFeatureConfig is a compatibility alias
type UAFeatureConfig = extension.UAFeatureConfig

// UAFeatureRule is a compatibility alias
type UAFeatureRule = extension.UAFeatureRule

// ScoringConfig is a compatibility alias (corresponds to extension.RulesScoringConfig)
type ScoringConfig = extension.RulesScoringConfig

// RiskLevel is a compatibility alias
type RiskLevel = extension.RiskLevel

// LoadRulesConfig is a compatibility forwarder
func LoadRulesConfig(path string) (*RulesConfig, error) {
	return extension.LoadRulesConfig(path)
}

// LoadRulesConfigByFilename is a compatibility forwarder
func LoadRulesConfigByFilename(filename string) (*RulesConfig, error) {
	return extension.LoadRulesConfigByFilename(filename)
}

// DefaultRulesConfig is a compatibility forwarder
func DefaultRulesConfig() *RulesConfig {
	return extension.DefaultRulesConfig()
}
