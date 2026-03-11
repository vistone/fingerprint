package agent

import (
	"math"
	"testing"
	"time"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/ml"
)

func TestNewReinforcementEngine(t *testing.T) {
	re := NewReinforcementEngine(nil)
	if re == nil {
		t.Fatal("NewReinforcementEngine returned nil")
	}
	if re.epsilon != DefaultRLConfig.EpsilonMax {
		t.Errorf("initial epsilon = %f, want %f", re.epsilon, DefaultRLConfig.EpsilonMax)
	}
	stats := re.Stats()
	if stats.StatesExplored != 0 || stats.Episodes != 0 {
		t.Errorf("fresh engine should have 0 states/episodes, got %+v", stats)
	}
}

func TestDiscretizeState(t *testing.T) {
	tests := []struct {
		name    string
		risk    float64
		conf    float64
		consist float64
		switch_ float64
		want    State
	}{
		{"all_zero", 0, 1.0, 1.0, 0, State{0, 0, 0, 0}},
		{"high_risk_low_conf", 0.9, 0.2, 0.1, 6.0, State{3, 3, 3, 3}},
		{"medium", 0.6, 0.55, 0.5, 1.0, State{1, 2, 1, 1}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			obs := &Observation{
				RiskAssessment: &core.RiskAssessment{Score: tc.risk},
				Classification: &ml.ClassificationResult{Confidence: tc.conf},
			}
			profile := &BehaviorSummary{
				ConsistencyScore: tc.consist,
				FPSwitchRate:     tc.switch_,
			}
			got := DiscretizeState(obs, profile)
			if got != tc.want {
				t.Errorf("DiscretizeState = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestQLearningUpdate(t *testing.T) {
	cfg := &RLConfig{
		Alpha:        0.5,
		Gamma:        0.9,
		EpsilonMax:   0.0, // no exploration for deterministic test
		EpsilonMin:   0.0,
		EpsilonDecay: 1.0,
	}
	re := NewReinforcementEngine(cfg)

	state := State{1, 1, 1, 1}
	nextState := State{0, 0, 0, 0}

	// Q(s,a) starts at 0. After update with reward=1.0:
	// Q(s, block) = 0 + 0.5 * (1.0 + 0.9*0 - 0) = 0.5
	re.Update(state, ActionBlock, 1.0, nextState)
	q := re.QValue(state, ActionBlock)
	if math.Abs(q-0.5) > 1e-9 {
		t.Errorf("Q(state, block) = %f, want 0.5", q)
	}

	// BestAction should now be Block for this state
	best := re.BestAction(state)
	if best != ActionBlock {
		t.Errorf("BestAction = %s, want %s", best, ActionBlock)
	}
}

func TestEpsilonDecay(t *testing.T) {
	cfg := &RLConfig{
		Alpha:        0.1,
		Gamma:        0.9,
		EpsilonMax:   1.0,
		EpsilonMin:   0.01,
		EpsilonDecay: 0.5, // aggressive decay for testing
	}
	re := NewReinforcementEngine(cfg)

	s := State{}
	for i := 0; i < 10; i++ {
		re.Update(s, ActionAllow, 0, s)
	}

	eps := re.Epsilon()
	if eps >= 1.0 {
		t.Errorf("epsilon should have decayed, got %f", eps)
	}
	if eps < cfg.EpsilonMin {
		t.Errorf("epsilon %f dropped below min %f", eps, cfg.EpsilonMin)
	}
}

func TestSelectActionGreedy(t *testing.T) {
	cfg := &RLConfig{
		Alpha:        0.5,
		Gamma:        0.9,
		EpsilonMax:   0.0, // fully greedy
		EpsilonMin:   0.0,
		EpsilonDecay: 1.0,
	}
	re := NewReinforcementEngine(cfg)

	state := State{2, 2, 2, 2}
	re.Update(state, ActionChallenge, 5.0, State{})

	action, explored := re.SelectAction(state)
	if explored {
		t.Error("should not explore with epsilon=0")
	}
	if action != ActionChallenge {
		t.Errorf("SelectAction = %s, want %s", action, ActionChallenge)
	}
}

func TestComputeReward(t *testing.T) {
	tests := []struct {
		action     ActionType
		threat     bool
		fpCost     float64
		fnCost     float64
		wantSign   int // +1, -1, or 0
	}{
		{ActionBlock, true, 1.0, 2.0, 1},   // true positive
		{ActionAllow, false, 1.0, 2.0, 1},  // true negative
		{ActionBlock, false, 1.0, 2.0, -1}, // false positive
		{ActionAllow, true, 1.0, 2.0, -1},  // false negative
	}

	for _, tc := range tests {
		r := ComputeReward(tc.action, tc.threat, tc.fpCost, tc.fnCost)
		gotSign := 0
		if r > 0 {
			gotSign = 1
		} else if r < 0 {
			gotSign = -1
		}
		if gotSign != tc.wantSign {
			t.Errorf("ComputeReward(%s, threat=%v) = %f (sign %d), want sign %d",
				tc.action, tc.threat, r, gotSign, tc.wantSign)
		}
	}
}

func TestQLearningConvergence(t *testing.T) {
	cfg := &RLConfig{
		Alpha:        0.1,
		Gamma:        0.95,
		EpsilonMax:   0.0,
		EpsilonMin:   0.0,
		EpsilonDecay: 1.0,
	}
	re := NewReinforcementEngine(cfg)

	// Repeatedly reward blocking in a high-risk state
	highRisk := State{3, 3, 3, 3}
	safe := State{0, 0, 0, 0}

	for i := 0; i < 200; i++ {
		re.Update(highRisk, ActionBlock, 1.0, safe)
		re.Update(safe, ActionAllow, 0.5, safe)
	}

	// Block should be best for high-risk, Allow for safe
	if re.BestAction(highRisk) != ActionBlock {
		t.Errorf("expected Block for high-risk state, got %s", re.BestAction(highRisk))
	}
	if re.BestAction(safe) != ActionAllow {
		t.Errorf("expected Allow for safe state, got %s", re.BestAction(safe))
	}
}

func TestRLIntegrationWithAgent(t *testing.T) {
	cfg := *DefaultAgentConfig
	cfg.RLConfig = DefaultRLConfig
	a := NewAgent(&cfg)

	if a.rl == nil {
		t.Fatal("agent RL engine not initialized")
	}

	now := time.Now()
	obs := makeObs("rl-obs-1", "client-rl", "fp-rl", now, 0.9, 0.1)
	dec := a.Process(nil, obs)
	if dec == nil {
		t.Fatal("Process returned nil")
	}

	stats := a.Stats()
	if stats.RLStats == nil {
		t.Error("expected RLStats in agent stats")
	}
}
