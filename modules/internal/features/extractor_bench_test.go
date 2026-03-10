package features

import (
	"testing"
)

// translated comment
func BenchmarkExtractFeature_Entropy(b *testing.B) {
	extractor := NewBaseFeatureExtractor(nil)
	config := DefaultFeatureConfig()
	data := []byte("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = extractor.ExtractFeature(FeatureEntropy, data, config)
	}
}

// translated comment
func BenchmarkExtractFeature_ToolMarker(b *testing.B) {
	extractor := NewBaseFeatureExtractor(nil)
	config := DefaultFeatureConfig()
	data := []byte("Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = extractor.ExtractFeature(FeatureToolMarker, data, config)
	}
}

// translated comment
func BenchmarkExtractFeature_HeadlessBrowser(b *testing.B) {
	extractor := NewBaseFeatureExtractor(nil)
	config := DefaultFeatureConfig()
	ua := "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = extractor.ExtractFeature(FeatureHeadlessBrowser, ua, config)
	}
}

// translated comment
func BenchmarkExtractFeature_OSPlatformContradiction(b *testing.B) {
	extractor := NewBaseFeatureExtractor(nil)
	config := DefaultFeatureConfig()
	data := map[string]string{
		"os":       "Windows",
		"platform": "macOS",
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = extractor.ExtractFeature(FeatureOSPlatformContradiction, data, config)
	}
}

// translated comment
func BenchmarkBaseFeatureExtractor_AllFeatures(b *testing.B) {
	extractor := NewBaseFeatureExtractor(nil)
	config := DefaultFeatureConfig()

	features := []FeatureType{
		FeatureEntropy,
		FeatureToolMarker,
		FeatureHeadlessBrowser,
		FeatureOSPlatformContradiction,
		FeatureUAOSContradiction,
		FeatureUAFeatureContradiction,
		FeatureMobileScreenContradiction,
	}

	data := []byte("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, feature := range features {
			_, _ = extractor.ExtractFeature(feature, data, config)
		}
	}
}

// translated comment
func BenchmarkLegacyFeatureAdapter_DetectAnomalies(b *testing.B) {
	adapter := NewLegacyFeatureAdapter(nil)
	data := []byte("Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = adapter.DetectAnomalies(data)
	}
}

// translated comment
func BenchmarkLegacyFeatureAdapter_CheckContradictions(b *testing.B) {
	adapter := NewLegacyFeatureAdapter(nil)
	attributes := map[string]string{
		"os":         "Windows",
		"platform":   "macOS",
		"user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = adapter.CheckContradictions(attributes)
	}
}

// translated comment
func BenchmarkLegacyFeatureAdapter_RecognizeFromHeaders(b *testing.B) {
	adapter := NewLegacyFeatureAdapter(nil)
	headers := map[string]string{
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0",
		"Accept":          "text/html,application/xhtml+xml",
		"Accept-Language": "en-US,en;q=0.9",
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = adapter.RecognizeFromHeaders(headers)
	}
}

// translated comment
func BenchmarkCalculateRiskScore(b *testing.B) {
	scores := map[FeatureType]float64{
		FeatureEntropy:                   0.8,
		FeatureToolMarker:                0.9,
		FeatureHeadlessBrowser:           0.7,
		FeatureOSPlatformContradiction:   0.6,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = calculateRiskScore(scores)
	}
}
