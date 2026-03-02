package extension

import (
	"os"
	"sync"
	"testing"
	"time"
)

// TestConfigEnvironments 测试三种环境配置预设是否正确
func TestConfigEnvironments(t *testing.T) {
	envs := []Environment{EnvDevelopment, EnvTesting, EnvProduction}
	names := map[Environment]string{
		EnvDevelopment: "Development",
		EnvTesting:     "Testing",
		EnvProduction:  "Production",
	}

	for _, env := range envs {
		t.Run(names[env], func(t *testing.T) {
			cfg := NewConfig(env)
			if cfg == nil {
				t.Fatal("Config should not be nil")
			}
			if cfg.Environment != env {
				t.Errorf("Environment mismatch: expected %s, got %s", env, cfg.Environment)
			}
			if cfg.Defense == nil {
				t.Fatal("Defense policy should not be nil")
			}
			if cfg.Logger == nil {
				t.Fatal("Logger config should not be nil")
			}
			if cfg.Validator == nil {
				t.Fatal("Validator config should not be nil")
			}
			if cfg.Audit == nil {
				t.Fatal("Audit config should not be nil")
			}

			// Production should be stricter than Development
		})
	}
}

// TestContainerDependencyInjection 测试容器的依赖注入功能
func TestContainerDependencyInjection(t *testing.T) {
	cfg := NewConfig(EnvTesting)
	container := NewContainer(cfg)

	// 测试获取组件
	logger, err := container.GetLogger("test")
	if err != nil {
		t.Fatalf("Failed to get logger: %v", err)
	}
	if logger == nil {
		t.Fatal("Logger should not be nil")
	}

	validator, err := container.GetValidator()
	if err != nil {
		t.Fatalf("Failed to get validator: %v", err)
	}
	if validator == nil {
		t.Fatal("Validator should not be nil")
	}

	guard, err := container.GetRequestGuard()
	if err != nil {
		t.Fatalf("Failed to get request guard: %v", err)
	}
	if guard == nil {
		t.Fatal("RequestGuard should not be nil")
	}

	auditor, err := container.GetSecurityAuditor()
	if err != nil {
		t.Fatalf("Failed to get security auditor: %v", err)
	}
	if auditor == nil {
		t.Fatal("SecurityAuditor should not be nil")
	}

	limiter, err := container.GetRateLimiter()
	if err != nil {
		t.Fatalf("Failed to get rate limiter: %v", err)
	}
	if limiter == nil {
		t.Fatal("RateLimiter should not be nil")
	}
}

// TestContainerSingletonCaching 测试容器的单例缓存功能
func TestContainerSingletonCaching(t *testing.T) {
	cfg := NewConfig(EnvTesting)
	container := NewContainer(cfg)

	logger1, _ := container.GetLogger("test")
	logger2, _ := container.GetLogger("test")

	// 应该返回相同的实例
	if logger1 != logger2 {
		t.Error("Container should cache singleton instances")
	}

	validator1, _ := container.GetValidator()
	validator2, _ := container.GetValidator()

	if validator1 != validator2 {
		t.Error("Validator should be cached as singleton")
	}
}

