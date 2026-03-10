package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// translated comment
// translated comment
type ConfigCenter struct {
	mu                sync.RWMutex
	current           *ManagedConfig
	history           []*VersionedConfig
	listeners         []ConfigChangeListener
	loaded            bool
	lastModTime       time.Time
	configPath        string
	maxHistorySize    int
	autoReload        bool
	reloadInterval    time.Duration
	stopReloadChan    chan struct{}
	reloadOnce        sync.Once // translated comment
	validationEnabled bool
}

// translated comment
type ManagedConfig struct {
	// translated comment
	BehaviorAnalysis *BehaviorAnalysisConfig `json:"behavior_analysis"`

	// translated comment
	RiskScoring *RiskScoringConfig `json:"risk_scoring"`

	// translated comment
	Features *FeatureExtractionConfig `json:"features"`

	// translated comment
	QUIC *QUICConfig `json:"quic"`

	// translated comment
	TLS *TLSConfig `json:"tls"`

	// translated comment
	Global *GlobalConfig `json:"global"`

	// translated comment
	Metadata *ConfigMetadata `json:"metadata"`
}

// translated comment
type ConfigMetadata struct {
	Version      string    `json:"version"`
	LastModified time.Time `json:"last_modified"`
	Author       string    `json:"author"`
	Description  string    `json:"description"`
}

// translated comment
type VersionedConfig struct {
	Timestamp    time.Time
	Version      string
	Config       *ManagedConfig
	ChangeReason string
	ChangedBy    string
}

// translated comment
type ConfigChangeListener interface {
	// translated comment
	OnConfigChange(old, new *ManagedConfig, changes []ConfigChange) error
}

// translated comment
type ConfigChange struct {
	Path      string      // translated comment
	OldValue  interface{} // translated comment
	NewValue  interface{} // translated comment
	Timestamp time.Time
}

// translated comment
func NewConfigCenter(configPath string) *ConfigCenter {
	return &ConfigCenter{
		current:           &ManagedConfig{},
		history:           make([]*VersionedConfig, 0),
		listeners:         make([]ConfigChangeListener, 0),
		configPath:        configPath,
		maxHistorySize:    50,
		autoReload:        false,
		reloadInterval:    30 * time.Second,
		stopReloadChan:    make(chan struct{}),
		validationEnabled: true,
	}
}

// translated comment
func (cc *ConfigCenter) Load() error {
	// translated comment
	cc.mu.RLock()
	configPath := cc.configPath
	validationEnabled := cc.validationEnabled
	cc.mu.RUnlock()

	if configPath == "" {
		return fmt.Errorf("config path not set")
	}

	// translated comment
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// translated comment
	var config ManagedConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// translated comment
	if validationEnabled {
		if errs := cc.validateConfig(&config); len(errs) > 0 {
			return fmt.Errorf("config validation failed: %v", errs)
		}
	}

	// translated comment
	fileInfo, err := os.Stat(configPath)
	if err != nil {
		return fmt.Errorf("failed to stat config file: %w", err)
	}

	// translated comment
	if config.Metadata == nil {
		config.Metadata = &ConfigMetadata{}
	}
	config.Metadata.LastModified = fileInfo.ModTime()

	// translated comment
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.lastModTime = fileInfo.ModTime()
	cc.recordVersion(&config, "initial_load", "system")
	cc.current = &config
	cc.loaded = true

	return nil
}

// translated comment
func (cc *ConfigCenter) Get() *ManagedConfig {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.current
}

// translated comment
func (cc *ConfigCenter) Update(newConfig *ManagedConfig, reason, changedBy string) error {
	cc.mu.Lock()

	if !cc.loaded {
		cc.mu.Unlock()
		return fmt.Errorf("config center not loaded")
	}

	// translated comment
	if cc.validationEnabled {
		if errs := cc.validateConfig(newConfig); len(errs) > 0 {
			cc.mu.Unlock()
			return fmt.Errorf("new config validation failed: %v", errs)
		}
	}

	// translated comment
	changes := cc.detectChanges(cc.current, newConfig)

	// translated comment
	oldConfig := cc.current

	// translated comment
	listeners := make([]ConfigChangeListener, len(cc.listeners))
	copy(listeners, cc.listeners)

	// translated comment
	cc.current = newConfig
	cc.recordVersion(newConfig, reason, changedBy)

	// translated comment
	if err := cc.saveToFileLocked(); err != nil {
		// translated comment
		cc.current = oldConfig
		cc.mu.Unlock()
		return fmt.Errorf("failed to save config: %w", err)
	}

	cc.mu.Unlock()

	// translated comment
	for _, listener := range listeners {
		if err := listener.OnConfigChange(oldConfig, newConfig, changes); err != nil {
			return fmt.Errorf("listener error: %w", err)
		}
	}

	return nil
}

// translated comment
func (cc *ConfigCenter) SaveToFile() error {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	return cc.saveToFileLocked()
}

