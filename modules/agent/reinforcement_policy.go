package agent

import (
	"math"
	"math/rand/v2"
	"sync"
)

func (nn *neuralNet) copyFrom(src *neuralNet) {
	for l := range nn.layers {
		for i := range nn.layers[l].weights {
			copy(nn.layers[l].weights[i], src.layers[l].weights[i])
		}
		copy(nn.layers[l].biases, src.layers[l].biases)
	}
}

// ---------------------------------------------------------------------------
// Experience Replay Buffer
// ---------------------------------------------------------------------------

// experience stores a single (s, a, r, s', done) transition.
type experience struct {
	state     []float64
	action    int
	reward    float64
	nextState []float64
	done      bool
}

// replayBuffer is a circular buffer of experiences for off-policy learning.
type replayBuffer struct {
	buf  []experience
	cap  int
	pos  int
	full bool
}

func newReplayBuffer(capacity int) *replayBuffer {
	return &replayBuffer{
		buf: make([]experience, capacity),
		cap: capacity,
	}
}

func (rb *replayBuffer) push(e experience) {
	rb.buf[rb.pos] = e
	rb.pos++
	if rb.pos >= rb.cap {
		rb.pos = 0
		rb.full = true
	}
}

func (rb *replayBuffer) size() int {
	if rb.full {
		return rb.cap
	}
	return rb.pos
}

// sample returns a random mini-batch of the given size.
func (rb *replayBuffer) sample(batchSize int) []experience {
	n := rb.size()
	if batchSize > n {
		batchSize = n
	}
	batch := make([]experience, batchSize)
	for i := 0; i < batchSize; i++ {
		batch[i] = rb.buf[rand.IntN(n)]
	}
	return batch
}

// ---------------------------------------------------------------------------
// Deep Q-Network Engine
// ---------------------------------------------------------------------------

// ReinforcementEngine implements a Deep Q-Network (DQN) for action-value estimation.
//
// Architecture:
//   - Online network: trained via mini-batch SGD on experience replay samples
//   - Target network: periodically synced copy used for stable TD target computation
//   - Experience replay buffer: breaks temporal correlation in training data
//   - Epsilon-greedy exploration with multiplicative decay
//
// The state is a continuous float vector (default 8-dim) extracted from observations,
// replacing the old discrete Q-table approach with a neural function approximator
// that can generalize across similar states.
type ReinforcementEngine struct {
	config     *RLConfig
	online     *neuralNet    // policy network (trained every step)
	target     *neuralNet    // target network (synced periodically)
	replay     *replayBuffer // experience replay buffer
	epsilon    float64
	step       int64
	totalLoss  float64 // accumulated training loss (for monitoring)
	trainSteps int64   // number of training mini-batches executed
	mu         sync.RWMutex
}

// NewReinforcementEngine creates a DQN engine with the given configuration.
func NewReinforcementEngine(cfg *RLConfig) *ReinforcementEngine {
	if cfg == nil {
		cfg = DefaultRLConfig
	}
	// Backward compat: if Alpha is set but LearningRate is default, use Alpha
	if cfg.LearningRate == 0 {
		cfg.LearningRate = cfg.Alpha
	}
	if cfg.LearningRate == 0 {
		cfg.LearningRate = 0.001
	}
	if len(cfg.HiddenLayers) == 0 {
		cfg.HiddenLayers = []int{64, 32}
	}
	if cfg.StateDim == 0 {
		cfg.StateDim = 8
	}
	if cfg.ReplayCapacity == 0 {
		cfg.ReplayCapacity = 10000
	}
	if cfg.BatchSize == 0 {
		cfg.BatchSize = 32
	}
	if cfg.TargetUpdateFreq == 0 {
		cfg.TargetUpdateFreq = 100
	}

	online := newNeuralNet(cfg.StateDim, cfg.HiddenLayers, numActions)
	target := newNeuralNet(cfg.StateDim, cfg.HiddenLayers, numActions)
	target.copyFrom(online)

	return &ReinforcementEngine{
		config:  cfg,
		online:  online,
		target:  target,
		replay:  newReplayBuffer(cfg.ReplayCapacity),
		epsilon: cfg.EpsilonMax,
	}
}

// ---------------------------------------------------------------------------
// Public API — SelectAction / Update / BestAction / QValue (preserves old interface)
// ---------------------------------------------------------------------------

