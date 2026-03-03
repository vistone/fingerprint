# 特征提取框架 - 快速参考卡

## 导入与初始化

```go
import (
  "fingerprint/internal/features"
  "fingerprint/internal/extension"
)

// 方式 1: 默认配置（推荐快速开始）
extractor := features.NewBaseFeatureExtractor(nil)

// 方式 2: 从 JSON 加载
appConfig := extension.NewUnifiedConfigFromEnv()
rulesConfig := appConfig.Rules
featureConfig := convertToFeatureConfig(rulesConfig) // 转换函数
extractor := features.NewBaseFeatureExtractor(featureConfig)

// 方式 3: 自定义配置
customConfig := &features.FeatureConfig{
  EntropyHighThreshold: 8.0,
  EntropyLowThreshold:  20,
  // ... 更多字段
}
extractor := features.NewBaseFeatureExtractor(customConfig)

// 方式 4: 兼容层（零迁移成本）
adapter := features.NewLegacyFeatureAdapter(nil)
```

---

## 特征提取 API

### 单个特征

```go
score, isAnomaly := extractor.ExtractFeature(
  features.FeatureHeadlessBrowser,  // 特征类型
  "Mozilla/5.0 HeadlessChrome",    // 输入数据
  nil,                              // 配置（nil 使用已初始化的）
)
// score: 0.0-1.0，isAnomaly: true/false
```

### 完整向量

```go
vector := extractor.ExtractFeatureVector(data, nil)
// 返回: &FeatureVector{
//   Scores: map[FeatureType]float64,    // 所有特征评分
//   Anomalies: []FeatureType,           // 异常特征列表
//   RiskScore: float64,                 // 综合风险评分
//   Hash: string,                       // 去重哈希
// }
```

---

## 特征类型速查表

| 特征类型 | 快速引用 | 输入类型 | 应用 |
|---------|--------|--------|------|
| 异常熵 | `FeatureEntropy` | `[]byte` \| `string` | 数据异常检测 |
| 工具标记 | `FeatureToolMarker` | `[]byte` \| `string` | 自动化工具识别 |
| 无头浏览器 | `FeatureHeadlessBrowser` | `string` \| `map[string]string` | 机器人检测 |
| OS/Platform 矛盾 | `FeatureOSPlatformContradiction` | `map[string]string` | 伪装检测 |
| UA/OS 矛盾 | `FeatureUAOSContradiction` | `map[string]string` | 指纹矛盾检测 |
| Mobile/Screen 矛盾 | `FeatureMobileScreenContradiction` | `map[string]string` | 设备一致性检查 |
| UA/Feature 矛盾 | `FeatureUAFeatureContradiction` | `map[string]string` | 兼容性检查 |

---

## 常用代码片段

### 1. 检测无头浏览器

```go
isBot := func(ua string) bool {
  score, _ := extractor.ExtractFeature(features.FeatureHeadlessBrowser, ua, nil)
  return score > 0.8
}

if isBot(userAgent) {
  blockRequest()
}
```

### 2. 全面异常检查

```go
data := map[string]interface{}{
  "user_agent": request.UserAgent(),
  "os": detectOS(request),
  "platform": request.Header.Get("Sec-CH-UA-Platform"),
  "is_mobile": detectMobile(request),
  "screen_width": request.Header.Get("Viewport-Width"),
}

vector := extractor.ExtractFeatureVector(data, nil)

if vector.RiskScore > 0.7 {
  triggerSecurityAlert(vector)
}
```

### 3. 多维度矛盾检测

```go
attributes := map[string]string{
  "os": "Windows",
  "platform": request.Header.Get("Platform"),
  "user_agent": request.UserAgent(),
  "is_mobile": strconv.FormatBool(isMobile),
  "screen_width": screenWidth,
}

// 逐一检测各类矛盾
contradictions := []features.FeatureType{
  features.FeatureOSPlatformContradiction,
  features.FeatureUAOSContradiction,
  features.FeatureMobileScreenContradiction,
}

anomalyCount := 0
for _, fType := range contradictions {
  _, isAnomaly := extractor.ExtractFeature(fType, attributes, nil)
  if isAnomaly {
    anomalyCount++
  }
}

if anomalyCount >= 2 {
  raiseSecurityEvent()
}
```

