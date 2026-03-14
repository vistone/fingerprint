// Package generator provides ML-driven browser fingerprint generation.
//
// The generator uses the trained neural pipeline to produce realistic
// fingerprints that pass forgery detection and cross-layer consistency
// checks. It supports:
//
//   - Random generation: pick a random browser profile and mutate it
//   - Targeted generation: specify browser type and OS
//   - Batch generation: produce multiple diverse fingerprints
//   - ML-validated generation: iteratively refine until quality passes
//   - Noise-injected generation: add controlled noise for privacy
//
// Usage:
//
//	gen := generator.NewSmartGenerator(mlService, nil)
//	result, err := gen.Generate(&generator.SmartGenerateConfig{
//	    TargetBrowser: "chrome",
//	    TargetOS:      "windows",
//	})
package generator

import (
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vistone/fingerprint/modules/ml"
	"github.com/vistone/fingerprint/modules/profiles"
)

// =========================================================================
// SmartGenerator — ML-driven fingerprint generator
// =========================================================================

// SmartGeneratorConfig configures the ML-driven generator.
type SmartGeneratorConfig struct {
	// MaxRetries is the maximum number of generation attempts.
	MaxRetries int

	// DefaultNoiseIntensity is the noise level for mutations [0,1].
	DefaultNoiseIntensity float64

	// QualityThreshold: generated fingerprints must achieve this quality
	// score (composite of forgery, consistency, confidence) in [0,1].
	QualityThreshold float64

	// EnableCaching caches recently generated profiles for dedup.
	EnableCaching bool

	// CacheSize is the max number of cached profiles.
	CacheSize int
}

// DefaultSmartGeneratorConfig provides sensible defaults.
var DefaultSmartGeneratorConfig = &SmartGeneratorConfig{
	MaxRetries:            20,
	DefaultNoiseIntensity: 0.05,
	QualityThreshold:      0.6,
	EnableCaching:         true,
	CacheSize:             100,
}

// SmartGenerator uses ML models to generate and validate fingerprints.
type SmartGenerator struct {
	mlService *ml.MLService
	config    *SmartGeneratorConfig
	mu        sync.RWMutex

	// Cache of recently generated profiles for dedup.
	cache []GeneratedProfile
	head  int

	// Statistics.
	totalGenerated atomic.Int64
	totalAttempts  atomic.Int64
	totalRejected  atomic.Int64
}

// GeneratedProfile is a cached generation result.
type GeneratedProfile struct {
	Profile     *profiles.ClientProfile
	Quality     float64
	Source      string
	GeneratedAt time.Time
}

// NewSmartGenerator creates a new ML-driven generator.
func NewSmartGenerator(mlService *ml.MLService, config *SmartGeneratorConfig) *SmartGenerator {
	if config == nil {
		config = DefaultSmartGeneratorConfig
	}
	cacheSize := config.CacheSize
	if cacheSize <= 0 {
		cacheSize = 100
	}

	return &SmartGenerator{
		mlService: mlService,
		config:    config,
		cache:     make([]GeneratedProfile, cacheSize),
	}
}

// SmartGenerateConfig controls a single generation request.
type SmartGenerateConfig struct {
	// TargetBrowser is the desired browser family (e.g., "chrome", "firefox").
	TargetBrowser string

	// TargetOS is the desired operating system (e.g., "windows", "macos").
	TargetOS string

	// NoiseIntensity overrides the default noise level [0,1].
	NoiseIntensity float64

	// MaxRetries overrides the default max retries.
	MaxRetries int

	// QualityThreshold overrides the default quality threshold.
	QualityThreshold float64
}

// SmartGenerateResult is the output of ML-driven generation.
type SmartGenerateResult struct {
	// Profile is the generated fingerprint.
	Profile *profiles.ClientProfile

	// QualityScore is the composite quality score [0,1].
	QualityScore float64

	// Validation contains detailed ML validation results.
	Validation *ml.ValidationResult

	// Attempts is how many candidates were evaluated.
	Attempts int

	// SourceProfileID is the base profile that was mutated.
	SourceProfileID string

	// GenerationTime is total generation duration.
	GenerationTime time.Duration
}

