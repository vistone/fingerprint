package extension

import (
	"fmt"
	"sync"
	"time"
)

// translated comment
//
// translated comment
//
//	config := NewConfig(EnvProduction)
//	container := NewContainer(config)
//
//	logger, err := container.GetLogger("myapp")
//	guard, err := container.GetRequestGuard()
//	validator, err := container.GetValidator()
//
// translated comment
// translated comment
// translated comment
// translated comment
type Container struct {
	mu sync.RWMutex

	// translated comment
	config *Config

	// translated comment
	singletons map[string]interface{}

	// translated comment
	factories map[string]func() (interface{}, error)

	// translated comment
	initialized bool
}

// translated comment
func NewContainer(config *Config) *Container {
	if config == nil {
		config = NewUnifiedConfigFromEnv()
	}

	// translated comment
	if err := config.Validate(); err != nil {
		// translated comment
		config = NewConfig(EnvDevelopment)
	}

	return &Container{
		config:     config,
		singletons: make(map[string]interface{}),
		factories:  make(map[string]func() (interface{}, error)),
	}
}

// translated comment
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

// translated comment
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

	// translated comment
	instance, err := factory()
	if err != nil {
		return nil, err
	}

	// translated comment
	c.mu.Lock()
	c.singletons[name] = instance
	c.mu.Unlock()

	return instance, nil
}

// translated comment
func (c *Container) GetConfig() *Config {
	return c.config
}

// translated comment
func (c *Container) GetLogger(name string) (*SimpleLogger, error) {
	key := "logger_" + name

	instance, err := c.Get(key)
	if err == nil {
		return instance.(*SimpleLogger), nil
	}

	// translated comment
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

// translated comment
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

// translated comment
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

// translated comment
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

// translated comment
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

// translated comment
// translated comment
func (c *Container) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.singletons = make(map[string]interface{})
	c.initialized = false
}

// translated comment
// translated comment
func (c *Container) Initialize() error {
	c.mu.Lock()
	if c.initialized {
		c.mu.Unlock()
		return nil
	}
	c.mu.Unlock()

	// translated comment
	// translated comment
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
