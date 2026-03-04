package config

import (
	"testing"
	"time"
)

// TestNewUnifiedConfigManager 测试创建统一配置管理器
func TestNewUnifiedConfigManager(t *testing.T) {
	ucm := NewUnifiedConfigManager("")
	if ucm == nil {
		t.Fatal("NewUnifiedConfigManager() returned nil")
	}

	if ucm.ConfigCenter == nil {
		t.Error("ConfigCenter is nil")
	}

	if ucm.enhanced != nil {
		t.Error("enhanced features should be nil by default")
	}
}

// TestUnifiedConfigManager_EnableEnhancedFeatures 测试启用增强功能
func TestUnifiedConfigManager_EnableEnhancedFeatures(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	// 默认未启用
	if ucm.enhanced != nil {
		t.Error("enhanced features should be nil by default")
	}

	// 启用增强功能
	ucm.EnableEnhancedFeatures()

	if ucm.enhanced == nil {
		t.Fatal("enhanced features should be enabled")
	}

	if ucm.enhanced.broadcastCh == nil {
		t.Error("broadcastCh is nil")
	}

	if ucm.enhanced.subscribers == nil {
		t.Error("subscribers is nil")
	}

	if ucm.enhanced.healthChecker == nil {
		t.Error("healthChecker is nil")
	}

	// 重复启用不应出错
	ucm.EnableEnhancedFeatures()
}

// TestUnifiedConfigManager_Subscribe 测试订阅功能
func TestUnifiedConfigManager_Subscribe(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	// 未启用增强功能时订阅应该失败
	_, err := ucm.Subscribe("test-subscriber")
	if err == nil {
		t.Error("Subscribe should fail when enhanced features not enabled")
	}

	// 启用增强功能
	ucm.EnableEnhancedFeatures()

	// 订阅应该成功
	eventCh, err := ucm.Subscribe("test-subscriber")
	if err != nil {
		t.Errorf("Subscribe() error = %v", err)
	}

	if eventCh == nil {
		t.Error("Subscribe() returned nil channel")
	}

	// 重复订阅应该失败
	_, err = ucm.Subscribe("test-subscriber")
	if err == nil {
		t.Error("Subscribe should fail for duplicate subscriber")
	}

	// 取消订阅
	err = ucm.Unsubscribe("test-subscriber")
	if err != nil {
		t.Errorf("Unsubscribe() error = %v", err)
	}

	// 取消不存在的订阅者应该失败
	err = ucm.Unsubscribe("non-existent")
	if err == nil {
		t.Error("Unsubscribe should fail for non-existent subscriber")
	}
}

// TestUnifiedConfigManager_GetBehaviorAnalysisConfig 测试获取行为分析配置
func TestUnifiedConfigManager_GetBehaviorAnalysisConfig(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	// 获取默认配置
	config := ucm.GetBehaviorAnalysisConfig()
	if config == nil {
		t.Fatal("GetBehaviorAnalysisConfig() returned nil")
	}

	// 验证默认值
	if config.MinRequestsForAnalysis <= 0 {
		t.Errorf("MinRequestsForAnalysis = %d, want > 0", config.MinRequestsForAnalysis)
	}

	if config.RegularityThreshold < 0 || config.RegularityThreshold > 1 {
		t.Errorf("RegularityThreshold = %f, want between 0 and 1", config.RegularityThreshold)
	}
}

// TestUnifiedConfigManager_GetRiskScoringConfig 测试获取风险评分配置
func TestUnifiedConfigManager_GetRiskScoringConfig(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	config := ucm.GetRiskScoringConfig()
	if config == nil {
		t.Fatal("GetRiskScoringConfig() returned nil")
	}

	if config.Weights == nil {
		t.Error("Weights is nil")
	}
}

// TestUnifiedConfigManager_GetFeatureExtractionConfig 测试获取特征提取配置
func TestUnifiedConfigManager_GetFeatureExtractionConfig(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	config := ucm.GetFeatureExtractionConfig()
	if config == nil {
		t.Fatal("GetFeatureExtractionConfig() returned nil")
	}

	if len(config.ToolMarkers) == 0 {
		t.Error("ToolMarkers is empty")
	}

	if len(config.HeadlessMarkers) == 0 {
		t.Error("HeadlessMarkers is empty")
	}
}

// TestUnifiedConfigManager_GetQUICConfig 测试获取 QUIC 配置
func TestUnifiedConfigManager_GetQUICConfig(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	config := ucm.GetQUICConfig()
	if config == nil {
		t.Fatal("GetQUICConfig() returned nil")
	}
}

// TestUnifiedConfigManager_GetTLSConfig 测试获取 TLS 配置
func TestUnifiedConfigManager_GetTLSConfig(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	config := ucm.GetTLSConfig()
	if config == nil {
		t.Fatal("GetTLSConfig() returned nil")
	}
}

