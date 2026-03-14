// Package agent Implements the Autonomous Security Agent.
//
// Paradigm shift: From passive "fingerprint recognition" to active "behavioral agent".
//
// Core architecture: Observe → Analyze → Decide → Act (OADA loop)
//
//   - Observer: Continuously collect fingerprint analysis event stream
//   - BehaviorAnalyzer: Build client behavioral profile, identify temporal anomalies
//   - StrategyEngine: Adaptive strategy engine, dynamically evolve detection rules based on threat patterns
//   - Memory: Agent memory system, store learned patterns and threat signatures
//
// Agent does not replace existing ML/Defense modules, but builds higher-level
// autonomous decision-making capabilities on top of them, forming a complete "Perception-Cognition-Decision-Execution" closed loop.
package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/defense"
	"github.com/vistone/fingerprint/modules/ml"
)

// ActionType defines the types of actions an agent can take.
type ActionType string

const (
	ActionAllow     ActionType = "allow"     // Allow (pass through)
	ActionMonitor   ActionType = "monitor"   // Monitor (no blocking, enhanced observation)
	ActionChallenge ActionType = "challenge" // Challenge with verification (e.g., JS validation, CAPTCHA)
	ActionThrottle  ActionType = "throttle"  // Rate limit
	ActionBlock     ActionType = "block"     // Block
)

// ThreatClass classifies the types of threats detected.
type ThreatClass string

const (
	ThreatNone              ThreatClass = "none"
	ThreatBot               ThreatClass = "bot"                // Automated tools / crawler bots
	ThreatFingerprintSpoof  ThreatClass = "fingerprint_spoof"  // Fingerprint spoofing
	ThreatSessionAnomaly    ThreatClass = "session_anomaly"    // Session anomaly
	ThreatBehavioralAnomaly ThreatClass = "behavioral_anomaly" // Behavioral anomaly
	ThreatEvasion           ThreatClass = "evasion"            // Active evasion attempts
)

// Observation represents a single observation event - the agent's perception input.
type Observation struct {
	ID        string
	ClientID  string // Client identifier (IP or session)
	Timestamp time.Time

	// Raw data from existing pipeline
	Features       *core.FeatureVector
	Classification *ml.ClassificationResult
	Detection      *defense.DetectionResult
	RiskAssessment *core.RiskAssessment

	// Fingerprint hash identifier
	FingerprintHash string

	// Metadata information
	Metadata map[string]string
}

// Decision represents the agent's decision result - enriched version of raw risk assessment.
type Decision struct {
	// Decision action to take
	Action ActionType `json:"action"`

	// Detected threat classification
	ThreatClass ThreatClass `json:"threat_class"`

	// Composite confidence score [0,1]
	Confidence float64 `json:"confidence"`

	// Behavioral profile summary
	BehaviorSummary *BehaviorSummary `json:"behavior_summary,omitempty"`

	// Triggered adaptive strategies
	TriggeredStrategies []string `json:"triggered_strategies,omitempty"`

	// Knowledge base match results - cross-layer consistency check
	KnowledgeMatch *MatchResult `json:"knowledge_match,omitempty"`

	// Insights (supplements RiskAssessment.Suggestions)
	Insights []string `json:"insights,omitempty"`

	// Processing latency in microseconds
	LatencyUs int64 `json:"latency_us"`
}

// BehaviorSummary summarizes the behavioral profile.
type BehaviorSummary struct {
	TotalObservations   int     `json:"total_observations"`
	UniqueFP            int     `json:"unique_fingerprints"`    // Number of distinct fingerprints
	FPSwitchRate        float64 `json:"fp_switch_rate"`         // Fingerprint switch rate (switches/min)
	AvgRequestInterval  float64 `json:"avg_request_interval_s"` // Average request interval (seconds)
	ConsistencyScore    float64 `json:"consistency_score"`      // Fingerprint consistency score [0,1]
	RiskTrend           float64 `json:"risk_trend"`             // Risk trend (-1~1, positive=worsening)
	SessionDurationSecs float64 `json:"session_duration_s"`
}

// AgentConfig defines the configuration for the agent.
type AgentConfig struct {
	// Behavioral analysis
	SessionWindow   time.Duration // Session window duration (default 30min)
	MaxObservations int           // Maximum observations to retain per client
	CleanupInterval time.Duration // Cleanup interval for expired sessions
	SessionTimeout  time.Duration // Session timeout (expires if inactive)

	// Strategy engine
	StrategyUpdateInterval time.Duration // Strategy auto-evolution interval
	MinObservationsToLearn int           // Minimum observations to trigger learning

	// Thresholds
	FPSwitchRateThreshold float64 // Fingerprint switch rate anomaly threshold
	ConsistencyThreshold  float64 // Consistency score anomaly threshold
	RequestBurstThreshold float64 // Request burst anomaly threshold (req/s)
	RiskEscalationFactor  float64 // Risk escalation factor

	// Reinforcement learning
	RLConfig *RLConfig // Q-learning config (nil = disabled)

	// Contextual bandit
	BanditConfig *BanditConfig // LinUCB config (nil = disabled)

	// Model persistence
	ModelStorePath string // path to model store directory (empty = no persistence)

	// Background goroutines
	Enabled bool // Whether to enable the Agent
}

