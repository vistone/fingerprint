package config

import (
	"testing"
)

// TestDeepCopy tests deep copy functionality
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

	// Validate deep copy effectiveness.
	if err := ValidateDeepCopy(original, copied); err != nil {
		t.Fatalf("ValidateDeepCopy failed: %v", err)
	}

	// Validate data integrity.
	if copied.BehaviorAnalysis.MinRequestsForAnalysis != 10 {
		t.Errorf("expected MinRequestsForAnalysis=10, got %d", copied.BehaviorAnalysis.MinRequestsForAnalysis)
	}

	// Validate that modifying the original config does not affect the copy.
	original.BehaviorAnalysis.MinRequestsForAnalysis = 20
	if copied.BehaviorAnalysis.MinRequestsForAnalysis != 10 {
		t.Error("deep copy failed: modifying original affected the copy")
	}
}

// TestDeepCopyNil tests deep copying nil config
func TestDeepCopyNil(t *testing.T) {
	copied, err := DeepCopy(nil)
	if err != nil {
		t.Fatalf("DeepCopy(nil) should not fail, got error: %v", err)
	}
	if copied != nil {
		t.Error("DeepCopy(nil) should return nil")
	}
}
