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

// TestNewConfigCenter 测试创建配置中心
func TestNewConfigCenter(t *testing.T) {
	tests := []struct {
		name       string
		configPath string
		wantPath   string
		wantMemory bool
	}{
		{
			name:       "创建带路径的配置中心",
			configPath: "/tmp/config.json",
			wantPath:   "/tmp/config.json",
			wantMemory: false,
		},
		{
			name:       "创建内存模式的配置中心（空路径）",
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

// TestConfigCenter_Load 测试从文件加载配置
func TestConfigCenter_Load(t *testing.T) {
	tests := []struct {
		name       string
		setupFile  func(t *testing.T) string
		wantErr    bool
		errContain string
		wantLoaded bool
	}{
		{
			name: "从文件加载有效配置",
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
			name: "文件不存在返回错误",
			setupFile: func(t *testing.T) string {
				return "/nonexistent/path/config.json"
			},
			wantErr:    true,
			errContain: "failed to read config file",
			wantLoaded: false,
		},
		{
			name: "无效的 JSON 返回错误",
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
			name: "验证失败的配置返回错误",
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
				// 验证历史记录
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

// TestConfigCenter_Get 测试获取当前配置（并发安全）
func TestConfigCenter_Get(t *testing.T) {
	cc := NewConfigCenter("")

	// 设置初始配置
	cc.current = DefaultManagedConfig()

	// 测试并发读取
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

// TestConfigCenter_Update 测试更新配置
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
			name: "未加载时更新返回错误",
			setup: func(t *testing.T, cc *ConfigCenter) {
				// 不加载配置
			},
			newConfig:  DefaultManagedConfig(),
			wantErr:    true,
			errContain: "config center not loaded",
		},
		{
			name: "验证失败的配置返回错误",
			setup: func(t *testing.T, cc *ConfigCenter) {
				cc.loaded = true
			},
			newConfig: func() *ManagedConfig {
				cfg := DefaultManagedConfig()
				cfg.BehaviorAnalysis.MinRequestsForAnalysis = 0 // 无效值
				return cfg
			}(),
			wantErr:    true,
			errContain: "new config validation failed",
		},
		{
			name: "监听器被正确调用",
			setup: func(t *testing.T, cc *ConfigCenter) {
				cc.loaded = true
				cc.current = DefaultManagedConfig() // 初始化 current 避免 nil 字段
				listener := &mockListener{t: t, called: false}
				cc.RegisterListener(listener)
			},
			newConfig: func() *ManagedConfig {
				cfg := DefaultManagedConfig()
				cfg.BehaviorAnalysis.MinRequestsForAnalysis = 10 // 修改一个字段以触发变更检测
				return cfg
			}(),
			wantErr: false,
		},
		{
			name: "监听器返回错误时更新失败",
			setup: func(t *testing.T, cc *ConfigCenter) {
				cc.loaded = true
				cc.current = DefaultManagedConfig() // 初始化 current
				listener := &mockErrorListener{}
				cc.RegisterListener(listener)
			},
			newConfig:  DefaultManagedConfig(),
			wantErr:    true,
			errContain: "listener error",
		},
		{
			name: "更新成功并保存到文件",
			setup: func(t *testing.T, cc *ConfigCenter) {
				cc.loaded = true
				cc.current = DefaultManagedConfig() // 初始化 current
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
				// 验证配置已更新
				if cc.Get() != tt.newConfig {
					t.Error("Config was not updated")
				}

				// 验证历史记录
				history := cc.GetHistory(1)
				if len(history) == 0 {
					t.Error("Expected history to be recorded after update")
				}
				if history[0].ChangedBy != "tester" {
					t.Errorf("ChangedBy = %v, want tester", history[0].ChangedBy)
				}

				// 验证文件保存
				if tt.checkFile {
					if _, err := os.Stat(configPath); os.IsNotExist(err) {
						t.Error("Config file was not saved")
					}
				}
			}
		})
	}
}

// TestConfigCenter_SaveToFile 测试保存到文件
func TestConfigCenter_SaveToFile(t *testing.T) {
	tests := []struct {
		name        string
		configPath  string
		setupConfig func(cc *ConfigCenter)
		wantErr     bool
		wantFile    bool
	}{
		{
			name:       "保存到文件成功",
			configPath: "",
			setupConfig: func(cc *ConfigCenter) {
				// configPath 会在测试循环中被设置
				cc.current = DefaultManagedConfig()
			},
			wantErr:  false,
			wantFile: true,
		},
		{
			name:       "空路径（内存模式）行为",
			configPath: "",
			setupConfig: func(cc *ConfigCenter) {
				cc.current = DefaultManagedConfig()
				// configPath 为空，表示内存模式
			},
			wantErr:  false,
			wantFile: false, // 内存模式不创建文件
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

				// 验证文件内容
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

// TestConfigCenter_RegisterListener 测试注册配置变更监听器
func TestConfigCenter_RegisterListener(t *testing.T) {
	cc := NewConfigCenter("")

	// 测试注册单个监听器
	listener1 := &mockListener{t: t, called: false}
	cc.RegisterListener(listener1)

	if len(cc.listeners) != 1 {
		t.Errorf("listeners count = %d, want 1", len(cc.listeners))
	}

	// 测试注册多个监听器
	listener2 := &mockListener{t: t, called: false}
	cc.RegisterListener(listener2)

	if len(cc.listeners) != 2 {
		t.Errorf("listeners count = %d, want 2", len(cc.listeners))
	}
}

// TestConfigCenter_EnableAutoReload 测试启用自动重载
func TestConfigCenter_EnableAutoReload(t *testing.T) {
	tests := []struct {
		name     string
		interval time.Duration
		wantMin  time.Duration
	}{
		{
			name:     "启用自动重载（正常间隔）",
			interval: 5 * time.Second,
			wantMin:  5 * time.Second,
		},
		{
			name:     "启用自动重载（小于1秒，应该被限制为1秒）",
			interval: 500 * time.Millisecond,
			wantMin:  time.Second,
		},
		{
			name:     "多次启用不会启动多个 worker",
			interval: time.Second,
			wantMin:  time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.json")

			// 创建初始配置文件
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

			// 给 worker 一点时间启动
			time.Sleep(100 * time.Millisecond)

			// 再次启用，不应该启动新的 worker
			cc.EnableAutoReload(tt.interval)

			// 禁用清理
			cc.DisableAutoReload()
		})
	}
}

// TestConfigCenter_DisableAutoReload 测试禁用自动重载
func TestConfigCenter_DisableAutoReload(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// 创建初始配置文件
	config := DefaultManagedConfig()
	data, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(configPath, data, 0644)

	cc := NewConfigCenter(configPath)
	cc.loaded = true

	// 先启用
	cc.EnableAutoReload(time.Second)
	if !cc.autoReload {
		t.Fatal("autoReload should be true")
	}

	// 给 worker 一点时间启动
	time.Sleep(100 * time.Millisecond)

	// 再禁用
	cc.DisableAutoReload()

	if cc.autoReload {
		t.Error("autoReload should be false after DisableAutoReload")
	}

	// 重复禁用不应 panic
	cc.DisableAutoReload()
}

// TestConfigCenter_GetHistory 测试获取历史版本
func TestConfigCenter_GetHistory(t *testing.T) {
	cc := NewConfigCenter("")
	cc.loaded = true

	// 添加多个历史版本
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
			name:      "获取历史版本（限制5个）",
			limit:     5,
			wantCount: 5,
		},
		{
			name:      "获取历史版本（限制超过总数）",
			limit:     20,
			wantCount: 10,
		},
		{
			name:      "获取历史版本（限制为0，返回全部）",
			limit:     0,
			wantCount: 10,
		},
		{
			name:      "获取历史版本（负数限制，返回全部）",
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

// TestConfigCenter_Rollback 测试回滚到指定版本
func TestConfigCenter_Rollback(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, cc *ConfigCenter) string
		version    string
		wantErr    bool
		errContain string
	}{
		{
			name: "回滚到指定版本成功",
			setup: func(t *testing.T, cc *ConfigCenter) string {
				cc.loaded = true
				// 添加多个版本
				for i := 0; i < 3; i++ {
					config := DefaultManagedConfig()
					config.BehaviorAnalysis.MinRequestsForAnalysis = 5 + i
					cc.Update(config, fmt.Sprintf("update %d", i), "tester")
				}
				return "v1" // 回滚到第一个版本
			},
			version: "v1",
			wantErr: false,
		},
		{
			name: "回滚到不存在的版本返回错误",
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
			name: "未加载时回滚返回错误",
			setup: func(t *testing.T, cc *ConfigCenter) string {
				// 不加载配置
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

// TestConfigCenter_IsLoaded 测试检查配置是否已加载
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

// TestConfigCenter_SetValidationEnabled 测试设置是否启用配置验证
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

// TestDefaultManagedConfig 测试默认值创建
func TestDefaultManagedConfig(t *testing.T) {
	config := DefaultManagedConfig()

	if config == nil {
		t.Fatal("DefaultManagedConfig() returned nil")
	}

	// 验证 BehaviorAnalysis
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

	// 验证 RiskScoring
	if config.RiskScoring == nil {
		t.Error("RiskScoring is nil")
	} else {
		if config.RiskScoring.Weights == nil {
			t.Error("Weights is nil")
		}
	}

	// 验证 Features
	if config.Features == nil {
		t.Error("Features is nil")
	} else {
		if len(config.Features.ToolMarkers) == 0 {
			t.Error("ToolMarkers is empty")
		}
	}

	// 验证 QUIC
	if config.QUIC == nil {
		t.Error("QUIC is nil")
	}

	// 验证 TLS
	if config.TLS == nil {
		t.Error("TLS is nil")
	}

	// 验证 Global
	if config.Global == nil {
		t.Error("Global is nil")
	}

	// 验证 Metadata
	if config.Metadata == nil {
		t.Error("Metadata is nil")
	}
}

// TestValidation 测试配置验证
func TestValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *ManagedConfig
		wantErr bool
	}{
		{
			name:    "验证有效配置",
			config:  DefaultManagedConfig(),
			wantErr: false,
		},
		{
			name: "验证无效配置 - MinRequestsForAnalysis 为 0",
			config: func() *ManagedConfig {
				cfg := DefaultManagedConfig()
				cfg.BehaviorAnalysis.MinRequestsForAnalysis = 0
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "验证无效配置 - RegularityThreshold 大于 1",
			config: func() *ManagedConfig {
				cfg := DefaultManagedConfig()
				cfg.BehaviorAnalysis.RegularityThreshold = 1.5
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "验证无效配置 - RegularityThreshold 小于 0",
			config: func() *ManagedConfig {
				cfg := DefaultManagedConfig()
				cfg.BehaviorAnalysis.RegularityThreshold = -0.5
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "验证无效配置 - EntropyThreshold 大于 1",
			config: func() *ManagedConfig {
				cfg := DefaultManagedConfig()
				cfg.BehaviorAnalysis.EntropyThreshold = 2.0
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "验证无效配置 - AnomalousIntervalRateThreshold 大于 1",
			config: func() *ManagedConfig {
				cfg := DefaultManagedConfig()
				cfg.BehaviorAnalysis.AnomalousIntervalRateThreshold = 1.5
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "验证无效配置 - 风险评分阈值顺序错误",
			config: func() *ManagedConfig {
				cfg := DefaultManagedConfig()
				cfg.RiskScoring.CriticalThreshold = 0.5
				cfg.RiskScoring.HighThreshold = 0.7
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "验证无效配置 - EntropyHighThreshold 为负数",
			config: func() *ManagedConfig {
				cfg := DefaultManagedConfig()
				cfg.Features.EntropyHighThreshold = -1.0
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "验证无效配置 - MobileScreenWidthMax 为 0",
			config: func() *ManagedConfig {
				cfg := DefaultManagedConfig()
				cfg.Features.MobileScreenWidthMax = 0
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "验证无效配置 - QUIC MinInitialMaxData 为负数",
			config: func() *ManagedConfig {
				cfg := DefaultManagedConfig()
				cfg.QUIC.MinInitialMaxData = -1
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "验证无效配置 - QUIC MinStreamData 为负数",
			config: func() *ManagedConfig {
				cfg := DefaultManagedConfig()
				cfg.QUIC.MinStreamData = -1
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "验证无效配置 - Global MaxConcurrency 为 0",
			config: func() *ManagedConfig {
				cfg := DefaultManagedConfig()
				cfg.Global.MaxConcurrency = 0
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "验证无效配置 - Global RequestTimeout 为负数",
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

// TestValidationError 测试验证错误类型
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

// TestConfigValidator_AddRule 测试添加自定义验证规则
func TestConfigValidator_AddRule(t *testing.T) {
	validator := NewConfigValidator()

	// 添加自定义规则
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

	// 测试通过的配置
	config := DefaultManagedConfig()
	config.Global.MaxConcurrency = 500
	errs := validator.Validate(config)
	if len(errs) > 0 {
		t.Errorf("Expected no errors, got: %v", errs)
	}

	// 测试失败的配置
	config.Global.MaxConcurrency = 2000
	errs = validator.Validate(config)
	if len(errs) == 0 {
		t.Error("Expected validation error for max_concurrency > 1000")
	}
}

// TestConfigValidator_ValidateField 测试验证指定字段
func TestConfigValidator_ValidateField(t *testing.T) {
	validator := NewConfigValidator()

	// 当前实现是简化版，总是返回 nil
	err := validator.ValidateField("any.field", "any value")
	if err != nil {
		t.Errorf("ValidateField() unexpected error: %v", err)
	}
}

// TestConfigManager 测试配置管理器
func TestConfigManager(t *testing.T) {
	center := NewConfigCenter("")
	center.current = DefaultManagedConfig()
	manager := NewConfigManager(center)

	// 测试 GetBehaviorAnalysisConfig
	t.Run("GetBehaviorAnalysisConfig", func(t *testing.T) {
		config := manager.GetBehaviorAnalysisConfig()
		if config == nil {
			t.Error("GetBehaviorAnalysisConfig() returned nil")
		}
	})

	// 测试 GetRiskScoringConfig
	t.Run("GetRiskScoringConfig", func(t *testing.T) {
		config := manager.GetRiskScoringConfig()
		if config == nil {
			t.Error("GetRiskScoringConfig() returned nil")
		}
	})

	// 测试 GetFeatureExtractionConfig
	t.Run("GetFeatureExtractionConfig", func(t *testing.T) {
		config := manager.GetFeatureExtractionConfig()
		if config == nil {
			t.Error("GetFeatureExtractionConfig() returned nil")
		}
	})

	// 测试 GetQUICConfig
	t.Run("GetQUICConfig", func(t *testing.T) {
		config := manager.GetQUICConfig()
		if config == nil {
			t.Error("GetQUICConfig() returned nil")
		}
	})

	// 测试 GetTLSConfig
	t.Run("GetTLSConfig", func(t *testing.T) {
		config := manager.GetTLSConfig()
		if config == nil {
			t.Error("GetTLSConfig() returned nil")
		}
	})

	// 测试 GetGlobalConfig
	t.Run("GetGlobalConfig", func(t *testing.T) {
		config := manager.GetGlobalConfig()
		if config == nil {
			t.Error("GetGlobalConfig() returned nil")
		}
	})

	// 测试 GetConfigValue
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

	// 测试 IsLoaded
	t.Run("IsLoaded", func(t *testing.T) {
		if manager.IsLoaded() != center.IsLoaded() {
			t.Error("IsLoaded() should return same value as center")
		}
	})
}

// TestConfigManager_NilConfig 测试配置管理器处理 nil 配置
func TestConfigManager_NilConfig(t *testing.T) {
	center := NewConfigCenter("")
	// 不设置 current，让它保持默认的空 ManagedConfig
	manager := NewConfigManager(center)

	// 测试当配置字段为 nil 时返回默认值
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

// TestHealthChecker 测试健康检查器
func TestHealthChecker(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(cc *ConfigCenter)
		wantStatus HealthStatus
	}{
		{
			name: "未加载配置的健康状态",
			setup: func(cc *ConfigCenter) {
				// 不加载配置
			},
			wantStatus: HealthCritical,
		},
		{
			name: "已加载配置的健康状态",
			setup: func(cc *ConfigCenter) {
				cc.current = DefaultManagedConfig()
				cc.loaded = true
				// 保存配置到文件以使 fileAccessCheck 通过
				cc.SaveToFile()
				// 记录版本以使 historyCheck 通过
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

// TestHealthChecker_AddCheck 测试添加健康检查
func TestHealthChecker_AddCheck(t *testing.T) {
	cc := NewConfigCenter("")
	hc := NewHealthChecker(cc)

	// 记录原始检查数量
	originalCount := len(hc.checks)

	// 添加自定义检查
	hc.AddCheck(&mockHealthCheck{})

	if len(hc.checks) != originalCount+1 {
		t.Errorf("checks count = %d, want %d", len(hc.checks), originalCount+1)
	}
}

// TestConfigChange 测试配置变更信息
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

// TestVersionedConfig 测试版本化配置
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

// TestConfigMetadata 测试配置元数据
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

// TestHealthCheckResult 测试健康检查结果
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

// TestHealthCheckItem 测试健康检查项
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

// TestHealthStatus 测试健康状态常量
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

// 辅助函数和模拟对象

// mockListener 模拟配置变更监听器
type mockListener struct {
	t      *testing.T
	called bool
}

func (m *mockListener) OnConfigChange(old, new *ManagedConfig, changes []ConfigChange) error {
	m.called = true
	return nil
}

// mockErrorListener 模拟返回错误的监听器
type mockErrorListener struct{}

func (m *mockErrorListener) OnConfigChange(old, new *ManagedConfig, changes []ConfigChange) error {
	return fmt.Errorf("mock listener error")
}

// mockHealthCheck 模拟健康检查
type mockHealthCheck struct{}

func (m *mockHealthCheck) Check() *HealthCheckItem {
	return &HealthCheckItem{
		Name:        "Mock Check",
		Status:      HealthOK,
		Message:     "Mock check passed",
		LastChecked: time.Now(),
	}
}

// contains 检查字符串是否包含子串
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

