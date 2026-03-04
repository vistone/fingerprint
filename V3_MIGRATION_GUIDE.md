# Fingerprint Go v3.0 Migration Guide

## 概述

v3.0 是基于 Go Workspace 的全新架构，直接移植自 Rust 版本的 workspace 设计。此版本包含 135+ 浏览器指纹配置、三层 ML 分类器、前端 SDK 和 API 网关。

## 架构对比

### v2.x (旧架构)
```
fingerprint/
├── types/          # 类型定义
├── profiles/       # 指纹配置
├── tls/            # TLS 相关
├── http/           # HTTP 相关
└── internal/       # 内部包
```

### v3.0 (新架构 - Go Workspace)
```
fingerprint/
├── go.work                    # Workspace 定义
├── fingerprint.go             # 统一 Facade API
├── modules/                   # 工作区模块
│   ├── core/                  # 核心类型和接口
│   ├── profiles/              # 135+ 浏览器指纹
│   ├── tls/                   # JA3/JA4 生成
│   ├── http/                  # JA4H/HTTP2 指纹
│   ├── ml/                    # ML 三层分类器
│   ├── defense/               # 安全防护系统
│   ├── frontend/              # 前端 SDK
│   └── gateway/               # API 网关
└── examples/v3/               # v3 API 示例
```

## 快速开始

### 1. 获取随机指纹
```go
import "github.com/vistone/fingerprint"

// 获取随机指纹
profile := fingerprint.GetRandom()
fmt.Println(profile.BrowserType)  // "chrome"
fmt.Println(profile.BrowserVersion) // "133.0.6943.98"

// 获取指定浏览器的指纹
chrome := fingerprint.GetByBrowser(fingerprint.BrowserChrome)
firefox := fingerprint.GetByBrowser(fingerprint.BrowserFirefox)
```

### 2. 指纹分析
```go
// 创建分析器
analyzer := fingerprint.NewAnalyzer()

// 提取特征
features := analyzer.ExtractFeatures(profile)

// 分类
classification := analyzer.Classify(features)
fmt.Println(classification.Family)      // "chrome"
fmt.Println(classification.Version)     // "133"
fmt.Println(classification.Confidence)  // 0.95

// 风险评估
risk := analyzer.EvaluateRisk(features, classification)
fmt.Println(risk.Level)  // "low"
```

### 3. 计算指纹哈希
```go
// JA3
ja3 := fingerprint.CalculateJA3(spec)
fmt.Println(ja3.Hash)

// JA4
ja4 := fingerprint.CalculateJA4(spec)
fmt.Println(ja4.Fingerprint)

// JA4H
ja4h := fingerprint.CalculateJA4H(headers, "GET")
fmt.Println(ja4h.Fingerprint)
```

### 4. 启动 API 网关
```go
config := fingerprint.DefaultGatewayConfig
config.Port = 8080

// 启动服务
fingerprint.StartGateway(config)
```

API 端点：
- `POST /api/v1/analyze` - 指纹分析
- `GET /api/v1/sdk.js` - 前端 SDK
- `POST /api/v1/collect` - 数据收集

### 5. 前端 SDK
```go
sdk := fingerprint.NewFrontendSDK(nil)

// 生成 JavaScript 代码
jsCode := sdk.GenerateJSCore()

// HTTP 处理器
http.HandleFunc("/collect", sdk.HandleCollect)
```

## 模块详解

### core 模块
核心类型和接口，所有其他模块都依赖它。

```go
import "github.com/vistone/fingerprint/modules/core"

// 基础类型
type BrowserType string
const (BrowserChrome, BrowserFirefox, BrowserSafari, ...)

// HTTP 头
type HTTPHeaders struct {
    Accept, AcceptLanguage, UserAgent string
    Custom map[string]string
}

// 特征向量
type FeatureVector struct {
    Features map[FeatureType]float64
}
```

### profiles 模块
135+ 浏览器指纹配置。

```go
import "github.com/vistone/fingerprint/modules/profiles"

// 获取所有配置
all := profiles.GetAll()

// 按浏览器类型获取
chromeProfiles := profiles.GetProfilesByBrowser(core.BrowserChrome)

// 获取数量
count := profiles.GetProfileCount()  // 135
```

### ml 模块
三层分层分类器。

```go
import "github.com/vistone/fingerprint/modules/ml"

// 创建分类器
classifier := ml.NewHierarchicalClassifier()
classifier.Initialize()

// 训练（使用合成数据）
ml.InitWithSyntheticData(classifier, 1000)

// 或使用真实数据集
dataset := ml.GenerateSyntheticDataset("mydata", 500)
trainingData := dataset.ToTrainingData()
classifier.Train(trainingData)

// 分类
result := classifier.Classify(features)
```

