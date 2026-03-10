package config

import "time"

// DefaultManagedConfig creates the default configuration
func DefaultManagedConfig() *ManagedConfig {
	return &ManagedConfig{
		BehaviorAnalysis: &BehaviorAnalysisConfig{
			MinRequestsForAnalysis:         5,
			RegularityThreshold:            0.3,
			EntropyThreshold:               0.5,
			AnomalousIntervalRateThreshold: 0.2,
			RequestHistoryCapacity:         100,
			SignalCapacity:                 50,
		},
		RiskScoring: &RiskScoringConfig{
			CriticalThreshold: 0.9,
			HighThreshold:     0.7,
			MediumThreshold:   0.5,
			LowThreshold:      0.3,
			MinConfidence:     0.5,
			Weights: &RiskWeights{
				Headless:         0.25,
				Anomaly:          0.20,
				Contradiction:    0.15,
				ECH:              0.10,
				ClientHints:      0.10,
				BehaviorAnomaly:  0.10,
				CipherSuiteRisk:  0.05,
				ExtensionAnomaly: 0.05,
			},
		},
		Features: &FeatureExtractionConfig{
			EntropyHighThreshold:  7.5,
			EntropyLowThreshold:   26,
			ToolMarkers:           []string{"HeadlessChrome", "PhantomJS", "webdriver", "selenium", "puppeteer"},
			HeadlessMarkers:       []string{"headlesschrome", "phantomjs", "selenium", "webdriver", "puppeteer", "playwright", "cypress", "jsdom", "zombie", "htmlunit"},
			MobileScreenWidthMax:  1920,
			DesktopScreenWidthMin: 800,
		},
		QUIC: &QUICConfig{
			MinInitialMaxData:      1024,
			MinStreamData:          256,
			SupportedVersions:      []uint32{0x00000001, 0x6b3343cf},
			TransportParamCapacity: 16,
		},
		TLS: &TLSConfig{
			WeakCipherSuites: []uint16{
				0x0001,
				0x0002,
				0x0004,
			},
			SupportedVersions:    []uint16{0x0303, 0x0304},
			GREASEExtensions:     []uint16{},
			AnomalyFlagsCapacity: 8,
		},
		Global: &GlobalConfig{
			MaxConcurrency: 100,
			RequestTimeout: 30000,
			CacheSize:      10000,
			DebugMode:      false,
			MaxInputSize:   1048576,
		},
		Metadata: &ConfigMetadata{
			Version:      "1.0.0",
			LastModified: time.Now(),
			Author:       "fingerprint-team",
			Description:  "Default configuration for fingerprint detection system",
		},
	}
}
