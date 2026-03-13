package config

import (
	"fmt"
	"sync"
	"time"
)

// UnifiedConfigManager is the unified configuration manager
// Combines the functionality of ConfigCenter, ConfigManager, and EnhancedConfigCenter
type UnifiedConfigManager struct {
	// Core configuration center (embedded for direct access)
	*ConfigCenter

	// Optional enhanced features
	enhanced *enhancedFeatures

	// Mutex to protect the enhanced field
	mu sync.RWMutex
}

// enhancedFeatures represents optional enhanced features
type enhancedFeatures struct {
	broadcastCh   chan ConfigChangeEvent
	subscribers   map[string]chan ConfigChangeEvent
	subscriberMu  sync.RWMutex
	healthChecker *ConfigHealthChecker
}

// NewUnifiedConfigManager creates a new unified configuration manager
func NewUnifiedConfigManager(configPath string) *UnifiedConfigManager {
	return &UnifiedConfigManager{
		ConfigCenter: NewConfigCenter(configPath),
	}
}

// EnableEnhancedFeatures enables enhanced features (event subscription, health check)
func (ucm *UnifiedConfigManager) EnableEnhancedFeatures() {
	if ucm.enhanced != nil {
		return // Already enabled
	}

	ucm.enhanced = &enhancedFeatures{
		broadcastCh: make(chan ConfigChangeEvent, 100),
		subscribers: make(map[string]chan ConfigChangeEvent),
	}

	// Initialize the health checker
	ucm.enhanced.healthChecker = &ConfigHealthChecker{
		center:     ucm.ConfigCenter,
		checkFuncs: []HealthCheckFunc{defaultHealthCheck},
		interval:   30 * time.Second,
		stopCh:     make(chan struct{}),
		lastStatus: ConfigHealthStatus{
			Healthy:       true,
			LastCheckTime: time.Now(),
		},
	}

	// Start the broadcast processor
	go ucm.broadcastProcessor()

	// Start the health checker
	go ucm.enhanced.healthChecker.start(ucm.enhanced.broadcastCh)
}

// DisableEnhancedFeatures disables enhanced features
func (ucm *UnifiedConfigManager) DisableEnhancedFeatures() {
	ucm.mu.Lock()
	defer ucm.mu.Unlock()

	if ucm.enhanced == nil {
		return
	}

	// Stop the health checker
	if ucm.enhanced.healthChecker != nil {
		ucm.enhanced.healthChecker.stop()
	}

	// Close the broadcast channel (if not already closed)
	if ucm.enhanced.broadcastCh != nil {
		select {
		case <-ucm.enhanced.broadcastCh:
			// Already closed
		default:
			close(ucm.enhanced.broadcastCh)
		}
	}

	// Close all subscriber channels
	ucm.enhanced.subscriberMu.Lock()
	for _, ch := range ucm.enhanced.subscribers {
		select {
		case <-ch:
			// Already closed
		default:
			close(ch)
		}
	}
	ucm.enhanced.subscribers = make(map[string]chan ConfigChangeEvent)
	ucm.enhanced.subscriberMu.Unlock()

	ucm.enhanced = nil
}

// Subscribe subscribes to configuration change events (requires enhanced features to be enabled)
func (ucm *UnifiedConfigManager) Subscribe(subscriberID string) (<-chan ConfigChangeEvent, error) {
	if ucm.enhanced == nil {
		return nil, fmt.Errorf("enhanced features not enabled, call EnableEnhancedFeatures() first")
	}

	ucm.enhanced.subscriberMu.Lock()
	defer ucm.enhanced.subscriberMu.Unlock()

	if _, exists := ucm.enhanced.subscribers[subscriberID]; exists {
		return nil, fmt.Errorf("subscriber %s already exists", subscriberID)
	}

	eventCh := make(chan ConfigChangeEvent, 10)
	ucm.enhanced.subscribers[subscriberID] = eventCh

	// Send subscription confirmation event
	ucm.enhanced.broadcastCh <- ConfigChangeEvent{
		Type:        ConfigChangeTypeSubscribe,
		Timestamp:   time.Now(),
		Description: fmt.Sprintf("Subscriber %s registered", subscriberID),
		Source:      subscriberID,
	}

	return eventCh, nil
}

