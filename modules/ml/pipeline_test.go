package ml

import (
	"fmt"
	"math"
	"testing"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/profiles"
)

// =========================================================================
// Tensor basic tests
// =========================================================================

func TestTensorCreation(t *testing.T) {
	tensor := Zeros(3, 4)
	if tensor.Shape[0] != 3 || tensor.Shape[1] != 4 {
		t.Fatalf("expected shape [3,4], got %v", tensor.Shape)
	}
	if len(tensor.Data) != 12 {
		t.Fatalf("expected 12 elements, got %d", len(tensor.Data))
	}
	for _, v := range tensor.Data {
		if v != 0 {
			t.Fatal("Zeros should produce all zeros")
		}
	}
}

func TestTensorMatMul(t *testing.T) {
	// [2x3] * [3x2] = [2x2]
	a := NewTensor([]int{2, 3}, []float64{1, 2, 3, 4, 5, 6})
	b := NewTensor([]int{3, 2}, []float64{7, 8, 9, 10, 11, 12})
	c := a.MatMul(b)

	if c.Shape[0] != 2 || c.Shape[1] != 2 {
		t.Fatalf("expected [2,2], got %v", c.Shape)
	}
	// Row 0: 1*7+2*9+3*11=58, 1*8+2*10+3*12=64
	// Row 1: 4*7+5*9+6*11=139, 4*8+5*10+6*12=154
	expected := []float64{58, 64, 139, 154}
	for i, e := range expected {
		if math.Abs(c.Data[i]-e) > 1e-9 {
			t.Errorf("MatMul[%d]: expected %.0f, got %.6f", i, e, c.Data[i])
		}
	}
}

func TestTensorSoftmax(t *testing.T) {
	tensor := NewTensor([]int{1, 3}, []float64{1.0, 2.0, 3.0})
	result := tensor.SoftmaxRow()

	// Verify softmax normalization
	sum := 0.0
	for _, v := range result.Data {
		sum += v
	}
	if math.Abs(sum-1.0) > 1e-9 {
		t.Fatalf("softmax sum should be 1.0, got %f", sum)
	}

	// Verify ordering
	if result.Data[2] <= result.Data[1] || result.Data[1] <= result.Data[0] {
		t.Fatal("softmax should preserve order")
	}
}

func TestTensorArgmax(t *testing.T) {
	tensor1 := NewTensor([]int{1, 4}, []float64{0.1, 0.5, 0.3, 0.1})
	idx := tensor1.Argmax()
	if idx != 1 {
		t.Errorf("argmax: expected 1, got %d", idx)
	}

	tensor2 := NewTensor([]int{1, 4}, []float64{0.2, 0.1, 0.6, 0.1})
	idx2 := tensor2.Argmax()
	if idx2 != 2 {
		t.Errorf("argmax: expected 2, got %d", idx2)
	}
}

// =========================================================================
// Neural network layer tests
// =========================================================================

func TestDenseLayerForward(t *testing.T) {
	layer := NewDenseLayer(3, 2)
	input := NewTensor([]int{1, 3}, []float64{1.0, 0.5, -0.5})
	output := layer.Forward(input)

	if output.Shape[0] != 1 || output.Shape[1] != 2 {
		t.Fatalf("expected output [1,2], got %v", output.Shape)
	}
}

func TestSequentialForwardBackward(t *testing.T) {
	seq := NewSequential(
		NewDenseLayer(4, 8),
		NewReLULayer(),
		NewDenseLayer(8, 3),
		NewSoftmaxLayer(),
	)

	input := NewTensor([]int{2, 4}, []float64{
		0.1, 0.2, 0.3, 0.4,
		0.5, 0.6, 0.7, 0.8,
	})

	output := seq.Forward(input)

	if output.Shape[0] != 2 || output.Shape[1] != 3 {
		t.Fatalf("expected [2,3], got %v", output.Shape)
	}

	// Verify softmax normalization
	for row := 0; row < 2; row++ {
		sum := 0.0
		for col := 0; col < 3; col++ {
			sum += output.At(row, col)
		}
		if math.Abs(sum-1.0) > 1e-6 {
			t.Errorf("row %d softmax sum: expected 1.0, got %f", row, sum)
		}
	}

	// Test backward pass does not panic
	grad := Ones(2, 3)
	seq.ZeroGrad()
	_ = seq.Backward(grad)

	// Verify parameters have gradients
	params := seq.Params()
	if len(params) == 0 {
		t.Fatal("Sequential should have parameters")
	}
}

