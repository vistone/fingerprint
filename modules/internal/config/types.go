package config

// BehaviorAnalysisConfig 行为分析配置
type BehaviorAnalysisConfig struct {
	// 分析所需的最小请求数
	MinRequestsForAnalysis int `json:"min_requests_for_analysis"`

	// 规律性阈值（0-1）
	RegularityThreshold float64 `json:"regularity_threshold"`

	// 熵阈值（0-1）
	EntropyThreshold float64 `json:"entropy_threshold"`

	// 异常间隔比率阈值（0-1）
	AnomalousIntervalRateThreshold float64 `json:"anomalous_interval_rate_threshold"`

	// 请求历史容量（预分配）
	RequestHistoryCapacity int `json:"request_history_capacity"`

	// 信号容量（预分配）
	SignalCapacity int `json:"signal_capacity"`
}

// RiskScoringConfig 风险评分配置
type RiskScoringConfig struct {
	// Critical 威胁级阈值
	CriticalThreshold float64 `json:"critical_threshold"`

	// High 威胁级阈值
	HighThreshold float64 `json:"high_threshold"`

	// Medium 威胁级阈值
	MediumThreshold float64 `json:"medium_threshold"`

	// Low 威胁级阈值
	LowThreshold float64 `json:"low_threshold"`

	// 最小置信度
	MinConfidence float64 `json:"min_confidence"`

	// 权重配置
	Weights *RiskWeights `json:"weights"`
}

// RiskWeights 风险评分权重
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

// FeatureExtractionConfig 特征提取配置
type FeatureExtractionConfig struct {
	// 高熵阈值（bits）
	EntropyHighThreshold float64 `json:"entropy_high_threshold"`

	// 低熵阈值（unique bytes）
	EntropyLowThreshold int `json:"entropy_low_threshold"`

	// 工具特征列表
	ToolMarkers []string `json:"tool_markers"`

	// 无头浏览器特征列表
	HeadlessMarkers []string `json:"headless_markers"`

	// 移动设备屏幕分辨率上限
	MobileScreenWidthMax int `json:"mobile_screen_width_max"`

	// 桌面设备屏幕分辨率下限
	DesktopScreenWidthMin int `json:"desktop_screen_width_min"`
}

// QUICConfig QUIC 配置
type QUICConfig struct {
	// 可疑的初始数据流大小下限
	MinInitialMaxData int `json:"min_initial_max_data"`

	// 可疑的流数据大小下限
	MinStreamData int `json:"min_stream_data"`

	// 支持的协议版本
	SupportedVersions []uint32 `json:"supported_versions"`

	// 检测参数容量
	TransportParamCapacity int `json:"transport_param_capacity"`
}

// TLSConfig TLS 配置
type TLSConfig struct {
	// 弱密码套件列表
	WeakCipherSuites []uint16 `json:"weak_cipher_suites"`

	// 支持的 TLS 版本
	SupportedVersions []uint16 `json:"supported_versions"`

	// GREASE 扩展列表
	GREASEExtensions []uint16 `json:"grease_extensions"`

	// 异常标记容量
	AnomalyFlagsCapacity int `json:"anomaly_flags_capacity"`
}

// GlobalConfig 全局配置
type GlobalConfig struct {
	// 最大并发处理数
	MaxConcurrency int `json:"max_concurrency"`

	// 请求超时（毫秒）
	RequestTimeout int `json:"request_timeout"`

	// 缓存大小
	CacheSize int `json:"cache_size"`

	// 是否启用调试模式
	DebugMode bool `json:"debug_mode"`

	// 最大输入大小（字节）
	MaxInputSize int `json:"max_input_size"`
}