// translated comment
func (cc *ConfigCenter) saveToFileLocked() error {
	// translated comment
	if cc.configPath == "" {
		return nil
	}

	data, err := json.MarshalIndent(cc.current, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// translated comment
	dir := filepath.Dir(cc.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// translated comment
	if err := os.WriteFile(cc.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// translated comment
func (cc *ConfigCenter) RegisterListener(listener ConfigChangeListener) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.listeners = append(cc.listeners, listener)
}

// translated comment
func (cc *ConfigCenter) EnableAutoReload(interval time.Duration) {
	cc.mu.Lock()
	cc.autoReload = true
	if interval < time.Second {
		interval = time.Second // translated comment
	}
	cc.reloadInterval = interval
	cc.mu.Unlock()

	// translated comment
	cc.reloadOnce.Do(func() {
		go cc.autoReloadWorker()
	})
}

// translated comment
func (cc *ConfigCenter) DisableAutoReload() {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	if cc.autoReload {
		cc.autoReload = false
		close(cc.stopReloadChan)
	}
}

// translated comment
func (cc *ConfigCenter) autoReloadWorker() {
	for {
		// translated comment
		cc.mu.RLock()
		stopChan := cc.stopReloadChan
		interval := cc.reloadInterval
		cc.mu.RUnlock()

		ticker := time.NewTicker(interval)
		stopped := false

		for !stopped {
			select {
			case <-stopChan:
				ticker.Stop()
				stopped = true
			case <-ticker.C:
				cc.mu.RLock()
				isAutoReload := cc.autoReload
				configPath := cc.configPath
				lastModTime := cc.lastModTime
				listeners := make([]ConfigChangeListener, len(cc.listeners))
				copy(listeners, cc.listeners)
				cc.mu.RUnlock()

				if !isAutoReload {
					continue
				}

				fileInfo, err := os.Stat(configPath)
				if err != nil {
					continue
				}

				if fileInfo.ModTime().After(lastModTime) {
					if err := cc.Load(); err == nil {
						// translated comment
						current := cc.Get()
						for _, listener := range listeners {
							listener.OnConfigChange(nil, current, nil)
						}
					}
				}
			}
		}
	}
}

// translated comment
func (cc *ConfigCenter) GetHistory(limit int) []*VersionedConfig {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	if limit <= 0 || limit > len(cc.history) {
		limit = len(cc.history)
	}

	result := make([]*VersionedConfig, limit)
	copy(result, cc.history[len(cc.history)-limit:])
	return result
}

// translated comment
func (cc *ConfigCenter) Rollback(version string, reason, rolledBackBy string) error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	if !cc.loaded {
		return fmt.Errorf("config center not loaded")
	}

	// translated comment
	var targetVersion *VersionedConfig
	for _, v := range cc.history {
		if v.Version == version {
			targetVersion = v
			break
		}
	}

	if targetVersion == nil {
		return fmt.Errorf("version not found: %s", version)
	}

	// translated comment
	oldConfig := cc.current
	newConfig := cc.copyConfig(targetVersion.Config)

	// translated comment
	changes := cc.detectChanges(cc.current, newConfig)

	// translated comment
	for _, listener := range cc.listeners {
		if err := listener.OnConfigChange(cc.current, newConfig, changes); err != nil {
			return fmt.Errorf("listener error: %w", err)
		}
	}

	// translated comment
	cc.current = newConfig
	cc.recordVersion(newConfig, reason+" (rolled back from "+version+")", rolledBackBy)

	// translated comment
	if err := cc.saveToFileLocked(); err != nil {
		// translated comment
		cc.current = oldConfig
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// translated comment
func (cc *ConfigCenter) recordVersion(config *ManagedConfig, reason, changedBy string) {
	version := fmt.Sprintf("v%d", len(cc.history)+1)

	versionedConfig := &VersionedConfig{
		Timestamp:    time.Now(),
		Version:      version,
		Config:       cc.copyConfig(config),
		ChangeReason: reason,
		ChangedBy:    changedBy,
	}

	cc.history = append(cc.history, versionedConfig)

	// translated comment
	if len(cc.history) > cc.maxHistorySize {
		cc.history = cc.history[len(cc.history)-cc.maxHistorySize:]
	}
}

// translated comment
func (cc *ConfigCenter) detectChanges(old, new *ManagedConfig) []ConfigChange {
	changes := make([]ConfigChange, 0)
	if old == nil || new == nil {
		return changes
	}

	// translated comment
	if old.BehaviorAnalysis != nil && new.BehaviorAnalysis != nil {
		if old.BehaviorAnalysis.MinRequestsForAnalysis != new.BehaviorAnalysis.MinRequestsForAnalysis {
			changes = append(changes, ConfigChange{
				Path:      "behavior_analysis.min_requests",
				OldValue:  old.BehaviorAnalysis.MinRequestsForAnalysis,
				NewValue:  new.BehaviorAnalysis.MinRequestsForAnalysis,
				Timestamp: time.Now(),
			})
		}
	}

	return changes
}

// translated comment
// translated comment
func (cc *ConfigCenter) copyConfig(config *ManagedConfig) *ManagedConfig {
	return config.Clone()
}

// translated comment
func (cc *ConfigCenter) validateConfig(config *ManagedConfig) []error {
	errs := make([]error, 0)

	if config.BehaviorAnalysis != nil {
		if config.BehaviorAnalysis.MinRequestsForAnalysis <= 0 {
			errs = append(errs, fmt.Errorf("MinRequestsForAnalysis must be > 0"))
		}
		if config.BehaviorAnalysis.RegularityThreshold < 0 || config.BehaviorAnalysis.RegularityThreshold > 1 {
			errs = append(errs, fmt.Errorf("RegularityThreshold must be between 0 and 1"))
		}
	}

	return errs
}

// translated comment
func (cc *ConfigCenter) IsLoaded() bool {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.loaded
}

// translated comment
func (cc *ConfigCenter) SetValidationEnabled(enabled bool) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.validationEnabled = enabled
}
