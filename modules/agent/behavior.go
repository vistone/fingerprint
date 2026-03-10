package agent

import (
	"math"
	"time"
)

// BehaviorAnalyzer 行为分析器
//
// 从观测历史中提取行为特征，构建安全行为画像。
// 关注维度：
//   - 指纹稳定性（切换频率、唯一指纹数）
//   - 请求模式（间隔、突发、周期性）
//   - 风险趋势（风险分数的时间序列变化方向）
//   - 分类一致性（ML 分类结果是否前后一致）
type BehaviorAnalyzer struct {
	config *AgentConfig
	memory *Memory
}

// NewBehaviorAnalyzer 创建行为分析器
func NewBehaviorAnalyzer(config *AgentConfig, memory *Memory) *BehaviorAnalyzer {
	return &BehaviorAnalyzer{
		config: config,
		memory: memory,
	}
}

// Analyze 分析指定客户端的行为画像
func (ba *BehaviorAnalyzer) Analyze(clientID string) *BehaviorSummary {
	session := ba.memory.GetSession(clientID)
	if session == nil || len(session.Observations) == 0 {
		return &BehaviorSummary{ConsistencyScore: 1.0}
	}

	obs := session.Observations
	n := len(obs)

	summary := &BehaviorSummary{
		TotalObservations: n,
		UniqueFP:          len(session.fingerprintSet),
	}

	if n < 2 {
		summary.ConsistencyScore = 1.0
		return summary
	}

	// 会话时长
	first := obs[0].Timestamp
	last := obs[n-1].Timestamp
	duration := last.Sub(first)
	summary.SessionDurationSecs = duration.Seconds()

	// 指纹切换频率
	switches := ba.countFPSwitches(obs)
	if duration > 0 {
		summary.FPSwitchRate = float64(switches) / duration.Minutes()
	}

	// 平均请求间隔
	summary.AvgRequestInterval = ba.avgInterval(obs)

	// 一致性得分
	summary.ConsistencyScore = ba.consistencyScore(obs)

	// 风险趋势
	summary.RiskTrend = ba.riskTrend(obs)

	return summary
}

// countFPSwitches 统计指纹切换次数
func (ba *BehaviorAnalyzer) countFPSwitches(obs []*Observation) int {
	switches := 0
	for i := 1; i < len(obs); i++ {
		if obs[i].FingerprintHash != obs[i-1].FingerprintHash {
			switches++
		}
	}
	return switches
}

// avgInterval 计算平均请求间隔（秒）
func (ba *BehaviorAnalyzer) avgInterval(obs []*Observation) float64 {
	if len(obs) < 2 {
		return 0
	}
	var total time.Duration
	for i := 1; i < len(obs); i++ {
		total += obs[i].Timestamp.Sub(obs[i-1].Timestamp)
	}
	return total.Seconds() / float64(len(obs)-1)
}

// consistencyScore 计算指纹一致性得分 [0,1]
//
// 基于两个因子：
// 1. 指纹多样性：唯一指纹数 / 总观测数（越低越一致）
// 2. ML 分类稳定性：主分类占比
func (ba *BehaviorAnalyzer) consistencyScore(obs []*Observation) float64 {
	if len(obs) == 0 {
		return 1.0
	}

	// 因子1: 指纹多样性（反向）
	fpSet := make(map[string]struct{})
	for _, o := range obs {
		fpSet[o.FingerprintHash] = struct{}{}
	}
	diversityPenalty := float64(len(fpSet)) / float64(len(obs))

	// 因子2: ML 分类一致性
	familyCounts := make(map[string]int)
	for _, o := range obs {
		if o.Classification != nil {
			familyCounts[string(o.Classification.Family)]++
		}
	}
	maxCount := 0
	for _, c := range familyCounts {
		if c > maxCount {
			maxCount = c
		}
	}
	classStability := 1.0
	if len(familyCounts) > 0 {
		classStability = float64(maxCount) / float64(len(obs))
	}

	// 加权合并
	score := 0.5*(1.0-diversityPenalty) + 0.5*classStability
	return math.Max(0, math.Min(1, score))
}

// riskTrend 计算风险趋势 [-1, 1]
//
// 使用最近观测的风险分数做简单线性回归斜率，
// 正值表示风险在上升，负值表示风险在下降。
func (ba *BehaviorAnalyzer) riskTrend(obs []*Observation) float64 {
	n := len(obs)
	if n < 3 {
		return 0
	}

	// 只取最近 20 条
	start := 0
	if n > 20 {
		start = n - 20
	}
	recent := obs[start:]
	m := len(recent)

	// 简单线性回归 y = a + bx
	var sumX, sumY, sumXY, sumX2 float64
	for i, o := range recent {
		x := float64(i)
		y := 0.0
		if o.RiskAssessment != nil {
			y = o.RiskAssessment.Score
		}
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	fm := float64(m)
	denom := fm*sumX2 - sumX*sumX
	if denom == 0 {
		return 0
	}

	slope := (fm*sumXY - sumX*sumY) / denom
	// 归一化到 [-1, 1]
	return math.Max(-1, math.Min(1, slope*10))
}
