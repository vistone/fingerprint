package extension

import (
	"testing"
)

// ========================================================================
// 集成示例的单元测试
// ========================================================================

// TestExample1_BasicUsage 测试基础示例能否运行
func TestExample1_BasicUsage(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Example 1 panic: %v", r)
		}
	}()

	Example1_BasicUsage()
	t.Log("✅ Example 1 通过")
}

// TestExample2_LoggingMiddleware 测试日志中间件示例
func TestExample2_LoggingMiddleware(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Example 2 panic: %v", r)
		}
	}()

	Example2_LoggingMiddleware()
	t.Log("✅ Example 2 通过")
}

// TestExample3_CachingMiddleware 测试缓存中间件示例
func TestExample3_CachingMiddleware(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Example 3 panic: %v", r)
		}
	}()

	Example3_CachingMiddleware()
	t.Log("✅ Example 3 通过")
}

// TestExample4_ABTest 测试 A/B 测试示例
func TestExample4_ABTest(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Example 4 panic: %v", r)
		}
	}()

	Example4_ABTest()
	t.Log("✅ Example 4 通过")
}

// TestExample5_SelectiveStrategy 测试迁移策略示例
func TestExample5_SelectiveStrategy(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Example 5 panic: %v", r)
		}
	}()

	Example5_SelectiveStrategy()
	t.Log("✅ Example 5 通过")
}

// TestShowBestPractices 测试最佳实践展示
func TestShowBestPractices(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("BestPractices panic: %v", r)
		}
	}()

	ShowBestPractices()
	t.Log("✅ BestPractices 通过")
}

// TestShowMigrationTimeline 测试迁移时间表展示
func TestShowMigrationTimeline(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("MigrationTimeline panic: %v", r)
		}
	}()

	ShowMigrationTimeline()
	t.Log("✅ MigrationTimeline 通过")
}

// ========================================================================
// ZapLoggerAdapter 的单元测试
// ========================================================================
// 注意：ZapLoggerAdapter 已在 pipeline_middleware_test.go 中定义和测试
// 本文件不再重复定义和测试

// ========================================================================
// 场景选择函数的测试
// ========================================================================

// TestShouldUseProcessWithPipeline 测试场景决策函数
func TestShouldUseProcessWithPipeline(t *testing.T) {
	testCases := []struct {
		name     string
		scenario *ProcessingScenario
		expected bool
	}{
		{
			name: "NeedsDetailedLogging",
			scenario: &ProcessingScenario{
				NeedsDetailedLogging: true,
			},
			expected: true,
		},
		{
			name: "NeedsDistributedTracing",
			scenario: &ProcessingScenario{
				NeedsDistributedTracing: true,
			},
			expected: true,
		},
		{
			name: "NeedsMetrics",
			scenario: &ProcessingScenario{
				NeedsMetrics: true,
			},
			expected: true,
		},
		{
			name: "NeedsCaching",
			scenario: &ProcessingScenario{
				NeedsCaching: true,
			},
			expected: true,
		},
		{
			name: "HighConcurrency_LowLatency",
			scenario: &ProcessingScenario{
				IsHighConcurrencyPath: true,
				CanTolerateLatency:    false,
			},
			expected: false,
		},
		{
			name: "UltraPerformanceSensitive",
			scenario: &ProcessingScenario{
				IsUltraPerformanceSensitive: true,
			},
			expected: false,
		},
		{
			name:     "Default",
			scenario: &ProcessingScenario{},
			expected: true,
		},
		{
			name:     "NilScenario",
			scenario: nil,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ShouldUseProcessWithPipeline(tc.scenario)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}

	t.Logf("✅ 场景决策函数: %d 个测试用例通过", len(testCases))
}

// ========================================================================
// 集成测试
// ========================================================================

// TestIntegration_AllExamplesRunWithoutPanic 集成测试：所有示例都能运行
func TestIntegration_AllExamplesRunWithoutPanic(t *testing.T) {
	// 这个测试验证所有示例都能在不 panic 的情况下执行

	examples := []struct {
		name string
		fn   func()
	}{
		{"Example1", Example1_BasicUsage},
		{"Example2", Example2_LoggingMiddleware},
		{"Example3", Example3_CachingMiddleware},
		{"Example4", Example4_ABTest},
		{"Example5", Example5_SelectiveStrategy},
		{"BestPractices", ShowBestPractices},
		{"MigrationTimeline", ShowMigrationTimeline},
	}

	for _, ex := range examples {
		t.Run(ex.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("%s panic: %v", ex.name, r)
				}
			}()

			ex.fn()
		})
	}

	t.Logf("✅ 所有 %d 个示例都能正常执行", len(examples))
}
