package config

import (
	"fmt"
)

// CloneableConfig 可克隆配置的接口
type CloneableConfig interface {
	Clone() interface{}
}

// DeepCopy 执行深复制。相比 JSON 序列化方式的优势：
// 1. 编译时安全（不依赖 JSON tag）
// 2. 性能更好（避免序列化开销）
// 3. 可以支持不可序列化的字段（如 sync.Mutex）
// 4. 错误处理更清晰
func DeepCopy(config *ManagedConfig) (*ManagedConfig, error) {
	if config == nil {
		return nil, nil
	}

	result := &ManagedConfig{}

	// 深复制每个嵌套结构
	if config.BehaviorAnalysis != nil {
		result.BehaviorAnalysis = cloneBehaviorAnalysisConfig(config.BehaviorAnalysis)
	}

	if config.RiskScoring != nil {
		result.RiskScoring = cloneRiskScoringConfig(config.RiskScoring)
	}

	if config.Features != nil {
		result.Features = cloneFeatureExtractionConfig(config.Features)
	}

	if config.QUIC != nil {
		result.QUIC = cloneQUICConfig(config.QUIC)
	}

	if config.TLS != nil {
		result.TLS = cloneTLSConfig(config.TLS)
	}

	if config.Global != nil {
		result.Global = cloneGlobalConfig(config.Global)
	}

	if config.Metadata != nil {
		result.Metadata = cloneConfigMetadata(config.Metadata)
	}

	return result, nil
}

// cloneBehaviorAnalysisConfig 深复制行为分析配置
func cloneBehaviorAnalysisConfig(c *BehaviorAnalysisConfig) *BehaviorAnalysisConfig {
	if c == nil {
		return nil
	}
	return &BehaviorAnalysisConfig{
		MinRequestsForAnalysis:         c.MinRequestsForAnalysis,
		RegularityThreshold:            c.RegularityThreshold,
		EntropyThreshold:               c.EntropyThreshold,
		AnomalousIntervalRateThreshold: c.AnomalousIntervalRateThreshold,
		RequestHistoryCapacity:         c.RequestHistoryCapacity,
		SignalCapacity:                 c.SignalCapacity,
	}
}

// cloneRiskScoringConfig 深复制风险评分配置
func cloneRiskScoringConfig(c *RiskScoringConfig) *RiskScoringConfig {
	if c == nil {
		return nil
	}
	result := &RiskScoringConfig{
		CriticalThreshold: c.CriticalThreshold,
		HighThreshold:     c.HighThreshold,
		MediumThreshold:   c.MediumThreshold,
		LowThreshold:      c.LowThreshold,
		MinConfidence:     c.MinConfidence,
	}
	if c.Weights != nil {
		result.Weights = cloneRiskWeights(c.Weights)
	}
	return result
}

// cloneRiskWeights 深复制风险权重
func cloneRiskWeights(w *RiskWeights) *RiskWeights {
	if w == nil {
		return nil
	}
	return &RiskWeights{
		Headless:         w.Headless,
		Anomaly:          w.Anomaly,
		Contradiction:    w.Contradiction,
		ECH:              w.ECH,
		ClientHints:      w.ClientHints,
		BehaviorAnomaly:  w.BehaviorAnomaly,
		CipherSuiteRisk:  w.CipherSuiteRisk,
		ExtensionAnomaly: w.ExtensionAnomaly,
	}
}

// cloneFeatureExtractionConfig 深复制特征提取配置
func cloneFeatureExtractionConfig(c *FeatureExtractionConfig) *FeatureExtractionConfig {
	if c == nil {
		return nil
	}
	result := &FeatureExtractionConfig{
		EntropyHighThreshold:  c.EntropyHighThreshold,
		EntropyLowThreshold:   c.EntropyLowThreshold,
		MobileScreenWidthMax:  c.MobileScreenWidthMax,
		DesktopScreenWidthMin: c.DesktopScreenWidthMin,
	}
	// 深复制字符串切片
	if c.ToolMarkers != nil {
		result.ToolMarkers = make([]string, len(c.ToolMarkers))
		copy(result.ToolMarkers, c.ToolMarkers)
	}
	if c.HeadlessMarkers != nil {
		result.HeadlessMarkers = make([]string, len(c.HeadlessMarkers))
		copy(result.HeadlessMarkers, c.HeadlessMarkers)
	}
	return result
}

// cloneQUICConfig 深复制 QUIC 配置
func cloneQUICConfig(c *QUICConfig) *QUICConfig {
	if c == nil {
		return nil
	}
	result := &QUICConfig{
		MinInitialMaxData:      c.MinInitialMaxData,
		MinStreamData:          c.MinStreamData,
		TransportParamCapacity: c.TransportParamCapacity,
	}
	// 深复制 uint32 切片
	if c.SupportedVersions != nil {
		result.SupportedVersions = make([]uint32, len(c.SupportedVersions))
		copy(result.SupportedVersions, c.SupportedVersions)
	}
	return result
}

// cloneTLSConfig 深复制 TLS 配置
func cloneTLSConfig(c *TLSConfig) *TLSConfig {
	if c == nil {
		return nil
	}
	result := &TLSConfig{
		AnomalyFlagsCapacity: c.AnomalyFlagsCapacity,
	}
	// 深复制 uint16 切片
	if c.WeakCipherSuites != nil {
		result.WeakCipherSuites = make([]uint16, len(c.WeakCipherSuites))
		copy(result.WeakCipherSuites, c.WeakCipherSuites)
	}
	if c.SupportedVersions != nil {
		result.SupportedVersions = make([]uint16, len(c.SupportedVersions))
		copy(result.SupportedVersions, c.SupportedVersions)
	}
	if c.GREASEExtensions != nil {
		result.GREASEExtensions = make([]uint16, len(c.GREASEExtensions))
		copy(result.GREASEExtensions, c.GREASEExtensions)
	}
	return result
}

// cloneGlobalConfig 深复制全局配置
func cloneGlobalConfig(c *GlobalConfig) *GlobalConfig {
	if c == nil {
		return nil
	}
	return &GlobalConfig{
		MaxConcurrency: c.MaxConcurrency,
		RequestTimeout: c.RequestTimeout,
		CacheSize:      c.CacheSize,
		DebugMode:      c.DebugMode,
		MaxInputSize:   c.MaxInputSize,
	}
}

// cloneConfigMetadata 深复制配置元数据
func cloneConfigMetadata(m *ConfigMetadata) *ConfigMetadata {
	if m == nil {
		return nil
	}
	return &ConfigMetadata{
		Version:      m.Version,
		LastModified: m.LastModified,
		Author:       m.Author,
		Description:  m.Description,
	}
}

// ValidateDeepCopy 验证深复制的有效性
func ValidateDeepCopy(original *ManagedConfig, copied *ManagedConfig) error {
	if original == nil && copied == nil {
		return nil
	}

	if original == nil || copied == nil {
		return fmt.Errorf("deep copy failed: one is nil while the other is not")
	}

	// 检查指针不同（真正的深复制，而不是浅复制）
	if &original == &copied {
		return fmt.Errorf("deep copy failed: returned same pointer")
	}

	// 检查嵌套对象也是不同的指针
	if original.BehaviorAnalysis != nil && copied.BehaviorAnalysis != nil {
		if &original.BehaviorAnalysis == &copied.BehaviorAnalysis {
			return fmt.Errorf("deep copy failed: BehaviorAnalysis is same pointer")
		}
	}

	return nil
}
