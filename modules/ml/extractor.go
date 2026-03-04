// Package ml 提供特征提取功能
package ml

import (
	"math"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/profiles"
)

// FeatureExtractor 特征提取器
type FeatureExtractor struct {
	// 特征权重
	weights map[core.FeatureType]float64
}

// NewFeatureExtractor 创建新的特征提取器
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

// ExtractFromProfile 从配置提取特征向量
func (fe *FeatureExtractor) ExtractFromProfile(profile *profiles.ClientProfile) *core.FeatureVector {
	fv := core.NewFeatureVector()

	// TLS 特征
	fv.Set(core.FeatureTLSVersion, float64(profile.TLSVersion))
	fv.Set(core.FeatureCipherSuites, float64(len(profile.CipherSuites)))
	fv.Set(core.FeatureExtensions, float64(len(profile.Extensions)))

	// HTTP/2 特征
	fv.Set(core.FeatureHTTP2Settings, fe.hashHTTP2Settings(profile.HTTP2Settings))
	fv.Set(core.FeatureEntropy, fe.calculateEntropy(profile))

	// HTTP 头特征
	if profile.Headers != nil {
		fv.Set(core.FeatureHTTPHeaders, fe.hashHeaders(profile.Headers))
		fv.Set(core.FeatureUserAgent, fe.hashUserAgent(profile.Headers.UserAgent))
	}

	return fv
}

// ExtractFromClientHello 从 ClientHello 提取特征
func (fe *FeatureExtractor) ExtractFromClientHello(spec core.ClientHelloSpec) *core.FeatureVector {
	fv := core.NewFeatureVector()

	fv.Set(core.FeatureTLSVersion, float64(spec.TLSVersion))
	fv.Set(core.FeatureCipherSuites, float64(len(spec.CipherSuites)))
	fv.Set(core.FeatureExtensions, float64(len(spec.Extensions)))

	return fv
}

// ExtractFromHTTPHeaders 从 HTTP 头提取特征
func (fe *FeatureExtractor) ExtractFromHTTPHeaders(headers *core.HTTPHeaders) *core.FeatureVector {
	fv := core.NewFeatureVector()

	if headers == nil {
		return fv
	}

	fv.Set(core.FeatureHTTPHeaders, fe.hashHeaders(headers))
	fv.Set(core.FeatureUserAgent, fe.hashUserAgent(headers.UserAgent))

	// 计算头信息熵
	entropy := 0.0
	headerMap := headers.ToMap()
	for _, value := range headerMap {
		entropy += fe.calculateStringEntropy(value)
	}
	fv.Set(core.FeatureEntropy, entropy)

	return fv
}

// ExtractFromFrontend 从前端指纹提取特征
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

