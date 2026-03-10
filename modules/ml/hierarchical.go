// Package ml provides three-layer hierarchical classifier implementation
package ml

import (
	"sync"

	"github.com/vistone/fingerprint/modules/core"
)

// HierarchicalClassifier three-layer hierarchical classifier
// Ported from Rust version ML architecture
// Layer 1: Protocol type classifier (TLS/HTTP/QUIC)
// Layer 2: Browser family classifier (Chrome/Firefox/Safari)
// Layer 3: Version recognition classifier (Chrome 120/121/122)
type HierarchicalClassifier struct {
	// Layer 1: Protocol type classifier
	protocolClassifier *ProtocolClassifier

	// Layer 2: Browser family classifiers, indexed by protocol type
	familyClassifiers map[core.ProtocolType]*FamilyClassifier

	// Layer 3: Version recognition classifiers, indexed by browser family
	versionClassifiers map[core.BrowserType]*VersionClassifier

	// Training status
	trained bool
	mu      sync.RWMutex
}

// NewHierarchicalClassifier create new three-layer hierarchical classifier
func NewHierarchicalClassifier() *HierarchicalClassifier {
	return &HierarchicalClassifier{
		protocolClassifier: NewProtocolClassifier(),
		familyClassifiers:  make(map[core.ProtocolType]*FamilyClassifier),
		versionClassifiers: make(map[core.BrowserType]*VersionClassifier),
		trained:            false,
	}
}

// Initialize initialize classifier hierarchy
func (hc *HierarchicalClassifier) Initialize() {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	// Initialize Layer 2: create browser family classifier for each protocol type
	protocols := []core.ProtocolType{
		core.ProtocolTLS,
		core.ProtocolHTTP,
		core.ProtocolHTTP2,
		core.ProtocolQUIC,
		core.ProtocolHTTP3,
	}

	for _, proto := range protocols {
		hc.familyClassifiers[proto] = NewFamilyClassifier(proto)
	}

	// Initialize Layer 3: create version classifier for each browser family
	families := []core.BrowserType{
		core.BrowserChrome,
		core.BrowserFirefox,
		core.BrowserSafari,
		core.BrowserOpera,
		core.BrowserEdge,
	}

	for _, family := range families {
		hc.versionClassifiers[family] = NewVersionClassifier(family)
	}
}

// TrainingData training data structure
type TrainingData struct {
	ProtocolFeatures [][]float64
	ProtocolLabels   []core.ProtocolType

	FamilyFeatures map[core.ProtocolType][][]float64
	FamilyLabels   map[core.ProtocolType][]core.BrowserType

	VersionFeatures map[core.BrowserType][][]float64
	VersionLabels   map[core.BrowserType][]string
}

// Train train three-layer hierarchical classifier
func (hc *HierarchicalClassifier) Train(data *TrainingData) error {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	// Train Layer 1: protocol classifier
	if err := hc.protocolClassifier.Train(data.ProtocolFeatures, data.ProtocolLabels); err != nil {
		return err
	}

	// Train Layer 2: browser family classifiers
	for proto, features := range data.FamilyFeatures {
		if classifier, ok := hc.familyClassifiers[proto]; ok {
			labels := data.FamilyLabels[proto]
			if err := classifier.Train(features, labels); err != nil {
				return err
			}
		}
	}

	// Train Layer 3: version classifiers
	for family, features := range data.VersionFeatures {
		if classifier, ok := hc.versionClassifiers[family]; ok {
			labels := data.VersionLabels[family]
			if err := classifier.Train(features, labels); err != nil {
				return err
			}
		}
	}

	hc.trained = true
	return nil
}

// ClassificationResult classification result
type ClassificationResult struct {
	Protocol   core.ProtocolType
	Family     core.BrowserType
	Version    string
	Confidence float64
	Labels     map[string]string
	LayerScores
}

// LayerScores scores for each layer
type LayerScores struct {
	ProtocolConfidence float64
	FamilyConfidence   float64
	VersionConfidence  float64
}

// Classify perform hierarchical classification
func (hc *HierarchicalClassifier) Classify(features *core.FeatureVector) *ClassificationResult {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	if !hc.trained {
		return &ClassificationResult{
			Labels: map[string]string{"error": "classifier not trained"},
		}
	}

	result := &ClassificationResult{
		Labels: make(map[string]string),
	}

	// Layer 1: protocol type classification
	protocolFeatures := hc.extractProtocolFeatures(features)
	protocol, protocolConf := hc.protocolClassifier.Predict(protocolFeatures)
	result.Protocol = protocol
	result.ProtocolConfidence = protocolConf
	result.Labels["protocol"] = string(protocol)

	// Layer 2: browser family classification
	familyClassifier, ok := hc.familyClassifiers[protocol]
	if !ok {
		result.Confidence = protocolConf
		return result
	}

	familyFeatures := hc.extractFamilyFeatures(features)
	family, familyConf := familyClassifier.Predict(familyFeatures)
	result.Family = family
	result.FamilyConfidence = familyConf
	result.Labels["family"] = string(family)

	// Layer 3: version recognition
	versionClassifier, ok := hc.versionClassifiers[family]
	if !ok {
		// When only two layers, use weighted average
		result.Confidence = 0.5*protocolConf + 0.5*familyConf
		return result
	}

	versionFeatures := hc.extractVersionFeatures(features)
	version, versionConf := versionClassifier.Predict(versionFeatures)
	result.Version = version
	result.VersionConfidence = versionConf
	result.Labels["version"] = version

	// Overall confidence: use weighted average to avoid exponential decay
	// Protocol weight 0.3, family weight 0.3, version weight 0.4
	result.Confidence = 0.3*protocolConf + 0.3*familyConf + 0.4*versionConf

	return result
}

