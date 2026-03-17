package ml

import "testing"

func TestShadowComparatorRecord(t *testing.T) {
	comparator := &shadowComparator{
		enabled:    true,
		sampleRate: 1.0,
		native:     &nativeInferenceBackend{pipeline: NewModelPipeline()},
		onnx:       &nativeInferenceBackend{pipeline: NewModelPipeline()},
	}

	primary := &PipelineResult{
		Browser: BrowserPrediction{Family: "chrome"},
		Forgery: ForgeryResult{ForgeryProb: 0.2},
		Threat:  ThreatPrediction{ThreatProb: 0.1, Action: ActAllow},
	}
	shadow := &PipelineResult{
		Browser: BrowserPrediction{Family: "chrome"},
		Forgery: ForgeryResult{ForgeryProb: 0.25},
		Threat:  ThreatPrediction{ThreatProb: 0.2, Action: ActAllow},
	}

	comparator.record("native", "onnx", primary, shadow, nil)
	stats := comparator.Snapshot()
	if stats == nil {
		t.Fatal("expected stats snapshot")
	}
	if stats.SampledCount != 1 {
		t.Fatalf("expected sampledCount=1, got %d", stats.SampledCount)
	}
	if stats.BrowserTop1AgreeRate != 1 {
		t.Fatalf("expected browser agree rate 1, got %f", stats.BrowserTop1AgreeRate)
	}
	if stats.ActionTop1AgreeRate != 1 {
		t.Fatalf("expected action agree rate 1, got %f", stats.ActionTop1AgreeRate)
	}
}
