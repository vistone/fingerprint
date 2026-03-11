package agent

import (
	"math"
	"math/rand/v2"
	"sync"
)

// BanditConfig configures the contextual bandit strategy selector.
type BanditConfig struct {
	Alpha     float64 // UCB exploration parameter, default 1.0
	Dimension int     // Context feature vector dimension, default 8
}

// DefaultBanditConfig provides sensible defaults.
var DefaultBanditConfig = &BanditConfig{
	Alpha:     1.0,
	Dimension: 8,
}

// ContextualBandit implements the LinUCB (Linear Upper Confidence Bound) algorithm
// for strategy selection in the agent.
//
// Each strategy is an "arm". The bandit maintains a linear model per arm that predicts
// the expected reward given a context vector. The UCB term encourages exploration of
// arms whose reward is uncertain.
//
// Context vector is constructed from the observation + behavior profile:
//
//	[risk_score, ml_confidence, consistency_score, fp_switch_rate,
//	 avg_interval, risk_trend, suspicion_score, observation_count_norm]
type ContextualBandit struct {
	arms   map[string]*linUCBArm // arm ID (strategy ID) → model
	config *BanditConfig
	mu     sync.RWMutex
}

// linUCBArm stores the per-arm model for LinUCB.
type linUCBArm struct {
	// A = d×d matrix (accumulated context outer products + identity)
	A [][]float64
	// b = d-dimensional vector (accumulated reward-weighted contexts)
	b []float64
	// Statistics
	pulls   int
	rewards float64
}

// NewContextualBandit creates a new LinUCB bandit.
func NewContextualBandit(cfg *BanditConfig) *ContextualBandit {
	if cfg == nil {
		cfg = DefaultBanditConfig
	}
	return &ContextualBandit{
		arms:   make(map[string]*linUCBArm),
		config: cfg,
	}
}

// newArm creates a new LinUCB arm with identity matrix A and zero vector b.
func (cb *ContextualBandit) newArm() *linUCBArm {
	d := cb.config.Dimension
	A := make([][]float64, d)
	for i := range A {
		A[i] = make([]float64, d)
		A[i][i] = 1.0 // Identity matrix
	}
	b := make([]float64, d)
	return &linUCBArm{A: A, b: b}
}

// RegisterArm registers a strategy as an arm in the bandit.
func (cb *ContextualBandit) RegisterArm(strategyID string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if _, ok := cb.arms[strategyID]; !ok {
		cb.arms[strategyID] = cb.newArm()
	}
}

// SelectArm selects the best strategy given a context using LinUCB.
// Returns the selected arm ID and its UCB score.
func (cb *ContextualBandit) SelectArm(context []float64) (string, float64) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if len(cb.arms) == 0 {
		return "", 0
	}

	ctx := cb.padContext(context)

	bestArm := ""
	bestUCB := math.Inf(-1)

	for id, arm := range cb.arms {
		ucb := cb.computeUCB(arm, ctx)
		if ucb > bestUCB {
			bestUCB = ucb
			bestArm = id
		}
	}

	return bestArm, bestUCB
}

// SelectArmAmong selects from a specific set of candidate strategy IDs.
func (cb *ContextualBandit) SelectArmAmong(context []float64, candidates []string) (string, float64) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if len(candidates) == 0 {
		return "", 0
	}

	ctx := cb.padContext(context)

	bestArm := candidates[0]
	bestUCB := math.Inf(-1)

	for _, id := range candidates {
		arm, ok := cb.arms[id]
		if !ok {
			continue
		}
		ucb := cb.computeUCB(arm, ctx)
		if ucb > bestUCB {
			bestUCB = ucb
			bestArm = id
		}
	}

	return bestArm, bestUCB
}

// UpdateReward provides feedback for a selected arm in a given context.
func (cb *ContextualBandit) UpdateReward(strategyID string, context []float64, reward float64) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	arm, ok := cb.arms[strategyID]
	if !ok {
		arm = cb.newArm()
		cb.arms[strategyID] = arm
	}

	ctx := cb.padContext(context)

	// A = A + x*x^T
	for i := 0; i < len(ctx); i++ {
		for j := 0; j < len(ctx); j++ {
			arm.A[i][j] += ctx[i] * ctx[j]
		}
	}

	// b = b + reward * x
	for i := 0; i < len(ctx); i++ {
		arm.b[i] += reward * ctx[i]
	}

	arm.pulls++
	arm.rewards += reward
}

// computeUCB calculates the UCB score for an arm: θ^T·x + α·sqrt(x^T·A^{-1}·x)
func (cb *ContextualBandit) computeUCB(arm *linUCBArm, ctx []float64) float64 {
	d := cb.config.Dimension
	invA := invertMatrix(arm.A, d)

	// θ = A^{-1} · b
	theta := matVecMul(invA, arm.b, d)

	// Expected reward: θ^T · x
	expected := dotProduct(theta, ctx, d)

	// Uncertainty: sqrt(x^T · A^{-1} · x)
	Ainv_x := matVecMul(invA, ctx, d)
	uncertainty := math.Sqrt(math.Abs(dotProduct(ctx, Ainv_x, d)))

	return expected + cb.config.Alpha*uncertainty
}

