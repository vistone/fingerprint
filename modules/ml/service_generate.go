package ml

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/vistone/fingerprint/modules/profiles"
)

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
