package agent

import (
	"math"
	"math/rand/v2"
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
