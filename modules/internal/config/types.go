package config

// translated comment
type BehaviorAnalysisConfig struct {
	// translated comment
	MinRequestsForAnalysis int `json:"min_requests_for_analysis"`

	// translated comment
	RegularityThreshold float64 `json:"regularity_threshold"`

	// translated comment
	EntropyThreshold float64 `json:"entropy_threshold"`

	// translated comment
	AnomalousIntervalRateThreshold float64 `json:"anomalous_interval_rate_threshold"`

	// translated comment
	RequestHistoryCapacity int `json:"request_history_capacity"`

	// translated comment
	SignalCapacity int `json:"signal_capacity"`
}

// translated comment
type RiskScoringConfig struct {
	// translated comment
	CriticalThreshold float64 `json:"critical_threshold"`

	// translated comment
	HighThreshold float64 `json:"high_threshold"`

	// translated comment
	MediumThreshold float64 `json:"medium_threshold"`

	// translated comment
	LowThreshold float64 `json:"low_threshold"`

	// translated comment
	MinConfidence float64 `json:"min_confidence"`

	// translated comment
	Weights *RiskWeights `json:"weights"`
}

// translated comment
type RiskWeights struct {
	Headless         float64 `json:"headless"`
	Anomaly          float64 `json:"anomaly"`
	Contradiction    float64 `json:"contradiction"`
	ECH              float64 `json:"ech"`
	ClientHints      float64 `json:"client_hints"`
	BehaviorAnomaly  float64 `json:"behavior_anomaly"`
	CipherSuiteRisk  float64 `json:"cipher_suite_risk"`
	ExtensionAnomaly float64 `json:"extension_anomaly"`
}

// translated comment
type FeatureExtractionConfig struct {
	// translated comment
	EntropyHighThreshold float64 `json:"entropy_high_threshold"`

	// translated comment
	EntropyLowThreshold int `json:"entropy_low_threshold"`

	// translated comment
	ToolMarkers []string `json:"tool_markers"`

	// translated comment
	HeadlessMarkers []string `json:"headless_markers"`

	// translated comment
	MobileScreenWidthMax int `json:"mobile_screen_width_max"`

	// translated comment
	DesktopScreenWidthMin int `json:"desktop_screen_width_min"`
}

// translated comment
type QUICConfig struct {
	// translated comment
	MinInitialMaxData int `json:"min_initial_max_data"`

	// translated comment
	MinStreamData int `json:"min_stream_data"`

	// translated comment
	SupportedVersions []uint32 `json:"supported_versions"`

	// translated comment
	TransportParamCapacity int `json:"transport_param_capacity"`
}

// translated comment
type TLSConfig struct {
	// translated comment
	WeakCipherSuites []uint16 `json:"weak_cipher_suites"`

	// translated comment
	SupportedVersions []uint16 `json:"supported_versions"`

	// translated comment
	GREASEExtensions []uint16 `json:"grease_extensions"`

	// translated comment
	AnomalyFlagsCapacity int `json:"anomaly_flags_capacity"`
}

// translated comment
type GlobalConfig struct {
	// translated comment
	MaxConcurrency int `json:"max_concurrency"`

	// translated comment
	RequestTimeout int `json:"request_timeout"`

	// translated comment
	CacheSize int `json:"cache_size"`

	// translated comment
	DebugMode bool `json:"debug_mode"`

	// translated comment
	MaxInputSize int `json:"max_input_size"`
}