// SelectAction chooses an action using epsilon-greedy over the DQN Q-values.
// Accepts a discrete State for backward compatibility (converts to vector internally).
func (re *ReinforcementEngine) SelectAction(state State) (ActionType, bool) {
	re.mu.RLock()
	defer re.mu.RUnlock()

	if rand.Float64() < re.epsilon {
		return indexAction[rand.IntN(numActions)], true
	}

	sv := state.toVector(re.config.StateDim)
	qvals := re.online.forward(sv)
	return indexAction[argmax(qvals)], false
}

// SelectActionContinuous chooses an action from a continuous state vector.
// This is the preferred API for DQN — avoids discretization loss.
func (re *ReinforcementEngine) SelectActionContinuous(stateVec []float64) (ActionType, bool) {
	re.mu.RLock()
	defer re.mu.RUnlock()

	if rand.Float64() < re.epsilon {
		return indexAction[rand.IntN(numActions)], true
	}

	qvals := re.online.forward(re.padState(stateVec))
	return indexAction[argmax(qvals)], false
}

// Update stores a transition and trains the DQN on a mini-batch.
// Accepts discrete State for backward compatibility.
func (re *ReinforcementEngine) Update(state State, action ActionType, reward float64, nextState State) {
	sv := state.toVector(re.config.StateDim)
	nsv := nextState.toVector(re.config.StateDim)
	re.UpdateContinuous(sv, action, reward, nsv, false)
}

// UpdateContinuous stores a (s, a, r, s', done) transition in the experience replay
// buffer and trains the online network on a mini-batch sample.
//
// Training uses the DQN loss:
//
//	L = (r + γ·max_a'[Q_target(s', a')] − Q_online(s, a))²
//
// The target network is synchronized every TargetUpdateFreq steps.
func (re *ReinforcementEngine) UpdateContinuous(stateVec []float64, action ActionType, reward float64, nextStateVec []float64, done bool) {
	re.mu.Lock()
	defer re.mu.Unlock()

	sv := re.padState(stateVec)
	nsv := re.padState(nextStateVec)

	// Store experience
	re.replay.push(experience{
		state:     sv,
		action:    actionIndex[action],
		reward:    reward,
		nextState: nsv,
		done:      done,
	})

	re.step++

	// Decay epsilon
	re.epsilon = math.Max(re.config.EpsilonMin, re.epsilon*re.config.EpsilonDecay)

	// Wait until we have enough experiences to form a meaningful batch
	if re.replay.size() < re.config.BatchSize {
		return
	}

	// Sample mini-batch and train
	batch := re.replay.sample(re.config.BatchSize)
	re.trainBatch(batch)

	// Periodically sync target network
	if re.step%int64(re.config.TargetUpdateFreq) == 0 {
		re.target.copyFrom(re.online)
	}
}

// trainBatch performs one SGD step on a mini-batch of experiences.
func (re *ReinforcementEngine) trainBatch(batch []experience) {
	lr := re.config.LearningRate
	gamma := re.config.Gamma
	batchLoss := 0.0

	for _, exp := range batch {
		// Forward pass through online network to get current Q(s, ·)
		qvals := re.online.forward(exp.state)

		// Compute TD target using target network
		var tdTarget float64
		if exp.done {
			tdTarget = exp.reward
		} else {
			targetQvals := re.target.forward(exp.nextState)
			tdTarget = exp.reward + gamma*sliceMax(targetQvals)
		}

		// Compute loss gradient: only the chosen action has non-zero gradient
		// dL/dQ[a] = 2 * (Q[a] - target) / batchSize  (MSE gradient)
		dLoss := make([]float64, numActions)
		td_error := qvals[exp.action] - tdTarget
		dLoss[exp.action] = 2.0 * td_error / float64(len(batch))

		// Clamp gradient to prevent exploding gradients
		if dLoss[exp.action] > 1.0 {
			dLoss[exp.action] = 1.0
		} else if dLoss[exp.action] < -1.0 {
			dLoss[exp.action] = -1.0
		}

		batchLoss += td_error * td_error

		// Backward pass + weight update
		re.online.backward(dLoss, lr)
	}

	re.totalLoss += batchLoss / float64(len(batch))
	re.trainSteps++
}