// Unsubscribe unsubscribes from configuration change events
func (ucm *UnifiedConfigManager) Unsubscribe(subscriberID string) error {
	if ucm.enhanced == nil {
		return fmt.Errorf("enhanced features not enabled")
	}

	ucm.enhanced.subscriberMu.Lock()
	defer ucm.enhanced.subscriberMu.Unlock()

	eventCh, exists := ucm.enhanced.subscribers[subscriberID]
	if !exists {
		return fmt.Errorf("subscriber %s not found", subscriberID)
	}

	close(eventCh)
	delete(ucm.enhanced.subscribers, subscriberID)

	return nil
}

// broadcastProcessor is the broadcast event processor
func (ucm *UnifiedConfigManager) broadcastProcessor() {
	for {
		// Check whether enhanced features exist
		ucm.mu.RLock()
		enhanced := ucm.enhanced
		ucm.mu.RUnlock()

		if enhanced == nil {
			return
		}

		// Safely read events from the broadcast channel
		event, ok := <-enhanced.broadcastCh
		if !ok {
			// Channel is closed, exit
			return
		}

		// Copy subscriber list (to avoid holding the lock while sending)
		enhanced.subscriberMu.RLock()
		subscribers := make(map[string]chan ConfigChangeEvent, len(enhanced.subscribers))
		for k, v := range enhanced.subscribers {
			subscribers[k] = v
		}
		enhanced.subscriberMu.RUnlock()

		// Asynchronously send to all subscribers
		for subscriberID, ch := range subscribers {
			select {
			case ch <- event:
				// Sent successfully
			default:
				// Channel is full or closed, skip
				_ = subscriberID
			}
		}
	}
}

// Update overrides the update method to add event broadcasting
func (ucm *UnifiedConfigManager) Update(newConfig *ManagedConfig, reason, changedBy string) error {
	// Call the base method
	if err := ucm.ConfigCenter.Update(newConfig, reason, changedBy); err != nil {
		return err
	}

	// If enhanced features are enabled, broadcast the update event
	if ucm.enhanced != nil {
		ucm.enhanced.broadcastCh <- ConfigChangeEvent{
			Type:        ConfigChangeTypeUpdate,
			Timestamp:   time.Now(),
			Config:      ucm.ConfigCenter.Get(),
			Description: reason,
			Source:      changedBy,
		}
	}

	return nil
}

// GetHealthStatus returns the health status
func (ucm *UnifiedConfigManager) GetHealthStatus() (ConfigHealthStatus, error) {
	if ucm.enhanced == nil {
		return ConfigHealthStatus{}, fmt.Errorf("enhanced features not enabled")
	}

	ucm.enhanced.healthChecker.mu.RLock()
	defer ucm.enhanced.healthChecker.mu.RUnlock()
	return ucm.enhanced.healthChecker.lastStatus, nil
}

// AddHealthCheck adds a health check function
func (ucm *UnifiedConfigManager) AddHealthCheck(checkFunc HealthCheckFunc) error {
	if ucm.enhanced == nil {
		return fmt.Errorf("enhanced features not enabled")
	}

	ucm.enhanced.healthChecker.mu.Lock()
	defer ucm.enhanced.healthChecker.mu.Unlock()
	ucm.enhanced.healthChecker.checkFuncs = append(ucm.enhanced.healthChecker.checkFuncs, checkFunc)
	return nil
}

// ============================================
// Convenient configuration getter methods (original ConfigManager functionality)
// ============================================

// GetBehaviorAnalysisConfig returns the behavior analysis configuration
func (ucm *UnifiedConfigManager) GetBehaviorAnalysisConfig() *BehaviorAnalysisConfig {
	config := ucm.ConfigCenter.Get()
	if config.BehaviorAnalysis == nil {
		return &BehaviorAnalysisConfig{
			MinRequestsForAnalysis:         5,
			RegularityThreshold:            0.3,
			EntropyThreshold:               0.5,
			AnomalousIntervalRateThreshold: 0.2,
			RequestHistoryCapacity:         100,
			SignalCapacity:                 50,
		}
	}
	return config.BehaviorAnalysis
}

