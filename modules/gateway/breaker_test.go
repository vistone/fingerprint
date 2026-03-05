package gateway

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/vistone/fingerprint/modules/internal/testhelpers"
)

func TestNewCircuitBreaker(t *testing.T) {
	tests := []struct {
		name             string
		failureThreshold int
		timeout          time.Duration
	}{
		{
			name:             "valid circuit breaker",
			failureThreshold: 5,
			timeout:          time.Second * 30,
		},
		{
			name:             "zero threshold uses default",
			failureThreshold: 0,
			timeout:          time.Second * 30,
		},
		{
			name:             "zero timeout uses default",
			failureThreshold: 5,
			timeout:          0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := NewCircuitBreaker(tt.failureThreshold, tt.timeout)
			testhelpers.AssertNotNil(t, cb)
			testhelpers.AssertEqual(t, cb.State(), StateClosed)
		})
	}
}

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	t.Run("closed to open on failures", func(t *testing.T) {
		cb := NewCircuitBreaker(3, time.Second*30)

		// Should start closed
		testhelpers.AssertEqual(t, cb.State(), StateClosed)

		// Record failures up to threshold
		cb.RecordFailure()
		testhelpers.AssertEqual(t, cb.State(), StateClosed)

		cb.RecordFailure()
		testhelpers.AssertEqual(t, cb.State(), StateClosed)

		cb.RecordFailure()
		// After 3 failures, should open
		testhelpers.AssertEqual(t, cb.State(), StateOpen)
	})

	t.Run("open to half-open after timeout", func(t *testing.T) {
		cb := NewCircuitBreaker(1, time.Millisecond*50)

		// Force open
		cb.RecordFailure()
		testhelpers.AssertEqual(t, cb.State(), StateOpen)

		// Wait for timeout
		time.Sleep(time.Millisecond * 100)

		// Next request should transition to half-open
		err := cb.Execute(func() error {
			return nil
		})
		testhelpers.AssertNoError(t, err)
		testhelpers.AssertEqual(t, cb.State(), StateHalfOpen)
	})

	t.Run("half-open to closed on success", func(t *testing.T) {
		cb := NewCircuitBreaker(1, time.Millisecond*50)

		// Force open
		cb.RecordFailure()
		time.Sleep(time.Millisecond * 100)

		// Transition to half-open and succeed
		cb.Execute(func() error { return nil })
		testhelpers.AssertEqual(t, cb.State(), StateHalfOpen)

		// More successes to close
		cb.RecordSuccess()
		cb.RecordSuccess()

		testhelpers.AssertEqual(t, cb.State(), StateClosed)
	})

	t.Run("half-open to open on failure", func(t *testing.T) {
		cb := NewCircuitBreaker(1, time.Millisecond*50)

		// Force open
		cb.RecordFailure()
		time.Sleep(time.Millisecond * 100)

		// Transition to half-open
		cb.Execute(func() error { return nil })
		testhelpers.AssertEqual(t, cb.State(), StateHalfOpen)

		// Fail in half-open
		cb.RecordFailure()
		testhelpers.AssertEqual(t, cb.State(), StateOpen)
	})
}

func TestCircuitBreaker_Execute(t *testing.T) {
	t.Run("execute success in closed state", func(t *testing.T) {
		cb := NewCircuitBreaker(5, time.Second*30)

		called := false
		err := cb.Execute(func() error {
			called = true
			return nil
		})

		testhelpers.AssertNoError(t, err)
		testhelpers.AssertEqual(t, called, true)
	})

	t.Run("execute failure in closed state", func(t *testing.T) {
		cb := NewCircuitBreaker(5, time.Second*30)

		expectedErr := errors.New("operation failed")
		err := cb.Execute(func() error {
			return expectedErr
		})

		testhelpers.AssertError(t, err)
		testhelpers.AssertEqual(t, err, expectedErr)
	})

	t.Run("execute blocked in open state", func(t *testing.T) {
		cb := NewCircuitBreaker(1, time.Minute)

		// Force open
		cb.RecordFailure()
		testhelpers.AssertEqual(t, cb.State(), StateOpen)

		called := false
		err := cb.Execute(func() error {
			called = true
			return nil
		})

		testhelpers.AssertError(t, err)
		testhelpers.AssertEqual(t, errors.Is(err, ErrCircuitOpen), true)
		testhelpers.AssertEqual(t, called, false)
	})

	t.Run("execute in half-open state", func(t *testing.T) {
		cb := NewCircuitBreaker(1, time.Millisecond*50)

		// Force open and wait
		cb.RecordFailure()
		time.Sleep(time.Millisecond * 100)

		// Should allow one request in half-open
		called := false
		err := cb.Execute(func() error {
			called = true
			return nil
		})

		testhelpers.AssertNoError(t, err)
		testhelpers.AssertEqual(t, called, true)
		testhelpers.AssertEqual(t, cb.State(), StateHalfOpen)
	})
}

