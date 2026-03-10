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
type Agent struct {
	config    *AgentConfig
	behavior  *BehaviorAnalyzer
	strategy  *StrategyEngine
	memory    *Memory
	knowledge *KnowledgeBase
	anomaly   *AnomalyDetector

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

	return &Agent{
		config:    config,
		behavior:  NewBehaviorAnalyzer(config, mem),
		strategy:  NewStrategyEngine(config, mem),
		memory:    mem,
		knowledge: kb,
		anomaly:   NewAnomalyDetector(kb),
		stopCh:    make(chan struct{}),
	}
}

// Start initiates the agent's background goroutines (cleanup, strategy evolution, etc.)
func (a *Agent) Start() {
	a.wg.Add(2)
	go a.cleanupLoop()
	go a.strategyEvolutionLoop()
}

// Stop 优雅停止智能体
func (a *Agent) Stop() {
	close(a.stopCh)
	a.wg.Wait()
}

// Process 处理一次观测事件并返回决策——OADA 主循环
//
// 这是 Gateway.Analyze() 调用的入口，同步执行，延迟通常 < 1ms。
func (a *Agent) Process(ctx context.Context, obs *Observation) *Decision {
	start := time.Now()

	// O: 记录观测
	a.memory.Record(obs)

	// A1: 行为分析
	profile := a.behavior.Analyze(obs.ClientID)

	// A2: 知识驱动异常检测——用全球指纹蓝图校验观测一致性
	matchResult := a.anomaly.Analyze(obs)

	// D: 策略决策（行为 + 知识双重输入）
	decision := a.strategy.Evaluate(obs, profile)

	// 将知识匹配结果合并到决策
	decision.KnowledgeMatch = matchResult
	if matchResult.SuspicionScore > 0.5 {
		// 知识校验高度可疑，提升威胁等级
		if decision.Action == ActionAllow {
			decision.Action = ActionMonitor
		} else if decision.Action == ActionMonitor {
			decision.Action = ActionChallenge
		}
		decision.ThreatClass = ThreatFingerprintSpoof
		decision.Insights = append(decision.Insights,
			fmt.Sprintf("知识库检测到 %d 处跨层矛盾，可疑度 %.2f",
				len(matchResult.Contradictions), matchResult.SuspicionScore))
	}

	decision.LatencyUs = time.Since(start).Microseconds()
	return decision
}

// GetBehaviorProfile 获取指定客户端的行为画像（供外部查询）
func (a *Agent) GetBehaviorProfile(clientID string) *BehaviorSummary {
	return a.behavior.Analyze(clientID)
}

// GetActiveStrategies 返回当前活跃策略列表
func (a *Agent) GetActiveStrategies() []StrategyInfo {
	return a.strategy.ListActive()
}

// Knowledge 返回智能体的全球指纹知识库
func (a *Agent) Knowledge() *KnowledgeBase {
	return a.knowledge
}

// Stats 返回智能体运行统计
func (a *Agent) Stats() AgentStats {
	return AgentStats{
		ActiveSessions:    a.memory.SessionCount(),
		TotalObservations: a.memory.TotalObservations(),
		ActiveStrategies:  len(a.strategy.ListActive()),
		LearnedPatterns:   a.strategy.LearnedPatternCount(),
	}
}

// AgentStats 运行统计
type AgentStats struct {
	ActiveSessions    int `json:"active_sessions"`
	TotalObservations int `json:"total_observations"`
	ActiveStrategies  int `json:"active_strategies"`
	LearnedPatterns   int `json:"learned_patterns"`
}

// cleanupLoop 后台清理过期会话
func (a *Agent) cleanupLoop() {
	defer a.wg.Done()
	ticker := time.NewTicker(a.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopCh:
			return
		case <-ticker.C:
			a.memory.Cleanup(a.config.SessionTimeout)
		}
	}
}

// strategyEvolutionLoop 后台策略自动演化
func (a *Agent) strategyEvolutionLoop() {
	defer a.wg.Done()
	ticker := time.NewTicker(a.config.StrategyUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopCh:
			return
		case <-ticker.C:
			a.strategy.Evolve()
		}
	}
}