// GetRiskScoringConfig returns the risk scoring configuration
func (ucm *UnifiedConfigManager) GetRiskScoringConfig() *RiskScoringConfig {
	config := ucm.ConfigCenter.Get()
	if config.RiskScoring == nil {
		return &RiskScoringConfig{
			CriticalThreshold: 0.9,
			HighThreshold:     0.7,
			MediumThreshold:   0.5,
			LowThreshold:      0.3,
			MinConfidence:     0.6,
			Weights: &RiskWeights{
				Headless:         0.3,
				Anomaly:          0.2,
				Contradiction:    0.2,
				ECH:              0.1,
				ClientHints:      0.1,
				BehaviorAnomaly:  0.2,
				CipherSuiteRisk:  0.1,
				ExtensionAnomaly: 0.1,
			},
		}
	}
	return config.RiskScoring
}

// GetFeatureExtractionConfig returns the feature extraction configuration
func (ucm *UnifiedConfigManager) GetFeatureExtractionConfig() *FeatureExtractionConfig {
	config := ucm.ConfigCenter.Get()
	if config.Features == nil {
		return &FeatureExtractionConfig{
			EntropyHighThreshold:  7.5,
			EntropyLowThreshold:   26,
			ToolMarkers:           []string{"HeadlessChrome", "PhantomJS", "webdriver", "selenium", "puppeteer"},
			HeadlessMarkers:       []string{"headlesschrome", "phantomjs", "selenium", "webdriver", "puppeteer", "playwright", "cypress", "jsdom", "zombie", "htmlunit"},
			MobileScreenWidthMax:  1920,
			DesktopScreenWidthMin: 800,
		}
	}
	return config.Features
}

// GetQUICConfig returns the QUIC configuration
func (ucm *UnifiedConfigManager) GetQUICConfig() *QUICConfig {
	config := ucm.ConfigCenter.Get()
	if config.QUIC == nil {
		return &QUICConfig{
			MinInitialMaxData:      1000000,
			MinStreamData:          100000,
			SupportedVersions:      []uint32{0x00000001, 0x00000002},
			TransportParamCapacity: 16,
		}
	}
	return config.QUIC
}

// GetTLSConfig returns the TLS configuration
func (ucm *UnifiedConfigManager) GetTLSConfig() *TLSConfig {
	config := ucm.ConfigCenter.Get()
	if config.TLS == nil {
		return &TLSConfig{
			WeakCipherSuites:     []uint16{0x002f, 0x0035, 0x000a},
			SupportedVersions:    []uint16{0x0301, 0x0302, 0x0303, 0x0304},
			GREASEExtensions:     []uint16{0x0a0a, 0x1a1a, 0x2a2a},
			AnomalyFlagsCapacity: 32,
		}
	}
	return config.TLS
}

// GetGlobalConfig returns the global configuration
func (ucm *UnifiedConfigManager) GetGlobalConfig() *GlobalConfig {
	config := ucm.ConfigCenter.Get()
	if config.Global == nil {
		return &GlobalConfig{
			MaxConcurrency: 100,
			RequestTimeout: 5000,
			CacheSize:      1000,
			DebugMode:      false,
			MaxInputSize:   1024 * 1024,
		}
	}
	return config.Global
}

// ============================================
// Convenient configuration update methods
// ============================================

// UpdateBehaviorAnalysisConfig updates the behavior analysis configuration
func (ucm *UnifiedConfigManager) UpdateBehaviorAnalysisConfig(newConfig *BehaviorAnalysisConfig, reason, changedBy string) error {
	config := ucm.ConfigCenter.Get()
	config.BehaviorAnalysis = newConfig
	return ucm.Update(config, reason, changedBy)
}

// UpdateRiskScoringConfig updates the risk scoring configuration
func (ucm *UnifiedConfigManager) UpdateRiskScoringConfig(newConfig *RiskScoringConfig, reason, changedBy string) error {
	config := ucm.ConfigCenter.Get()
	config.RiskScoring = newConfig
	return ucm.Update(config, reason, changedBy)
}

// UpdateFeatureExtractionConfig updates the feature extraction configuration
func (ucm *UnifiedConfigManager) UpdateFeatureExtractionConfig(newConfig *FeatureExtractionConfig, reason, changedBy string) error {
	config := ucm.ConfigCenter.Get()
	config.Features = newConfig
	return ucm.Update(config, reason, changedBy)
}

// ============================================
// Compatibility adaptation layer
// ============================================
