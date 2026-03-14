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
	"log/slog"
	"math"
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
// After recording, it checks for accuracy drift and triggers auto-evolution
// if the learner has detected a significant accuracy drop.
func (s *MLService) Feedback(sample *FeedbackSample) {
	s.feedbackCount.Add(1)
	if s.learner != nil {
		s.learner.AddSample(sample)
		// Auto-evolve when drift is detected
		if s.learner.DriftDetected() {
			slog.Info("ml: accuracy drift detected, triggering auto-evolution")
			if _, err := s.Evolve(nil); err != nil {
				slog.Warn("ml: auto-evolution failed", "error", err)
			}
		}
	}
}

// Learner returns the online learner instance (nil if not configured).
func (s *MLService) Learner() *OnlineLearner {
	return s.learner
}

// =========================================================================
// Generation — ML-guided fingerprint creation
// =========================================================================

// GenerateConfig controls ML-guided fingerprint generation.
