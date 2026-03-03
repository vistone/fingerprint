package config

import (
	"reflect"
	"testing"
	"time"
)

// TestManagedConfigClone 测试 ManagedConfig 深拷贝
func TestManagedConfigClone(t *testing.T) {
	original := &ManagedConfig{
		BehaviorAnalysis: &BehaviorAnalysisConfig{
			MinRequestsForAnalysis:         10,
			RegularityThreshold:            0.5,
			EntropyThreshold:               0.3,
			AnomalousIntervalRateThreshold: 0.2,
			RequestHistoryCapacity:         1000,
			SignalCapacity:                 100,
		},
		RiskScoring: &RiskScoringConfig{
			CriticalThreshold: 0.9,
			HighThreshold:     0.7,
			MediumThreshold:   0.5,
			LowThreshold:      0.3,
			MinConfidence:     0.8,
			Weights: &RiskWeights{
				Headless: 0.5,
				Anomaly:  0.3,
			},
		},
		Features: &FeatureExtractionConfig{
			EntropyHighThreshold:  7.0,
			EntropyLowThreshold:   10,
			ToolMarkers:           []string{"curl", "wget"},
			HeadlessMarkers:       []string{"headless", "phantom"},
			MobileScreenWidthMax:  768,
			DesktopScreenWidthMin: 1024,
		},
		QUIC: &QUICConfig{
			MinInitialMaxData:      1000,
			MinStreamData:          500,
			SupportedVersions:      []uint32{1, 2, 3},
			TransportParamCapacity: 100,
		},
		TLS: &TLSConfig{
			WeakCipherSuites:     []uint16{0x002f, 0x0035},
			SupportedVersions:    []uint16{0x0303, 0x0304},
			GREASEExtensions:     []uint16{0x0a0a, 0x1a1a},
			AnomalyFlagsCapacity: 50,
		},
		Global: &GlobalConfig{
			MaxConcurrency: 100,
			RequestTimeout: 5000,
			CacheSize:      10000,
			DebugMode:      true,
			MaxInputSize:   1048576,
		},
		Metadata: &ConfigMetadata{
			Version:      "1.0.0",
			LastModified: time.Now(),
			Author:       "test",
			Description:  "Test config",
		},
	}

	clone := original.Clone()

	// 验证值相等
	if !reflect.DeepEqual(original, clone) {
		t.Error("Clone should be equal to original")
	}

	// 验证深拷贝（修改 clone 不影响 original）
	clone.BehaviorAnalysis.MinRequestsForAnalysis = 999
	if original.BehaviorAnalysis.MinRequestsForAnalysis == 999 {
		t.Error("Clone is not deep copy - BehaviorAnalysis modified")
	}

	clone.RiskScoring.Weights.Headless = 999.0
	if original.RiskScoring.Weights.Headless == 999.0 {
		t.Error("Clone is not deep copy - RiskWeights modified")
	}

	clone.Features.ToolMarkers[0] = "modified"
	if original.Features.ToolMarkers[0] == "modified" {
		t.Error("Clone is not deep copy - ToolMarkers slice modified")
	}

	clone.QUIC.SupportedVersions[0] = 999
	if original.QUIC.SupportedVersions[0] == 999 {
		t.Error("Clone is not deep copy - QUIC SupportedVersions modified")
	}

	clone.TLS.WeakCipherSuites[0] = 0x9999
	if original.TLS.WeakCipherSuites[0] == 0x9999 {
		t.Error("Clone is not deep copy - TLS WeakCipherSuites modified")
	}
}

// TestNilConfigClone 测试 nil 配置的克隆
func TestNilConfigClone(t *testing.T) {
	var nilConfig *ManagedConfig
	if nilConfig.Clone() != nil {
		t.Error("Nil config clone should return nil")
	}

	nilBehavior := &ManagedConfig{BehaviorAnalysis: nil}
	clone := nilBehavior.Clone()
	if clone.BehaviorAnalysis != nil {
		t.Error("Nil BehaviorAnalysis clone should return nil")
	}
}

