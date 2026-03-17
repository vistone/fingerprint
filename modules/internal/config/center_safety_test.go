package config

import (
	"testing"
	"time"
)

type rollbackReadListener struct {
	center *ConfigCenter
	done   chan struct{}
}

func (l *rollbackReadListener) OnConfigChange(_, _ *ManagedConfig, _ []ConfigChange) error {
	_ = l.center.Get()
	close(l.done)
	return nil
}

func TestConfigCenterGet_ReturnsClone(t *testing.T) {
	cc := NewConfigCenter("")
	cc.current = DefaultManagedConfig()

	got := cc.Get()
	got.BehaviorAnalysis.MinRequestsForAnalysis = 999

	again := cc.Get()
	if again.BehaviorAnalysis.MinRequestsForAnalysis == 999 {
		t.Fatal("Get should return cloned config, not mutable internal state")
	}
}

func TestConfigCenterRollback_ListenerRunsOutsideLock(t *testing.T) {
	cc := NewConfigCenter("")
	cc.loaded = true

	base := DefaultManagedConfig()
	base.BehaviorAnalysis.MinRequestsForAnalysis = 5
	if err := cc.Update(base, "v1", "tester"); err != nil {
		t.Fatalf("seed update failed: %v", err)
	}
	next := DefaultManagedConfig()
	next.BehaviorAnalysis.MinRequestsForAnalysis = 8
	if err := cc.Update(next, "v2", "tester"); err != nil {
		t.Fatalf("seed update failed: %v", err)
	}

	done := make(chan struct{}, 1)
	cc.RegisterListener(&rollbackReadListener{center: cc, done: done})

	rollbackDone := make(chan error, 1)
	go func() {
		rollbackDone <- cc.Rollback("v1", "rollback", "tester")
	}()

	select {
	case err := <-rollbackDone:
		if err != nil {
			t.Fatalf("rollback failed: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("rollback timed out, listener likely invoked under lock")
	}

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("listener was not invoked")
	}
}