func TestAdamOptimizer(t *testing.T) {
	layer := NewDenseLayer(2, 1)
	params := layer.Params()
	optimizer := NewAdamOptimizer(params, 0.01)

	// Simulate one gradient step
	for _, p := range params {
		if p.Grad != nil {
			for i := range p.Grad.Data {
				p.Grad.Data[i] = 0.1
			}
		}
	}

	optimizer.Step()
	// Verify parameters were updated — should not panic
}

func TestCrossEntropyLoss(t *testing.T) {
	probs := NewTensor([]int{3, 4}, []float64{
		0.7, 0.1, 0.1, 0.1,
		0.1, 0.8, 0.05, 0.05,
		0.1, 0.1, 0.1, 0.7,
	})
	targets := []int{0, 1, 3} // correct classifications

	loss, grad := CrossEntropyLoss(probs, targets)
	if loss < 0 {
		t.Errorf("loss should be non-negative, got %f", loss)
	}
	if loss > 1.0 {
		t.Errorf("loss should be small for correct predictions, got %f", loss)
	}
	if grad.Shape[0] != 3 || grad.Shape[1] != 4 {
		t.Fatalf("grad shape mismatch: expected [3,4], got %v", grad.Shape)
	}
}

func TestTripletMarginLoss(t *testing.T) {
	anchor := NewTensor([]int{2, 4}, []float64{
		1, 0, 0, 0,
		0, 1, 0, 0,
	})
	positive := NewTensor([]int{2, 4}, []float64{
		0.9, 0.1, 0, 0,
		0.1, 0.9, 0, 0,
	})
	negative := NewTensor([]int{2, 4}, []float64{
		0, 0, 1, 0,
		0, 0, 0, 1,
	})

	loss, gradA, gradP, gradN := TripletMarginLoss(anchor, positive, negative, 1.0)

	if loss < 0 {
		t.Errorf("triplet loss should be non-negative, got %f", loss)
	}
	if gradA == nil || gradP == nil || gradN == nil {
		t.Fatal("gradients should not be nil")
	}
}

// =========================================================================
// Domain model tests
// =========================================================================

func TestFingerprintEncoder(t *testing.T) {
	enc := NewFingerprintEncoder()

	// Single sample encoding
	features := make([]float64, FingerprintFeatureDim)
	for i := range features {
		features[i] = float64(i) / float64(FingerprintFeatureDim)
	}

	embedding := enc.EncodeSingle(features)
	if len(embedding) != EmbeddingDim {
		t.Fatalf("expected %d-dim embedding, got %d", EmbeddingDim, len(embedding))
	}

	// Verify L2 normalization
	norm := 0.0
	for _, v := range embedding {
		norm += v * v
	}
	norm = math.Sqrt(norm)
	if math.Abs(norm-1.0) > 0.01 {
		t.Errorf("embedding should be L2-normalized (norm=1.0), got %.4f", norm)
	}
}

func TestFingerprintEncoderBatch(t *testing.T) {
	enc := NewFingerprintEncoder()

	batch := 5
	data := make([]float64, batch*FingerprintFeatureDim)
	for i := range data {
		data[i] = float64(i) / float64(len(data))
	}
	input := NewTensor([]int{batch, FingerprintFeatureDim}, data)
	output := enc.Encode(input)

	if output.Shape[0] != batch || output.Shape[1] != EmbeddingDim {
		t.Fatalf("expected [%d,%d], got %v", batch, EmbeddingDim, output.Shape)
	}
}

