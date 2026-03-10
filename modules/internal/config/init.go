package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// translated comment
var GlobalConfigCenter *ConfigCenter
var GlobalConfigManager *ConfigManager
var GlobalHealthChecker *HealthChecker

// translated comment
// translated comment
func InitializeConfigCenter() error {
	// translated comment
	configPath := os.Getenv("FINGERPRINT_CONFIG_PATH")
	if configPath == "" {
		// translated comment
		configPath = filepath.Join("internal", "config", "config.json")
	}

	// translated comment
	center := NewConfigCenter(configPath)

	// translated comment
	if err := center.Load(); err != nil {
		return fmt.Errorf("failed to initialize config center: %w", err)
	}

	// translated comment
	manager := NewConfigManager(center)

	// translated comment
	healthChecker := NewHealthChecker(center)

	// translated comment
	GlobalConfigCenter = center
	GlobalConfigManager = manager
	GlobalHealthChecker = healthChecker

	return nil
}

// translated comment
func InitializeConfigCenterWithDefaults() error {
	center := NewConfigCenter("")

	// translated comment
	center.current = DefaultManagedConfig()
	center.loaded = true
	center.recordVersion(center.current, "initialization", "system")

	// translated comment
	GlobalConfigCenter = center
	GlobalConfigManager = NewConfigManager(center)
	GlobalHealthChecker = NewHealthChecker(center)

	return nil
}

// translated comment
func GetConfigCenter() *ConfigCenter {
	if GlobalConfigCenter == nil {
		// translated comment
		if err := InitializeConfigCenter(); err != nil {
			// translated comment
			InitializeConfigCenterWithDefaults()
		}
	}
	return GlobalConfigCenter
}

// translated comment
func GetConfigManager() *ConfigManager {
	if GlobalConfigManager == nil {
		GetConfigCenter()
	}
	return GlobalConfigManager
}

// translated comment
func GetHealthChecker() *HealthChecker {
	if GlobalHealthChecker == nil {
		GetConfigCenter()
	}
	return GlobalHealthChecker
}
