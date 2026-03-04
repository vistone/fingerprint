package config

import (
	"fmt"
	"sync"
	"time"
)

// UnifiedConfigManager 统一的配置管理器
// 合并了 ConfigCenter、ConfigManager 和 EnhancedConfigCenter 的功能
type UnifiedConfigManager struct {
	// 核心配置中心（内嵌，直接访问）
	*ConfigCenter

	// 可选的增强功能
	enhanced *enhancedFeatures

	// 保护 enhanced 字段的互斥锁
	mu sync.RWMutex
}

// enhancedFeatures 增强功能（可选）
type enhancedFeatures struct {
	broadcastCh   chan ConfigChangeEvent
	subscribers   map[string]chan ConfigChangeEvent
	subscriberMu  sync.RWMutex
	healthChecker *ConfigHealthChecker
}

// NewUnifiedConfigManager 创建统一的配置管理器
func NewUnifiedConfigManager(configPath string) *UnifiedConfigManager {
	return &UnifiedConfigManager{
		ConfigCenter: NewConfigCenter(configPath),
	}
}

// EnableEnhancedFeatures 启用增强功能（事件订阅、健康检查）
func (ucm *UnifiedConfigManager) EnableEnhancedFeatures() {
	if ucm.enhanced != nil {
		return // 已启用
	}

	ucm.enhanced = &enhancedFeatures{
		broadcastCh: make(chan ConfigChangeEvent, 100),
		subscribers: make(map[string]chan ConfigChangeEvent),
	}

	// 初始化健康检查器
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

	// 启动广播处理器
	go ucm.broadcastProcessor()

	// 启动健康检查
	go ucm.enhanced.healthChecker.start(ucm.enhanced.broadcastCh)
}

// DisableEnhancedFeatures 禁用增强功能
func (ucm *UnifiedConfigManager) DisableEnhancedFeatures() {
	ucm.mu.Lock()
	defer ucm.mu.Unlock()
	
	if ucm.enhanced == nil {
		return
	}

	// 停止健康检查
	if ucm.enhanced.healthChecker != nil {
		ucm.enhanced.healthChecker.stop()
	}

	// 关闭广播通道（如果还没关闭）
	if ucm.enhanced.broadcastCh != nil {
		select {
		case <-ucm.enhanced.broadcastCh:
			// 已经关闭
		default:
			close(ucm.enhanced.broadcastCh)
		}
	}

	// 关闭所有订阅者通道
	ucm.enhanced.subscriberMu.Lock()
	for _, ch := range ucm.enhanced.subscribers {
		select {
		case <-ch:
			// 已经关闭
		default:
			close(ch)
		}
	}
	ucm.enhanced.subscribers = make(map[string]chan ConfigChangeEvent)
	ucm.enhanced.subscriberMu.Unlock()

	ucm.enhanced = nil
}

// Subscribe 订阅配置变更事件（需要启用增强功能）
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

	// 发送订阅确认事件
	ucm.enhanced.broadcastCh <- ConfigChangeEvent{
		Type:        ConfigChangeTypeSubscribe,
		Timestamp:   time.Now(),
		Description: fmt.Sprintf("Subscriber %s registered", subscriberID),
		Source:      subscriberID,
	}

	return eventCh, nil
}

// Unsubscribe 取消订阅
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

// broadcastProcessor 广播事件处理器
func (ucm *UnifiedConfigManager) broadcastProcessor() {
	for {
		// 检查 enhanced 是否存在
		ucm.mu.RLock()
		enhanced := ucm.enhanced
		ucm.mu.RUnlock()
		
		if enhanced == nil {
			return
		}
		
		// 安全的从广播通道读取事件
		event, ok := <-enhanced.broadcastCh
		if !ok {
			// 通道已关闭，退出
			return
		}
		
		// 复制订阅者列表（避免在发送时持有锁）
		enhanced.subscriberMu.RLock()
		subscribers := make(map[string]chan ConfigChangeEvent, len(enhanced.subscribers))
		for k, v := range enhanced.subscribers {
			subscribers[k] = v
		}
		enhanced.subscriberMu.RUnlock()

		// 异步发送给所有订阅者
		for subscriberID, ch := range subscribers {
			select {
			case ch <- event:
				// 发送成功
			default:
				// 通道已满或关闭，跳过
				_ = subscriberID
			}
		}
	}
}

// Update 重写更新方法，添加事件广播
func (ucm *UnifiedConfigManager) Update(newConfig *ManagedConfig, reason, changedBy string) error {
	// 调用基类方法
	if err := ucm.ConfigCenter.Update(newConfig, reason, changedBy); err != nil {
		return err
	}

	// 如果启用了增强功能，广播更新事件
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

// GetHealthStatus 获取健康状态
func (ucm *UnifiedConfigManager) GetHealthStatus() (ConfigHealthStatus, error) {
	if ucm.enhanced == nil {
		return ConfigHealthStatus{}, fmt.Errorf("enhanced features not enabled")
	}

	ucm.enhanced.healthChecker.mu.RLock()
	defer ucm.enhanced.healthChecker.mu.RUnlock()
	return ucm.enhanced.healthChecker.lastStatus, nil
}

// AddHealthCheck 添加健康检查函数
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
// 便捷的获取配置方法（原 ConfigManager 功能）
// ============================================

// GetBehaviorAnalysisConfig 获取行为分析配置
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

// GetRiskScoringConfig 获取风险评分配置
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

// GetFeatureExtractionConfig 获取特征提取配置
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

// GetQUICConfig 获取 QUIC 配置
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

// GetTLSConfig 获取 TLS 配置
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

// GetGlobalConfig 获取全局配置
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
// 便捷的更新配置方法
// ============================================

// UpdateBehaviorAnalysisConfig 更新行为分析配置
func (ucm *UnifiedConfigManager) UpdateBehaviorAnalysisConfig(newConfig *BehaviorAnalysisConfig, reason, changedBy string) error {
	config := ucm.ConfigCenter.Get()
	config.BehaviorAnalysis = newConfig
	return ucm.Update(config, reason, changedBy)
}

// UpdateRiskScoringConfig 更新风险评分配置
func (ucm *UnifiedConfigManager) UpdateRiskScoringConfig(newConfig *RiskScoringConfig, reason, changedBy string) error {
	config := ucm.ConfigCenter.Get()
	config.RiskScoring = newConfig
	return ucm.Update(config, reason, changedBy)
}

// UpdateFeatureExtractionConfig 更新特征提取配置
func (ucm *UnifiedConfigManager) UpdateFeatureExtractionConfig(newConfig *FeatureExtractionConfig, reason, changedBy string) error {
	config := ucm.ConfigCenter.Get()
	config.Features = newConfig
	return ucm.Update(config, reason, changedBy)
}

// ============================================
// 兼容性适配层
// ============================================