### gateway 模块
高性能 API 网关。

```go
import "github.com/vistone/fingerprint/modules/gateway"

gw := gateway.NewGateway(&gateway.GatewayConfig{
    Port: 8080,
    RateLimitRequests: 1000,
    CacheEnabled: true,
})

// 分析请求
resp, err := gw.Analyze(ctx, &gateway.AnalyzeRequest{
    TLSVersion: 0x0303,
    CipherSuites: []uint16{0x1301, 0x1302},
    Headers: headers,
})

// 启动服务
gw.Start()
```

## 从 v2.x 迁移

### 类型变更
| v2.x | v3.0 |
|------|------|
| `types.BrowserType` | `fingerprint.BrowserType` 或 `core.BrowserType` |
| `types.HTTPHeaders` | `fingerprint.HTTPHeaders` 或 `core.HTTPHeaders` |
| `profiles.ClientProfile` | `fingerprint.ClientProfile` 或 `profiles.ClientProfile` |

### API 变更
| v2.x | v3.0 |
|------|------|
| `GetRandomFingerprint()` | `fingerprint.GetRandom()` |
| `GetFingerprintByBrowser()` | `fingerprint.GetByBrowser()` |
| `CalculateJA3()` | `fingerprint.CalculateJA3()` |
| `NewBehaviorAnalyzer()` | `fingerprint.NewAnalyzer()` |

### 示例迁移
```go
// v2.x
result := fingerprint.GetRandomFingerprint()
analyzer := fingerprint.NewBehaviorAnalyzer()
signals := analyzer.AnalyzeTemporalPattern(data)

// v3.0
result := fingerprint.GetRandom()
analyzer := fingerprint.NewAnalyzer()
classification := analyzer.Classify(features)
risk := analyzer.EvaluateRisk(features, classification)
```

## 新功能

### 1. ML 分层分类器
- 三层架构：Protocol → Family → Version
- 95%+ 准确率（训练后）
- 支持自定义训练数据

### 2. 扩展指纹库
- 135+ 真实浏览器指纹
- 覆盖 Chrome, Firefox, Safari, Edge, Opera
- 支持 Windows, macOS, Linux, iOS, Android

### 3. 前端 SDK
- Canvas/WebGL/Audio/Font 指纹收集
- 噪声注入（反追踪）
- WebRTC IP 泄漏检测

### 4. API 网关
- REST API 端点
- 限流和缓存
- SDK 脚本服务

### 5. 安全防护
- 被动检测（无头浏览器、自动化工具）
- 主动防护（噪声注入）
- 风险评分系统

## 性能优化

### 编译优化
```bash
# 编译所有模块
go build ./...

# 运行测试
go test ./...

# 编译示例
go run ./examples/v3
```

### 运行时优化
- 零分配特征提取
- 内存池复用
- 并发安全的分类器

## 与 Rust 版本的对比

| 功能 | Rust | Go v3.0 |
|------|------|---------|
| Workspace | 20 crates | 8 modules |
| 指纹数量 | 97+ | 135+ |
| ML 分类器 | 3层 (95%+) | 已移植 |
| 前端指纹 | Canvas/WebGL/Audio/Font | ✅ 完整支持 |
| API 网关 | ✅ | ✅ |
| 协议支持 | HTTP/1/2/3 | ✅ |
| JA3/JA4/JA4H | ✅ | ✅ |
| 云原生 | 一般 | ✅ 更适合 |

## 贡献指南

### 添加新指纹配置
```go
// modules/profiles/your_profiles.go
package profiles

func init() {
    Register(ClientProfile{
        ID: "my_browser_100",
        BrowserType: core.BrowserChrome,
        // ...
    })
}
```

### 添加 ML 特征
```go
// modules/ml/extractor.go
func (fe *FeatureExtractor) ExtractFromProfile(profile *profiles.ClientProfile) *core.FeatureVector {
    fv := core.NewFeatureVector()
    fv.Set(core.FeatureYourFeature, value)
    return fv
}
```

## 下一步

- [ ] 加载真实训练数据（从 Rust 版本导入）
- [ ] gRPC API 支持
- [ ] 分布式分类器
- [ ] 更多浏览器支持

## 参考

- [Rust 版本](https://github.com/vistone/fingerprint-rust)
- [Go Workspace](https://go.dev/ref/mod#workspaces)
- [MODULES.md](./MODULES.md)
