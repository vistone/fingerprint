# 配置管理系统完整指南

## 🎯 设计目标

完整的配置管理系统，支持：
- ✅ 動態下發配置更新
- ✅ 配置熱更新（無需重啟）
- ✅ 版本管理與回滾機制
- ✅ 配置校驗和健康檢查
- ✅ 消除硬編碼配置值

## 📦 核心组件

### 1. ConfigCenter - 配置中心
集中管理全部配置，提供加载、更新、版本管理功能。

```go
import "fingerprint/internal/config"

// 初始化
center := config.NewConfigCenter("path/to/config.json")
if err := center.Load(); err != nil {
    log.Fatal(err)
}

// 获取当前配置
cfg := center.Get()
fmt.Println(cfg.BehaviorAnalysis.MinRequestsForAnalysis)  // 5
```

### 2. ConfigManager - 配置管理器
提供类型安全的配置访问接口。

```go
manager := config.NewConfigManager(center)

// 获取特定模块配置
behaviorCfg := manager.GetBehaviorAnalysisConfig()
riskCfg := manager.GetRiskScoringConfig()

// 获取特定配置值
val, _ := manager.GetConfigValue("behavior_analysis.min_requests")
fmt.Println(val)  // 5
```

### 3. ConfigValidator - 配置验证器
验证配置的完整性和正确性。

```go
validator := config.NewConfigValidator()

// 验证完整配置
errs := validator.Validate(newConfig)
if len(errs) > 0 {
    for _, err := range errs {
        fmt.Println(err)
    }
}

// 添加自定义验证规则
validator.AddRule("custom_check", func(cfg *config.ManagedConfig) error {
    if cfg.BehaviorAnalysis.MinRequestsForAnalysis < 3 {
        return fmt.Errorf("min_requests must be >= 3")
    }
    return nil
})
```

### 4. HealthChecker - 健康检查
监控配置系统的整体健康状态。

```go
healthChecker := config.NewHealthChecker(center)

// 执行健康检查
result := healthChecker.CheckHealth()
fmt.Printf("Status: %s\n", result.Status)  // "ok", "warning", "critical"
fmt.Printf("Overall: %v\n", result.Overall)

// 查看详细检查项
for _, item := range result.Checks {
    fmt.Printf("✓ %s: %s\n", item.Name, item.Message)
}
```

## 🔄 配置更新流程

### A. 动态配置更新（无需重启）

```go
// 1. 创建新配置
newConfig := center.Get()  // 获取当前配置副本
newConfig.BehaviorAnalysis.MinRequestsForAnalysis = 10

// 2. 验证配置
validator := config.NewConfigValidator()
if errs := validator.Validate(newConfig); len(errs) > 0 {
    log.Fatal("validation failed")
}

// 3. 应用更新
if err := center.Update(newConfig, "increase min requests", "admin"); err != nil {
    log.Fatal(err)
}

// 更改立即生效，无需重启应用
```

### B. 版本管理与回滚

```go
// 查看版本历史
history := center.GetHistory(10)  // 获取最近10个版本
for _, vc := range history {
    fmt.Printf("Version: %s\n", vc.Version)
    fmt.Printf("Timestamp: %v\n", vc.Timestamp)
    fmt.Printf("Changed by: %s\n", vc.ChangedBy)
    fmt.Printf("Reason: %s\n", vc.ChangeReason)
}

// 回滚到之前的版本
if err := center.Rollback("v5", "revert problematic config", "admin"); err != nil {
    log.Fatal(err)
}
```

## 📡 配置变更监听与热更新

### 注册监听器

```go
// 自定义监听器
type MyConfigListener struct{}

func (mcl *MyConfigListener) OnConfigChange(old, new *config.ManagedConfig, changes []config.ConfigChange) error {
    fmt.Println("Config changed!")
    
    for _, change := range changes {
        fmt.Printf("  Path: %s\n", change.Path)
        fmt.Printf("  Old: %v → New: %v\n", change.OldValue, change.NewValue)
    }
    
    // 根据变更应用新配置
    // ...
    
    return nil
}

// 注册监听器
center.RegisterListener(&MyConfigListener{})

// 现在任何配置更新都会自动触发监听器
```

### 启用自动重新加载

```go
// 每30秒检查一次配置文件是否有改动
// 如果有改动，自动加载并通知所有监听器
center.EnableAutoReload(30 * time.Second)

// 禁用自动重新加载
center.DisableAutoReload()
```

## 🔧 配置结构

### BehaviorAnalysisConfig
```json
{
  "behavior_analysis": {
    "min_requests_for_analysis": 5,
    "regularity_threshold": 0.3,
    "entropy_threshold": 0.5,
    "anomalous_interval_rate_threshold": 0.2,
    "request_history_capacity": 100,
    "signal_capacity": 50
  }
}
```

### RiskScoringConfig
```json
{
  "risk_scoring": {
    "critical_threshold": 0.9,
    "high_threshold": 0.7,
    "medium_threshold": 0.5,
    "low_threshold": 0.3,
    "min_confidence": 0.5,
    "weights": {
      "headless": 0.25,
      "anomaly": 0.20,
      "contradiction": 0.15,
      "ech": 0.10,
      "client_hints": 0.10,
      "behavior_anomaly": 0.10,
      "cipher_suite_risk": 0.05,
      "extension_anomaly": 0.05
    }
  }
}
```

### FeatureExtractionConfig
```json
{
  "features": {
    "entropy_high_threshold": 7.5,
    "entropy_low_threshold": 26,
    "tool_markers": ["HeadlessChrome", "PhantomJS", "webdriver", "selenium"],
    "headless_markers": ["headlesschrome", "phantomjs", "selenium", "webdriver"],
    "mobile_screen_width_max": 1920,
    "desktop_screen_width_min": 800
  }
}
```

