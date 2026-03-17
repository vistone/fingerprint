package config

import ic "github.com/vistone/fingerprint/modules/internal/config"

// ConfigCenter configuration center.
type ConfigCenter = ic.ConfigCenter

// ConfigManager configuration manager.
type ConfigManager = ic.ConfigManager

// HealthChecker configuration health checker.
type HealthChecker = ic.HealthChecker

// ManagedConfig managed configuration object.
type ManagedConfig = ic.ManagedConfig

// ConfigChange configuration change info.
type ConfigChange = ic.ConfigChange

// ConfigChangeListener configuration change listener interface.
type ConfigChangeListener = ic.ConfigChangeListener

// BehaviorAnalysisConfig defines behavior analysis settings.
type BehaviorAnalysisConfig = ic.BehaviorAnalysisConfig

// RiskScoringConfig risk scoring configuration.
type RiskScoringConfig = ic.RiskScoringConfig

// FeatureExtractionConfig defines feature extraction settings.
type FeatureExtractionConfig = ic.FeatureExtractionConfig

// QUICConfig defines QUIC fingerprint settings.
type QUICConfig = ic.QUICConfig

// TLSConfig defines TLS fingerprint settings.
type TLSConfig = ic.TLSConfig

// GlobalConfig global configuration.
type GlobalConfig = ic.GlobalConfig

// InitializeConfigCenter initialize global configuration center.
func InitializeConfigCenter() error {
	return ic.InitializeConfigCenter()
}

// InitializeConfigCenterWithDefaults initialize with default configuration.
func InitializeConfigCenterWithDefaults() error {
	return ic.InitializeConfigCenterWithDefaults()
}

// GetConfigCenter get global configuration center.
func GetConfigCenter() *ConfigCenter {
	return ic.GetConfigCenter()
}

// GetConfigManager get global configuration manager.
func GetConfigManager() *ConfigManager {
	return ic.GetConfigManager()
}

// GetHealthChecker get global health checker.
func GetHealthChecker() *HealthChecker {
	return ic.GetHealthChecker()
}
