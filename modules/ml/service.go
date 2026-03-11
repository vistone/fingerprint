// Package ml — service.go provides a central ML service that unifies
// inference, validation, feedback, and generation across the entire project.
//
// MLService is the single entry point for all AI capabilities:
//
//   - Infer:    Run the 4-stage neural pipeline on any fingerprint
//   - Validate: Check if a generated fingerprint looks realistic
//   - Feedback: Feed real-world observations back for online learning
//   - Generate: Produce ML-guided fingerprint feature vectors
//   - Evolve:   Trigger incremental model evolution from new profiles
//
// Typical usage:
//
//	svc, _ := ml.NewMLService(&ml.ServiceConfig{ModelStorePath: "./models"})
//	result := svc.Infer(profile, nil)    // inference
//	ok     := svc.Validate(profile)      // validation
//	svc.Feedback(profile, 0.9)           // reward feedback
package ml

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/profiles"
)

// =========================================================================
// ServiceConfig
// =========================================================================

// ServiceConfig configures the central ML service.
type ServiceConfig struct {
	// ModelStorePath is the directory for persisted model snapshots.
	ModelStorePath string

	// MaxStoreVersions is the maximum number of model versions retained.
	MaxStoreVersions int

	// AutoLoadLatest loads the latest model snapshot on startup.
	AutoLoadLatest bool

	// FeedbackBufferSize is the capacity of the asynchronous feedback buffer.
	FeedbackBufferSize int

	// EvolveInterval is how often the service checks for evolution triggers.
	EvolveInterval time.Duration

	// DriftThreshold is the accuracy drop that triggers automatic evolution.
	DriftThreshold float64

	// ValidationForgeryThreshold: generated fingerprints above this forgery
	// probability are rejected.
	ValidationForgeryThreshold float64

	// ValidationConsistencyMin: generated fingerprints below this cross-layer
	// consistency score are rejected.
	ValidationConsistencyMin float64
}

// DefaultServiceConfig provides sensible defaults.
var DefaultServiceConfig = &ServiceConfig{
	ModelStorePath:             "./models",
	MaxStoreVersions:           10,
	AutoLoadLatest:             true,
	FeedbackBufferSize:         1024,
	EvolveInterval:             24 * time.Hour,
	DriftThreshold:             0.05,
	ValidationForgeryThreshold: 0.3,
	ValidationConsistencyMin:   0.7,
}

// =========================================================================
// MLService — central AI brain
// =========================================================================

// MLService is the project-wide ML service singleton.
type MLService struct {
	config   *ServiceConfig
	pipeline *ModelPipeline
	store    *ModelStore
	learner  *OnlineLearner
	mu       sync.RWMutex

	// stats
	inferCount    atomic.Int64
	feedbackCount atomic.Int64
	evolveCount   atomic.Int64
}

// NewMLService creates and initialises the ML service.
func NewMLService(config *ServiceConfig) (*MLService, error) {
	if config == nil {
		config = DefaultServiceConfig
	}

	svc := &MLService{
		config:   config,
		pipeline: NewModelPipeline(),
	}

	// Initialise model store.
	if config.ModelStorePath != "" {
		storeConfig := DefaultStoreConfig(config.ModelStorePath)
		if config.MaxStoreVersions > 0 {
			storeConfig.MaxVersions = config.MaxStoreVersions
		}
		store, err := NewModelStore(storeConfig)
		if err != nil {
			return nil, fmt.Errorf("ml service: init store: %w", err)
		}
		svc.store = store

		// Auto-load latest.
		if config.AutoLoadLatest {
			if loaded, err := svc.pipeline.LoadFromStore(store); err != nil {
				return nil, fmt.Errorf("ml service: load model: %w", err)
			} else if loaded {
				svc.pipeline.trained = true
			}
		}
	}

	// Initialise online learner.
	svc.learner = NewOnlineLearner(&OnlineLearnerConfig{
		BufferSize:     config.FeedbackBufferSize,
		DriftThreshold: config.DriftThreshold,
	})

	return svc, nil
}

// Pipeline returns the underlying model pipeline (for advanced use).
func (s *MLService) Pipeline() *ModelPipeline {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pipeline
}

