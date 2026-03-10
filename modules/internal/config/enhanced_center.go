package config

import (
	"fmt"
	"sync"
	"time"
)

// translated comment
type EnhancedConfigCenter struct {
	center         *ConfigCenter
	broadcastCh    chan ConfigChangeEvent
	subscribers    map[string]chan ConfigChangeEvent
	subscriberMu   sync.RWMutex
	healthChecker  *ConfigHealthChecker
	notificationMu sync.RWMutex
}

// translated comment
type ConfigChangeEvent struct {
	Type        ConfigChangeType
	Timestamp   time.Time
	Config      *ManagedConfig
	Changes     []ConfigChange
	Source      string
	Description string
}

// translated comment
type ConfigChangeType string

const (
	ConfigChangeTypeUpdate    ConfigChangeType = "update"
	ConfigChangeTypeReload    ConfigChangeType = "reload"
	ConfigChangeTypeRollback  ConfigChangeType = "rollback"
	ConfigChangeTypeSubscribe ConfigChangeType = "subscribe"
	ConfigChangeTypeHealth    ConfigChangeType = "health"
)

// translated comment
type ConfigHealthStatus struct {
	Healthy        bool
	LastCheckTime  time.Time
	Issues         []string
	Version        string
	LastUpdateTime time.Time
}

// translated comment
type ConfigHealthChecker struct {
	center     *ConfigCenter
	checkFuncs []HealthCheckFunc
	interval   time.Duration
	stopCh     chan struct{}
	mu         sync.RWMutex
	lastStatus ConfigHealthStatus
}

// translated comment
type HealthCheckFunc func(*ManagedConfig) []string

// translated comment
// translated comment
func WrapConfigCenter(baseCenter *ConfigCenter) *EnhancedConfigCenter {
	enhanced := &EnhancedConfigCenter{
		center:      baseCenter,
		broadcastCh: make(chan ConfigChangeEvent, 100),
		subscribers: make(map[string]chan ConfigChangeEvent),
	}

	// translated comment
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

	// translated comment
	go enhanced.broadcastProcessor()

	// translated comment
	go enhanced.healthChecker.start(enhanced.broadcastCh)

	return enhanced
}

// translated comment
func (ecc *EnhancedConfigCenter) Subscribe(subscriberID string) (<-chan ConfigChangeEvent, error) {
	ecc.subscriberMu.Lock()
	defer ecc.subscriberMu.Unlock()

	if _, exists := ecc.subscribers[subscriberID]; exists {
		return nil, fmt.Errorf("subscriber %s already exists", subscriberID)
	}

	eventCh := make(chan ConfigChangeEvent, 10)
	ecc.subscribers[subscriberID] = eventCh

	// translated comment
	ecc.broadcastCh <- ConfigChangeEvent{
		Type:        ConfigChangeTypeSubscribe,
		Timestamp:   time.Now(),
		Description: fmt.Sprintf("Subscriber %s registered", subscriberID),
		Source:      subscriberID,
	}

	return eventCh, nil
}

// translated comment
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

// translated comment
func (ecc *EnhancedConfigCenter) broadcastProcessor() {
	for event := range ecc.broadcastCh {
		// translated comment
		ecc.subscriberMu.RLock()
		subscribers := make(map[string]chan ConfigChangeEvent, len(ecc.subscribers))
		for k, v := range ecc.subscribers {
			subscribers[k] = v
		}
		ecc.subscriberMu.RUnlock()

		// translated comment
		for _, ch := range subscribers {
			select {
			case ch <- event:
				// translated comment
			default:
				// translated comment
			}
		}
	}
}

// translated comment
func (ecc *EnhancedConfigCenter) Get() *ManagedConfig {
	return ecc.center.Get()
}

// translated comment
func (ecc *EnhancedConfigCenter) Update(newConfig *ManagedConfig, reason, changedBy string) error {
	// translated comment
	if err := ecc.center.Update(newConfig, reason, changedBy); err != nil {
		return err
	}

	// translated comment
	ecc.broadcastCh <- ConfigChangeEvent{
		Type:        ConfigChangeTypeUpdate,
		Timestamp:   time.Now(),
		Config:      ecc.center.Get(),
		Description: reason,
		Source:      changedBy,
	}

	return nil
}

// translated comment
func (ecc *EnhancedConfigCenter) GetHealthStatus() ConfigHealthStatus {
	ecc.healthChecker.mu.RLock()
	defer ecc.healthChecker.mu.RUnlock()
	return ecc.healthChecker.lastStatus
}

// translated comment
func (ecc *EnhancedConfigCenter) AddHealthCheck(checkFunc HealthCheckFunc) {
	ecc.healthChecker.mu.Lock()
	defer ecc.healthChecker.mu.Unlock()
	ecc.healthChecker.checkFuncs = append(ecc.healthChecker.checkFuncs, checkFunc)
}

// translated comment
func (ecc *EnhancedConfigCenter) Close() error {
	// translated comment
	ecc.healthChecker.stop()

	// translated comment
	close(ecc.broadcastCh)

	// translated comment
	ecc.subscriberMu.Lock()
	for _, ch := range ecc.subscribers {
		close(ch)
	}
	ecc.subscribers = make(map[string]chan ConfigChangeEvent)
	ecc.subscriberMu.Unlock()

	return nil
}

// translated comment
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

// translated comment
func (hc *ConfigHealthChecker) stop() {
	select {
	case <-hc.stopCh:
		// translated comment
	default:
		close(hc.stopCh)
	}
}

// translated comment
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

	// translated comment
	if len(allIssues) > 0 && broadcastCh != nil {
		select {
		case broadcastCh <- ConfigChangeEvent{
			Type:        ConfigChangeTypeHealth,
			Timestamp:   time.Now(),
			Config:      config,
			Description: fmt.Sprintf("Health check found %d issues", len(allIssues)),
		}:
		default:
			// translated comment
		}
	}
}

// translated comment
func defaultHealthCheck(config *ManagedConfig) []string {
	var issues []string

	// translated comment
	if config.BehaviorAnalysis == nil {
		issues = append(issues, "BehaviorAnalysis config is missing")
	}
	if config.RiskScoring == nil {
		issues = append(issues, "RiskScoring config is missing")
	}
	if config.Features == nil {
		issues = append(issues, "Features config is missing")
	}

	// translated comment
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
