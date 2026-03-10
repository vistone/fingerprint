// Package ml provides feature extraction functionality
package ml

import (
	"math"
	"sort"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/profiles"
)

// FeatureExtractor feature extractor
type FeatureExtractor struct {
	// Feature weights
	weights map[core.FeatureType]float64
}

// NewFeatureExtractor create new feature extractor
func NewFeatureExtractor() *FeatureExtractor {
	return &FeatureExtractor{
		weights: map[core.FeatureType]float64{
			core.FeatureTLSVersion:      1.0,
			core.FeatureCipherSuites:    1.0,
			core.FeatureExtensions:      1.0,
			core.FeatureHTTP2Settings:   0.8,
			core.FeatureHTTPHeaders:     0.9,
			core.FeatureUserAgent:       1.0,
			core.FeatureCanvas:          0.7,
			core.FeatureWebGL:           0.7,
			core.FeatureAudio:           0.6,
			core.FeatureFonts:           0.5,
			core.FeatureStorage:         0.4,
			core.FeatureWebRTC:          0.5,
			core.FeatureHardware:        0.3,
			core.FeatureTiming:          0.4,
			core.FeatureHeadlessBrowser: 0.8,
			core.FeatureEntropy:         0.9,
			core.FeatureToolMarker:      0.8,
			core.FeatureBehaviorPattern: 0.7,
		},
	}
}

// ExtractFromProfile extract feature vector from profile
func (fe *FeatureExtractor) ExtractFromProfile(profile *profiles.ClientProfile) *core.FeatureVector {
	fv := core.NewFeatureVector()

	// TLS features
	fv.Set(core.FeatureTLSVersion, float64(profile.TLSVersion))
	fv.Set(core.FeatureCipherSuites, float64(len(profile.CipherSuites)))
	fv.Set(core.FeatureExtensions, float64(len(profile.Extensions)))

	// HTTP/2 features
	fv.Set(core.FeatureHTTP2Settings, fe.hashHTTP2Settings(profile.HTTP2Settings))
	fv.Set(core.FeatureEntropy, fe.calculateEntropy(profile))

	// HTTP header features
	if profile.Headers != nil {
		fv.Set(core.FeatureHTTPHeaders, fe.hashHeaders(profile.Headers))
		fv.Set(core.FeatureUserAgent, fe.hashUserAgent(profile.Headers.UserAgent))
	}

	return fv
}

// ExtractFromClientHello extract features from ClientHello
func (fe *FeatureExtractor) ExtractFromClientHello(spec core.ClientHelloSpec) *core.FeatureVector {
	fv := core.NewFeatureVector()

	fv.Set(core.FeatureTLSVersion, float64(spec.TLSVersion))
	fv.Set(core.FeatureCipherSuites, float64(len(spec.CipherSuites)))
	fv.Set(core.FeatureExtensions, float64(len(spec.Extensions)))

	return fv
}

// ExtractFromHTTPHeaders extract features from HTTP headers
func (fe *FeatureExtractor) ExtractFromHTTPHeaders(headers *core.HTTPHeaders) *core.FeatureVector {
	fv := core.NewFeatureVector()

	if headers == nil {
		return fv
	}

	fv.Set(core.FeatureHTTPHeaders, fe.hashHeaders(headers))
	fv.Set(core.FeatureUserAgent, fe.hashUserAgent(headers.UserAgent))

	// Calculate header information entropy
	entropy := 0.0
	headerMap := headers.ToMap()
	for _, value := range headerMap {
		entropy += fe.calculateStringEntropy(value)
	}
	fv.Set(core.FeatureEntropy, entropy)

	return fv
}

// ExtractFromFrontend extract features from frontend fingerprint
func (fe *FeatureExtractor) ExtractFromFrontend(data FrontendFingerprintData) *core.FeatureVector {
	fv := core.NewFeatureVector()

	fv.Set(core.FeatureCanvas, data.Canvas.Entropy)
	fv.Set(core.FeatureWebGL, data.WebGL.Entropy)
	fv.Set(core.FeatureAudio, data.Audio.Entropy)
	fv.Set(core.FeatureFonts, float64(len(data.Fonts.List)))
	fv.Set(core.FeatureStorage, fe.hashStorage(data.Storage))
	fv.Set(core.FeatureWebRTC, fe.hashWebRTC(data.WebRTC))
	fv.Set(core.FeatureHardware, fe.hashHardware(data.Hardware))
	fv.Set(core.FeatureTiming, data.Timing.Precision)

	return fv
}

// ExtractCombined extract combined features
func (fe *FeatureExtractor) ExtractCombined(serverData ServerFingerprintData, frontendData FrontendFingerprintData) *core.FeatureVector {
	fv := core.NewFeatureVector()

	// Server-side features
	tlsFV := fe.ExtractFromClientHello(serverData.ClientHello)
	httpFV := fe.ExtractFromHTTPHeaders(serverData.Headers)

	// Frontend features
	frontendFV := fe.ExtractFromFrontend(frontendData)

	// Merge features
	for ft, v := range tlsFV.Features {
		fv.Set(ft, v*fe.weights[ft])
	}
	for ft, v := range httpFV.Features {
		if existing, ok := fv.Features[ft]; ok {
			fv.Set(ft, (existing+v)/2)
		} else {
			fv.Set(ft, v*fe.weights[ft])
		}
	}
	for ft, v := range frontendFV.Features {
		fv.Set(ft, v*fe.weights[ft])
	}

	return fv
}

// hashHTTP2Settings hash HTTP/2 Settings
func (fe *FeatureExtractor) hashHTTP2Settings(settings core.HTTP2Settings) float64 {
	// Simplified hash calculation
	hash := float64(settings.HeaderTableSize) * 0.1
	hash += float64(settings.MaxConcurrentStreams) * 0.01
	hash += float64(settings.InitialWindowSize) * 0.001
	return hash
}

