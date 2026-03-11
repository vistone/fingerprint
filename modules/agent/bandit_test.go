package agent

import (
	"math"
	"testing"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/ml"
)

func TestNewContextualBandit(t *testing.T) {
	cb := NewContextualBandit(nil)
	if cb == nil {
		t.Fatal("NewContextualBandit returned nil")
	}
	stats := cb.Stats()
	if stats.TotalArms != 0 {
		t.Errorf("expected 0 arms, got %d", stats.TotalArms)
	}
}

func TestRegisterArm(t *testing.T) {
	cb := NewContextualBandit(nil)
	cb.RegisterArm("strategy-a")
	cb.RegisterArm("strategy-b")
	cb.RegisterArm("strategy-a") // duplicate — should not add twice

	stats := cb.Stats()
	if stats.TotalArms != 2 {
		t.Errorf("expected 2 arms, got %d", stats.TotalArms)
	}
}

func TestSelectArmSingle(t *testing.T) {
	cb := NewContextualBandit(nil)
	cb.RegisterArm("only-arm")

	ctx := []float64{0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5}
	arm, _ := cb.SelectArm(ctx)
	if arm != "only-arm" {
		t.Errorf("expected only-arm, got %s", arm)
	}
}

func TestSelectArmAmongSubset(t *testing.T) {
	cb := NewContextualBandit(nil)
	cb.RegisterArm("a")
	cb.RegisterArm("b")
	cb.RegisterArm("c")

	// Give arm "b" a big reward so it should be preferred
	ctx := []float64{0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5}
	for i := 0; i < 20; i++ {
		cb.UpdateReward("b", ctx, 10.0)
		cb.UpdateReward("a", ctx, 0.1)
		cb.UpdateReward("c", ctx, 0.1)
	}

	arm, _ := cb.SelectArmAmong(ctx, []string{"a", "b"})
	if arm != "b" {
		t.Errorf("expected b (high reward), got %s", arm)
	}
}

func TestUpdateRewardCreatesArm(t *testing.T) {
	cb := NewContextualBandit(nil)
	ctx := make([]float64, 8)

	// Updating a non-existent arm should create it
	cb.UpdateReward("new-arm", ctx, 1.0)

	stats := cb.Stats()
	if stats.TotalArms != 1 {
		t.Errorf("expected 1 arm (auto-created), got %d", stats.TotalArms)
	}
	as, ok := stats.Arms["new-arm"]
	if !ok {
		t.Fatal("new-arm not in stats")
	}
	if as.Pulls != 1 {
		t.Errorf("expected 1 pull, got %d", as.Pulls)
	}
}

func TestBuildContext(t *testing.T) {
	obs := &Observation{
		RiskAssessment: &core.RiskAssessment{Score: 0.8},
		Classification: &ml.ClassificationResult{Confidence: 0.7},
	}
	profile := &BehaviorSummary{
		ConsistencyScore:   0.9,
		FPSwitchRate:       2.5,
		AvgRequestInterval: 0.1, // 10 req/s
		RiskTrend:          0.5,
		TotalObservations:  50,
	}
	match := &MatchResult{
		SuspicionScore: 0.3,
	}

	ctx := BuildContext(obs, profile, match)

	if len(ctx) != 8 {
		t.Fatalf("expected 8-dim context, got %d", len(ctx))
	}
	if ctx[0] != 0.8 {
		t.Errorf("ctx[0] risk = %f, want 0.8", ctx[0])
	}
	if ctx[1] != 0.7 {
		t.Errorf("ctx[1] conf = %f, want 0.7", ctx[1])
	}
	if ctx[6] != 0.3 {
		t.Errorf("ctx[6] suspicion = %f, want 0.3", ctx[6])
	}
}

func TestBanditLearnsPreference(t *testing.T) {
	cb := NewContextualBandit(&BanditConfig{Alpha: 0.5, Dimension: 4})
	cb.RegisterArm("good")
	cb.RegisterArm("bad")

	ctx := []float64{1.0, 0.5, 0.3, 0.8}

	// Train: arm "good" gets positive reward, "bad" gets negative
	for i := 0; i < 50; i++ {
		cb.UpdateReward("good", ctx, 1.0)
		cb.UpdateReward("bad", ctx, -0.5)
	}

	arm, _ := cb.SelectArm(ctx)
	if arm != "good" {
		t.Errorf("expected learned preference for 'good', got '%s'", arm)
	}
}

func TestInvertMatrixIdentity(t *testing.T) {
	d := 3
	I := make([][]float64, d)
	for i := 0; i < d; i++ {
		I[i] = make([]float64, d)
		I[i][i] = 1.0
	}

	inv := invertMatrix(I, d)

	for i := 0; i < d; i++ {
		for j := 0; j < d; j++ {
			expected := 0.0
			if i == j {
				expected = 1.0
			}
			if math.Abs(inv[i][j]-expected) > 1e-9 {
				t.Errorf("inv[%d][%d] = %f, want %f", i, j, inv[i][j], expected)
			}
		}
	}
}

func TestPadContext(t *testing.T) {
	cb := NewContextualBandit(&BanditConfig{Alpha: 1.0, Dimension: 4})

	// Short context should be padded
	short := []float64{1.0, 2.0}
	padded := cb.padContext(short)
	if len(padded) != 4 {
		t.Errorf("expected len 4, got %d", len(padded))
	}
	if padded[0] != 1.0 || padded[1] != 2.0 || padded[2] != 0 || padded[3] != 0 {
		t.Errorf("unexpected padded: %v", padded)
	}

	// Long context should be truncated
	long := []float64{1, 2, 3, 4, 5, 6}
	trimmed := cb.padContext(long)
	if len(trimmed) != 4 {
		t.Errorf("expected len 4, got %d", len(trimmed))
	}
}

func TestRandomBandit(t *testing.T) {
	rb := NewRandomBandit()

	// Empty should return ""
	if rb.SelectArm() != "" {
		t.Error("expected empty string from empty random bandit")
	}

	rb.RegisterArm("x")
	rb.RegisterArm("y")

	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		seen[rb.SelectArm()] = true
	}
	if len(seen) < 2 {
		t.Error("expected both arms to be selected at least once in 100 trials")
	}
}

func TestBanditSelectArmEmpty(t *testing.T) {
	cb := NewContextualBandit(nil)
	arm, ucb := cb.SelectArm([]float64{0.5})
	if arm != "" || ucb != 0 {
		t.Errorf("expected empty result for no arms, got arm=%q ucb=%f", arm, ucb)
	}
}

func TestSelectArmAmongEmpty(t *testing.T) {
	cb := NewContextualBandit(nil)
	cb.RegisterArm("a")
	arm, ucb := cb.SelectArmAmong([]float64{0.5}, nil)
	if arm != "" || ucb != 0 {
		t.Errorf("expected empty result for no candidates, got arm=%q ucb=%f", arm, ucb)
	}
}
