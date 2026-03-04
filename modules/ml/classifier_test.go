// Package ml 分类器测试
package ml

import (
	"math"
	"testing"

	"github.com/vistone/fingerprint/modules/core"
)

func TestSimpleClassifier(t *testing.T) {
	sc := NewSimpleClassifier(5)
	
	// 训练数据
	features := [][]float64{
		{1, 0, 0, 0, 0},
		{0.9, 0.1, 0, 0, 0},
		{0, 1, 0, 0, 0},
		{0, 0.9, 0.1, 0, 0},
	}
	labels := []string{"class_a", "class_a", "class_b", "class_b"}
	
	err := sc.Train(features, labels)
	if err != nil {
		t.Fatalf("Train failed: %v", err)
	}
	
	// 预测
	label, confidence := sc.Predict([]float64{0.95, 0, 0, 0, 0})
	if label != "class_a" {
		t.Errorf("Expected class_a, got %s", label)
	}
	if confidence <= 0 {
		t.Error("Confidence should be > 0")
	}
}

func TestSimpleClassifierPredictTopK(t *testing.T) {
	sc := NewSimpleClassifier(2)
	
	features := [][]float64{
		{1, 0},
		{0, 1},
	}
	labels := []string{"a", "b"}
	
	sc.Train(features, labels)
	
	predictions := sc.PredictTopK([]float64{0.8, 0.2}, 2)
	if len(predictions) != 2 {
		t.Errorf("Expected 2 predictions, got %d", len(predictions))
	}
	
	// 第一个预测应该是 "a"
	if predictions[0].Label != "a" {
		t.Errorf("Expected 'a' as top prediction, got %s", predictions[0].Label)
	}
	
	// 置信度应该按降序排列
	if len(predictions) == 2 && predictions[0].Confidence < predictions[1].Confidence {
		t.Error("Predictions should be sorted by confidence (descending)")
	}
}

func TestProtocolClassifier(t *testing.T) {
	pc := NewProtocolClassifier()
	
	features := [][]float64{
		{0x0303, 8, 10, 65536, 10, 15},
		{0x0304, 6, 8, 65536, 10, 20},
	}
	labels := []core.ProtocolType{core.ProtocolTLS, core.ProtocolHTTP3}
	
	err := pc.Train(features, labels)
	if err != nil {
		t.Fatalf("Train failed: %v", err)
	}
	
	// 预测
	protocol, conf := pc.Predict([]float64{0x0303, 8, 10, 65536, 10, 15})
	if protocol != core.ProtocolTLS {
		t.Errorf("Expected TLS, got %s", protocol)
	}
	if conf <= 0 {
		t.Error("Confidence should be > 0")
	}
}

func TestHierarchicalClassifierInitialize(t *testing.T) {
	hc := NewHierarchicalClassifier()
	hc.Initialize()
	
	// 初始化后应该创建了子分类器
	// 注意：由于字段是私有的，我们只能测试行为
	result := hc.Classify(core.NewFeatureVector())
	if result == nil {
		t.Error("Classify should return result")
	}
	
	if result.Labels["error"] != "classifier not trained" {
		t.Error("Should indicate not trained before training")
	}
}

func TestHierarchicalClassifierTrain(t *testing.T) {
	hc := NewHierarchicalClassifier()
	hc.Initialize()
	
	// 创建简单的训练数据
	trainingData := &TrainingData{
		ProtocolFeatures: [][]float64{
			{0x0303, 8, 10, 65536, 10, 15},
			{0x0304, 6, 8, 65536, 10, 20},
		},
		ProtocolLabels: []core.ProtocolType{core.ProtocolTLS, core.ProtocolHTTP3},
		FamilyFeatures: map[core.ProtocolType][][]float64{
			core.ProtocolTLS: {
				{100, 50, 8, 10, 65536},
				{200, 60, 8, 10, 65536},
			},
		},
		FamilyLabels: map[core.ProtocolType][]core.BrowserType{
			core.ProtocolTLS: {core.BrowserChrome, core.BrowserFirefox},
		},
		VersionFeatures: map[core.BrowserType][][]float64{
			core.BrowserChrome: {
				{0x0303, 8, 11, 65536, 100},
			},
		},
		VersionLabels: map[core.BrowserType][]string{
			core.BrowserChrome: {"133"},
		},
	}
	
	err := hc.Train(trainingData)
	if err != nil {
		t.Fatalf("Train failed: %v", err)
	}
	
	// 训练后应该可以进行分类
	fv := core.NewFeatureVector()
	fv.Set(core.FeatureTLSVersion, 0x0303)
	fv.Set(core.FeatureCipherSuites, 8)
	fv.Set(core.FeatureExtensions, 10)
	
	result := hc.Classify(fv)
	if result.Labels["error"] == "classifier not trained" {
		t.Error("Should not show 'not trained' after training")
	}
}