// ExtractCombined 提取综合特征
func (fe *FeatureExtractor) ExtractCombined(serverData ServerFingerprintData, frontendData FrontendFingerprintData) *core.FeatureVector {
	fv := core.NewFeatureVector()

	// 服务端特征
	tlsFV := fe.ExtractFromClientHello(serverData.ClientHello)
	httpFV := fe.ExtractFromHTTPHeaders(serverData.Headers)

	// 前端特征
	frontendFV := fe.ExtractFromFrontend(frontendData)

	// 合并特征
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

// hashHTTP2Settings 哈希 HTTP/2 Settings
func (fe *FeatureExtractor) hashHTTP2Settings(settings core.HTTP2Settings) float64 {
	// 简化的哈希计算
	hash := float64(settings.HeaderTableSize) * 0.1
	hash += float64(settings.MaxConcurrentStreams) * 0.01
	hash += float64(settings.InitialWindowSize) * 0.001
	return hash
}

// hashHeaders 哈希 HTTP 头
func (fe *FeatureExtractor) hashHeaders(headers *core.HTTPHeaders) float64 {
	headerMap := headers.ToMap()
	var sum float64
	for name, value := range headerMap {
		sum += float64(len(name)) * float64(len(value))
	}
	return math.Log1p(sum)
}

// hashUserAgent 哈希 User-Agent
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

// hashStorage 哈希存储指纹
func (fe *FeatureExtractor) hashStorage(storage StorageData) float64 {
	return float64(storage.LocalStorageSize+storage.SessionStorageSize) * 0.001
}

// hashWebRTC 哈希 WebRTC 指纹
func (fe *FeatureExtractor) hashWebRTC(webrtc WebRTCData) float64 {
	if webrtc.IPLeaked {
		return 1.0
	}
	return 0.0
}

// hashHardware 哈希硬件指纹
func (fe *FeatureExtractor) hashHardware(hardware HardwareData) float64 {
	return float64(hardware.Cores) * 0.1
}

// calculateEntropy 计算配置熵
func (fe *FeatureExtractor) calculateEntropy(profile *profiles.ClientProfile) float64 {
	entropy := 0.0
	entropy += float64(len(profile.CipherSuites))
	entropy += float64(len(profile.Extensions)) * 0.5
	entropy += float64(len(profile.SupportedCurves)) * 0.3
	return entropy
}

// calculateStringEntropy 计算字符串熵
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

// Normalize 特征归一化
func (fe *FeatureExtractor) Normalize(fv *core.FeatureVector) *core.FeatureVector {
	normalized := core.NewFeatureVector()

	// 找到最大值
	maxVal := 0.0
	for _, v := range fv.Features {
		if v > maxVal {
			maxVal = v
		}
	}

	if maxVal == 0 {
		return normalized
	}

	// 归一化
	for ft, v := range fv.Features {
		normalized.Set(ft, v/maxVal)
	}

	return normalized
}

// FrontendFingerprintData 前端指纹数据
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

// CanvasData Canvas 指纹数据
type CanvasData struct {
	Entropy float64
	Hash    string
}

// WebGLData WebGL 指纹数据
type WebGLData struct {
	Entropy    float64
	Vendor     string
	Renderer   string
	Extensions []string
}

// AudioData Audio 指纹数据
type AudioData struct {
	Entropy   float64
	SampleRate float64
}

// FontsData 字体指纹数据
type FontsData struct {
	List []string
}

// StorageData 存储指纹数据
type StorageData struct {
	LocalStorageSize  int
	SessionStorageSize int
}

// WebRTCData WebRTC 指纹数据
type WebRTCData struct {
	IPLeaked bool
	LocalIPs []string
}

// HardwareData 硬件指纹数据
type HardwareData struct {
	Cores       int
	Memory      int
	TouchPoints int
}

// TimingData 时间指纹数据
type TimingData struct {
	Precision float64
}

// ServerFingerprintData 服务端指纹数据
type ServerFingerprintData struct {
	ClientHello core.ClientHelloSpec
	Headers     *core.HTTPHeaders
}

// FeatureImportance 特征重要性分析
type FeatureImportance struct {
	Feature    core.FeatureType
	Importance float64
}

// AnalyzeFeatureImportance 分析特征重要性
func (fe *FeatureExtractor) AnalyzeFeatureImportance(trainingData []*core.FeatureVector, labels []string) []FeatureImportance {
	// 简化的特征重要性分析
	importance := make(map[core.FeatureType]float64)

	// 统计每个特征的变化程度
	for ft := range fe.weights {
		var values []float64
		for _, fv := range trainingData {
			if v, ok := fv.Features[ft]; ok {
				values = append(values, v)
			}
		}
		importance[ft] = fe.calculateVariance(values)
	}

	// 转换为切片并排序
	var result []FeatureImportance
	for ft, imp := range importance {
		result = append(result, FeatureImportance{
			Feature:    ft,
			Importance: imp * fe.weights[ft],
		})
	}

	// 按重要性排序
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Importance < result[j].Importance {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

// calculateVariance 计算方差
func (fe *FeatureExtractor) calculateVariance(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// 计算平均值
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	// 计算方差
	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}

	return variance / float64(len(values))
}

// SelectTopFeatures 选择最重要的特征
func (fe *FeatureExtractor) SelectTopFeatures(importance []FeatureImportance, topK int) []core.FeatureType {
	if topK > len(importance) {
		topK = len(importance)
	}

	var result []core.FeatureType
	for i := 0; i < topK; i++ {
		result = append(result, importance[i].Feature)
	}
	return result
}