// TestUnifiedConfigManager_GetGlobalConfig 测试获取全局配置
func TestUnifiedConfigManager_GetGlobalConfig(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	config := ucm.GetGlobalConfig()
	if config == nil {
		t.Fatal("GetGlobalConfig() returned nil")
	}

	if config.MaxConcurrency <= 0 {
		t.Errorf("MaxConcurrency = %d, want > 0", config.MaxConcurrency)
	}
}

// TestUnifiedConfigManager_UpdateBehaviorAnalysisConfig 测试更新行为分析配置
func TestUnifiedConfigManager_UpdateBehaviorAnalysisConfig(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	newConfig := &BehaviorAnalysisConfig{
		MinRequestsForAnalysis:         10,
		RegularityThreshold:            0.5,
		EntropyThreshold:               0.6,
		AnomalousIntervalRateThreshold: 0.3,
	}

	// 需要先设置 loaded 为 true 才能更新
	ucm.ConfigCenter.loaded = true

	err := ucm.UpdateBehaviorAnalysisConfig(newConfig, "test update", "test")
	if err != nil {
		t.Errorf("UpdateBehaviorAnalysisConfig() error = %v", err)
	}

	// 验证更新后的配置
	config := ucm.GetBehaviorAnalysisConfig()
	if config.MinRequestsForAnalysis != 10 {
		t.Errorf("MinRequestsForAnalysis = %d, want 10", config.MinRequestsForAnalysis)
	}
}

// TestUnifiedConfigManager_GetHealthStatus 测试获取健康状态
func TestUnifiedConfigManager_GetHealthStatus(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	// 未启用增强功能时应该失败
	_, err := ucm.GetHealthStatus()
	if err == nil {
		t.Error("GetHealthStatus should fail when enhanced features not enabled")
	}

	// 启用增强功能
	ucm.EnableEnhancedFeatures()

	// 给健康检查一点时间启动
	time.Sleep(100 * time.Millisecond)

	// 应该能获取健康状态
	status, err := ucm.GetHealthStatus()
	if err != nil {
		t.Errorf("GetHealthStatus() error = %v", err)
	}

	// 空配置的健康状态
	_ = status
}

// TestUnifiedConfigManager_DisableEnhancedFeatures 测试禁用增强功能
func TestUnifiedConfigManager_DisableEnhancedFeatures(t *testing.T) {
	ucm := NewUnifiedConfigManager("")

	// 禁用未启用的功能不应出错
	ucm.DisableEnhancedFeatures()

	// 启用后再禁用
	ucm.EnableEnhancedFeatures()
	ucm.DisableEnhancedFeatures()

	if ucm.enhanced != nil {
		t.Error("enhanced features should be nil after disabling")
	}
}

// TestUnifiedConfigManager_WithConfigPath 测试带配置路径的管理器
func TestUnifiedConfigManager_WithConfigPath(t *testing.T) {
	// 使用空路径（内存模式）
	ucm := NewUnifiedConfigManager("")

	if ucm.configPath != "" {
		t.Error("configPath should be empty for memory mode")
	}

	// 验证可以正常操作
	config := ucm.GetBehaviorAnalysisConfig()
	if config == nil {
		t.Error("GetBehaviorAnalysisConfig() returned nil")
	}
}

// TestUnifiedConfigManager_ConcurrentAccess 测试并发访问
func TestUnifiedConfigManager_ConcurrentAccess(t *testing.T) {
	ucm := NewUnifiedConfigManager("")
	ucm.EnableEnhancedFeatures()

	// 并发读取
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_ = ucm.GetBehaviorAnalysisConfig()
			_ = ucm.GetRiskScoringConfig()
			_ = ucm.GetFeatureExtractionConfig()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// OK
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent access timeout")
		}
	}
}

// TestCompatibility_NewConfigManager 测试兼容性的 NewConfigManager
func TestCompatibility_NewConfigManager(t *testing.T) {
	center := NewConfigCenter("")
	cm := NewConfigManager(center)

	if cm == nil {
		t.Fatal("NewConfigManager() returned nil")
	}

	if cm.center != center {
		t.Error("center not set correctly")
	}

	// 验证 ConfigManager 的基本功能
	config := cm.GetBehaviorAnalysisConfig()
	if config == nil {
		t.Error("GetBehaviorAnalysisConfig() returned nil")
	}
}

// TestCompatibility_WrapConfigCenter 测试兼容性的 WrapConfigCenter
func TestCompatibility_WrapConfigCenter(t *testing.T) {
	center := NewConfigCenter("")
	enhanced := WrapConfigCenter(center)

	if enhanced == nil {
		t.Fatal("WrapConfigCenter() returned nil")
	}

	if enhanced.center != center {
		t.Error("center not set correctly")
	}

	// 验证增强功能已启用
	_, err := enhanced.Subscribe("test")
	if err != nil {
		t.Errorf("Subscribe() error = %v", err)
	}
}
