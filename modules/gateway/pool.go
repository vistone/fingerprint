package gateway

import (
	"errors"
	"sync"
	"time"
)

var (
	// ErrPoolExhausted is returned when the pool is at maximum capacity
	ErrPoolExhausted = errors.New("connection pool exhausted")
	// ErrNilConnection is returned when trying to put a nil connection
	ErrNilConnection = errors.New("cannot put nil connection")
)

// Config holds configuration for the connection pool
type Config struct {
	MaxSize     int
	TTL         time.Duration
	MaxIdleTime time.Duration
}

// PooledConnection wraps a connection with metadata
type PooledConnection struct {
	Connection   interface{}
	CreationTime time.Time
	LastUsed     time.Time
	Healthy      bool
}

// ConnectionPool manages a pool of reusable connections
type ConnectionPool struct {
	mu            sync.RWMutex
	connections   []*PooledConnection
	maxSize       int
	ttl           time.Duration
	maxIdleTime   time.Duration
	factory       func() (interface{}, error)
	healthCheck   func(interface{}) bool
	running       bool
	stopHealthCheck chan struct{}
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(config Config) *ConnectionPool {
	if config.MaxSize <= 0 {
		config.MaxSize = 100
	}
	if config.TTL <= 0 {
		config.TTL = time.Minute * 10
	}
	if config.MaxIdleTime <= 0 {
		config.MaxIdleTime = time.Minute * 5
	}

	return &ConnectionPool{
		connections:   make([]*PooledConnection, 0, config.MaxSize),
		maxSize:       config.MaxSize,
		ttl:           config.TTL,
		maxIdleTime:   config.MaxIdleTime,
		stopHealthCheck: make(chan struct{}),
	}
}

// SetFactory sets the connection factory function
func (p *ConnectionPool) SetFactory(factory func() (interface{}, error)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.factory = factory
}

// SetHealthCheck sets the health check function
func (p *ConnectionPool) SetHealthCheck(healthCheck func(interface{}) bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.healthCheck = healthCheck
}

// Get retrieves a connection from the pool
func (p *ConnectionPool) Get() (interface{}, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Try to get an existing connection
	for len(p.connections) > 0 {
		// Get from the end for O(1) removal
		idx := len(p.connections) - 1
		pc := p.connections[idx]
		p.connections = p.connections[:idx]

		// Check if connection is still valid
		if time.Since(pc.CreationTime) > p.ttl {
			continue // Expired, skip
		}

		pc.LastUsed = time.Now()
		return pc.Connection, nil
	}

	// No available connection, create new one if factory is set
	if p.factory != nil {
		if len(p.connections) >= p.maxSize {
			return nil, ErrPoolExhausted
		}
		conn, err := p.factory()
		if err != nil {
			return nil, err
		}
		return conn, nil
	}

	return nil, ErrPoolExhausted
}

// Put returns a connection to the pool
func (p *ConnectionPool) Put(conn interface{}) error {
	if conn == nil {
		return ErrNilConnection
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Don't exceed max size
	if len(p.connections) >= p.maxSize {
		return nil // Silently drop excess connections
	}

	pc := &PooledConnection{
		Connection:   conn,
		CreationTime: time.Now(),
		LastUsed:     time.Now(),
		Healthy:      true,
	}

	p.connections = append(p.connections, pc)
	return nil
}

// Len returns the current number of pooled connections
func (p *ConnectionPool) Len() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.connections)
}

// IsRunning returns true if health check is running
func (p *ConnectionPool) IsRunning() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.running
}

// StartHealthCheck starts the background health check goroutine
func (p *ConnectionPool) StartHealthCheck(interval time.Duration) {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return
	}
	p.running = true
	p.stopHealthCheck = make(chan struct{})
	p.mu.Unlock()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				p.performHealthCheck()
			case <-p.stopHealthCheck:
				return
			}
		}
	}()
}

// StopHealthCheck stops the background health check
func (p *ConnectionPool) StopHealthCheck() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		p.running = false
		close(p.stopHealthCheck)
	}
}

// performHealthCheck removes unhealthy or idle connections
func (p *ConnectionPool) performHealthCheck() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.healthCheck == nil {
		return
	}

	valid := make([]*PooledConnection, 0, len(p.connections))
	now := time.Now()

	for _, pc := range p.connections {
		// Remove expired connections
		if now.Sub(pc.CreationTime) > p.ttl {
			continue
		}

		// Remove idle connections
		if now.Sub(pc.LastUsed) > p.maxIdleTime {
			continue
		}

		// Check health
		if !p.healthCheck(pc.Connection) {
			continue
		}

		valid = append(valid, pc)
	}

	p.connections = valid
}

// Close closes the pool and clears all connections
func (p *ConnectionPool) Close() error {
	p.StopHealthCheck()

	p.mu.Lock()
	defer p.mu.Unlock()

	p.connections = p.connections[:0]
	return nil
}
