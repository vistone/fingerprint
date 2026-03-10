package features

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
)

// translated comment
type FeatureType string

const (
	// translated comment
	FeatureEntropy FeatureType = "entropy"
	// translated comment
	FeatureToolMarker FeatureType = "tool_marker"
	// translated comment
	FeatureOSPlatformContradiction FeatureType = "os_platform_contradiction"
	// translated comment
	FeatureUAOSContradiction FeatureType = "ua_os_contradiction"
	// translated comment
	FeatureMobileScreenContradiction FeatureType = "mobile_screen_contradiction"
	// translated comment
	FeatureUAFeatureContradiction FeatureType = "ua_feature_contradiction"
	// translated comment
	FeatureHeadlessBrowser FeatureType = "headless_browser"
)

// translated comment
type FeatureExtractor interface {
	// translated comment
	// translated comment
	ExtractFeature(featureType FeatureType, data interface{}, config *FeatureConfig) (float64, bool)

	// translated comment
	GetFeatureName(featureType FeatureType) string
}

// translated comment
type FeatureConfig struct {
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

// translated comment
type BaseFeatureExtractor struct {
	config *FeatureConfig
}

// translated comment
func NewBaseFeatureExtractor(config *FeatureConfig) *BaseFeatureExtractor {
	if config == nil {
		config = DefaultFeatureConfig()
	}
	return &BaseFeatureExtractor{config: config}
}

// translated comment
// translated comment
func (b *BaseFeatureExtractor) ExtractFeature(featureType FeatureType, data interface{}, config *FeatureConfig) (float64, bool) {
	// translated comment
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

// translated comment
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

// translated comment
func (b *BaseFeatureExtractor) extractEntropyFeature(data interface{}, cfg *FeatureConfig) (float64, bool) {
	var bytes []byte

	// translated comment
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

	// translated comment
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

	// translated comment
	if uniqueBytes < cfg.EntropyLowThreshold {
		return 0.95, true
	}

	// translated comment
	if len(bytes) >= 20 {
		n := float64(len(bytes))
		entropy := 0.0
		for _, count := range byteCounts {
			if count > 0 {
				p := float64(count) / n
				entropy -= p * math.Log2(p)
			}
		}

		// translated comment
		if entropy > cfg.EntropyHighThreshold {
			return 0.85, true
		}
	}

	return 0.0, false
}

// translated comment
// translated comment
func (b *BaseFeatureExtractor) extractToolMarkerFeature(data interface{}, cfg *FeatureConfig) (float64, bool) {
	var text string

	switch v := data.(type) {
	case []byte:
		// translated comment
		if len(v) > 1024 {
			// translated comment
			text = string(v)
		} else {
			// translated comment
			return b.extractToolMarkerFromBytes(v, cfg.ToolMarkers)
		}
	case string:
		text = v
	default:
		return 0.0, false
	}

	// translated comment
	textLower := strings.ToLower(text)
	for _, pattern := range cfg.ToolMarkers {
		if strings.Contains(textLower, strings.ToLower(pattern)) {
			return 0.9, true
		}
	}

	return 0.0, false
}

// translated comment
func (b *BaseFeatureExtractor) extractToolMarkerFromBytes(data []byte, patterns []string) (float64, bool) {
	dataLower := bytes.ToLower(data)
	for _, pattern := range patterns {
		if bytes.Contains(dataLower, bytes.ToLower([]byte(pattern))) {
			return 0.9, true
		}
	}
	return 0.0, false
}

// translated comment
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

	// translated comment
	if strings.Contains(os, "Windows") && !strings.Contains(platform, "Win") {
		return 0.8, true
	}
	// translated comment
	if strings.Contains(os, "Mac") && !strings.Contains(platform, "Mac") {
		return 0.8, true
	}
	// translated comment
	if strings.Contains(os, "Linux") && !strings.Contains(platform, "Linux") && !strings.Contains(platform, "X11") {
		return 0.8, true
	}

	return 0.0, false
}

// translated comment
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

	// translated comment
	if strings.Contains(uaLower, "windows") && strings.Contains(osLower, "mac") {
		return 0.9, true
	}
	// translated comment
	if strings.Contains(uaLower, "macintosh") && strings.Contains(osLower, "windows") {
		return 0.9, true
	}
	// translated comment
	if strings.Contains(uaLower, "x11; linux") && strings.Contains(osLower, "windows") {
		return 0.9, true
	}

	return 0.0, false
}

