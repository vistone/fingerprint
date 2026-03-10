package config

// BehaviorAnalysisConfig represents the behavior analysis configuration
type BehaviorAnalysisConfig struct {
	// Minimum number of requests required for analysis
	MinRequestsForAnalysis int `json:"min_requests_for_analysis"`

	// Regularity threshold (0-1)
	RegularityThreshold float64 `json:"regularity_threshold"`

	// Entropy threshold (0-1)
	EntropyThreshold float64 `json:"entropy_threshold"`

	// Anomalous interval rate threshold (0-1)
	AnomalousIntervalRateThreshold float64 `json:"anomalous_interval_rate_threshold"`

	// Request history capacity (pre-allocated)
	RequestHistoryCapacity int `json:"request_history_capacity"`

	// Signal capacity (pre-allocated)
	SignalCapacity int `json:"signal_capacity"`
}

// RiskScoringConfig represents the risk scoring configuration
type RiskScoringConfig struct {
	// Critical threat level threshold
	CriticalThreshold float64 `json:"critical_threshold"`

	// High threat level threshold
	HighThreshold float64 `json:"high_threshold"`

	// Medium threat level threshold
	MediumThreshold float64 `json:"medium_threshold"`

	// Low threat level threshold
	LowThreshold float64 `json:"low_threshold"`

	// Minimum confidence
	MinConfidence float64 `json:"min_confidence"`

	// Weights configuration
	Weights *RiskWeights `json:"weights"`
}

// RiskWeights represents the risk scoring weights
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

// FeatureExtractionConfig represents the feature extraction configuration
type FeatureExtractionConfig struct {
	// High entropy threshold (bits)
	EntropyHighThreshold float64 `json:"entropy_high_threshold"`

	// Low entropy threshold (unique bytes)
	EntropyLowThreshold int `json:"entropy_low_threshold"`

	// Tool marker list
	ToolMarkers []string `json:"tool_markers"`

	// Headless browser marker list
	HeadlessMarkers []string `json:"headless_markers"`

	// Maximum mobile device screen width
	MobileScreenWidthMax int `json:"mobile_screen_width_max"`

	// Minimum desktop device screen width
	DesktopScreenWidthMin int `json:"desktop_screen_width_min"`
}

// QUICConfig represents the QUIC configuration
type QUICConfig struct {
	// Minimum suspicious initial max data size
	MinInitialMaxData int `json:"min_initial_max_data"`

	// Minimum suspicious stream data size
	MinStreamData int `json:"min_stream_data"`

	// Supported protocol versions
	SupportedVersions []uint32 `json:"supported_versions"`

	// Detection parameter capacity
	TransportParamCapacity int `json:"transport_param_capacity"`
}

// TLSConfig represents the TLS configuration
type TLSConfig struct {
	// Weak cipher suite list
	WeakCipherSuites []uint16 `json:"weak_cipher_suites"`

	// Supported TLS versions
	SupportedVersions []uint16 `json:"supported_versions"`

	// GREASE extension list
	GREASEExtensions []uint16 `json:"grease_extensions"`

	// Anomaly flags capacity
	AnomalyFlagsCapacity int `json:"anomaly_flags_capacity"`
}

// GlobalConfig represents the global configuration
type GlobalConfig struct {
	// Maximum concurrency
	MaxConcurrency int `json:"max_concurrency"`

	// Request timeout (milliseconds)
	RequestTimeout int `json:"request_timeout"`

	// Cache size
	CacheSize int `json:"cache_size"`

	// Whether to enable debug mode
	DebugMode bool `json:"debug_mode"`

	// Maximum input size (bytes)
	MaxInputSize int `json:"max_input_size"`
}
