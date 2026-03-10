package config

import "github.com/vistone/fingerprint/modules/internal/extension"

// translated comment
// translated comment
// translated comment

// translated comment
type RulesConfig = extension.RulesConfig

// translated comment
type EntropyConfig = extension.EntropyConfig

// translated comment
type ToolMarkersConfig = extension.ToolMarkersConfig

// translated comment
type MarkerInfo = extension.MarkerInfo

// translated comment
type HeadlessBrowserConfig = extension.HeadlessBrowserConfig

// translated comment
type OSPlatformConfig = extension.OSPlatformConfig

// translated comment
type OSRule = extension.OSRule

// translated comment
type UAOSConfig = extension.UAOSConfig

// translated comment
type UARuleOS = extension.UARuleOS

// translated comment
type MobileScreenConfig = extension.MobileScreenConfig

// translated comment
type MobileScreenRule = extension.MobileScreenRule

// translated comment
type UAFeatureConfig = extension.UAFeatureConfig

// translated comment
type UAFeatureRule = extension.UAFeatureRule

// translated comment
type ScoringConfig = extension.RulesScoringConfig

// translated comment
type RiskLevel = extension.RiskLevel

// translated comment
func LoadRulesConfig(path string) (*RulesConfig, error) {
	return extension.LoadRulesConfig(path)
}

// translated comment
func LoadRulesConfigByFilename(filename string) (*RulesConfig, error) {
	return extension.LoadRulesConfigByFilename(filename)
}

// translated comment
func DefaultRulesConfig() *RulesConfig {
	return extension.DefaultRulesConfig()
}
