package ml

import (
	"math"
	"testing"

	"github.com/vistone/fingerprint/modules/core"
)

func TestNewOnlineClassifier(t *testing.T) {
	oc := NewOnlineClassifier(5)
	if oc == nil {
		t.Fatal("NewOnlineClassifier returned nil")
	}
	if oc.ClassCount() != 0 {
		t.Errorf("expected 0 classes, got %d", oc.ClassCount())
	}
}

func TestPartialFitNewClass(t *testing.T) {
	oc := NewOnlineClassifier(3)

	oc.PartialFit([]float64{1, 0, 0}, "class_a")
	oc.PartialFit([]float64{0, 1, 0}, "class_b")

	if oc.ClassCount() != 2 {
		t.Errorf("expected 2 classes, got %d", oc.ClassCount())
	}
	if oc.SampleCount() != 2 {
		t.Errorf("expected 2 samples, got %d", oc.SampleCount())
	}
}

func TestPartialFitCentroidUpdate(t *testing.T) {
	oc := NewOnlineClassifier(2)

	// First sample: centroid = [2, 0]
	oc.PartialFit([]float64{2, 0}, "a")
	// Second sample: centroid = [2, 0] + ([4, 0] - [2, 0])/2 = [3, 0]
	oc.PartialFit([]float64{4, 0}, "a")

	label, _ := oc.Predict([]float64{3, 0})
	if label != "a" {
		t.Errorf("expected class 'a', got %s", label)
	}
}

func TestOnlinePrediction(t *testing.T) {
	oc := NewOnlineClassifier(3)

	// Add two well-separated classes
	for i := 0; i < 10; i++ {
		oc.PartialFit([]float64{1, 0, 0}, "left")
		oc.PartialFit([]float64{0, 0, 1}, "right")
	}

	label, conf := oc.Predict([]float64{0.9, 0, 0.1})
	if label != "left" {
		t.Errorf("expected 'left', got %s", label)
	}
	if conf <= 0 {
		t.Error("confidence should be positive")
	}
}

func TestPartialFitBatch(t *testing.T) {
	oc := NewOnlineClassifier(2)

	features := [][]float64{{1, 0}, {0, 1}, {1, 0.1}}
	labels := []string{"a", "b", "a"}

	oc.PartialFitBatch(features, labels)

	if oc.ClassCount() != 2 {
		t.Errorf("expected 2 classes, got %d", oc.ClassCount())
	}
	if oc.SampleCount() != 3 {
		t.Errorf("expected 3 samples, got %d", oc.SampleCount())
	}
}

func TestWeightedPartialFit(t *testing.T) {
	oc := NewOnlineClassifier(2)

	oc.PartialFit([]float64{0, 0}, "a")
	// High-weight update should pull centroid more
	oc.WeightedPartialFit([]float64{10, 10}, "a", 1.0)

	// Centroid should be closer to [10,10] than [0,0]
	label, _ := oc.Predict([]float64{8, 8})
	if label != "a" {
		t.Errorf("expected 'a', got %s", label)
	}
}

func TestWeightedPartialFitZeroWeight(t *testing.T) {
	oc := NewOnlineClassifier(2)
	oc.PartialFit([]float64{1, 0}, "a")

	initialCount := oc.SampleCount()
	oc.WeightedPartialFit([]float64{0, 1}, "a", 0) // should be ignored
	if oc.SampleCount() != initialCount {
		t.Error("zero-weight update should not change sample count")
	}
}

func TestConceptDriftDetection(t *testing.T) {
	oc := NewOnlineClassifier(2)

	// Not enough data yet
	if oc.DriftDetected() {
		t.Error("should not detect drift with no data")
	}

	// Record many correct outcomes
	for i := 0; i < 60; i++ {
		oc.RecordOutcome(true)
	}
	if oc.DriftDetected() {
		t.Error("should not detect drift with high accuracy")
	}

	// Now record many incorrect outcomes to trigger drift
	for i := 0; i < 200; i++ {
		oc.RecordOutcome(false)
	}
	if !oc.DriftDetected() {
		t.Error("should detect drift after many incorrect predictions")
	}
}

