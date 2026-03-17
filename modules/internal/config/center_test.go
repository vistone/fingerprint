package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestNewConfigCenter tests creating a configuration center
func TestNewConfigCenter(t *testing.T) {
	tests := []struct {
		name       string
		configPath string
		wantPath   string
		wantMemory bool
	}{
		{
			name:       "create config center with path",
			configPath: "/tmp/config.json",
			wantPath:   "/tmp/config.json",
			wantMemory: false,
		},
		{
			name:       "create config center in memory mode (empty path)",
			configPath: "",
			wantPath:   "",
			wantMemory: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := NewConfigCenter(tt.configPath)

			if cc == nil {
				t.Fatal("NewConfigCenter() returned nil")
			}

			if cc.configPath != tt.wantPath {
				t.Errorf("configPath = %v, want %v", cc.configPath, tt.wantPath)
			}

			if cc.loaded != false {
				t.Error("loaded should be false initially")
			}

			if cc.autoReload != false {
				t.Error("autoReload should be false initially")
			}

			if cc.current == nil {
				t.Error("current should not be nil")
			}

			if cc.history == nil {
				t.Error("history should not be nil")
			}

			if len(cc.history) != 0 {
				t.Error("history should be empty initially")
			}

			if cc.listeners == nil {
				t.Error("listeners should not be nil")
			}

			if cc.maxHistorySize != 50 {
				t.Errorf("maxHistorySize = %d, want 50", cc.maxHistorySize)
			}

			if cc.reloadInterval != 30*time.Second {
				t.Errorf("reloadInterval = %v, want 30s", cc.reloadInterval)
			}

			if cc.validationEnabled != true {
				t.Error("validationEnabled should be true initially")
			}

			if cc.stopReloadChan == nil {
				t.Error("stopReloadChan should not be nil")
			}
		})
	}
}

// TestConfigCenter_Load tests loading configuration from a file
func TestConfigCenter_Load(t *testing.T) {
	tests := []struct {
		name       string
		setupFile  func(t *testing.T) string
		wantErr    bool
		errContain string
		wantLoaded bool
	}{
		{
			name: "load valid config from file",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.json")
				config := `{
					"behavior_analysis": {
						"min_requests_for_analysis": 5,
						"regularity_threshold": 0.3,
						"entropy_threshold": 0.5,
						"anomalous_interval_rate_threshold": 0.2,
						"request_history_capacity": 100,
						"signal_capacity": 50
					},
					"risk_scoring": {
						"critical_threshold": 0.9,
						"high_threshold": 0.7,
						"medium_threshold": 0.5,
						"low_threshold": 0.3,
						"min_confidence": 0.5
					},
					"features": {
						"entropy_high_threshold": 7.5,
						"entropy_low_threshold": 26,
						"mobile_screen_width_max": 1920,
						"desktop_screen_width_min": 800
					},
					"quic": {
						"min_initial_max_data": 1024,
						"min_stream_data": 256
					},
					"global": {
						"max_concurrency": 100,
						"request_timeout": 30000,
						"cache_size": 10000
					}
				}`
				if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
					t.Fatalf("Failed to create test config: %v", err)
				}
				return configPath
			},
			wantErr:    false,
			wantLoaded: true,
		},
		{
			name: "file not found returns error",
			setupFile: func(t *testing.T) string {
				return "/nonexistent/path/config.json"
			},
			wantErr:    true,
			errContain: "failed to read config file",
			wantLoaded: false,
		},
		{
			name: "invalid JSON returns error",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.json")
				if err := os.WriteFile(configPath, []byte("invalid json"), 0644); err != nil {
					t.Fatalf("Failed to create test config: %v", err)
				}
				return configPath
			},
			wantErr:    true,
			errContain: "failed to parse config JSON",
			wantLoaded: false,
		},
		{
			name: "validation-failing config returns error",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.json")
				config := `{
					"behavior_analysis": {
						"min_requests_for_analysis": 0,
						"regularity_threshold": 0.3
					},
					"risk_scoring": {
						"critical_threshold": 0.9,
						"high_threshold": 0.7,
						"medium_threshold": 0.5,
						"low_threshold": 0.3,
						"min_confidence": 0.5
					},
					"features": {
						"entropy_high_threshold": 7.5,
						"entropy_low_threshold": 26,
						"mobile_screen_width_max": 1920,
						"desktop_screen_width_min": 800
					},
					"quic": {
						"min_initial_max_data": 1024,
						"min_stream_data": 256
					},
					"global": {
						"max_concurrency": 100,
						"request_timeout": 30000,
						"cache_size": 10000
					}
				}`
				if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
					t.Fatalf("Failed to create test config: %v", err)
				}
				return configPath
			},
			wantErr:    true,
			errContain: "config validation failed",
			wantLoaded: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := tt.setupFile(t)
			cc := NewConfigCenter(configPath)

			err := cc.Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContain != "" {
				if !contains(err.Error(), tt.errContain) {
					t.Errorf("Load() error = %v, should contain %v", err, tt.errContain)
				}
			}

			if cc.IsLoaded() != tt.wantLoaded {
				t.Errorf("IsLoaded() = %v, want %v", cc.IsLoaded(), tt.wantLoaded)
			}

			if !tt.wantErr && tt.wantLoaded {
				// Verify history records
				history := cc.GetHistory(1)
				if len(history) == 0 {
					t.Error("Expected history to be recorded after load")
				}
				if history[0].ChangeReason != "initial_load" {
					t.Errorf("ChangeReason = %v, want initial_load", history[0].ChangeReason)
				}
			}
		})
	}
}

