package generator

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/ml"
	"github.com/vistone/fingerprint/modules/profiles"
)

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
