// Package fingerprint — ml_api.go provides top-level ML convenience APIs
// that expose the full neural pipeline without requiring users to import
// the ml or gateway packages directly.
//
// This bridges the gap between the low-level ML module and end-users
// who just want smart fingerprint operations:
//
//   - MLAnalyze: deep neural analysis of any fingerprint
//   - MLGenerate: produce ML-validated fingerprints
//   - MLValidate: check if a fingerprint looks realistic
//   - MLService: access the centralized AI brain
//
// Usage:
//
//	svc := fingerprint.NewMLService(nil)
//	result := svc.Analyze(profile)
//	gen, _ := svc.Generate("chrome", "windows")
//	ok := svc.Validate(profile)
package fingerprint

import (
	"github.com/vistone/fingerprint/modules/generator"
	"github.com/vistone/fingerprint/modules/ml"
	"github.com/vistone/fingerprint/modules/profiles"
)

// =========================================================================
// MLFacade — top-level ML entry point
// =========================================================================

// MLFacade provides a simplified interface to the entire ML system.
type MLFacade struct {
	service   *ml.MLService
	generator *generator.SmartGenerator
	evolution *ml.ProfileEvolutionEngine
}

// MLFacadeConfig configures the top-level ML facade.
type MLFacadeConfig struct {
	// ModelStorePath is the directory for model snapshots.
	ModelStorePath string

	// AutoLoadModel loads the latest model on startup.
	AutoLoadModel bool

	// GeneratorConfig overrides generator defaults.
	GeneratorConfig *generator.SmartGeneratorConfig

	// EvolutionConfig overrides evolution defaults.
	EvolutionConfig *ml.EvolutionConfig
}

// DefaultMLFacadeConfig provides sensible defaults.
var DefaultMLFacadeConfig = &MLFacadeConfig{
	ModelStorePath: "./models",
	AutoLoadModel:  true,
}

// NewMLFacade creates a new top-level ML facade.
// This is the recommended way to access ML capabilities.
func NewMLFacade(config *MLFacadeConfig) (*MLFacade, error) {
	if config == nil {
		config = DefaultMLFacadeConfig
	}

	svcConfig := &ml.ServiceConfig{
		ModelStorePath: config.ModelStorePath,
		AutoLoadLatest: config.AutoLoadModel,
	}
	svc, err := ml.NewMLService(svcConfig)
	if err != nil {
		return nil, err
	}

	f := &MLFacade{
		service:   svc,
		generator: generator.NewSmartGenerator(svc, config.GeneratorConfig),
		evolution: ml.NewProfileEvolutionEngine(config.EvolutionConfig),
	}

	return f, nil
}

// Service returns the underlying ML service for advanced usage.
func (f *MLFacade) Service() *ml.MLService {
	return f.service
}

// Generator returns the smart generator for advanced usage.
func (f *MLFacade) Generator() *generator.SmartGenerator {
	return f.generator
}

// Evolution returns the profile evolution engine.
func (f *MLFacade) Evolution() *ml.ProfileEvolutionEngine {
	return f.evolution
}

// =========================================================================
// Analysis APIs
// =========================================================================

// MLAnalyzeResult is a comprehensive ML analysis result.
type MLAnalyzeResult struct {
	// Browser identification
	BrowserFamily     string
	BrowserConfidence float64

	// Forgery detection
	ForgeryProb float64
	ForgeryType string
	IsForged    bool

	// Threat assessment
	ThreatClass  string
	ThreatAction string

	// Cross-layer consistency
	ConsistencyScore float64

	// Embedding (for similarity search)
	Embedding []float64

	// Raw pipeline result (for advanced usage)
	Raw *ml.PipelineResult
}

// MLAnalyze performs deep neural analysis on a fingerprint profile.
func (f *MLFacade) MLAnalyze(profile *profiles.ClientProfile) *MLAnalyzeResult {
	result := f.service.Infer(profile, nil)
	return pipelineToAnalyzeResult(result)
}

// MLAnalyzeWithBehavior performs analysis with behavioral context.
func (f *MLFacade) MLAnalyzeWithBehavior(profile *profiles.ClientProfile, behavior []float64) *MLAnalyzeResult {
	result := f.service.Infer(profile, behavior)
	return pipelineToAnalyzeResult(result)
}