func TestAccuracy(t *testing.T) {
	oc := NewOnlineClassifier(2)

	if oc.Accuracy() != 0 {
		t.Error("accuracy should be 0 with no data")
	}

	oc.RecordOutcome(true)
	oc.RecordOutcome(true)
	oc.RecordOutcome(false)

	acc := oc.Accuracy()
	expected := 2.0 / 3.0
	if math.Abs(acc-expected) > 0.01 {
		t.Errorf("accuracy = %f, want ~%f", acc, expected)
	}
}

func TestOnlineClassifierFromBase(t *testing.T) {
	base := NewSimpleClassifier(2)
	base.Train(
		[][]float64{{1, 0}, {0, 1}},
		[]string{"a", "b"},
	)

	oc := NewOnlineClassifierFrom(base, map[string]int{"a": 1, "b": 1})
	if oc.ClassCount() != 2 {
		t.Errorf("expected 2 classes, got %d", oc.ClassCount())
	}

	label, _ := oc.Predict([]float64{0.9, 0.1})
	if label != "a" {
		t.Errorf("expected 'a', got %s", label)
	}

	// Incremental update should work on top of existing centroids
	oc.PartialFit([]float64{0.5, 0.5}, "c")
	if oc.ClassCount() != 3 {
		t.Errorf("expected 3 classes after online update, got %d", oc.ClassCount())
	}
}

func TestOnlineHierarchicalClassifier(t *testing.T) {
	ohc := NewOnlineHierarchicalClassifier()

	// Train protocol layer
	for i := 0; i < 10; i++ {
		ohc.UpdateProtocol([]float64{1, 0, 0, 0, 0, 0, 0, 0, 0, 0}, "tls")
		ohc.UpdateProtocol([]float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 1}, "http")
	}

	// Train family layer for TLS
	for i := 0; i < 10; i++ {
		ohc.UpdateFamily("tls", make([]float64, 15), string(core.BrowserChrome))
	}

	// Train version layer
	for i := 0; i < 10; i++ {
		ohc.UpdateVersion(string(core.BrowserChrome), make([]float64, 20), "120.0")
	}

	stats := ohc.Stats()
	if stats.Protocol.Classes < 2 {
		t.Errorf("expected >= 2 protocol classes, got %d", stats.Protocol.Classes)
	}
	if len(stats.Family) == 0 {
		t.Error("expected at least one family layer")
	}
}

func TestOnlineHierarchicalClassify(t *testing.T) {
	ohc := NewOnlineHierarchicalClassifier()

	// Train protocol
	for i := 0; i < 10; i++ {
		ohc.UpdateProtocol([]float64{1, 0, 0, 0, 0, 0, 0, 0, 0, 0}, "tls")
		ohc.UpdateProtocol([]float64{0, 0, 0, 0, 0, 1, 0, 0, 0, 0}, "http")
	}

	// Train family for tls
	for i := 0; i < 10; i++ {
		tlsFamily := make([]float64, 15)
		tlsFamily[0] = 1.0
		ohc.UpdateFamily("tls", tlsFamily, string(core.BrowserChrome))
	}

	// Classify a TLS-like feature vector
	features := make([]float64, 20)
	features[0] = 1.0

	result := ohc.Classify(features)
	if result.Protocol != core.ProtocolType("tls") {
		t.Errorf("expected protocol tls, got %s", result.Protocol)
	}
	if result.Confidence <= 0 {
		t.Error("expected positive confidence")
	}
}

func TestOnlineHierarchicalStatsDetail(t *testing.T) {
	ohc := NewOnlineHierarchicalClassifier()

	ohc.UpdateProtocol([]float64{1, 0, 0, 0, 0, 0, 0, 0, 0, 0}, "tls")
	ohc.UpdateFamily("tls", make([]float64, 15), "chrome")
	ohc.UpdateVersion("chrome", make([]float64, 20), "120.0")

	stats := ohc.Stats()
	if stats.Protocol.Samples != 1 {
		t.Errorf("expected 1 protocol sample, got %d", stats.Protocol.Samples)
	}
	fs, ok := stats.Family["tls"]
	if !ok {
		t.Fatal("expected tls family stats")
	}
	if fs.Samples != 1 {
		t.Errorf("expected 1 family sample, got %d", fs.Samples)
	}
	vs, ok := stats.Version["chrome"]
	if !ok {
		t.Fatal("expected chrome version stats")
	}
	if vs.Samples != 1 {
		t.Errorf("expected 1 version sample, got %d", vs.Samples)
	}
}
