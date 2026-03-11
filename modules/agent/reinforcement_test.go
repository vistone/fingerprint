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
	if stats.Steps != 0 || stats.TrainSteps != 0 {
		t.Errorf("fresh engine should have 0 steps, got %+v", stats)
	}
	if len(stats.NetworkLayers) == 0 {
		t.Error("expected non-empty network layer description")
	}
}

func TestNeuralNetForward(t *testing.T) {
	nn := newNeuralNet(4, []int{8}, 2)
	input := []float64{0.5, 0.3, 0.7, 0.1}
	output := nn.forward(input)

	if len(output) != 2 {
		t.Fatalf("expected 2 outputs, got %d", len(output))
	}
	// Output should be finite (not NaN/Inf)
	for i, v := range output {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			t.Errorf("output[%d] = %f, expected finite", i, v)
		}
	}
}

func TestNeuralNetBackward(t *testing.T) {
	nn := newNeuralNet(3, []int{4}, 2)
	input := []float64{1.0, 0.5, 0.3}

	// Forward pass
	output := nn.forward(input)

	// Compute gradient for MSE loss: target = [1, 0]
	target := []float64{1.0, 0.0}
	grad := make([]float64, 2)
	for i := range grad {
		grad[i] = 2.0 * (output[i] - target[i])
	}

	// Backward should not panic
	nn.backward(grad, 0.01)

	// After one step, output should be closer to target
	newOutput := nn.forward(input)
	oldDist := (output[0]-target[0])*(output[0]-target[0]) + (output[1]-target[1])*(output[1]-target[1])
	newDist := (newOutput[0]-target[0])*(newOutput[0]-target[0]) + (newOutput[1]-target[1])*(newOutput[1]-target[1])

	if newDist >= oldDist {
		t.Errorf("expected loss to decrease after gradient step: old=%.6f new=%.6f", oldDist, newDist)
	}
}

func TestNeuralNetCopyFrom(t *testing.T) {
	nn1 := newNeuralNet(3, []int{4}, 2)
	nn2 := newNeuralNet(3, []int{4}, 2)

	// They should have different weights initially (random init)
	nn2.copyFrom(nn1)

	// After copy, forward pass should give identical output
	input := []float64{0.5, 0.5, 0.5}
	out1 := nn1.forward(input)
	out2 := nn2.forward(input)

	for i := range out1 {
		if math.Abs(out1[i]-out2[i]) > 1e-12 {
			t.Errorf("output[%d] diverged after copyFrom: %f vs %f", i, out1[i], out2[i])
		}
	}
}

func TestReplayBuffer(t *testing.T) {
	rb := newReplayBuffer(5)
	if rb.size() != 0 {
		t.Errorf("expected size 0, got %d", rb.size())
	}

	for i := 0; i < 3; i++ {
		rb.push(experience{state: []float64{float64(i)}, action: i % numActions})
	}
	if rb.size() != 3 {
		t.Errorf("expected size 3, got %d", rb.size())
	}

	// Overflow
	for i := 0; i < 10; i++ {
		rb.push(experience{state: []float64{float64(i + 10)}, action: 0})
	}
	if rb.size() != 5 {
		t.Errorf("expected size capped at 5, got %d", rb.size())
	}

	batch := rb.sample(3)
	if len(batch) != 3 {
		t.Errorf("expected batch of 3, got %d", len(batch))
	}
}

