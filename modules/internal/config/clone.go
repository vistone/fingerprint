package config

// clone.go - Efficient deep copy implementation for configuration
// Avoid JSON serialization to provide better performance and type safety

// Clone creates a deep copy of ManagedConfig
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

// Clone creates a deep copy of BehaviorAnalysisConfig
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

// Clone creates a deep copy of RiskScoringConfig
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

// Clone creates a deep copy of RiskWeights
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

// Clone creates a deep copy of FeatureExtractionConfig
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

	// Deep copy slices.
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

// Clone creates a deep copy of QUICConfig
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

// Clone creates a deep copy of TLSConfig
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

// Clone creates a deep copy of GlobalConfig
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

// Clone creates a deep copy of ConfigMetadata
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
