package config

import (
	"fmt"
	"sync"
	"time"
)

// EnhancedConfigCenter is an enhanced wrapper for the configuration center
type EnhancedConfigCenter struct {
	center         *ConfigCenter
	broadcastCh    chan ConfigChangeEvent
	subscribers    map[string]chan ConfigChangeEvent
	subscriberMu   sync.RWMutex
	healthChecker  *ConfigHealthChecker
	notificationMu sync.RWMutex
}

// ConfigChangeEvent represents a configuration change event
type ConfigChangeEvent struct {
	Type        ConfigChangeType
	Timestamp   time.Time
	Config      *ManagedConfig
	Changes     []ConfigChange
	Source      string
	Description string
}

// ConfigChangeType represents a configuration change type
type ConfigChangeType string

const (
	ConfigChangeTypeUpdate    ConfigChangeType = "update"
	ConfigChangeTypeReload    ConfigChangeType = "reload"
	ConfigChangeTypeRollback  ConfigChangeType = "rollback"
	ConfigChangeTypeSubscribe ConfigChangeType = "subscribe"
	ConfigChangeTypeHealth    ConfigChangeType = "health"
)

// ConfigHealthStatus represents the configuration health status
type ConfigHealthStatus struct {
	Healthy        bool
	LastCheckTime  time.Time
	Issues         []string
	Version        string
	LastUpdateTime time.Time
}

// ConfigHealthChecker is a configuration health checker
type ConfigHealthChecker struct {
	center     *ConfigCenter
	checkFuncs []HealthCheckFunc
	interval   time.Duration
	stopCh     chan struct{}
	mu         sync.RWMutex
	lastStatus ConfigHealthStatus
}

// HealthCheckFunc is a health check function
type HealthCheckFunc func(*ManagedConfig) []string

// WrapConfigCenter wraps an existing configuration center
// Deprecated: Use NewUnifiedConfigManager instead
func WrapConfigCenter(baseCenter *ConfigCenter) *EnhancedConfigCenter {
	enhanced := &EnhancedConfigCenter{
		center:      baseCenter,
		broadcastCh: make(chan ConfigChangeEvent, 100),
		subscribers: make(map[string]chan ConfigChangeEvent),
	}

	// Initialize the health checker
	enhanced.healthChecker = &ConfigHealthChecker{
		center:     baseCenter,
		checkFuncs: []HealthCheckFunc{defaultHealthCheck},
		interval:   30 * time.Second,
		stopCh:     make(chan struct{}),
		lastStatus: ConfigHealthStatus{
			Healthy:       true,
			LastCheckTime: time.Now(),
		},
	}

	// Start the broadcast processor
	go enhanced.broadcastProcessor()

	// Start the health checker
	go enhanced.healthChecker.start(enhanced.broadcastCh)

	return enhanced
}

// Subscribe subscribes to configuration change events
func (ecc *EnhancedConfigCenter) Subscribe(subscriberID string) (<-chan ConfigChangeEvent, error) {
	ecc.subscriberMu.Lock()
	defer ecc.subscriberMu.Unlock()

	if _, exists := ecc.subscribers[subscriberID]; exists {
		return nil, fmt.Errorf("subscriber %s already exists", subscriberID)
	}

	eventCh := make(chan ConfigChangeEvent, 10)
	ecc.subscribers[subscriberID] = eventCh

	// Send subscription confirmation event
	ecc.broadcastCh <- ConfigChangeEvent{
		Type:        ConfigChangeTypeSubscribe,
		Timestamp:   time.Now(),
		Description: fmt.Sprintf("Subscriber %s registered", subscriberID),
		Source:      subscriberID,
	}

	return eventCh, nil
}

// Unsubscribe unsubscribes from configuration change events
func (ecc *EnhancedConfigCenter) Unsubscribe(subscriberID string) error {
	ecc.subscriberMu.Lock()
	defer ecc.subscriberMu.Unlock()

	eventCh, exists := ecc.subscribers[subscriberID]
	if !exists {
		return fmt.Errorf("subscriber %s not found", subscriberID)
	}

	close(eventCh)
	delete(ecc.subscribers, subscriberID)

	return nil
}

// broadcastProcessor is the broadcast event processor
func (ecc *EnhancedConfigCenter) broadcastProcessor() {
	for event := range ecc.broadcastCh {
		// Copy subscriber list (to avoid holding the lock while sending)
		ecc.subscriberMu.RLock()
		subscribers := make(map[string]chan ConfigChangeEvent, len(ecc.subscribers))
		for k, v := range ecc.subscribers {
			subscribers[k] = v
		}
		ecc.subscriberMu.RUnlock()

		// Asynchronously send to all subscribers
		for _, ch := range subscribers {
			select {
			case ch <- event:
				// Sent successfully
			default:
				// Channel is full, skip
			}
		}
	}
}

