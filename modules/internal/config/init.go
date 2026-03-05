package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// GlobalConfigCenter 全局配置中心实例
var GlobalConfigCenter *ConfigCenter
var GlobalConfigManager *ConfigManager
var GlobalHealthChecker *HealthChecker

// InitializeConfigCenter 初始化全局配置中心
// 支持环境变量 FINGERPRINT_CONFIG_PATH 来指定配置文件路径
func InitializeConfigCenter() error {
	// 获取配置文件路径
	configPath := os.Getenv("FINGERPRINT_CONFIG_PATH")
	if configPath == "" {
		// 默认路径
		configPath = filepath.Join("internal", "config", "config.json")
	}

	// 创建配置中心
	center := NewConfigCenter(configPath)

	// 加载配置
	if err := center.Load(); err != nil {
		return fmt.Errorf("failed to initialize config center: %w", err)
	}

	// 创建配置管理器
	manager := NewConfigManager(center)

	// 创建健康检查器
	healthChecker := NewHealthChecker(center)

	// 保存全局实例
	GlobalConfigCenter = center
	GlobalConfigManager = manager
	GlobalHealthChecker = healthChecker

	return nil
}

// InitializeConfigCenterWithDefaults 使用默认配置初始化
func InitializeConfigCenterWithDefaults() error {
	center := NewConfigCenter("")

	// 设置默认配置
	center.current = DefaultManagedConfig()
	center.loaded = true
	center.recordVersion(center.current, "initialization", "system")

	// 创建管理器和检查器
	GlobalConfigCenter = center
	GlobalConfigManager = NewConfigManager(center)
	GlobalHealthChecker = NewHealthChecker(center)

	return nil
}

// GetConfigCenter 获取全局配置中心
func GetConfigCenter() *ConfigCenter {
	if GlobalConfigCenter == nil {
		// 尝试初始化
		if err := InitializeConfigCenter(); err != nil {
			// 如果初始化失败，使用默认配置
			InitializeConfigCenterWithDefaults()
		}
	}
	return GlobalConfigCenter
}

// GetConfigManager 获取全局配置管理器
func GetConfigManager() *ConfigManager {
	if GlobalConfigManager == nil {
		GetConfigCenter()
	}
	return GlobalConfigManager
}

// GetHealthChecker 获取全局健康检查器
func GetHealthChecker() *HealthChecker {
	if GlobalHealthChecker == nil {
		GetConfigCenter()
	}
	return GlobalHealthChecker
}
