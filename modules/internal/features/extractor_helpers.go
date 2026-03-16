package features

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
)

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

	// Parse width
	var width int
	_, err := fmt.Sscanf(screenWidth, "%d", &width)
	if err != nil {
		return 0.0, false
	}

	// Mobile devices should not report excessively large widths
	if isMobile == "true" && width > cfg.MobileScreenWidthMax {
		return 0.85, true
	}
	// Desktop devices should not report extremely small widths
	if isMobile == "false" && width < cfg.DesktopScreenWidthMin {
		return 0.85, true
	}

	return 0.0, false
}

// extractUAFeatureContradictionFeature checks User-Agent/feature contradictions
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

	// Chrome 60 should not claim WebGL2 support
	if strings.Contains(uaLower, "chrome/60") && strings.Contains(featuresLower, "webgl2") {
		return 0.8, true
	}
	// Mobile UA should not claim desktop-only features
	if strings.Contains(uaLower, "mobile") && strings.Contains(featuresLower, "desktop") {
		return 0.8, true
	}

	return 0.0, false
}

// extractHeadlessBrowserFeature detects headless browser indicators
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

// memEquals compares two byte slices for equality
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

// toStringMap converts interface{} into map[string]string
// Supports map[string]string and map[string]interface{} inputs
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

// FeatureVector represents a complete feature vector
type FeatureVector struct {
	// Score per feature (0.0-1.0)
	Scores map[FeatureType]float64
	// MD5 hash (used for deduplication)
	Hash string
	// Detected anomaly types
	Anomalies []FeatureType
	// Overall risk score (0.0-1.0)
	RiskScore float64
}

// ExtractFeatureVector builds a full feature vector from multiple data sources
// Note: this method is concurrency-safe, and config only applies to this call
func (b *BaseFeatureExtractor) ExtractFeatureVector(data map[string]interface{}, config *FeatureConfig) *FeatureVector {
	// Use provided config or default config without mutating extractor state
	cfg := b.config
	if config != nil {
		cfg = config
	}

	vector := &FeatureVector{
		Scores:    make(map[FeatureType]float64),
		Anomalies: []FeatureType{},
	}

	// Define the feature list to extract
	featuresToExtract := []FeatureType{
		FeatureEntropy,
		FeatureToolMarker,
		FeatureHeadlessBrowser,
		FeatureOSPlatformContradiction,
		FeatureUAOSContradiction,
		FeatureMobileScreenContradiction,
		FeatureUAFeatureContradiction,
	}

	// Extract all features
	for _, fType := range featuresToExtract {
		score, isAnomaly := b.ExtractFeature(fType, data, cfg)
		vector.Scores[fType] = score
		if isAnomaly {
			vector.Anomalies = append(vector.Anomalies, fType)
		}
	}

	// Compute overall risk score
	vector.RiskScore = calculateRiskScore(vector.Scores)

	// Compute MD5 hash of the feature vector for deduplication
	var hashInput strings.Builder
	hashInput.Grow(len(featuresToExtract) * 20)
	for _, fType := range featuresToExtract {
		fmt.Fprintf(&hashInput, "%s:%.2f;", fType, vector.Scores[fType])
	}
	h := md5.Sum([]byte(hashInput.String()))
	vector.Hash = hex.EncodeToString(h[:])

	return vector
}

// calculateRiskScore computes an aggregated risk score
func calculateRiskScore(scores map[FeatureType]float64) float64 {
	if len(scores) == 0 {
		return 0.0
	}

	// Max-score strategy: use the highest-risk feature score
	maxScore := 0.0
	for _, score := range scores {
		if score > maxScore {
			maxScore = score
		}
	}

	// Apply weighting based on anomaly count
	anomalyCount := 0
	for _, score := range scores {
		if score > 0.5 {
			anomalyCount++
		}
	}

	// Multiple anomalies increase risk
	weightedScore := maxScore * (1.0 + float64(anomalyCount-1)*0.1)
	if weightedScore > 1.0 {
		weightedScore = 1.0
	}

	return weightedScore
}
