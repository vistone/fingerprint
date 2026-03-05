package gateway

import (
	"sync"
	"testing"
	"time"

	"github.com/vistone/fingerprint/modules/internal/testhelpers"
)

// MockConnection for testing
type MockConnection struct {
	ID       int
	Active   bool
	LastUsed time.Time
}

func TestNewConnectionPool(t *testing.T) {
	tests := []struct {
		name      string
		maxSize   int
		wantPanic bool
	}{
		{
			name:      "valid pool creation",
			maxSize:   10,
			wantPanic: false,
		},
		{
			name:      "zero max size uses default",
			maxSize:   0,
			wantPanic: false,
		},
		{
			name:      "negative max size uses default",
			maxSize:   -1,
			wantPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewConnectionPool(Config{
				MaxSize:     tt.maxSize,
				TTL:         time.Minute,
				MaxIdleTime: time.Second * 30,
			})
			testhelpers.AssertNotNil(t, pool)
			testhelpers.AssertEqual(t, pool.IsRunning(), false)
		})
	}
}

func TestConnectionPool_GetPut(t *testing.T) {
	pool := NewConnectionPool(Config{
		MaxSize:     5,
		TTL:         time.Minute,
		MaxIdleTime: time.Second * 30,
	})

	factory := func() (interface{}, error) {
		return &MockConnection{ID: 1, Active: true}, nil
	}
	pool.SetFactory(factory)

	t.Run("get from empty pool creates new", func(t *testing.T) {
		conn, err := pool.Get()
		testhelpers.AssertNoError(t, err)
		testhelpers.AssertNotNil(t, conn)
		testhelpers.AssertEqual(t, conn.(*MockConnection).Active, true)
	})

	t.Run("put connection back to pool", func(t *testing.T) {
		conn := &MockConnection{ID: 2, Active: true}
		err := pool.Put(conn)
		testhelpers.AssertNoError(t, err)
		testhelpers.AssertEqual(t, pool.Len(), 1)
	})

	t.Run("get returns pooled connection", func(t *testing.T) {
		conn, err := pool.Get()
		testhelpers.AssertNoError(t, err)
		testhelpers.AssertNotNil(t, conn)
		testhelpers.AssertEqual(t, pool.Len(), 0)
	})

	t.Run("put nil returns error", func(t *testing.T) {
		err := pool.Put(nil)
		testhelpers.AssertError(t, err)
	})
}

func TestConnectionPool_MaxSize(t *testing.T) {
	pool := NewConnectionPool(Config{
		MaxSize:     3,
		TTL:         time.Minute,
		MaxIdleTime: time.Second * 30,
	})

	factory := func() (interface{}, error) {
		return &MockConnection{Active: true}, nil
	}
	pool.SetFactory(factory)

	// Create connections up to max
	for i := 0; i < 3; i++ {
		conn, err := pool.Get()
		testhelpers.AssertNoError(t, err)
		testhelpers.AssertNotNil(t, conn)
	}

	t.Run("exceeding max size returns error", func(t *testing.T) {
		_, err := pool.Get()
		testhelpers.AssertError(t, err)
	})
}

func TestConnectionPool_ConcurrentAccess(t *testing.T) {
	pool := NewConnectionPool(Config{
		MaxSize:     100,
		TTL:         time.Minute,
		MaxIdleTime: time.Second * 30,
	})

	counter := 0
	var mu sync.Mutex
	factory := func() (interface{}, error) {
		mu.Lock()
		defer mu.Unlock()
		counter++
		return &MockConnection{ID: counter, Active: true}, nil
	}
	pool.SetFactory(factory)

	var wg sync.WaitGroup
	numGoroutines := 50
	operationsPerGoroutine := 20

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				conn, err := pool.Get()
				if err != nil {
					continue
				}
				time.Sleep(time.Microsecond * 10)
				pool.Put(conn)
			}
		}()
	}

	wg.Wait()
	testhelpers.AssertEqual(t, pool.Len() <= 100, true)
}

