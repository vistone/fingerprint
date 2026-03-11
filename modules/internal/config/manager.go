package config

import (
	"fmt"
)

// ConfigManager is a configuration manager that provides a unified configuration access interface
// Note: ConfigManager itself does not need locks because ConfigCenter already has a complete locking mechanism
type ConfigManager struct {
	center *ConfigCenter
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(center *ConfigCenter) *ConfigManager {
	return &ConfigManager{
		center: center,
	}
}

// GetBehaviorAnalysisConfig returns the behavior analysis configuration
func (cm *ConfigManager) GetBehaviorAnalysisConfig() *BehaviorAnalysisConfig {
	config := cm.center.Get()
	if config.BehaviorAnalysis == nil {
		return &BehaviorAnalysisConfig{}
	}
	return config.BehaviorAnalysis
}

// GetRiskScoringConfig returns the risk scoring configuration
func (cm *ConfigManager) GetRiskScoringConfig() *RiskScoringConfig {
	config := cm.center.Get()
	if config.RiskScoring == nil {
		return &RiskScoringConfig{}
	}
	return config.RiskScoring
}

// GetFeatureExtractionConfig returns the feature extraction configuration
func (cm *ConfigManager) GetFeatureExtractionConfig() *FeatureExtractionConfig {
	config := cm.center.Get()
	if config.Features == nil {
		return &FeatureExtractionConfig{}
	}
	return config.Features
}

// GetQUICConfig returns the QUIC configuration
func (cm *ConfigManager) GetQUICConfig() *QUICConfig {
	config := cm.center.Get()
	if config.QUIC == nil {
		return &QUICConfig{}
	}
	return config.QUIC
}

// GetTLSConfig returns the TLS configuration
func (cm *ConfigManager) GetTLSConfig() *TLSConfig {
	config := cm.center.Get()
	if config.TLS == nil {
		return &TLSConfig{}
	}
	return config.TLS
}

// GetGlobalConfig returns the global configuration
func (cm *ConfigManager) GetGlobalConfig() *GlobalConfig {
	config := cm.center.Get()
	if config.Global == nil {
		return &GlobalConfig{}
	}
	return config.Global
}

// UpdateBehaviorAnalysisConfig updates the behavior analysis configuration
func (cm *ConfigManager) UpdateBehaviorAnalysisConfig(newConfig *BehaviorAnalysisConfig, reason, changedBy string) error {
	config := cm.center.Get()
	config.BehaviorAnalysis = newConfig

	return cm.center.Update(config, reason, changedBy)
}

// UpdateRiskScoringConfig updates the risk scoring configuration
func (cm *ConfigManager) UpdateRiskScoringConfig(newConfig *RiskScoringConfig, reason, changedBy string) error {
	config := cm.center.Get()
	config.RiskScoring = newConfig

	return cm.center.Update(config, reason, changedBy)
}

// UpdateFeatureExtractionConfig updates the feature extraction configuration
func (cm *ConfigManager) UpdateFeatureExtractionConfig(newConfig *FeatureExtractionConfig, reason, changedBy string) error {
	config := cm.center.Get()
	config.Features = newConfig

	return cm.center.Update(config, reason, changedBy)
}

// GetConfigValue returns the configuration value at the specified path
func (cm *ConfigManager) GetConfigValue(path string) (interface{}, error) {
	// Simplified implementation - supports basic path queries
	// In practice, reflection or a JSON path library should be used

	switch path {
	case "behavior_analysis.min_requests":
		return cm.center.Get().BehaviorAnalysis.MinRequestsForAnalysis, nil
	case "behavior_analysis.regularity_threshold":
		return cm.center.Get().BehaviorAnalysis.RegularityThreshold, nil
	case "quic.min_initial_max_data":
		return cm.center.Get().QUIC.MinInitialMaxData, nil
	default:
		return nil, fmt.Errorf("unknown config path: %s", path)
	}
}

// IsLoaded checks whether the configuration has been loaded
func (cm *ConfigManager) IsLoaded() bool {
	return cm.center.IsLoaded()
}