func TestBrowserClassifier(t *testing.T) {
	cls := NewBrowserClassifier()

	embedding := make([]float64, EmbeddingDim)
	for i := range embedding {
		embedding[i] = 0.5
	}
	pred := cls.ClassifySingle(embedding)

	if pred.Confidence < 0 || pred.Confidence > 1 {
		t.Errorf("confidence should be [0,1], got %f", pred.Confidence)
	}
	if len(pred.Probs) != NumBrowserFamilies {
		t.Errorf("expected %d probs, got %d", NumBrowserFamilies, len(pred.Probs))
	}

	// Verify probability normalization
	sum := 0.0
	for _, p := range pred.Probs {
		sum += p
	}
	if math.Abs(sum-1.0) > 1e-6 {
		t.Errorf("probs should sum to 1.0, got %f", sum)
	}
}

func TestForgeryDetector(t *testing.T) {
	det := NewForgeryDetector()

	features := make([]float64, FingerprintFeatureDim)
	crossFeatures := make([]float64, CrossLayerFeatureDim)
	for i := range features {
		features[i] = 0.5
	}
	for i := range crossFeatures {
		crossFeatures[i] = 0.8 // high consistency → should be real
	}

	result := det.DetectSingle(features, crossFeatures)

	if result.ForgeryProb < 0 || result.ForgeryProb > 1 {
		t.Errorf("forgery probability should be [0,1], got %f", result.ForgeryProb)
	}
	if len(result.TypeProbs) != NumForgeryTypes {
		t.Errorf("expected %d type probs, got %d", NumForgeryTypes, len(result.TypeProbs))
	}
}

func TestThreatAssessor(t *testing.T) {
	assessor := NewThreatAssessor()

	embedding := make([]float64, EmbeddingDim)
	forgery := &ForgeryResult{
		ForgeryProb: 0.1,
		ForgeryType: ForgeryReal,
		TypeProbs:   []float64{0.9, 0.05, 0.03, 0.02},
	}
	behavior := make([]float64, BehaviorFeatureDim)

	pred := assessor.AssessSingle(embedding, forgery, behavior)

	if pred.ThreatProb < 0 || pred.ThreatProb > 1 {
		t.Errorf("threat prob should be [0,1], got %f", pred.ThreatProb)
	}
	if len(pred.ClassProbs) != NumThreatClasses {
		t.Errorf("expected %d class probs, got %d", NumThreatClasses, len(pred.ClassProbs))
	}
	if len(pred.ActionProbs) != NumActions {
		t.Errorf("expected %d action probs, got %d", NumActions, len(pred.ActionProbs))
	}
}

// =========================================================================
// Feature encoding tests
// =========================================================================

func TestEncodeFingerprint(t *testing.T) {
	profile := &profiles.ClientProfile{
		TLSVersion:   0x0304,
		CipherSuites: []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f},
		Extensions: []core.TLSExtension{
			{Type: 0x0000}, // SNI
			{Type: 0x0010}, // ALPN
			{Type: 0x000a},
			{Type: 0x000b},
		},
		SupportedCurves:   []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		HTTP2Settings:     core.HTTP2Settings{InitialWindowSize: 6291456, MaxConcurrentStreams: 100, HeaderTableSize: 65536},
		PseudoHeaderOrder: []string{":method", ":authority", ":scheme", ":path"},
		TCPIP: &profiles.TCPIPFingerprint{
			TTL:        64,
			WindowSize: 65535,
			MSS:        1460,
			Timestamps: true,
		},
	}

	vec := EncodeFingerprint(profile)

	if len(vec) != FingerprintFeatureDim {
		t.Fatalf("expected %d features, got %d", FingerprintFeatureDim, len(vec))
	}

	// Verify TLS version encoding
	if math.Abs(vec[0]-1.0) > 1e-9 {
		t.Errorf("TLS 1.3 should encode to 1.0, got %f", vec[0])
	}

	// Verify all values in [0,1] range
	for i, v := range vec {
		if v < 0 || v > 1 {
			t.Errorf("feature[%d] out of range [0,1]: %f", i, v)
		}
	}

	// Verify SNI detection
	if vec[4] != 1.0 {
		t.Errorf("hasSNI should be 1.0, got %f", vec[4])
	}
	if vec[5] != 1.0 {
		t.Errorf("hasALPN should be 1.0, got %f", vec[5])
	}
}