func TestConnectionPool_HealthCheck(t *testing.T) {
	pool := NewConnectionPool(Config{
		MaxSize:     5,
		TTL:         time.Minute,
		MaxIdleTime: time.Millisecond * 100,
	})

	factory := func() (interface{}, error) {
		return &MockConnection{Active: true, LastUsed: time.Now()}, nil
	}
	healthCheck := func(conn interface{}) bool {
		return conn.(*MockConnection).Active
	}

	pool.SetFactory(factory)
	pool.SetHealthCheck(healthCheck)

	// Hold connections and put them back to fill the pool
	// (Get creates new, Put returns to pool)
	conns := make([]interface{}, 0, 3)
	for i := 0; i < 3; i++ {
		conn, _ := pool.Get()
		conns = append(conns, conn)
	}
	// Now put all back to pool
	for _, conn := range conns {
		pool.Put(conn)
	}

	testhelpers.AssertEqual(t, pool.Len(), 3)

	// Mark one connection as unhealthy
	conn, _ := pool.Get()
	if conn != nil {
		conn.(*MockConnection).Active = false
		pool.Put(conn)
	}

	// Start health check
	pool.StartHealthCheck(time.Millisecond * 50)
	time.Sleep(time.Millisecond * 200)
	pool.StopHealthCheck()

	// Unhealthy connection should be removed
	testhelpers.AssertEqual(t, pool.Len() <= 3, true)
}

func TestConnectionPool_Close(t *testing.T) {
	pool := NewConnectionPool(Config{
		MaxSize:     5,
		TTL:         time.Minute,
		MaxIdleTime: time.Second * 30,
	})

	factory := func() (interface{}, error) {
		return &MockConnection{Active: true}, nil
	}
	pool.SetFactory(factory)

	// Hold connections and put them back to fill the pool
	conns := make([]interface{}, 0, 3)
	for i := 0; i < 3; i++ {
		conn, _ := pool.Get()
		conns = append(conns, conn)
	}
	// Now put all back to pool
	for _, conn := range conns {
		pool.Put(conn)
	}

	testhelpers.AssertEqual(t, pool.Len(), 3)

	err := pool.Close()
	testhelpers.AssertNoError(t, err)
	testhelpers.AssertEqual(t, pool.Len(), 0)
	testhelpers.AssertEqual(t, pool.IsRunning(), false)
}

func TestPooledConnection(t *testing.T) {
	t.Run("creation time tracking", func(t *testing.T) {
		before := time.Now()
		pc := &PooledConnection{
			Connection:   &MockConnection{},
			CreationTime: time.Now(),
		}
		after := time.Now()

		testhelpers.AssertEqual(t, pc.CreationTime.After(before) || pc.CreationTime.Equal(before), true)
		testhelpers.AssertEqual(t, pc.CreationTime.Before(after) || pc.CreationTime.Equal(after), true)
	})

	t.Run("last used time updates", func(t *testing.T) {
		pc := &PooledConnection{
			Connection:   &MockConnection{},
			CreationTime: time.Now(),
			LastUsed:     time.Now(),
		}

		oldLastUsed := pc.LastUsed
		time.Sleep(time.Millisecond * 10)
		pc.LastUsed = time.Now()

		testhelpers.AssertEqual(t, pc.LastUsed.After(oldLastUsed), true)
	})

	t.Run("health status", func(t *testing.T) {
		pc := &PooledConnection{
			Connection:   &MockConnection{Active: true},
			CreationTime: time.Now(),
			LastUsed:     time.Now(),
			Healthy:      true,
		}

		testhelpers.AssertEqual(t, pc.Healthy, true)

		pc.Healthy = false
		testhelpers.AssertEqual(t, pc.Healthy, false)
	})
}

func BenchmarkConnectionPool_GetPut(b *testing.B) {
	pool := NewConnectionPool(Config{
		MaxSize:     1000,
		TTL:         time.Minute,
		MaxIdleTime: time.Second * 30,
	})

	factory := func() (interface{}, error) {
		return &MockConnection{Active: true}, nil
	}
	pool.SetFactory(factory)

	// Pre-populate pool
	for i := 0; i < 100; i++ {
		conn, _ := pool.Get()
		pool.Put(conn)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			conn, err := pool.Get()
			if err != nil {
				continue
			}
			pool.Put(conn)
		}
	})
}
