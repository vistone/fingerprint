package config

import (
	"sync"
	"testing"
	"time"
)

// TestNewUnifiedConfigManager tests creating a unified configuration manager
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

// TestUnifiedConfigManager_EnableEnhancedFeatures tests enabling enhanced features
func TestUnifiedConfigManager_EnableEnhancedFeatures(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	// Not enabled by default
	if ucm.enhanced != nil {
		t.Error("enhanced features should be nil by default")
	}

	// Enable enhanced features
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

	// Repeated enabling should not cause errors
	ucm.EnableEnhancedFeatures()
}

// TestUnifiedConfigManager_Subscribe tests the subscription functionality
func TestUnifiedConfigManager_Subscribe(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	// Subscribing should fail when enhanced features are not enabled
	_, err := ucm.Subscribe("test-subscriber")
	if err == nil {
		t.Error("Subscribe should fail when enhanced features not enabled")
	}

	// Enable enhanced features
	ucm.EnableEnhancedFeatures()

	// Subscription should succeed
	eventCh, err := ucm.Subscribe("test-subscriber")
	if err != nil {
		t.Errorf("Subscribe() error = %v", err)
	}

	if eventCh == nil {
		t.Error("Subscribe() returned nil channel")
	}

	// Duplicate subscription should fail
	_, err = ucm.Subscribe("test-subscriber")
	if err == nil {
		t.Error("Subscribe should fail for duplicate subscriber")
	}

	// Unsubscribe
	err = ucm.Unsubscribe("test-subscriber")
	if err != nil {
		t.Errorf("Unsubscribe() error = %v", err)
	}

	// Unsubscribing a non-existent subscriber should fail
	err = ucm.Unsubscribe("non-existent")
	if err == nil {
		t.Error("Unsubscribe should fail for non-existent subscriber")
	}
}

// TestUnifiedConfigManager_GetBehaviorAnalysisConfig tests getting the behavior analysis configuration
func TestUnifiedConfigManager_GetBehaviorAnalysisConfig(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	// Get the default configuration
	config := ucm.GetBehaviorAnalysisConfig()
	if config == nil {
		t.Fatal("GetBehaviorAnalysisConfig() returned nil")
	}

	// Verify default values
	if config.MinRequestsForAnalysis <= 0 {
		t.Errorf("MinRequestsForAnalysis = %d, want > 0", config.MinRequestsForAnalysis)
	}

	if config.RegularityThreshold < 0 || config.RegularityThreshold > 1 {
		t.Errorf("RegularityThreshold = %f, want between 0 and 1", config.RegularityThreshold)
	}
}

// TestUnifiedConfigManager_GetRiskScoringConfig tests getting the risk scoring configuration
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

// TestUnifiedConfigManager_GetFeatureExtractionConfig tests getting the feature extraction configuration
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

// TestUnifiedConfigManager_GetQUICConfig tests getting the QUIC configuration
func TestUnifiedConfigManager_GetQUICConfig(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	config := ucm.GetQUICConfig()
	if config == nil {
		t.Fatal("GetQUICConfig() returned nil")
	}
}

// TestUnifiedConfigManager_GetTLSConfig tests getting the TLS configuration
func TestUnifiedConfigManager_GetTLSConfig(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	config := ucm.GetTLSConfig()
	if config == nil {
		t.Fatal("GetTLSConfig() returned nil")
	}
}

// TestUnifiedConfigManager_GetGlobalConfig tests getting the global configuration
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

// TestUnifiedConfigManager_UpdateBehaviorAnalysisConfig tests updating the behavior analysis configuration
func TestUnifiedConfigManager_UpdateBehaviorAnalysisConfig(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	newConfig := &BehaviorAnalysisConfig{
		MinRequestsForAnalysis:         10,
		RegularityThreshold:            0.5,
		EntropyThreshold:               0.6,
		AnomalousIntervalRateThreshold: 0.3,
	}

	// Need to set loaded to true before updating
	ucm.ConfigCenter.loaded = true

	err := ucm.UpdateBehaviorAnalysisConfig(newConfig, "test update", "test")
	if err != nil {
		t.Errorf("UpdateBehaviorAnalysisConfig() error = %v", err)
	}

	// Verify the updated configuration
	config := ucm.GetBehaviorAnalysisConfig()
	if config.MinRequestsForAnalysis != 10 {
		t.Errorf("MinRequestsForAnalysis = %d, want 10", config.MinRequestsForAnalysis)
	}
}

// TestUnifiedConfigManager_GetHealthStatus tests getting the health status
func TestUnifiedConfigManager_GetHealthStatus(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	// Should fail when enhanced features are not enabled
	_, err := ucm.GetHealthStatus()
	if err == nil {
		t.Error("GetHealthStatus should fail when enhanced features not enabled")
	}

	// Enable enhanced features
	ucm.EnableEnhancedFeatures()

	// Give the health checker a moment to start
	time.Sleep(100 * time.Millisecond)

	// Should be able to get the health status
	status, err := ucm.GetHealthStatus()
	if err != nil {
		t.Errorf("GetHealthStatus() error = %v", err)
	}

	// Health status of an empty configuration
	_ = status
}