// Store returns the model store.
func (s *MLService) Store() *ModelStore {
	return s.store
}

// IsReady returns true if the pipeline has trained/loaded weights.
func (s *MLService) IsReady() bool {
	return s.pipeline.Trained()
}

// =========================================================================
// Inference
// =========================================================================

// Infer runs the full 4-stage neural pipeline on a client profile.
func (s *MLService) Infer(profile *profiles.ClientProfile, behavior []float64) *PipelineResult {
	s.inferCount.Add(1)
	return s.pipeline.Infer(profile, behavior)
}

// InferFromFeatures runs inference from a pre-extracted feature vector.
func (s *MLService) InferFromFeatures(fv *core.FeatureVector, behavior []float64) *PipelineResult {
	s.inferCount.Add(1)
	return s.pipeline.InferFromFeatures(fv, behavior)
}

// InferBatch runs batch inference for multiple profiles.
func (s *MLService) InferBatch(profs []*profiles.ClientProfile, behaviors [][]float64) []*PipelineResult {
	s.inferCount.Add(int64(len(profs)))
	return s.pipeline.InferBatch(profs, behaviors)
}

// =========================================================================
// Validation — check if a fingerprint looks realistic
// =========================================================================

// ValidationResult describes how realistic a fingerprint looks.
type ValidationResult struct {
	// Valid is true if the fingerprint passes all checks.
	Valid bool

	// ForgeryProb is the forgery probability [0,1].
	ForgeryProb float64

	// ForgeryType identifies the detected forgery category.
	ForgeryType ForgeryType

	// ConsistencyScore is the cross-layer consistency [0,1].
	ConsistencyScore float64

	// BrowserFamily is the identified browser.
	BrowserFamily string

	// Confidence is the classification confidence.
	Confidence float64

	// Suggestions lists improvements for invalid fingerprints.
	Suggestions []string
}

// Validate checks whether a fingerprint profile looks realistic according
// to the trained models. This is used by the generator to reject bad
// candidates and by TLS/HTTP modules to verify output quality.
func (s *MLService) Validate(profile *profiles.ClientProfile) *ValidationResult {
	result := s.pipeline.Infer(profile, nil)

	vr := &ValidationResult{
		ForgeryProb:      result.Forgery.ForgeryProb,
		ForgeryType:      result.Forgery.ForgeryType,
		ConsistencyScore: crossLayerConsistencyScore(result.CrossFeatures),
		BrowserFamily:    string(result.Browser.Family),
		Confidence:       result.Browser.Confidence,
	}

	vr.Valid = true
	if vr.ForgeryProb > s.config.ValidationForgeryThreshold {
		vr.Valid = false
		vr.Suggestions = append(vr.Suggestions,
			fmt.Sprintf("forgery probability %.2f exceeds threshold %.2f",
				vr.ForgeryProb, s.config.ValidationForgeryThreshold))
	}
	if vr.ConsistencyScore < s.config.ValidationConsistencyMin {
		vr.Valid = false
		vr.Suggestions = append(vr.Suggestions,
			fmt.Sprintf("cross-layer consistency %.2f below minimum %.2f",
				vr.ConsistencyScore, s.config.ValidationConsistencyMin))
	}
	if vr.Confidence < 0.3 {
		vr.Suggestions = append(vr.Suggestions,
			fmt.Sprintf("low classification confidence %.2f — fingerprint may be ambiguous", vr.Confidence))
	}

	return vr
}

// ValidateFeatures validates from a raw feature vector.
func (s *MLService) ValidateFeatures(features []float64) *ValidationResult {
	crossFeatures := ComputeCrossLayerFeatures(features)
	embedding := s.pipeline.encoder.EncodeSingle(features)
	browser := s.pipeline.classifier.ClassifySingle(embedding)
	forgery := s.pipeline.detector.DetectSingle(features, crossFeatures)

	vr := &ValidationResult{
		ForgeryProb:      forgery.ForgeryProb,
		ForgeryType:      forgery.ForgeryType,
		ConsistencyScore: crossLayerConsistencyScore(crossFeatures),
		BrowserFamily:    string(browser.Family),
		Confidence:       browser.Confidence,
		Valid:            true,
	}

	if vr.ForgeryProb > s.config.ValidationForgeryThreshold {
		vr.Valid = false
		vr.Suggestions = append(vr.Suggestions, "forgery probability too high")
	}
	if vr.ConsistencyScore < s.config.ValidationConsistencyMin {
		vr.Valid = false
		vr.Suggestions = append(vr.Suggestions, "cross-layer consistency too low")
	}
	return vr
}