## 🚀 全局初始化

### 方法1：从配置文件初始化（推荐）

```go
package main

import (
    "log"
    "fingerprint/internal/config"
)

func init() {
    if err := config.InitializeConfigCenter(); err != nil {
        log.Fatal(err)
    }
}

func main() {
    // 使用全局配置中心
    center := config.GetConfigCenter()
    cfg := center.Get()
    
    // 或使用配置管理器
    manager := config.GetConfigManager()
    behaviorCfg := manager.GetBehaviorAnalysisConfig()
}
```

### 方法2：使用默认配置初始化

```go
// 当配置文件不可用时使用默认配置
if err := config.InitializeConfigCenterWithDefaults(); err != nil {
    log.Fatal(err)
}
```

## 💡 实际应用示例

### 示例1：在 BehaviorAnalyzer 中使用配置中心

**原始代码（硬编码）：**
```go
func NewBehaviorAnalyzer(config *BehaviorAnalysisConfig) *BehaviorAnalyzer {
    if config == nil {
        config = &BehaviorAnalysisConfig{
            MinRequestsForAnalysis: 5,  // ❌ 硬编码
            RegularityThreshold:    0.3, // ❌ 硬编码
            EntropyThreshold:       0.5, // ❌ 硬编码
        }
    }
    // ...
}
```

**改进代码（配置中心）：**
```go
import "fingerprint/internal/config"

func NewBehaviorAnalyzer(cfg *BehaviorAnalysisConfig) *BehaviorAnalyzer {
    if cfg == nil {
        // 从配置中心读取
        manager := config.GetConfigManager()
        cfg = manager.GetBehaviorAnalysisConfig()
    }
    // ...
}
```

### 示例2：在 QUIC 分析中使用配置中心

**改进代码：**
```go
func (qs *QUICSignatureAnalyzer) AnalyzeQUICInitial(initial QUICInitialData) (*QUICResult, error) {
    // 从配置中心读取限制
    manager := config.GetConfigManager()
    quicCfg := manager.GetQUICConfig()
    
    // 使用配置的最小值限制，而不是硬编码的 1024, 256
    if initial.InitialMaxData < uint64(quicCfg.MinInitialMaxData) || 
       initial.InitialMaxStreamData < uint64(quicCfg.MinStreamData) {
        result.AnomalyFlags = append(result.AnomalyFlags, "SUSPICIOUS_LIMITS")
    }
    
    // ...
}
```

### 示例3：动态更新风险评分阈值

```go
// 运行时动态调整威胁级别阈值
func UpdateRiskThresholds(critical, high, medium, low float64) error {
    manager := config.GetConfigManager()
    
    cfg := manager.GetRiskScoringConfig()
    cfg.CriticalThreshold = critical
    cfg.HighThreshold = high
    cfg.MediumThreshold = medium
    cfg.LowThreshold = low
    
    return manager.UpdateRiskScoringConfig(
        cfg,
        "adjust risk thresholds for higher sensitivity",
        "operator",
    )
}
```

## 📊 健康检查示例

```go
func PerformHealthCheck() {
    healthChecker := config.GetHealthChecker()
    result := healthChecker.CheckHealth()
    
    if result.Overall {
        fmt.Println("✅ Configuration system is healthy")
    } else {
        fmt.Println("⚠️  Configuration system has issues:")
        for _, check := range result.Checks {
            if check.Status != config.HealthOK {
                fmt.Printf("  ❌ %s: %s\n", check.Name, check.Message)
            }
        }
    }
}
```

## ⚙️ 环境变量支持

```bash
# 指定配置文件路径
export FINGERPRINT_CONFIG_PATH=/etc/fingerprint/config.json

# 启用调试模式
export FINGERPRINT_DEBUG=true

# 启用自动重新加载
export FINGERPRINT_AUTO_RELOAD=true
export FINGERPRINT_RELOAD_INTERVAL=30
```

## 🔐 最佳实践

1. **集中化配置管理**
   - 所有配置都应该从配置中心读取
   - 避免在代码中硬编码配置值

2. **版本控制**
   - 保留配置历史用于审计和调试
   - 定期检查历史版本

3. **验证配置**
   - 在更新前始终验证新配置
   - 添加自定义验证规则以确保数据一致性

4. **监听变更**
   - 为关键模块注册配置变更监听器
   - 实现优雅的配置热更新

5. **健康检查**
   - 定期执行健康检查
   - 监控配置文件可访问性
   - 跟踪配置版本历史

## 📝 迁移指南

### 步骤1：初始化配置系统
在应用启动时调用 `InitializeConfigCenter()`

### 步骤2：更新代码以使用配置中心
将所有硬编码的配置值替换为：
```go
manager := config.GetConfigManager()
cfg := manager.GetBehaviorAnalysisConfig()
minRequests := cfg.MinRequestsForAnalysis
```

### 步骤3：测试配置更新
验证应用能够正确处理动态配置更新

### 步骤4：部署配置文件
将 `internal/config/config.json` 部署到生产环境

## 🐛 故障排除

### 配置加载失败
```go
center := config.GetConfigCenter()
if !center.IsLoaded() {
    log.Fatal("Configuration not loaded")
}
```

### 验证失败
```go
validator := config.NewConfigValidator()
errs := validator.Validate(cfg)
for _, err := range errs {
    log.Printf("Validation error: %v", err)
}
```

### 健康检查报警
```go
healthChecker := config.GetHealthChecker()
result := healthChecker.CheckHealth()
for _, item := range result.Checks {
    if item.Status != config.HealthOK {
        log.Printf("Health warning: %s - %s", item.Name, item.Message)
    }
}
```

---

更多信息请参考 API 文档和示例代码。