// ClassifyBatch batch classification
func (hc *HierarchicalClassifier) ClassifyBatch(features []*core.FeatureVector) []*ClassificationResult {
	results := make([]*ClassificationResult, len(features))
	for i, fv := range features {
		results[i] = hc.Classify(fv)
	}
	return results
}

// extractProtocolFeatures extract protocol-related features
func (hc *HierarchicalClassifier) extractProtocolFeatures(fv *core.FeatureVector) []float64 {
	// Extract protocol recognition related features
	features := []float64{
		fv.Get(core.FeatureTLSVersion),
		fv.Get(core.FeatureCipherSuites),
		fv.Get(core.FeatureExtensions),
		fv.Get(core.FeatureHTTP2Settings),
		fv.Get(core.FeatureHTTPHeaders),
		fv.Get(core.FeatureEntropy),
	}
	return features
}

// extractFamilyFeatures extract browser family related features
func (hc *HierarchicalClassifier) extractFamilyFeatures(fv *core.FeatureVector) []float64 {
	// Extract browser recognition related features
	features := []float64{
		fv.Get(core.FeatureUserAgent),
		fv.Get(core.FeatureHTTPHeaders),
		fv.Get(core.FeatureCipherSuites),
		fv.Get(core.FeatureExtensions),
		fv.Get(core.FeatureHTTP2Settings),
		fv.Get(core.FeatureCanvas),
		fv.Get(core.FeatureWebGL),
		fv.Get(core.FeatureBehaviorPattern),
	}
	return features
}

// extractVersionFeatures extract version recognition related features
func (hc *HierarchicalClassifier) extractVersionFeatures(fv *core.FeatureVector) []float64 {
	// Extract version recognition related features
	features := []float64{
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
	return features
}

// GetConfidenceThresholds get confidence thresholds for each layer
func (hc *HierarchicalClassifier) GetConfidenceThresholds() (protocol, family, version float64) {
	return 0.7, 0.8, 0.6
}

// IsHighConfidence determine if the classification is high confidence
func (r *ClassificationResult) IsHighConfidence() bool {
	return r.Confidence >= 0.8 &&
		r.ProtocolConfidence >= 0.7 &&
		r.FamilyConfidence >= 0.8 &&
		r.VersionConfidence >= 0.6
}

// FingerprintMatcher fingerprint matcher
type FingerprintMatcher struct {
	classifier *HierarchicalClassifier
}

// NewFingerprintMatcher create new fingerprint matcher
func NewFingerprintMatcher() *FingerprintMatcher {
	return &FingerprintMatcher{
		classifier: NewHierarchicalClassifier(),
	}
}

// Initialize initialize matcher
func (fm *FingerprintMatcher) Initialize() {
	fm.classifier.Initialize()
}

// Train train matcher
func (fm *FingerprintMatcher) Train(data *TrainingData) error {
	return fm.classifier.Train(data)
}

// Match match fingerprint
func (fm *FingerprintMatcher) Match(features *core.FeatureVector) *ClassificationResult {
	return fm.classifier.Classify(features)
}

// MatchWithProfile match with known profiles
func (fm *FingerprintMatcher) MatchWithProfile(features *core.FeatureVector, profiles []core.FingerprintSpec) *ClassificationResult {
	result := fm.classifier.Classify(features)

	// Find closest match in known profiles
	bestMatch := ""
	bestScore := 0.0

	for _, profile := range profiles {
		score := fm.calculateMatchScore(result, profile)
		if score > bestScore {
			bestScore = score
			bestMatch = profile.GetID()
		}
	}

	result.Labels["best_match"] = bestMatch
	result.Labels["match_score"] = ""

	return result
}

// calculateMatchScore calculate match score
func (fm *FingerprintMatcher) calculateMatchScore(result *ClassificationResult, profile core.FingerprintSpec) float64 {
	score := 0.0
	if result.Family == profile.GetBrowserType() {
		score += 0.5
	}
	if result.Labels["best_match"] == profile.GetID() {
		score += 0.5
	}
	return score
}