// Get returns the configuration
func (ecc *EnhancedConfigCenter) Get() *ManagedConfig {
	return ecc.center.Get()
}

// Update is the enhanced version of configuration update
func (ecc *EnhancedConfigCenter) Update(newConfig *ManagedConfig, reason, changedBy string) error {
	// Call the base method
	if err := ecc.center.Update(newConfig, reason, changedBy); err != nil {
		return err
	}

	// Broadcast the update event
	ecc.broadcastCh <- ConfigChangeEvent{
		Type:        ConfigChangeTypeUpdate,
		Timestamp:   time.Now(),
		Config:      ecc.center.Get(),
		Description: reason,
		Source:      changedBy,
	}

	return nil
}

// GetHealthStatus returns the health status
func (ecc *EnhancedConfigCenter) GetHealthStatus() ConfigHealthStatus {
	ecc.healthChecker.mu.RLock()
	defer ecc.healthChecker.mu.RUnlock()
	return ecc.healthChecker.lastStatus
}

// AddHealthCheck adds a health check function
func (ecc *EnhancedConfigCenter) AddHealthCheck(checkFunc HealthCheckFunc) {
	ecc.healthChecker.mu.Lock()
	defer ecc.healthChecker.mu.Unlock()
	ecc.healthChecker.checkFuncs = append(ecc.healthChecker.checkFuncs, checkFunc)
}

// Close closes the enhanced configuration center
func (ecc *EnhancedConfigCenter) Close() error {
	// Stop the health checker
	ecc.healthChecker.stop()

	// Close the broadcast channel
	close(ecc.broadcastCh)

	// Close all subscriber channels
	ecc.subscriberMu.Lock()
	for _, ch := range ecc.subscribers {
		close(ch)
	}
	ecc.subscribers = make(map[string]chan ConfigChangeEvent)
	ecc.subscriberMu.Unlock()

	return nil
}

// start starts the health checker
func (hc *ConfigHealthChecker) start(broadcastCh chan ConfigChangeEvent) {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-hc.stopCh:
			return
		case <-ticker.C:
			hc.performCheck(broadcastCh)
		}
	}
}

// stop stops the health checker
func (hc *ConfigHealthChecker) stop() {
	select {
	case <-hc.stopCh:
		// Already closed
	default:
		close(hc.stopCh)
	}
}

// performCheck performs a health check
func (hc *ConfigHealthChecker) performCheck(broadcastCh chan ConfigChangeEvent) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	config := hc.center.Get()
	if config == nil || config.Metadata == nil {
		hc.lastStatus = ConfigHealthStatus{
			Healthy:       false,
			LastCheckTime: time.Now(),
			Issues:        []string{"config or metadata is nil"},
		}
		return
	}

	var allIssues []string

	for _, checkFunc := range hc.checkFuncs {
		issues := checkFunc(config)
		allIssues = append(allIssues, issues...)
	}

	hc.lastStatus = ConfigHealthStatus{
		Healthy:        len(allIssues) == 0,
		LastCheckTime:  time.Now(),
		Issues:         allIssues,
		Version:        config.Metadata.Version,
		LastUpdateTime: config.Metadata.LastModified,
	}

	// If there are issues, broadcast a health event
	if len(allIssues) > 0 && broadcastCh != nil {
		select {
		case broadcastCh <- ConfigChangeEvent{
			Type:        ConfigChangeTypeHealth,
			Timestamp:   time.Now(),
			Config:      config,
			Description: fmt.Sprintf("Health check found %d issues", len(allIssues)),
		}:
		default:
			// Channel is full, do not block
		}
	}
}

// defaultHealthCheck is the default health check
func defaultHealthCheck(config *ManagedConfig) []string {
	var issues []string

	// Check whether required configurations exist
	if config.BehaviorAnalysis == nil {
		issues = append(issues, "BehaviorAnalysis config is missing")
	}
	if config.RiskScoring == nil {
		issues = append(issues, "RiskScoring config is missing")
	}
	if config.Features == nil {
		issues = append(issues, "Features config is missing")
	}

	// Check the validity of configuration values
	if config.BehaviorAnalysis != nil {
		if config.BehaviorAnalysis.MinRequestsForAnalysis <= 0 {
			issues = append(issues, "MinRequestsForAnalysis must be positive")
		}
		if config.BehaviorAnalysis.RegularityThreshold < 0 || config.BehaviorAnalysis.RegularityThreshold > 1 {
			issues = append(issues, "RegularityThreshold must be between 0 and 1")
		}
	}

	return issues
}
