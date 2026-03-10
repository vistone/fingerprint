package config

import (
	"fmt"
)

// translated comment
// translated comment
type ConfigManager struct {
	center *ConfigCenter
}

// translated comment
func NewConfigManager(center *ConfigCenter) *ConfigManager {
	return &ConfigManager{
		center: center,
	}
}

// translated comment
func (cm *ConfigManager) GetBehaviorAnalysisConfig() *BehaviorAnalysisConfig {
	config := cm.center.Get()
	if config.BehaviorAnalysis == nil {
		return &BehaviorAnalysisConfig{}
	}
	return config.BehaviorAnalysis
}

// translated comment
func (cm *ConfigManager) GetRiskScoringConfig() *RiskScoringConfig {
	config := cm.center.Get()
	if config.RiskScoring == nil {
		return &RiskScoringConfig{}
	}
	return config.RiskScoring
}

// translated comment
func (cm *ConfigManager) GetFeatureExtractionConfig() *FeatureExtractionConfig {
	config := cm.center.Get()
	if config.Features == nil {
		return &FeatureExtractionConfig{}
	}
	return config.Features
}

// translated comment
func (cm *ConfigManager) GetQUICConfig() *QUICConfig {
	config := cm.center.Get()
	if config.QUIC == nil {
		return &QUICConfig{}
	}
	return config.QUIC
}

// translated comment
func (cm *ConfigManager) GetTLSConfig() *TLSConfig {
	config := cm.center.Get()
	if config.TLS == nil {
		return &TLSConfig{}
	}
	return config.TLS
}

// translated comment
func (cm *ConfigManager) GetGlobalConfig() *GlobalConfig {
	config := cm.center.Get()
	if config.Global == nil {
		return &GlobalConfig{}
	}
	return config.Global
}

// translated comment
func (cm *ConfigManager) UpdateBehaviorAnalysisConfig(newConfig *BehaviorAnalysisConfig, reason, changedBy string) error {
	config := cm.center.Get()
	config.BehaviorAnalysis = newConfig

	return cm.center.Update(config, reason, changedBy)
}

// translated comment
func (cm *ConfigManager) UpdateRiskScoringConfig(newConfig *RiskScoringConfig, reason, changedBy string) error {
	config := cm.center.Get()
	config.RiskScoring = newConfig

	return cm.center.Update(config, reason, changedBy)
}

// translated comment
func (cm *ConfigManager) UpdateFeatureExtractionConfig(newConfig *FeatureExtractionConfig, reason, changedBy string) error {
	config := cm.center.Get()
	config.Features = newConfig

	return cm.center.Update(config, reason, changedBy)
}

// translated comment
func (cm *ConfigManager) GetConfigValue(path string) (interface{}, error) {
	// translated comment
	// translated comment

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

// translated comment
func (cm *ConfigManager) IsLoaded() bool {
	return cm.center.IsLoaded()
}