// MLAnalyzeBatch analyses multiple profiles in batch.
func (f *MLFacade) MLAnalyzeBatch(profs []*profiles.ClientProfile) []*MLAnalyzeResult {
	results := f.service.InferBatch(profs, nil)
	analyzed := make([]*MLAnalyzeResult, len(results))
	for i, r := range results {
		analyzed[i] = pipelineToAnalyzeResult(r)
	}
	return analyzed
}

func pipelineToAnalyzeResult(r *ml.PipelineResult) *MLAnalyzeResult {
	if r == nil {
		return &MLAnalyzeResult{}
	}
	return &MLAnalyzeResult{
		BrowserFamily:     string(r.Browser.Family),
		BrowserConfidence: r.Browser.Confidence,
		ForgeryProb:       r.Forgery.ForgeryProb,
		ForgeryType:       r.Forgery.ForgeryType.String(),
		IsForged:          r.Forgery.ForgeryProb > 0.5,
		ThreatClass:       r.Threat.ThreatClass.String(),
		ThreatAction:      r.Threat.Action.String(),
		ConsistencyScore:  1.0 - avg(r.CrossFeatures),
		Embedding:         r.Embedding,
		Raw:               r,
	}
}

func avg(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

// =========================================================================
// Generation APIs
// =========================================================================

// MLGenerate produces a ML-validated fingerprint for the target browser/OS.
func (f *MLFacade) MLGenerate(browser, os string) (*generator.SmartGenerateResult, error) {
	return f.generator.Generate(&generator.SmartGenerateConfig{
		TargetBrowser: browser,
		TargetOS:      os,
	})
}

// MLGenerateRandom produces a random ML-validated fingerprint.
func (f *MLFacade) MLGenerateRandom() (*generator.SmartGenerateResult, error) {
	return f.generator.Generate(nil)
}

// MLGenerateBatch produces multiple diverse ML-validated fingerprints.
func (f *MLFacade) MLGenerateBatch(n int, browser, os string) ([]*generator.SmartGenerateResult, error) {
	return f.generator.GenerateBatch(n, &generator.SmartGenerateConfig{
		TargetBrowser: browser,
		TargetOS:      os,
	})
}

// =========================================================================
// Validation APIs
// =========================================================================

// MLValidate checks whether a fingerprint looks realistic.
func (f *MLFacade) MLValidate(profile *profiles.ClientProfile) *ml.ValidationResult {
	return f.service.Validate(profile)
}

// MLValidateAll validates all registered profiles and returns rankings.
func (f *MLFacade) MLValidateAll() []generator.ProfileQualityRank {
	return f.generator.RankProfiles()
}

// =========================================================================
// Learning & Evolution APIs
// =========================================================================

// MLFeedback records an observation for online learning.
func (f *MLFacade) MLFeedback(profile *profiles.ClientProfile, reward float64) {
	f.service.Feedback(&ml.FeedbackSample{
		Profile: profile,
		Reward:  reward,
	})
}

// MLEvolve triggers incremental model evolution.
func (f *MLFacade) MLEvolve(registry *profiles.ProfileRegistry) (*ml.TrainingMetrics, error) {
	return f.service.Evolve(registry)
}

// MLTrain runs full training from scratch.
func (f *MLFacade) MLTrain(registry *profiles.ProfileRegistry) error {
	return f.service.Train(registry)
}

// MLCheckHealth evaluates the profile ecosystem health.
func (f *MLFacade) MLCheckHealth() *ml.EvolutionHealthReport {
	return f.evolution.CheckHealth()
}

// =========================================================================
// Similarity & Embedding APIs
// =========================================================================

// MLFindSimilar finds profiles most similar to a target using embeddings.
func (f *MLFacade) MLFindSimilar(target *profiles.ClientProfile, topN int) []generator.SimilarProfile {
	return f.generator.FindSimilarProfiles(target, topN)
}

// MLEmbedding computes the 32-dim embedding for a profile.
func (f *MLFacade) MLEmbedding(profile *profiles.ClientProfile) []float64 {
	result := f.service.Infer(profile, nil)
	return result.Embedding
}

// =========================================================================
// Stats
// =========================================================================

// MLStats returns combined ML statistics.
type MLStats struct {
	Service   *ml.ServiceStats
	Generator *generator.GeneratorStats
}

// MLStats returns combined ML system statistics.
func (f *MLFacade) MLStats() *MLStats {
	return &MLStats{
		Service:   f.service.Stats(),
		Generator: f.generator.Stats(),
	}
}
