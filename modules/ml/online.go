package ml

import (
	"math"
	"sync"

	"github.com/vistone/fingerprint/modules/core"
)

// OnlineClassifier wraps a SimpleClassifier and supports incremental (online) learning.
//
// Instead of retraining from scratch, it updates class centroids incrementally as new
// samples arrive. This is essential for adapting to concept drift in real-time fingerprint
// classification where browser versions and fingerprint patterns evolve continuously.
type OnlineClassifier struct {
	classifier  *SimpleClassifier
	classCounts map[string]int // running sample count per class
	mu          sync.Mutex

	// Concept drift detection
	recentCorrect int // correct predictions in recent window
	recentTotal   int // total predictions in recent window
	windowSize    int // sliding window size for accuracy tracking
	driftThreshold float64 // accuracy below this triggers drift alert
}

// NewOnlineClassifier creates a new online learning classifier.
// featureCount is the dimensionality of the feature vector.
func NewOnlineClassifier(featureCount int) *OnlineClassifier {
	return &OnlineClassifier{
		classifier:     NewSimpleClassifier(featureCount),
		classCounts:    make(map[string]int),
		windowSize:     200,
		driftThreshold: 0.6,
	}
}

// NewOnlineClassifierFrom wraps an existing trained SimpleClassifier for online updates.
// initialCounts maps each class label to the number of training samples used to build
// its centroid; this is required so incremental updates are weighted correctly.
func NewOnlineClassifierFrom(base *SimpleClassifier, initialCounts map[string]int) *OnlineClassifier {
	counts := make(map[string]int, len(initialCounts))
	for k, v := range initialCounts {
		counts[k] = v
	}
	return &OnlineClassifier{
		classifier:     base,
		classCounts:    counts,
		windowSize:     200,
		driftThreshold: 0.6,
	}
}

// PartialFit incrementally updates the classifier with a single new sample.
//
// For an existing class, the centroid is updated using the online mean formula:
//
//	new_centroid = old_centroid + (x - old_centroid) / (n + 1)
//
// For a new (unseen) class, a new centroid is created from the sample.
func (oc *OnlineClassifier) PartialFit(features []float64, label string) {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	oc.classifier.mu.Lock()
	defer oc.classifier.mu.Unlock()

	centroid, exists := oc.classifier.classes[label]
	count := oc.classCounts[label]

	if !exists {
		// New class: initialize centroid from this sample
		newCentroid := make([]float64, len(features))
		copy(newCentroid, features)
		oc.classifier.classes[label] = newCentroid
		oc.classCounts[label] = 1
		oc.classifier.classCount = len(oc.classifier.classes)
		return
	}

	// Incremental centroid update
	n := float64(count + 1)
	for i := 0; i < len(centroid) && i < len(features); i++ {
		centroid[i] += (features[i] - centroid[i]) / n
	}
	oc.classCounts[label] = count + 1
}

// PartialFitBatch updates the classifier with a batch of samples.
func (oc *OnlineClassifier) PartialFitBatch(features [][]float64, labels []string) {
	for i := range features {
		oc.PartialFit(features[i], labels[i])
	}
}

// Predict returns the predicted label and confidence.
func (oc *OnlineClassifier) Predict(features []float64) (string, float64) {
	return oc.classifier.Predict(features)
}

// PredictTopK returns top-K predictions.
func (oc *OnlineClassifier) PredictTopK(features []float64, k int) []Prediction {
	return oc.classifier.PredictTopK(features, k)
}

// RecordOutcome records whether a prediction was correct for drift detection.
func (oc *OnlineClassifier) RecordOutcome(correct bool) {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	oc.recentTotal++
	if correct {
		oc.recentCorrect++
	}

	// Keep a sliding window
	if oc.recentTotal > oc.windowSize {
		// Approximate halving to maintain recency bias
		oc.recentCorrect /= 2
		oc.recentTotal /= 2
	}
}

// DriftDetected returns true if the recent accuracy has dropped below the drift threshold.
func (oc *OnlineClassifier) DriftDetected() bool {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	if oc.recentTotal < oc.windowSize/4 {
		return false // Not enough data to judge
	}
	accuracy := float64(oc.recentCorrect) / float64(oc.recentTotal)
	return accuracy < oc.driftThreshold
}

// Accuracy returns the recent prediction accuracy in [0,1].
func (oc *OnlineClassifier) Accuracy() float64 {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	if oc.recentTotal == 0 {
		return 0
	}
	return float64(oc.recentCorrect) / float64(oc.recentTotal)
}

// ClassCount returns the number of distinct classes known to the classifier.
func (oc *OnlineClassifier) ClassCount() int {
	oc.mu.Lock()
	defer oc.mu.Unlock()
	return len(oc.classCounts)
}

// SampleCount returns the total number of samples that have been used for training.
func (oc *OnlineClassifier) SampleCount() int {
	oc.mu.Lock()
	defer oc.mu.Unlock()
	total := 0
	for _, c := range oc.classCounts {
		total += c
	}
	return total
}

// OnlineHierarchicalClassifier wraps three OnlineClassifiers for the three-layer
// hierarchical classification pipeline: Protocol → Family → Version.
type OnlineHierarchicalClassifier struct {
	Protocol *OnlineClassifier
	Family   map[string]*OnlineClassifier // keyed by protocol type
	Version  map[string]*OnlineClassifier // keyed by browser family
	mu       sync.RWMutex
}