// padContext ensures context has the required dimension.
func (cb *ContextualBandit) padContext(ctx []float64) []float64 {
	d := cb.config.Dimension
	if len(ctx) >= d {
		return ctx[:d]
	}
	padded := make([]float64, d)
	copy(padded, ctx)
	return padded
}

// BuildContext constructs the context vector from an observation and behavior profile.
func BuildContext(obs *Observation, profile *BehaviorSummary, matchResult *MatchResult) []float64 {
	ctx := make([]float64, 8)

	if obs.RiskAssessment != nil {
		ctx[0] = obs.RiskAssessment.Score
	}
	if obs.Classification != nil {
		ctx[1] = obs.Classification.Confidence
	}
	if profile != nil {
		ctx[2] = profile.ConsistencyScore
		ctx[3] = math.Min(profile.FPSwitchRate/10.0, 1.0) // normalize to [0,1]
		if profile.AvgRequestInterval > 0 {
			ctx[4] = math.Min(1.0/profile.AvgRequestInterval/20.0, 1.0) // req/s normalized
		}
		ctx[5] = (profile.RiskTrend + 1.0) / 2.0 // normalize [-1,1] → [0,1]
	}
	if matchResult != nil {
		ctx[6] = matchResult.SuspicionScore
	}
	if profile != nil {
		ctx[7] = math.Min(float64(profile.TotalObservations)/100.0, 1.0) // observation count normalized
	}

	return ctx
}

// Stats returns bandit statistics.
func (cb *ContextualBandit) Stats() BanditStats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	arms := make(map[string]ArmStats, len(cb.arms))
	for id, arm := range cb.arms {
		avgReward := 0.0
		if arm.pulls > 0 {
			avgReward = arm.rewards / float64(arm.pulls)
		}
		arms[id] = ArmStats{
			Pulls:     arm.pulls,
			AvgReward: avgReward,
		}
	}

	return BanditStats{
		TotalArms: len(cb.arms),
		Arms:      arms,
	}
}

// BanditStats contains contextual bandit runtime statistics.
type BanditStats struct {
	TotalArms int                 `json:"total_arms"`
	Arms      map[string]ArmStats `json:"arms"`
}

// ArmStats contains per-arm statistics.
type ArmStats struct {
	Pulls     int     `json:"pulls"`
	AvgReward float64 `json:"avg_reward"`
}

// --- Linear algebra helpers (small d×d, no external dependency needed) ---

// invertMatrix computes the inverse of a d×d matrix using Gauss-Jordan elimination.
// Falls back to identity if singular.
func invertMatrix(A [][]float64, d int) [][]float64 {
	// Build augmented matrix [A | I]
	aug := make([][]float64, d)
	for i := 0; i < d; i++ {
		aug[i] = make([]float64, 2*d)
		copy(aug[i][:d], A[i])
		aug[i][d+i] = 1.0
	}

	for col := 0; col < d; col++ {
		// Partial pivoting
		maxRow := col
		maxVal := math.Abs(aug[col][col])
		for row := col + 1; row < d; row++ {
			if v := math.Abs(aug[row][col]); v > maxVal {
				maxVal = v
				maxRow = row
			}
		}
		if maxVal < 1e-12 {
			// Near-singular: return identity (safe fallback for UCB)
			identity := make([][]float64, d)
			for i := 0; i < d; i++ {
				identity[i] = make([]float64, d)
				identity[i][i] = 1.0
			}
			return identity
		}
		aug[col], aug[maxRow] = aug[maxRow], aug[col]

		// Scale pivot row
		pivot := aug[col][col]
		for j := 0; j < 2*d; j++ {
			aug[col][j] /= pivot
		}

		// Eliminate column
		for row := 0; row < d; row++ {
			if row == col {
				continue
			}
			factor := aug[row][col]
			for j := 0; j < 2*d; j++ {
				aug[row][j] -= factor * aug[col][j]
			}
		}
	}

	// Extract inverse
	inv := make([][]float64, d)
	for i := 0; i < d; i++ {
		inv[i] = make([]float64, d)
		copy(inv[i], aug[i][d:])
	}
	return inv
}

// matVecMul computes A·x for d×d matrix A and d-vector x.
func matVecMul(A [][]float64, x []float64, d int) []float64 {
	result := make([]float64, d)
	for i := 0; i < d; i++ {
		for j := 0; j < d; j++ {
			result[i] += A[i][j] * x[j]
		}
	}
	return result
}

// dotProduct computes x^T · y.
func dotProduct(x, y []float64, d int) float64 {
	sum := 0.0
	for i := 0; i < d; i++ {
		sum += x[i] * y[i]
	}
	return sum
}

// RandomBandit is a simple baseline that selects arms uniformly at random.
// Useful for A/B testing against the LinUCB bandit.
type RandomBandit struct {
	arms []string
	mu   sync.RWMutex
}

// NewRandomBandit creates a uniform random arm selector.
func NewRandomBandit() *RandomBandit {
	return &RandomBandit{}
}

// RegisterArm adds an arm.
func (rb *RandomBandit) RegisterArm(id string) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.arms = append(rb.arms, id)
}

// SelectArm returns a uniformly random arm.
func (rb *RandomBandit) SelectArm() string {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	if len(rb.arms) == 0 {
		return ""
	}
	return rb.arms[rand.IntN(len(rb.arms))]
}