// hashHeaders hash HTTP headers
func (fe *FeatureExtractor) hashHeaders(headers *core.HTTPHeaders) float64 {
	headerMap := headers.ToMap()
	var sum float64
	for name, value := range headerMap {
		sum += float64(len(name)) * float64(len(value))
	}
	return math.Log1p(sum)
}

// hashUserAgent hash User-Agent
func (fe *FeatureExtractor) hashUserAgent(ua string) float64 {
	if ua == "" {
		return 0
	}
	var sum float64
	for i, c := range ua {
		sum += float64(c) * float64(i+1)
	}
	return sum / float64(len(ua))
}

// hashStorage hash storage fingerprint
func (fe *FeatureExtractor) hashStorage(storage StorageData) float64 {
	return float64(storage.LocalStorageSize+storage.SessionStorageSize) * 0.001
}

// hashWebRTC hash WebRTC fingerprint
func (fe *FeatureExtractor) hashWebRTC(webrtc WebRTCData) float64 {
	if webrtc.IPLeaked {
		return 1.0
	}
	return 0.0
}

// hashHardware hash hardware fingerprint
func (fe *FeatureExtractor) hashHardware(hardware HardwareData) float64 {
	return float64(hardware.Cores) * 0.1
}

// calculateEntropy calculate configuration entropy
func (fe *FeatureExtractor) calculateEntropy(profile *profiles.ClientProfile) float64 {
	entropy := 0.0
	entropy += float64(len(profile.CipherSuites))
	entropy += float64(len(profile.Extensions)) * 0.5
	entropy += float64(len(profile.SupportedCurves)) * 0.3
	return entropy
}

// calculateStringEntropy calculate string entropy
func (fe *FeatureExtractor) calculateStringEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}
	freq := make(map[rune]int)
	for _, c := range s {
		freq[c]++
	}

	var entropy float64
	length := float64(len(s))
	for _, count := range freq {
		p := float64(count) / length
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}
	return entropy
}

// Normalize feature normalization
func (fe *FeatureExtractor) Normalize(fv *core.FeatureVector) *core.FeatureVector {
	normalized := core.NewFeatureVector()

	// Find maximum value
	maxVal := 0.0
	for _, v := range fv.Features {
		if v > maxVal {
			maxVal = v
		}
	}

	if maxVal == 0 {
		return normalized
	}

	// Normalize
	for ft, v := range fv.Features {
		normalized.Set(ft, v/maxVal)
	}

	return normalized
}

// FrontendFingerprintData frontend fingerprint data
type FrontendFingerprintData struct {
	Canvas   CanvasData
	WebGL    WebGLData
	Audio    AudioData
	Fonts    FontsData
	Storage  StorageData
	WebRTC   WebRTCData
	Hardware HardwareData
	Timing   TimingData
}

// CanvasData Canvas fingerprint data
type CanvasData struct {
	Entropy float64
	Hash    string
}

// WebGLData WebGL fingerprint data
type WebGLData struct {
	Entropy    float64
	Vendor     string
	Renderer   string
	Extensions []string
}

// AudioData Audio fingerprint data
type AudioData struct {
	Entropy    float64
	SampleRate float64
}

// FontsData font fingerprint data
type FontsData struct {
	List []string
}

// StorageData storage fingerprint data
type StorageData struct {
	LocalStorageSize   int
	SessionStorageSize int
}

// WebRTCData WebRTC fingerprint data
type WebRTCData struct {
	IPLeaked bool
	LocalIPs []string
}

// HardwareData hardware fingerprint data
type HardwareData struct {
	Cores       int
	Memory      int
	TouchPoints int
}

// TimingData timing fingerprint data
type TimingData struct {
	Precision float64
}

// ServerFingerprintData server-side fingerprint data
type ServerFingerprintData struct {
	ClientHello core.ClientHelloSpec
	Headers     *core.HTTPHeaders
}

// FeatureImportance feature importance analysis
type FeatureImportance struct {
	Feature    core.FeatureType
	Importance float64
}

// AnalyzeFeatureImportance analyze feature importance
func (fe *FeatureExtractor) AnalyzeFeatureImportance(trainingData []*core.FeatureVector, labels []string) []FeatureImportance {
	// Simplified feature importance analysis
	importance := make(map[core.FeatureType]float64)

	// Calculate variance for each feature
	for ft := range fe.weights {
		var values []float64
		for _, fv := range trainingData {
			if v, ok := fv.Features[ft]; ok {
				values = append(values, v)
			}
		}
		importance[ft] = fe.calculateVariance(values)
	}

	// Convert to slice and sort
	result := make([]FeatureImportance, 0, len(importance))
	for ft, imp := range importance {
		result = append(result, FeatureImportance{
			Feature:    ft,
			Importance: imp * fe.weights[ft],
		})
	}

	// Sort by importance
	sort.Slice(result, func(i, j int) bool {
		return result[i].Importance > result[j].Importance
	})

	return result
}

// calculateVariance calculate variance
func (fe *FeatureExtractor) calculateVariance(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// Calculate mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	// Calculate variance
	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}

	return variance / float64(len(values))
}

// SelectTopFeatures select most important features
func (fe *FeatureExtractor) SelectTopFeatures(importance []FeatureImportance, topK int) []core.FeatureType {
	if topK > len(importance) {
		topK = len(importance)
	}

	result := make([]core.FeatureType, 0, topK)
	for i := 0; i < topK; i++ {
		result = append(result, importance[i].Feature)
	}
	return result
}