// NewOnlineHierarchicalClassifier creates a three-layer online classifier.
func NewOnlineHierarchicalClassifier() *OnlineHierarchicalClassifier {
	return &OnlineHierarchicalClassifier{
		Protocol: NewOnlineClassifier(10),
		Family:   make(map[string]*OnlineClassifier),
		Version:  make(map[string]*OnlineClassifier),
	}
}

// UpdateProtocol incrementally trains the protocol layer.
func (ohc *OnlineHierarchicalClassifier) UpdateProtocol(features []float64, label string) {
	ohc.Protocol.PartialFit(features, label)
}

// UpdateFamily incrementally trains the family layer for a given protocol.
func (ohc *OnlineHierarchicalClassifier) UpdateFamily(protocol string, features []float64, label string) {
	ohc.mu.Lock()
	fc, ok := ohc.Family[protocol]
	if !ok {
		fc = NewOnlineClassifier(15)
		ohc.Family[protocol] = fc
	}
	ohc.mu.Unlock()
	fc.PartialFit(features, label)
}

// UpdateVersion incrementally trains the version layer for a given browser family.
func (ohc *OnlineHierarchicalClassifier) UpdateVersion(family string, features []float64, label string) {
	ohc.mu.Lock()
	vc, ok := ohc.Version[family]
	if !ok {
		vc = NewOnlineClassifier(20)
		ohc.Version[family] = vc
	}
	ohc.mu.Unlock()
	vc.PartialFit(features, label)
}

// Classify performs three-layer classification with the online-updated models.
func (ohc *OnlineHierarchicalClassifier) Classify(features []float64) *ClassificationResult {
	result := &ClassificationResult{}

	// Layer 1: Protocol
	protocolFeatures := features
	if len(features) > 10 {
		protocolFeatures = features[:10]
	}
	proto, protoConf := ohc.Protocol.Predict(protocolFeatures)
	result.Protocol = core.ProtocolType(proto)
	result.ProtocolConfidence = protoConf

	// Layer 2: Family
	ohc.mu.RLock()
	fc, ok := ohc.Family[proto]
	ohc.mu.RUnlock()
	if ok {
		familyFeatures := features
		if len(features) > 15 {
			familyFeatures = features[:15]
		}
		family, familyConf := fc.Predict(familyFeatures)
		result.Family = core.BrowserType(family)
		result.FamilyConfidence = familyConf

		// Layer 3: Version
		ohc.mu.RLock()
		vc, ok := ohc.Version[family]
		ohc.mu.RUnlock()
		if ok {
			versionFeatures := features
			if len(features) > 20 {
				versionFeatures = features[:20]
			}
			version, versionConf := vc.Predict(versionFeatures)
			result.Version = version
			result.VersionConfidence = versionConf
		}
	}

	// Composite confidence
	result.Confidence = 0.3*result.ProtocolConfidence + 0.3*result.FamilyConfidence + 0.4*result.VersionConfidence

	return result
}

// Stats returns per-layer statistics.
func (ohc *OnlineHierarchicalClassifier) Stats() OnlineHierarchicalStats {
	ohc.mu.RLock()
	defer ohc.mu.RUnlock()

	familyStats := make(map[string]OnlineLayerStats, len(ohc.Family))
	for k, v := range ohc.Family {
		familyStats[k] = OnlineLayerStats{
			Classes: v.ClassCount(),
			Samples: v.SampleCount(),
		}
	}

	versionStats := make(map[string]OnlineLayerStats, len(ohc.Version))
	for k, v := range ohc.Version {
		versionStats[k] = OnlineLayerStats{
			Classes: v.ClassCount(),
			Samples: v.SampleCount(),
		}
	}

	return OnlineHierarchicalStats{
		Protocol: OnlineLayerStats{
			Classes: ohc.Protocol.ClassCount(),
			Samples: ohc.Protocol.SampleCount(),
		},
		Family:  familyStats,
		Version: versionStats,
	}
}

// OnlineHierarchicalStats contains per-layer statistics.
type OnlineHierarchicalStats struct {
	Protocol OnlineLayerStats            `json:"protocol"`
	Family   map[string]OnlineLayerStats `json:"family"`
	Version  map[string]OnlineLayerStats `json:"version"`
}

// OnlineLayerStats contains statistics for a single classifier layer.
type OnlineLayerStats struct {
	Classes int `json:"classes"`
	Samples int `json:"samples"`
}

// WeightedOnlineUpdate performs an importance-weighted partial fit.
// Samples with higher weight shift the centroid more aggressively.
// weight should be in (0, 1]; values > 1 are clamped.
func (oc *OnlineClassifier) WeightedPartialFit(features []float64, label string, weight float64) {
	if weight <= 0 {
		return
	}
	weight = math.Min(weight, 1.0)

	oc.mu.Lock()
	defer oc.mu.Unlock()

	oc.classifier.mu.Lock()
	defer oc.classifier.mu.Unlock()

	centroid, exists := oc.classifier.classes[label]
	count := oc.classCounts[label]

	if !exists {
		newCentroid := make([]float64, len(features))
		copy(newCentroid, features)
		oc.classifier.classes[label] = newCentroid
		oc.classCounts[label] = 1
		oc.classifier.classCount = len(oc.classifier.classes)
		return
	}

	// Weighted incremental update: higher weight → larger step toward the new sample
	effectiveN := float64(count+1) / weight
	for i := 0; i < len(centroid) && i < len(features); i++ {
		centroid[i] += (features[i] - centroid[i]) / effectiveN
	}
	oc.classCounts[label] = count + 1
}
