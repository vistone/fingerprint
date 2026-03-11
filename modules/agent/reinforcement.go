package agent

import (
	"math"
	"math/rand/v2"
	"sync"
)

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

// RLConfig configures the Deep Q-Network reinforcement learning engine.
type RLConfig struct {
	// Neural network architecture
	HiddenLayers []int   // Hidden layer sizes, default [64, 32]
	LearningRate float64 // SGD learning rate, default 0.001
	// RL parameters
	Gamma        float64 // Discount factor [0,1], default 0.95
	EpsilonMax   float64 // Initial exploration rate, default 0.3
	EpsilonMin   float64 // Minimum exploration rate, default 0.01
	EpsilonDecay float64 // Multiplicative decay per step, default 0.995
	// DQN-specific
	ReplayCapacity   int // Experience replay buffer capacity, default 10000
	BatchSize        int // Mini-batch size for training, default 32
	TargetUpdateFreq int // Copy weights to target network every N steps, default 100
	StateDim         int // Continuous state dimension, default 8
	// Compatibility
	Alpha float64 // Alias for LearningRate (kept for config backward compat)
}

// DefaultRLConfig provides sensible defaults for DQN.
var DefaultRLConfig = &RLConfig{
	HiddenLayers:     []int{64, 32},
	LearningRate:     0.001,
	Gamma:            0.95,
	EpsilonMax:       0.3,
	EpsilonMin:       0.01,
	EpsilonDecay:     0.995,
	ReplayCapacity:   10000,
	BatchSize:        32,
	TargetUpdateFreq: 100,
	StateDim:         8,
	Alpha:            0.001,
}

// ---------------------------------------------------------------------------
// Action mapping (unchanged public API)
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// State (kept for backward compat; DQN uses continuous vectors internally)
// ---------------------------------------------------------------------------

// State represents a discretized agent state. The DQN engine converts this to a
// continuous vector internally but the struct is preserved for API compatibility.
type State struct {
	ThreatBucket      int
	RiskBucket        int
	ConsistencyBucket int
	SwitchRateBucket  int
}

// stateKey returns a compact integer key (legacy, used only for stats).
func (s State) stateKey() int {
	return s.ThreatBucket*64 + s.RiskBucket*16 + s.ConsistencyBucket*4 + s.SwitchRateBucket
}

// toVector converts a discrete State into a continuous float vector for the neural network.
func (s State) toVector(dim int) []float64 {
	v := make([]float64, dim)
	if dim > 0 {
		v[0] = float64(s.ThreatBucket) / 3.0
	}
	if dim > 1 {
		v[1] = float64(s.RiskBucket) / 3.0
	}
	if dim > 2 {
		v[2] = float64(s.ConsistencyBucket) / 3.0
	}
	if dim > 3 {
		v[3] = float64(s.SwitchRateBucket) / 3.0
	}
	return v
}