func TestClassificationResultIsHighConfidence(t *testing.T) {
	tests := []struct {
		name     string
		result   ClassificationResult
		expected bool
	}{
		{
			name: "high confidence",
			result: ClassificationResult{
				Confidence: 0.85,
				LayerScores: LayerScores{
					ProtocolConfidence: 0.75,
					FamilyConfidence:   0.85,
					VersionConfidence:  0.65,
				},
			},
			expected: true,
		},
		{
			name: "low confidence",
			result: ClassificationResult{
				Confidence: 0.50,
				LayerScores: LayerScores{
					ProtocolConfidence: 0.75,
					FamilyConfidence:   0.85,
					VersionConfidence:  0.65,
				},
			},
			expected: false,
		},
		{
			name: "low protocol confidence",
			result: ClassificationResult{
				Confidence: 0.85,
				LayerScores: LayerScores{
					ProtocolConfidence: 0.60,
					FamilyConfidence:   0.85,
					VersionConfidence:  0.65,
				},
			},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.IsHighConfidence()
			if got != tt.expected {
				t.Errorf("IsHighConfidence() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGenerateSyntheticDataset(t *testing.T) {
	dataset := GenerateSyntheticDataset("test", 50)
	
	if len(dataset.Samples) != 50 {
		t.Errorf("Expected 50 samples, got %d", len(dataset.Samples))
	}
	
	if dataset.Name != "test" {
		t.Errorf("Expected name 'test', got %s", dataset.Name)
	}
	
	// 检查统计信息
	if dataset.Statistics.TotalSamples != 50 {
		t.Errorf("Statistics.TotalSamples = %d, want 50", dataset.Statistics.TotalSamples)
	}
	
	// 验证有样本被分配到不同类别
	if len(dataset.Statistics.ProtocolCounts) == 0 {
		t.Error("Should have protocol counts")
	}
	if len(dataset.Statistics.FamilyCounts) == 0 {
		t.Error("Should have family counts")
	}
}

func TestDatasetToTrainingData(t *testing.T) {
	dataset := GenerateSyntheticDataset("test", 10)
	trainingData := dataset.ToTrainingData()
	
	if len(trainingData.ProtocolFeatures) != 10 {
		t.Errorf("Expected 10 protocol features, got %d", len(trainingData.ProtocolFeatures))
	}
	if len(trainingData.ProtocolLabels) != 10 {
		t.Errorf("Expected 10 protocol labels, got %d", len(trainingData.ProtocolLabels))
	}
}

func BenchmarkSimpleClassifierPredict(b *testing.B) {
	sc := NewSimpleClassifier(10)
	
	// 预训练
	features := make([][]float64, 100)
	labels := make([]string, 100)
	for i := 0; i < 100; i++ {
		features[i] = []float64{float64(i), float64(i + 1)}
		if i%2 == 0 {
			labels[i] = "even"
		} else {
			labels[i] = "odd"
		}
	}
	sc.Train(features, labels)
	
	query := []float64{50, 51}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = sc.Predict(query)
	}
}

func BenchmarkHierarchicalClassifierClassify(b *testing.B) {
	hc := NewHierarchicalClassifier()
	hc.Initialize()
	
	// 初始化并训练
	dataset := GenerateSyntheticDataset("bench", 100)
	trainingData := dataset.ToTrainingData()
	hc.Train(trainingData)
	
	fv := core.NewFeatureVector()
	fv.Set(core.FeatureTLSVersion, 0x0303)
	fv.Set(core.FeatureCipherSuites, 8)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hc.Classify(fv)
	}
}

func TestWeightedDistance(t *testing.T) {
	sc := NewSimpleClassifier(2)
	sc.weights = []float64{1.0, 2.0}
	
	// 测试不同权重的距离计算
	a := []float64{0, 0}
	b := []float64{3, 4}
	
	dist := sc.weightedDistance(a, b)
	
	// 期望距离: sqrt(1*3^2 + 2*4^2) = sqrt(9 + 32) = sqrt(41) ≈ 6.4
	expected := math.Sqrt(41)
	if math.Abs(dist-expected) > 0.001 {
		t.Errorf("weightedDistance = %v, want %v", dist, expected)
	}
}

func TestEmptyClassifier(t *testing.T) {
	sc := NewSimpleClassifier(5)
	
	// 未训练的分类器应该返回空结果
	label, conf := sc.Predict([]float64{1, 2, 3, 4, 5})
	if label != "" {
		t.Errorf("Expected empty label for untrained classifier, got %s", label)
	}
	if conf != 0 {
		t.Errorf("Expected 0 confidence for untrained classifier, got %f", conf)
	}
}
