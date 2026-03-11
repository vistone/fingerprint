package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// GlobalConfigCenter is the global configuration center instance
var GlobalConfigCenter *ConfigCenter
var GlobalConfigManager *ConfigManager
var GlobalHealthChecker *HealthChecker

// InitializeConfigCenter initializes the global configuration center
// Supports the FINGERPRINT_CONFIG_PATH environment variable to specify the configuration file path
func InitializeConfigCenter() error {
	// Get the configuration file path
	configPath := os.Getenv("FINGERPRINT_CONFIG_PATH")
	if configPath == "" {
		// Default path
		configPath = filepath.Join("internal", "config", "config.json")
	}

	// Create the configuration center
	center := NewConfigCenter(configPath)

	// Load the configuration
	if err := center.Load(); err != nil {
		return fmt.Errorf("failed to initialize config center: %w", err)
	}

	// Create the configuration manager
	manager := NewConfigManager(center)

	// Create the health checker
	healthChecker := NewHealthChecker(center)

	// Save global instances
	GlobalConfigCenter = center
	GlobalConfigManager = manager
	GlobalHealthChecker = healthChecker

	return nil
}

// InitializeConfigCenterWithDefaults initializes with default configuration
func InitializeConfigCenterWithDefaults() error {
	center := NewConfigCenter("")

	// Set default configuration
	center.current = DefaultManagedConfig()
	center.loaded = true
	center.recordVersion(center.current, "initialization", "system")

	// Create the manager and checker
	GlobalConfigCenter = center
	GlobalConfigManager = NewConfigManager(center)
	GlobalHealthChecker = NewHealthChecker(center)

	return nil
}

// GetConfigCenter returns the global configuration center
func GetConfigCenter() *ConfigCenter {
	if GlobalConfigCenter == nil {
		// Try to initialize
		if err := InitializeConfigCenter(); err != nil {
			// If initialization fails, use default configuration
			InitializeConfigCenterWithDefaults()
		}
	}
	return GlobalConfigCenter
}

// GetConfigManager returns the global configuration manager
func GetConfigManager() *ConfigManager {
	if GlobalConfigManager == nil {
		GetConfigCenter()
	}
	return GlobalConfigManager
}

// GetHealthChecker returns the global health checker
func GetHealthChecker() *HealthChecker {
	if GlobalHealthChecker == nil {
		GetConfigCenter()
	}
	return GlobalHealthChecker
}
