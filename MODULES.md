# Go Workspace 模块架构 (v3.0)

本文档描述指纹库 v3.0 的 Go Workspace 重构架构，直接移植自 Rust 版本的 workspace 设计。

## 架构概览

```
fingerprint/
├── go.work                    # Go Workspace 定义
├── go.mod                     # 主模块依赖
├── fingerprint.go             # 统一 API Facade
├── modules/                   # 工作区模块
│   ├── core/                  # 核心类型和接口
│   ├── profiles/              # 浏览器指纹配置
│   ├── tls/                   # TLS 指纹生成 (JA3/JA4)
│   ├── http/                  # HTTP 指纹生成 (JA4H)
│   ├── ml/                    # ML 分层分类器
│   ├── defense/               # 安全防护系统
│   ├── frontend/              # 前端指纹 SDK
│   └── gateway/               # API 网关服务
└── examples/v3/               # v3 API 示例
```

## 模块详情

### core - 核心模块

**路径**: `github.com/vistone/fingerprint/modules/core`

**功能**:
- 基础类型定义 (BrowserType, OperatingSystem, ProtocolType)
- HTTPHeaders 和操作方法
- 核心接口 (FingerprintSpec, TLSClient)
- 特征类型和特征向量
- 风险等级和评估结果
- 工具函数 (哈希、随机选择、切片操作)

**关键类型**:
```go
type BrowserType string
const (BrowserChrome, BrowserFirefox, BrowserSafari, ...)

type HTTPHeaders struct {
    Accept, AcceptLanguage, AcceptEncoding string
    UserAgent, SecFetchSite, SecFetchMode  string
    Custom map[string]string
}

type FeatureVector struct {
    Features map[FeatureType]float64
    Metadata map[string]interface{}
}
```

### profiles - 指纹配置模块

**路径**: `github.com/vistone/fingerprint/modules/profiles`

**功能**:
- ClientProfile 结构定义
- ProfileRegistry 注册表管理
- 内置浏览器指纹配置 (Chrome, Firefox, Safari)
- HTTP/2 Settings 和 Priorities

**关键类型**:
```go
type ClientProfile struct {
    ID, Name, Description string
    BrowserType, BrowserVersion string
    OS, OSVersion, OSArch string
    TLSVersion uint16
    CipherSuites []uint16
    Extensions []core.TLSExtension
    HTTP2Settings core.HTTP2Settings
    Headers *core.HTTPHeaders
}
```

### tls - TLS 指纹模块

**路径**: `github.com/vistone/fingerprint/modules/tls`

**功能**:
- JA3 指纹计算
- JA4 指纹计算
- ClientHello 分析器
- GREASE 值过滤

**API**:
```go
func CalculateJA3(spec core.ClientHelloSpec) *JA3Result
func CalculateJA4(spec core.ClientHelloSpec) *JA4Result
type Analyzer struct { ... }
func (a *Analyzer) AnalyzeJA3() *JA3Result
func (a *Analyzer) AnalyzeJA4() *JA4Result
```

### http - HTTP 指纹模块

**路径**: `github.com/vistone/fingerprint/modules/http`

**功能**:
- JA4H 指纹计算
- HTTP/2 指纹生成
- HTTP 头分析

**API**:
```go
func CalculateJA4H(headers *core.HTTPHeaders, method string) *JA4HResult
func FingerprintHTTP2(settings core.HTTP2Settings, priorities []core.HTTP2Priority, flow uint32) *HTTP2Fingerprint
```

### ml - 机器学习模块

**路径**: `github.com/vistone/fingerprint/modules/ml`

**功能**:
- 三层分层分类器 (移植自 Rust)
- 特征提取器
- 协议/家族/版本分类
- 训练数据结构

**三层分类器**:
```go
type HierarchicalClassifier struct {
    protocolClassifier *ProtocolClassifier      // Layer 1
    familyClassifiers map[ProtocolType]*FamilyClassifier  // Layer 2
    versionClassifiers map[BrowserType]*VersionClassifier // Layer 3
}
```

**API**:
```go
hc := ml.NewHierarchicalClassifier()
hc.Initialize()
result := hc.Classify(featureVector)
// result: Protocol, Family, Version, Confidence
```

### defense - 安全防护模块

**路径**: `github.com/vistone/fingerprint/modules/defense`

