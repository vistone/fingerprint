// Package defense provides security protection features
// includes passive detection, active protection and risk assessment
package defense

import (
	"math"
	"strings"
	"sync"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/ml"
)

// Detector security detector
type Detector struct {
	classifier *ml.HierarchicalClassifier
	extractor  *ml.FeatureExtractor
	rules      []DetectionRule
	mu         sync.RWMutex
}

// DetectionRule detection rule
type DetectionRule struct {
	Name        string
	Description string
	Condition   func(*core.FeatureVector) bool
	RiskScore   float64
}

// NewDetector creates new security detector
func NewDetector() *Detector {
	d := &Detector{
		classifier: ml.NewHierarchicalClassifier(),
		extractor:  ml.NewFeatureExtractor(),
		rules:      make([]DetectionRule, 0),
	}
	d.initializeRules()
	return d
}

// initializeRules initializes detection rules
func (d *Detector) initializeRules() {
	d.rules = []DetectionRule{
		{
			Name:        "headless_browser",
			Description: "detect headless browser",
			Condition: func(fv *core.FeatureVector) bool {
				return fv.Get(core.FeatureHeadlessBrowser) > 0.5
			},
			RiskScore: 0.7,
		},
		{
			Name:        "high_entropy",
			Description: "detect exceptionally high entropy values",
			Condition: func(fv *core.FeatureVector) bool {
				return fv.Get(core.FeatureEntropy) > 10.0
			},
			RiskScore: 0.5,
		},
		{
			Name:        "automation_tool",
			Description: "detect automation tool markers",
			Condition: func(fv *core.FeatureVector) bool {
				return fv.Get(core.FeatureToolMarker) > 0.3
			},
			RiskScore: 0.8,
		},
		{
			Name:        "inconsistent_behavior",
			Description: "detect inconsistent behavior",
			Condition: func(fv *core.FeatureVector) bool {
				return fv.Get(core.FeatureBehaviorPattern) < 0.3
			},
			RiskScore: 0.6,
		},
	}
}

// Detect executes detection
func (d *Detector) Detect(features *core.FeatureVector) *DetectionResult {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := &DetectionResult{
		Findings:    make([]Finding, 0),
		RiskFactors: make([]core.RiskFactor, 0),
		Labels:      make(map[string]string),
	}

	// run all rules
	for _, rule := range d.rules {
		if rule.Condition(features) {
			finding := Finding{
				Rule:        rule.Name,
				Description: rule.Description,
				RiskScore:   rule.RiskScore,
			}
			result.Findings = append(result.Findings, finding)
			result.RiskFactors = append(result.RiskFactors, core.RiskFactor{
				Name:        rule.Name,
				Weight:      rule.RiskScore,
				Description: rule.Description,
			})
			result.TotalRisk += rule.RiskScore
		}
	}

	// ML classification detection
	if d.classifier != nil {
		classification := d.classifier.Classify(features)
		result.Classification = classification
		result.Labels["protocol"] = string(classification.Protocol)
		result.Labels["family"] = string(classification.Family)
		result.Labels["version"] = classification.Version
	}

	// calculate final risk level
	result.RiskLevel = d.calculateRiskLevel(result.TotalRisk)
	result.RiskScore = math.Min(result.TotalRisk, 1.0)

	return result
}

// calculateRiskLevel calculates risk level
func (d *Detector) calculateRiskLevel(score float64) core.RiskLevel {
	return core.RiskLevelFromScore(score)
}

// DetectionResult detection result
type DetectionResult struct {
	Findings       []Finding
	RiskFactors    []core.RiskFactor
	Classification *ml.ClassificationResult
	RiskScore      float64
	RiskLevel      core.RiskLevel
	TotalRisk      float64
	Labels         map[string]string
}

// Finding detection finding
type Finding struct {
	Rule        string
	Description string
	RiskScore   float64
}

// IsThreat checks if it is a threat
func (r *DetectionResult) IsThreat() bool {
	return r.RiskLevel >= core.RiskLevelMedium
}

// PassiveDetector passive detector
type PassiveDetector struct {
	*Detector
	database *FingerprintDatabase
}

// FingerprintDatabase fingerprint database
type FingerprintDatabase struct {
	knownFingerprints map[string]FingerprintRecord
	mu                sync.RWMutex
}

// FingerprintRecord fingerprint record
type FingerprintRecord struct {
	ID        string
	Features  *core.FeatureVector
	Labels    map[string]string
	RiskScore float64
	FirstSeen int64
	LastSeen  int64
	SeenCount int
}

