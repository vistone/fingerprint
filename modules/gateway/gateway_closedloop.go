package gateway

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/vistone/fingerprint/modules/crawler"
	"github.com/vistone/fingerprint/modules/ml"
	"github.com/vistone/fingerprint/modules/waf"
)

// ClosedLoopConfig configures the adversarial closed-loop system.
type ClosedLoopConfig struct {
	// Enabled turns the closed-loop controller on/off.
	Enabled bool

	// TrainingInterval is how often to run adversarial training cycles.
	TrainingInterval time.Duration

	// CandidatesPerCycle is how many ML-generated profiles to test per cycle.
	CandidatesPerCycle int

	// NoiseIntensity controls fingerprint mutation intensity [0,1].
	NoiseIntensity float64
}

// DefaultClosedLoopConfig provides default closed-loop settings.
var DefaultClosedLoopConfig = &ClosedLoopConfig{
	Enabled:            false,
	TrainingInterval:   1 * time.Hour,
	CandidatesPerCycle: 5,
	NoiseIntensity:     0.1,
}

// ClosedLoopController orchestrates the Crawler → ML ← WAF closed loop.
//
// Architecture:
//
//	Crawler  ──result──▶  ML OnlineLearner  ◀──detection── WAF
//	  ▲                       │                              ▲
//	  └── ML.Generate() ◀────┘──── ML.Infer() ─────────────┘
//
// The crawler tests fingerprints against real targets, feeding success/failure
// results back to the ML system. The WAF detects suspicious fingerprints and
// feeds detection patterns to ML. ML evolves models that improve both
// fingerprint generation (for crawler) and detection (for WAF).
type ClosedLoopController struct {
	config    *ClosedLoopConfig
	crawler   *crawler.Crawler
	waf       *waf.WAF
	mlService *ml.MLService

	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
	stats  *ClosedLoopStats
	logger *slog.Logger
}

// ClosedLoopStats tracks closed-loop training statistics.
type ClosedLoopStats struct {
	CyclesCompleted     int64
	ProfilesGenerated   int64
	DetectionsProcessed int64
	ModelsEvolved       int64
	LastCycleTime       time.Time
}

// NewClosedLoopController creates the closed-loop orchestrator.
func NewClosedLoopController(config *ClosedLoopConfig, mlSvc *ml.MLService) *ClosedLoopController {
	if config == nil {
		config = DefaultClosedLoopConfig
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &ClosedLoopController{
		config:    config,
		mlService: mlSvc,
		ctx:       ctx,
		cancel:    cancel,
		stats:     &ClosedLoopStats{},
		logger:    slog.Default().With("component", "closed-loop"),
	}
}

// SetCrawler sets the crawler instance for the closed loop.
func (c *ClosedLoopController) SetCrawler(cr *crawler.Crawler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.crawler = cr
}

// SetWAF sets the WAF instance for the closed loop.
func (c *ClosedLoopController) SetWAF(w *waf.WAF) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.waf = w
}

// Start begins the closed-loop adversarial training cycle.
func (c *ClosedLoopController) Start() {
	if !c.config.Enabled {
		return
	}

	go c.trainingLoop()
	c.logger.Info("closed-loop controller started",
		"interval", c.config.TrainingInterval,
		"candidates_per_cycle", c.config.CandidatesPerCycle)
}

// Stop halts the closed-loop controller.
func (c *ClosedLoopController) Stop() {
	c.cancel()
}

// Stats returns current closed-loop statistics.
func (c *ClosedLoopController) Stats() *ClosedLoopStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	s := *c.stats
	return &s
}

// trainingLoop runs periodic adversarial training cycles.
func (c *ClosedLoopController) trainingLoop() {
	interval := c.config.TrainingInterval
	if interval <= 0 {
		interval = DefaultClosedLoopConfig.TrainingInterval
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.runTrainingCycle()
		}
	}
}

// runTrainingCycle executes one adversarial training cycle:
//  1. Generate fingerprints with ML
//  2. Validate them against the detection pipeline
//  3. Feed results back to ML for model evolution
func (c *ClosedLoopController) runTrainingCycle() {
	c.mu.RLock()
	mlSvc := c.mlService
	wafInst := c.waf
	c.mu.RUnlock()

	if mlSvc == nil {
		return
	}

	c.logger.Info("starting adversarial training cycle")
	startTime := time.Now()

	generated := c.generateCandidates(mlSvc)
	if len(generated) == 0 {
		return
	}

	// Validate generated profiles against WAF detection
	c.validateAgainstWAF(generated, wafInst, mlSvc)

	// Check if drift was detected and trigger evolution
	c.checkAndEvolve(mlSvc)

	c.mu.Lock()
	c.stats.CyclesCompleted++
	c.stats.LastCycleTime = time.Now()
	c.mu.Unlock()

	c.logger.Info("adversarial training cycle complete",
		"cycle", c.stats.CyclesCompleted,
		"duration", time.Since(startTime))
}

// generateCandidates creates ML-generated fingerprint profiles.
func (c *ClosedLoopController) generateCandidates(mlSvc *ml.MLService) []*ml.GenerateResult {
	candidates := make([]*ml.GenerateResult, 0, c.config.CandidatesPerCycle)

	for i := 0; i < c.config.CandidatesPerCycle; i++ {
		genCfg := &ml.GenerateConfig{
			MaxAttempts:    5,
			NoiseIntensity: c.config.NoiseIntensity,
		}
		result, err := mlSvc.Generate(genCfg)
		if err != nil {
			c.logger.Debug("generation attempt failed", "error", err)
			continue
		}
		if result.Profile != nil {
			candidates = append(candidates, result)
			c.mu.Lock()
			c.stats.ProfilesGenerated++
			c.mu.Unlock()
		}
	}
	return candidates
}

// validateAgainstWAF tests generated profiles against WAF detection.
func (c *ClosedLoopController) validateAgainstWAF(
	candidates []*ml.GenerateResult,
	wafInst *waf.WAF,
	mlSvc *ml.MLService,
) {
	for _, gen := range candidates {
		// Use ML validation as a proxy for WAF testing
		vr := mlSvc.Validate(gen.Profile)

		reward := (1.0 - vr.ForgeryProb) * vr.ConsistencyScore
		mlSvc.Feedback(&ml.FeedbackSample{
			Profile:   gen.Profile,
			Reward:    reward,
			Label:     fmt.Sprintf("generated_%s", gen.SourceProfileID),
			Timestamp: time.Now(),
		})

		c.logger.Debug("validated generated profile",
			"profile", gen.Profile.ID,
			"forgery_prob", vr.ForgeryProb,
			"consistency", vr.ConsistencyScore,
			"reward", reward)
	}
}

// checkAndEvolve checks if the learner has detected drift and triggers
// model evolution to incorporate new observations.
func (c *ClosedLoopController) checkAndEvolve(mlSvc *ml.MLService) {
	learner := mlSvc.Learner()
	if learner == nil || !learner.DriftDetected() {
		return
	}

	c.logger.Info("drift detected, triggering model evolution")
	if _, err := mlSvc.Evolve(nil); err != nil {
		c.logger.Warn("model evolution failed", "error", err)
		return
	}

	c.mu.Lock()
	c.stats.ModelsEvolved++
	c.mu.Unlock()

	c.logger.Info("model evolution completed")
}
