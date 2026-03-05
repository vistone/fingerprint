package config

// clone.go - 高效的配置深拷贝实现
// 避免使用 JSON 序列化，提供更好的性能和类型安全

// Clone 创建 ManagedConfig 的深拷贝
func (c *ManagedConfig) Clone() *ManagedConfig {
	if c == nil {
		return nil
	}

	return &ManagedConfig{
		BehaviorAnalysis: c.BehaviorAnalysis.Clone(),
		RiskScoring:      c.RiskScoring.Clone(),
		Features:         c.Features.Clone(),
		QUIC:             c.QUIC.Clone(),
		TLS:              c.TLS.Clone(),
		Global:           c.Global.Clone(),
		Metadata:         c.Metadata.Clone(),
	}
}

// Clone 创建 BehaviorAnalysisConfig 的深拷贝
func (c *BehaviorAnalysisConfig) Clone() *BehaviorAnalysisConfig {
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

// Clone 创建 RiskScoringConfig 的深拷贝
func (c *RiskScoringConfig) Clone() *RiskScoringConfig {
	if c == nil {
		return nil
	}

	clone := &RiskScoringConfig{
		CriticalThreshold: c.CriticalThreshold,
		HighThreshold:     c.HighThreshold,
		MediumThreshold:   c.MediumThreshold,
		LowThreshold:      c.LowThreshold,
		MinConfidence:     c.MinConfidence,
	}

	if c.Weights != nil {
		clone.Weights = c.Weights.Clone()
	}

	return clone
}

// Clone 创建 RiskWeights 的深拷贝
func (w *RiskWeights) Clone() *RiskWeights {
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

// Clone 创建 FeatureExtractionConfig 的深拷贝
func (c *FeatureExtractionConfig) Clone() *FeatureExtractionConfig {
	if c == nil {
		return nil
	}

	clone := &FeatureExtractionConfig{
		EntropyHighThreshold:  c.EntropyHighThreshold,
		EntropyLowThreshold:   c.EntropyLowThreshold,
		MobileScreenWidthMax:  c.MobileScreenWidthMax,
		DesktopScreenWidthMin: c.DesktopScreenWidthMin,
	}

	// 深拷贝切片
	if c.ToolMarkers != nil {
		clone.ToolMarkers = make([]string, len(c.ToolMarkers))
		copy(clone.ToolMarkers, c.ToolMarkers)
	}

	if c.HeadlessMarkers != nil {
		clone.HeadlessMarkers = make([]string, len(c.HeadlessMarkers))
		copy(clone.HeadlessMarkers, c.HeadlessMarkers)
	}

	return clone
}

// Clone 创建 QUICConfig 的深拷贝
func (c *QUICConfig) Clone() *QUICConfig {
	if c == nil {
		return nil
	}

	clone := &QUICConfig{
		MinInitialMaxData:      c.MinInitialMaxData,
		MinStreamData:          c.MinStreamData,
		TransportParamCapacity: c.TransportParamCapacity,
	}

	if c.SupportedVersions != nil {
		clone.SupportedVersions = make([]uint32, len(c.SupportedVersions))
		copy(clone.SupportedVersions, c.SupportedVersions)
	}

	return clone
}

// Clone 创建 TLSConfig 的深拷贝
func (c *TLSConfig) Clone() *TLSConfig {
	if c == nil {
		return nil
	}

	clone := &TLSConfig{
		AnomalyFlagsCapacity: c.AnomalyFlagsCapacity,
	}

	if c.WeakCipherSuites != nil {
		clone.WeakCipherSuites = make([]uint16, len(c.WeakCipherSuites))
		copy(clone.WeakCipherSuites, c.WeakCipherSuites)
	}

	if c.SupportedVersions != nil {
		clone.SupportedVersions = make([]uint16, len(c.SupportedVersions))
		copy(clone.SupportedVersions, c.SupportedVersions)
	}

	if c.GREASEExtensions != nil {
		clone.GREASEExtensions = make([]uint16, len(c.GREASEExtensions))
		copy(clone.GREASEExtensions, c.GREASEExtensions)
	}

	return clone
}

// Clone 创建 GlobalConfig 的深拷贝
func (c *GlobalConfig) Clone() *GlobalConfig {
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

// Clone 创建 ConfigMetadata 的深拷贝
func (m *ConfigMetadata) Clone() *ConfigMetadata {
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