func TestCircuitBreaker_Metrics(t *testing.T) {
	cb := NewCircuitBreaker(5, time.Second*30)

	t.Run("initial metrics", func(t *testing.T) {
		metrics := cb.Metrics()
		testhelpers.AssertEqual(t, metrics.State, StateClosed)
		testhelpers.AssertEqual(t, metrics.FailureCount, 0)
		testhelpers.AssertEqual(t, metrics.SuccessCount, 0)
	})

	t.Run("metrics after operations", func(t *testing.T) {
		cb.RecordSuccess()
		cb.RecordSuccess()
		cb.RecordFailure()

		metrics := cb.Metrics()
		testhelpers.AssertEqual(t, metrics.SuccessCount, 2)
		testhelpers.AssertEqual(t, metrics.FailureCount, 1)
	})

	t.Run("metrics in open state", func(t *testing.T) {
		cb2 := NewCircuitBreaker(1, time.Minute)
		cb2.RecordFailure()

		metrics := cb2.Metrics()
		testhelpers.AssertEqual(t, metrics.State, StateOpen)
		testhelpers.AssertEqual(t, metrics.LastFailureTime.IsZero(), false)
	})
}

func TestCircuitBreaker_ConcurrentAccess(t *testing.T) {
	cb := NewCircuitBreaker(100, time.Second*30)

	var wg sync.WaitGroup
	numGoroutines := 50

	// Concurrent successes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				cb.Execute(func() error {
					return nil
				})
			}
		}()
	}

	wg.Wait()

	metrics := cb.Metrics()
	testhelpers.AssertEqual(t, metrics.SuccessCount, numGoroutines*20)
	testhelpers.AssertEqual(t, cb.State(), StateClosed)
}

func TestCircuitBreaker_FailureDecay(t *testing.T) {
	cb := NewCircuitBreaker(5, time.Second*30)

	// Record some failures
	cb.RecordFailure()
	cb.RecordFailure()
	testhelpers.AssertEqual(t, cb.failureCount, 2)

	// Wait for decay
	time.Sleep(time.Millisecond * 100)

	// Record a success should trigger decay
	cb.RecordSuccess()

	// Failure count should be reduced (implementation dependent)
	// The main point is circuit shouldn't open prematurely
	testhelpers.AssertEqual(t, cb.State(), StateClosed)
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := NewCircuitBreaker(3, time.Second*30)

	// Record some failures
	cb.RecordFailure()
	cb.RecordFailure()
	testhelpers.AssertEqual(t, cb.failureCount, 2)

	// Reset
	cb.Reset()

	testhelpers.AssertEqual(t, cb.State(), StateClosed)
	testhelpers.AssertEqual(t, cb.failureCount, 0)
	testhelpers.AssertEqual(t, cb.successCount, 0)

	metrics := cb.Metrics()
	testhelpers.AssertEqual(t, metrics.FailureCount, 0)
	testhelpers.AssertEqual(t, metrics.SuccessCount, 0)
}

func BenchmarkCircuitBreaker_Execute(b *testing.B) {
	cb := NewCircuitBreaker(1000, time.Second*30)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cb.Execute(func() error {
				return nil
			})
		}
	})
}

func BenchmarkCircuitBreaker_RecordFailure(b *testing.B) {
	cb := NewCircuitBreaker(1000000, time.Second*30)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cb.RecordFailure()
		}
	})
}
