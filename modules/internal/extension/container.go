package extension

import (
	"fmt"
	"sync"
	"time"
)

// Container 依赖注入容器，管理所有组件的生命周期和依赖关系
//
// 使用示例：
//
//	config := NewConfig(EnvProduction)
//	container := NewContainer(config)
//
//	logger, err := container.GetLogger("myapp")
//	guard, err := container.GetRequestGuard()
//	validator, err := container.GetValidator()
//
// 特点：
//   - 单例模式：每个组件只创建一次
//   - 懒加载：组件在首次使用时才创建
//   - 线程安全：使用互斥锁保护共享状态
type Container struct {
	mu sync.RWMutex

	// 配置对象
	config *Config

	// 单例缓存
	singletons map[string]interface{}

	// 工厂函数注册
	factories map[string]func() (interface{}, error)

	// 初始化标记
	initialized bool
}

// NewContainer 创建依赖注入容器
func NewContainer(config *Config) *Container {
	if config == nil {
		config = NewUnifiedConfigFromEnv()
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		// 继续，使用有效的默认配置
		config = NewConfig(EnvDevelopment)
	}

	return &Container{
		config:     config,
		singletons: make(map[string]interface{}),
		factories:  make(map[string]func() (interface{}, error)),
	}
}

// Register 注册工厂函数
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

// Get 获取组件实例（单例）
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

	// 创建新实例
	instance, err := factory()
	if err != nil {
		return nil, err
	}

	// 存储到单例缓存
	c.mu.Lock()
	c.singletons[name] = instance
	c.mu.Unlock()

	return instance, nil
}

// GetConfig 获取配置对象
func (c *Container) GetConfig() *Config {
	return c.config
}

// GetLogger 获取日志记录器实例
func (c *Container) GetLogger(name string) (*SimpleLogger, error) {
	key := "logger_" + name

	instance, err := c.Get(key)
	if err == nil {
		return instance.(*SimpleLogger), nil
	}

	// 注册并创建新的日志记录器
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

// GetValidator 获取验证器实例
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

// GetSecurityAuditor 获取安全审计器实例
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

// GetRequestGuard 获取请求守卫实例
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

// GetRateLimiter 获取速率限制器实例
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

// Reset 重置容器，清除所有单例
// 注意：仅在测试中使用
func (c *Container) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.singletons = make(map[string]interface{})
	c.initialized = false
}

// Initialize 初始化容器的所有默认组件
// 预先加载关键组件以提高启动性能
func (c *Container) Initialize() error {
	c.mu.Lock()
	if c.initialized {
		c.mu.Unlock()
		return nil
	}
	c.mu.Unlock()

	// 预先初始化关键组件
	// 使用类型化访问器确保正确的工厂注册和初始化
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
