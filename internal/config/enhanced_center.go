package config

import (
	"fmt"
	"sync"
	"time"

	"github.com/vistone/fingerprint/internal/logger"
)

// EnhancedConfigCenter 配置中心增强包装器
type EnhancedConfigCenter struct {
	center         *ConfigCenter
	broadcastCh    chan ConfigChangeEvent
	subscribers    map[string]chan ConfigChangeEvent
	subscriberMu   sync.RWMutex
	healthChecker  *ConfigHealthChecker
	notificationMu sync.RWMutex
}

// ConfigChangeEvent 配置变更事件
type ConfigChangeEvent struct {
	Type        ConfigChangeType
	Timestamp   time.Time
	Config      *ManagedConfig
	Changes     []ConfigChange
	Source      string
	Description string
}

// ConfigChangeType 配置变更类型
type ConfigChangeType string

const (
	ConfigChangeTypeUpdate    ConfigChangeType = "update"
	ConfigChangeTypeReload    ConfigChangeType = "reload"
	ConfigChangeTypeRollback  ConfigChangeType = "rollback"
	ConfigChangeTypeSubscribe ConfigChangeType = "subscribe"
	ConfigChangeTypeHealth    ConfigChangeType = "health"
)

// ConfigHealthStatus 配置健康状态
type ConfigHealthStatus struct {
	Healthy        bool
	LastCheckTime  time.Time
	Issues         []string
	Version        string
	LastUpdateTime time.Time
}

// ConfigHealthChecker 配置健康检查器
type ConfigHealthChecker struct {
	center     *EnhancedConfigCenter
	checkFuncs []HealthCheckFunc
	interval   time.Duration
	stopCh     chan struct{}
	mu         sync.RWMutex
	lastStatus ConfigHealthStatus
}

// HealthCheckFunc 健康检查函数
type HealthCheckFunc func(*ManagedConfig) []string

// WrapConfigCenter 包装现有的配置中心
func WrapConfigCenter(baseCenter *ConfigCenter) *EnhancedConfigCenter {
	enhanced := &EnhancedConfigCenter{
		center:      baseCenter,
		broadcastCh: make(chan ConfigChangeEvent, 100),
		subscribers: make(map[string]chan ConfigChangeEvent),
	}

	// 初始化健康检查器
	enhanced.healthChecker = &ConfigHealthChecker{
		center:     enhanced,
		checkFuncs: []HealthCheckFunc{defaultHealthCheck},
		interval:   30 * time.Second,
		stopCh:     make(chan struct{}),
		lastStatus: ConfigHealthStatus{
			Healthy:       true,
			LastCheckTime: time.Now(),
		},
	}

	// 启动广播处理器
	go enhanced.broadcastProcessor()

	// 启动健康检查
	go enhanced.healthChecker.start()

	return enhanced
}

// Subscribe 订阅配置变更事件
func (ecc *EnhancedConfigCenter) Subscribe(subscriberID string) (<-chan ConfigChangeEvent, error) {
	ecc.subscriberMu.Lock()
	defer ecc.subscriberMu.Unlock()

	if _, exists := ecc.subscribers[subscriberID]; exists {
		return nil, fmt.Errorf("subscriber %s already exists", subscriberID)
	}

	eventCh := make(chan ConfigChangeEvent, 10)
	ecc.subscribers[subscriberID] = eventCh

	// 发送订阅确认事件
	ecc.broadcastCh <- ConfigChangeEvent{
		Type:        ConfigChangeTypeSubscribe,
		Timestamp:   time.Now(),
		Description: fmt.Sprintf("Subscriber %s registered", subscriberID),
		Source:      subscriberID,
	}

	return eventCh, nil
}

// Unsubscribe 取消订阅
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

// broadcastProcessor 广播事件处理器
func (ecc *EnhancedConfigCenter) broadcastProcessor() {
	for event := range ecc.broadcastCh {
		ecc.notificationMu.RLock()
		subscribers := make(map[string]chan ConfigChangeEvent)
		for k, v := range ecc.subscribers {
			subscribers[k] = v
		}
		ecc.notificationMu.RUnlock()

		// 异步发送给所有订阅者
		for subscriberID, ch := range subscribers {
			select {
			case ch <- event:
				// 成功发送
			default:
				// 通道满，记录日志
				logger.Warn("Config event channel full for subscriber",
					"subscriber", subscriberID,
					"event_type", event.Type)
			}
		}
	}
}

// Get 获取配置
func (ecc *EnhancedConfigCenter) Get() *ManagedConfig {
	return ecc.center.Get()
}

// Update 增强版更新配置
func (ecc *EnhancedConfigCenter) Update(newConfig *ManagedConfig, reason, changedBy string) error {
	// 调用基类方法
	if err := ecc.center.Update(newConfig, reason, changedBy); err != nil {
		return err
	}

	// 广播更新事件
	ecc.broadcastCh <- ConfigChangeEvent{
		Type:        ConfigChangeTypeUpdate,
		Timestamp:   time.Now(),
		Config:      ecc.center.Get(),
		Description: reason,
		Source:      changedBy,
	}

	return nil
}

// GetHealthStatus 获取健康状态
func (ecc *EnhancedConfigCenter) GetHealthStatus() ConfigHealthStatus {
	ecc.healthChecker.mu.RLock()
	defer ecc.healthChecker.mu.RUnlock()
	return ecc.healthChecker.lastStatus
}

// AddHealthCheck 添加健康检查函数
func (ecc *EnhancedConfigCenter) AddHealthCheck(checkFunc HealthCheckFunc) {
	ecc.healthChecker.mu.Lock()
	defer ecc.healthChecker.mu.Unlock()
	ecc.healthChecker.checkFuncs = append(ecc.healthChecker.checkFuncs, checkFunc)
}

// Close 关闭增强配置中心
func (ecc *EnhancedConfigCenter) Close() error {
	// 停止健康检查
	ecc.healthChecker.stop()

	// 关闭广播通道
	close(ecc.broadcastCh)

	// 关闭所有订阅者通道
	ecc.subscriberMu.Lock()
	for _, ch := range ecc.subscribers {
		close(ch)
	}
	ecc.subscribers = make(map[string]chan ConfigChangeEvent)
	ecc.subscriberMu.Unlock()

	return nil
}

// start 启动健康检查
func (hc *ConfigHealthChecker) start() {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-hc.stopCh:
			return
		case <-ticker.C:
			hc.performCheck()
		}
	}
}

// stop 停止健康检查
func (hc *ConfigHealthChecker) stop() {
	close(hc.stopCh)
}

// performCheck 执行健康检查
func (hc *ConfigHealthChecker) performCheck() {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	config := hc.center.center.Get() // 修正方法调用
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

	// 如果有问题，广播健康事件
	if len(allIssues) > 0 {
		hc.center.broadcastCh <- ConfigChangeEvent{
			Type:        ConfigChangeTypeHealth,
			Timestamp:   time.Now(),
			Config:      config,
			Description: fmt.Sprintf("Health check found %d issues", len(allIssues)),
		}
	}
}

// defaultHealthCheck 默认健康检查
func defaultHealthCheck(config *ManagedConfig) []string {
	var issues []string

	// 检查必要配置是否存在
	if config.BehaviorAnalysis == nil {
		issues = append(issues, "BehaviorAnalysis config is missing")
	}
	if config.RiskScoring == nil {
		issues = append(issues, "RiskScoring config is missing")
	}
	if config.Features == nil {
		issues = append(issues, "Features config is missing")
	}

	// 检查配置值的有效性
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
