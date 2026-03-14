package features

import (
	"bytes"
	"math"
	"strings"
)

// FeatureType 特征类别枚举
type FeatureType string

const (
	// 熵特征
	FeatureEntropy FeatureType = "entropy"
	// 工具特征
	FeatureToolMarker FeatureType = "tool_marker"
	// 操作系统平台矛盾
	FeatureOSPlatformContradiction FeatureType = "os_platform_contradiction"
	// User-Agent 和操作系统矛盾
	FeatureUAOSContradiction FeatureType = "ua_os_contradiction"
	// 移动设备屏幕分辨率矛盾
	FeatureMobileScreenContradiction FeatureType = "mobile_screen_contradiction"
	// User-Agent 特性矛盾
	FeatureUAFeatureContradiction FeatureType = "ua_feature_contradiction"
	// 无头浏览器特征
	FeatureHeadlessBrowser FeatureType = "headless_browser"
)

// FeatureExtractor 统一特征提取器接口
type FeatureExtractor interface {
	// ExtractFeature 从数据中提取指定类型的特征
	// 返回特征值（0.0-1.0）和是否检测到异常
	ExtractFeature(featureType FeatureType, data interface{}, config *FeatureConfig) (float64, bool)

	// GetFeatureName 获取特征的人类可读名称
	GetFeatureName(featureType FeatureType) string
}

// FeatureConfig 特征提取配置
type FeatureConfig struct {
	// 高熵阈值（bits）
	EntropyHighThreshold float64 `json:"entropy_high_threshold"`
	// 低熵阈值（unique bytes 数量）
	EntropyLowThreshold int `json:"entropy_low_threshold"`
	// 工具特征匹配列表
	ToolMarkers []string `json:"tool_markers"`
	// 无头浏览器特征列表
	HeadlessMarkers []string `json:"headless_markers"`
	// 移动设备屏幕分辨率上限（超过该值可疑）
	MobileScreenWidthMax int `json:"mobile_screen_width_max"`
	// 桌面设备屏幕分辨率下限（低于该值可疑）
	DesktopScreenWidthMin int `json:"desktop_screen_width_min"`
}

// DefaultFeatureConfig 默认特征配置
func DefaultFeatureConfig() *FeatureConfig {
	return &FeatureConfig{
		EntropyHighThreshold:  7.5,
		EntropyLowThreshold:   26,
		ToolMarkers:           []string{"HeadlessChrome", "PhantomJS", "webdriver", "selenium", "puppeteer"},
		HeadlessMarkers:       []string{"headlesschrome", "phantomjs", "selenium", "webdriver", "puppeteer", "playwright", "cypress", "jsdom", "zombie", "htmlunit"},
		MobileScreenWidthMax:  1920,
		DesktopScreenWidthMin: 800,
	}
}

// BaseFeatureExtractor 基础特征提取器实现
type BaseFeatureExtractor struct {
	config *FeatureConfig
}

// NewBaseFeatureExtractor 创建新的基础特征提取器
func NewBaseFeatureExtractor(config *FeatureConfig) *BaseFeatureExtractor {
	if config == nil {
		config = DefaultFeatureConfig()
	}
	return &BaseFeatureExtractor{config: config}
}

// ExtractFeature 实现统一的特征提取接口
// 注意：此方法并发安全，但 config 参数仅在本次调用中生效，不会修改提取器的默认配置
func (b *BaseFeatureExtractor) ExtractFeature(featureType FeatureType, data interface{}, config *FeatureConfig) (float64, bool) {
	// 使用传入的 config 或默认 config，不修改提取器的状态
	cfg := b.config
	if config != nil {
		cfg = config
	}

	switch featureType {
	case FeatureEntropy:
		return b.extractEntropyFeature(data, cfg)
	case FeatureToolMarker:
		return b.extractToolMarkerFeature(data, cfg)
	case FeatureOSPlatformContradiction:
		return b.extractOSPlatformContradictionFeature(data)
	case FeatureUAOSContradiction:
		return b.extractUAOSContradictionFeature(data)
	case FeatureMobileScreenContradiction:
		return b.extractMobileScreenContradictionFeature(data, cfg)
	case FeatureUAFeatureContradiction:
		return b.extractUAFeatureContradictionFeature(data)
	case FeatureHeadlessBrowser:
		return b.extractHeadlessBrowserFeature(data, cfg)
	default:
		return 0.0, false
	}
}

// GetFeatureName 获取特征的人类可读名称
func (b *BaseFeatureExtractor) GetFeatureName(featureType FeatureType) string {
	switch featureType {
	case FeatureEntropy:
		return "Entropy Anomaly"
	case FeatureToolMarker:
		return "Tool Marker Detection"
	case FeatureOSPlatformContradiction:
		return "OS/Platform Contradiction"
	case FeatureUAOSContradiction:
		return "UA/OS Contradiction"
	case FeatureMobileScreenContradiction:
		return "Mobile/Screen Contradiction"
	case FeatureUAFeatureContradiction:
		return "UA/Feature Contradiction"
	case FeatureHeadlessBrowser:
		return "Headless Browser Detection"
	default:
		return "Unknown Feature"
	}
}

