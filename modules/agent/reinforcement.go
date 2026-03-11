package agent

import (
	"math"
	"math/rand/v2"
	"sync"
)

// RLConfig configures the reinforcement learning engine.
type RLConfig struct {
	Alpha      float64 // Learning rate (0,1], default 0.1
	Gamma      float64 // Discount factor [0,1], default 0.95
	EpsilonMax float64 // Initial exploration rate, default 0.3
	EpsilonMin float64 // Minimum exploration rate after decay, default 0.01
	EpsilonDecay float64 // Decay per episode, default 0.995
}

// DefaultRLConfig provides sensible defaults.
var DefaultRLConfig = &RLConfig{
	Alpha:        0.1,
	Gamma:        0.95,
	EpsilonMax:   0.3,
	EpsilonMin:   0.01,
	EpsilonDecay: 0.995,
}

// State represents a discretized agent state for Q-learning.
// Discretized from continuous observations to keep the state space manageable.
type State struct {
	ThreatBucket      int // 0=none, 1=low, 2=medium, 3=high
	RiskBucket        int // 0=safe, 1=elevated, 2=high, 3=critical
	ConsistencyBucket int // 0=solid, 1=moderate, 2=weak, 3=anomalous
	SwitchRateBucket  int // 0=stable, 1=moderate, 2=frequent, 3=rapid
}

// stateKey returns a compact integer key for map lookup.
func (s State) stateKey() int {
	return s.ThreatBucket*64 + s.RiskBucket*16 + s.ConsistencyBucket*4 + s.SwitchRateBucket
}

// actionIndex maps ActionType to a numeric index.
var actionIndex = map[ActionType]int{
	ActionAllow:     0,
	ActionMonitor:   1,
	ActionChallenge: 2,
	ActionThrottle:  3,
	ActionBlock:     4,
}

// indexAction maps numeric index back to ActionType.
var indexAction = [5]ActionType{
	ActionAllow,
	ActionMonitor,
	ActionChallenge,
	ActionThrottle,
	ActionBlock,
}

const numActions = 5

// ReinforcementEngine implements tabular Q-learning for the agent's action selection.
//
// The agent discretizes its observations into a finite State and uses a Q-table
// to learn the expected cumulative reward for each (state, action) pair.
// Epsilon-greedy exploration ensures the agent discovers better policies
// while exploiting what it has already learned.
type ReinforcementEngine struct {
	config  *RLConfig
	qtable  map[int][numActions]float64 // stateKey → Q-values per action
	epsilon float64
	episode int64
	mu      sync.RWMutex
}

// NewReinforcementEngine creates a new Q-learning engine.
func NewReinforcementEngine(cfg *RLConfig) *ReinforcementEngine {
	if cfg == nil {
		cfg = DefaultRLConfig
	}
	return &ReinforcementEngine{
		config:  cfg,
		qtable:  make(map[int][numActions]float64),
		epsilon: cfg.EpsilonMax,
	}
}

// DiscretizeState converts continuous observation signals into a discrete State.
func DiscretizeState(obs *Observation, profile *BehaviorSummary) State {
	s := State{}

	// Threat bucket from risk assessment
	if obs.RiskAssessment != nil {
		switch {
		case obs.RiskAssessment.Score > 0.75:
			s.RiskBucket = 3
		case obs.RiskAssessment.Score > 0.5:
			s.RiskBucket = 2
		case obs.RiskAssessment.Score > 0.25:
			s.RiskBucket = 1
		}
	}

	// Threat class bucket from classification confidence
	if obs.Classification != nil {
		switch {
		case obs.Classification.Confidence < 0.3:
			s.ThreatBucket = 3
		case obs.Classification.Confidence < 0.5:
			s.ThreatBucket = 2
		case obs.Classification.Confidence < 0.7:
			s.ThreatBucket = 1
		}
	}

	if profile != nil {
		// Consistency bucket
		switch {
		case profile.ConsistencyScore < 0.2:
			s.ConsistencyBucket = 3
		case profile.ConsistencyScore < 0.4:
			s.ConsistencyBucket = 2
		case profile.ConsistencyScore < 0.7:
			s.ConsistencyBucket = 1
		}

		// Switch rate bucket
		switch {
		case profile.FPSwitchRate > 5.0:
			s.SwitchRateBucket = 3
		case profile.FPSwitchRate > 2.0:
			s.SwitchRateBucket = 2
		case profile.FPSwitchRate > 0.5:
			s.SwitchRateBucket = 1
		}
	}

	return s
}

