package config

import (
	"testing"
	"time"
)

// translated comment
func TestNewUnifiedConfigManager(t *testing.T) {
	ucm := NewUnifiedConfigManager("")
	if ucm == nil {
		t.Fatal("NewUnifiedConfigManager() returned nil")
	}

	if ucm.ConfigCenter == nil {
		t.Error("ConfigCenter is nil")
	}

	if ucm.enhanced != nil {
		t.Error("enhanced features should be nil by default")
	}
}

// translated comment
func TestUnifiedConfigManager_EnableEnhancedFeatures(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	// translated comment
	if ucm.enhanced != nil {
		t.Error("enhanced features should be nil by default")
	}

	// translated comment
	ucm.EnableEnhancedFeatures()

	if ucm.enhanced == nil {
		t.Fatal("enhanced features should be enabled")
	}

	if ucm.enhanced.broadcastCh == nil {
		t.Error("broadcastCh is nil")
	}

	if ucm.enhanced.subscribers == nil {
		t.Error("subscribers is nil")
	}

	if ucm.enhanced.healthChecker == nil {
		t.Error("healthChecker is nil")
	}

	// translated comment
	ucm.EnableEnhancedFeatures()
}

// translated comment
func TestUnifiedConfigManager_Subscribe(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	// translated comment
	_, err := ucm.Subscribe("test-subscriber")
	if err == nil {
		t.Error("Subscribe should fail when enhanced features not enabled")
	}

	// translated comment
	ucm.EnableEnhancedFeatures()

	// translated comment
	eventCh, err := ucm.Subscribe("test-subscriber")
	if err != nil {
		t.Errorf("Subscribe() error = %v", err)
	}

	if eventCh == nil {
		t.Error("Subscribe() returned nil channel")
	}

	// translated comment
	_, err = ucm.Subscribe("test-subscriber")
	if err == nil {
		t.Error("Subscribe should fail for duplicate subscriber")
	}

	// translated comment
	err = ucm.Unsubscribe("test-subscriber")
	if err != nil {
		t.Errorf("Unsubscribe() error = %v", err)
	}

	// translated comment
	err = ucm.Unsubscribe("non-existent")
	if err == nil {
		t.Error("Unsubscribe should fail for non-existent subscriber")
	}
}

// translated comment
func TestUnifiedConfigManager_GetBehaviorAnalysisConfig(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	// translated comment
	config := ucm.GetBehaviorAnalysisConfig()
	if config == nil {
		t.Fatal("GetBehaviorAnalysisConfig() returned nil")
	}

	// translated comment
	if config.MinRequestsForAnalysis <= 0 {
		t.Errorf("MinRequestsForAnalysis = %d, want > 0", config.MinRequestsForAnalysis)
	}

	if config.RegularityThreshold < 0 || config.RegularityThreshold > 1 {
		t.Errorf("RegularityThreshold = %f, want between 0 and 1", config.RegularityThreshold)
	}
}

// translated comment
func TestUnifiedConfigManager_GetRiskScoringConfig(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	config := ucm.GetRiskScoringConfig()
	if config == nil {
		t.Fatal("GetRiskScoringConfig() returned nil")
	}

	if config.Weights == nil {
		t.Error("Weights is nil")
	}
}

// translated comment
func TestUnifiedConfigManager_GetFeatureExtractionConfig(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	config := ucm.GetFeatureExtractionConfig()
	if config == nil {
		t.Fatal("GetFeatureExtractionConfig() returned nil")
	}

	if len(config.ToolMarkers) == 0 {
		t.Error("ToolMarkers is empty")
	}

	if len(config.HeadlessMarkers) == 0 {
		t.Error("HeadlessMarkers is empty")
	}
}

// translated comment
func TestUnifiedConfigManager_GetQUICConfig(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	config := ucm.GetQUICConfig()
	if config == nil {
		t.Fatal("GetQUICConfig() returned nil")
	}
}

// translated comment
func TestUnifiedConfigManager_GetTLSConfig(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	config := ucm.GetTLSConfig()
	if config == nil {
		t.Fatal("GetTLSConfig() returned nil")
	}
}