// crossLayerConsistencyScore computes a 0–1 score from cross-layer features.
// Higher is more consistent (less suspicious).
func crossLayerConsistencyScore(crossFeatures []float64) float64 {
	if len(crossFeatures) < CrossLayerFeatureDim {
		return 0.5
	}
	// Cross-layer features are anomaly indicators; lower values = more consistent.
	sum := 0.0
	for _, v := range crossFeatures {
		sum += v
	}
	avg := sum / float64(len(crossFeatures))
	// Invert: high anomaly → low consistency.
	return math.Max(0, 1.0-avg)
}

// =========================================================================
// Feedback — online learning loop
// =========================================================================

// FeedbackSample is a single observation fed back into the learning system.
type FeedbackSample struct {
	Profile   *profiles.ClientProfile
	Features  []float64
	Label     string  // ground-truth browser family (optional)
	Reward    float64 // 0–1 reward signal
	Timestamp time.Time
}

// Feedback records an observation for online learning.
func (s *MLService) Feedback(sample *FeedbackSample) {
	s.feedbackCount.Add(1)
	if s.learner != nil {
		s.learner.AddSample(sample)
	}
}

// =========================================================================
// Generation — ML-guided fingerprint creation
// =========================================================================

// GenerateConfig controls ML-guided fingerprint generation.
type GenerateConfig struct {
	// TargetBrowser is the desired browser family.
	TargetBrowser string

	// TargetOS is the desired operating system.
	TargetOS string

	// MaxAttempts is how many candidates to try before giving up.
	MaxAttempts int

	// NoiseIntensity controls how much random variation to inject [0,1].
	NoiseIntensity float64
}

// DefaultGenerateConfig provides sensible generation defaults.
var DefaultGenerateConfig = &GenerateConfig{
	TargetBrowser:  "",
	TargetOS:       "",
	MaxAttempts:    10,
	NoiseIntensity: 0.05,
}

// GenerateResult is the output of ML-guided fingerprint generation.
type GenerateResult struct {
	// Profile is the generated fingerprint.
	Profile *profiles.ClientProfile

	// Validation is how the generated fingerprint scored.
	Validation *ValidationResult

	// Attempts is how many candidates were evaluated.
	Attempts int

	// SourceProfileID is the base profile that was mutated.
	SourceProfileID string
}

// Generate produces a fingerprint that passes forgery/consistency validation.
// It works by selecting a base profile, applying controlled mutations, and
// validating with the neural pipeline until a candidate passes.
func (s *MLService) Generate(config *GenerateConfig) (*GenerateResult, error) {
	if config == nil {
		config = DefaultGenerateConfig
	}
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = 10
	}

	// Select candidate base profiles.
	candidates := s.selectCandidates(config)
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no candidate profiles for browser=%q os=%q", config.TargetBrowser, config.TargetOS)
	}

	bestResult := &GenerateResult{Attempts: 0}
	bestScore := -1.0

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Pick a random base.
		base := candidates[rand.Intn(len(candidates))]
		mutated := s.mutateProfile(&base, config.NoiseIntensity)

		// Validate.
		vr := s.Validate(mutated)
		bestResult.Attempts = attempt + 1

		// Score: lower forgery + higher consistency + higher confidence.
		score := (1.0 - vr.ForgeryProb) * vr.ConsistencyScore * vr.Confidence
		if score > bestScore {
			bestScore = score
			bestResult.Profile = mutated
			bestResult.Validation = vr
			bestResult.SourceProfileID = base.ID
		}

		if vr.Valid {
			return bestResult, nil
		}
	}

	// Return the best candidate even if none passed.
	return bestResult, nil
}