// SelectAction chooses an action using epsilon-greedy policy.
// Returns the selected action and whether it was an exploratory choice.
func (re *ReinforcementEngine) SelectAction(state State) (ActionType, bool) {
	re.mu.RLock()
	defer re.mu.RUnlock()

	// Epsilon-greedy: explore with probability epsilon
	if rand.Float64() < re.epsilon {
		return indexAction[rand.IntN(numActions)], true
	}

	// Exploit: choose action with highest Q-value
	key := state.stateKey()
	qvals := re.qtable[key]
	bestIdx := 0
	bestQ := qvals[0]
	for i := 1; i < numActions; i++ {
		if qvals[i] > bestQ {
			bestQ = qvals[i]
			bestIdx = i
		}
	}
	return indexAction[bestIdx], false
}

// Update performs a Q-learning update: Q(s,a) += α * [reward + γ*max_a'(Q(s',a')) - Q(s,a)]
func (re *ReinforcementEngine) Update(state State, action ActionType, reward float64, nextState State) {
	re.mu.Lock()
	defer re.mu.Unlock()

	key := state.stateKey()
	nextKey := nextState.stateKey()
	aIdx := actionIndex[action]

	// Current Q-value
	qvals := re.qtable[key]
	currentQ := qvals[aIdx]

	// Max Q-value for next state
	nextQvals := re.qtable[nextKey]
	maxNextQ := nextQvals[0]
	for i := 1; i < numActions; i++ {
		if nextQvals[i] > maxNextQ {
			maxNextQ = nextQvals[i]
		}
	}

	// Q-learning update
	tdTarget := reward + re.config.Gamma*maxNextQ
	qvals[aIdx] = currentQ + re.config.Alpha*(tdTarget-currentQ)
	re.qtable[key] = qvals

	// Decay epsilon
	re.episode++
	re.epsilon = math.Max(re.config.EpsilonMin, re.epsilon*re.config.EpsilonDecay)
}

// ComputeReward calculates a scalar reward from the outcome of an action.
//
// Positive reward: correct action (e.g., blocking a confirmed threat, allowing legitimate traffic).
// Negative reward: wrong action (e.g., blocking legitimate users, allowing confirmed threats).
func ComputeReward(action ActionType, wasActualThreat bool, falsePositiveCost, falseNegativeCost float64) float64 {
	blocked := action == ActionBlock || action == ActionThrottle || action == ActionChallenge

	switch {
	case blocked && wasActualThreat:
		return 1.0 // True positive: correctly blocked threat
	case !blocked && !wasActualThreat:
		return 0.5 // True negative: correctly allowed legitimate traffic
	case blocked && !wasActualThreat:
		return -falsePositiveCost // False positive: blocked legitimate user
	default: // !blocked && wasActualThreat
		return -falseNegativeCost // False negative: missed a threat
	}
}

// QValue returns the current Q-value for a (state, action) pair.
func (re *ReinforcementEngine) QValue(state State, action ActionType) float64 {
	re.mu.RLock()
	defer re.mu.RUnlock()
	return re.qtable[state.stateKey()][actionIndex[action]]
}

// BestAction returns the greedy (exploitation-only) action for a state.
func (re *ReinforcementEngine) BestAction(state State) ActionType {
	re.mu.RLock()
	defer re.mu.RUnlock()

	qvals := re.qtable[state.stateKey()]
	bestIdx := 0
	for i := 1; i < numActions; i++ {
		if qvals[i] > qvals[bestIdx] {
			bestIdx = i
		}
	}
	return indexAction[bestIdx]
}

// Epsilon returns the current exploration rate.
func (re *ReinforcementEngine) Epsilon() float64 {
	re.mu.RLock()
	defer re.mu.RUnlock()
	return re.epsilon
}

// Stats returns Q-table statistics.
func (re *ReinforcementEngine) Stats() RLStats {
	re.mu.RLock()
	defer re.mu.RUnlock()
	return RLStats{
		StatesExplored: len(re.qtable),
		Episodes:       re.episode,
		Epsilon:        re.epsilon,
	}
}

// RLStats contains reinforcement learning runtime statistics.
type RLStats struct {
	StatesExplored int     `json:"states_explored"`
	Episodes       int64   `json:"episodes"`
	Epsilon        float64 `json:"epsilon"`
}