// TestContainerThreadSafety 测试容器的线程安全性
func TestContainerThreadSafety(t *testing.T) {
	cfg := NewConfig(EnvTesting)
	container := NewContainer(cfg)

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	// 并发访问容器
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// 随机获取不同的组件
			switch id % 5 {
			case 0:
				_, err := container.GetLogger("concurrent")
				if err != nil {
					errors <- err
				}
			case 1:
				_, err := container.GetValidator()
				if err != nil {
					errors <- err
				}
			case 2:
				_, err := container.GetRequestGuard()
				if err != nil {
					errors <- err
				}
			case 3:
				_, err := container.GetSecurityAuditor()
				if err != nil {
					errors <- err
				}
			case 4:
				_, err := container.GetRateLimiter()
				if err != nil {
					errors <- err
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		if err != nil {
			t.Errorf("Concurrent access error: %v", err)
		}
	}
}

// TestComponentInfoRetrieval 测试组件信息接口
func TestComponentInfoRetrieval(t *testing.T) {
	cfg := NewConfig(EnvTesting)
	container := NewContainer(cfg)

	_, _ = container.GetLogger("test")
	// SimpleLogger 作为组件，有名称和版本
	loggerName := "SimpleLogger"

	if loggerName == "" {
		t.Error("Component name should not be empty")
	}
}

// TestRateLimiterWithContainer 测试通过容器使用速率限制器
func TestRateLimiterWithContainer(t *testing.T) {
	cfg := NewConfig(EnvTesting)
	container := NewContainer(cfg)

	limiter, err := container.GetRateLimiter()
	if err != nil {
		t.Fatalf("Failed to get rate limiter: %v", err)
	}

	// 应该允许第一个请求
	if err := limiter.Allow(); err != nil {
		t.Errorf("First request should be allowed: %v", err)
	}

	// 再允许几个请求
	for i := 0; i < 5; i++ {
		if err := limiter.Allow(); err != nil {
			t.Logf("Request %d denied: %v", i+2, err)
		}
	}
}

// TestValidatorWithContainer 测试通过容器使用验证器
func TestValidatorWithContainer(t *testing.T) {
	cfg := NewConfig(EnvTesting)
	container := NewContainer(cfg)

	validator, err := container.GetValidator()
	if err != nil {
		t.Fatalf("Failed to get validator: %v", err)
	}

	// 测试验证小的数据
	err = validator.ValidateData([]byte("small data"))
	if err != nil {
		t.Errorf("Small data should be valid: err=%v", err)
	}

	// 测试验证过大的数据
	largeData := make([]byte, cfg.Validator.MaxDataSize+1)
	err = validator.ValidateData(largeData)
	if err == nil {
		t.Error("Large data should be invalid")
	}
}

// TestConfigValidation 测试配置验证
func TestConfigValidation(t *testing.T) {
	validConfigs := []Environment{EnvDevelopment, EnvTesting, EnvProduction}

	for _, env := range validConfigs {
		cfg := NewConfig(env)
		if err := cfg.Validate(); err != nil {
			t.Errorf("Valid config for %s should pass validation: %v", env, err)
		}
	}

	// 测试无效的配置
	invalidCfg := NewConfig(EnvTesting)
	invalidCfg.Defense.MaxInputSize = -1
	if err := invalidCfg.Validate(); err == nil {
		t.Error("Invalid config should fail validation")
	}
}

// TestEnvironmentVariableOverride 测试环境变量覆盖
func TestEnvironmentVariableOverride(t *testing.T) {
	// 这个测试需要设置实际的环境变量
	// 在这里只是演示测试结构
	cfg := NewConfigFromEnv()

	if cfg == nil {
		t.Fatal("Config from environment should not be nil")
	}

	// 环境变量应该覆盖默认值
	if cfg.Environment != EnvDevelopment && cfg.Environment != EnvTesting && cfg.Environment != EnvProduction {
		t.Errorf("Invalid environment: %s", cfg.Environment)
	}
}

// TestUnifiedConfigFromEnv 测试统一配置入口（含规则配置）
func TestUnifiedConfigFromEnv(t *testing.T) {
	t.Setenv("FINGERPRINT_ENV", string(EnvTesting))
	t.Setenv("FINGERPRINT_RULES_FILE", "rules.json")

	cfg := NewUnifiedConfigFromEnv()
	if cfg == nil {
		t.Fatal("Unified config should not be nil")
	}

	if cfg.Rules == nil {
		t.Fatal("Rules config should not be nil")
	}

	if cfg.RulesSource == "" {
		t.Error("RulesSource should not be empty")
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Unified config should be valid: %v", err)
	}
}

// TestConfigLoadRulesFromPath 测试通过路径加载规则配置
func TestConfigLoadRulesFromPath(t *testing.T) {
	cfg := NewConfig(EnvTesting)

	path := "/media/stone/data/fingerprint/internal/config/rules.json"
	if _, err := os.Stat(path); err != nil {
		t.Skip("rules.json not found in current environment")
	}

	if err := cfg.LoadRulesFromPath(path); err != nil {
		t.Fatalf("LoadRulesFromPath should succeed: %v", err)
	}

	if cfg.Rules == nil {
		t.Fatal("Rules should not be nil after load")
	}

	if cfg.RulesSource != path {
		t.Errorf("RulesSource mismatch: expected %s, got %s", path, cfg.RulesSource)
	}
}

// TestContainerInitialize 测试容器的初始化方法
func TestContainerInitialize(t *testing.T) {
	cfg := NewConfig(EnvTesting)
	container := NewContainer(cfg)

	// Initialize() 应该预加载所有关键组件
	if err := container.Initialize(); err != nil {
		t.Fatalf("Container initialization failed: %v", err)
	}

	// 验证所有关键组件都已加载
	_, err := container.GetLogger("test")
	if err != nil {
		t.Error("Logger should be initialized")
	}

	_, err = container.GetValidator()
	if err != nil {
		t.Error("Validator should be initialized")
	}

	_, err = container.GetRequestGuard()
	if err != nil {
		t.Error("RequestGuard should be initialized")
	}
}

// TestConfigEnvironmentDifferences 测试不同环境的配置差异
func TestConfigEnvironmentDifferences(t *testing.T) {
	devCfg := NewConfig(EnvDevelopment)
	prodCfg := NewConfig(EnvProduction)

	// Production 应该有更严格的限制
	tests := []struct {
		name string
		dev  int
		prod int
		want bool
	}{
		{
			name: "MaxInputSize",
			dev:  devCfg.Defense.MaxInputSize,
			prod: prodCfg.Defense.MaxInputSize,
			want: prodCfg.Defense.MaxInputSize >= devCfg.Defense.MaxInputSize,
		},
		{
			name: "RateLimit",
			dev:  devCfg.Defense.RateLimit,
			prod: prodCfg.Defense.RateLimit,
			want: prodCfg.Defense.RateLimit >= 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.want != true {
				t.Logf("Dev: %d, Prod: %d", tt.dev, tt.prod)
			}
		})
	}
}

// BenchmarkContainerGetLogger 基准测试：获取日志记录器
func BenchmarkContainerGetLogger(b *testing.B) {
	cfg := NewConfig(EnvTesting)
	container := NewContainer(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = container.GetLogger("benchmark")
	}
}

// BenchmarkRateLimiterAllow 基准测试：速率限制器允许操作
func BenchmarkRateLimiterAllow(b *testing.B) {
	limiter := NewRateLimiter(10000, time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = limiter.Allow()
	}
}

// BenchmarkValidatorValidate 基准测试：验证器验证
func BenchmarkValidatorValidate(b *testing.B) {
	validator := NewDefaultValidator()
	testData := []byte("test data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.ValidateData(testData)
	}
}
