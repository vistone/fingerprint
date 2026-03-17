// Package ml provides training data loading and model training functionality
package ml

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/vistone/fingerprint/modules/core"
)

// TrainingSample training sample
type TrainingSample struct {
	ID       string                 `json:"id"`
	Features *core.FeatureVector    `json:"features"`
	Label    TrainingLabel          `json:"label"`
	Metadata map[string]interface{} `json:"metadata"`
}

// TrainingLabel training label
type TrainingLabel struct {
	Protocol core.ProtocolType `json:"protocol"`
	Family   core.BrowserType  `json:"family"`
	Version  string            `json:"version"`
	OS       string            `json:"os"`
}

// Dataset dataset
type Dataset struct {
	Name        string           `json:"name"`
	Version     string           `json:"version"`
	Description string           `json:"description"`
	Samples     []TrainingSample `json:"samples"`
	Statistics  DatasetStats     `json:"statistics"`
}

// DatasetStats dataset statistics
type DatasetStats struct {
	TotalSamples   int            `json:"total_samples"`
	ProtocolCounts map[string]int `json:"protocol_counts"`
	FamilyCounts   map[string]int `json:"family_counts"`
	VersionCounts  map[string]int `json:"version_counts"`
}

// DataLoader training data loader
type DataLoader struct {
	dataPath string
}

// NewDataLoader create new data loader
func NewDataLoader(dataPath string) *DataLoader {
	return &DataLoader{dataPath: dataPath}
}

// LoadDataset load dataset
func (dl *DataLoader) LoadDataset(filename string) (*Dataset, error) {
	path := filepath.Join(dl.dataPath, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read dataset file: %w", err)
	}

	var dataset Dataset
	if err := json.Unmarshal(data, &dataset); err != nil {
		return nil, fmt.Errorf("unmarshal dataset: %w", err)
	}

	return &dataset, nil
}

// LoadMultipleDatasets load multiple datasets and merge
func (dl *DataLoader) LoadMultipleDatasets(filenames []string) (*Dataset, error) {
	merged := &Dataset{
		Name:    "merged",
		Version: "1.0",
		Samples: make([]TrainingSample, 0),
		Statistics: DatasetStats{
			ProtocolCounts: make(map[string]int),
			FamilyCounts:   make(map[string]int),
			VersionCounts:  make(map[string]int),
		},
	}

	for _, filename := range filenames {
		dataset, err := dl.LoadDataset(filename)
		if err != nil {
			return nil, fmt.Errorf("load dataset %s: %w", filename, err)
		}
		merged.Samples = append(merged.Samples, dataset.Samples...)
	}

	merged.updateStatistics()
	return merged, nil
}

// updateStatistics update dataset statistics
func (d *Dataset) updateStatistics() {
	d.Statistics.TotalSamples = len(d.Samples)
	d.Statistics.ProtocolCounts = make(map[string]int)
	d.Statistics.FamilyCounts = make(map[string]int)
	d.Statistics.VersionCounts = make(map[string]int)

	for _, sample := range d.Samples {
		d.Statistics.ProtocolCounts[string(sample.Label.Protocol)]++
		d.Statistics.FamilyCounts[string(sample.Label.Family)]++
		d.Statistics.VersionCounts[sample.Label.Version]++
	}
}

// ToTrainingData convert to training data format
func (d *Dataset) ToTrainingData() *TrainingData {
	// Group by protocol
	protocolFeatures := make([][]float64, 0)
	protocolLabels := make([]core.ProtocolType, 0)

	familyFeatures := make(map[core.ProtocolType][][]float64)
	familyLabels := make(map[core.ProtocolType][]core.BrowserType)

	versionFeatures := make(map[core.BrowserType][][]float64)
	versionLabels := make(map[core.BrowserType][]string)

	for _, sample := range d.Samples {
		// Layer 1: Protocol
		protocolFeatures = append(protocolFeatures, d.extractProtocolFeatures(sample.Features))
		protocolLabels = append(protocolLabels, sample.Label.Protocol)

		// Layer 2: Family
		proto := sample.Label.Protocol
		if _, ok := familyFeatures[proto]; !ok {
			familyFeatures[proto] = make([][]float64, 0)
			familyLabels[proto] = make([]core.BrowserType, 0)
		}
		familyFeatures[proto] = append(familyFeatures[proto], d.extractFamilyFeatures(sample.Features))
		familyLabels[proto] = append(familyLabels[proto], sample.Label.Family)

		// Layer 3: Version
		family := sample.Label.Family
		if _, ok := versionFeatures[family]; !ok {
			versionFeatures[family] = make([][]float64, 0)
			versionLabels[family] = make([]string, 0)
		}
		versionFeatures[family] = append(versionFeatures[family], d.extractVersionFeatures(sample.Features))
		versionLabels[family] = append(versionLabels[family], sample.Label.Version)
	}

	return &TrainingData{
		ProtocolFeatures: protocolFeatures,
		ProtocolLabels:   protocolLabels,
		FamilyFeatures:   familyFeatures,
		FamilyLabels:     familyLabels,
		VersionFeatures:  versionFeatures,
		VersionLabels:    versionLabels,
	}
}

