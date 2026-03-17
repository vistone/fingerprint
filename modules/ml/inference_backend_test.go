package ml

import "testing"

func TestBuildResultFromONNXOutputs(t *testing.T) {
	features := make([]float64, FingerprintFeatureDim)
	cross := make([]float64, CrossLayerFeatureDim)
	embedding := make([]float64, EmbeddingDim)
	familyLogits := []float64{0.1, 1.2, -0.4, 0.3, 0.2, 0.0, -0.1}
	typeLogits := []float64{-1.0, 2.0, 0.1, -0.2}
	threatLogits := []float64{0.2, 0.7, -0.1, 0.3, 0.0, -0.4}
	actionLogits := []float64{0.1, -0.2, 0.3, 0.9, -0.4}

	result, err := buildResultFromONNXOutputs(onnxOutputBundle{
		Features:      features,
		CrossFeatures: cross,
		Embedding:     embedding,
		FamilyLogits:  familyLogits,
		ForgeryProb:   0.8,
		TypeLogits:    typeLogits,
		ThreatLogits:  threatLogits,
		ActionLogits:  actionLogits,
	})
	if err != nil {
		t.Fatalf("buildResultFromONNXOutputs returned error: %v", err)
	}

	if !result.Forgery.IsForgery {
		t.Fatal("expected forged result when forgery_prob > 0.5")
	}
	if len(result.Browser.Probs) != NumBrowserFamilies {
		t.Fatalf("expected %d browser probabilities, got %d", NumBrowserFamilies, len(result.Browser.Probs))
	}
	if len(result.Threat.ActionProbs) != NumActions {
		t.Fatalf("expected %d action probabilities, got %d", NumActions, len(result.Threat.ActionProbs))
	}
}
