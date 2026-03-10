package agent

import (
	"fmt"
	"sync"
	"time"
)

// StrategyEngine 自适应策略引擎
//
// 管理一组动态检测策略，每条策略描述一种威胁模式的检测条件和响应动作。
// 引擎具备两种能力：
//
//  1. 规则评估：对每次观测 + 行为画像执行所有活跃策略
//  2. 自演化（Evolve）：根据积累的观测统计自动生成/淘汰策略
type StrategyEngine struct {
	config     *AgentConfig
	memory     *Memory
	strategies []*Strategy
	mu         sync.RWMutex

	// 学习到的模式计数
	learnedPatterns int
}

// Strategy 单条自适应策略
type Strategy struct {
	ID          string
	Name        string
	Description string
	ThreatClass ThreatClass
	Action      ActionType
	Priority    int // 越高越先评估

	// 条件函数：输入观测 + 行为画像，返回是否命中
	Condition func(obs *Observation, profile *BehaviorSummary) bool

	// 统计
	HitCount  int
	MissCount int
	CreatedAt time.Time
	Learned   bool // 是否由演化自动生成
	Enabled   bool
}

// StrategyInfo 策略摘要（对外暴露，不含函数）
type StrategyInfo struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	ThreatClass ThreatClass `json:"threat_class"`
	Action      ActionType  `json:"action"`
	Priority    int         `json:"priority"`
	HitCount    int         `json:"hit_count"`
	Learned     bool        `json:"learned"`
	Enabled     bool        `json:"enabled"`
}

// NewStrategyEngine 创建策略引擎并注册内置策略
func NewStrategyEngine(config *AgentConfig, memory *Memory) *StrategyEngine {
	se := &StrategyEngine{
		config:     config,
		memory:     memory,
		strategies: make([]*Strategy, 0, 16),
	}
	se.registerBuiltinStrategies()
	return se
}

// registerBuiltinStrategies 注册内置策略（基线）
func (se *StrategyEngine) registerBuiltinStrategies() {
	cfg := se.config
	se.strategies = append(se.strategies,
		// S1: 指纹快速切换 → 可能为指纹伪造/轮换工具
		&Strategy{
			ID:          "builtin_fp_switch",
			Name:        "指纹快速切换检测",
			Description: "客户端在短时间内频繁切换指纹，疑似使用反检测工具",
			ThreatClass: ThreatFingerprintSpoof,
			Action:      ActionChallenge,
			Priority:    90,
			Condition: func(obs *Observation, profile *BehaviorSummary) bool {
				return profile.FPSwitchRate > cfg.FPSwitchRateThreshold
			},
			CreatedAt: time.Now(),
			Enabled:   true,
		},
		// S2: 低一致性 → 行为异常
		&Strategy{
			ID:          "builtin_low_consistency",
			Name:        "低指纹一致性检测",
			Description: "客户端指纹/分类结果一致性极低，行为模式异常",
			ThreatClass: ThreatBehavioralAnomaly,
			Action:      ActionMonitor,
			Priority:    70,
			Condition: func(obs *Observation, profile *BehaviorSummary) bool {
				return profile.TotalObservations >= 3 &&
					profile.ConsistencyScore < cfg.ConsistencyThreshold
			},
			CreatedAt: time.Now(),
			Enabled:   true,
		},
		// S3: 请求突发 → 可能为自动化工具
		&Strategy{
			ID:          "builtin_request_burst",
			Name:        "请求突发检测",
			Description: "请求频率异常高，疑似自动化工具或爬虫",
			ThreatClass: ThreatBot,
			Action:      ActionThrottle,
			Priority:    80,
			Condition: func(obs *Observation, profile *BehaviorSummary) bool {
				if profile.AvgRequestInterval <= 0 || profile.TotalObservations < 5 {
					return false
				}
				reqPerSec := 1.0 / profile.AvgRequestInterval
				return reqPerSec > cfg.RequestBurstThreshold
			},
			CreatedAt: time.Now(),
			Enabled:   true,
		},
		// S4: 风险持续上升 → 升级响应
		&Strategy{
			ID:          "builtin_risk_escalation",
			Name:        "风险上升趋势检测",
			Description: "客户端风险评分持续上升，需要升级防护",
			ThreatClass: ThreatEvasion,
			Action:      ActionChallenge,
			Priority:    85,
			Condition: func(obs *Observation, profile *BehaviorSummary) bool {
				return profile.RiskTrend > 0.5 && profile.TotalObservations >= 5
			},
			CreatedAt: time.Now(),
			Enabled:   true,
		},
		// S5: 高危规则触发 + 低 ML 置信 → 阻断
		&Strategy{
			ID:          "builtin_high_risk_block",
			Name:        "高危综合阻断",
			Description: "多个高危信号同时触发：规则命中+ML低置信+行为异常",
			ThreatClass: ThreatEvasion,
			Action:      ActionBlock,
			Priority:    95,
			Condition: func(obs *Observation, profile *BehaviorSummary) bool {
				if obs.RiskAssessment == nil || obs.Classification == nil {
					return false
				}
				highRisk := obs.RiskAssessment.Score > 0.7
				lowConf := obs.Classification.Confidence < 0.3
				badBehavior := profile.ConsistencyScore < 0.3
				return highRisk && lowConf && badBehavior
			},
			CreatedAt: time.Now(),
			Enabled:   true,
		},
	)
}