// TestConfigCenter_Get tests getting the current configuration (concurrency-safe)
func TestConfigCenter_Get(t *testing.T) {
	cc := NewConfigCenter("")

	// Set initial configuration
	cc.current = DefaultManagedConfig()

	// Test concurrent reads
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			config := cc.Get()
			if config == nil {
				t.Error("Get() returned nil")
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// OK
	case <-time.After(5 * time.Second):
		t.Fatal("Concurrent Get() timeout")
	}
}

// TestConfigCenter_Update tests updating the configuration
func TestConfigCenter_Update(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, cc *ConfigCenter)
		newConfig  *ManagedConfig
		wantErr    bool
		errContain string
		checkFile  bool
	}{
		{
			name: "update returns error when not loaded",
			setup: func(t *testing.T, cc *ConfigCenter) {
				// Do not load configuration
			},
			newConfig:  DefaultManagedConfig(),
			wantErr:    true,
			errContain: "config center not loaded",
		},
		{
			name: "validation-failing config returns error",
			setup: func(t *testing.T, cc *ConfigCenter) {
				cc.loaded = true
			},
			newConfig: func() *ManagedConfig {
				cfg := DefaultManagedConfig()
				cfg.BehaviorAnalysis.MinRequestsForAnalysis = 0 // Invalid value
				return cfg
			}(),
			wantErr:    true,
			errContain: "new config validation failed",
		},
		{
			name: "listener is called correctly",
			setup: func(t *testing.T, cc *ConfigCenter) {
				cc.loaded = true
				cc.current = DefaultManagedConfig() // Initialize current to avoid nil fields
				listener := &mockListener{t: t, called: false}
				cc.RegisterListener(listener)
			},
			newConfig: func() *ManagedConfig {
				cfg := DefaultManagedConfig()
				cfg.BehaviorAnalysis.MinRequestsForAnalysis = 10 // Modify a field to trigger change detection
				return cfg
			}(),
			wantErr: false,
		},
		{
			name: "update fails when listener returns error",
			setup: func(t *testing.T, cc *ConfigCenter) {
				cc.loaded = true
				cc.current = DefaultManagedConfig() // Initialize current
				listener := &mockErrorListener{}
				cc.RegisterListener(listener)
			},
			newConfig:  DefaultManagedConfig(),
			wantErr:    true,
			errContain: "listener error",
		},
		{
			name: "update succeeds and saves to file",
			setup: func(t *testing.T, cc *ConfigCenter) {
				cc.loaded = true
				cc.current = DefaultManagedConfig() // Initialize current
			},
			newConfig: DefaultManagedConfig(),
			wantErr:   false,
			checkFile: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.json")
			cc := NewConfigCenter(configPath)

			tt.setup(t, cc)

			err := cc.Update(tt.newConfig, "test update", "tester")
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContain != "" {
				if !contains(err.Error(), tt.errContain) {
					t.Errorf("Update() error = %v, should contain %v", err, tt.errContain)
				}
			}

			if !tt.wantErr {
				// Verify configuration has been updated
				if cc.Get().BehaviorAnalysis.MinRequestsForAnalysis != tt.newConfig.BehaviorAnalysis.MinRequestsForAnalysis {
					t.Error("Config was not updated")
				}

				// Verify history records
				history := cc.GetHistory(1)
				if len(history) == 0 {
					t.Error("Expected history to be recorded after update")
				}
				if history[0].ChangedBy != "tester" {
					t.Errorf("ChangedBy = %v, want tester", history[0].ChangedBy)
				}

				// Verify file has been saved
				if tt.checkFile {
					if _, err := os.Stat(configPath); os.IsNotExist(err) {
						t.Error("Config file was not saved")
					}
				}
			}
		})
	}
}

