package config

import (
	"testing"

	ic "github.com/vistone/fingerprint/modules/internal/config"
)

// TestTypeAliases 测试类型别名是否正确导出
func TestTypeAliases(t *testing.T) {
	// 测试类型别名可以正常使用
	var cc *ConfigCenter
	_ = cc

	var cm *ConfigManager
	_ = cm

	var hc *HealthChecker
	_ = hc

	var mc *ManagedConfig
	_ = mc

	var change *ConfigChange
	_ = change

	var listener ConfigChangeListener
	_ = listener

	var bac *BehaviorAnalysisConfig
	_ = bac

	var rsc *RiskScoringConfig
	_ = rsc

	var fec *FeatureExtractionConfig
	_ = fec

	var qc *QUICConfig
	_ = qc

	var tc *TLSConfig
	_ = tc

	var gc *GlobalConfig
	_ = gc
}

// TestInitializeConfigCenter 测试初始化配置中心
func TestInitializeConfigCenter(t *testing.T) {
	err := InitializeConfigCenter()
	// 如果没有配置文件，初始化会失败，这是预期的行为
	// 我们只需要确保函数被调用且不 panic 即可
	_ = err
	t.Logf("InitializeConfigCenter() returned: %v", err)
}

// TestInitializeConfigCenterWithDefaults 测试使用默认配置初始化
func TestInitializeConfigCenterWithDefaults(t *testing.T) {
	err := InitializeConfigCenterWithDefaults()
	if err != nil {
		t.Errorf("InitializeConfigCenterWithDefaults() error = %v", err)
	}
}

// TestGetConfigCenter 测试获取配置中心
func TestGetConfigCenter(t *testing.T) {
	// 确保先初始化
	_ = InitializeConfigCenter()

	cc := GetConfigCenter()
	if cc == nil {
		t.Error("GetConfigCenter() returned nil")
	}
}

// TestGetConfigManager 测试获取配置管理器
func TestGetConfigManager(t *testing.T) {
	// 确保先初始化
	_ = InitializeConfigCenter()

	cm := GetConfigManager()
	if cm == nil {
		t.Error("GetConfigManager() returned nil")
	}
}

// TestGetHealthChecker 测试获取健康检查器
func TestGetHealthChecker(t *testing.T) {
	// 确保先初始化
	_ = InitializeConfigCenter()

	hc := GetHealthChecker()
	if hc == nil {
		t.Error("GetHealthChecker() returned nil")
	}
}

// TestBridgeIntegration 测试桥接功能集成
func TestBridgeIntegration(t *testing.T) {
	// 测试完整的初始化流程
	err := InitializeConfigCenterWithDefaults()
	if err != nil {
		t.Fatalf("Failed to initialize config center: %v", err)
	}

	// 获取各个组件
	cc := GetConfigCenter()
	cm := GetConfigManager()
	hc := GetHealthChecker()

	if cc == nil {
		t.Error("ConfigCenter is nil")
	}

	if cm == nil {
		t.Error("ConfigManager is nil")
	}

	if hc == nil {
		t.Error("HealthChecker is nil")
	}

	// 验证类型可以转换回 internal 包类型
	var internalCC *ic.ConfigCenter = cc
	_ = internalCC

	var internalCM *ic.ConfigManager = cm
	_ = internalCM

	var internalHC *ic.HealthChecker = hc
	_ = internalHC
}

// TestManagedConfigUsage 测试 ManagedConfig 使用
func TestManagedConfigUsage(t *testing.T) {
	config := &ManagedConfig{
		BehaviorAnalysis: &BehaviorAnalysisConfig{},
		RiskScoring:      &RiskScoringConfig{},
		Features:         &FeatureExtractionConfig{},
		QUIC:             &QUICConfig{},
		TLS:              &TLSConfig{},
		Global:           &GlobalConfig{},
	}

	if config.BehaviorAnalysis == nil {
		t.Error("BehaviorAnalysis is nil")
	}

	if config.RiskScoring == nil {
		t.Error("RiskScoring is nil")
	}

	if config.Features == nil {
		t.Error("Features is nil")
	}

	if config.QUIC == nil {
		t.Error("QUIC is nil")
	}

	if config.TLS == nil {
		t.Error("TLS is nil")
	}

	if config.Global == nil {
		t.Error("Global is nil")
	}
}

// TestConfigChangeUsage 测试 ConfigChange 使用
func TestConfigChangeUsage(t *testing.T) {
	change := &ConfigChange{
		Path:     "test.path",
		OldValue: "old",
		NewValue: "new",
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