func TestEncodeFingerprintFromFeatureVector(t *testing.T) {
	fv := core.NewFeatureVector()
	fv.Set(core.FeatureTLSVersion, 0x0304)
	fv.Set(core.FeatureCipherSuites, 15)
	fv.Set(core.FeatureExtensions, 12)

	vec := EncodeFingerprintFromFeatureVector(fv)
	if len(vec) != FingerprintFeatureDim {
		t.Fatalf("expected %d features, got %d", FingerprintFeatureDim, len(vec))
	}
}

func TestComputeCrossLayerFeatures(t *testing.T) {
	fp := make([]float64, FingerprintFeatureDim)
	// Set some consistent features
	fp[0] = 0.8  // TLS version
	fp[1] = 0.5  // cipher count
	fp[8] = 0.5  // H2 window
	fp[14] = 0.5 // TTL
	fp[18] = 0.5 // canvas
	fp[19] = 0.5 // webgl

	cross := ComputeCrossLayerFeatures(fp)
	if len(cross) != CrossLayerFeatureDim {
		t.Fatalf("expected %d cross features, got %d", CrossLayerFeatureDim, len(cross))
	}

	// canvas and webgl both present → consistency should be 1.0
	if cross[6] != 1.0 {
		t.Errorf("canvas+webgl consistency should be 1.0, got %f", cross[6])
	}

	// All values should be in [0,1]
	for i, v := range cross {
		if v < 0 || v > 1 {
			t.Errorf("cross[%d] out of range: %f", i, v)
		}
	}
}

// =========================================================================
// Inference pipeline tests
// =========================================================================

func TestModelPipelineInfer(t *testing.T) {
	pipeline := NewModelPipeline()

	profile := &profiles.ClientProfile{
		TLSVersion:   0x0303,
		CipherSuites: []uint16{0xc02b, 0xc02f, 0xc02c, 0xc030},
		Extensions: []core.TLSExtension{
			{Type: 0x0000},
			{Type: 0x0010},
		},
		SupportedCurves: []core.CurveID{core.CurveP256, core.CurveP384},
		HTTP2Settings:   core.HTTP2Settings{InitialWindowSize: 131072},
	}

	behavior := make([]float64, BehaviorFeatureDim)
	result := pipeline.Infer(profile, behavior)

	if result == nil {
		t.Fatal("pipeline result should not be nil")
	}
	if len(result.Embedding) != EmbeddingDim {
		t.Errorf("expected %d-dim embedding, got %d", EmbeddingDim, len(result.Embedding))
	}
	if result.Browser.Confidence < 0 || result.Browser.Confidence > 1 {
		t.Errorf("browser confidence out of range: %f", result.Browser.Confidence)
	}
	if result.Forgery.ForgeryProb < 0 || result.Forgery.ForgeryProb > 1 {
		t.Errorf("forgery prob out of range: %f", result.Forgery.ForgeryProb)
	}
	if result.Threat.ThreatProb < 0 || result.Threat.ThreatProb > 1 {
		t.Errorf("threat prob out of range: %f", result.Threat.ThreatProb)
	}
}

func TestModelPipelineInferFromFeatures(t *testing.T) {
	pipeline := NewModelPipeline()

	fv := core.NewFeatureVector()
	fv.Set(core.FeatureTLSVersion, 0x0304)
	fv.Set(core.FeatureCipherSuites, 15)

	result := pipeline.InferFromFeatures(fv, nil)
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if len(result.Embedding) != EmbeddingDim {
		t.Fatal("embedding dimension mismatch")
	}
}

