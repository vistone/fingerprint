package crawler

import (
	"fmt"
	"log/slog"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/vistone/fingerprint/modules/ml"
	"github.com/vistone/fingerprint/modules/profiles"
)

// CrawlerMLAdapter bridges the crawler with the ML service for
// adaptive profile selection and crawl-result feedback learning.
type CrawlerMLAdapter struct {
	mlService    *ml.MLService
	profileStats map[string]*profilePerf
	mu           sync.RWMutex
	logger       *slog.Logger
}

// profilePerf tracks per-profile success/failure statistics.
type profilePerf struct {
	ProfileID   string
	Successes   int64
	Failures    int64
	Blocks      int64
	TotalReward float64
	LastUsed    time.Time
}

// NewCrawlerMLAdapter creates a new ML adapter for the crawler.
func NewCrawlerMLAdapter(svc *ml.MLService) *CrawlerMLAdapter {
	return &CrawlerMLAdapter{
		mlService:    svc,
		profileStats: make(map[string]*profilePerf),
		logger:       slog.Default().With("component", "crawler-ml"),
	}
}

// RecordResult records a crawl result for ML online learning.
func (a *CrawlerMLAdapter) RecordResult(result *CrawlResult) {
	if result == nil || result.Fingerprint == nil {
		return
	}

	pid := result.Fingerprint.ID

	a.mu.Lock()
	a.updateProfileStats(pid, result)
	a.mu.Unlock()

	a.feedToML(result, pid)
}

// updateProfileStats updates local performance statistics for a profile.
func (a *CrawlerMLAdapter) updateProfileStats(pid string, result *CrawlResult) {
	perf, ok := a.profileStats[pid]
	if !ok {
		perf = &profilePerf{ProfileID: pid}
		a.profileStats[pid] = perf
	}
	perf.LastUsed = time.Now()

	switch {
	case result.Blocked:
		perf.Blocks++
	case result.StatusCode >= 200 && result.StatusCode < 300:
		perf.Successes++
		perf.TotalReward += 1.0
	default:
		perf.Failures++
		perf.TotalReward += 0.3
	}
}

// feedToML sends a crawl result as feedback to the ML service.
func (a *CrawlerMLAdapter) feedToML(result *CrawlResult, pid string) {
	if a.mlService == nil {
		return
	}

	feedback := &ml.CrawlerFeedback{
		ProfileID:     pid,
		Profile:       result.Fingerprint,
		URL:           result.URL,
		Success:       result.StatusCode >= 200 && result.StatusCode < 300,
		Blocked:       result.Blocked,
		BlockReason:   result.BlockReason,
		DetectionInfo: result.DetectionInfo,
		Duration:      result.Duration,
		Timestamp:     result.Timestamp,
	}

	sample := feedback.ToFeedbackSample()
	a.mlService.Feedback(sample)

	a.logger.Debug("ml feedback sent",
		"profile", pid,
		"reward", sample.Reward,
		"blocked", result.Blocked)
}

// SelectBestProfile selects the best-performing profile using UCB1
// (Upper Confidence Bound) for exploration/exploitation balance.
func (a *CrawlerMLAdapter) SelectBestProfile(pool []*profiles.ClientProfile) *profiles.ClientProfile {
	if len(pool) == 0 {
		return nil
	}
	if len(pool) == 1 {
		return pool[0]
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	totalPulls := a.totalPullCount(pool)

	// If insufficient data, pick an unexplored profile first.
	if totalPulls < int64(len(pool)) {
		if p := a.findUnexplored(pool); p != nil {
			return p
		}
	}

	return a.selectByUCB1(pool, totalPulls)
}

// totalPullCount sums total attempts across all candidate profiles.
func (a *CrawlerMLAdapter) totalPullCount(pool []*profiles.ClientProfile) int64 {
	var total int64
	for _, p := range pool {
		if perf, ok := a.profileStats[p.ID]; ok {
			total += perf.Successes + perf.Failures + perf.Blocks
		}
	}
	return total
}

// findUnexplored returns the first profile with no recorded stats.
func (a *CrawlerMLAdapter) findUnexplored(pool []*profiles.ClientProfile) *profiles.ClientProfile {
	for _, p := range pool {
		if _, ok := a.profileStats[p.ID]; !ok {
			return p
		}
	}
	return nil
}

// selectByUCB1 picks the profile with the highest UCB1 score.
func (a *CrawlerMLAdapter) selectByUCB1(pool []*profiles.ClientProfile, totalPulls int64) *profiles.ClientProfile {
	type scored struct {
		profile *profiles.ClientProfile
		score   float64
	}

	scores := make([]scored, 0, len(pool))
	for _, p := range pool {
		perf, ok := a.profileStats[p.ID]
		if !ok {
			scores = append(scores, scored{p, math.MaxFloat64})
			continue
		}
		n := float64(perf.Successes + perf.Failures + perf.Blocks)
		if n == 0 {
			scores = append(scores, scored{p, math.MaxFloat64})
			continue
		}
		avgReward := perf.TotalReward / n
		exploration := math.Sqrt(2.0 * math.Log(float64(totalPulls)+1) / n)
		scores = append(scores, scored{p, avgReward + exploration})
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})
	return scores[0].profile
}

// GenerateProfile uses the ML service to generate a new fingerprint profile.
func (a *CrawlerMLAdapter) GenerateProfile(cfg *ml.GenerateConfig) (*profiles.ClientProfile, error) {
	if a.mlService == nil {
		return nil, fmt.Errorf("ml service not available")
	}

	result, err := a.mlService.Generate(cfg)
	if err != nil {
		return nil, fmt.Errorf("generate profile: %w", err)
	}
	if result.Profile == nil {
		return nil, fmt.Errorf("generation produced no valid profile")
	}

	a.logger.Info("generated ml profile",
		"id", result.Profile.ID,
		"source", result.SourceProfileID,
		"attempts", result.Attempts,
		"valid", result.Validation.Valid)

	return result.Profile, nil
}

// ProfileStats returns current performance statistics for all tracked profiles.
func (a *CrawlerMLAdapter) ProfileStats() map[string]*profilePerf {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make(map[string]*profilePerf, len(a.profileStats))
	for k, v := range a.profileStats {
		cp := *v
		out[k] = &cp
	}
	return out
}
