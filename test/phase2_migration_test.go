package fingerprint_test

import (
	"testing"

	fp "github.com/vistone/fingerprint"
)

// TestPhase2MigrationBasicFunctionality 验证 Phase 2 迁移模块的基本功能
// 确保所有 10 个迁移的模块支持根包转发函数正常工作
func TestPhase2MigrationBasicFunctionality(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "JA3 MatchJA3",
			test: func(t *testing.T) {
				result := fp.MatchJA3("hash1", "hash2")
				// 验证函数存在且可以调用
				_ = result
			},
		},
		{
			name: "JA4 ComputeJA4ByProfileName",
			test: func(t *testing.T) {
				result, err := fp.ComputeJA4ByProfileName("chrome")
				// 验证函数可以调用
				if err != nil {
					t.Logf("ComputeJA4ByProfileName returned error: %v", err)
				}
				if result != nil {
					t.Logf("ComputeJA4ByProfileName returned result: %#v", result)
				}
			},
		},
		{
			name: "JA4S NewJA4SAnalyzer",
			test: func(t *testing.T) {
				analyzer := fp.NewJA4SAnalyzer()
				if analyzer == nil {
					t.Error("NewJA4SAnalyzer() returned nil")
				}
			},
		},
		{
			name: "ECH NewECHAnalyzer",
			test: func(t *testing.T) {
				analyzer := fp.NewECHAnalyzer()
				if analyzer == nil {
					t.Error("NewECHAnalyzer() returned nil")
				}
			},
		},
		{
			name: "JA4H NewJA4HAnalyzer",
			test: func(t *testing.T) {
				analyzer := fp.NewJA4HAnalyzer()
				if analyzer == nil {
					t.Error("NewJA4HAnalyzer() returned nil")
				}
			},
		},
		{
			name: "HTTP2 NewHTTP2SignatureAnalyzer",
			test: func(t *testing.T) {
				analyzer := fp.NewHTTP2SignatureAnalyzer()
				if analyzer == nil {
					t.Error("NewHTTP2SignatureAnalyzer() returned nil")
				}
			},
		},
		{
			name: "Policy NewPermissionsPolicyAnalyzer",
			test: func(t *testing.T) {
				analyzer := fp.NewPermissionsPolicyAnalyzer()
				if analyzer == nil {
					t.Error("NewPermissionsPolicyAnalyzer() returned nil")
				}

				// 验证方法可以调用
				result := analyzer.ParsePermissionsPolicy("camera=()")
				if result == nil {
					t.Error("ParsePermissionsPolicy() returned nil")
				}
			},
		},
		{
			name: "Behavior NewBehaviorAnalyzer",
			test: func(t *testing.T) {
				config := &fp.BehaviorAnalysisConfig{
					TimeWindowSize: 60,
				}
				analyzer := fp.NewBehaviorAnalyzer(config)
				if analyzer == nil {
					t.Error("NewBehaviorAnalyzer() returned nil")
				}

				// 验证方法可以调用
				signals := analyzer.GetAllSignals()
				if signals == nil {
					t.Error("GetAllSignals() returned nil")
				}
			},
		},
		{
			name: "QUIC NewQUICSignatureAnalyzer",
			test: func(t *testing.T) {
				analyzer := fp.NewQUICSignatureAnalyzer()
				if analyzer == nil {
					t.Error("NewQUICSignatureAnalyzer() returned nil")
				}
			},
		},
		{
			name: "TCP NewTCPIPAnalyzer",
			test: func(t *testing.T) {
				analyzer := fp.NewTCPIPAnalyzer()
				if analyzer == nil {
					t.Error("NewTCPIPAnalyzer() returned nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

// TestPhase2TypeAliasesWork 验证根包的类型别名正常工作
func TestPhase2TypeAliasesWork(t *testing.T) {
	t.Run("BehaviorAnalysisConfig type alias", func(t *testing.T) {
		config := &fp.BehaviorAnalysisConfig{
			TimeWindowSize: 60,
		}
		if config.TimeWindowSize != 60 {
			t.Error("BehaviorAnalysisConfig field access failed")
		}
	})

	t.Run("RequestBehavior type alias", func(t *testing.T) {
		req := &fp.RequestBehavior{
			TLSVersion: "1.3",
		}
		if req.TLSVersion != "1.3" {
			t.Error("RequestBehavior field access failed")
		}
	})

	t.Run("PermissionDirective type alias", func(t *testing.T) {
		dir := &fp.PermissionDirective{
			FeatureName: "camera",
		}
		if dir.FeatureName != "camera" {
			t.Error("PermissionDirective field access failed")
		}
	})
}
