package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ConfigCenter is the configuration center that centrally manages all configurations
// When configPath is empty, ConfigCenter operates in memory mode where configurations are only stored in memory and not persisted to files.
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
	reloadOnce        sync.Once // Ensures autoReloadWorker is started only once
	validationEnabled bool
}

// ManagedConfig represents a managed configuration object
type ManagedConfig struct {
	// Behavior analysis configuration
	BehaviorAnalysis *BehaviorAnalysisConfig `json:"behavior_analysis"`

	// Risk scoring configuration
	RiskScoring *RiskScoringConfig `json:"risk_scoring"`

	// Feature extraction configuration
	Features *FeatureExtractionConfig `json:"features"`

	// QUIC configuration
	QUIC *QUICConfig `json:"quic"`

	// TLS configuration
	TLS *TLSConfig `json:"tls"`

	// Global limits and thresholds
	Global *GlobalConfig `json:"global"`

	// Metadata
	Metadata *ConfigMetadata `json:"metadata"`
}

// ConfigMetadata represents configuration metadata
type ConfigMetadata struct {
	Version      string    `json:"version"`
	LastModified time.Time `json:"last_modified"`
	Author       string    `json:"author"`
	Description  string    `json:"description"`
}

// VersionedConfig represents a configuration with version information
type VersionedConfig struct {
	Timestamp    time.Time
	Version      string
	Config       *ManagedConfig
	ChangeReason string
	ChangedBy    string
}

// ConfigChangeListener is the configuration change listener interface
type ConfigChangeListener interface {
	// OnConfigChange is called when the configuration changes
	OnConfigChange(old, new *ManagedConfig, changes []ConfigChange) error
}

// ConfigChange represents configuration change information
type ConfigChange struct {
	Path      string      // Changed configuration path (e.g., "behavior_analysis.min_requests")
	OldValue  interface{} // Old value
	NewValue  interface{} // New value
	Timestamp time.Time
}

// NewConfigCenter creates a new configuration center
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

// Load loads the configuration from a file
func (cc *ConfigCenter) Load() error {
	// Check the configuration path
	cc.mu.RLock()
	configPath := cc.configPath
	validationEnabled := cc.validationEnabled
	cc.mu.RUnlock()

	if configPath == "" {
		return fmt.Errorf("config path not set")
	}

	// 1. Read the file without holding the lock (IO operation)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 2. Parse JSON (CPU operation, no lock needed)
	var config ManagedConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// 3. Validate the configuration
	if validationEnabled {
		if errs := cc.validateConfig(&config); len(errs) > 0 {
			return fmt.Errorf("config validation failed: %v", errs)
		}
	}

	// 4. Get the file modification time
	fileInfo, err := os.Stat(configPath)
	if err != nil {
		return fmt.Errorf("failed to stat config file: %w", err)
	}

	// 5. Initialize metadata
	if config.Metadata == nil {
		config.Metadata = &ConfigMetadata{}
	}
	config.Metadata.LastModified = fileInfo.ModTime()

	// 6. Acquire the lock and update state
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.lastModTime = fileInfo.ModTime()
	cc.recordVersion(&config, "initial_load", "system")
	cc.current = &config
	cc.loaded = true

	return nil
}

// Get returns the current configuration (read-only)
func (cc *ConfigCenter) Get() *ManagedConfig {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.current.Clone()
}

// Update updates the configuration
func (cc *ConfigCenter) Update(newConfig *ManagedConfig, reason, changedBy string) error {
	cc.mu.Lock()

	if !cc.loaded {
		cc.mu.Unlock()
		return fmt.Errorf("config center not loaded")
	}

	// Validate the new configuration
	if cc.validationEnabled {
		if errs := cc.validateConfig(newConfig); len(errs) > 0 {
			cc.mu.Unlock()
			return fmt.Errorf("new config validation failed: %v", errs)
		}
	}

	// Detect changes
	changes := cc.detectChanges(cc.current, newConfig)

	// Save old configuration for notification
	oldConfig := cc.current

	// Copy the listener list (to avoid holding the lock during notification)
	listeners := make([]ConfigChangeListener, len(cc.listeners))
	copy(listeners, cc.listeners)

	// Update the current configuration
	cc.current = newConfig
	cc.recordVersion(newConfig, reason, changedBy)

	// Save to file (completed before releasing the lock)
	if err := cc.saveToFileLocked(); err != nil {
		// Roll back to the old configuration
		cc.current = oldConfig
		cc.mu.Unlock()
		return fmt.Errorf("failed to save config: %w", err)
	}

	cc.mu.Unlock()

	// Notify listeners (outside the lock to avoid deadlocks)
	for _, listener := range listeners {
		if err := listener.OnConfigChange(oldConfig, newConfig, changes); err != nil {
			return fmt.Errorf("listener error: %w", err)
		}
	}

	return nil
}