// extractEntropyFeature 从字节数据提取熵特征
func (b *BaseFeatureExtractor) extractEntropyFeature(data interface{}, cfg *FeatureConfig) (float64, bool) {
	var bytes []byte

	// 类型转换
	switch v := data.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return 0.0, false
	}

	if len(bytes) < 10 {
		return 0.0, false
	}

	// 计算低熵（重复字节）
	var byteCounts [256]int
	for _, b := range bytes {
		byteCounts[b]++
	}
	uniqueBytes := 0
	for _, count := range byteCounts {
		if count > 0 {
			uniqueBytes++
		}
	}

	// 低熵异常
	if uniqueBytes < cfg.EntropyLowThreshold {
		return 0.95, true
	}

	// 计算高熵（Shannon 熵）
	if len(bytes) >= 20 {
		n := float64(len(bytes))
		entropy := 0.0
		for _, count := range byteCounts {
			if count > 0 {
				p := float64(count) / n
				entropy -= p * math.Log2(p)
			}
		}

		// 高熵异常
		if entropy > cfg.EntropyHighThreshold {
			return 0.85, true
		}
	}

	return 0.0, false
}

// extractToolMarkerFeature 检测工具特征
// 优化：对大文本使用 strings.Contains 进行高效匹配
func (b *BaseFeatureExtractor) extractToolMarkerFeature(data interface{}, cfg *FeatureConfig) (float64, bool) {
	var text string

	switch v := data.(type) {
	case []byte:
		// 如果数据量很大，先尝试转为字符串进行高效匹配
		if len(v) > 1024 {
			// 对大文本使用更高效的 Boyer-Moore 类算法（strings.Contains 内部实现）
			text = string(v)
		} else {
			// 小数据量使用逐字节匹配
			return b.extractToolMarkerFromBytes(v, cfg.ToolMarkers)
		}
	case string:
		text = v
	default:
		return 0.0, false
	}

	// 使用 strings.Contains 进行高效匹配（内部使用优化的字符串搜索算法）
	textLower := strings.ToLower(text)
	for _, pattern := range cfg.ToolMarkers {
		if strings.Contains(textLower, strings.ToLower(pattern)) {
			return 0.9, true
		}
	}

	return 0.0, false
}

// extractToolMarkerFromBytes 对小字节切片进行工具特征检测
func (b *BaseFeatureExtractor) extractToolMarkerFromBytes(data []byte, patterns []string) (float64, bool) {
	dataLower := bytes.ToLower(data)
	for _, pattern := range patterns {
		if bytes.Contains(dataLower, bytes.ToLower([]byte(pattern))) {
			return 0.9, true
		}
	}
	return 0.0, false
}

// extractOSPlatformContradictionFeature 操作系统和平台矛盾
func (b *BaseFeatureExtractor) extractOSPlatformContradictionFeature(data interface{}) (float64, bool) {
	attrs := toStringMap(data)
	if attrs == nil {
		return 0.0, false
	}

	os := attrs["os"]
	platform := attrs["platform"]

	if os == "" || platform == "" {
		return 0.0, false
	}

	// Windows OS 不应搭配非 Win platform
	if strings.Contains(os, "Windows") && !strings.Contains(platform, "Win") {
		return 0.8, true
	}
	// Mac OS 不应搭配非 Mac platform
	if strings.Contains(os, "Mac") && !strings.Contains(platform, "Mac") {
		return 0.8, true
	}
	// Linux 不应搭配非 X11/Linux platform
	if strings.Contains(os, "Linux") && !strings.Contains(platform, "Linux") && !strings.Contains(platform, "X11") {
		return 0.8, true
	}

	return 0.0, false
}

// extractUAOSContradictionFeature User-Agent 和操作系统矛盾
func (b *BaseFeatureExtractor) extractUAOSContradictionFeature(data interface{}) (float64, bool) {
	attrs := toStringMap(data)
	if attrs == nil {
		return 0.0, false
	}

	ua := attrs["user_agent"]
	os := attrs["os"]

	if ua == "" || os == "" {
		return 0.0, false
	}

	uaLower := strings.ToLower(ua)
	osLower := strings.ToLower(os)

	// Windows UA 声称 Mac OS
	if strings.Contains(uaLower, "windows") && strings.Contains(osLower, "mac") {
		return 0.9, true
	}
	// Mac UA 声称 Windows OS
	if strings.Contains(uaLower, "macintosh") && strings.Contains(osLower, "windows") {
		return 0.9, true
	}
	// Linux UA 声称 Windows OS
	if strings.Contains(uaLower, "x11; linux") && strings.Contains(osLower, "windows") {
		return 0.9, true
	}

	return 0.0, false
}

// extractMobileScreenContradictionFeature 移动设备屏幕分辨率矛盾