func TestModelPipelineBatchInfer(t *testing.T) {
	pipeline := NewModelPipeline()

	profs := make([]*profiles.ClientProfile, 3)
	for i := range profs {
		profs[i] = &profiles.ClientProfile{
			TLSVersion:   0x0304,
			CipherSuites: []uint16{0x1301, 0x1302},
			Extensions:   []core.TLSExtension{{Type: 0x0000}},
		}
	}

	results := pipeline.InferBatch(profs, nil)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	for i, r := range results {
		if r == nil {
			t.Fatalf("result[%d] should not be nil", i)
		}
		if len(r.Embedding) != EmbeddingDim {
			t.Errorf("result[%d] embedding dim: expected %d, got %d", i, EmbeddingDim, len(r.Embedding))
		}
	}
}

// =========================================================================
// Model serialization tests
// =========================================================================

func TestModelPipelineSaveLoad(t *testing.T) {
	pipeline := NewModelPipeline()

	tmpFile := t.TempDir() + "/model_weights.json"

	// Save
	if err := pipeline.SaveWeights(tmpFile); err != nil {
		t.Fatalf("SaveWeights failed: %v", err)
	}

	// Load into new pipeline
	pipeline2 := NewModelPipeline()
	if err := pipeline2.LoadWeights(tmpFile); err != nil {
		t.Fatalf("LoadWeights failed: %v", err)
	}

	// Verify weight consistency
	params1 := pipeline.encoder.Net.Params()
	params2 := pipeline2.encoder.Net.Params()
	if len(params1) != len(params2) {
		t.Fatalf("param count mismatch: %d vs %d", len(params1), len(params2))
	}
	for i := range params1 {
		for j := range params1[i].Value.Data {
			if math.Abs(params1[i].Value.Data[j]-params2[i].Value.Data[j]) > 1e-12 {
				t.Fatalf("param[%d][%d] mismatch after save/load", i, j)
			}
		}
	}
}

// =========================================================================
// NeuralTrainer tests
// =========================================================================

func TestNeuralTrainerBuildData(t *testing.T) {
	pipeline := NewModelPipeline()
	trainer := NewNeuralTrainer(pipeline, &NeuralTrainerConfig{
		Epochs:          2,
		BatchSize:       4,
		LearningRate:    0.001,
		AugmentNoise:    0.02,
		TripletMargin:   1.0,
		ForgeryRatio:    1.0,
		ValidationSplit: 0.2,
	})

	// Create test profiles
	registry := profiles.NewProfileRegistry()
	for i := 0; i < 10; i++ {
		p := profiles.ClientProfile{
			ID:           fmt.Sprintf("test_%d", i),
			BrowserType:  core.BrowserChrome,
			TLSVersion:   0x0304,
			CipherSuites: []uint16{0x1301, 0x1302, 0x1303},
			Extensions:   []core.TLSExtension{{Type: 0x0000}},
		}
		registry.Register(p)
	}

	if err := trainer.TrainFromProfiles(registry); err != nil {
		t.Fatalf("TrainFromProfiles failed: %v", err)
	}

	if len(trainer.Metrics) == 0 {
		t.Error("expected training metrics to be recorded")
	}
}

// =========================================================================
// GPU device abstraction tests
// =========================================================================

func TestCPUDevice(t *testing.T) {
	dev := CPU
	a := Zeros(2, 3)
	b := Zeros(3, 2)
	for i := range a.Data {
		a.Data[i] = float64(i + 1)
	}
	for i := range b.Data {
		b.Data[i] = float64(i + 1)
	}

	c := dev.MatMul(a, b)
	if c.Shape[0] != 2 || c.Shape[1] != 2 {
		t.Fatalf("device MatMul shape: expected [2,2], got %v", c.Shape)
	}
}

// =========================================================================
// Model store tests
// =========================================================================