// QValue returns the DQN's estimated Q-value for a (state, action) pair.
func (re *ReinforcementEngine) QValue(state State, action ActionType) float64 {
	re.mu.RLock()
	defer re.mu.RUnlock()
	sv := state.toVector(re.config.StateDim)
	qvals := re.online.forward(sv)
	return qvals[actionIndex[action]]
}

// QValueContinuous returns Q-values for a continuous state vector.
func (re *ReinforcementEngine) QValueContinuous(stateVec []float64) []float64 {
	re.mu.RLock()
	defer re.mu.RUnlock()
	qvals := re.online.forward(re.padState(stateVec))
	result := make([]float64, numActions)
	copy(result, qvals)
	return result
}

// BestAction returns the greedy action for a discrete state (no exploration).
func (re *ReinforcementEngine) BestAction(state State) ActionType {
	re.mu.RLock()
	defer re.mu.RUnlock()
	sv := state.toVector(re.config.StateDim)
	qvals := re.online.forward(sv)
	return indexAction[argmax(qvals)]
}

// BestActionContinuous returns the greedy action for a continuous state vector.
func (re *ReinforcementEngine) BestActionContinuous(stateVec []float64) ActionType {
	re.mu.RLock()
	defer re.mu.RUnlock()
	qvals := re.online.forward(re.padState(stateVec))
	return indexAction[argmax(qvals)]
}

// Epsilon returns the current exploration rate.
func (re *ReinforcementEngine) Epsilon() float64 {
	re.mu.RLock()
	defer re.mu.RUnlock()
	return re.epsilon
}

// Stats returns DQN training statistics.
func (re *ReinforcementEngine) Stats() RLStats {
	re.mu.RLock()
	defer re.mu.RUnlock()

	avgLoss := 0.0
	if re.trainSteps > 0 {
		avgLoss = re.totalLoss / float64(re.trainSteps)
	}
	return RLStats{
		Steps:         re.step,
		TrainSteps:    re.trainSteps,
		Epsilon:       re.epsilon,
		AvgLoss:       avgLoss,
		ReplaySize:    re.replay.size(),
		NetworkLayers: re.describeNetwork(),
		// Backward compat fields
		StatesExplored: re.replay.size(),
		Episodes:       re.step,
	}
}

// describeNetwork returns a human-readable description of the DQN architecture.
func (re *ReinforcementEngine) describeNetwork() []int {
	dims := make([]int, 0, len(re.online.layers)+1)
	if len(re.online.layers) > 0 {
		dims = append(dims, len(re.online.layers[0].weights[0])) // input dim
	}
	for _, layer := range re.online.layers {
		dims = append(dims, len(layer.biases))
	}
	return dims
}

// RLStats contains DQN reinforcement learning runtime statistics.
type RLStats struct {
	Steps         int64   `json:"steps"`
	TrainSteps    int64   `json:"train_steps"`
	Epsilon       float64 `json:"epsilon"`
	AvgLoss       float64 `json:"avg_loss"`
	ReplaySize    int     `json:"replay_size"`
	NetworkLayers []int   `json:"network_layers"`
	// Backward compatibility
	StatesExplored int   `json:"states_explored"`
	Episodes       int64 `json:"episodes"`
}

// ---------------------------------------------------------------------------
// Reward computation (unchanged)
// ---------------------------------------------------------------------------

// ComputeReward calculates a scalar reward from the outcome of an action.
func ComputeReward(action ActionType, wasActualThreat bool, falsePositiveCost, falseNegativeCost float64) float64 {
	blocked := action == ActionBlock || action == ActionThrottle || action == ActionChallenge

	switch {
	case blocked && wasActualThreat:
		return 1.0
	case !blocked && !wasActualThreat:
		return 0.5
	case blocked && !wasActualThreat:
		return -falsePositiveCost
	default:
		return -falseNegativeCost
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (re *ReinforcementEngine) padState(v []float64) []float64 {
	d := re.config.StateDim
	if len(v) >= d {
		return v[:d]
	}
	padded := make([]float64, d)
	copy(padded, v)
	return padded
}

func argmax(s []float64) int {
	best := 0
	for i := 1; i < len(s); i++ {
		if s[i] > s[best] {
			best = i
		}
	}
	return best
}

func sliceMax(s []float64) float64 {
	m := s[0]
	for _, v := range s[1:] {
		if v > m {
			m = v
		}
	}
	return m
}
