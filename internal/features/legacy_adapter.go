package features

import "strings"

// LegacyFeatureAdapter 遗留特征适配器
// 为现有 defense.go 的 API 提供兼容层，使用新的 Feature Extractor 引擎
type LegacyFeatureAdapter struct {
	extractor *BaseFeatureExtractor
	config    *FeatureConfig
}

// NewLegacyFeatureAdapter 创建新的遗留特征适配器
func NewLegacyFeatureAdapter(config *FeatureConfig) *LegacyFeatureAdapter {
	if config == nil {
		config = DefaultFeatureConfig()
	}
	return &LegacyFeatureAdapter{
		extractor: NewBaseFeatureExtractor(config),
		config:    config,
	}
}

// DetectAnomalies 兼容 defense.go 的异常检测 API
func (a *LegacyFeatureAdapter) DetectAnomalies(data []byte) bool {
	_, isAnomaly := a.extractor.ExtractFeature(FeatureEntropy, data, a.config)
	if isAnomaly {
		return true
	}

	_, isAnomaly = a.extractor.ExtractFeature(FeatureToolMarker, data, a.config)
	return isAnomaly
}

// HasLowEntropy 检查数据熵值是否过低（兼容 defense.go）
func (a *LegacyFeatureAdapter) HasLowEntropy(data []byte) bool {
	score, isAnomaly := a.extractor.ExtractFeature(FeatureEntropy, data, a.config)
	return isAnomaly && score > 0.9
}

// HasExcessiveEntropy 检查数据熵值是否过高（兼容 defense.go）
func (a *LegacyFeatureAdapter) HasExcessiveEntropy(data []byte) bool {
	score, isAnomaly := a.extractor.ExtractFeature(FeatureEntropy, data, a.config)
	return isAnomaly && score > 0.8 && score < 0.95
}

// ContainsSpoofingMarkers 检查是否包含已知自动化工具特征（兼容 defense.go）
func (a *LegacyFeatureAdapter) ContainsSpoofingMarkers(data []byte) bool {
	_, isAnomaly := a.extractor.ExtractFeature(FeatureToolMarker, data, a.config)
	return isAnomaly
}

// DetectHeadlessBrowser 检测 User-Agent 是否为无头浏览器
func (a *LegacyFeatureAdapter) DetectHeadlessBrowser(userAgent string) bool {
	_, isAnomaly := a.extractor.ExtractFeature(FeatureHeadlessBrowser, userAgent, a.config)
	return isAnomaly
}

// CheckContradictions 检查指纹属性是否存在矛盾
func (a *LegacyFeatureAdapter) CheckContradictions(attributes map[string]string) bool {
	if len(attributes) == 0 {
		return false
	}

	// 检查操作系统与平台矛盾
	if os, ok := attributes["os"]; ok {
		if platform, ok := attributes["platform"]; ok {
			score, isAnomaly := a.extractor.ExtractFeature(
				FeatureOSPlatformContradiction,
				map[string]string{"os": os, "platform": platform},
				a.config,
			)
			if isAnomaly {
				_ = score // 使用 score 防止未使用警告
				return true
			}
		}
	}

	// 检查 User-Agent 与操作系统矛盾
	if ua, ok := attributes["user_agent"]; ok {
		if os, ok := attributes["os"]; ok {
			score, isAnomaly := a.extractor.ExtractFeature(
				FeatureUAOSContradiction,
				map[string]string{"user_agent": ua, "os": os},
				a.config,
			)
			if isAnomaly {
				_ = score
				return true
			}
		}
	}

	// 检查 User-Agent 与特性矛盾
	if ua, ok := attributes["user_agent"]; ok {
		if features, ok := attributes["features"]; ok {
			score, isAnomaly := a.extractor.ExtractFeature(
				FeatureUAFeatureContradiction,
				map[string]string{"user_agent": ua, "features": features},
				a.config,
			)
			if isAnomaly {
				_ = score
				return true
			}
		}
	}

	// 检查移动设备与屏幕分辨率矛盾
	if isMobile, ok := attributes["is_mobile"]; ok {
		if screenWidth, ok := attributes["screen_width"]; ok {
			score, isAnomaly := a.extractor.ExtractFeature(
				FeatureMobileScreenContradiction,
				map[string]string{"is_mobile": isMobile, "screen_width": screenWidth},
				a.config,
			)
			if isAnomaly {
				_ = score
				return true
			}
		}
	}

	return false
}

