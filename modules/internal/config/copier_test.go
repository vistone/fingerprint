package config

import (
	"testing"
)

// TestDeepCopy 测试深复制功能
func TestDeepCopy(t *testing.T) {
	original := &ManagedConfig{
		BehaviorAnalysis: &BehaviorAnalysisConfig{
			MinRequestsForAnalysis: 10,
			RegularityThreshold:    0.8,
			RequestHistoryCapacity: 1000,
			SignalCapacity:         500,
		},
		Global: &GlobalConfig{
			MaxConcurrency: 100,
			RequestTimeout: 5000,
		},
	}

	copied, err := DeepCopy(original)
	if err != nil {
		t.Fatalf("DeepCopy failed: %v", err)
	}

	// 验证深复制的有效性
	if err := ValidateDeepCopy(original, copied); err != nil {
		t.Fatalf("ValidateDeepCopy failed: %v", err)
	}

	// 验证数据完整性
	if copied.BehaviorAnalysis.MinRequestsForAnalysis != 10 {
		t.Errorf("expected MinRequestsForAnalysis=10, got %d", copied.BehaviorAnalysis.MinRequestsForAnalysis)
	}

	// 验证修改原始配置不会影响副本
	original.BehaviorAnalysis.MinRequestsForAnalysis = 20
	if copied.BehaviorAnalysis.MinRequestsForAnalysis != 10 {
		t.Error("deep copy failed: modifying original affected the copy")
	}
}

// TestDeepCopyNil 测试深复制 nil 配置
func TestDeepCopyNil(t *testing.T) {
	copied, err := DeepCopy(nil)
	if err != nil {
		t.Fatalf("DeepCopy(nil) should not fail, got error: %v", err)
	}
	if copied != nil {
		t.Error("DeepCopy(nil) should return nil")
	}
}
