package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/vistone/fingerprint/internal/config"
)

// DemoListener 演示监听器
type DemoListener struct{}

func (dl *DemoListener) OnConfigChange(old, new *config.ManagedConfig, changes []config.ConfigChange) error {
	fmt.Println("Configuration changed!")
	for _, change := range changes {
		fmt.Printf("  Path: %s\n", change.Path)
		fmt.Printf("  Old Value: %v\n", change.OldValue)
		fmt.Printf("  New Value: %v\n", change.NewValue)
	}
	return nil
}

// 示例1：基础配置加载和使用
func ExampleBasicConfigUsage() {
	fmt.Println("=== Example 1: Basic Configuration Usage ===\n")

	// 初始化配置中心
	if err := config.InitializeConfigCenterWithDefaults(); err != nil {
		log.Fatalf("Failed to initialize config: %v\n", err)
	}

	// 获取全局配置中心
	center := config.GetConfigCenter()
	cfg := center.Get()

	// 使用配置
	fmt.Printf("Behavior Analysis Min Requests: %d\n", cfg.BehaviorAnalysis.MinRequestsForAnalysis)
	fmt.Printf("Risk Scoring Critical Threshold: %.2f\n", cfg.RiskScoring.CriticalThreshold)
	fmt.Printf("QUIC Min Initial Max Data: %d\n", cfg.QUIC.MinInitialMaxData)
	fmt.Println()
}

// 示例2：使用配置管理器
func ExampleConfigManager() {
	fmt.Println("=== Example 2: Using Config Manager ===\n")

	center := config.GetConfigCenter()
	manager := config.NewConfigManager(center)

	// 获取特定模块配置
	behaviorCfg := manager.GetBehaviorAnalysisConfig()
	riskCfg := manager.GetRiskScoringConfig()
	featureCfg := manager.GetFeatureExtractionConfig()

	fmt.Printf("Behavior - Min Requests: %d\n", behaviorCfg.MinRequestsForAnalysis)
	fmt.Printf("Behavior - Regularity Threshold: %.2f\n", behaviorCfg.RegularityThreshold)
	fmt.Printf("Risk - Critical Threshold: %.2f\n", riskCfg.CriticalThreshold)
	fmt.Printf("Features - Entropy High Threshold: %.2f\n", featureCfg.EntropyHighThreshold)
	fmt.Println()
}

// 示例3：配置更新
func ExampleConfigUpdate() {
	fmt.Println("=== Example 3: Configuration Update ===\n")

	center := config.GetConfigCenter()
	manager := config.NewConfigManager(center)

	// 获取当前配置
	currentCfg := manager.GetBehaviorAnalysisConfig()
	fmt.Printf("Current Min Requests: %d\n", currentCfg.MinRequestsForAnalysis)

	// 创建新配置
	newCfg := *currentCfg
	newCfg.MinRequestsForAnalysis = 10
	newCfg.RegularityThreshold = 0.4

	// 更新配置
	if err := manager.UpdateBehaviorAnalysisConfig(&newCfg, "increase thresholds", "admin"); err != nil {
		log.Fatalf("Failed to update config: %v\n", err)
	}

	// 验证更新
	updatedCfg := manager.GetBehaviorAnalysisConfig()
	fmt.Printf("Updated Min Requests: %d\n", updatedCfg.MinRequestsForAnalysis)
	fmt.Printf("Updated Regularity Threshold: %.2f\n", updatedCfg.RegularityThreshold)
	fmt.Println()
}

// 示例4：配置验证
func ExampleConfigValidation() {
	fmt.Println("=== Example 4: Configuration Validation ===\n")

	// 创建验证器
	validator := config.NewConfigValidator()

	// 验证当前配置
	center := config.GetConfigCenter()
	cfg := center.Get()

	errs := validator.Validate(cfg)
	if len(errs) == 0 {
		fmt.Println("✓ All validations passed")
	} else {
		fmt.Printf("✗ %d validation errors:\n", len(errs))
		for _, err := range errs {
			fmt.Printf("  - %v\n", err)
		}
	}
	fmt.Println()
}

// 示例5：版本管理和回滚
func ExampleVersionManagement() {
	fmt.Println("=== Example 5: Version Management and Rollback ===\n")

	center := config.GetConfigCenter()

	// 创建几个新版本
	cfg := center.Get()

	for i := 1; i <= 3; i++ {
		newCfg := *cfg
		newCfg.BehaviorAnalysis.MinRequestsForAnalysis = 5 + i
		center.Update(&newCfg, fmt.Sprintf("Version %d update", i), "developer")
	}

	// 查看版本历史
	history := center.GetHistory(10)
	fmt.Printf("Version History (%d versions):\n", len(history))
	for _, vc := range history {
		fmt.Printf("  %s - Changed by: %s, Min Requests: %d\n",
			vc.Version, vc.ChangedBy, vc.Config.BehaviorAnalysis.MinRequestsForAnalysis)
	}

	// 回滚到早期版本
	if len(history) > 2 {
		earlierVersion := history[len(history)-3].Version
		fmt.Printf("\nRolling back to %s...\n", earlierVersion)
		if err := center.Rollback(earlierVersion, "testing rollback", "admin"); err != nil {
			log.Fatalf("Failed to rollback: %v\n", err)
		}
		fmt.Printf("Current Min Requests after rollback: %d\n", center.Get().BehaviorAnalysis.MinRequestsForAnalysis)
	}
	fmt.Println()
}