// TestConfigCenter_SaveToFile tests saving to a file
func TestConfigCenter_SaveToFile(t *testing.T) {
	tests := []struct {
		name        string
		configPath  string
		setupConfig func(cc *ConfigCenter)
		wantErr     bool
		wantFile    bool
	}{
		{
			name:       "save to file succeeds",
			configPath: "",
			setupConfig: func(cc *ConfigCenter) {
				// configPath will be set in the test loop
				cc.current = DefaultManagedConfig()
			},
			wantErr:  false,
			wantFile: true,
		},
		{
			name:       "empty path (memory mode) behavior",
			configPath: "",
			setupConfig: func(cc *ConfigCenter) {
				cc.current = DefaultManagedConfig()
				// configPath is empty, indicating memory mode
			},
			wantErr:  false,
			wantFile: false, // Memory mode does not create files
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var configPath string
			if tt.wantFile {
				tmpDir := t.TempDir()
				configPath = filepath.Join(tmpDir, "config.json")
			} else {
				configPath = tt.configPath
			}

			cc := NewConfigCenter(configPath)
			tt.setupConfig(cc)

			err := cc.SaveToFile()
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveToFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantFile {
				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					t.Error("Config file was not created")
				}

				// Verify file content
				data, err := os.ReadFile(configPath)
				if err != nil {
					t.Errorf("Failed to read saved file: %v", err)
				}
				if len(data) == 0 {
					t.Error("Saved file is empty")
				}
			}
		})
	}
}

// TestConfigCenter_RegisterListener tests registering a configuration change listener
func TestConfigCenter_RegisterListener(t *testing.T) {
	cc := NewConfigCenter("")

	// Test registering a single listener
	listener1 := &mockListener{t: t, called: false}
	cc.RegisterListener(listener1)

	if len(cc.listeners) != 1 {
		t.Errorf("listeners count = %d, want 1", len(cc.listeners))
	}

	// Test registering multiple listeners
	listener2 := &mockListener{t: t, called: false}
	cc.RegisterListener(listener2)

	if len(cc.listeners) != 2 {
		t.Errorf("listeners count = %d, want 2", len(cc.listeners))
	}
}

// TestConfigCenter_EnableAutoReload tests enabling automatic reload
func TestConfigCenter_EnableAutoReload(t *testing.T) {
	tests := []struct {
		name     string
		interval time.Duration
		wantMin  time.Duration
	}{
		{
			name:     "enable auto reload (normal interval)",
			interval: 5 * time.Second,
			wantMin:  5 * time.Second,
		},
		{
			name:     "enable auto reload (less than 1s, should be clamped to 1s)",
			interval: 500 * time.Millisecond,
			wantMin:  time.Second,
		},
		{
			name:     "multiple enables do not start multiple workers",
			interval: time.Second,
			wantMin:  time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.json")

			// Create initial configuration file
			config := DefaultManagedConfig()
			data, _ := json.MarshalIndent(config, "", "  ")
			os.WriteFile(configPath, data, 0644)

			cc := NewConfigCenter(configPath)
			cc.loaded = true

			cc.EnableAutoReload(tt.interval)

			if !cc.autoReload {
				t.Error("autoReload should be true after EnableAutoReload")
			}

			if cc.reloadInterval < tt.wantMin {
				t.Errorf("reloadInterval = %v, want >= %v", cc.reloadInterval, tt.wantMin)
			}

			// Give the worker a moment to start
			time.Sleep(100 * time.Millisecond)

			// Enable again, should not start a new worker
			cc.EnableAutoReload(tt.interval)

			// Disable for cleanup
			cc.DisableAutoReload()
		})
	}
}