// TestFeatureExtractionConfigClone 测试 FeatureExtractionConfig 克隆
func TestFeatureExtractionConfigClone(t *testing.T) {
	original := &FeatureExtractionConfig{
		EntropyHighThreshold:  7.0,
		ToolMarkers:           []string{"curl", "wget", "python-requests"},
		HeadlessMarkers:       []string{"puppeteer", "selenium"},
		MobileScreenWidthMax:  768,
		DesktopScreenWidthMin: 1024,
	}

	clone := original.Clone()

	// 验证切片独立
	clone.ToolMarkers = append(clone.ToolMarkers, "new-marker")
	if len(original.ToolMarkers) != 3 {
		t.Error("ToolMarkers should not be affected by clone modification")
	}

	// 修改 clone 的 slice 元素
	clone.ToolMarkers[0] = "modified"
	if original.ToolMarkers[0] == "modified" {
		t.Error("ToolMarkers elements should be independent")
	}
}

// TestQUICConfigClone 测试 QUICConfig 克隆
func TestQUICConfigClone(t *testing.T) {
	original := &QUICConfig{
		MinInitialMaxData:      10000,
		MinStreamData:          5000,
		SupportedVersions:      []uint32{0x00000001, 0x00000002},
		TransportParamCapacity: 200,
	}

	clone := original.Clone()

	// 验证切片独立
	clone.SupportedVersions[0] = 0x99999999
	if original.SupportedVersions[0] == 0x99999999 {
		t.Error("SupportedVersions should be independent")
	}
}

// TestTLSConfigClone 测试 TLSConfig 克隆
func TestTLSConfigClone(t *testing.T) {
	original := &TLSConfig{
		WeakCipherSuites:     []uint16{0x002f, 0x0035, 0x003c},
		SupportedVersions:    []uint16{0x0301, 0x0302, 0x0303},
		GREASEExtensions:     []uint16{0x0a0a, 0x1a1a, 0x2a2a},
		AnomalyFlagsCapacity: 100,
	}

	clone := original.Clone()

	// 验证所有切片独立
	clone.WeakCipherSuites[0] = 0x9999
	clone.SupportedVersions[0] = 0x9999
	clone.GREASEExtensions[0] = 0x9999

	if original.WeakCipherSuites[0] == 0x9999 {
		t.Error("WeakCipherSuites should be independent")
	}
	if original.SupportedVersions[0] == 0x9999 {
		t.Error("SupportedVersions should be independent")
	}
	if original.GREASEExtensions[0] == 0x9999 {
		t.Error("GREASEExtensions should be independent")
	}
}

// BenchmarkManagedConfigClone 基准测试：ManagedConfig 深拷贝
func BenchmarkManagedConfigClone(b *testing.B) {
	config := &ManagedConfig{
		BehaviorAnalysis: &BehaviorAnalysisConfig{
			MinRequestsForAnalysis: 10,
			RegularityThreshold:    0.5,
		},
		RiskScoring: &RiskScoringConfig{
			CriticalThreshold: 0.9,
			Weights: &RiskWeights{
				Headless: 0.5,
				Anomaly:  0.3,
			},
		},
		Features: &FeatureExtractionConfig{
			ToolMarkers:     []string{"curl", "wget", "python-requests"},
			HeadlessMarkers: []string{"puppeteer", "selenium"},
		},
		QUIC: &QUICConfig{
			SupportedVersions: []uint32{1, 2, 3},
		},
		TLS: &TLSConfig{
			WeakCipherSuites:  []uint16{0x002f, 0x0035},
			SupportedVersions: []uint16{0x0303, 0x0304},
		},
		Global: &GlobalConfig{
			MaxConcurrency: 100,
			CacheSize:      10000,
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = config.Clone()
	}
}

// BenchmarkFeatureExtractionConfigClone 基准测试：FeatureExtractionConfig 深拷贝
func BenchmarkFeatureExtractionConfigClone(b *testing.B) {
	config := &FeatureExtractionConfig{
		ToolMarkers:     []string{"curl", "wget", "python-requests", "node-fetch"},
		HeadlessMarkers: []string{"puppeteer", "selenium", "playwright"},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = config.Clone()
	}
}
