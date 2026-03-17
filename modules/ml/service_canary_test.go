package ml

import "testing"

func TestCanaryRouterSelect(t *testing.T) {
	primary := &nativeInferenceBackend{pipeline: NewModelPipeline()}
	canary := &canaryRouter{
		enabled: true,
		rate:    1.0,
		backend: primary,
	}

	selected := canary.pick(primary)
	if selected == nil {
		t.Fatal("expected selected backend")
	}
	if canary.Snapshot().CanaryRouted != 1 {
		t.Fatalf("expected canary routed count 1, got %d", canary.Snapshot().CanaryRouted)
	}
}