// SaveToFile saves the configuration to a file
func (cc *ConfigCenter) SaveToFile() error {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	return cc.saveToFileLocked()
}

// saveToFileLocked saves the configuration to a file (caller must already hold the lock)
func (cc *ConfigCenter) saveToFileLocked() error {
	// If there is no config file path, skip file saving (memory mode only)
	if cc.configPath == "" {
		return nil
	}

	data, err := json.MarshalIndent(cc.current, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Create the directory (if it does not exist)
	dir := filepath.Dir(cc.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write to file
	if err := os.WriteFile(cc.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// RegisterListener registers a configuration change listener
func (cc *ConfigCenter) RegisterListener(listener ConfigChangeListener) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.listeners = append(cc.listeners, listener)
}

// EnableAutoReload enables automatic reloading
func (cc *ConfigCenter) EnableAutoReload(interval time.Duration) {
	cc.mu.Lock()
	cc.autoReload = true
	if interval < time.Second {
		interval = time.Second // Minimum interval of 1 second to prevent overly frequent checks
	}
	cc.reloadInterval = interval
	cc.mu.Unlock()

	// Use sync.Once to ensure only one autoReloadWorker is started
	cc.reloadOnce.Do(func() {
		go cc.autoReloadWorker()
	})
}

// DisableAutoReload disables automatic reloading
func (cc *ConfigCenter) DisableAutoReload() {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	if cc.autoReload {
		cc.autoReload = false
		close(cc.stopReloadChan)
	}
}

// autoReloadWorker is the automatic reload worker
func (cc *ConfigCenter) autoReloadWorker() {
	for {
		// Get the current stop channel and interval
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
						// Notify listeners that the configuration has been reloaded
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

// GetHistory returns the configuration version history
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

// Rollback rolls back to the specified version
func (cc *ConfigCenter) Rollback(version string, reason, rolledBackBy string) error {
	cc.mu.Lock()

	if !cc.loaded {
		cc.mu.Unlock()
		return fmt.Errorf("config center not loaded")
	}

	// Find the specified version
	var targetVersion *VersionedConfig
	for _, v := range cc.history {
		if v.Version == version {
			targetVersion = v
			break
		}
	}

	if targetVersion == nil {
		cc.mu.Unlock()
		return fmt.Errorf("version not found: %s", version)
	}

	// Create a configuration copy
	oldConfig := cc.current
	newConfig := cc.copyConfig(targetVersion.Config)

	// Detect changes
	changes := cc.detectChanges(cc.current, newConfig)

	listeners := make([]ConfigChangeListener, len(cc.listeners))
	copy(listeners, cc.listeners)

	// Update the current configuration
	cc.current = newConfig
	cc.recordVersion(newConfig, reason+" (rolled back from "+version+")", rolledBackBy)

	// Save to file
	if err := cc.saveToFileLocked(); err != nil {
		// Roll back to the old configuration
		cc.current = oldConfig
		cc.mu.Unlock()
		return fmt.Errorf("failed to save config: %w", err)
	}

	cc.mu.Unlock()

	// Notify listeners outside lock to avoid deadlock and long lock hold.
	for _, listener := range listeners {
		if err := listener.OnConfigChange(oldConfig, newConfig, changes); err != nil {
			return fmt.Errorf("listener error: %w", err)
		}
	}

	return nil
}

// recordVersion records a configuration version
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

	// Limit history size
	if len(cc.history) > cc.maxHistorySize {
		cc.history = cc.history[len(cc.history)-cc.maxHistorySize:]
	}
}

// detectChanges detects configuration changes
func (cc *ConfigCenter) detectChanges(old, new *ManagedConfig) []ConfigChange {
	changes := make([]ConfigChange, 0)
	if old == nil || new == nil {
		return changes
	}

	// Behavior analysis configuration changes
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

// copyConfig deep copies the configuration
// Uses a hand-written Clone method, which is faster than JSON serialization
func (cc *ConfigCenter) copyConfig(config *ManagedConfig) *ManagedConfig {
	return config.Clone()
}

// validateConfig validates the configuration (basic validation)
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

// IsLoaded checks whether the configuration has been loaded
func (cc *ConfigCenter) IsLoaded() bool {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.loaded
}

// SetValidationEnabled sets whether configuration validation is enabled
func (cc *ConfigCenter) SetValidationEnabled(enabled bool) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.validationEnabled = enabled
}
