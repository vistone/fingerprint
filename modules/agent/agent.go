// Package agent 实现自主安全智能体（Autonomous Security Agent）
//
// 范式转移：从被动"指纹识别"到主动"行为智能体"
//
// 核心架构：Observe → Analyze → Decide → Act (OADA 循环)
//
//   - Observer: 持续收集指纹分析事件流
//   - BehaviorAnalyzer: 构建客户端行为画像，识别时序异常
//   - StrategyEngine: 自适应策略引擎，根据威胁模式动态演化检测规则
//   - Memory: 智能体记忆系统，存储学习到的模式与威胁签名
//
// Agent 不替换现有 ML/Defense 模块，而是在其之上构建更高层的
// 自主决策能力，形成"感知-认知-决策-执行"完整闭环。
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

// ActionType 智能体动作类型
type ActionType string

const (
	ActionAllow     ActionType = "allow"     // 放行
	ActionMonitor   ActionType = "monitor"   // 监控（不阻断但加强观察）
	ActionChallenge ActionType = "challenge" // 挑战验证（如 JS 验证、验证码）
	ActionThrottle  ActionType = "throttle"  // 限速
	ActionBlock     ActionType = "block"     // 阻断
)

// ThreatClass 威胁分类
type ThreatClass string

const (
	ThreatNone              ThreatClass = "none"
	ThreatBot               ThreatClass = "bot"                // 自动化工具/爬虫
	ThreatFingerprintSpoof  ThreatClass = "fingerprint_spoof"  // 指纹伪造
	ThreatSessionAnomaly    ThreatClass = "session_anomaly"    // 会话异常
	ThreatBehavioralAnomaly ThreatClass = "behavioral_anomaly" // 行为异常
	ThreatEvasion           ThreatClass = "evasion"            // 主动逃避检测
)

// Observation 单次观测事件——智能体的感知输入
type Observation struct {
	ID        string
	ClientID  string // 客户端标识 (IP 或 session)
	Timestamp time.Time

	// 来自现有管线的原始数据
	Features       *core.FeatureVector
	Classification *ml.ClassificationResult
	Detection      *defense.DetectionResult
	RiskAssessment *core.RiskAssessment

	// 指纹标识
	FingerprintHash string

	// 元信息
	Metadata map[string]string
}

// Decision 智能体决策结果——丰富了原始风险评估
type Decision struct {
	// 决策动作
	Action ActionType `json:"action"`

	// 威胁分类
	ThreatClass ThreatClass `json:"threat_class"`

	// 综合置信度 [0,1]
	Confidence float64 `json:"confidence"`

	// 行为画像摘要
	BehaviorSummary *BehaviorSummary `json:"behavior_summary,omitempty"`

	// 被触发的自适应规则
	TriggeredStrategies []string `json:"triggered_strategies,omitempty"`

	// 知识库匹配结果——跨层一致性校验
	KnowledgeMatch *MatchResult `json:"knowledge_match,omitempty"`

	// 建议（补充 RiskAssessment.Suggestions）
	Insights []string `json:"insights,omitempty"`

	// 处理耗时
	LatencyUs int64 `json:"latency_us"`
}

// BehaviorSummary 行为画像摘要
type BehaviorSummary struct {
	TotalObservations   int     `json:"total_observations"`
	UniqueFP            int     `json:"unique_fingerprints"`    // 不同指纹数
	FPSwitchRate        float64 `json:"fp_switch_rate"`         // 指纹切换频率 (switches/min)
	AvgRequestInterval  float64 `json:"avg_request_interval_s"` // 平均请求间隔(秒)
	ConsistencyScore    float64 `json:"consistency_score"`      // 指纹一致性得分 [0,1]
	RiskTrend           float64 `json:"risk_trend"`             // 风险趋势 (-1~1, 正=恶化)
	SessionDurationSecs float64 `json:"session_duration_s"`
}

// AgentConfig 智能体配置
type AgentConfig struct {
	// 行为分析
	SessionWindow   time.Duration // 会话窗口长度（默认 30min）
	MaxObservations int           // 每客户端最大保留观测数
	CleanupInterval time.Duration // 过期会话清理间隔
	SessionTimeout  time.Duration // 会话超时（无活动后过期）

	// 策略引擎
	StrategyUpdateInterval time.Duration // 策略自动演化间隔
	MinObservationsToLearn int           // 触发学习的最少观测数

	// 阈值
	FPSwitchRateThreshold float64 // 指纹切换频率异常阈值
	ConsistencyThreshold  float64 // 一致性得分异常阈值
	RequestBurstThreshold float64 // 请求突发异常阈值 (req/s)
	RiskEscalationFactor  float64 // 风险升级因子

	// 后台协程
	Enabled bool // 是否启用 Agent
}

// DefaultAgentConfig 默认配置
var DefaultAgentConfig = &AgentConfig{
	SessionWindow:          30 * time.Minute,
	MaxObservations:        500,
	CleanupInterval:        5 * time.Minute,
	SessionTimeout:         30 * time.Minute,
	StrategyUpdateInterval: 10 * time.Minute,
	MinObservationsToLearn: 20,
	FPSwitchRateThreshold:  3.0,  // 每分钟 3 次指纹切换视为异常
	ConsistencyThreshold:   0.4,  // 一致性低于 0.4 视为可疑
	RequestBurstThreshold:  10.0, // 每秒 10 请求视为突发
	RiskEscalationFactor:   1.5,
	Enabled:                true,
}

// Agent 自主安全智能体
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

// NewAgent 创建新的安全智能体
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

// Start 启动智能体后台协程（清理、策略演化等）
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
