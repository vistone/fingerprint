package agent

import (
	"time"

	"github.com/vistone/fingerprint/modules/ml"
)

func (a *Agent) ReportReward(obs *Observation, profile *BehaviorSummary, action ActionType,
	wasActualThreat bool, matchResult *MatchResult) {

	if a.rl != nil {
		stateVec := ExtractStateVector(obs, profile)
		reward := ComputeReward(action, wasActualThreat, 0.8, 1.0)
		// Terminal transition: done=true so no bootstrap from next state
		a.rl.UpdateContinuous(stateVec, action, reward, stateVec, true)
	}

	if a.bandit != nil {
		ctx := BuildContext(obs, profile, matchResult)
		reward := ComputeReward(action, wasActualThreat, 0.8, 1.0)
		// Update all triggered strategies that led to this action
		for _, s := range a.strategy.ListActive() {
			if s.Action == action {
				a.bandit.UpdateReward(s.ID, ctx, reward)
			}
		}
	}
}

// cleanupLoop background goroutine for cleaning up expired sessions.
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

// strategyEvolutionLoop background goroutine for automatic strategy evolution.
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

// Pipeline returns the agent's neural network model pipeline.
func (a *Agent) Pipeline() *ml.ModelPipeline {
	return a.pipeline
}

// extractBehaviorVector converts behavioral profile to 8-dim feature vector for the model pipeline.
func (a *Agent) extractBehaviorVector(profile *BehaviorSummary) []float64 {
	vec := make([]float64, ml.BehaviorFeatureDim)
	if profile == nil {
		return vec
	}
	vec[0] = float64(profile.TotalObservations) / 100.0
	vec[1] = float64(profile.UniqueFP) / 10.0
	vec[2] = profile.FPSwitchRate / 5.0
	vec[3] = profile.AvgRequestInterval / 60.0
	vec[4] = profile.ConsistencyScore
	vec[5] = (profile.RiskTrend + 1.0) / 2.0 // normalize -1~1 → 0~1
	vec[6] = profile.SessionDurationSecs / 3600.0
	// vec[7] reserved

	for i := range vec {
		if vec[i] < 0 {
			vec[i] = 0
		} else if vec[i] > 1 {
			vec[i] = 1
		}
	}
	return vec
}

// threatActionToAgentAction maps model ActionClass to agent ActionType.
func threatActionToAgentAction(action ml.ActionClass) ActionType {
	switch action {
	case ml.ActAllow:
		return ActionAllow
	case ml.ActMonitor:
		return ActionMonitor
	case ml.ActChallenge:
		return ActionChallenge
	case ml.ActThrottle:
		return ActionThrottle
	case ml.ActBlock:
		return ActionBlock
	default:
		return ActionAllow
	}
}

// forgeryTypeName returns human-readable forgery type name.
func forgeryTypeName(ft ml.ForgeryType) string {
	switch ft {
	case ml.ForgeryReal:
		return "real"
	case ml.ForgeryHeadless:
		return "headless"
	case ml.ForgeryAntiDetect:
		return "antidetect"
	case ml.ForgeryProxy:
		return "proxy"
	default:
		return "unknown"
	}
}

// threatClassName returns human-readable threat class name.
func threatClassName(tc ml.ThreatClass) string {
	switch tc {
	case ml.ThreatNone:
		return "none"
	case ml.ThreatBot:
		return "bot"
	case ml.ThreatFingerprintSpoof:
		return "fingerprint_spoof"
	case ml.ThreatSessionAnomaly:
		return "session_anomaly"
	case ml.ThreatBehavioralAnomaly:
		return "behavioral_anomaly"
	case ml.ThreatEvasion:
		return "evasion"
	default:
		return "unknown"
	}
}