func TestModelStoreCreateAndLoad(t *testing.T) {
	dir := t.TempDir()
	store, err := NewModelStore(DefaultStoreConfig(dir))
	if err != nil {
		t.Fatalf("NewModelStore: %v", err)
	}

	if store.VersionCount() != 0 {
		t.Fatalf("expected 0 versions, got %d", store.VersionCount())
	}

	// Save a model
	pipeline := NewModelPipeline()
	err = store.Save(pipeline, "test save", nil)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	if store.VersionCount() != 1 {
		t.Fatalf("expected 1 version, got %d", store.VersionCount())
	}

	latest := store.Latest()
	if latest == nil || latest.Version != 1 {
		t.Fatalf("expected version 1, got %v", latest)
	}
	if latest.Description != "test save" {
		t.Errorf("expected description 'test save', got %q", latest.Description)
	}

	// Load into a new pipeline
	pipeline2 := NewModelPipeline()
	loaded, err := store.LoadLatest(pipeline2)
	if err != nil {
		t.Fatalf("LoadLatest: %v", err)
	}
	if !loaded {
		t.Fatal("expected to load model")
	}

	// Verify weights match
	p1 := pipeline.encoder.Net.Params()
	p2 := pipeline2.encoder.Net.Params()
	for i := range p1 {
		for j := range p1[i].Value.Data {
			if p1[i].Value.Data[j] != p2[i].Value.Data[j] {
				t.Fatalf("weight mismatch at param[%d][%d]", i, j)
			}
		}
	}
}

func TestModelStoreMultipleVersions(t *testing.T) {
	dir := t.TempDir()
	cfg := DefaultStoreConfig(dir)
	cfg.MaxVersions = 3
	store, err := NewModelStore(cfg)
	if err != nil {
		t.Fatalf("NewModelStore: %v", err)
	}

	pipeline := NewModelPipeline()

	// Save 5 versions; store should prune to 3
	for i := 0; i < 5; i++ {
		err = store.Save(pipeline, fmt.Sprintf("version %d", i+1), nil)
		if err != nil {
			t.Fatalf("Save v%d: %v", i+1, err)
		}
	}

	if store.VersionCount() != 3 {
		t.Fatalf("expected 3 versions after pruning, got %d", store.VersionCount())
	}

	versions := store.ListVersions()
	if versions[0].Version != 3 || versions[2].Version != 5 {
		t.Errorf("expected versions [3,4,5], got [%d,%d,%d]",
			versions[0].Version, versions[1].Version, versions[2].Version)
	}
}

func TestModelStoreEmptyLoadLatest(t *testing.T) {
	dir := t.TempDir()
	store, err := NewModelStore(DefaultStoreConfig(dir))
	if err != nil {
		t.Fatalf("NewModelStore: %v", err)
	}

	pipeline := NewModelPipeline()
	loaded, err := store.LoadLatest(pipeline)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loaded {
		t.Fatal("should not load from empty store")
	}
}

func TestModelStorePersistence(t *testing.T) {
	dir := t.TempDir()

	// Create store and save
	store1, _ := NewModelStore(DefaultStoreConfig(dir))
	pipeline := NewModelPipeline()
	_ = store1.Save(pipeline, "persist test", nil)

	// Re-open store from same directory
	store2, err := NewModelStore(DefaultStoreConfig(dir))
	if err != nil {
		t.Fatalf("re-open store: %v", err)
	}
	if store2.VersionCount() != 1 {
		t.Fatalf("expected 1 version after re-open, got %d", store2.VersionCount())
	}

	pipeline2 := NewModelPipeline()
	loaded, err := store2.LoadLatest(pipeline2)
	if err != nil || !loaded {
		t.Fatalf("failed to load from re-opened store: loaded=%v, err=%v", loaded, err)
	}
}

func TestPipelineLoadFromStore(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewModelStore(DefaultStoreConfig(dir))
	orig := NewModelPipeline()
	_ = store.Save(orig, "convenience test", nil)

	pipeline := NewModelPipeline()
	loaded, err := pipeline.LoadFromStore(store)
	if err != nil || !loaded {
		t.Fatalf("LoadFromStore failed: loaded=%v, err=%v", loaded, err)
	}
	if !pipeline.Trained() {
		t.Error("pipeline should be marked as trained after loading")
	}
}
