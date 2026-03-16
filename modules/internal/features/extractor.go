package features

import (
	"bytes"
	"math"
	"strings"
)

// FeatureType enumerates feature categories
type FeatureType string

const (
	// Entropy feature
	FeatureEntropy FeatureType = "entropy"
	// Tool marker feature
	FeatureToolMarker FeatureType = "tool_marker"
	// OS/platform contradiction feature
	FeatureOSPlatformContradiction FeatureType = "os_platform_contradiction"
	// User-Agent/OS contradiction feature
	FeatureUAOSContradiction FeatureType = "ua_os_contradiction"
	// Mobile device/screen resolution contradiction feature
	FeatureMobileScreenContradiction FeatureType = "mobile_screen_contradiction"
	// User-Agent/feature contradiction feature
	FeatureUAFeatureContradiction FeatureType = "ua_feature_contradiction"
	// Headless browser feature
	FeatureHeadlessBrowser FeatureType = "headless_browser"
)

// FeatureExtractor defines the unified feature extraction interface
type FeatureExtractor interface {
	// ExtractFeature extracts a feature of the given type from input data
	// Returns the feature score (0.0-1.0) and whether an anomaly is detected
	ExtractFeature(featureType FeatureType, data interface{}, config *FeatureConfig) (float64, bool)

	// GetFeatureName returns a human-readable feature name
	GetFeatureName(featureType FeatureType) string
}

// FeatureConfig holds feature extraction settings
type FeatureConfig struct {
	// High entropy threshold (bits)
	EntropyHighThreshold float64 `json:"entropy_high_threshold"`
	// Low entropy threshold (unique byte count)
	EntropyLowThreshold int `json:"entropy_low_threshold"`
	// Tool marker pattern list
	ToolMarkers []string `json:"tool_markers"`
	// Headless browser marker list
	HeadlessMarkers []string `json:"headless_markers"`
	// Max mobile screen width (larger values are suspicious)
	MobileScreenWidthMax int `json:"mobile_screen_width_max"`
	// Min desktop screen width (smaller values are suspicious)
	DesktopScreenWidthMin int `json:"desktop_screen_width_min"`
}

// DefaultFeatureConfig returns default feature settings
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

// BaseFeatureExtractor implements core feature extraction
type BaseFeatureExtractor struct {
	config *FeatureConfig
}

// NewBaseFeatureExtractor creates a new base extractor
func NewBaseFeatureExtractor(config *FeatureConfig) *BaseFeatureExtractor {
	if config == nil {
		config = DefaultFeatureConfig()
	}
	return &BaseFeatureExtractor{config: config}
}

// ExtractFeature implements the unified extraction interface
// Note: this method is concurrency-safe, and config only applies to this call without mutating default extractor state
func (b *BaseFeatureExtractor) ExtractFeature(featureType FeatureType, data interface{}, config *FeatureConfig) (float64, bool) {
	// Use the provided config or fallback to the default without mutating extractor state
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

// GetFeatureName returns a human-readable feature name
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

// extractEntropyFeature extracts entropy-based features from byte data
func (b *BaseFeatureExtractor) extractEntropyFeature(data interface{}, cfg *FeatureConfig) (float64, bool) {
	var bytes []byte

	// Type conversion
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

	// Calculate low entropy (repeated-byte pattern)
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

	// Low-entropy anomaly
	if uniqueBytes < cfg.EntropyLowThreshold {
		return 0.95, true
	}

	// Calculate high entropy (Shannon entropy)
	if len(bytes) >= 20 {
		n := float64(len(bytes))
		entropy := 0.0
		for _, count := range byteCounts {
			if count > 0 {
				p := float64(count) / n
				entropy -= p * math.Log2(p)
			}
		}

		// High-entropy anomaly
		if entropy > cfg.EntropyHighThreshold {
			return 0.85, true
		}
	}

	return 0.0, false
}

// extractToolMarkerFeature detects automation tool markers
// Optimization: use strings.Contains for efficient matching on large text
func (b *BaseFeatureExtractor) extractToolMarkerFeature(data interface{}, cfg *FeatureConfig) (float64, bool) {
	var text string

	switch v := data.(type) {
	case []byte:
		// For large payloads, convert to string first for efficient matching
		if len(v) > 1024 {
			// For large text, rely on optimized string search used by strings.Contains
			text = string(v)
		} else {
			// For small payloads, use byte-level matching
			return b.extractToolMarkerFromBytes(v, cfg.ToolMarkers)
		}
	case string:
		text = v
	default:
		return 0.0, false
	}

	// Use strings.Contains for efficient matching
	textLower := strings.ToLower(text)
	for _, pattern := range cfg.ToolMarkers {
		if strings.Contains(textLower, strings.ToLower(pattern)) {
			return 0.9, true
		}
	}

	return 0.0, false
}

// extractToolMarkerFromBytes detects tool markers on small byte slices
func (b *BaseFeatureExtractor) extractToolMarkerFromBytes(data []byte, patterns []string) (float64, bool) {
	dataLower := bytes.ToLower(data)
	for _, pattern := range patterns {
		if bytes.Contains(dataLower, bytes.ToLower([]byte(pattern))) {
			return 0.9, true
		}
	}
	return 0.0, false
}

// extractOSPlatformContradictionFeature checks OS/platform contradictions
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

	// Windows OS should pair with a Win platform
	if strings.Contains(os, "Windows") && !strings.Contains(platform, "Win") {
		return 0.8, true
	}
	// Mac OS should pair with a Mac platform
	if strings.Contains(os, "Mac") && !strings.Contains(platform, "Mac") {
		return 0.8, true
	}
	// Linux OS should pair with X11/Linux platform
	if strings.Contains(os, "Linux") && !strings.Contains(platform, "Linux") && !strings.Contains(platform, "X11") {
		return 0.8, true
	}

	return 0.0, false
}

// extractUAOSContradictionFeature checks User-Agent/OS contradictions
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

	// Windows UA claims a Mac OS
	if strings.Contains(uaLower, "windows") && strings.Contains(osLower, "mac") {
		return 0.9, true
	}
	// Mac UA claims a Windows OS
	if strings.Contains(uaLower, "macintosh") && strings.Contains(osLower, "windows") {
		return 0.9, true
	}
	// Linux UA claims a Windows OS
	if strings.Contains(uaLower, "x11; linux") && strings.Contains(osLower, "windows") {
		return 0.9, true
	}

	return 0.0, false
}

// extractMobileScreenContradictionFeature checks mobile/screen contradictions