// Evaluate 评估所有活跃策略，返回最终决策
func (se *StrategyEngine) Evaluate(obs *Observation, profile *BehaviorSummary) *Decision {
	se.mu.RLock()
	defer se.mu.RUnlock()

	decision := &Decision{
		Action:      ActionAllow,
		ThreatClass: ThreatNone,
		Confidence:  1.0,
	}

	if profile != nil {
		decision.BehaviorSummary = profile
	}

	// 收集所有命中的策略，取最高优先级的 Action
	var triggered []*Strategy
	for _, s := range se.strategies {
		if !s.Enabled {
			continue
		}
		if s.Condition(obs, profile) {
			triggered = append(triggered, s)
			s.HitCount++
		} else {
			s.MissCount++
		}
	}

	if len(triggered) == 0 {
		return decision
	}

	// 找最高优先级策略
	best := triggered[0]
	for _, s := range triggered[1:] {
		if s.Priority > best.Priority {
			best = s
		}
	}

	decision.Action = best.Action
	decision.ThreatClass = best.ThreatClass

	// 置信度 = 命中策略的 hit ratio（命中率越高越可靠）
	if best.HitCount+best.MissCount > 0 {
		decision.Confidence = float64(best.HitCount) / float64(best.HitCount+best.MissCount)
	}

	for _, s := range triggered {
		decision.TriggeredStrategies = append(decision.TriggeredStrategies, s.ID)
		decision.Insights = append(decision.Insights, s.Description)
	}

	return decision
}

// Evolve 策略自演化——根据积累的观测统计自动生成新策略或淘汰无效策略
func (se *StrategyEngine) Evolve() {
	se.mu.Lock()
	defer se.mu.Unlock()

	// 1. 淘汰长期无命中的学习策略
	var kept []*Strategy
	for _, s := range se.strategies {
		if s.Learned && s.HitCount == 0 && s.MissCount > 100 {
			continue // 淘汰
		}
		kept = append(kept, s)
	}
	se.strategies = kept

	// 2. 从观测统计中学习新模式
	se.learnFromObservations()
}

// learnFromObservations 从观测数据中学习新的威胁模式
func (se *StrategyEngine) learnFromObservations() {
	sessions := se.memory.AllSessions()
	if len(sessions) < se.config.MinObservationsToLearn {
		return
	}

	// 统计全局指纹切换频率分布
	var totalSwitchRate float64
	var sessionCount int

	for _, clientID := range sessions {
		session := se.memory.GetSession(clientID)
		if session == nil || len(session.Observations) < 3 {
			continue
		}
		// 简易切换率计算
		switches := 0
		obs := session.Observations
		for i := 1; i < len(obs); i++ {
			if obs[i].FingerprintHash != obs[i-1].FingerprintHash {
				switches++
			}
		}
		duration := obs[len(obs)-1].Timestamp.Sub(obs[0].Timestamp)
		if duration.Minutes() > 0 {
			totalSwitchRate += float64(switches) / duration.Minutes()
			sessionCount++
		}
	}

	if sessionCount == 0 {
		return
	}

	avgSwitchRate := totalSwitchRate / float64(sessionCount)

	// 如果全局平均切换率远高于阈值的一半，生成更严格的策略
	if avgSwitchRate > se.config.FPSwitchRateThreshold*0.5 {
		newThreshold := avgSwitchRate * 1.5
		strategyID := fmt.Sprintf("learned_adaptive_fp_switch_%d", se.learnedPatterns)

		// 检查是否已有类似策略
		for _, s := range se.strategies {
			if s.ThreatClass == ThreatFingerprintSpoof && s.Learned {
				return // 已有，跳过
			}
		}

		se.strategies = append(se.strategies, &Strategy{
			ID:          strategyID,
			Name:        "自适应指纹切换检测",
			Description: fmt.Sprintf("基于观测学习: 切换率 > %.1f/min 视为异常 (全局均值 %.1f/min)", newThreshold, avgSwitchRate),
			ThreatClass: ThreatFingerprintSpoof,
			Action:      ActionMonitor,
			Priority:    60,
			Condition: func(obs *Observation, profile *BehaviorSummary) bool {
				return profile.FPSwitchRate > newThreshold
			},
			CreatedAt: time.Now(),
			Learned:   true,
			Enabled:   true,
		})
		se.learnedPatterns++
	}
}

// ListActive 返回所有活跃策略摘要
func (se *StrategyEngine) ListActive() []StrategyInfo {
	se.mu.RLock()
	defer se.mu.RUnlock()

	var result []StrategyInfo
	for _, s := range se.strategies {
		if !s.Enabled {
			continue
		}
		result = append(result, StrategyInfo{
			ID:          s.ID,
			Name:        s.Name,
			Description: s.Description,
			ThreatClass: s.ThreatClass,
			Action:      s.Action,
			Priority:    s.Priority,
			HitCount:    s.HitCount,
			Learned:     s.Learned,
			Enabled:     s.Enabled,
		})
	}
	return result
}

// LearnedPatternCount 返回已学习的模式数量
func (se *StrategyEngine) LearnedPatternCount() int {
	se.mu.RLock()
	defer se.mu.RUnlock()
	return se.learnedPatterns
}
