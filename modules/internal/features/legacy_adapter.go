package features

import "strings"

// LegacyFeatureAdapter provides backward-compatible feature APIs
// It wraps the new feature extractor engine to preserve the defense.go API
type LegacyFeatureAdapter struct {
	extractor *BaseFeatureExtractor
	config    *FeatureConfig
}

// NewLegacyFeatureAdapter creates a legacy feature adapter
func NewLegacyFeatureAdapter(config *FeatureConfig) *LegacyFeatureAdapter {
	if config == nil {
		config = DefaultFeatureConfig()
	}
	return &LegacyFeatureAdapter{
		extractor: NewBaseFeatureExtractor(config),
		config:    config,
	}
}

// DetectAnomalies keeps compatibility with the defense.go anomaly API
func (a *LegacyFeatureAdapter) DetectAnomalies(data []byte) bool {
	_, isAnomaly := a.extractor.ExtractFeature(FeatureEntropy, data, a.config)
	if isAnomaly {
		return true
	}

	_, isAnomaly = a.extractor.ExtractFeature(FeatureToolMarker, data, a.config)
	return isAnomaly
}

// HasLowEntropy checks whether entropy is too low (defense.go compatible)
func (a *LegacyFeatureAdapter) HasLowEntropy(data []byte) bool {
	score, isAnomaly := a.extractor.ExtractFeature(FeatureEntropy, data, a.config)
	return isAnomaly && score > 0.9
}

// HasExcessiveEntropy checks whether entropy is too high (defense.go compatible)
func (a *LegacyFeatureAdapter) HasExcessiveEntropy(data []byte) bool {
	score, isAnomaly := a.extractor.ExtractFeature(FeatureEntropy, data, a.config)
	return isAnomaly && score > 0.8 && score < 0.95
}

// ContainsSpoofingMarkers checks known automation markers (defense.go compatible)
func (a *LegacyFeatureAdapter) ContainsSpoofingMarkers(data []byte) bool {
	_, isAnomaly := a.extractor.ExtractFeature(FeatureToolMarker, data, a.config)
	return isAnomaly
}

// DetectHeadlessBrowser checks whether User-Agent indicates a headless browser
func (a *LegacyFeatureAdapter) DetectHeadlessBrowser(userAgent string) bool {
	_, isAnomaly := a.extractor.ExtractFeature(FeatureHeadlessBrowser, userAgent, a.config)
	return isAnomaly
}

// CheckContradictions validates contradictions in fingerprint attributes
func (a *LegacyFeatureAdapter) CheckContradictions(attributes map[string]string) bool {
	if len(attributes) == 0 {
		return false
	}

	// Check OS/platform contradictions
	if os, ok := attributes["os"]; ok {
		if platform, ok := attributes["platform"]; ok {
			score, isAnomaly := a.extractor.ExtractFeature(
				FeatureOSPlatformContradiction,
				map[string]string{"os": os, "platform": platform},
				a.config,
			)
			if isAnomaly {
				_ = score // Keep score to avoid unused variable warning
				return true
			}
		}
	}

	// Check User-Agent/OS contradictions
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

	// Check User-Agent/feature contradictions
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

	// Check mobile/screen resolution contradictions
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

// RecognitionResultLegacy stores passive recognition results (defense.go compatible)
type RecognitionResultLegacy struct {
	Browser        string
	OS             string
	BrowserVersion string
	Confidence     float64
	IsMobile       bool
	IsBot          bool
}

// RecognizeFromHeaders infers browser fingerprint info from HTTP headers (defense.go compatible)
func (a *LegacyFeatureAdapter) RecognizeFromHeaders(headers map[string]string) *RecognitionResultLegacy {
	result := &RecognitionResultLegacy{}

	ua := headers["User-Agent"]
	if ua == "" {
		result.Confidence = 0.0
		return result
	}

	// Detect bot indicators
	if a.DetectHeadlessBrowser(ua) {
		result.IsBot = true
		result.Confidence = 0.9
		return result
	}

	uaLower := strings.ToLower(ua)

	// Detect mobile device
	result.IsMobile = strings.Contains(uaLower, "mobile") ||
		strings.Contains(uaLower, "android") ||
		strings.Contains(uaLower, "iphone") ||
		strings.Contains(uaLower, "ipad")

	// Identify browser
	result.Browser, result.BrowserVersion = detectBrowserFromUALegacy(ua)

	// Identify operating system
	result.OS = detectOSFromUALegacy(ua)

	// Compute confidence
	result.Confidence = calculateConfidenceLegacy(headers)

	return result
}

// Helper: detect browser family and version from User-Agent
func detectBrowserFromUALegacy(ua string) (string, string) {
	uaLower := strings.ToLower(ua)

	// Edge must be checked before Chrome
	if strings.Contains(uaLower, "edg/") || strings.Contains(uaLower, "edge/") {
		version := extractVersionFromUALegacy(ua, "Edg/")
		if version == "" {
			version = extractVersionFromUALegacy(ua, "Edge/")
		}
		return "Edge", version
	}

	// Opera must be checked before Chrome
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

// Extract version token from User-Agent
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

// Detect operating system from User-Agent
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

// Compute recognition confidence
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