// selectCandidates returns base profiles matching the generation config.
func (s *MLService) selectCandidates(config *GenerateConfig) []profiles.ClientProfile {
	all := profiles.GetAll()
	if config.TargetBrowser == "" && config.TargetOS == "" {
		return all
	}

	var filtered []profiles.ClientProfile
	for _, p := range all {
		browserMatch := config.TargetBrowser == "" ||
			string(p.BrowserType) == config.TargetBrowser
		osMatch := config.TargetOS == "" ||
			string(p.OS) == config.TargetOS
		if browserMatch && osMatch {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

// mutateProfile applies controlled noise to a profile clone for diversity.
func (s *MLService) mutateProfile(base *profiles.ClientProfile, intensity float64) *profiles.ClientProfile {
	// Deep-clone the profile (shallow copy + pointer fields).
	cloned := *base
	cloned.ID = fmt.Sprintf("%s_gen_%d", base.ID, time.Now().UnixNano())

	// Mutate TCP/IP parameters with small noise.
	if cloned.TCPIP != nil {
		tcp := *cloned.TCPIP
		if rand.Float64() < intensity*2 {
			delta := int16(float64(tcp.WindowSize) * (rand.Float64() - 0.5) * intensity)
			tcp.WindowSize = uint16(int16(tcp.WindowSize) + delta)
		}
		if rand.Float64() < intensity*2 {
			delta := int16(float64(tcp.MSS) * (rand.Float64() - 0.5) * intensity)
			tcp.MSS = uint16(int16(tcp.MSS) + delta)
		}
		cloned.TCPIP = &tcp
	}

	// Mutate HTTP/2 settings slightly.
	{
		h2 := cloned.HTTP2Settings
		if h2.InitialWindowSize > 0 && rand.Float64() < intensity {
			delta := int32(float64(h2.InitialWindowSize) * (rand.Float64() - 0.5) * intensity * 0.1)
			h2.InitialWindowSize = uint32(int32(h2.InitialWindowSize) + delta)
		}
		cloned.HTTP2Settings = h2
	}

	return &cloned
}

// =========================================================================
// Evolution — incremental model improvement
// =========================================================================

// Evolve triggers incremental training with the current profile registry.
// Returns the training metrics and an error if the evolution fails.
func (s *MLService) Evolve(registry *profiles.ProfileRegistry) (*TrainingMetrics, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.evolveCount.Add(1)

	trainer := NewNeuralTrainer(s.pipeline, nil)

	if s.store != nil {
		_, metrics, err := trainer.EvolveAndSave(registry, s.store, nil)
		if err != nil {
			return nil, fmt.Errorf("ml service: evolve: %w", err)
		}
		return metrics, nil
	}

	// No store — evolve in-memory only.
	metrics, err := trainer.Evolve(registry, nil)
	if err != nil {
		return nil, fmt.Errorf("ml service: evolve: %w", err)
	}
	return metrics, nil
}

// Train runs full training from scratch with the given profile registry.
func (s *MLService) Train(registry *profiles.ProfileRegistry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	trainer := NewNeuralTrainer(s.pipeline, nil)

	if s.store != nil {
		_, err := trainer.TrainAndSave(registry, s.store)
		if err != nil {
			return fmt.Errorf("ml service: train: %w", err)
		}
		return nil
	}

	if err := trainer.TrainFromProfiles(registry); err != nil {
		return fmt.Errorf("ml service: train: %w", err)
	}
	return nil
}

// =========================================================================
// Stats
// =========================================================================

// ServiceStats holds ML service statistics.
type ServiceStats struct {
	InferCount    int64
	FeedbackCount int64
	EvolveCount   int64
	ModelReady    bool
	ModelVersions int
	LearnerStats  *OnlineLearnerStats
}

// Stats returns current service statistics.
func (s *MLService) Stats() *ServiceStats {
	st := &ServiceStats{
		InferCount:    s.inferCount.Load(),
		FeedbackCount: s.feedbackCount.Load(),
		EvolveCount:   s.evolveCount.Load(),
		ModelReady:    s.IsReady(),
	}
	if s.store != nil {
		st.ModelVersions = s.store.VersionCount()
	}
	if s.learner != nil {
		st.LearnerStats = s.learner.Stats()
	}
	return st
}