// extractProtocolFeatures extract protocol layer features
func (d *Dataset) extractProtocolFeatures(fv *core.FeatureVector) []float64 {
	return []float64{
		fv.Get(core.FeatureTLSVersion),
		fv.Get(core.FeatureCipherSuites),
		fv.Get(core.FeatureExtensions),
		fv.Get(core.FeatureHTTP2Settings),
		fv.Get(core.FeatureHTTPHeaders),
		fv.Get(core.FeatureEntropy),
	}
}

// extractFamilyFeatures extract family layer features
func (d *Dataset) extractFamilyFeatures(fv *core.FeatureVector) []float64 {
	return []float64{
		fv.Get(core.FeatureUserAgent),
		fv.Get(core.FeatureHTTPHeaders),
		fv.Get(core.FeatureCipherSuites),
		fv.Get(core.FeatureExtensions),
		fv.Get(core.FeatureHTTP2Settings),
		fv.Get(core.FeatureCanvas),
		fv.Get(core.FeatureWebGL),
		fv.Get(core.FeatureBehaviorPattern),
	}
}

// extractVersionFeatures extract version layer features
func (d *Dataset) extractVersionFeatures(fv *core.FeatureVector) []float64 {
	return []float64{
		fv.Get(core.FeatureTLSVersion),
		fv.Get(core.FeatureCipherSuites),
		fv.Get(core.FeatureExtensions),
		fv.Get(core.FeatureHTTP2Settings),
		fv.Get(core.FeatureHTTPHeaders),
		fv.Get(core.FeatureUserAgent),
		fv.Get(core.FeatureCanvas),
		fv.Get(core.FeatureWebGL),
		fv.Get(core.FeatureFonts),
		fv.Get(core.FeatureAudio),
	}
}

// SaveDataset save dataset to file
func (d *Dataset) SaveDataset(path string) error {
	d.updateStatistics()

	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal dataset: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write dataset: %w", err)
	}

	return nil
}

// PretrainedModel pretrained model
type PretrainedModel struct {
	Name            string                          `json:"name"`
	Version         string                          `json:"version"`
	Description     string                          `json:"description"`
	ProtocolCenters map[string][]float64            `json:"protocol_centers"`
	FamilyCenters   map[string]map[string][]float64 `json:"family_centers"`  // protocol -> family -> center
	VersionCenters  map[string]map[string][]float64 `json:"version_centers"` // family -> version -> center
	FeatureWeights  []float64                       `json:"feature_weights"`
}

// ModelLoader model loader
type ModelLoader struct {
	modelPath string
}

// NewModelLoader create new model loader
func NewModelLoader(modelPath string) *ModelLoader {
	return &ModelLoader{modelPath: modelPath}
}

// LoadModel load pretrained model
func (ml *ModelLoader) LoadModel(filename string) (*PretrainedModel, error) {
	path := filepath.Join(ml.modelPath, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read model file: %w", err)
	}

	var model PretrainedModel
	if err := json.Unmarshal(data, &model); err != nil {
		return nil, fmt.Errorf("unmarshal model: %w", err)
	}

	return &model, nil
}

// SaveModel save model
func (ml *ModelLoader) SaveModel(model *PretrainedModel, filename string) error {
	path := filepath.Join(ml.modelPath, filename)

	data, err := json.MarshalIndent(model, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal model: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write model: %w", err)
	}

	return nil
}

// ToClassifier convert pretrained model to classifier
func (pm *PretrainedModel) ToClassifier() *HierarchicalClassifier {
	hc := NewHierarchicalClassifier()
	hc.Initialize()

	// Load protocol centers
	for proto, center := range pm.ProtocolCenters {
		// Need to access private fields, actual implementation needs to provide methods
		_ = proto
		_ = center
	}

	return hc
}

// Trainer model trainer
type Trainer struct {
	classifier *HierarchicalClassifier
	dataset    *Dataset
}

// NewTrainer create new trainer
func NewTrainer(classifier *HierarchicalClassifier) *Trainer {
	return &Trainer{
		classifier: classifier,
	}
}