// RecognitionResultLegacy 被动识别结果（兼容 defense.go）
type RecognitionResultLegacy struct {
	Browser        string
	OS             string
	BrowserVersion string
	Confidence     float64
	IsMobile       bool
	IsBot          bool
}

// RecognizeFromHeaders 从 HTTP 请求头识别浏览器指纹（兼容 defense.go）
func (a *LegacyFeatureAdapter) RecognizeFromHeaders(headers map[string]string) *RecognitionResultLegacy {
	result := &RecognitionResultLegacy{}

	ua := headers["User-Agent"]
	if ua == "" {
		result.Confidence = 0.0
		return result
	}

	// 检测机器人
	if a.DetectHeadlessBrowser(ua) {
		result.IsBot = true
		result.Confidence = 0.9
		return result
	}

	uaLower := strings.ToLower(ua)

	// 检测移动设备
	result.IsMobile = strings.Contains(uaLower, "mobile") ||
		strings.Contains(uaLower, "android") ||
		strings.Contains(uaLower, "iphone") ||
		strings.Contains(uaLower, "ipad")

	// 识别浏览器
	result.Browser, result.BrowserVersion = detectBrowserFromUALegacy(ua)

	// 识别操作系统
	result.OS = detectOSFromUALegacy(ua)

	// 计算置信度
	result.Confidence = calculateConfidenceLegacy(headers)

	return result
}

// 辅助函数：从 User-Agent 检测浏览器类型和版本
func detectBrowserFromUALegacy(ua string) (string, string) {
	uaLower := strings.ToLower(ua)

	// Edge 必须在 Chrome 之前检测
	if strings.Contains(uaLower, "edg/") || strings.Contains(uaLower, "edge/") {
		version := extractVersionFromUALegacy(ua, "Edg/")
		if version == "" {
			version = extractVersionFromUALegacy(ua, "Edge/")
		}
		return "Edge", version
	}

	// Opera 必须在 Chrome 之前检测
	if strings.Contains(uaLower, "opr/") {
		version := extractVersionFromUALegacy(ua, "OPR/")
		return "Opera", version
	}

	// Chrome
	if strings.Contains(uaLower, "chrome/") && !strings.Contains(uaLower, "chromium") {
		version := extractVersionFromUALegacy(ua, "Chrome/")
		return "Chrome", version
	}

	// Firefox
	if strings.Contains(uaLower, "firefox/") {
		version := extractVersionFromUALegacy(ua, "Firefox/")
		return "Firefox", version
	}

	// Safari
	if strings.Contains(uaLower, "safari/") {
		version := extractVersionFromUALegacy(ua, "Version/")
		return "Safari", version
	}

	return "Chrome", ""
}

// 从 User-Agent 提取版本号
func extractVersionFromUALegacy(ua, prefix string) string {
	idx := strings.Index(ua, prefix)
	if idx == -1 {
		return ""
	}
	start := idx + len(prefix)
	end := start
	for end < len(ua) && ua[end] != '.' && ua[end] != ' ' && ua[end] != ';' {
		end++
	}
	if end > start {
		return ua[start:end]
	}
	return ""
}

// 从 User-Agent 检测操作系统
func detectOSFromUALegacy(ua string) string {
	if strings.Contains(ua, "Windows NT 10.0") {
		return "Windows 10"
	}
	if strings.Contains(ua, "Macintosh; Intel Mac OS X 15") {
		return "MacOS 15"
	}
	if strings.Contains(ua, "Macintosh; Intel Mac OS X 14") {
		return "MacOS 14"
	}
	if strings.Contains(ua, "Macintosh; Intel Mac OS X 13") {
		return "MacOS 13"
	}
	if strings.Contains(ua, "X11; Linux") {
		return "Linux"
	}
	return "Windows 10"
}

// 计算识别置信度
func calculateConfidenceLegacy(headers map[string]string) float64 {
	score := 0.5

	if headers["User-Agent"] != "" {
		score += 0.2
	}
	if headers["Accept"] != "" {
		score += 0.1
	}
	if headers["Accept-Language"] != "" {
		score += 0.1
	}
	if headers["Sec-CH-UA"] != "" {
		score += 0.1
	}

	if score > 1.0 {
		score = 1.0
	}
	return score
}