// translated comment
func TestUnifiedConfigManager_GetGlobalConfig(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	config := ucm.GetGlobalConfig()
	if config == nil {
		t.Fatal("GetGlobalConfig() returned nil")
	}

	if config.MaxConcurrency <= 0 {
		t.Errorf("MaxConcurrency = %d, want > 0", config.MaxConcurrency)
	}
}

// translated comment
func TestUnifiedConfigManager_UpdateBehaviorAnalysisConfig(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	newConfig := &BehaviorAnalysisConfig{
		MinRequestsForAnalysis:         10,
		RegularityThreshold:            0.5,
		EntropyThreshold:               0.6,
		AnomalousIntervalRateThreshold: 0.3,
	}

	// translated comment
	ucm.ConfigCenter.loaded = true

	err := ucm.UpdateBehaviorAnalysisConfig(newConfig, "test update", "test")
	if err != nil {
		t.Errorf("UpdateBehaviorAnalysisConfig() error = %v", err)
	}

	// translated comment
	config := ucm.GetBehaviorAnalysisConfig()
	if config.MinRequestsForAnalysis != 10 {
		t.Errorf("MinRequestsForAnalysis = %d, want 10", config.MinRequestsForAnalysis)
	}
}

// translated comment
func TestUnifiedConfigManager_GetHealthStatus(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	// translated comment
	_, err := ucm.GetHealthStatus()
	if err == nil {
		t.Error("GetHealthStatus should fail when enhanced features not enabled")
	}

	// translated comment
	ucm.EnableEnhancedFeatures()

	// translated comment
	time.Sleep(100 * time.Millisecond)

	// translated comment
	status, err := ucm.GetHealthStatus()
	if err != nil {
		t.Errorf("GetHealthStatus() error = %v", err)
	}

	// translated comment
	_ = status
}

// translated comment
func TestUnifiedConfigManager_DisableEnhancedFeatures(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	// translated comment
	ucm.DisableEnhancedFeatures()

	// translated comment
	ucm.EnableEnhancedFeatures()
	ucm.DisableEnhancedFeatures()

	if ucm.enhanced != nil {
		t.Error("enhanced features should be nil after disabling")
	}
}

// translated comment
func TestUnifiedConfigManager_WithConfigPath(t *testing.T) {
	// translated comment
	ucm := NewUnifiedConfigManager("")

	if ucm.configPath != "" {
		t.Error("configPath should be empty for memory mode")
	}

	// translated comment
	config := ucm.GetBehaviorAnalysisConfig()
	if config == nil {
		t.Error("GetBehaviorAnalysisConfig() returned nil")
	}
}

// translated comment
func TestUnifiedConfigManager_ConcurrentAccess(t *testing.T) {
	ucm := NewUnifiedConfigManager("")
	ucm.EnableEnhancedFeatures()

	// translated comment
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_ = ucm.GetBehaviorAnalysisConfig()
			_ = ucm.GetRiskScoringConfig()
			_ = ucm.GetFeatureExtractionConfig()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// OK
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent access timeout")
		}
	}
}

// translated comment
func TestCompatibility_NewConfigManager(t *testing.T) {
	center := NewConfigCenter("")
	cm := NewConfigManager(center)

	if cm == nil {
		t.Fatal("NewConfigManager() returned nil")
	}

	if cm.center != center {
		t.Error("center not set correctly")
	}

	// translated comment
	config := cm.GetBehaviorAnalysisConfig()
	if config == nil {
		t.Error("GetBehaviorAnalysisConfig() returned nil")
	}
}

// translated comment
func TestCompatibility_WrapConfigCenter(t *testing.T) {
	center := NewConfigCenter("")
	enhanced := WrapConfigCenter(center)

	if enhanced == nil {
		t.Fatal("WrapConfigCenter() returned nil")
	}

	if enhanced.center != center {
		t.Error("center not set correctly")
	}

	// translated comment
	_, err := enhanced.Subscribe("test")
	if err != nil {
		t.Errorf("Subscribe() error = %v", err)
	}
}