// translated comment
func (b *BaseFeatureExtractor) extractMobileScreenContradictionFeature(data interface{}, cfg *FeatureConfig) (float64, bool) {
	attrs := toStringMap(data)
	if attrs == nil {
		return 0.0, false
	}

	isMobile := attrs["is_mobile"]
	screenWidth := attrs["screen_width"]

	if isMobile == "" || screenWidth == "" {
		return 0.0, false
	}

	// translated comment
	var width int
	_, err := fmt.Sscanf(screenWidth, "%d", &width)
	if err != nil {
		return 0.0, false
	}

	// translated comment
	if isMobile == "true" && width > cfg.MobileScreenWidthMax {
		return 0.85, true
	}
	// translated comment
	if isMobile == "false" && width < cfg.DesktopScreenWidthMin {
		return 0.85, true
	}

	return 0.0, false
}

// translated comment
func (b *BaseFeatureExtractor) extractUAFeatureContradictionFeature(data interface{}) (float64, bool) {
	attrs := toStringMap(data)
	if attrs == nil {
		return 0.0, false
	}

	ua := attrs["user_agent"]
	features := attrs["features"]

	if ua == "" || features == "" {
		return 0.0, false
	}

	uaLower := strings.ToLower(ua)
	featuresLower := strings.ToLower(features)

	// translated comment
	if strings.Contains(uaLower, "chrome/60") && strings.Contains(featuresLower, "webgl2") {
		return 0.8, true
	}
	// translated comment
	if strings.Contains(uaLower, "mobile") && strings.Contains(featuresLower, "desktop") {
		return 0.8, true
	}

	return 0.0, false
}

// translated comment
func (b *BaseFeatureExtractor) extractHeadlessBrowserFeature(data interface{}, cfg *FeatureConfig) (float64, bool) {
	var ua string

	switch v := data.(type) {
	case string:
		ua = v
	case map[string]string:
		ua = v["user_agent"]
	default:
		return 0.0, false
	}

	if ua == "" {
		return 0.0, false
	}

	uaLower := strings.ToLower(ua)
	for _, marker := range cfg.HeadlessMarkers {
		if strings.Contains(uaLower, strings.ToLower(marker)) {
			return 0.95, true
		}
	}

	return 0.0, false
}

// translated comment
func memEquals(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// translated comment
// translated comment
func toStringMap(data interface{}) map[string]string {
	switch v := data.(type) {
	case map[string]string:
		return v
	case map[string]interface{}:
		result := make(map[string]string, len(v))
		for key, val := range v {
			if s, ok := val.(string); ok {
				result[key] = s
			}
		}
		return result
	default:
		return nil
	}
}

// translated comment
type FeatureVector struct {
	// translated comment
	Scores map[FeatureType]float64
	// translated comment
	Hash string
	// translated comment
	Anomalies []FeatureType
	// translated comment
	RiskScore float64
}

// translated comment
// translated comment
func (b *BaseFeatureExtractor) ExtractFeatureVector(data map[string]interface{}, config *FeatureConfig) *FeatureVector {
	// translated comment
	cfg := b.config
	if config != nil {
		cfg = config
	}

	vector := &FeatureVector{
		Scores:    make(map[FeatureType]float64),
		Anomalies: []FeatureType{},
	}

	// translated comment
	featuresToExtract := []FeatureType{
		FeatureEntropy,
		FeatureToolMarker,
		FeatureHeadlessBrowser,
		FeatureOSPlatformContradiction,
		FeatureUAOSContradiction,
		FeatureMobileScreenContradiction,
		FeatureUAFeatureContradiction,
	}

	// translated comment
	for _, fType := range featuresToExtract {
		score, isAnomaly := b.ExtractFeature(fType, data, cfg)
		vector.Scores[fType] = score
		if isAnomaly {
			vector.Anomalies = append(vector.Anomalies, fType)
		}
	}

	// translated comment
	vector.RiskScore = calculateRiskScore(vector.Scores)

	// translated comment
	var hashInput strings.Builder
	hashInput.Grow(len(featuresToExtract) * 20)
	for _, fType := range featuresToExtract {
		fmt.Fprintf(&hashInput, "%s:%.2f;", fType, vector.Scores[fType])
	}
	h := md5.Sum([]byte(hashInput.String()))
	vector.Hash = hex.EncodeToString(h[:])

	return vector
}

// translated comment
func calculateRiskScore(scores map[FeatureType]float64) float64 {
	if len(scores) == 0 {
		return 0.0
	}

	// translated comment
	maxScore := 0.0
	for _, score := range scores {
		if score > maxScore {
			maxScore = score
		}
	}

	// translated comment
	anomalyCount := 0
	for _, score := range scores {
		if score > 0.5 {
			anomalyCount++
		}
	}

	// translated comment
	weightedScore := maxScore * (1.0 + float64(anomalyCount-1)*0.1)
	if weightedScore > 1.0 {
		weightedScore = 1.0
	}

	return weightedScore
}
