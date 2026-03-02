# Feature Extractor 快速开始指南

## 概述

本指南说明如何在现有和新的代码中使用 P0 阶段新引入的统一 Feature Extractor 框架。

## 快速开始

### 1. 使用默认配置

最简单的方式 - 无需任何配置文件：

```go
package main

import (
  "fingerprint/internal/features"
)

func main() {
  // 创建使用默认配置的提取器
  extractor := features.NewBaseFeatureExtractor(nil)
  
  // 检测无头浏览器
  userAgent := "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 HeadlessChrome/120.0"
  score, isBot := extractor.ExtractFeature(
    features.FeatureHeadlessBrowser,
    userAgent,
    nil,
  )
  
  if isBot {
    println("Bot detected with risk score:", score)
  }
}
```

### 2. 从 JSON 配置加载

使用可配置规则（支持动态更新）：

```go
package main

import (
  "fingerprint/internal/extension"
  "fingerprint/internal/features"
)

func main() {
  // 统一配置入口（运行配置 + 规则配置）
  appConfig := extension.NewUnifiedConfigFromEnv()
  rulesConfig := appConfig.Rules
  
  // 转换为特征提取配置
  featureConfig := &features.FeatureConfig{
    EntropyHighThreshold:   rulesConfig.Entropy.HighThreshold,
    EntropyLowThreshold:    rulesConfig.Entropy.LowThreshold,
    ToolMarkers:            getToolMarkers(rulesConfig),
    HeadlessMarkers:        rulesConfig.HeadlessBrowserUA.Markers,
    MobileScreenWidthMax:   rulesConfig.MobileScreenContradiction.MobileScreenWidthMax,
    DesktopScreenWidthMin:  rulesConfig.MobileScreenContradiction.DesktopScreenWidthMin,
  }
  
  extractor := features.NewBaseFeatureExtractor(featureConfig)
  
  // ... 使用 extractor
}

func getToolMarkers(cfg *config.RulesConfig) []string {
  markers := make([]string, len(cfg.ToolMarkers.Patterns))
  for i, p := range cfg.ToolMarkers.Patterns {
    markers[i] = p.Marker
  }
  return markers
}
```

### 3. 提取完整特征向量

一次获取所有特征和综合风险评分：

```go
package main

import (
  "fingerprint/internal/features"
  "fmt"
)

func main() {
  extractor := features.NewBaseFeatureExtractor(nil)
  
  // 准备多源数据
  data := map[string]interface{}{
    "raw_binary": []byte("test data with patterns..."),
    "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0",
    "os": "Windows",
    "platform": "Win32",
    "is_mobile": "false",
    "screen_width": "1920",
  }
  
  // 一次提取所有特征
  vector := extractor.ExtractFeatureVector(data, nil)
  
  fmt.Printf("Risk Score: %.2f\n", vector.RiskScore)
  fmt.Printf("Anomalies: %v\n", vector.Anomalies)
  fmt.Printf("Feature Scores:\n")
  for featureType, score := range vector.Scores {
    fmt.Printf("  %s: %.2f\n", featureType, score)
  }
  fmt.Printf("Hash (for deduplication): %s\n", vector.Hash)
}
```

## 特征类型参考

| 特征类型 | 说明 | 应用场景 |
|---------|------|---------|
| `FeatureEntropy` | 数据熵异常（低/高） | 检测重复模式或过度随机化 |
| `FeatureToolMarker` | 自动化工具特征 | 识别 Selenium、Puppeteer 等工具 |
| `FeatureHeadlessBrowser` | 无头浏览器 UA | 检测 HeadlessChrome、PhantomJS |
| `FeatureOSPlatformContradiction` | OS 与 Platform 矛盾 | 检测伪造的操作系统声明 |
| `FeatureUAOSContradiction` | UA 与 OS 矛盾 | 检测 UA 字符串与系统不匹配 |
| `FeatureMobileScreenContradiction` | 移动/屏幕分辨率矛盾 | 检测移动设备用超大分辨率等 |
| `FeatureUAFeatureContradiction` | UA 与特性矛盾 | 检测旧浏览器声称新特性等 |

## 现有代码迁移