// DiscretizeState converts continuous observation signals into a discrete State.
func DiscretizeState(obs *Observation, profile *BehaviorSummary) State {
	s := State{}

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
		switch {
		case profile.ConsistencyScore < 0.2:
			s.ConsistencyBucket = 3
		case profile.ConsistencyScore < 0.4:
			s.ConsistencyBucket = 2
		case profile.ConsistencyScore < 0.7:
			s.ConsistencyBucket = 1
		}

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

// ExtractStateVector builds a continuous state vector directly from raw observation
// signals, bypassing discretization. This is the preferred input for the DQN.
func ExtractStateVector(obs *Observation, profile *BehaviorSummary) []float64 {
	v := make([]float64, 8)
	if obs.RiskAssessment != nil {
		v[0] = obs.RiskAssessment.Score
	}
	if obs.Classification != nil {
		v[1] = obs.Classification.Confidence
	}
	if profile != nil {
		v[2] = profile.ConsistencyScore
		v[3] = math.Min(profile.FPSwitchRate/10.0, 1.0)
		if profile.AvgRequestInterval > 0 {
			v[4] = math.Min(1.0/profile.AvgRequestInterval/20.0, 1.0)
		}
		v[5] = (profile.RiskTrend + 1.0) / 2.0
		v[6] = math.Min(float64(profile.TotalObservations)/100.0, 1.0)
		if profile.TotalObservations > 0 {
			v[7] = float64(profile.UniqueFP) / float64(profile.TotalObservations)
		}
	}
	return v
}

// ---------------------------------------------------------------------------
// Neural Network — Multi-layer Perceptron
// ---------------------------------------------------------------------------

// neuralNet is a fully connected feedforward network with ReLU hidden activations
// and linear output layer. Weights are trained via backpropagation (SGD).
type neuralNet struct {
	layers []nnLayer // layers[0] = first hidden, layers[len-1] = output
}

// nnLayer stores weights, biases, and cached activations for one layer.
type nnLayer struct {
	// weights[i][j] = weight from input j to neuron i
	weights [][]float64
	biases  []float64
	// Cached during forward pass for backprop
	input  []float64 // input to this layer
	preAct []float64 // pre-activation (z = Wx + b)
	output []float64 // post-activation (a = relu(z) or linear for output)
	relu   bool      // true for hidden layers
}

// newNeuralNet creates a feedforward net: inputDim → hidden1 → hidden2 → … → outputDim.
// Weights are initialized with He initialization (√(2/fan_in)).
func newNeuralNet(inputDim int, hiddenLayers []int, outputDim int) *neuralNet {
	sizes := make([]int, 0, len(hiddenLayers)+2)
	sizes = append(sizes, inputDim)
	sizes = append(sizes, hiddenLayers...)
	sizes = append(sizes, outputDim)

	nn := &neuralNet{layers: make([]nnLayer, len(sizes)-1)}
	for l := 0; l < len(sizes)-1; l++ {
		fanIn := sizes[l]
		fanOut := sizes[l+1]
		scale := math.Sqrt(2.0 / float64(fanIn)) // He init

		layer := nnLayer{
			weights: make([][]float64, fanOut),
			biases:  make([]float64, fanOut),
			relu:    l < len(sizes)-2, // last layer is linear
		}
		for i := 0; i < fanOut; i++ {
			layer.weights[i] = make([]float64, fanIn)
			for j := 0; j < fanIn; j++ {
				layer.weights[i][j] = rand.NormFloat64() * scale
			}
		}
		nn.layers[l] = layer
	}
	return nn
}

// forward computes the network output for a given input. Caches intermediate
// values for backpropagation.
func (nn *neuralNet) forward(input []float64) []float64 {
	x := input
	for l := range nn.layers {
		layer := &nn.layers[l]
		layer.input = make([]float64, len(x))
		copy(layer.input, x)

		n := len(layer.biases)
		layer.preAct = make([]float64, n)
		layer.output = make([]float64, n)

		for i := 0; i < n; i++ {
			sum := layer.biases[i]
			for j := 0; j < len(x); j++ {
				sum += layer.weights[i][j] * x[j]
			}
			layer.preAct[i] = sum
			if layer.relu {
				layer.output[i] = math.Max(0, sum)
			} else {
				layer.output[i] = sum // linear output
			}
		}
		x = layer.output
	}
	return x
}

// backward computes gradients and updates weights via SGD given the loss gradient
// dLoss_dOutput (partial derivatives of loss w.r.t. each output neuron).
func (nn *neuralNet) backward(dLoss_dOutput []float64, lr float64) {
	delta := dLoss_dOutput

	for l := len(nn.layers) - 1; l >= 0; l-- {
		layer := &nn.layers[l]
		n := len(layer.biases)

		// Compute gradient through activation
		grad := make([]float64, n)
		for i := 0; i < n; i++ {
			if layer.relu {
				if layer.preAct[i] > 0 {
					grad[i] = delta[i]
				} // else 0 (ReLU derivative)
			} else {
				grad[i] = delta[i] // linear
			}
		}

		// Propagate delta to previous layer
		if l > 0 {
			prevSize := len(layer.input)
			newDelta := make([]float64, prevSize)
			for j := 0; j < prevSize; j++ {
				for i := 0; i < n; i++ {
					newDelta[j] += grad[i] * layer.weights[i][j]
				}
			}
			delta = newDelta
		}

		// Update weights and biases (SGD)
		for i := 0; i < n; i++ {
			for j := 0; j < len(layer.input); j++ {
				layer.weights[i][j] -= lr * grad[i] * layer.input[j]
			}
			layer.biases[i] -= lr * grad[i]
		}
	}
}

// copyFrom copies all weights and biases from another network (for target network sync).
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
