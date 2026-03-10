package features

import "strings"

// translated comment
// translated comment
type LegacyFeatureAdapter struct {
	extractor *BaseFeatureExtractor
	config    *FeatureConfig
}

// translated comment
func NewLegacyFeatureAdapter(config *FeatureConfig) *LegacyFeatureAdapter {
	if config == nil {
		config = DefaultFeatureConfig()
	}
	return &LegacyFeatureAdapter{
		extractor: NewBaseFeatureExtractor(config),
		config:    config,
	}
}

// translated comment
func (a *LegacyFeatureAdapter) DetectAnomalies(data []byte) bool {
	_, isAnomaly := a.extractor.ExtractFeature(FeatureEntropy, data, a.config)
	if isAnomaly {
		return true
	}

	_, isAnomaly = a.extractor.ExtractFeature(FeatureToolMarker, data, a.config)
	return isAnomaly
}

// translated comment
func (a *LegacyFeatureAdapter) HasLowEntropy(data []byte) bool {
	score, isAnomaly := a.extractor.ExtractFeature(FeatureEntropy, data, a.config)
	return isAnomaly && score > 0.9
}

// translated comment
func (a *LegacyFeatureAdapter) HasExcessiveEntropy(data []byte) bool {
	score, isAnomaly := a.extractor.ExtractFeature(FeatureEntropy, data, a.config)
	return isAnomaly && score > 0.8 && score < 0.95
}

// translated comment
func (a *LegacyFeatureAdapter) ContainsSpoofingMarkers(data []byte) bool {
	_, isAnomaly := a.extractor.ExtractFeature(FeatureToolMarker, data, a.config)
	return isAnomaly
}

// translated comment
func (a *LegacyFeatureAdapter) DetectHeadlessBrowser(userAgent string) bool {
	_, isAnomaly := a.extractor.ExtractFeature(FeatureHeadlessBrowser, userAgent, a.config)
	return isAnomaly
}

// translated comment
func (a *LegacyFeatureAdapter) CheckContradictions(attributes map[string]string) bool {
	if len(attributes) == 0 {
		return false
	}

	// translated comment
	if os, ok := attributes["os"]; ok {
		if platform, ok := attributes["platform"]; ok {
			score, isAnomaly := a.extractor.ExtractFeature(
				FeatureOSPlatformContradiction,
				map[string]string{"os": os, "platform": platform},
				a.config,
			)
			if isAnomaly {
				_ = score // translated comment
				return true
			}
		}
	}

	// translated comment
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

	// translated comment
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

	// translated comment
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

// translated comment
type RecognitionResultLegacy struct {
	Browser        string
	OS             string
	BrowserVersion string
	Confidence     float64
	IsMobile       bool
	IsBot          bool
}

// translated comment
func (a *LegacyFeatureAdapter) RecognizeFromHeaders(headers map[string]string) *RecognitionResultLegacy {
	result := &RecognitionResultLegacy{}

	ua := headers["User-Agent"]
	if ua == "" {
		result.Confidence = 0.0
		return result
	}

	// translated comment
	if a.DetectHeadlessBrowser(ua) {
		result.IsBot = true
		result.Confidence = 0.9
		return result
	}

	uaLower := strings.ToLower(ua)

	// translated comment
	result.IsMobile = strings.Contains(uaLower, "mobile") ||
		strings.Contains(uaLower, "android") ||
		strings.Contains(uaLower, "iphone") ||
		strings.Contains(uaLower, "ipad")

	// translated comment
	result.Browser, result.BrowserVersion = detectBrowserFromUALegacy(ua)

	// translated comment
	result.OS = detectOSFromUALegacy(ua)

	// translated comment
	result.Confidence = calculateConfidenceLegacy(headers)

	return result
}

// translated comment
func detectBrowserFromUALegacy(ua string) (string, string) {
	uaLower := strings.ToLower(ua)

	// translated comment
	if strings.Contains(uaLower, "edg/") || strings.Contains(uaLower, "edge/") {
		version := extractVersionFromUALegacy(ua, "Edg/")
		if version == "" {
			version = extractVersionFromUALegacy(ua, "Edge/")
		}
		return "Edge", version
	}

	// translated comment
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

// translated comment
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

// translated comment
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

// translated comment
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
