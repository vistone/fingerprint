package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ConfigCenter 配置中心 - 集中管理所有配置
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
	reloadOnce        sync.Once // 确保 autoReloadWorker 只启动一次
	validationEnabled bool
}

// ManagedConfig 被管理的配置对象
type ManagedConfig struct {
	// 行为分析配置
	BehaviorAnalysis *BehaviorAnalysisConfig `json:"behavior_analysis"`

	// 风险评分配置
	RiskScoring *RiskScoringConfig `json:"risk_scoring"`

	// 特征提取配置
	Features *FeatureExtractionConfig `json:"features"`

	// QUIC 配置
	QUIC *QUICConfig `json:"quic"`

	// TLS 配置
	TLS *TLSConfig `json:"tls"`

	// 全局限制和阈值
	Global *GlobalConfig `json:"global"`

	// 元数据
	Metadata *ConfigMetadata `json:"metadata"`
}

// ConfigMetadata 配置元数据
type ConfigMetadata struct {
	Version      string    `json:"version"`
	LastModified time.Time `json:"last_modified"`
	Author       string    `json:"author"`
	Description  string    `json:"description"`
}

// VersionedConfig 带版本信息的配置
type VersionedConfig struct {
	Timestamp    time.Time
	Version      string
	Config       *ManagedConfig
	ChangeReason string
	ChangedBy    string
}

// ConfigChangeListener 配置变更监听器接口
type ConfigChangeListener interface {
	// OnConfigChange 当配置变更时被调用
	OnConfigChange(old, new *ManagedConfig, changes []ConfigChange) error
}

// ConfigChange 配置变更信息
type ConfigChange struct {
	Path      string      // 变更的配置路径（如 "behavior_analysis.min_requests"）
	OldValue  interface{} // 旧值
	NewValue  interface{} // 新值
	Timestamp time.Time
}

// NewConfigCenter 创建新的配置中心
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

// Load 从文件加载配置
func (cc *ConfigCenter) Load() error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	data, err := os.ReadFile(cc.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 验证 JSON 格式
	var config ManagedConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// 验证配置
	if cc.validationEnabled {
		if errs := cc.validateConfig(&config); len(errs) > 0 {
			return fmt.Errorf("config validation failed: %v", errs)
		}
	}

	// 初始化元数据
	if config.Metadata == nil {
		config.Metadata = &ConfigMetadata{}
	}

	// 获取文件修改时间
	fileInfo, _ := os.Stat(cc.configPath)
	if fileInfo != nil {
		cc.lastModTime = fileInfo.ModTime()
		config.Metadata.LastModified = cc.lastModTime
	}

	// 保存到历史
	cc.recordVersion(&config, "initial_load", "system")

	cc.current = &config
	cc.loaded = true

	return nil
}

// Get 获取当前配置（只读）
func (cc *ConfigCenter) Get() *ManagedConfig {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.current
}

// Update 更新配置
func (cc *ConfigCenter) Update(newConfig *ManagedConfig, reason, changedBy string) error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	if !cc.loaded {
		return fmt.Errorf("config center not loaded")
	}

	// 验证新配置
	if cc.validationEnabled {
		if errs := cc.validateConfig(newConfig); len(errs) > 0 {
			return fmt.Errorf("new config validation failed: %v", errs)
		}
	}

	// 检测变更
	changes := cc.detectChanges(cc.current, newConfig)

	// 通知监听器
	for _, listener := range cc.listeners {
		if err := listener.OnConfigChange(cc.current, newConfig, changes); err != nil {
			return fmt.Errorf("listener error: %w", err)
		}
	}

	// 保存旧配置
	oldConfig := cc.current

	// 更新当前配置
	cc.current = newConfig
	cc.recordVersion(newConfig, reason, changedBy)

	// 保存到文件
	if err := cc.SaveToFile(); err != nil {
		// 回滚到旧配置
		cc.current = oldConfig
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// SaveToFile 保存配置到文件
func (cc *ConfigCenter) SaveToFile() error {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	data, err := json.MarshalIndent(cc.current, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 创建目录（如果不存在）
	dir := filepath.Dir(cc.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(cc.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// RegisterListener 注册配置变更监听器
func (cc *ConfigCenter) RegisterListener(listener ConfigChangeListener) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.listeners = append(cc.listeners, listener)
}

// EnableAutoReload 启用自动重新加载
func (cc *ConfigCenter) EnableAutoReload(interval time.Duration) {
	cc.mu.Lock()
	cc.autoReload = true
	if interval < time.Second {
		interval = time.Second // 最小间隔 1 秒，防止过于频繁的检查
	}
	cc.reloadInterval = interval
	cc.mu.Unlock()

	// 使用 sync.Once 确保只启动一个 autoReloadWorker
	cc.reloadOnce.Do(func() {
		go cc.autoReloadWorker()
	})
}

// DisableAutoReload 禁用自动重新加载
func (cc *ConfigCenter) DisableAutoReload() {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	if cc.autoReload {
		cc.autoReload = false
		close(cc.stopReloadChan)
	}
}

// autoReloadWorker 自动重新加载 worker
func (cc *ConfigCenter) autoReloadWorker() {
	for {
		// 获取当前的停止通道和间隔
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
						// 通知监听器配置已重新加载
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

// GetHistory 获取配置版本历史
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

// Rollback 回滚到指定版本
func (cc *ConfigCenter) Rollback(version string, reason, rolledBackBy string) error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	// 查找指定版本
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

	// 创建配置副本
	oldConfig := cc.current
	newConfig := cc.copyConfig(targetVersion.Config)

	// 检测变更
	changes := cc.detectChanges(cc.current, newConfig)

	// 通知监听器
	for _, listener := range cc.listeners {
		if err := listener.OnConfigChange(cc.current, newConfig, changes); err != nil {
			return fmt.Errorf("listener error: %w", err)
		}
	}

	// 更新当前配置
	cc.current = newConfig
	cc.recordVersion(newConfig, reason+" (rolled back from "+version+")", rolledBackBy)

	// 保存到文件
	if err := cc.SaveToFile(); err != nil {
		// 回滚到旧配置
		cc.current = oldConfig
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// recordVersion 记录配置版本
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

	// 限制历史大小
	if len(cc.history) > cc.maxHistorySize {
		cc.history = cc.history[len(cc.history)-cc.maxHistorySize:]
	}
}

// detectChanges 检测配置变更
func (cc *ConfigCenter) detectChanges(old, new *ManagedConfig) []ConfigChange {
	// 简化实现 - 实际应该使用反射进行深层比较
	changes := make([]ConfigChange, 0)
	if old == nil {
		return changes
	}

	// 行为分析配置变更
	if old.BehaviorAnalysis != new.BehaviorAnalysis {
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

// copyConfig 深复制配置
// 使用手写的 Clone 方法，性能优于 JSON 序列化
func (cc *ConfigCenter) copyConfig(config *ManagedConfig) *ManagedConfig {
	return config.Clone()
}

// validateConfig 验证配置（基础验证）
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

// IsLoaded 检查配置是否已加载
func (cc *ConfigCenter) IsLoaded() bool {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.loaded
}

// SetValidationEnabled 设置是否启用配置验证
func (cc *ConfigCenter) SetValidationEnabled(enabled bool) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.validationEnabled = enabled
}
