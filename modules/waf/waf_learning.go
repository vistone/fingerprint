package waf

import (
	"log/slog"
	"sync"
	"time"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/ml"
	"github.com/vistone/fingerprint/modules/profiles"
)

// LearningPipeline bridges WAF detection results to the ML learning system.
// It runs ML inference during request analysis and feeds detection outcomes
// back to the OnlineLearner for continuous model improvement.
type LearningPipeline struct {
	mlService *ml.MLService
	stats     *LearningStats
	mu        sync.RWMutex
	logger    *slog.Logger
}

// LearningStats tracks learning pipeline statistics.
type LearningStats struct {
	SamplesProcessed int64
	DetectionsFed    int64
	InferencesRun    int64
	LastFeedbackTime time.Time
}

// NewLearningPipeline creates a new WAF learning pipeline.
func NewLearningPipeline(svc *ml.MLService) *LearningPipeline {
	return &LearningPipeline{
		mlService: svc,
		stats:     &LearningStats{},
		logger:    slog.Default().With("component", "waf-learning"),
	}
}

// RunInference runs ML inference on a fingerprint profile extracted from
// a request, returning a risk score adjustment based on forgery probability.
func (lp *LearningPipeline) RunInference(
	profile *profiles.ClientProfile,
	riskFactors []core.RiskFactor,
) float64 {
	if lp.mlService == nil || !lp.mlService.IsReady() {
		return 0
	}

	lp.mu.Lock()
	lp.stats.InferencesRun++
	lp.mu.Unlock()

	result := lp.mlService.Infer(profile, nil)
	if result == nil {
		return 0
	}

	// Use forgery probability as risk adjustment (scaled to 0.3 max).
	adjustment := result.Forgery.ForgeryProb * 0.3

	if result.Forgery.ForgeryProb > 0.7 {
		lp.logger.Debug("ml detected high forgery probability",
			"forgery_prob", result.Forgery.ForgeryProb,
			"browser", result.Browser.Family)
	}

	return adjustment
}

// FeedDetection feeds a WAF detection result back to the ML system.
// This allows the model to learn which fingerprint patterns are associated
// with detected or blocked requests.
func (lp *LearningPipeline) FeedDetection(feedback *ml.WAFDetectionFeedback) {
	if lp.mlService == nil || feedback == nil {
		return
	}

	lp.mu.Lock()
	lp.stats.DetectionsFed++
	lp.stats.SamplesProcessed++
	lp.stats.LastFeedbackTime = time.Now()
	lp.mu.Unlock()

	sample := feedback.ToFeedbackSample()
	lp.mlService.Feedback(sample)

	lp.logger.Debug("waf detection fed to ml",
		"risk_score", feedback.RiskScore,
		"layers", feedback.DetectionLayers,
		"blocked", feedback.Blocked)
}

// LearningStats returns current learning pipeline statistics.
func (lp *LearningPipeline) LearningStats() *LearningStats {
	lp.mu.RLock()
	defer lp.mu.RUnlock()
	s := *lp.stats
	return &s
}
