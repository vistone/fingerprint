// Package ml provides machine learning fingerprint classification functionality
// Implements three-layer hierarchical classifier architecture, ported from Rust version ML design
package ml

import (
	"math"
	"sort"
	"sync"

	"github.com/vistone/fingerprint/modules/core"
)

// Classifier classifier interface
type Classifier interface {
	// Train train classifier
	Train(features [][]float64, labels []string) error
	// Predict predict label
	Predict(features []float64) (string, float64)
	// PredictTopK return top-K prediction results
	PredictTopK(features []float64, k int) []Prediction
}

// Prediction prediction result
type Prediction struct {
	Label      string
	Confidence float64
}

// SimpleClassifier simple classifier (distance-based)
type SimpleClassifier struct {
	classes    map[string][]float64 // Class centroids
	weights    []float64            // Feature weights
	classCount int
	mu         sync.RWMutex
}

// NewSimpleClassifier create new simple classifier
func NewSimpleClassifier(featureCount int) *SimpleClassifier {
	weights := make([]float64, featureCount)
	for i := range weights {
		weights[i] = 1.0
	}
	return &SimpleClassifier{
		classes:    make(map[string][]float64),
		weights:    weights,
		classCount: 0,
	}
}

// Train train classifier
func (c *SimpleClassifier) Train(features [][]float64, labels []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Calculate centroid for each class
	classSums := make(map[string][]float64)
	classCounts := make(map[string]int)

	for i, feat := range features {
		label := labels[i]
		if _, ok := classSums[label]; !ok {
			classSums[label] = make([]float64, len(feat))
		}
		for j, v := range feat {
			classSums[label][j] += v
		}
		classCounts[label]++
	}

	// Calculate average
	for label, sum := range classSums {
		count := classCounts[label]
		center := make([]float64, len(sum))
		for i, v := range sum {
			center[i] = v / float64(count)
		}
		c.classes[label] = center
	}

	c.classCount = len(c.classes)
	return nil
}

// Predict predict label
func (c *SimpleClassifier) Predict(features []float64) (string, float64) {
	predictions := c.PredictTopK(features, 1)
	if len(predictions) == 0 {
		return "", 0
	}
	return predictions[0].Label, predictions[0].Confidence
}

// PredictTopK return top-K prediction results
func (c *SimpleClassifier) PredictTopK(features []float64, k int) []Prediction {
	c.mu.RLock()
	defer c.mu.RUnlock()

	type distanceResult struct {
		label    string
		distance float64
	}

	// Calculate weighted distance to each class
	results := make([]distanceResult, 0, len(c.classes))
	for label, center := range c.classes {
		dist := c.weightedDistance(features, center)
		results = append(results, distanceResult{label: label, distance: dist})
	}

	// Sort by distance (smaller distance = more similar)
	sort.Slice(results, func(i, j int) bool {
		return results[i].distance < results[j].distance
	})

	// Convert to confidence scores
	predictions := make([]Prediction, 0, min(k, len(results)))
	for i := 0; i < k && i < len(results); i++ {
		// Convert distance to confidence (smaller distance = higher confidence)
		confidence := 1.0 / (1.0 + results[i].distance)
		predictions = append(predictions, Prediction{
			Label:      results[i].label,
			Confidence: confidence,
		})
	}

	return predictions
}

// weightedDistance calculate weighted Euclidean distance
func (c *SimpleClassifier) weightedDistance(a, b []float64) float64 {
	if len(a) != len(b) {
		return math.MaxFloat64
	}

	sum := 0.0
	for i := 0; i < len(a) && i < len(c.weights); i++ {
		diff := a[i] - b[i]
		sum += c.weights[i] * diff * diff
	}

	return math.Sqrt(sum)
}

// ProtocolClassifier protocol type classifier (layer 1)
type ProtocolClassifier struct {
	classifier *SimpleClassifier
}

// NewProtocolClassifier create new protocol classifier
func NewProtocolClassifier() *ProtocolClassifier {
	return &ProtocolClassifier{
		classifier: NewSimpleClassifier(10), // 10 protocol-related features
	}
}

// Train train protocol classifier
func (pc *ProtocolClassifier) Train(features [][]float64, labels []core.ProtocolType) error {
	strLabels := make([]string, len(labels))
	for i, l := range labels {
		strLabels[i] = string(l)
	}
	return pc.classifier.Train(features, strLabels)
}

// Predict predict protocol type
func (pc *ProtocolClassifier) Predict(features []float64) (core.ProtocolType, float64) {
	label, conf := pc.classifier.Predict(features)
	return core.ProtocolType(label), conf
}

// FamilyClassifier browser family classifier (layer 2)
type FamilyClassifier struct {
	classifier *SimpleClassifier
	protocol   core.ProtocolType
}

// NewFamilyClassifier create new browser family classifier
func NewFamilyClassifier(protocol core.ProtocolType) *FamilyClassifier {
	return &FamilyClassifier{
		classifier: NewSimpleClassifier(15), // 15 browser-related features
		protocol:   protocol,
	}
}

// Train train browser family classifier
func (fc *FamilyClassifier) Train(features [][]float64, labels []core.BrowserType) error {
	strLabels := make([]string, len(labels))
	for i, l := range labels {
		strLabels[i] = string(l)
	}
	return fc.classifier.Train(features, strLabels)
}

// Predict predict browser family
func (fc *FamilyClassifier) Predict(features []float64) (core.BrowserType, float64) {
	label, conf := fc.classifier.Predict(features)
	return core.BrowserType(label), conf
}

// VersionClassifier version recognition classifier (layer 3)
type VersionClassifier struct {
	classifier    *SimpleClassifier
	browserFamily core.BrowserType
}

// NewVersionClassifier create new version classifier
func NewVersionClassifier(family core.BrowserType) *VersionClassifier {
	return &VersionClassifier{
		classifier:    NewSimpleClassifier(20), // 20 version-related features
		browserFamily: family,
	}
}

// Train train version classifier
func (vc *VersionClassifier) Train(features [][]float64, labels []string) error {
	return vc.classifier.Train(features, labels)
}

// Predict predict version
func (vc *VersionClassifier) Predict(features []float64) (string, float64) {
	return vc.classifier.Predict(features)
}

// min returns the smaller of a and b
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