// NewPassiveDetector creates new passive detector
func NewPassiveDetector() *PassiveDetector {
	return &PassiveDetector{
		Detector: NewDetector(),
		database: &FingerprintDatabase{
			knownFingerprints: make(map[string]FingerprintRecord),
		},
	}
}

// Lookup looks up known fingerprints
func (pd *PassiveDetector) Lookup(fingerprintHash string) (FingerprintRecord, bool) {
	pd.database.mu.RLock()
	defer pd.database.mu.RUnlock()

	record, ok := pd.database.knownFingerprints[fingerprintHash]
	return record, ok
}

// Store stores fingerprint
func (pd *PassiveDetector) Store(record FingerprintRecord) {
	pd.database.mu.Lock()
	defer pd.database.mu.Unlock()

	pd.database.knownFingerprints[record.ID] = record
}

// ActiveProtector active protector
type ActiveProtector struct {
	noiseGenerators map[string]NoiseGenerator
	mu              sync.RWMutex
}

// NoiseGenerator noise generator interface
type NoiseGenerator interface {
	Generate(seed int64) interface{}
}

// NewActiveProtector creates new active protector
func NewActiveProtector() *ActiveProtector {
	return &ActiveProtector{
		noiseGenerators: make(map[string]NoiseGenerator),
	}
}

// RegisterNoiseGenerator registers noise generator
func (ap *ActiveProtector) RegisterNoiseGenerator(name string, generator NoiseGenerator) {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	ap.noiseGenerators[name] = generator
}

// GenerateNoise generates noise
func (ap *ActiveProtector) GenerateNoise(generatorName string, seed int64) interface{} {
	ap.mu.RLock()
	defer ap.mu.RUnlock()

	if gen, ok := ap.noiseGenerators[generatorName]; ok {
		return gen.Generate(seed)
	}
	return nil
}

// ProtectionConfig protection configuration
type ProtectionConfig struct {
	EnableCanvasNoise bool
	EnableAudioNoise  bool
	EnableWebGLNoise  bool
	EnableTimingNoise bool
	NoiseLevel        float64
}

// DefaultProtectionConfig default protection configuration
var DefaultProtectionConfig = &ProtectionConfig{
	EnableCanvasNoise: true,
	EnableAudioNoise:  true,
	EnableWebGLNoise:  true,
	EnableTimingNoise: true,
	NoiseLevel:        0.1,
}

// ApplyProtection applies protection measures
func (ap *ActiveProtector) ApplyProtection(config *ProtectionConfig) *ProtectionResult {
	result := &ProtectionResult{
		AppliedMeasures: make([]string, 0),
		NoiseData:       make(map[string]interface{}),
	}

	if config.EnableCanvasNoise {
		noise := ap.GenerateNoise("canvas", 0)
		result.NoiseData["canvas"] = noise
		result.AppliedMeasures = append(result.AppliedMeasures, "canvas_noise")
	}

	if config.EnableAudioNoise {
		noise := ap.GenerateNoise("audio", 0)
		result.NoiseData["audio"] = noise
		result.AppliedMeasures = append(result.AppliedMeasures, "audio_noise")
	}

	if config.EnableWebGLNoise {
		noise := ap.GenerateNoise("webgl", 0)
		result.NoiseData["webgl"] = noise
		result.AppliedMeasures = append(result.AppliedMeasures, "webgl_noise")
	}

	if config.EnableTimingNoise {
		noise := ap.GenerateNoise("timing", 0)
		result.NoiseData["timing"] = noise
		result.AppliedMeasures = append(result.AppliedMeasures, "timing_noise")
	}

	return result
}

// ProtectionResult protection result
type ProtectionResult struct {
	AppliedMeasures []string
	NoiseData       map[string]interface{}
}

// RiskEngine risk engine
type RiskEngine struct {
	passiveDetector *PassiveDetector
	activeProtector *ActiveProtector
	weights         RiskWeights
}

// RiskWeights risk weights
type RiskWeights struct {
	BehaviorWeight    float64
	FingerprintWeight float64
	ReputationWeight  float64
	AnomalyWeight     float64
}

// DefaultRiskWeights default risk weights
var DefaultRiskWeights = RiskWeights{
	BehaviorWeight:    0.3,
	FingerprintWeight: 0.3,
	ReputationWeight:  0.2,
	AnomalyWeight:     0.2,
}

