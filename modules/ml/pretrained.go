// Package ml provides pretrained models and initialization functionality
package ml

import (
	"github.com/vistone/fingerprint/modules/core"
)

// InitWithSyntheticData initialize classifier with synthetic data
func InitWithSyntheticData(classifier *HierarchicalClassifier, sampleCount int) error {
	// Generate synthetic dataset
	dataset := GenerateSyntheticDataset("synthetic_init", sampleCount)
	
	// Convert to training data
	trainingData := dataset.ToTrainingData()
	
	// Train classifier
	return classifier.Train(trainingData)
}

// InitPretrainedClassifier create and initialize pretrained classifier
func InitPretrainedClassifier() *HierarchicalClassifier {
	hc := NewHierarchicalClassifier()
	hc.Initialize()
	
	// Initialize with synthetic data (should load real training data in practice)
	InitWithSyntheticData(hc, 1000)
	
	return hc
}

// DefaultClassifier default pretrained classifier (singleton)
var DefaultClassifier = InitPretrainedClassifier()

// QuickClassify quickly classify using default classifier
func QuickClassify(features *core.FeatureVector) *ClassificationResult {
	return DefaultClassifier.Classify(features)
}

// ExportBuiltinModel export builtin model
func ExportBuiltinModel() *PretrainedModel {
	model := &PretrainedModel{
		Name:        "fingerprint_builtin",
		Version:     "1.0.0",
		Description: "Built-in pretrained model with synthetic data",
		ProtocolCenters: map[string][]float64{
			"tls":   {0x0303, 8.0, 10.0, 65536.0, 10.0, 15.0},
			"http2": {0x0303, 8.0, 10.0, 65536.0, 10.0, 15.0},
			"http3": {0x0304, 6.0, 8.0, 65536.0, 10.0, 20.0},
		},
		FamilyCenters: map[string]map[string][]float64{
			"tls": {
				"chrome":  {100.0, 50.0, 8.0, 10.0, 65536.0, 50.0, 80.0, 0.5},
				"firefox": {200.0, 60.0, 8.0, 10.0, 65536.0, 60.0, 70.0, 0.5},
				"safari":  {300.0, 55.0, 8.0, 10.0, 65536.0, 55.0, 75.0, 0.5},
			},
		},
		VersionCenters: map[string]map[string][]float64{
			"chrome": {
				"133": {0x0303, 8.0, 11.0, 65536.0, 10.0, 100.0, 50.0, 80.0, 100.0, 0.0},
				"131": {0x0303, 8.0, 10.0, 65536.0, 10.0, 100.0, 50.0, 80.0, 100.0, 0.0},
				"120": {0x0303, 8.0, 10.0, 65536.0, 10.0, 100.0, 50.0, 80.0, 100.0, 0.0},
			},
			"firefox": {
				"133": {0x0303, 11.0, 11.0, 131072.0, 10.0, 200.0, 60.0, 70.0, 100.0, 0.0},
				"132": {0x0303, 11.0, 11.0, 131072.0, 10.0, 200.0, 60.0, 70.0, 100.0, 0.0},
			},
			"safari": {
				"18.0": {0x0303, 8.0, 9.0, 2097152.0, 10.0, 300.0, 55.0, 75.0, 100.0, 0.0},
				"17.0": {0x0303, 8.0, 9.0, 2097152.0, 10.0, 300.0, 55.0, 75.0, 100.0, 0.0},
			},
		},
		FeatureWeights: []float64{1.0, 1.0, 0.8, 0.6, 0.9, 0.7, 0.7, 0.5},
	}
	
	return model
}

// ModelInfo model information
type ModelInfo struct {
	Name           string
	Version        string
	Description    string
	ProtocolCount  int
	FamilyCount    int
	VersionCount   int
	TotalCenters   int
}

// GetModelInfo get model information
func GetModelInfo(model *PretrainedModel) *ModelInfo {
	protocolCount := len(model.ProtocolCenters)
	
	familyCount := 0
	for _, families := range model.FamilyCenters {
		familyCount += len(families)
	}
	
	versionCount := 0
	for _, versions := range model.VersionCenters {
		versionCount += len(versions)
	}
	
	return &ModelInfo{
		Name:          model.Name,
		Version:       model.Version,
		Description:   model.Description,
		ProtocolCount: protocolCount,
		FamilyCount:   familyCount,
		VersionCount:  versionCount,
		TotalCenters:  protocolCount + familyCount + versionCount,
	}
}
