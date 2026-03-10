package fingerprint_test

import (
	"testing"

	"github.com/vistone/fingerprint/modules/internal/config"
)

func TestConfigCenterBasic(t *testing.T) {
	// 使用默认configurationinitialize
	if err := config.InitializeConfigCenterWithDefaults(); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	center := config.GetConfigCenter()
	if center == nil {
		t.Fatal("Config center should not be nil")
	}

	if !center.IsLoaded() {
		t.Fatal("Config center should be loaded")
	}

	cfg := center.Get()
	if cfg == nil {
		t.Fatal("Config should not be nil")
	}
}

func TestConfigManager(t *testing.T) {
	if err := config.InitializeConfigCenterWithDefaults(); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	manager := config.GetConfigManager()
	if manager == nil {
		t.Fatal("Config manager should not be nil")
	}

	behaviorCfg := manager.GetBehaviorAnalysisConfig()
	if behaviorCfg.MinRequestsForAnalysis != 5 {
		t.Fatalf("Expected MinRequestsForAnalysis=5, got %d", behaviorCfg.MinRequestsForAnalysis)
	}
}

func TestConfigValidator(t *testing.T) {
	validator := config.NewConfigValidator()
	cfg := config.DefaultManagedConfig()

	errs := validator.Validate(cfg)
	if len(errs) > 0 {
		t.Fatalf("Valid config should not have errors, got %d", len(errs))
	}
}

func TestHealthChecker(t *testing.T) {
	if err := config.InitializeConfigCenterWithDefaults(); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	healthChecker := config.GetHealthChecker()
	if healthChecker == nil {
		t.Fatal("Health checker should not be nil")
	}

	result := healthChecker.CheckHealth()
	if result.Status == config.HealthCritical {
		t.Fatal("Valid config should not be critical")
	}

	if !result.Overall {
		t.Fatal("Valid config should have overall health = true")
	}
}