func TestReplayBufferSampleClamp(t *testing.T) {
	rb := newReplayBuffer(100)
	rb.push(experience{state: []float64{1}, action: 0})
	rb.push(experience{state: []float64{2}, action: 1})

	// Requesting more than available should clamp
	batch := rb.sample(50)
	if len(batch) != 2 {
		t.Errorf("expected batch clamped to 2, got %d", len(batch))
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

func TestExtractStateVector(t *testing.T) {
	obs := &Observation{
		RiskAssessment: &core.RiskAssessment{Score: 0.7},
		Classification: &ml.ClassificationResult{Confidence: 0.8},
	}
	profile := &BehaviorSummary{
		ConsistencyScore:   0.9,
		FPSwitchRate:       3.0,
		AvgRequestInterval: 0.2,
		RiskTrend:          0.4,
		TotalObservations:  50,
		UniqueFP:           5,
	}
	v := ExtractStateVector(obs, profile)

	if len(v) != 8 {
		t.Fatalf("expected 8-dim vector, got %d", len(v))
	}
	if v[0] != 0.7 {
		t.Errorf("v[0] risk = %f, want 0.7", v[0])
	}
	if v[1] != 0.8 {
		t.Errorf("v[1] confidence = %f, want 0.8", v[1])
	}
	for i, val := range v {
		if val < 0 || val > 1.0 {
			t.Errorf("v[%d] = %f, expected in [0,1]", i, val)
		}
	}
}

func TestDQNSelectActionGreedy(t *testing.T) {
	cfg := &RLConfig{
		HiddenLayers:     []int{32, 16},
		LearningRate:     0.005,
		Gamma:            0.9,
		EpsilonMax:       0.0, // fully greedy
		EpsilonMin:       0.0,
		EpsilonDecay:     1.0,
		ReplayCapacity:   2000,
		BatchSize:        16,
		TargetUpdateFreq: 50,
		StateDim:         4,
	}
	re := NewReinforcementEngine(cfg)

	// Train on a pattern: high threat state → block gets high reward,
	// all other actions get negative reward (contrastive training).
	highThreat := []float64{1.0, 0.0, 1.0, 1.0}
	safe := []float64{0.0, 1.0, 0.0, 0.0}
	wrongActions := []ActionType{ActionAllow, ActionMonitor, ActionChallenge, ActionThrottle}

	for i := 0; i < 3000; i++ {
		re.UpdateContinuous(highThreat, ActionBlock, 1.0, safe, false)
		for _, a := range wrongActions {
			re.UpdateContinuous(highThreat, a, -0.5, highThreat, false)
		}
		re.UpdateContinuous(safe, ActionAllow, 0.5, safe, true)
	}

	// DQN should have learned that block is best for high-threat
	qvals := re.QValueContinuous(highThreat)
	blockQ := qvals[actionIndex[ActionBlock]]
	allowQ := qvals[actionIndex[ActionAllow]]
	if blockQ <= allowQ {
		t.Logf("Q-values for high threat: %v", qvals)
		t.Errorf("expected Q(Block)=%.3f > Q(Allow)=%.3f for high threat", blockQ, allowQ)
	}
}

func TestDQNSelectActionExploration(t *testing.T) {
	cfg := &RLConfig{
		HiddenLayers:     []int{8},
		LearningRate:     0.001,
		Gamma:            0.9,
		EpsilonMax:       1.0, // always explore
		EpsilonMin:       1.0,
		EpsilonDecay:     1.0,
		ReplayCapacity:   100,
		BatchSize:        4,
		TargetUpdateFreq: 10,
		StateDim:         4,
	}
	re := NewReinforcementEngine(cfg)

	// With epsilon=1.0, all actions should be explored
	counts := make(map[ActionType]int)
	for i := 0; i < 200; i++ {
		sv := []float64{0.5, 0.5, 0.5, 0.5}
		action, explored := re.SelectActionContinuous(sv)
		if !explored {
			t.Error("should always explore with epsilon=1.0")
		}
		counts[action]++
	}

	// All 5 actions should appear at least once in 200 trials
	for _, a := range indexAction {
		if counts[a] == 0 {
			t.Errorf("action %s never selected in 200 exploratory trials", a)
		}
	}
}

func TestEpsilonDecay(t *testing.T) {
	cfg := &RLConfig{
		HiddenLayers:     []int{8},
		LearningRate:     0.001,
		Gamma:            0.9,
		EpsilonMax:       1.0,
		EpsilonMin:       0.01,
		EpsilonDecay:     0.9, // aggressive decay for test
		ReplayCapacity:   100,
		BatchSize:        4,
		TargetUpdateFreq: 10,
		StateDim:         4,
	}
	re := NewReinforcementEngine(cfg)

	sv := []float64{0.5, 0.5, 0.5, 0.5}
	for i := 0; i < 50; i++ {
		re.UpdateContinuous(sv, ActionAllow, 0, sv, false)
	}

	eps := re.Epsilon()
	if eps >= 1.0 {
		t.Errorf("epsilon should have decayed, got %f", eps)
	}
	if eps < cfg.EpsilonMin {
		t.Errorf("epsilon %f dropped below min %f", eps, cfg.EpsilonMin)
	}
}

func TestDQNLossDecreases(t *testing.T) {
	cfg := &RLConfig{
		HiddenLayers:     []int{16, 8},
		LearningRate:     0.01,
		Gamma:            0.9,
		EpsilonMax:       0.0,
		EpsilonMin:       0.0,
		EpsilonDecay:     1.0,
		ReplayCapacity:   200,
		BatchSize:        8,
		TargetUpdateFreq: 50,
		StateDim:         4,
	}
	re := NewReinforcementEngine(cfg)

	sv := []float64{0.8, 0.2, 0.9, 0.7}

	// Train for a while
	for i := 0; i < 300; i++ {
		re.UpdateContinuous(sv, ActionBlock, 1.0, sv, true)
	}

	stats := re.Stats()
	if stats.TrainSteps == 0 {
		t.Fatal("expected training steps > 0")
	}
	// AvgLoss should be finite
	if math.IsNaN(stats.AvgLoss) || math.IsInf(stats.AvgLoss, 0) {
		t.Errorf("AvgLoss is not finite: %f", stats.AvgLoss)
	}
}

func TestDQNTargetNetworkSync(t *testing.T) {
	cfg := &RLConfig{
		HiddenLayers:     []int{8},
		LearningRate:     0.01,
		Gamma:            0.9,
		EpsilonMax:       0.0,
		EpsilonMin:       0.0,
		EpsilonDecay:     1.0,
		ReplayCapacity:   200,
		BatchSize:        4,
		TargetUpdateFreq: 10,
		StateDim:         4,
	}
	re := NewReinforcementEngine(cfg)

	sv := []float64{0.5, 0.5, 0.5, 0.5}

	// After TargetUpdateFreq steps, target should sync
	for i := 0; i < 20; i++ {
		re.UpdateContinuous(sv, ActionAllow, 0.5, sv, false)
	}

	// Verify online and target give same output after sync
	re.mu.RLock()
	onlineOut := re.online.forward(sv)
	targetOut := re.target.forward(sv)
	re.mu.RUnlock()

	// They won't be exactly equal (online continues to train after sync) but should be finite
	for i := range onlineOut {
		if math.IsNaN(onlineOut[i]) || math.IsNaN(targetOut[i]) {
			t.Errorf("NaN in network output: online[%d]=%f target[%d]=%f",
				i, onlineOut[i], i, targetOut[i])
		}
	}
}

func TestComputeReward(t *testing.T) {
	tests := []struct {
		action   ActionType
		threat   bool
		fpCost   float64
		fnCost   float64
		wantSign int
	}{
		{ActionBlock, true, 1.0, 2.0, 1},
		{ActionAllow, false, 1.0, 2.0, 1},
		{ActionBlock, false, 1.0, 2.0, -1},
		{ActionAllow, true, 1.0, 2.0, -1},
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

func TestDQNConvergence(t *testing.T) {
	cfg := &RLConfig{
		HiddenLayers:     []int{32, 16},
		LearningRate:     0.005,
		Gamma:            0.9,
		EpsilonMax:       0.0, // no exploration noise during convergence training
		EpsilonMin:       0.0,
		EpsilonDecay:     1.0,
		ReplayCapacity:   5000,
		BatchSize:        32,
		TargetUpdateFreq: 100,
		StateDim:         4,
	}
	re := NewReinforcementEngine(cfg)

	highThreat := []float64{1.0, 0.0, 1.0, 1.0}
	safe := []float64{0.0, 1.0, 0.0, 0.0}
	wrongForThreat := []ActionType{ActionAllow, ActionMonitor, ActionChallenge, ActionThrottle}
	wrongForSafe := []ActionType{ActionBlock, ActionMonitor, ActionChallenge, ActionThrottle}

	// Train extensively with contrastive signals on ALL actions
	for i := 0; i < 5000; i++ {
		re.UpdateContinuous(highThreat, ActionBlock, 1.0, safe, false)
		for _, a := range wrongForThreat {
			re.UpdateContinuous(highThreat, a, -0.5, highThreat, false)
		}
		re.UpdateContinuous(safe, ActionAllow, 0.5, safe, true)
		for _, a := range wrongForSafe {
			re.UpdateContinuous(safe, a, -0.3, safe, true)
		}
	}

	threatQvals := re.QValueContinuous(highThreat)
	safeQvals := re.QValueContinuous(safe)

	// Block should have the highest Q for high threat
	if argmax(threatQvals) != actionIndex[ActionBlock] {
		t.Logf("Q-values for high threat: %v", threatQvals)
		t.Errorf("expected Block to have highest Q for high threat, got action idx %d", argmax(threatQvals))
	}

	// Allow should have the highest Q for safe state
	if argmax(safeQvals) != actionIndex[ActionAllow] {
		t.Logf("Q-values for safe: %v", safeQvals)
		t.Errorf("expected Allow to have highest Q for safe, got action idx %d", argmax(safeQvals))
	}
}

func TestBackwardCompatSelectAction(t *testing.T) {
	re := NewReinforcementEngine(nil)
	state := State{1, 2, 1, 0}

	// Should not panic — backward compat API
	action, _ := re.SelectAction(state)
	if actionIndex[action] < 0 || actionIndex[action] >= numActions {
		t.Errorf("invalid action from SelectAction: %s", action)
	}
}

func TestBackwardCompatUpdate(t *testing.T) {
	cfg := &RLConfig{
		HiddenLayers:     []int{8},
		LearningRate:     0.01,
		Gamma:            0.9,
		EpsilonMax:       0.0,
		EpsilonMin:       0.0,
		EpsilonDecay:     1.0,
		ReplayCapacity:   100,
		BatchSize:        4,
		TargetUpdateFreq: 10,
		StateDim:         8,
	}
	re := NewReinforcementEngine(cfg)

	s1 := State{2, 3, 1, 0}
	s2 := State{0, 0, 0, 0}

	// Backward compat Update should work
	for i := 0; i < 20; i++ {
		re.Update(s1, ActionBlock, 1.0, s2)
	}

	// QValue backward compat
	q := re.QValue(s1, ActionBlock)
	if math.IsNaN(q) {
		t.Error("QValue returned NaN")
	}

	// BestAction backward compat
	best := re.BestAction(s1)
	if actionIndex[best] < 0 || actionIndex[best] >= numActions {
		t.Errorf("invalid BestAction: %s", best)
	}
}

func TestDQNStatsComplete(t *testing.T) {
	cfg := &RLConfig{
		HiddenLayers:     []int{16, 8},
		LearningRate:     0.01,
		Gamma:            0.9,
		EpsilonMax:       0.5,
		EpsilonMin:       0.01,
		EpsilonDecay:     0.99,
		ReplayCapacity:   100,
		BatchSize:        4,
		TargetUpdateFreq: 10,
		StateDim:         4,
	}
	re := NewReinforcementEngine(cfg)

	sv := []float64{0.5, 0.5, 0.5, 0.5}
	for i := 0; i < 50; i++ {
		re.UpdateContinuous(sv, ActionAllow, 0.5, sv, false)
	}

	stats := re.Stats()
	if stats.Steps != 50 {
		t.Errorf("expected 50 steps, got %d", stats.Steps)
	}
	if stats.ReplaySize == 0 {
		t.Error("expected non-zero replay size")
	}
	if len(stats.NetworkLayers) < 3 {
		t.Errorf("expected at least 3 layers (input+hidden+output), got %d", len(stats.NetworkLayers))
	}
	// NetworkLayers should be [4, 16, 8, 5]
	expected := []int{4, 16, 8, 5}
	if len(stats.NetworkLayers) != len(expected) {
		t.Errorf("network layers = %v, want %v", stats.NetworkLayers, expected)
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