### 如果使用了 defense.go

使用兼容适配器无需修改代码：

```go
// 旧代码保持不变，但现在使用更优的检测引擎
import "fingerprint/internal/features"

// 创建适配器
adapter := features.NewLegacyFeatureAdapter(nil)

// 所有原有 API 都工作，但内部使用新的 Feature Extractor
if adapter.DetectHeadlessBrowser("HeadlessChrome/120") {
  // ... 处理
}

attrs := map[string]string{
  "os": "Windows",
  "platform": "Mac", // 矛盾
}
if adapter.CheckContradictions(attrs) {
  // ... 处理
}
```

## 配置文件格式

规则配置文件：默认 `rules.json`（由 `extension.NewUnifiedConfigFromEnv()` 统一加载）

**关键配置项**：

```json
{
  "entropy": {
    "enabled": true,
    "high_threshold": 7.5,      // Shannon 熵阈值（bits）
    "low_threshold": 26         // 最少 unique bytes
  },
  "tool_markers": {
    "enabled": true,
    "patterns": [
      { "marker": "HeadlessChrome", "severity": "critical" },
      { "marker": "webdriver", "severity": "critical" }
    ]
  },
  "mobile_screen_contradiction": {
    "mobile_screen_width_max": 1920,    // 移动设备最大宽度
    "desktop_screen_width_min": 800     // 桌面设备最小宽度
  }
}
```

## 常见用例

### 用例 1: 检测机器人流量

```go
extractor := features.NewBaseFeatureExtractor(nil)

headers := map[string]string{
  "User-Agent": request.Header.Get("User-Agent"),
  "Accept": request.Header.Get("Accept"),
}

// 检测无头浏览器
if score, isBot := extractor.ExtractFeature(
  features.FeatureHeadlessBrowser,
  headers["User-Agent"],
  nil,
); isBot && score > 0.8 {
  return blockRequest()
}
```

### 用例 2: 全面指纹异常检查

```go
extractor := features.NewBaseFeatureExtractor(nil)

data := map[string]interface{}{
  "user_agent": request.UserAgent(),
  "os": detectOS(request),
  "platform": request.Header.Get("Sec-CH-UA-Platform"),
  "is_mobile": detectMobile(request),
  "screen_width": request.Header.Get("Viewport-Width"),
}

vector := extractor.ExtractFeatureVector(data, nil)

// 多重异常提高可信度
if vector.RiskScore > 0.7 && len(vector.Anomalies) > 2 {
  triggerSecurityEvent(vector)
}
```

### 用例 3: 自定义规则阈值

```go
customConfig := &features.FeatureConfig{
  EntropyHighThreshold: 8.0,      // 更严格
  EntropyLowThreshold:  20,       // 更严格
  MobileScreenWidthMax: 1440,     // 调整分辨率
}

extractor := features.NewBaseFeatureExtractor(customConfig)

// 使用自定义配置
vector := extractor.ExtractFeatureVector(data, customConfig)
```

## 性能考虑

- **单特征提取** (~1-10µs): 适合实时决策
- **完整向量** (~50-100µs): 适合离线分析
- **配置加载** (~1ms): 建议在启动时缓存，不在请求路径中

## 故障排除

### Q: 配置文件找不到？

A: 使用 `extension.NewUnifiedConfigFromEnv()`，并通过 `FINGERPRINT_RULES_FILE` 指定规则文件路径

### Q: 评分总是 0？

A: 检查特征类型是否与数据类型匹配（如 `FeatureEntropy` 需要 `[]byte` 或 `string`）

### Q: 如何调整敏感度？

A: 创建自定义 `FeatureConfig` 并修改阈值参数

## 下一步

- 查看 [P0 完整实施文档](04-p0-feature-extractor-guide.md)
- 阅读源代码注释（`internal/features/extractor.go`）
- 浏览规则配置格式详解（默认 `rules.json`，统一入口 `extension.NewUnifiedConfigFromEnv()`）

---

**相关链接**  
- [缺口分析](../../1-analysis/02-fingerprint-gap-analysis-2026-02-28.md)
- [开发流程规则](01-fingerprint-project-rules.md)