// TestUnifiedConfigManager_DisableEnhancedFeatures tests disabling enhanced features
func TestUnifiedConfigManager_DisableEnhancedFeatures(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	// Disabling features that are not enabled should not cause errors
	ucm.DisableEnhancedFeatures()

	// Enable and then disable
	ucm.EnableEnhancedFeatures()
	ucm.DisableEnhancedFeatures()

	if ucm.enhanced != nil {
		t.Error("enhanced features should be nil after disabling")
	}
}

func TestUnifiedConfigManager_UpdateWhileDisablingDoesNotPanic(t *testing.T) {
	for i := 0; i < 100; i++ {
		ucm := NewUnifiedConfigManager("")
		ucm.current = DefaultManagedConfig()
		ucm.loaded = true
		ucm.EnableEnhancedFeatures()

		panicCh := make(chan interface{}, 2)
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					panicCh <- r
				}
			}()
			_ = ucm.Update(DefaultManagedConfig(), "test update", "tester")
		}()

		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					panicCh <- r
				}
			}()
			ucm.DisableEnhancedFeatures()
		}()

		wg.Wait()

		select {
		case p := <-panicCh:
			t.Fatalf("unexpected panic while disabling during update: %v", p)
		default:
		}
	}
}

func TestUnifiedConfigManager_UnsubscribeWhileBroadcastingDoesNotPanic(t *testing.T) {
	for i := 0; i < 50; i++ {
		ucm := NewUnifiedConfigManager("")
		ucm.current = DefaultManagedConfig()
		ucm.loaded = true
		ucm.EnableEnhancedFeatures()

		ch, err := ucm.Subscribe("test-subscriber")
		if err != nil {
			t.Fatalf("Subscribe() error = %v", err)
		}

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				_ = ucm.Update(DefaultManagedConfig(), "test update", "tester")
			}
		}()

		go func() {
			defer wg.Done()
			<-ch
			_ = ucm.Unsubscribe("test-subscriber")
		}()

		wg.Wait()
		ucm.DisableEnhancedFeatures()
	}
}

func TestUnifiedConfigManager_DisableClosesBroadcastChannelEvenWhenBuffered(t *testing.T) {
	ucm := NewUnifiedConfigManager("")
	ucm.EnableEnhancedFeatures()

	enhanced := ucm.enhanced
	enhanced.broadcastCh <- ConfigChangeEvent{
		Type:        ConfigChangeTypeUpdate,
		Timestamp:   time.Now(),
		Description: "buffered event",
		Source:      "test",
	}

	ucm.DisableEnhancedFeatures()

	select {
	case _, ok := <-enhanced.broadcastCh:
		if ok {
			select {
			case _, ok = <-enhanced.broadcastCh:
				if ok {
					t.Fatal("expected broadcast channel to be closed after draining buffered event")
				}
			case <-time.After(100 * time.Millisecond):
				t.Fatal("broadcast channel remained open after draining buffered event")
			}
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("broadcast channel remained open after disable")
	}
}

// TestUnifiedConfigManager_WithConfigPath tests the manager with a config path
func TestUnifiedConfigManager_WithConfigPath(t *testing.T) {
	// Use an empty path (memory mode)
	ucm := NewUnifiedConfigManager("")

	if ucm.configPath != "" {
		t.Error("configPath should be empty for memory mode")
	}

	// Verify normal operation
	config := ucm.GetBehaviorAnalysisConfig()
	if config == nil {
		t.Error("GetBehaviorAnalysisConfig() returned nil")
	}
}

// TestUnifiedConfigManager_ConcurrentAccess tests concurrent access
func TestUnifiedConfigManager_ConcurrentAccess(t *testing.T) {
	ucm := NewUnifiedConfigManager("")
	ucm.EnableEnhancedFeatures()

	// Concurrent reads
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

// TestCompatibility_NewConfigManager tests the compatibility of NewConfigManager
func TestCompatibility_NewConfigManager(t *testing.T) {
	center := NewConfigCenter("")
	cm := NewConfigManager(center)

	if cm == nil {
		t.Fatal("NewConfigManager() returned nil")
	}

	if cm.center != center {
		t.Error("center not set correctly")
	}

	// Verify basic ConfigManager functionality
	config := cm.GetBehaviorAnalysisConfig()
	if config == nil {
		t.Error("GetBehaviorAnalysisConfig() returned nil")
	}
}

// TestCompatibility_WrapConfigCenter tests the compatibility of WrapConfigCenter
func TestCompatibility_WrapConfigCenter(t *testing.T) {
	center := NewConfigCenter("")
	enhanced := WrapConfigCenter(center)

	if enhanced == nil {
		t.Fatal("WrapConfigCenter() returned nil")
	}

	if enhanced.center != center {
		t.Error("center not set correctly")
	}

	// Verify that enhanced features are enabled
	_, err := enhanced.Subscribe("test")
	if err != nil {
		t.Errorf("Subscribe() error = %v", err)
	}
}
