package config

import (
	"fmt"
)

// translated comment
type CloneableConfig interface {
	Clone() interface{}
}

// translated comment
// translated comment
// translated comment
// translated comment
// translated comment
func DeepCopy(config *ManagedConfig) (*ManagedConfig, error) {
	if config == nil {
		return nil, nil
	}

	result := &ManagedConfig{}

	// translated comment
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

// translated comment
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

// translated comment
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

// translated comment
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

// translated comment
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
	// translated comment
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

// translated comment
func cloneQUICConfig(c *QUICConfig) *QUICConfig {
	if c == nil {
		return nil
	}
	result := &QUICConfig{
		MinInitialMaxData:      c.MinInitialMaxData,
		MinStreamData:          c.MinStreamData,
		TransportParamCapacity: c.TransportParamCapacity,
	}
	// translated comment
	if c.SupportedVersions != nil {
		result.SupportedVersions = make([]uint32, len(c.SupportedVersions))
		copy(result.SupportedVersions, c.SupportedVersions)
	}
	return result
}

// translated comment
func cloneTLSConfig(c *TLSConfig) *TLSConfig {
	if c == nil {
		return nil
	}
	result := &TLSConfig{
		AnomalyFlagsCapacity: c.AnomalyFlagsCapacity,
	}
	// translated comment
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

// translated comment
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

// translated comment
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

// translated comment
func ValidateDeepCopy(original *ManagedConfig, copied *ManagedConfig) error {
	if original == nil && copied == nil {
		return nil
	}

	if original == nil || copied == nil {
		return fmt.Errorf("deep copy failed: one is nil while the other is not")
	}

	// translated comment
	if &original == &copied {
		return fmt.Errorf("deep copy failed: returned same pointer")
	}

	// translated comment
	if original.BehaviorAnalysis != nil && copied.BehaviorAnalysis != nil {
		if &original.BehaviorAnalysis == &copied.BehaviorAnalysis {
			return fmt.Errorf("deep copy failed: BehaviorAnalysis is same pointer")
		}
	}

	return nil
}