// 示例6：健康检查
func ExampleHealthCheck() {
	fmt.Println("=== Example 6: Health Check ===\n")

	center := config.GetConfigCenter()
	healthChecker := config.NewHealthChecker(center)

	// 执行健康检查
	result := healthChecker.CheckHealth()

	fmt.Printf("Overall Status: %s\n", result.Status)
	fmt.Printf("Overall Health: %v\n", result.Overall)
	fmt.Printf("Message: %s\n\n", result.Message)

	// 显示详细检查项
	fmt.Println("Detailed Check Results:")
	for _, item := range result.Checks {
		status := "✓"
		if item.Status != config.HealthOK {
			status = "✗"
		}
		fmt.Printf("  %s %s: %s\n", status, item.Name, item.Message)
	}
	fmt.Println()
}

// 示例7：配置变更监听
func ExampleConfigListener() {
	fmt.Println("=== Example 7: Configuration Change Listener ===\n")

	// 注册监听器
	center := config.GetConfigCenter()
	center.RegisterListener(&DemoListener{})

	// 触发配置变更
	cfg := center.Get()
	newCfg := *cfg
	newCfg.BehaviorAnalysis.MinRequestsForAnalysis = 8
	center.Update(&newCfg, "test listener", "admin")
	fmt.Println()
}

// 示例8：实际应用 - QUIC 分析
func ExampleQUICAnalysisWithConfig() {
	fmt.Println("=== Example 8: QUIC Analysis with Configuration ===\n")

	manager := config.GetConfigManager()
	quicCfg := manager.GetQUICConfig()

	// 模拟 QUIC Initial 包分析
	suspiciousMaxData := uint64(100)
	suspiciousStreamData := uint64(50)

	fmt.Printf("QUIC Configuration:\n")
	fmt.Printf("  Min Initial Max Data: %d\n", quicCfg.MinInitialMaxData)
	fmt.Printf("  Min Stream Data: %d\n", quicCfg.MinStreamData)

	fmt.Printf("\nAnalyzing QUIC Initial:\n")
	fmt.Printf("  Initial Max Data: %d\n", suspiciousMaxData)
	fmt.Printf("  Stream Data: %d\n", suspiciousStreamData)

	if suspiciousMaxData < uint64(quicCfg.MinInitialMaxData) ||
		suspiciousStreamData < uint64(quicCfg.MinStreamData) {
		fmt.Println("  ⚠️  SUSPICIOUS_LIMITS detected!")
	}
	fmt.Println()
}

// 示例9：实际应用 - 风险评分
func ExampleRiskScoringWithConfig() {
	fmt.Println("=== Example 9: Risk Scoring with Configuration ===\n")

	manager := config.GetConfigManager()
	riskCfg := manager.GetRiskScoringConfig()

	// 模拟风险评分
	riskScore := 0.85

	fmt.Printf("Risk Scoring Configuration:\n")
	fmt.Printf("  Critical Threshold: %.2f\n", riskCfg.CriticalThreshold)
	fmt.Printf("  High Threshold: %.2f\n", riskCfg.HighThreshold)
	fmt.Printf("  Medium Threshold: %.2f\n", riskCfg.MediumThreshold)

	fmt.Printf("\nEvaluating Risk Score: %.2f\n", riskScore)

	if riskScore >= riskCfg.CriticalThreshold {
		fmt.Println("  🔴 CRITICAL threat level")
	} else if riskScore >= riskCfg.HighThreshold {
		fmt.Println("  🟠 HIGH threat level")
	} else if riskScore >= riskCfg.MediumThreshold {
		fmt.Println("  🟡 MEDIUM threat level")
	} else {
		fmt.Println("  🟢 LOW threat level")
	}
	fmt.Println()
}

func main() {
	separator := strings.Repeat("=", 60)
	fmt.Println("\n" + separator)
	fmt.Println("Configuration Management System Examples")
	fmt.Println(separator + "\n")

	// 运行所有示例
	ExampleBasicConfigUsage()
	ExampleConfigManager()
	ExampleConfigValidation()
	ExampleConfigUpdate()
	ExampleVersionManagement()
	ExampleHealthCheck()
	ExampleQUICAnalysisWithConfig()
	ExampleRiskScoringWithConfig()

	fmt.Println(separator)
	fmt.Println("All examples completed successfully!")
	fmt.Println(separator + "\n")
}
