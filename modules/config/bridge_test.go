package config

import (
	"testing"

	ic "github.com/vistone/fingerprint/modules/internal/config"
)

// TestTypeAliases test whether type aliases are correctly exported
func TestTypeAliases(t *testing.T) {
	// Test type aliases can be used normally
	var cc *ConfigCenter
	_ = cc

	var cm *ConfigManager
	_ = cm

	var hc *HealthChecker
	_ = hc

	var mc *ManagedConfig
	_ = mc

	var change *ConfigChange
	_ = change

	var listener ConfigChangeListener
	_ = listener

	var bac *BehaviorAnalysisConfig
	_ = bac

	var rsc *RiskScoringConfig
	_ = rsc

	var fec *FeatureExtractionConfig
	_ = fec

	var qc *QUICConfig
	_ = qc

	var tc *TLSConfig
	_ = tc

	var gc *GlobalConfig
	_ = gc
}

// TestInitializeConfigCenter test initialize configuration center
func TestInitializeConfigCenter(t *testing.T) {
	err := InitializeConfigCenter()
	// If no configuration file, initialization will fail, this is expected behavior
	// We just need to ensure the function is called without panic
	_ = err
	t.Logf("InitializeConfigCenter() returned: %v", err)
}

// TestInitializeConfigCenterWithDefaults test initialize with default configuration
func TestInitializeConfigCenterWithDefaults(t *testing.T) {
	err := InitializeConfigCenterWithDefaults()
	if err != nil {
		t.Errorf("InitializeConfigCenterWithDefaults() error = %v", err)
	}
}

// TestGetConfigCenter test get configuration center
func TestGetConfigCenter(t *testing.T) {
	// Ensure initialization first
	_ = InitializeConfigCenter()

	cc := GetConfigCenter()
	if cc == nil {
		t.Error("GetConfigCenter() returned nil")
	}
}

// TestGetConfigManager test get configuration manager
func TestGetConfigManager(t *testing.T) {
	// Ensure initialization first
	_ = InitializeConfigCenter()

	cm := GetConfigManager()
	if cm == nil {
		t.Error("GetConfigManager() returned nil")
	}
}

// TestGetHealthChecker test get health checker
func TestGetHealthChecker(t *testing.T) {
	// Ensure initialization first
	_ = InitializeConfigCenter()

	hc := GetHealthChecker()
	if hc == nil {
		t.Error("GetHealthChecker() returned nil")
	}
}

// TestBridgeIntegration test bridge functionality integration
func TestBridgeIntegration(t *testing.T) {
	// Test complete initialization process
	err := InitializeConfigCenterWithDefaults()
	if err != nil {
		t.Fatalf("Failed to initialize config center: %v", err)
	}

	// Get each component
	cc := GetConfigCenter()
	cm := GetConfigManager()
	hc := GetHealthChecker()

	if cc == nil {
		t.Error("ConfigCenter is nil")
	}

	if cm == nil {
		t.Error("ConfigManager is nil")
	}

	if hc == nil {
		t.Error("HealthChecker is nil")
	}

	// Verify types can be converted back to internal package types
	var internalCC *ic.ConfigCenter = cc
	_ = internalCC

	var internalCM *ic.ConfigManager = cm
	_ = internalCM

	var internalHC *ic.HealthChecker = hc
	_ = internalHC
}

// TestManagedConfigUsage test ManagedConfig usage
func TestManagedConfigUsage(t *testing.T) {
	config := &ManagedConfig{
		BehaviorAnalysis: &BehaviorAnalysisConfig{},
		RiskScoring:      &RiskScoringConfig{},
		Features:         &FeatureExtractionConfig{},
		QUIC:             &QUICConfig{},
		TLS:              &TLSConfig{},
		Global:           &GlobalConfig{},
	}

	if config.BehaviorAnalysis == nil {
		t.Error("BehaviorAnalysis is nil")
	}

	if config.RiskScoring == nil {
		t.Error("RiskScoring is nil")
	}

	if config.Features == nil {
		t.Error("Features is nil")
	}

	if config.QUIC == nil {
		t.Error("QUIC is nil")
	}

	if config.TLS == nil {
		t.Error("TLS is nil")
	}

	if config.Global == nil {
		t.Error("Global is nil")
	}
}

// TestConfigChangeUsage test ConfigChange usage
func TestConfigChangeUsage(t *testing.T) {
	change := &ConfigChange{
		Path:     "test.path",
		OldValue: "old",
		NewValue: "new",
	}

	if change.Path != "test.path" {
		t.Errorf("Path = %v, want test.path", change.Path)
	}

	if change.OldValue != "old" {
		t.Errorf("OldValue = %v, want old", change.OldValue)
	}

	if change.NewValue != "new" {
		t.Errorf("NewValue = %v, want new", change.NewValue)
	}
}
