package agent

import (
	"math"
	"time"
)

// BehaviorAnalyzer behavior analyzer
//
// Extracts behavioral features from observation history, builds security behavior profile.
// Focus dimensions:
//   - Fingerprint stability (switching frequency, unique fingerprint count)
//   - Request patterns (interval, burst, periodicity)
//   - Risk trend (time series change direction of risk scores)
//   - Classification consistency (whether ML classification results are consistent)
type BehaviorAnalyzer struct {
	config *AgentConfig
	memory *Memory
}

// NewBehaviorAnalyzer create behavior analyzer
func NewBehaviorAnalyzer(config *AgentConfig, memory *Memory) *BehaviorAnalyzer {
	return &BehaviorAnalyzer{
		config: config,
		memory: memory,
	}
}

// Analyze analyze behavior profile of specified client
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

	// Session duration
	first := obs[0].Timestamp
	last := obs[n-1].Timestamp
	duration := last.Sub(first)
	summary.SessionDurationSecs = duration.Seconds()

	// Fingerprint switching frequency
	switches := ba.countFPSwitches(obs)
	if duration > 0 {
		summary.FPSwitchRate = float64(switches) / duration.Minutes()
	}

	// Average request interval
	summary.AvgRequestInterval = ba.avgInterval(obs)

	// Consistency score
	summary.ConsistencyScore = ba.consistencyScore(obs)

	// Risk trend
	summary.RiskTrend = ba.riskTrend(obs)

	return summary
}

// countFPSwitches count fingerprint switches
func (ba *BehaviorAnalyzer) countFPSwitches(obs []*Observation) int {
	switches := 0
	for i := 1; i < len(obs); i++ {
		if obs[i].FingerprintHash != obs[i-1].FingerprintHash {
			switches++
		}
	}
	return switches
}

// avgInterval calculate average request interval (seconds)
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

// consistencyScore calculate fingerprint consistency score [0,1]
//
// Based on two factors:
// 1. Fingerprint diversity: unique fingerprint count / total observations (lower is more consistent)
// 2. ML classification stability: dominant classification proportion
func (ba *BehaviorAnalyzer) consistencyScore(obs []*Observation) float64 {
	if len(obs) == 0 {
		return 1.0
	}

	// Factor 1: Fingerprint diversity (inverted)
	fpSet := make(map[string]struct{})
	for _, o := range obs {
		fpSet[o.FingerprintHash] = struct{}{}
	}
	diversityPenalty := float64(len(fpSet)) / float64(len(obs))

	// Factor 2: ML classification consistency
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

	// Weighted merge
	score := 0.5*(1.0-diversityPenalty) + 0.5*classStability
	return math.Max(0, math.Min(1, score))
}

// riskTrend calculate risk trend [-1, 1]
//
// Uses simple linear regression slope on risk scores of recent observations,
// positive value indicates risk is rising, negative value indicates risk is falling.
func (ba *BehaviorAnalyzer) riskTrend(obs []*Observation) float64 {
	n := len(obs)
	if n < 3 {
		return 0
	}

	// Only take most recent 20
	start := 0
	if n > 20 {
		start = n - 20
	}
	recent := obs[start:]
	m := len(recent)

	// Simple linear regression y = a + bx
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
	// Normalize to [-1, 1]
	return math.Max(-1, math.Min(1, slope*10))
}