### 4. 获取特征人类可读名称

```go
name := extractor.GetFeatureName(features.FeatureHeadlessBrowser)
// 返回: "Headless Browser Detection"
```

---

## 配置参数默认值

```go
// 熵特征
EntropyHighThreshold: 7.5     // Shannon 熵（bits）
EntropyLowThreshold: 26        // 最少 unique bytes

// 工具标记（默认 5 个）
ToolMarkers: []string{
  "HeadlessChrome", "PhantomJS", "webdriver", "selenium", "puppeteer",
}

// 无头浏览器标记（默认 10 个）
HeadlessMarkers: []string{
  "headlesschrome", "phantomjs", "selenium", "webdriver", "puppeteer",
  "playwright", "cypress", "jsdom", "zombie", "htmlunit",
}

// 屏幕分辨率
MobileScreenWidthMax: 1920      // 移动设备最大宽度
DesktopScreenWidthMin: 800      // 桌面设备最小宽度
```

---

## 风险评分解读

```
分数范围     风险等级     建议行动
[0.0, 0.3)  ✅ 低       正常放行
[0.3, 0.6)  ⚠️ 中       记录日志
[0.6, 0.8)  🔴 高       需要验证
[0.8, 1.0]  🚫 极危险   拦截请求
```

---

## 兼容层 API（defense.go）

```go
adapter := features.NewLegacyFeatureAdapter(nil)

// 所有原有方法继续工作
adapter.DetectHeadlessBrowser(ua)
adapter.DetectAnomalies([]byte{...})
adapter.CheckContradictions(map[string]string{...})
adapter.RecognizeFromHeaders(map[string]string{...})
```

---

## 调试技巧

### 查看特征向量详情

```go
vector := extractor.ExtractFeatureVector(data, nil)
fmt.Printf("Risk Score: %.2f\n", vector.RiskScore)
fmt.Printf("Anomalies: %v\n", vector.Anomalies)
fmt.Printf("Hash: %s (for dedup)\n", vector.Hash)

// 逐特征查看评分
for fType, score := range vector.Scores {
  fmt.Printf("%s: %.2f\n", fType, score)
}
```

### 特征类型名称

```go
// 获取人类可读的特征名称
for _, fType := range vector.Anomalies {
  name := extractor.GetFeatureName(fType)
  fmt.Printf("Anomaly: %s\n", name)
}
```

---

## 性能参考

| 操作 | 耗时 |
|------|------|
| 单特征提取 | ~1-10µs |
| 完整向量 (7 特征) | ~50-100µs |
| 配置加载 | ~1ms |
| 兼容层调用 | ~2-5µs |

---

## 常见问题

**Q: 如何调整敏感度？**  
A: 创建自定义 `FeatureConfig` 修改 `EntropyHighThreshold` 等参数

**Q: 配置文件找不到？**  
A: 使用 `extension.NewUnifiedConfigFromEnv()`，并通过 `FINGERPRINT_RULES_FILE` 指定规则文件路径

**Q: 评分总是 0？**  
A: 检查输入数据类型是否与特征类型匹配（如 `FeatureEntropy` 需要 `[]byte`）

**Q: 如何去重指纹？**  
A: 使用 `FeatureVector.Hash`（MD5，用于去重）

**Q: 性能有影响吗？**  
A: 单请求耗时增加 < 100µs，可忽略不计

---

## 相关文档

- 📖 [快速开始指南](../../2-guides/developer/03-feature-extractor-quickstart.md)

---

**版本**: v1.0  
**更新**: 2026-02-28  
**维护**: fingerprint 项目组
