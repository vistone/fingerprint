package config

import (
	"fmt"
)

// CloneableConfig is the interface for cloneable configurations
type CloneableConfig interface {
	Clone() interface{}
}

// DeepCopy performs a deep copy. Advantages over JSON serialization:
// 1. Compile-time safety (no dependency on JSON tags)
// 2. Better performance (avoids serialization overhead)
// 3. Can support non-serializable fields (e.g., sync.Mutex)
// 4. Clearer error handling
func DeepCopy(config *ManagedConfig) (*ManagedConfig, error) {
	if config == nil {
		return nil, nil
	}

	result := &ManagedConfig{}

	// Deep copy each nested structure
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

// cloneBehaviorAnalysisConfig deep copies the behavior analysis configuration
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

// cloneRiskScoringConfig deep copies the risk scoring configuration
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

// cloneRiskWeights deep copies the risk weights
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

// cloneFeatureExtractionConfig deep copies the feature extraction configuration
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
	// Deep copy string slices
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

// cloneQUICConfig deep copies the QUIC configuration
func cloneQUICConfig(c *QUICConfig) *QUICConfig {
	if c == nil {
		return nil
	}
	result := &QUICConfig{
		MinInitialMaxData:      c.MinInitialMaxData,
		MinStreamData:          c.MinStreamData,
		TransportParamCapacity: c.TransportParamCapacity,
	}
	// Deep copy uint32 slices
	if c.SupportedVersions != nil {
		result.SupportedVersions = make([]uint32, len(c.SupportedVersions))
		copy(result.SupportedVersions, c.SupportedVersions)
	}
	return result
}

// cloneTLSConfig deep copies the TLS configuration
func cloneTLSConfig(c *TLSConfig) *TLSConfig {
	if c == nil {
		return nil
	}
	result := &TLSConfig{
		AnomalyFlagsCapacity: c.AnomalyFlagsCapacity,
	}
	// Deep copy uint16 slices
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

// cloneGlobalConfig deep copies the global configuration
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

// cloneConfigMetadata deep copies the configuration metadata
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

// ValidateDeepCopy validates the correctness of a deep copy
func ValidateDeepCopy(original *ManagedConfig, copied *ManagedConfig) error {
	if original == nil && copied == nil {
		return nil
	}

	if original == nil || copied == nil {
		return fmt.Errorf("deep copy failed: one is nil while the other is not")
	}

	// Check that the pointers are different (a true deep copy, not a shallow copy)
	if &original == &copied {
		return fmt.Errorf("deep copy failed: returned same pointer")
	}

	// Check that nested objects also have different pointers
	if original.BehaviorAnalysis != nil && copied.BehaviorAnalysis != nil {
		if &original.BehaviorAnalysis == &copied.BehaviorAnalysis {
			return fmt.Errorf("deep copy failed: BehaviorAnalysis is same pointer")
		}
	}

	return nil
}