// Generate produces a single ML-validated fingerprint.
func (g *SmartGenerator) Generate(config *SmartGenerateConfig) (*SmartGenerateResult, error) {
	start := time.Now()

	if config == nil {
		config = &SmartGenerateConfig{}
	}

	maxRetries := config.MaxRetries
	if maxRetries <= 0 {
		maxRetries = g.config.MaxRetries
	}
	noiseIntensity := config.NoiseIntensity
	if noiseIntensity <= 0 {
		noiseIntensity = g.config.DefaultNoiseIntensity
	}
	qualThreshold := config.QualityThreshold
	if qualThreshold <= 0 {
		qualThreshold = g.config.QualityThreshold
	}

	// Select candidate base profiles.
	candidates := g.selectCandidates(config.TargetBrowser, config.TargetOS)
	if len(candidates) == 0 {
		return nil, ErrNoProfilesAvailable
	}

	best := &SmartGenerateResult{Attempts: 0}
	bestScore := -1.0

	for attempt := 0; attempt < maxRetries; attempt++ {
		g.totalAttempts.Add(1)

		// Pick a random base and mutate.
		base := candidates[rand.Intn(len(candidates))]
		mutated := g.mutateProfile(&base, noiseIntensity)

		// ML validation.
		vr := g.mlService.Validate(mutated)
		best.Attempts = attempt + 1

		// Composite quality score.
		score := computeQualityScore(vr)

		if score > bestScore {
			bestScore = score
			best.Profile = mutated
			best.QualityScore = score
			best.Validation = vr
			best.SourceProfileID = base.ID
		}

		if score >= qualThreshold && vr.Valid {
			break
		}

		g.totalRejected.Add(1)
	}

	best.GenerationTime = time.Since(start)
	g.totalGenerated.Add(1)

	// Cache the result.
	if g.config.EnableCaching && best.Profile != nil {
		g.addToCache(GeneratedProfile{
			Profile:     best.Profile,
			Quality:     best.QualityScore,
			Source:      best.SourceProfileID,
			GeneratedAt: time.Now(),
		})
	}

	return best, nil
}

// GenerateBatch produces multiple diverse fingerprints.
func (g *SmartGenerator) GenerateBatch(n int, config *SmartGenerateConfig) ([]*SmartGenerateResult, error) {
	if n <= 0 {
		n = 1
	}

	results := make([]*SmartGenerateResult, 0, n)
	usedSources := make(map[string]bool)

	for i := 0; i < n; i++ {
		result, err := g.Generate(config)
		if err != nil {
			return results, err
		}

		// Try to diversify — avoid the same source profile twice.
		if usedSources[result.SourceProfileID] && i < n*2 {
			i-- // retry
			continue
		}
		usedSources[result.SourceProfileID] = true
		results = append(results, result)
	}

	return results, nil
}

// GenerateForBrowser is a convenience method for generating by browser.
func (g *SmartGenerator) GenerateForBrowser(browser string) (*SmartGenerateResult, error) {
	return g.Generate(&SmartGenerateConfig{TargetBrowser: browser})
}

// GenerateForOS is a convenience method for generating by OS.
func (g *SmartGenerator) GenerateForOS(os string) (*SmartGenerateResult, error) {
	return g.Generate(&SmartGenerateConfig{TargetOS: os})
}

// =========================================================================
// Internal helpers
// =========================================================================

// selectCandidates returns base profiles matching the criteria.
func (g *SmartGenerator) selectCandidates(browser, os string) []profiles.ClientProfile {
	all := profiles.GetAll()
	if browser == "" && os == "" {
		return all
	}

	browserLower := strings.ToLower(browser)
	osLower := strings.ToLower(os)

	var filtered []profiles.ClientProfile
	for _, p := range all {
		browserMatch := browserLower == "" ||
			strings.Contains(strings.ToLower(string(p.BrowserType)), browserLower)
		osMatch := osLower == "" ||
			strings.Contains(strings.ToLower(string(p.OS)), osLower)
		if browserMatch && osMatch {
			filtered = append(filtered, p)
		}
	}

	// Fallback to all if filter too strict.
	if len(filtered) == 0 {
		return all
	}
	return filtered
}

// mutateProfile applies controlled mutations to create a variant.
