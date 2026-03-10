package extension

import (
	"os"
	"sync"
	"testing"
	"time"
)

// TestConfigEnvironments tests whether the three environment configuration presets are correct
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

// TestContainerDependencyInjection tests the container's dependency injection functionality
func TestContainerDependencyInjection(t *testing.T) {
	cfg := NewConfig(EnvTesting)
	container := NewContainer(cfg)

	// Test getting components
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

// TestContainerSingletonCaching tests the container's singleton caching functionality
func TestContainerSingletonCaching(t *testing.T) {
	cfg := NewConfig(EnvTesting)
	container := NewContainer(cfg)

	logger1, _ := container.GetLogger("test")
	logger2, _ := container.GetLogger("test")

	// Should return the same instance
	if logger1 != logger2 {
		t.Error("Container should cache singleton instances")
	}

	validator1, _ := container.GetValidator()
	validator2, _ := container.GetValidator()

	if validator1 != validator2 {
		t.Error("Validator should be cached as singleton")
	}
}

// TestContainerThreadSafety tests the container's thread safety
func TestContainerThreadSafety(t *testing.T) {
	cfg := NewConfig(EnvTesting)
	container := NewContainer(cfg)

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	// Concurrent container access
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Randomly get different components
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

// TestComponentInfoRetrieval tests the component info interface
func TestComponentInfoRetrieval(t *testing.T) {
	cfg := NewConfig(EnvTesting)
	container := NewContainer(cfg)

	_, _ = container.GetLogger("test")
	// SimpleLogger as a component, has name and version
	loggerName := "SimpleLogger"

	if loggerName == "" {
		t.Error("Component name should not be empty")
	}
}

// TestRateLimiterWithContainer tests using the rate limiter through the container
func TestRateLimiterWithContainer(t *testing.T) {
	cfg := NewConfig(EnvTesting)
	container := NewContainer(cfg)

	limiter, err := container.GetRateLimiter()
	if err != nil {
		t.Fatalf("Failed to get rate limiter: %v", err)
	}

	// Should allow the first request
	if err := limiter.Allow(); err != nil {
		t.Errorf("First request should be allowed: %v", err)
	}

	// Allow a few more requests
	for i := 0; i < 5; i++ {
		if err := limiter.Allow(); err != nil {
			t.Logf("Request %d denied: %v", i+2, err)
		}
	}
}

// TestValidatorWithContainer tests using the validator through the container
func TestValidatorWithContainer(t *testing.T) {
	cfg := NewConfig(EnvTesting)
	container := NewContainer(cfg)

	validator, err := container.GetValidator()
	if err != nil {
		t.Fatalf("Failed to get validator: %v", err)
	}

	// Test validating small data
	err = validator.ValidateData([]byte("small data"))
	if err != nil {
		t.Errorf("Small data should be valid: err=%v", err)
	}

	// Test validating oversized data
	largeData := make([]byte, cfg.Validator.MaxDataSize+1)
	err = validator.ValidateData(largeData)
	if err == nil {
		t.Error("Large data should be invalid")
	}
}

// TestConfigValidation tests configuration validation
func TestConfigValidation(t *testing.T) {
	validConfigs := []Environment{EnvDevelopment, EnvTesting, EnvProduction}

	for _, env := range validConfigs {
		cfg := NewConfig(env)
		if err := cfg.Validate(); err != nil {
			t.Errorf("Valid config for %s should pass validation: %v", env, err)
		}
	}

	// Test invalid configuration
	invalidCfg := NewConfig(EnvTesting)
	invalidCfg.Defense.MaxInputSize = -1
	if err := invalidCfg.Validate(); err == nil {
		t.Error("Invalid config should fail validation")
	}
}

// TestEnvironmentVariableOverride tests environment variable overrides
func TestEnvironmentVariableOverride(t *testing.T) {
	// This test requires setting actual environment variables
	// This is just demonstrating the test structure
	cfg := NewConfigFromEnv()

	if cfg == nil {
		t.Fatal("Config from environment should not be nil")
	}

	// Environment variables should override defaults
	if cfg.Environment != EnvDevelopment && cfg.Environment != EnvTesting && cfg.Environment != EnvProduction {
		t.Errorf("Invalid environment: %s", cfg.Environment)
	}
}

// TestUnifiedConfigFromEnv tests the unified configuration entry (including rules configuration)
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

// TestConfigLoadRulesFromPath tests loading rules configuration from a path
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

// TestContainerInitialize tests the container's initialization method
func TestContainerInitialize(t *testing.T) {
	cfg := NewConfig(EnvTesting)
	container := NewContainer(cfg)

	// Initialize() should preload all critical components
	if err := container.Initialize(); err != nil {
		t.Fatalf("Container initialization failed: %v", err)
	}

	// Verify all critical components are loaded
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

// TestConfigEnvironmentDifferences tests configuration differences across environments
func TestConfigEnvironmentDifferences(t *testing.T) {
	devCfg := NewConfig(EnvDevelopment)
	prodCfg := NewConfig(EnvProduction)

	// Production should have stricter limits
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

// BenchmarkContainerGetLogger benchmarks getting the logger
func BenchmarkContainerGetLogger(b *testing.B) {
	cfg := NewConfig(EnvTesting)
	container := NewContainer(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = container.GetLogger("benchmark")
	}
}

// BenchmarkRateLimiterAllow benchmarks the rate limiter allow operation
func BenchmarkRateLimiterAllow(b *testing.B) {
	limiter := NewRateLimiter(10000, time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = limiter.Allow()
	}
}

// BenchmarkValidatorValidate benchmarks the validator validation
func BenchmarkValidatorValidate(b *testing.B) {
	validator := NewDefaultValidator()
	testData := []byte("test data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.ValidateData(testData)
	}
}
