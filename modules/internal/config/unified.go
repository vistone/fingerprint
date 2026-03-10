package config

import (
	"fmt"
	"sync"
	"time"
)

// translated comment
// translated comment
type UnifiedConfigManager struct {
	// translated comment
	*ConfigCenter

	// translated comment
	enhanced *enhancedFeatures

	// translated comment
	mu sync.RWMutex
}

// translated comment
type enhancedFeatures struct {
	broadcastCh   chan ConfigChangeEvent
	subscribers   map[string]chan ConfigChangeEvent
	subscriberMu  sync.RWMutex
	healthChecker *ConfigHealthChecker
}

// translated comment
func NewUnifiedConfigManager(configPath string) *UnifiedConfigManager {
	return &UnifiedConfigManager{
		ConfigCenter: NewConfigCenter(configPath),
	}
}

// translated comment
func (ucm *UnifiedConfigManager) EnableEnhancedFeatures() {
	if ucm.enhanced != nil {
		return // translated comment
	}

	ucm.enhanced = &enhancedFeatures{
		broadcastCh: make(chan ConfigChangeEvent, 100),
		subscribers: make(map[string]chan ConfigChangeEvent),
	}

	// translated comment
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

	// translated comment
	go ucm.broadcastProcessor()

	// translated comment
	go ucm.enhanced.healthChecker.start(ucm.enhanced.broadcastCh)
}

// translated comment
func (ucm *UnifiedConfigManager) DisableEnhancedFeatures() {
	ucm.mu.Lock()
	defer ucm.mu.Unlock()
	
	if ucm.enhanced == nil {
		return
	}

	// translated comment
	if ucm.enhanced.healthChecker != nil {
		ucm.enhanced.healthChecker.stop()
	}

	// translated comment
	if ucm.enhanced.broadcastCh != nil {
		select {
		case <-ucm.enhanced.broadcastCh:
			// translated comment
		default:
			close(ucm.enhanced.broadcastCh)
		}
	}

	// translated comment
	ucm.enhanced.subscriberMu.Lock()
	for _, ch := range ucm.enhanced.subscribers {
		select {
		case <-ch:
			// translated comment
		default:
			close(ch)
		}
	}
	ucm.enhanced.subscribers = make(map[string]chan ConfigChangeEvent)
	ucm.enhanced.subscriberMu.Unlock()

	ucm.enhanced = nil
}

// translated comment
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

	// translated comment
	ucm.enhanced.broadcastCh <- ConfigChangeEvent{
		Type:        ConfigChangeTypeSubscribe,
		Timestamp:   time.Now(),
		Description: fmt.Sprintf("Subscriber %s registered", subscriberID),
		Source:      subscriberID,
	}

	return eventCh, nil
}

// translated comment
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

// translated comment
func (ucm *UnifiedConfigManager) broadcastProcessor() {
	for {
		// translated comment
		ucm.mu.RLock()
		enhanced := ucm.enhanced
		ucm.mu.RUnlock()
		
		if enhanced == nil {
			return
		}
		
		// translated comment
		event, ok := <-enhanced.broadcastCh
		if !ok {
			// translated comment
			return
		}
		
		// translated comment
		enhanced.subscriberMu.RLock()
		subscribers := make(map[string]chan ConfigChangeEvent, len(enhanced.subscribers))
		for k, v := range enhanced.subscribers {
			subscribers[k] = v
		}
		enhanced.subscriberMu.RUnlock()

		// translated comment
		for subscriberID, ch := range subscribers {
			select {
			case ch <- event:
				// translated comment
			default:
				// translated comment
				_ = subscriberID
			}
		}
	}
}

// translated comment
func (ucm *UnifiedConfigManager) Update(newConfig *ManagedConfig, reason, changedBy string) error {
	// translated comment
	if err := ucm.ConfigCenter.Update(newConfig, reason, changedBy); err != nil {
		return err
	}

	// translated comment
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

// translated comment
func (ucm *UnifiedConfigManager) GetHealthStatus() (ConfigHealthStatus, error) {
	if ucm.enhanced == nil {
		return ConfigHealthStatus{}, fmt.Errorf("enhanced features not enabled")
	}

	ucm.enhanced.healthChecker.mu.RLock()
	defer ucm.enhanced.healthChecker.mu.RUnlock()
	return ucm.enhanced.healthChecker.lastStatus, nil
}

// translated comment
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
// translated comment
// ============================================

// translated comment
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

// translated comment
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

// translated comment
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

// translated comment
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

// translated comment
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

// translated comment
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
// translated comment
// ============================================

// translated comment
func (ucm *UnifiedConfigManager) UpdateBehaviorAnalysisConfig(newConfig *BehaviorAnalysisConfig, reason, changedBy string) error {
	config := ucm.ConfigCenter.Get()
	config.BehaviorAnalysis = newConfig
	return ucm.Update(config, reason, changedBy)
}

// translated comment
func (ucm *UnifiedConfigManager) UpdateRiskScoringConfig(newConfig *RiskScoringConfig, reason, changedBy string) error {
	config := ucm.ConfigCenter.Get()
	config.RiskScoring = newConfig
	return ucm.Update(config, reason, changedBy)
}

// translated comment
func (ucm *UnifiedConfigManager) UpdateFeatureExtractionConfig(newConfig *FeatureExtractionConfig, reason, changedBy string) error {
	config := ucm.ConfigCenter.Get()
	config.Features = newConfig
	return ucm.Update(config, reason, changedBy)
}

// ============================================
// translated comment
// ============================================