// DefaultAgentConfig provides the default configuration.
var DefaultAgentConfig = &AgentConfig{
	SessionWindow:          30 * time.Minute,
	MaxObservations:        500,
	CleanupInterval:        5 * time.Minute,
	SessionTimeout:         30 * time.Minute,
	StrategyUpdateInterval: 10 * time.Minute,
	MinObservationsToLearn: 20,
	FPSwitchRateThreshold:  3.0,  // 3+ fingerprint switches per minute considered anomalous
	ConsistencyThreshold:   0.4,  // Consistency score below 0.4 indicates suspicious activity
	RequestBurstThreshold:  10.0, // 10+ requests per second considered burst
	RiskEscalationFactor:   1.5,
	Enabled:                true,
}

// Agent is the Autonomous Security Agent implementation.
//
// Agent is the model orchestrator + memory manager + decision authority:
//   - Maintains per-client session memory (observations, behavioral state)
//   - Orchestrates ModelPipeline for end-to-end inference (encode→classify→forgery detect→threat assess)
//   - Makes final security decisions based on model outputs
//   - Provides feedback loop for continuous model improvement
type Agent struct {
	config    *AgentConfig
	behavior  *BehaviorAnalyzer
	strategy  *StrategyEngine
	memory    *Memory
	knowledge *KnowledgeBase
	anomaly   *AnomalyDetector

	// Neural network model pipeline (core inference engine)
	pipeline *ml.ModelPipeline

	// Intelligence subsystems (legacy, kept for backward compat)
	rl     *ReinforcementEngine // Q-learning (nil if disabled)
	bandit *ContextualBandit    // LinUCB strategy selector (nil if disabled)

	stopCh chan struct{}
	wg     sync.WaitGroup
	mu     sync.RWMutex
}

// NewAgent creates a new Autonomous Security Agent with the provided configuration.
func NewAgent(config *AgentConfig) *Agent {
	if config == nil {
		config = DefaultAgentConfig
	}

	mem := NewMemory(config.SessionWindow, config.MaxObservations)
	kb := NewKnowledgeBase()

	a := &Agent{
		config:    config,
		behavior:  NewBehaviorAnalyzer(config, mem),
		strategy:  NewStrategyEngine(config, mem),
		memory:    mem,
		knowledge: kb,
		anomaly:   NewAnomalyDetector(kb),
		pipeline:  ml.NewModelPipeline(),
		stopCh:    make(chan struct{}),
	}

	// Auto-load model from store if configured
	if config.ModelStorePath != "" {
		storeCfg := ml.DefaultStoreConfig(config.ModelStorePath)
		store, err := ml.NewModelStore(storeCfg)
		if err == nil {
			_, _ = a.pipeline.LoadFromStore(store)
		}
	}

	// Initialize intelligence subsystems
	if config.RLConfig != nil {
		a.rl = NewReinforcementEngine(config.RLConfig)
	}
	if config.BanditConfig != nil {
		a.bandit = NewContextualBandit(config.BanditConfig)
		// Register builtin strategies as bandit arms
		for _, s := range a.strategy.ListActive() {
			a.bandit.RegisterArm(s.ID)
		}
	}

	return a
}

// Start initiates the agent's background goroutines (cleanup, strategy evolution, etc.)
func (a *Agent) Start() {
	a.wg.Add(2)
	go a.cleanupLoop()
	go a.strategyEvolutionLoop()
}

// Stop gracefully stops the agent and waits for background goroutines to finish.
func (a *Agent) Stop() {
	close(a.stopCh)
	a.wg.Wait()
}

