package waf

import (
	"sync"
	"time"

	"github.com/vistone/fingerprint/modules/agent"
	"github.com/vistone/fingerprint/modules/defense"
	"github.com/vistone/fingerprint/modules/ml"
)

// WAF firewall instance
type WAF struct {
	config *WAFConfig

	// Layer protection engines
	networkEngine  *NetworkEngine
	tlsEngine      *TLSEngine
	httpEngine     *HTTPEngine
	behaviorEngine *BehaviorEngine
	deviceEngine   *DeviceEngine

	// Core components
	classifier *ml.HierarchicalClassifier
	mlService  *ml.MLService
	detector   *defense.Detector
	agent      *agent.Agent
	riskEngine *defense.RiskEngine

	// Learning pipeline
	learningPipeline *LearningPipeline

	// Block management
	blockList   *BlockList
	rateLimiter *RateLimiter

	// Status
	mu      sync.RWMutex
	running bool
	stats   *WAFStats

	decisionMu      sync.RWMutex
	recentDecisions []WAFDecision
}

// NewWAF creates a new WAF instance
func NewWAF(config *WAFConfig) *WAF {
	if config == nil {
		config = DefaultWAFConfig
	}

	waf := &WAF{
		config:          config,
		stats:           &WAFStats{},
		recentDecisions: make([]WAFDecision, 0, 64),
	}

	// Initialize classifier
	waf.classifier = ml.NewHierarchicalClassifier()
	waf.classifier.Initialize()

	// Initialize detector
	waf.detector = defense.NewDetector()

	// Initialize risk engine
	waf.riskEngine = defense.NewRiskEngine()

	// Initialize layer engines
	if config.NetworkLayerEnabled {
		waf.networkEngine = NewNetworkEngine()
		waf.networkEngine.trustedProxies = append([]string(nil), config.TrustedProxies...)
	}
	if config.TLSLayerEnabled {
		waf.tlsEngine = NewTLSEngine(config.BlacklistJA3, config.BlacklistJA4)
	}
	if config.HTTPLayerEnabled {
		waf.httpEngine = NewHTTPEngine()
	}
	if config.BehaviorLayerEnabled {
		waf.behaviorEngine = NewBehaviorEngine()
	}
	if config.DeviceLayerEnabled {
		waf.deviceEngine = NewDeviceEngine()
	}

	// Initialize ML service
	if config.MLEnabled && config.MLClassifierPath != "" {
		svc, _ := ml.NewMLService(&ml.ServiceConfig{
			ModelStorePath: config.MLClassifierPath,
			AutoLoadLatest: true,
		})
		waf.mlService = svc
	}

	// Initialize learning pipeline
	if waf.mlService != nil {
		waf.learningPipeline = NewLearningPipeline(waf.mlService)
	}

	// Initialize autonomous agent
	if config.AgentEnabled {
		waf.agent = agent.NewAgent(&agent.AgentConfig{
			Enabled: true,
		})
		waf.agent.Start()
	}

	// Initialize block list
	waf.blockList = NewBlockList(config.BlockDuration)

	// Initialize rate limiter
	waf.rateLimiter = NewRateLimiter(config.RateLimitRPS, config.RateLimitBurst, time.Second)

	return waf
}

// Stats retrieves statistics
func (w *WAF) Stats() WAFStats {
	return *w.stats
}

// MLService returns the WAF's ML service instance (nil if not configured).
func (w *WAF) MLService() *ml.MLService {
	return w.mlService
}

// LearningPipelineStats returns learning pipeline statistics.
func (w *WAF) LearningPipelineStats() *LearningStats {
	if w.learningPipeline == nil {
		return nil
	}
	return w.learningPipeline.LearningStats()
}

// Stop stops the WAF
func (w *WAF) Stop() {
	if w.agent != nil {
		w.agent.Stop()
	}
	if w.blockList != nil {
		w.blockList.Stop()
	}
}