// NewRiskEngine creates new risk engine
func NewRiskEngine() *RiskEngine {
	return &RiskEngine{
		passiveDetector: NewPassiveDetector(),
		activeProtector: NewActiveProtector(),
		weights:         DefaultRiskWeights,
	}
}

// Evaluate assesses risk
func (re *RiskEngine) Evaluate(features *core.FeatureVector, classification *ml.ClassificationResult) *core.RiskAssessment {
	// passive detection
	detectionResult := re.passiveDetector.Detect(features)

	// calculate comprehensive risk score
	behaviorScore := detectionResult.RiskScore * re.weights.BehaviorWeight
	fingerprintScore := (1.0 - classification.Confidence) * re.weights.FingerprintWeight

	// total risk score
	totalScore := behaviorScore + fingerprintScore

	// generate suggestions
	suggestions := re.generateSuggestions(detectionResult, classification)

	return &core.RiskAssessment{
		Score:       totalScore,
		Level:       re.calculateRiskLevel(totalScore),
		Factors:     detectionResult.RiskFactors,
		Suggestions: suggestions,
	}
}

// generateSuggestions generates protection suggestions
func (re *RiskEngine) generateSuggestions(detection *DetectionResult, classification *ml.ClassificationResult) []string {
	var suggestions []string

	if detection.RiskScore > 0.7 {
		suggestions = append(suggestions, "enable enhanced protection mode")
		suggestions = append(suggestions, "increase noise injection level")
	}

	if classification.Confidence < 0.5 {
		suggestions = append(suggestions, "require additional verification")
	}

	for _, finding := range detection.Findings {
		if strings.Contains(finding.Rule, "headless") {
			suggestions = append(suggestions, "detect headless browser features")
		}
		if strings.Contains(finding.Rule, "automation") {
			suggestions = append(suggestions, "verify non-human behavior")
		}
	}

	return suggestions
}

// calculateRiskLevel calculates risk level
func (re *RiskEngine) calculateRiskLevel(score float64) core.RiskLevel {
	return core.RiskLevelFromScore(score)
}

// DefenseSystem comprehensive defense system
type DefenseSystem struct {
	Detector   *Detector
	Protector  *ActiveProtector
	RiskEngine *RiskEngine
}

// NewDefenseSystem creates new defense system
func NewDefenseSystem() *DefenseSystem {
	return &DefenseSystem{
		Detector:   NewDetector(),
		Protector:  NewActiveProtector(),
		RiskEngine: NewRiskEngine(),
	}
}

// Analyze analyzes and returns protection suggestions
func (ds *DefenseSystem) Analyze(features *core.FeatureVector, classification *ml.ClassificationResult) *DefenseAdvice {
	// detection
	detection := ds.Detector.Detect(features)

	// risk assessment
	risk := ds.RiskEngine.Evaluate(features, classification)

	// generate protection configuration
	protectionConfig := ds.generateProtectionConfig(risk)

	// apply protection
	protection := ds.Protector.ApplyProtection(protectionConfig)

	return &DefenseAdvice{
		Detection:   detection,
		Risk:        risk,
		Protection:  protection,
		Recommended: protectionConfig,
	}
}

// generateProtectionConfig generates protection configuration based on risk
func (ds *DefenseSystem) generateProtectionConfig(risk *core.RiskAssessment) *ProtectionConfig {
	config := &ProtectionConfig{
		EnableCanvasNoise: true,
		EnableAudioNoise:  true,
		EnableWebGLNoise:  true,
		EnableTimingNoise: true,
		NoiseLevel:        0.1,
	}

	// adjust configuration based on risk level
	switch risk.Level {
	case core.RiskLevelCritical:
		config.NoiseLevel = 0.5
		config.EnableTimingNoise = true
	case core.RiskLevelHigh:
		config.NoiseLevel = 0.3
	case core.RiskLevelMedium:
		config.NoiseLevel = 0.2
	case core.RiskLevelLow:
		config.NoiseLevel = 0.1
		config.EnableTimingNoise = false
	default:
		config.EnableCanvasNoise = false
		config.EnableAudioNoise = false
		config.EnableWebGLNoise = false
	}

	return config
}

// DefenseAdvice defense advice
type DefenseAdvice struct {
	Detection   *DetectionResult
	Risk        *core.RiskAssessment
	Protection  *ProtectionResult
	Recommended *ProtectionConfig
}