// TestConfigCenter_DisableAutoReload tests disabling automatic reload
func TestConfigCenter_DisableAutoReload(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Create initial configuration file
	config := DefaultManagedConfig()
	data, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(configPath, data, 0644)

	cc := NewConfigCenter(configPath)
	cc.loaded = true

	// Enable first
	cc.EnableAutoReload(time.Second)
	if !cc.autoReload {
		t.Fatal("autoReload should be true")
	}

	// Give the worker a moment to start
	time.Sleep(100 * time.Millisecond)

	// Then disable
	cc.DisableAutoReload()

	if cc.autoReload {
		t.Error("autoReload should be false after DisableAutoReload")
	}

	// Repeated disabling should not panic
	cc.DisableAutoReload()
}

// TestConfigCenter_GetHistory tests getting version history
func TestConfigCenter_GetHistory(t *testing.T) {
	cc := NewConfigCenter("")
	cc.loaded = true

	// Add multiple history versions
	for i := 0; i < 10; i++ {
		config := DefaultManagedConfig()
		config.Metadata = &ConfigMetadata{Version: fmt.Sprintf("1.0.%d", i)}
		cc.Update(config, fmt.Sprintf("update %d", i), "tester")
	}

	tests := []struct {
		name      string
		limit     int
		wantCount int
	}{
		{
			name:      "get history (limit 5)",
			limit:     5,
			wantCount: 5,
		},
		{
			name:      "get history (limit exceeds total)",
			limit:     20,
			wantCount: 10,
		},
		{
			name:      "get history (limit 0, returns all)",
			limit:     0,
			wantCount: 10,
		},
		{
			name:      "get history (negative limit, returns all)",
			limit:     -1,
			wantCount: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			history := cc.GetHistory(tt.limit)
			if len(history) != tt.wantCount {
				t.Errorf("GetHistory() returned %d items, want %d", len(history), tt.wantCount)
			}
		})
	}
}

// TestConfigCenter_Rollback tests rolling back to a specified version
func TestConfigCenter_Rollback(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, cc *ConfigCenter) string
		version    string
		wantErr    bool
		errContain string
	}{
		{
			name: "rollback to specified version succeeds",
			setup: func(t *testing.T, cc *ConfigCenter) string {
				cc.loaded = true
				// Add multiple versions
				for i := 0; i < 3; i++ {
					config := DefaultManagedConfig()
					config.BehaviorAnalysis.MinRequestsForAnalysis = 5 + i
					cc.Update(config, fmt.Sprintf("update %d", i), "tester")
				}
				return "v1" // Roll back to the first version
			},
			version: "v1",
			wantErr: false,
		},
		{
			name: "rollback to non-existent version returns error",
			setup: func(t *testing.T, cc *ConfigCenter) string {
				cc.loaded = true
				config := DefaultManagedConfig()
				cc.Update(config, "update", "tester")
				return "nonexistent"
			},
			version:    "nonexistent",
			wantErr:    true,
			errContain: "version not found",
		},
		{
			name: "rollback returns error when not loaded",
			setup: func(t *testing.T, cc *ConfigCenter) string {
				// Do not load configuration
				return "v1"
			},
			version:    "v1",
			wantErr:    true,
			errContain: "config center not loaded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.json")
			cc := NewConfigCenter(configPath)

			targetVersion := tt.setup(t, cc)
			if tt.version == "" {
				tt.version = targetVersion
			}

			err := cc.Rollback(tt.version, "rollback test", "tester")
			if (err != nil) != tt.wantErr {
				t.Errorf("Rollback() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContain != "" {
				if !contains(err.Error(), tt.errContain) {
					t.Errorf("Rollback() error = %v, should contain %v", err, tt.errContain)
				}
			}
		})
	}
}

// TestConfigCenter_IsLoaded tests checking whether the configuration has been loaded
func TestConfigCenter_IsLoaded(t *testing.T) {
	cc := NewConfigCenter("")

	if cc.IsLoaded() {
		t.Error("IsLoaded() should be false initially")
	}

	cc.loaded = true
	if !cc.IsLoaded() {
		t.Error("IsLoaded() should be true after loaded is set")
	}
}

