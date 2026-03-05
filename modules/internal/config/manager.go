package config

import (
	"fmt"
)

// ConfigManager 配置管理器 - 提供统一的配置访问接口
// 注意：ConfigManager 本身不需要锁，因为 ConfigCenter 已经有完整的锁机制
type ConfigManager struct {
	center *ConfigCenter
}

// NewConfigManager 创建配置管理器
func NewConfigManager(center *ConfigCenter) *ConfigManager {
	return &ConfigManager{
		center: center,
	}
}

// GetBehaviorAnalysisConfig 获取行为分析配置
func (cm *ConfigManager) GetBehaviorAnalysisConfig() *BehaviorAnalysisConfig {
	config := cm.center.Get()
	if config.BehaviorAnalysis == nil {
		return &BehaviorAnalysisConfig{}
	}
	return config.BehaviorAnalysis
}

// GetRiskScoringConfig 获取风险评分配置
func (cm *ConfigManager) GetRiskScoringConfig() *RiskScoringConfig {
	config := cm.center.Get()
	if config.RiskScoring == nil {
		return &RiskScoringConfig{}
	}
	return config.RiskScoring
}

// GetFeatureExtractionConfig 获取特征提取配置
func (cm *ConfigManager) GetFeatureExtractionConfig() *FeatureExtractionConfig {
	config := cm.center.Get()
	if config.Features == nil {
		return &FeatureExtractionConfig{}
	}
	return config.Features
}

// GetQUICConfig 获取 QUIC 配置
func (cm *ConfigManager) GetQUICConfig() *QUICConfig {
	config := cm.center.Get()
	if config.QUIC == nil {
		return &QUICConfig{}
	}
	return config.QUIC
}

// GetTLSConfig 获取 TLS 配置
func (cm *ConfigManager) GetTLSConfig() *TLSConfig {
	config := cm.center.Get()
	if config.TLS == nil {
		return &TLSConfig{}
	}
	return config.TLS
}

// GetGlobalConfig 获取全局配置
func (cm *ConfigManager) GetGlobalConfig() *GlobalConfig {
	config := cm.center.Get()
	if config.Global == nil {
		return &GlobalConfig{}
	}
	return config.Global
}

// UpdateBehaviorAnalysisConfig 更新行为分析配置
func (cm *ConfigManager) UpdateBehaviorAnalysisConfig(newConfig *BehaviorAnalysisConfig, reason, changedBy string) error {
	config := cm.center.Get()
	config.BehaviorAnalysis = newConfig

	return cm.center.Update(config, reason, changedBy)
}

// UpdateRiskScoringConfig 更新风险评分配置
func (cm *ConfigManager) UpdateRiskScoringConfig(newConfig *RiskScoringConfig, reason, changedBy string) error {
	config := cm.center.Get()
	config.RiskScoring = newConfig

	return cm.center.Update(config, reason, changedBy)
}

// UpdateFeatureExtractionConfig 更新特征提取配置
func (cm *ConfigManager) UpdateFeatureExtractionConfig(newConfig *FeatureExtractionConfig, reason, changedBy string) error {
	config := cm.center.Get()
	config.Features = newConfig

	return cm.center.Update(config, reason, changedBy)
}

// GetConfigValue 获取指定路径的配置值
func (cm *ConfigManager) GetConfigValue(path string) (interface{}, error) {
	// 简化实现 - 支持基本的路径查询
	// 实际应该使用反射或 JSON 路径库

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

// IsLoaded 检查配置是否已加载
func (cm *ConfigManager) IsLoaded() bool {
	return cm.center.IsLoaded()
}