**功能**:
- 被动检测 (无头浏览器、高熵值、自动化工具)
- 主动防护 (噪声注入)
- 风险引擎
- 防护建议

**API**:
```go
detector := defense.NewDetector()
result := detector.Detect(features)

protector := defense.NewActiveProtector()
protection := protector.ApplyProtection(config)

riskEngine := defense.NewRiskEngine()
assessment := riskEngine.Evaluate(features, classification)
```

### frontend - 前端 SDK 模块

**路径**: `github.com/vistone/fingerprint/modules/frontend`

**功能**:
- JavaScript SDK 代码生成
- Canvas/WebGL/Audio/Font 指纹收集
- 会话管理
- HTTP 处理函数

**API**:
```go
sdk := frontend.NewSDK(config)
jsCode := sdk.GenerateJSCore()
injector := sdk.GenerateJSInjector(endpoint)

// HTTP Handler
http.HandleFunc("/collect", sdk.HandleCollect)
```

**生成的 JavaScript 功能**:
- Canvas 指纹（带噪声注入）
- WebGL 指纹（Vendor, Renderer, Extensions）
- Audio 指纹（OscillatorNode 分析）
- 字体检测（基于 Canvas 测量）
- WebRTC IP 泄漏检测
- 硬件信息（Cores, Memory, Touch）

### gateway - API 网关模块

**路径**: `github.com/vistone/fingerprint/modules/gateway`

**功能**:
- REST API 端点
- 限流器 (Token Bucket)
- 指纹缓存 (TTL-based)
- ML 分类集成
- 风险评估
- SDK 脚本服务

**API**:
```go
gw := gateway.NewGateway(config)
response, err := gw.Analyze(ctx, &gateway.AnalyzeRequest{
    TLSVersion:   0x0303,
    CipherSuites: []uint16{...},
    Headers:      headers,
    Frontend:     frontendData,
})

// 启动服务
gw.Start()
```

**端点**:
- `POST /api/v1/analyze` - 指纹分析
- `GET /api/v1/sdk.js` - SDK 脚本
- `POST /api/v1/collect` - 前端数据收集

## 主 Facade API

主包提供统一的便捷 API：

```go
import "github.com/vistone/fingerprint"

// 获取指纹
profile := fingerprint.GetRandom()
chrome := fingerprint.GetByBrowser(fingerprint.BrowserChrome)

// 分析
analyzer := fingerprint.NewAnalyzer()
features := analyzer.ExtractFeatures(profile)
classification := analyzer.Classify(features)
risk := analyzer.EvaluateRisk(features, classification)

// 计算指纹
ja3 := fingerprint.CalculateJA3(spec)
ja4 := fingerprint.CalculateJA4(spec)
ja4h := fingerprint.CalculateJA4H(headers, "GET")

// 快速分析
result := fingerprint.QuickAnalyze(headers, method)

// 启动网关
fingerprint.StartGateway(config)
```

## 使用示例

见 `examples/v3/main.go`:

```bash
cd /media/stone/data/fingerprint
go run ./examples/v3
```

## 编译所有模块

```bash
# 编译单个模块
cd modules/core && go build .
cd modules/profiles && go build .
...

# 编译整个工作区
go build ./...

# 运行示例
go run ./examples/v3
```

## 与 Rust 版本的对比

| 功能 | Rust (fingerprint-rust) | Go (fingerprint-go) |
|------|------------------------|---------------------|
| 架构 | Workspace (20 crates) | Workspace (8 modules) |
| 前端指纹 | Canvas/WebGL/Audio/Font | ✅ 完整支持 |
| ML 分类器 | 3层分层 (95%+ 准确率) | ✅ 已移植 |
| API 网关 | fingerprint-gateway | ✅ 已实现 |
| 协议支持 | HTTP/1, HTTP/2, HTTP/3 | ✅ 同等支持 |
| 指纹格式 | JA3, JA4, JA4H | ✅ 已支持 |
| 云原生 | 一般 | ✅ 更适合 |

## 下一步

- [ ] 训练数据导入 (复用 Rust 的 models/)
- [ ] 更多浏览器指纹配置
- [ ] gRPC API 支持
- [ ] 与 Rust 版本共享指纹数据库
- [ ] 统一文档标准

## 参考

- Rust 版本: https://github.com/vistone/fingerprint-rust
- Go Workspace: https://go.dev/ref/mod#workspaces