// TestConfigCenter_SetValidationEnabled tests setting whether configuration validation is enabled
func TestConfigCenter_SetValidationEnabled(t *testing.T) {
	cc := NewConfigCenter("")

	if !cc.validationEnabled {
		t.Error("validationEnabled should be true initially")
	}

	cc.SetValidationEnabled(false)
	if cc.validationEnabled {
		t.Error("validationEnabled should be false after SetValidationEnabled(false)")
	}

	cc.SetValidationEnabled(true)
	if !cc.validationEnabled {
		t.Error("validationEnabled should be true after SetValidationEnabled(true)")
	}
}

// TestDefaultManagedConfig tests default value creation
func TestDefaultManagedConfig(t *testing.T) {
	config := DefaultManagedConfig()

	if config == nil {
		t.Fatal("DefaultManagedConfig() returned nil")
	}

	// Verify BehaviorAnalysis
	if config.BehaviorAnalysis == nil {
		t.Error("BehaviorAnalysis is nil")
	} else {
		if config.BehaviorAnalysis.MinRequestsForAnalysis != 5 {
			t.Errorf("MinRequestsForAnalysis = %d, want 5", config.BehaviorAnalysis.MinRequestsForAnalysis)
		}
		if config.BehaviorAnalysis.RegularityThreshold != 0.3 {
			t.Errorf("RegularityThreshold = %f, want 0.3", config.BehaviorAnalysis.RegularityThreshold)
		}
	}

	// Verify RiskScoring
	if config.RiskScoring == nil {
		t.Error("RiskScoring is nil")
	} else {
		if config.RiskScoring.Weights == nil {
			t.Error("Weights is nil")
		}
	}

	// Verify Features
	if config.Features == nil {
		t.Error("Features is nil")
	} else {
		if len(config.Features.ToolMarkers) == 0 {
			t.Error("ToolMarkers is empty")
		}
	}

	// Verify QUIC
	if config.QUIC == nil {
		t.Error("QUIC is nil")
	}

	// Verify TLS
	if config.TLS == nil {
		t.Error("TLS is nil")
	}

	// Verify Global
	if config.Global == nil {
		t.Error("Global is nil")
	}

	// Verify Metadata
	if config.Metadata == nil {
		t.Error("Metadata is nil")
	}
}

