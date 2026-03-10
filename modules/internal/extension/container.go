package extension

import (
	"fmt"
	"sync"
	"time"
)

// Container is the dependency injection container that manages the lifecycle and dependencies of all components
//
// Usage example:
//
//	config := NewConfig(EnvProduction)
//	container := NewContainer(config)
//
//	logger, err := container.GetLogger("myapp")
//	guard, err := container.GetRequestGuard()
//	validator, err := container.GetValidator()
//
// Features:
//   - Singleton pattern: each component is created only once
//   - Lazy loading: components are created on first use
//   - Thread-safe: uses mutexes to protect shared state
type Container struct {
	mu sync.RWMutex

	// Configuration object
	config *Config

	// Singleton cache
	singletons map[string]interface{}

	// Factory function registry
	factories map[string]func() (interface{}, error)

	// Initialization flag
	initialized bool
}

// NewContainer creates a dependency injection container
func NewContainer(config *Config) *Container {
	if config == nil {
		config = NewUnifiedConfigFromEnv()
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		// Continue with valid default configuration
		config = NewConfig(EnvDevelopment)
	}

	return &Container{
		config:     config,
		singletons: make(map[string]interface{}),
		factories:  make(map[string]func() (interface{}, error)),
	}
}

// Register registers a factory function
func (c *Container) Register(name string, factory func() (interface{}, error)) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.factories[name]; exists {
		return NewError(ErrCodeAlreadyRegistered,
			fmt.Sprintf("factory already registered: %s", name))
	}

	c.factories[name] = factory
	return nil
}

// Get retrieves a component instance (singleton)
func (c *Container) Get(name string) (interface{}, error) {
	c.mu.RLock()
	singleton, exists := c.singletons[name]
	factory, factoryExists := c.factories[name]
	c.mu.RUnlock()

	if exists {
		return singleton, nil
	}

	if !factoryExists {
		return nil, NewError(ErrCodeNotFound,
			fmt.Sprintf("component not registered: %s", name))
	}

	// Create new instance
	instance, err := factory()
	if err != nil {
		return nil, err
	}

	// Store in singleton cache
	c.mu.Lock()
	c.singletons[name] = instance
	c.mu.Unlock()

	return instance, nil
}

// GetConfig returns the configuration object
func (c *Container) GetConfig() *Config {
	return c.config
}

// GetLogger returns the logger instance
func (c *Container) GetLogger(name string) (*SimpleLogger, error) {
	key := "logger_" + name

	instance, err := c.Get(key)
	if err == nil {
		return instance.(*SimpleLogger), nil
	}

	// Register and create a new logger
	c.Register(key, func() (interface{}, error) {
		logger := NewSimpleLogger(name)
		logger.SetLevel(c.config.Logger.Level)
		return logger, nil
	})

	instance, err = c.Get(key)
	if err != nil {
		return nil, err
	}

	return instance.(*SimpleLogger), nil
}

// GetValidator returns the validator instance
func (c *Container) GetValidator() (*DefaultValidator, error) {
	instance, err := c.Get("validator")
	if err == nil {
		return instance.(*DefaultValidator), nil
	}

	c.Register("validator", func() (interface{}, error) {
		validator := NewDefaultValidator()
		validator.MaxDataSize = c.config.Validator.MaxDataSize
		validator.StrictMode = c.config.Validator.StrictMode
		return validator, nil
	})

	instance, err = c.Get("validator")
	if err != nil {
		return nil, err
	}

	return instance.(*DefaultValidator), nil
}

// GetSecurityAuditor returns the security auditor instance
func (c *Container) GetSecurityAuditor() (*SecurityAuditor, error) {
	instance, err := c.Get("auditor")
	if err == nil {
		return instance.(*SecurityAuditor), nil
	}

	c.Register("auditor", func() (interface{}, error) {
		return NewSecurityAuditor(c.config.Audit.MaxEvents), nil
	})

	instance, err = c.Get("auditor")
	if err != nil {
		return nil, err
	}

	return instance.(*SecurityAuditor), nil
}

// GetRequestGuard returns the request guard instance
func (c *Container) GetRequestGuard() (*RequestGuard, error) {
	instance, err := c.Get("request_guard")
	if err == nil {
		return instance.(*RequestGuard), nil
	}

	c.Register("request_guard", func() (interface{}, error) {
		return NewRequestGuard(c.config.Defense), nil
	})

	instance, err = c.Get("request_guard")
	if err != nil {
		return nil, err
	}

	return instance.(*RequestGuard), nil
}

// GetRateLimiter returns the rate limiter instance
func (c *Container) GetRateLimiter() (*RateLimiter, error) {
	instance, err := c.Get("rate_limiter")
	if err == nil {
		return instance.(*RateLimiter), nil
	}

	c.Register("rate_limiter", func() (interface{}, error) {
		limiter := NewRateLimiter(
			c.config.Defense.RateLimit,
			time.Minute,
		)
		return limiter, nil
	})

	instance, err = c.Get("rate_limiter")
	if err != nil {
		return nil, err
	}

	return instance.(*RateLimiter), nil
}

// Reset resets the container, clearing all singletons
// Note: only for use in tests
func (c *Container) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.singletons = make(map[string]interface{})
	c.initialized = false
}

// Initialize initializes all default components in the container
// Pre-loads critical components to improve startup performance
func (c *Container) Initialize() error {
	c.mu.Lock()
	if c.initialized {
		c.mu.Unlock()
		return nil
	}
	c.mu.Unlock()

	// Pre-initialize critical components
	// Use typed accessors to ensure correct factory registration and initialization
	if _, err := c.GetValidator(); err != nil {
		return fmt.Errorf("init validator: %w", err)
	}

	if _, err := c.GetSecurityAuditor(); err != nil {
		return fmt.Errorf("init security auditor: %w", err)
	}

	if _, err := c.GetRequestGuard(); err != nil {
		return fmt.Errorf("init request guard: %w", err)
	}

	if _, err := c.GetRateLimiter(); err != nil {
		return fmt.Errorf("init rate limiter: %w", err)
	}

	c.mu.Lock()
	c.initialized = true
	c.mu.Unlock()

	return nil
}
