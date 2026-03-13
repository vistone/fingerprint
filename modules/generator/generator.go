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
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vistone/fingerprint/modules/core"
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
func (g *SmartGenerator) mutateProfile(base *profiles.ClientProfile, intensity float64) *profiles.ClientProfile {
	cloned := *base
	cloned.ID = fmt.Sprintf("%s_gen_%d", base.ID, time.Now().UnixNano())

	// Mutate TCP/IP parameters.
	if cloned.TCPIP != nil {
		tcp := *cloned.TCPIP
		if rand.Float64() < intensity*3 {
			delta := int16(float64(tcp.WindowSize) * (rand.Float64() - 0.5) * intensity)
			tcp.WindowSize = uint16(int16(tcp.WindowSize) + delta)
			if tcp.WindowSize < 1024 {
				tcp.WindowSize = 1024
			}
		}
		if rand.Float64() < intensity*3 {
			delta := int16(float64(tcp.MSS) * (rand.Float64() - 0.5) * intensity)
			tcp.MSS = uint16(int16(tcp.MSS) + delta)
			if tcp.MSS < 536 {
				tcp.MSS = 536
			}
		}
		cloned.TCPIP = &tcp
	}

	// Mutate HTTP/2 settings.
	{
		h2 := cloned.HTTP2Settings
		if h2.InitialWindowSize > 0 && rand.Float64() < intensity*2 {
			delta := int32(float64(h2.InitialWindowSize) * (rand.Float64() - 0.5) * intensity * 0.1)
			h2.InitialWindowSize = uint32(int32(h2.InitialWindowSize) + delta)
		}
		if h2.MaxConcurrentStreams > 0 && rand.Float64() < intensity {
			delta := int32(float64(h2.MaxConcurrentStreams) * (rand.Float64() - 0.5) * intensity * 0.1)
			h2.MaxConcurrentStreams = uint32(int32(h2.MaxConcurrentStreams) + delta)
		}
		cloned.HTTP2Settings = h2
	}

	// Mutate connection flow.
	if cloned.ConnectionFlow > 0 && rand.Float64() < intensity {
		delta := int32(float64(cloned.ConnectionFlow) * (rand.Float64() - 0.5) * intensity * 0.05)
		cloned.ConnectionFlow = uint32(int32(cloned.ConnectionFlow) + delta)
	}

	return &cloned
}

// computeQualityScore computes a composite quality score from validation.
func computeQualityScore(vr *ml.ValidationResult) float64 {
	if vr == nil {
		return 0
	}
	// Weighted composite:
	// 40% forgery (lower is better)
	// 30% consistency (higher is better)
	// 30% confidence (higher is better)
	forgeryScore := 1.0 - vr.ForgeryProb
	score := 0.4*forgeryScore + 0.3*vr.ConsistencyScore + 0.3*vr.Confidence
	return score
}

// addToCache adds a generated profile to the ring buffer cache.
func (g *SmartGenerator) addToCache(gp GeneratedProfile) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.cache[g.head] = gp
	g.head = (g.head + 1) % len(g.cache)
}

// =========================================================================
// Statistics
// =========================================================================

// GeneratorStats holds generator statistics.
type GeneratorStats struct {
	TotalGenerated    int64
	TotalAttempts     int64
	TotalRejected     int64
	AvgAttemptsPerGen float64
	SuccessRate       float64
}

// Stats returns current generator statistics.
func (g *SmartGenerator) Stats() *GeneratorStats {
	generated := g.totalGenerated.Load()
	attempts := g.totalAttempts.Load()
	rejected := g.totalRejected.Load()

	stats := &GeneratorStats{
		TotalGenerated: generated,
		TotalAttempts:  attempts,
		TotalRejected:  rejected,
	}

	if generated > 0 {
		stats.AvgAttemptsPerGen = float64(attempts) / float64(generated)
		stats.SuccessRate = float64(generated) / float64(generated+rejected)
	}

	return stats
}

// =========================================================================
// Profile quality ranking
// =========================================================================

// RankProfiles evaluates and ranks all profiles by ML quality score.
func (g *SmartGenerator) RankProfiles() []ProfileQualityRank {
	all := profiles.GetAll()
	ranks := make([]ProfileQualityRank, 0, len(all))

	for i := range all {
		vr := g.mlService.Validate(&all[i])
		score := computeQualityScore(vr)
		ranks = append(ranks, ProfileQualityRank{
			ProfileID:    all[i].ID,
			BrowserType:  string(all[i].BrowserType),
			QualityScore: score,
			ForgeryProb:  vr.ForgeryProb,
			Consistency:  vr.ConsistencyScore,
			Confidence:   vr.Confidence,
		})
	}

	sort.Slice(ranks, func(i, j int) bool {
		return ranks[i].QualityScore > ranks[j].QualityScore
	})

	return ranks
}

// ProfileQualityRank ranks a profile by its ML quality metrics.
type ProfileQualityRank struct {
	ProfileID    string
	BrowserType  string
	QualityScore float64
	ForgeryProb  float64
	Consistency  float64
	Confidence   float64
}

// =========================================================================
// Feature vector generation helpers
// =========================================================================

// GenerateFeatureVector produces only the feature vector (no full profile)
// for a given browser type. Useful for embedding similarity searches.
func (g *SmartGenerator) GenerateFeatureVector(browser string) ([]float64, error) {
	result, err := g.Generate(&SmartGenerateConfig{TargetBrowser: browser})
	if err != nil {
		return nil, err
	}
	return ml.EncodeFingerprint(result.Profile), nil
}

// GenerateEmbedding produces a 32-dim embedding for a target browser.
func (g *SmartGenerator) GenerateEmbedding(browser string) ([]float64, error) {
	result, err := g.Generate(&SmartGenerateConfig{TargetBrowser: browser})
	if err != nil {
		return nil, err
	}
	inferResult := g.mlService.Infer(result.Profile, nil)
	return inferResult.Embedding, nil
}

// FindSimilarProfiles finds profiles most similar to a given profile
// using embedding distance.
func (g *SmartGenerator) FindSimilarProfiles(target *profiles.ClientProfile, topN int) []SimilarProfile {
	targetResult := g.mlService.Infer(target, nil)
	targetEmb := targetResult.Embedding

	all := profiles.GetAll()
	type ranked struct {
		profile  profiles.ClientProfile
		distance float64
	}
	var items []ranked

	for i := range all {
		result := g.mlService.Infer(&all[i], nil)
		dist := embeddingDistance(targetEmb, result.Embedding)
		items = append(items, ranked{profile: all[i], distance: dist})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].distance < items[j].distance
	})

	if topN > len(items) {
		topN = len(items)
	}

	results := make([]SimilarProfile, topN)
	for i := 0; i < topN; i++ {
		results[i] = SimilarProfile{
			Profile:    items[i].profile,
			Distance:   items[i].distance,
			Similarity: 1.0 / (1.0 + items[i].distance),
		}
	}
	return results
}

// SimilarProfile represents a profile similar to a target.
type SimilarProfile struct {
	Profile    profiles.ClientProfile
	Distance   float64
	Similarity float64
}

// embeddingDistance computes L2 distance between two embeddings.
func embeddingDistance(a, b []float64) float64 {
	if len(a) != len(b) {
		return 1e9
	}
	sum := 0.0
	for i := range a {
		d := a[i] - b[i]
		sum += d * d
	}
	if sum < 0 {
		return 0
	}
	return sum // return squared distance for efficiency
}

// Unexport core and ml for the compiler.
var _ = core.BrowserChrome