// LoadDataset load dataset
func (t *Trainer) LoadDataset(dataset *Dataset) {
	t.dataset = dataset
}

// Train train model
func (t *Trainer) Train() error {
	if t.dataset == nil {
		return fmt.Errorf("no dataset loaded")
	}

	trainingData := t.dataset.ToTrainingData()
	return t.classifier.Train(trainingData)
}

// TrainWithProgress training with progress callback
func (t *Trainer) TrainWithProgress(progress func(epoch, total int, loss float64)) error {
	// Simplified implementation, should support multi-epoch training in practice
	return t.Train()
}

// ExportModel export trained model
func (t *Trainer) ExportModel(name, version string) *PretrainedModel {
	model := &PretrainedModel{
		Name:            name,
		Version:         version,
		ProtocolCenters: make(map[string][]float64),
		FamilyCenters:   make(map[string]map[string][]float64),
		VersionCenters:  make(map[string]map[string][]float64),
	}

	// Export centers from classifier
	// Note: actual implementation needs to access classifier internal state

	return model
}

// GenerateSyntheticDataset generate synthetic training data
func GenerateSyntheticDataset(name string, sampleCount int) *Dataset {
	dataset := newSyntheticDataset(name, sampleCount)
	for i := 0; i < sampleCount; i++ {
		browser := syntheticBrowsers[i%len(syntheticBrowsers)]
		protocol := syntheticProtocols[i%len(syntheticProtocols)]
		dataset.Samples = append(dataset.Samples, buildSyntheticSample(i, browser, protocol))
	}

	dataset.updateStatistics()
	return dataset
}

type syntheticBrowserConfig struct {
	family  core.BrowserType
	version string
}

var syntheticBrowsers = []syntheticBrowserConfig{
	{core.BrowserChrome, "133"},
	{core.BrowserChrome, "131"},
	{core.BrowserFirefox, "133"},
	{core.BrowserFirefox, "132"},
	{core.BrowserSafari, "18.0"},
	{core.BrowserSafari, "17.0"},
}

var syntheticProtocols = []core.ProtocolType{
	core.ProtocolTLS,
	core.ProtocolHTTP2,
	core.ProtocolHTTP3,
}

func newSyntheticDataset(name string, sampleCount int) *Dataset {
	return &Dataset{
		Name:        name,
		Version:     "1.0",
		Description: "Synthetic training dataset",
		Samples:     make([]TrainingSample, 0, sampleCount),
		Statistics: DatasetStats{
			ProtocolCounts: make(map[string]int),
			FamilyCounts:   make(map[string]int),
			VersionCounts:  make(map[string]int),
		},
	}
}

func buildSyntheticSample(i int, browser syntheticBrowserConfig, protocol core.ProtocolType) TrainingSample {
	fv := buildSyntheticFeatureVector(i, browser.family)
	return TrainingSample{
		ID:       fmt.Sprintf("sample_%d", i),
		Features: fv,
		Label: TrainingLabel{
			Protocol: protocol,
			Family:   browser.family,
			Version:  browser.version,
		},
		Metadata: map[string]interface{}{
			"synthetic": true,
			"index":     i,
		},
	}
}

func buildSyntheticFeatureVector(i int, family core.BrowserType) *core.FeatureVector {
	fv := core.NewFeatureVector()
	fv.Set(core.FeatureTLSVersion, 0x0303)
	fv.Set(core.FeatureCipherSuites, float64(8+i%5))
	fv.Set(core.FeatureExtensions, float64(10+i%8))
	fv.Set(core.FeatureHTTP2Settings, float64(65536+i*1000))
	fv.Set(core.FeatureHTTPHeaders, float64(10+i%3))

	switch family {
	case core.BrowserChrome:
		fv.Set(core.FeatureUserAgent, 100.0+float64(i%20))
		fv.Set(core.FeatureCanvas, 50.0+float64(i%10))
		fv.Set(core.FeatureWebGL, 80.0+float64(i%15))
	case core.BrowserFirefox:
		fv.Set(core.FeatureUserAgent, 200.0+float64(i%20))
		fv.Set(core.FeatureCanvas, 60.0+float64(i%10))
		fv.Set(core.FeatureWebGL, 70.0+float64(i%15))
	case core.BrowserSafari:
		fv.Set(core.FeatureUserAgent, 300.0+float64(i%20))
		fv.Set(core.FeatureCanvas, 55.0+float64(i%10))
		fv.Set(core.FeatureWebGL, 75.0+float64(i%15))
	}
	return fv
}