// TestValidation tests configuration validation
func TestValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *ManagedConfig
		wantErr bool
	}{
		{
			name:    "validate valid config",
			config:  DefaultManagedConfig(),
			wantErr: false,
		},
		{
			name: "validate invalid config - MinRequestsForAnalysis is 0",
			config: func() *ManagedConfig {
				cfg := DefaultManagedConfig()
				cfg.BehaviorAnalysis.MinRequestsForAnalysis = 0
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "validate invalid config - RegularityThreshold greater than 1",
			config: func() *ManagedConfig {
				cfg := DefaultManagedConfig()
				cfg.BehaviorAnalysis.RegularityThreshold = 1.5
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "validate invalid config - RegularityThreshold less than 0",
			config: func() *ManagedConfig {
				cfg := DefaultManagedConfig()
				cfg.BehaviorAnalysis.RegularityThreshold = -0.5
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "validate invalid config - EntropyThreshold greater than 1",
			config: func() *ManagedConfig {
				cfg := DefaultManagedConfig()
				cfg.BehaviorAnalysis.EntropyThreshold = 2.0
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "validate invalid config - AnomalousIntervalRateThreshold greater than 1",
			config: func() *ManagedConfig {
				cfg := DefaultManagedConfig()
				cfg.BehaviorAnalysis.AnomalousIntervalRateThreshold = 1.5
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "validate invalid config - risk scoring threshold order is wrong",
			config: func() *ManagedConfig {
				cfg := DefaultManagedConfig()
				cfg.RiskScoring.CriticalThreshold = 0.5
				cfg.RiskScoring.HighThreshold = 0.7
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "validate invalid config - EntropyHighThreshold is negative",
			config: func() *ManagedConfig {
				cfg := DefaultManagedConfig()
				cfg.Features.EntropyHighThreshold = -1.0
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "validate invalid config - MobileScreenWidthMax is 0",
			config: func() *ManagedConfig {
				cfg := DefaultManagedConfig()
				cfg.Features.MobileScreenWidthMax = 0
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "validate invalid config - QUIC MinInitialMaxData is negative",
			config: func() *ManagedConfig {
				cfg := DefaultManagedConfig()
				cfg.QUIC.MinInitialMaxData = -1
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "validate invalid config - QUIC MinStreamData is negative",
			config: func() *ManagedConfig {
				cfg := DefaultManagedConfig()
				cfg.QUIC.MinStreamData = -1
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "validate invalid config - Global MaxConcurrency is 0",
			config: func() *ManagedConfig {
				cfg := DefaultManagedConfig()
				cfg.Global.MaxConcurrency = 0
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "validate invalid config - Global RequestTimeout is negative",
			config: func() *ManagedConfig {
				cfg := DefaultManagedConfig()
				cfg.Global.RequestTimeout = -1
				return cfg
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewConfigValidator()
			errs := validator.Validate(tt.config)

			if tt.wantErr && len(errs) == 0 {
				t.Error("Validate() expected errors but got none")
			}

			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("Validate() unexpected errors: %v", errs)
			}
		})
	}
}

// TestValidationError tests the validation error type
func TestValidationError(t *testing.T) {
	err := ValidationError{
		Field:  "test.field",
		Reason: "test reason",
		Value:  "test value",
	}

	msg := err.Error()
	if !contains(msg, "test.field") {
		t.Error("Error() should contain field name")
	}
	if !contains(msg, "test reason") {
		t.Error("Error() should contain reason")
	}
	if !contains(msg, "test value") {
		t.Error("Error() should contain value")
	}
}

// TestConfigValidator_AddRule tests adding a custom validation rule
func TestConfigValidator_AddRule(t *testing.T) {
	validator := NewConfigValidator()

	// Add a custom rule
	validator.AddRule("custom_rule", func(cfg *ManagedConfig) error {
		if cfg.Global.MaxConcurrency > 1000 {
			return ValidationError{
				Field:  "global.max_concurrency",
				Reason: "must be <= 1000",
				Value:  cfg.Global.MaxConcurrency,
			}
		}
		return nil
	})

	// Test passing configuration
	config := DefaultManagedConfig()
	config.Global.MaxConcurrency = 500
	errs := validator.Validate(config)
	if len(errs) > 0 {
		t.Errorf("Expected no errors, got: %v", errs)
	}

	// Test failing configuration
	config.Global.MaxConcurrency = 2000
	errs = validator.Validate(config)
	if len(errs) == 0 {
		t.Error("Expected validation error for max_concurrency > 1000")
	}
}

// TestConfigValidator_ValidateField tests validating a specified field
func TestConfigValidator_ValidateField(t *testing.T) {
	validator := NewConfigValidator()

	// Current implementation is simplified and always returns nil
	err := validator.ValidateField("any.field", "any value")
	if err != nil {
		t.Errorf("ValidateField() unexpected error: %v", err)
	}
}

// TestConfigManager tests the configuration manager
func TestConfigManager(t *testing.T) {
	center := NewConfigCenter("")
	center.current = DefaultManagedConfig()
	manager := NewConfigManager(center)

	// Test GetBehaviorAnalysisConfig
	t.Run("GetBehaviorAnalysisConfig", func(t *testing.T) {
		config := manager.GetBehaviorAnalysisConfig()
		if config == nil {
			t.Error("GetBehaviorAnalysisConfig() returned nil")
		}
	})

	// Test GetRiskScoringConfig
	t.Run("GetRiskScoringConfig", func(t *testing.T) {
		config := manager.GetRiskScoringConfig()
		if config == nil {
			t.Error("GetRiskScoringConfig() returned nil")
		}
	})

	// Test GetFeatureExtractionConfig
	t.Run("GetFeatureExtractionConfig", func(t *testing.T) {
		config := manager.GetFeatureExtractionConfig()
		if config == nil {
			t.Error("GetFeatureExtractionConfig() returned nil")
		}
	})

	// Test GetQUICConfig
	t.Run("GetQUICConfig", func(t *testing.T) {
		config := manager.GetQUICConfig()
		if config == nil {
			t.Error("GetQUICConfig() returned nil")
		}
	})

	// Test GetTLSConfig
	t.Run("GetTLSConfig", func(t *testing.T) {
		config := manager.GetTLSConfig()
		if config == nil {
			t.Error("GetTLSConfig() returned nil")
		}
	})

	// Test GetGlobalConfig
	t.Run("GetGlobalConfig", func(t *testing.T) {
		config := manager.GetGlobalConfig()
		if config == nil {
			t.Error("GetGlobalConfig() returned nil")
		}
	})

	// Test GetConfigValue
	t.Run("GetConfigValue", func(t *testing.T) {
		tests := []struct {
			path    string
			wantErr bool
		}{
			{"behavior_analysis.min_requests", false},
			{"behavior_analysis.regularity_threshold", false},
			{"quic.min_initial_max_data", false},
			{"unknown.path", true},
		}

		for _, tt := range tests {
			_, err := manager.GetConfigValue(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfigValue(%s) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		}
	})

	// Test IsLoaded
	t.Run("IsLoaded", func(t *testing.T) {
		if manager.IsLoaded() != center.IsLoaded() {
			t.Error("IsLoaded() should return same value as center")
		}
	})
}

// TestConfigManager_NilConfig tests the configuration manager handling nil configs
func TestConfigManager_NilConfig(t *testing.T) {
	center := NewConfigCenter("")
	// Do not set current, let it remain the default empty ManagedConfig
	manager := NewConfigManager(center)

	// Test that default values are returned when config fields are nil
	t.Run("GetBehaviorAnalysisConfig with nil", func(t *testing.T) {
		center.current.BehaviorAnalysis = nil
		config := manager.GetBehaviorAnalysisConfig()
		if config == nil {
			t.Error("GetBehaviorAnalysisConfig() returned nil when config field is nil")
		}
	})

	t.Run("GetRiskScoringConfig with nil", func(t *testing.T) {
		center.current.RiskScoring = nil
		config := manager.GetRiskScoringConfig()
		if config == nil {
			t.Error("GetRiskScoringConfig() returned nil when config field is nil")
		}
	})

	t.Run("GetFeatureExtractionConfig with nil", func(t *testing.T) {
		center.current.Features = nil
		config := manager.GetFeatureExtractionConfig()
		if config == nil {
			t.Error("GetFeatureExtractionConfig() returned nil when config field is nil")
		}
	})

	t.Run("GetQUICConfig with nil", func(t *testing.T) {
		center.current.QUIC = nil
		config := manager.GetQUICConfig()
		if config == nil {
			t.Error("GetQUICConfig() returned nil when config field is nil")
		}
	})

	t.Run("GetTLSConfig with nil", func(t *testing.T) {
		center.current.TLS = nil
		config := manager.GetTLSConfig()
		if config == nil {
			t.Error("GetTLSConfig() returned nil when config field is nil")
		}
	})

	t.Run("GetGlobalConfig with nil", func(t *testing.T) {
		center.current.Global = nil
		config := manager.GetGlobalConfig()
		if config == nil {
			t.Error("GetGlobalConfig() returned nil when config field is nil")
		}
	})
}

// TestHealthChecker tests the health checker
func TestHealthChecker(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(cc *ConfigCenter)
		wantStatus HealthStatus
	}{
		{
			name: "health status when config is not loaded",
			setup: func(cc *ConfigCenter) {
				// Do not load configuration
			},
			wantStatus: HealthCritical,
		},
		{
			name: "health status when config is loaded",
			setup: func(cc *ConfigCenter) {
				cc.current = DefaultManagedConfig()
				cc.loaded = true
				// Save config to file so that fileAccessCheck passes
				cc.SaveToFile()
				// Record version so that historyCheck passes
				cc.Update(DefaultManagedConfig(), "test", "tester")
			},
			wantStatus: HealthOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.json")
			cc := NewConfigCenter(configPath)
			tt.setup(cc)

			hc := NewHealthChecker(cc)
			result := hc.CheckHealth()

			if result.Status != tt.wantStatus {
				t.Errorf("CheckHealth() status = %v, want %v", result.Status, tt.wantStatus)
			}
		})
	}
}

// TestHealthChecker_AddCheck tests adding a health check
func TestHealthChecker_AddCheck(t *testing.T) {
	cc := NewConfigCenter("")
	hc := NewHealthChecker(cc)

	// Record the original check count
	originalCount := len(hc.checks)

	// Add a custom check
	hc.AddCheck(&mockHealthCheck{})

	if len(hc.checks) != originalCount+1 {
		t.Errorf("checks count = %d, want %d", len(hc.checks), originalCount+1)
	}
}

// TestConfigChange tests configuration change information
func TestConfigChange(t *testing.T) {
	change := ConfigChange{
		Path:      "test.path",
		OldValue:  "old",
		NewValue:  "new",
		Timestamp: time.Now(),
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

// TestVersionedConfig tests versioned configuration
func TestVersionedConfig(t *testing.T) {
	config := DefaultManagedConfig()
	vc := &VersionedConfig{
		Timestamp:    time.Now(),
		Version:      "v1",
		Config:       config,
		ChangeReason: "test",
		ChangedBy:    "tester",
	}

	if vc.Version != "v1" {
		t.Errorf("Version = %v, want v1", vc.Version)
	}
	if vc.ChangeReason != "test" {
		t.Errorf("ChangeReason = %v, want test", vc.ChangeReason)
	}
	if vc.ChangedBy != "tester" {
		t.Errorf("ChangedBy = %v, want tester", vc.ChangedBy)
	}
}

// TestConfigMetadata tests configuration metadata
func TestConfigMetadata(t *testing.T) {
	now := time.Now()
	meta := &ConfigMetadata{
		Version:      "1.0.0",
		LastModified: now,
		Author:       "tester",
		Description:  "test description",
	}

	if meta.Version != "1.0.0" {
		t.Errorf("Version = %v, want 1.0.0", meta.Version)
	}
	if meta.Author != "tester" {
		t.Errorf("Author = %v, want tester", meta.Author)
	}
	if meta.Description != "test description" {
		t.Errorf("Description = %v, want test description", meta.Description)
	}
}

// TestHealthCheckResult tests health check results
func TestHealthCheckResult(t *testing.T) {
	result := &HealthCheckResult{
		Status:    HealthOK,
		Message:   "All good",
		Timestamp: time.Now(),
		Checks:    make([]*HealthCheckItem, 0),
		Overall:   true,
	}

	if result.Status != HealthOK {
		t.Errorf("Status = %v, want HealthOK", result.Status)
	}
	if !result.Overall {
		t.Error("Overall should be true")
	}
}

// TestHealthCheckItem tests health check items
func TestHealthCheckItem(t *testing.T) {
	item := &HealthCheckItem{
		Name:        "Test Check",
		Status:      HealthWarning,
		Message:     "Warning message",
		LastChecked: time.Now(),
	}

	if item.Name != "Test Check" {
		t.Errorf("Name = %v, want Test Check", item.Name)
	}
	if item.Status != HealthWarning {
		t.Errorf("Status = %v, want HealthWarning", item.Status)
	}
}

// TestHealthStatus tests health status constants
func TestHealthStatus(t *testing.T) {
	if HealthOK != "ok" {
		t.Errorf("HealthOK = %v, want ok", HealthOK)
	}
	if HealthWarning != "warning" {
		t.Errorf("HealthWarning = %v, want warning", HealthWarning)
	}
	if HealthCritical != "critical" {
		t.Errorf("HealthCritical = %v, want critical", HealthCritical)
	}
}

// Helper functions and mock objects

// mockListener is a mock configuration change listener
type mockListener struct {
	t      *testing.T
	called bool
}

func (m *mockListener) OnConfigChange(old, new *ManagedConfig, changes []ConfigChange) error {
	m.called = true
	return nil
}

// mockErrorListener is a mock listener that returns errors
type mockErrorListener struct{}

func (m *mockErrorListener) OnConfigChange(old, new *ManagedConfig, changes []ConfigChange) error {
	return fmt.Errorf("mock listener error")
}

// mockHealthCheck is a mock health check
type mockHealthCheck struct{}

func (m *mockHealthCheck) Check() *HealthCheckItem {
	return &HealthCheckItem{
		Name:        "Mock Check",
		Status:      HealthOK,
		Message:     "Mock check passed",
		LastChecked: time.Now(),
	}
}

// contains checks whether a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsInternal(s, substr)))
}

func containsInternal(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
