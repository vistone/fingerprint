package pipeline

import (
	"context"
	"testing"
	"time"
)

func TestTimeoutMiddleware_DoesNotMutateOriginalStageDataAfterTimeout(t *testing.T) {
	middleware := NewTimeoutMiddleware(10 * time.Millisecond)
	data := &StageData{
		Input:   "input",
		Output:  "original",
		Context: map[string]interface{}{"key": "value"},
	}
	mutated := make(chan struct{}, 1)

	err := middleware.Process(context.Background(), "slow-stage", data, func(ctx context.Context, stageData *StageData) error {
		time.Sleep(30 * time.Millisecond)
		stageData.Output = "mutated-after-timeout"
		stageData.Context["late"] = "write"
		mutated <- struct{}{}
		return nil
	})
	if err == nil {
		t.Fatal("expected timeout error")
	}

	<-mutated

	if data.Output != "original" {
		t.Fatalf("expected original output to remain unchanged after timeout, got %v", data.Output)
	}
	if _, exists := data.Context["late"]; exists {
		t.Fatal("expected original context to remain unchanged after timeout")
	}
}