// Process handles an observation event and returns a decision - main OADA loop.
//
// Inference pipeline:
//
//	Observation → Behavioral analysis → Model inference (encode→classify→forgery detect→threat assess) → Strategy decision → Output
//
// This is the entry point called by Gateway.Analyze(), executes synchronously,
// typical latency < 1ms.
func (a *Agent) Process(ctx context.Context, obs *Observation) *Decision {
	start := time.Now()

	// O: Record observation
	a.memory.Record(obs)

	// A1: Behavioral analysis
	profile := a.behavior.Analyze(obs.ClientID)

	// A2: Knowledge-driven anomaly detection
	matchResult := a.anomaly.Analyze(obs)

	// D: Strategy decision (behavioral + knowledge dual input)
	decision := a.strategy.Evaluate(obs, profile)

	// Merge knowledge match result into decision
	decision.KnowledgeMatch = matchResult
	if matchResult.SuspicionScore > 0.5 {
		if decision.Action == ActionAllow {
			decision.Action = ActionMonitor
		} else if decision.Action == ActionMonitor {
			decision.Action = ActionChallenge
		}
		decision.ThreatClass = ThreatFingerprintSpoof
		decision.Insights = append(decision.Insights,
			fmt.Sprintf("Knowledge base detected %d cross-layer contradictions, suspicion score %.2f",
				len(matchResult.Contradictions), matchResult.SuspicionScore))
	}

	// Neural model pipeline: run end-to-end inference if features available and model is trained
	if a.pipeline != nil && a.pipeline.Trained() && obs.Features != nil {
		behaviorVec := a.extractBehaviorVector(profile)
		result := a.pipeline.InferFromFeatures(obs.Features, behaviorVec)

		// Enhance decision with model output (only escalate when model confidence is high enough)
		if result.Forgery.ForgeryProb > 0.6 && result.Forgery.ForgeryProb > result.Forgery.TypeProbs[int(ml.ForgeryReal)] {
			decision.Insights = append(decision.Insights,
				fmt.Sprintf("NN forgery detector: %.1f%% probability, type=%s",
					result.Forgery.ForgeryProb*100, forgeryTypeName(result.Forgery.ForgeryType)))
			if decision.Action == ActionAllow {
				decision.Action = ActionMonitor
			}
		}
		if result.Forgery.ForgeryProb > 0.8 {
			if decision.Action == ActionMonitor {
				decision.Action = ActionChallenge
			}
			decision.ThreatClass = ThreatFingerprintSpoof
		}

		// Model threat assessment can escalate actions (only when confidence > 0.7)
		modelAction := threatActionToAgentAction(result.Threat.Action)
		if result.Threat.ActionConfidence > 0.7 && actionIndex[modelAction] > actionIndex[decision.Action] {
			decision.Action = modelAction
			decision.Insights = append(decision.Insights,
				fmt.Sprintf("NN threat assessor escalated to %s (confidence=%.2f, class=%s)",
					modelAction, result.Threat.ThreatProb, threatClassName(result.Threat.ThreatClass)))
		}

		decision.Insights = append(decision.Insights,
			fmt.Sprintf("NN browser: %s (confidence=%.2f)",
				result.Browser.Family, result.Browser.Confidence))
	}

	// DQN override: if reinforcement learning is active, let the neural network refine the action
	if a.rl != nil {
		stateVec := ExtractStateVector(obs, profile)
		rlAction, explored := a.rl.SelectActionContinuous(stateVec)
		if explored {
			decision.Insights = append(decision.Insights, "DQN exploring alternative action")
		}
		if actionIndex[rlAction] > actionIndex[decision.Action] {
			qvals := a.rl.QValueContinuous(stateVec)
			decision.Action = rlAction
			decision.Insights = append(decision.Insights,
				fmt.Sprintf("DQN escalated to %s (Q=%.3f)", rlAction, qvals[actionIndex[rlAction]]))
		}
	}

	// Bandit-driven strategy prioritization
	if a.bandit != nil && len(decision.TriggeredStrategies) > 1 {
		ctx := BuildContext(obs, profile, matchResult)
		bestArm, score := a.bandit.SelectArmAmong(ctx, decision.TriggeredStrategies)
		if bestArm != "" {
			decision.Insights = append(decision.Insights,
				fmt.Sprintf("Bandit selected strategy %s (UCB=%.3f)", bestArm, score))
		}
	}

	decision.LatencyUs = time.Since(start).Microseconds()
	return decision
}

// GetBehaviorProfile retrieves the behavioral profile for the specified client (for external queries).
func (a *Agent) GetBehaviorProfile(clientID string) *BehaviorSummary {
	return a.behavior.Analyze(clientID)
}

// GetActiveStrategies returns the list of currently active strategies.
func (a *Agent) GetActiveStrategies() []StrategyInfo {
	return a.strategy.ListActive()
}

// Knowledge returns the agent's global fingerprint knowledge base.
func (a *Agent) Knowledge() *KnowledgeBase {
	return a.knowledge
}

// Stats returns the agent's runtime statistics.
func (a *Agent) Stats() AgentStats {
	stats := AgentStats{
		ActiveSessions:    a.memory.SessionCount(),
		TotalObservations: a.memory.TotalObservations(),
		ActiveStrategies:  len(a.strategy.ListActive()),
		LearnedPatterns:   a.strategy.LearnedPatternCount(),
	}
	if a.rl != nil {
		rlStats := a.rl.Stats()
		stats.RLStats = &rlStats
	}
	if a.bandit != nil {
		banditStats := a.bandit.Stats()
		stats.BanditStats = &banditStats
	}
	return stats
}

// AgentStats contains runtime statistics.
type AgentStats struct {
	ActiveSessions    int          `json:"active_sessions"`
	TotalObservations int          `json:"total_observations"`
	ActiveStrategies  int          `json:"active_strategies"`
	LearnedPatterns   int          `json:"learned_patterns"`
	RLStats           *RLStats     `json:"rl_stats,omitempty"`
	BanditStats       *BanditStats `json:"bandit_stats,omitempty"`
}

// ReportReward provides external feedback to the RL and Bandit subsystems.
// Call this when the ground truth about an observation is known (e.g., confirmed threat or false positive).
