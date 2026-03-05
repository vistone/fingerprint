package config

import ic "github.com/vistone/fingerprint/modules/internal/config"

// ConfigCenter 配置中心。
type ConfigCenter = ic.ConfigCenter

// ConfigManager 配置管理器。
type ConfigManager = ic.ConfigManager

// HealthChecker 配置健康检查器。
type HealthChecker = ic.HealthChecker

// ManagedConfig 被管理配置对象。
type ManagedConfig = ic.ManagedConfig

// ConfigChange 配置变更信息。
type ConfigChange = ic.ConfigChange

// ConfigChangeListener 配置变更监听器接口。
type ConfigChangeListener = ic.ConfigChangeListener

// BehaviorAnalysisConfig 行为分析配置。
type BehaviorAnalysisConfig = ic.BehaviorAnalysisConfig

// RiskScoringConfig 风险评分配置。
type RiskScoringConfig = ic.RiskScoringConfig

// FeatureExtractionConfig 特征提取配置。
type FeatureExtractionConfig = ic.FeatureExtractionConfig

// QUICConfig QUIC 配置。
type QUICConfig = ic.QUICConfig

// TLSConfig TLS 配置。
type TLSConfig = ic.TLSConfig

// GlobalConfig 全局配置。
type GlobalConfig = ic.GlobalConfig

// InitializeConfigCenter 初始化全局配置中心。
func InitializeConfigCenter() error {
	return ic.InitializeConfigCenter()
}

// InitializeConfigCenterWithDefaults 使用默认配置初始化。
func InitializeConfigCenterWithDefaults() error {
	return ic.InitializeConfigCenterWithDefaults()
}

// GetConfigCenter 获取全局配置中心。
func GetConfigCenter() *ConfigCenter {
	return ic.GetConfigCenter()
}

// GetConfigManager 获取全局配置管理器。
func GetConfigManager() *ConfigManager {
	return ic.GetConfigManager()
}

// GetHealthChecker 获取全局健康检查器。
func GetHealthChecker() *HealthChecker {
	return ic.GetHealthChecker()
}
