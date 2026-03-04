# fingerprint

[![Go Reference](https://pkg.go.dev/badge/github.com/vistone/fingerprint.svg)](https://pkg.go.dev/github.com/vistone/fingerprint)
[![License](https://img.shields.io/badge/License-BSD_3--Clause-blue.svg)](https://opensource.org/licenses/BSD-3-Clause)
[![Version](https://img.shields.io/badge/version-2.0.0-blue.svg)](https://github.com/vistone/fingerprint/releases/tag/v2.0.0)
[![Test Coverage](https://img.shields.io/badge/coverage-78.6%25-brightgreen.svg)](./TESTING_STANDARDS.md)
[![Go Version](https://img.shields.io/badge/go-1.25.7+-blue.svg)](https://golang.org)

高性能浏览器 TLS 指纹库，提供全面的浏览器指纹识别和模拟能力。

## 核心功能

### 已实现 ✅

- ✅ **71+ 真实浏览器指纹** - Chrome、Firefox、Safari、Opera、Edge 等准确版本
- ✅ **JA3/JA4 指纹生成** - 标准 TLS ClientHello 指纹（MD5 和 SHA256）
- ✅ **JA4S（服务端指纹）** - ServerHello 响应特征识别与异常检测
- ✅ **JA4H（HTTP 请求指纹）** - 应用层请求头特征识别与浏览器匹配
- ✅ **HTTP/2 帧签名** - Settings/Priority/Headers 完整签名与客户端识别
- ✅ **QUIC 签名分析** - QUIC Initial 包指纹、传输参数、帧序列特征识别
- ✅ **HTTP/3 支持** - 基于 QUIC v1/v2 的 HTTP/3 流量识别与异常检测
- ✅ **被动识别** - 从 HTTP 请求头识别浏览器类型、版本、操作系统
- ✅ **主动防护** - Canvas/Audio/WebGL/Font 噪声注入，防止精确指纹追踪
- ✅ **异常检测** - 检测无头浏览器、机器人、可疑指纹和数据熵异常
- ✅ **矛盾检测** - 识别 UA/OS/Platform 等属性间的逻辑不一致
- ✅ **HTTP/2 配置** - 完整的 Settings/Priority/Pseudo-Header-Order 支持
- ✅ **User-Agent 匹配** - 自动生成与指纹匹配的 User-Agent
- ✅ **全球语言支持** - 30+ 种语言的 Accept-Language
- ✅ **操作系统随机化** - 随机选择操作系统
- ✅ **高性能** - 零分配的关键操作，并发安全
- ✅ **配置驱动检测** - 基于 JSON 的可热更新规则配置引擎
- ✅ **Client Hints 支持** - 低熵/高熵提示、Accept-CH 协商、跨源委托、权限策略
- ✅ **ECH 分析** - Encrypted Client Hello 检测、影响评估、替代策略建议
- ✅ **综合风险评分** - 8 维度风险评估、5 级威胁判定、上下文感知计算
- ✅ **UA-CH 协商策略深化** - 完整的 Client Hints 生命周期管理
- ✅ **行为信号分析** - 时序模式、协议分布、连接行为多维分析
- ✅ **WebSocket 指纹检测** - 握手特征分析、异常检测、风险评分

### 规划中 🔄

- 📌 后续优化和扩展功能

## 最新更新（2026.02.28）

🎉 **高级指纹分析能力**

新增三大核心指纹分析模块：

### JA4S - TLS ServerHello 指纹分析

识别服务端 TLS 配置特征，检测异常服务器行为：

```go
import "github.com/vistone/fingerprint"

analyzer := fingerprint.NewJA4SAnalyzer()

// 从 ServerHello 字节数据生成指纹
result, _ := analyzer.AnalyzeServerHelloBytes(serverHelloBytes)

// 从结构化数据生成指纹
result, _ = analyzer.AnalyzeServerHello(fingerprint.ServerHelloData{
    TLSVersion:  0x0304,
    CipherSuite: 0x1302,
    Extensions:  []uint16{0x002b, 0x0033},
    Compression: 0,
})

fmt.Printf("JA4S Hash: %s\n", result.Hash)
fmt.Printf("Risk Score: %.2f\n", result.RiskScore)
// 输出: JA4S Hash: 24f6341540cd29ce904e7896...
//      Risk Score: 0.00
```plaintext

**异常检测能力**：

- 非标准 TLS 版本
- 已知弱密码套件
- 异常扩展配置
- 服务端配置匹配

### HTTP/2 帧签名分析

分析 HTTP/2 帧序列特征，识别客户端实现：

```go
frames := []fingerprint.HTTP2FrameData{
    {
        Type: "SETTINGS",
        Settings: map[string]interface{}{
            "HEADER_TABLE_SIZE":   4096,
            "INITIAL_WINDOW_SIZE": 65535,
        },
    },
    {
        Type:    "HEADERS",
        FrameID: 1,
        Headers: []string{":method", ":scheme", ":authority", ":path"},
    },
}

analyzer := fingerprint.NewHTTP2SignatureAnalyzer()
result, _ := analyzer.AnalyzeHTTP2Stream(frames)

fmt.Printf("Frame Sequence: %s\n", result.FrameSequence)
fmt.Printf("HTTP/2 Hash: %s\n", result.Hash)
// 输出: Frame Sequence: set-hea
//      HTTP/2 Hash: b7494379847d44d3382cdb3b...
```plaintext

**分析维度**：

- Settings 帧参数配置
- Priority 优先级树结构
- Headers 帧头顺序
- Window Update 窗口策略

### JA4H - HTTP 请求头指纹

应用层请求特征识别，浏览器行为匹配：

```go
request := fingerprint.HTTP2RequestData{
    Method:   "GET",
    Path:     "/api/v1/users",
    Protocol: "HTTP/2",
    Headers: []struct {
        Name  string
        Value string
    }{
        {Name: "Host", Value: "example.com"},
        {Name: "User-Agent", Value: "Mozilla/5.0..."},
        {Name: "Accept", Value: "text/html,..."},
        {Name: "Accept-Language", Value: "en-US,en;q=0.9"},
        {Name: "Accept-Encoding", Value: "gzip, deflate"},
    },
}

analyzer := fingerprint.NewJA4HAnalyzer()
result, _ := analyzer.AnalyzeHTTPRequest(request)

fmt.Printf("JA4H Hash: %s\n", result.Hash)
fmt.Printf("Risk Score: %.2f\n", result.RiskScore)
// 输出: JA4H Hash: 51790179c203f52835e0dcca...
//      Risk Score: 0.00
```plaintext

**异常检测能力**：

- 请求头顺序异常
- 缺失常见请求头
- User-Agent 不一致
- SQL 注入迹象
- 方法与路径矛盾

---

### WebSocket 指纹检测 🆕

分析 WebSocket 握手特征，检测可疑连接和自动化工具：

```go
import "github.com/vistone/fingerprint/http/websocket"

// 创建分析器
analyzer := websocket.NewAnalyzer(&websocket.AnalyzerConfig{
    NormalizeHeaders: true,
})

// 分析 WebSocket 升级请求
result := analyzer.AnalyzeRequest(
    []string{"Upgrade", "Connection", "Sec-WebSocket-Key", "Sec-WebSocket-Version"},
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36...",
)

// 获取指纹信息
fmt.Printf("Header Count: %d\n", result.HeaderCount)
fmt.Printf("Normalized: %v\n", result.IsNormalized)

// 异常检测
detector := websocket.NewDetector(&websocket.DetectionConfig{
    RiskThreshold: 60,
})
report := detector.Detect(result)

if report.IsAnomaly {
    fmt.Printf("⚠️  异常类型: %v\n", report.AnomalyTypes)
    fmt.Printf("风险分数: %.2f\n", report.RiskScore)
}
```

**检测能力**：

- 无效握手请求检测
- 可疑请求头分析
- 已知机器人 User-Agent 识别
- 熵值异常检测
- 请求头顺序异常

**测试覆盖**：87.8%

---

### 行为信号分析 - 时序模式与协议分布识别 🆕

通过分析跨请求的行为模式来识别机器人和异常活动：

```go
import "github.com/vistone/fingerprint"

// 创建行为分析器
analyzer := fingerprint.NewBehaviorAnalyzer(nil)

// 添加请求行为数据
for _, req := range requests {
    analyzer.AddRequest(fingerprint.RequestBehavior{
        Timestamp:         req.Timestamp,
        TLSVersion:        "1.3",
        CipherSuite:       "TLS_AES_256_GCM_SHA384",
        HTTPVersion:       "2",
        ReusingConnection: true,
        SNI:               "example.com",
    })
}

// 生成行为信号
signals := analyzer.GenerateBehaviorSignals("example.com")
riskScore := analyzer.GetRiskScore()

fmt.Printf("风险评分: %.2f\n", riskScore)
fmt.Printf("检测信号数: %d\n", len(signals))

// 获取完整分析报告
report := analyzer.GetAnalysisSummary()
```plaintext

**分析维度**：

1. **时序模式分析** - 请求间隔规律性
   - 规律性指数（0-1）：区分机器人(>0.8) vs 真实用户(<0.3)
   - 基于变异系数(CV)的统计分析
   - 异常间隔检测（3-sigma规则）

2. **协议分布分析** - TLS/HTTP版本和密码套件多样性
   - TLS版本分布
   - Cipher Suite组合
   - Extension签名一致性
   - 熵值计算

3. **连接复用行为** - 连接复用率异常检测
   - 真实用户：60-80%复用率
   - 机器人脚本：>95%复用率

**异常检测示例**：

```go
// 机器人特征：高规律性 + 单一协议 + 极高连接复用
pattern := analyzer.AnalyzeTemporalPattern("origin")
proportion := analyzer.AnalyzeProtocolProportion("origin")

if pattern.RegularityIndex > 0.8 && 
   len(proportion.TLSVersions) == 1 && 
   riskScore > 0.7 {
    fmt.Println("⚠️  检测到可疑的机器人行为")
}
```plaintext

**性能指标**：

- 添加请求：1.5μs (0 allocs)
- 时序分析：4.6μs (98% 内存优化)  
- 协议分析：< 50μs
- 信号生成：< 100μs
- 吞吐量：> 1M req/sec

**测试覆盖**：94.9% (78.6% 包平均)

---


分析 QUIC Initial 包特征，识别 HTTP/3 客户端：

```go
initial := fingerprint.QUICInitialData{
    Version: 0x00000001, // QUIC v1
    TransportParams: map[string]interface{}{
        "initial_max_data":                     10485760,
        "initial_max_stream_data_bidi_local":   1048576,
        "initial_max_stream_data_bidi_remote":  1048576,
        "initial_max_streams_bidi":             100,
    },
    FrameTypes: []uint64{
        0x06, // CRYPTO
        0x02, // ACK
    },
    SourceConnectionID:   []byte{0x01, 0x02, 0x03, 0x04},
    InitialMaxData:       10485760,
    InitialMaxStreamData: 1048576,
}

analyzer := fingerprint.NewQUICSignatureAnalyzer()
result, _ := analyzer.AnalyzeQUICInitial(initial)

// 或使用便捷函数
result, _ = fingerprint.ComputeQUICSignature(initial)

fmt.Printf("QUIC Hash: %s\n", result.Hash)
fmt.Printf("Version: %s (HTTP/3: %v)\n", result.QUICVersion, result.IsHTTP3)
fmt.Printf("Risk Score: %.2f\n", result.RiskScore)
// 输出: QUIC Hash: 1300acc37a7e074f1de5286b...
//      Version: v1 (HTTP/3: true)
//      Risk Score: 0.00
```plaintext

**异常检测能力**：

- 草稿版本（draft versions）
- 缺失 CRYPTO 帧
- 可疑的传输参数限制
- 无效连接 ID 长度
- 异常帧序列

---

### ECH - Encrypted Client Hello 分析

分析 TLS 1.3 ECH 扩展，评估对指纹识别的影响：

```go
import "github.com/vistone/fingerprint"

// 构建 ClientHello 数据
data := fingerprint.ClientHelloData{
    TLSVersion:   0x0304, // TLS 1.3
    CipherSuites: []uint16{0x1301, 0x1302, 0x1303},
    Extensions: []fingerprint.ExtensionData{
        {Type: 0xfd00, Data: []byte{0xfe, 0x0d, 0x01, 0x02}}, // ECH outer
        {Type: 0x000d, Data: []byte{}}, // signature_algorithms
        {Type: 0x0033, Data: []byte{}}, // key_share
    },
    HasSNI: false, // ECH 加密了 SNI
}

result, _ := fingerprint.AnalyzeECH(data)

fmt.Printf("ECH 检测: %v\n", result.ECHPresent)
fmt.Printf("ECH 类型: %s\n", result.ECHType)
fmt.Printf("影响等级: %s\n", result.Impact.ImpactLevel)
fmt.Printf("SNI 可见: %v\n", result.Impact.SNIVisible)
fmt.Printf("可见字段签名: %s\n", result.VisibleFieldsSignature)
// 输出: ECH 检测: true
//      ECH 类型: outer
//      影响等级: high
//      SNI 可见: false
//      可见字段签名: faa38b794acbd77a
```plaintext

**分析维度**：

- **ECH 类型识别**：GREASE（兼容测试）、Outer（完整加密）、Inner（内部 Hello）
- **影响评估**：SNI 可见性、受影响的指纹方法、仍可用的替代方法
- **异常检测**：配置错误（ECH 但 SNI 仍可见）、协议异常（旧 TLS 版本使用 ECH）
- **替代策略**：基于可见字段的指纹（Cipher Suites、扩展顺序）、结合 HTTP/2、QUIC 特征

**应对策略示例**：

```go
if result.ECHPresent && result.Impact.ImpactLevel == "high" {
    fmt.Println("建议的替代策略:")
    for _, strategy := range result.AlternativeStrategies {
        fmt.Printf("  - %s\n", strategy)
    }
}
// 输出:
//   - 使用 JA3/JA4 基于可见字段的指纹
//   - 分析 Cipher Suite 和扩展顺序
//   - 结合 HTTP/2 帧签名和 QUIC 特征
//   - 应用层行为分析（请求模式、时序）
//   - IP 信誉和地理位置分析
//   - 实施多层防御策略，不依赖单一指纹方法
```plaintext

---

### 统一特征提取框架（基础架构）

| 新增功能 | 说明 |
| --------- | ------ |
| **统一 Feature Extractor** | 可扩展的特征提取接口，支持多种实现（规则引擎、ML 等） |
| **JSON 规则配置** | `internal/config/rules.json`：阈值、工具标记、规则完全可配置（由 `internal/extension` 统一加载） |
| **统一配置入口** | `extension.NewUnifiedConfigFromEnv()`：自动加载运行配置 + 规则配置 |
| **兼容适配层** | `internal/features/legacy_adapter.go`：现有代码零成本迁移 |
| **特征向量** | `FeatureVector`：完整的特征记录、异常类型、风险评分 |

**使用新特征提取器**（详见[快速指南](docs/2-guides/developer/03-feature-extractor-quickstart.md)）：

```go
import (
    "fingerprint/internal/extension"
    "fingerprint/internal/features"
)

// 使用默认配置
extractor := features.NewBaseFeatureExtractor(nil)
score, isBot := extractor.ExtractFeature(
  features.FeatureHeadlessBrowser,
  userAgent,
  nil,
)

// 或从配置文件加载
appConfig := extension.NewUnifiedConfigFromEnv()
rulesConfig := appConfig.Rules
// ... 转换为 FeatureConfig ...
vector := extractor.ExtractFeatureVector(data, featureConfig)
```plaintext

**下一步** → [缺口分析](docs/1-analysis/02-fingerprint-gap-analysis-2026-02-28.md)

---

## 安装

```bash
go get github.com/vistone/fingerprint
```plaintext

## 快速开始

### 最简单的方式（推荐）⭐

```go
package main

import (
    "log"
    "github.com/vistone/fingerprint"
)

func main() {
    // 一行代码，获取指纹和完整的 HTTP Headers
    result, err := fingerprint.GetRandomFingerprint()
    if err != nil {
        log.Fatal(err)
    }
    
    // result.Profile - TLS 指纹配置
    // result.Headers - 完整的 HTTP Headers（包括 User-Agent、Accept-Language）
    // result.HelloClientID - Client Hello ID
    
    // 使用指纹进行 TLS 握手
    spec, _ := result.Profile.GetClientHelloSpec()
    
    // 使用 Headers
    headers := result.Headers.ToMap()
}
```plaintext

### 指定浏览器类型

```go
// 随机获取 Chrome 指纹
result, _ := fingerprint.GetRandomFingerprintByBrowser("chrome")

// 获取 Edge 指纹
result, _ := fingerprint.GetRandomFingerprintByBrowser("edge")

// 指定浏览器和操作系统
result, _ := fingerprint.GetRandomFingerprintByBrowserWithOS(
    "firefox",
    fingerprint.OSWindows10,
)
```plaintext

### JA3 指纹

```go
// 从指纹名称计算 JA3
ja3, err := fingerprint.ComputeJA3ByProfileName("chrome_133")
fmt.Println(ja3.Hash)       // MD5 哈希，如 "9a79e2a445c2b2c22c4dac65501fa1cd"
fmt.Println(ja3.RawString)  // 原始字符串，如 "772,4865-4866-...,..."

// 从 ClientProfile 计算 JA3
ja3, err = fingerprint.ComputeJA3FromProfile(result.Profile)

// 查找与 JA3 匹配的指纹
matches := fingerprint.FindProfileByJA3("9a79e2a445c2b2c22c4dac65501fa1cd")
```plaintext

### JA4 指纹

```go
// 从指纹名称计算 JA4
ja4, err := fingerprint.ComputeJA4ByProfileName("chrome_133")
fmt.Println(ja4.Hash)      // JA4 哈希，如 "t13i1615h3_8daaf6152771_22334254f9f7"
fmt.Println(ja4.JA4A)      // JA4_a 部分，如 "t13i1615h3"
fmt.Println(ja4.RawString) // JA4_r 原始字符串

// 从 ClientProfile 计算 JA4
ja4, err = fingerprint.ComputeJA4FromProfile(result.Profile)
```plaintext

### 被动识别

```go
// 从 HTTP 请求头识别浏览器
recognizer := fingerprint.NewPassiveRecognizer()
result := recognizer.RecognizeFromHeaders(map[string]string{
    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/133.0.0.0",
    "Accept":     "text/html,...",
    "Sec-CH-UA":  `"Google Chrome";v="133"`,
})
fmt.Println(result.Browser)       // "chrome"
fmt.Println(result.BrowserVersion) // "133"
fmt.Println(result.OS)            // "Windows NT 10.0; Win64; x64"
fmt.Println(result.Confidence)    // 置信度 0.0-1.0

// 仅从 User-Agent 识别
result = fingerprint.RecognizeFromUserAgent("Mozilla/5.0 ...")
```plaintext

### 异常检测

```go
detector := fingerprint.NewAnomalyDetector()

// 检测无头浏览器
isHeadless := detector.DetectHeadlessBrowser("Mozilla/5.0 HeadlessChrome/120")

// 检测数据异常
isAnomalous := detector.DetectAnomalies([]byte("fingerprint data"))
```plaintext

### 矛盾检测

```go
detector := fingerprint.NewContradictionDetector()

hasContradiction := detector.CheckContradictions(map[string]string{
    "os":            "Windows NT 10.0",
    "platform":      "MacIntel",      // 矛盾！Windows OS 但 Mac Platform
    "user_agent":    "Mozilla/5.0 (Windows NT 10.0)...",
    "is_mobile":     "true",
    "screen_width":  "2560",         // 矛盾！移动设备使用超大屏幕
})
```plaintext

### 噪声注入（主动防护）

```go
// 使用默认配置
profile := fingerprint.GenerateBrowserNoiseProfile()

// 自定义配置
config := fingerprint.NoiseConfig{
    Intensity:    0.3,
    EnableCanvas: true,
    EnableAudio:  true,
    EnableWebGL:  true,
    EnableFont:   true,
    EnableScreen: false,
}
injector := fingerprint.NewNoiseInjector(config)

// 生成各类噪声参数
canvasNoise := injector.GenerateCanvasNoise()
// canvasNoise.PixelOffsetR/G/B - 像素偏移

audioNoise := injector.GenerateAudioNoise()
// audioNoise.NoiseLevel - 音频噪声级别

webglNoise := injector.GenerateWebGLNoise()
// webglNoise.MaxTextureSizeOffset - 纹理尺寸偏移
```plaintext

### 自定义 Headers

```go
result, _ := fingerprint.GetRandomFingerprint()

// 设置自定义 header
result.Headers.Set("Cookie", "session_id=abc123")
result.Headers.Set("Authorization", "Bearer token")

// 批量设置
result.Headers.SetHeaders(map[string]string{
    "Cookie":      "session_id=abc123",
    "X-API-Key":   "your-api-key",
})

// 自动合并，直接使用
headers := result.Headers.ToMap()
```plaintext

## 支持的指纹

### 浏览器指纹（71 个）

**Chrome 系列** (19 个)

- Chrome 103, 104, 105, 106, 107, 108, 109, 110, 111, 112
- Chrome 116_PSK, 116_PSK_PQ, 117, 120, 124
- Chrome 130_PSK, 131, 131_PSK, 133, 133_PSK

**Firefox 系列** (12 个)

- Firefox 102, 104, 105, 106, 108, 110, 117, 120, 123, 132, 133, 135

**Safari 系列** (9 个)

- Safari 15.6.1, 16.0, iPad 15.6
- Safari iOS 15.5, 15.6, 16.0, 17.0, 18.0, 18.5

**Opera 系列** (3 个)

- Opera 89, 90, 91

**Edge 系列** (5 个) 🆕

- Edge 99, 101, 120, 131, 133

**移动端和自定义** (23 个)

- Zalando (2), Nike (2), MMS (3), Mesh (4), Confirmed (3)
- OkHttp4 Android (7), Cloudflare (1)

## API 参考

### 核心函数

```go
// 随机指纹（推荐）
GetRandomFingerprint() (*FingerprintResult, error)
GetRandomFingerprintWithOS(os OperatingSystem) (*FingerprintResult, error)
GetRandomFingerprintByBrowser(browserType string) (*FingerprintResult, error)
GetRandomFingerprintByBrowserWithOS(browserType string, os OperatingSystem) (*FingerprintResult, error)

// User-Agent
GetUserAgentByProfileName(profileName string) (string, error)
GetUserAgentByProfileNameWithOS(profileName string, os OperatingSystem) (string, error)
GetUserAgentFromProfile(profile ClientProfile) (string, error)

// Headers
GenerateHeaders(browserType BrowserType, userAgent string, isMobile bool) *HTTPHeaders
RandomLanguage() string
RandomOS() OperatingSystem

// JA3 指纹 🆕
ComputeJA3ByProfileName(profileName string) (*JA3Result, error)
ComputeJA3FromProfile(profile ClientProfile) (*JA3Result, error)
ComputeJA3FromSpec(spec tls.ClientHelloSpec) (*JA3Result, error)
FindProfileByJA3(ja3Hash string) []string
MatchJA3(hash1, hash2 string) bool

// JA4 指纹 🆕
ComputeJA4ByProfileName(profileName string) (*JA4Result, error)
ComputeJA4FromProfile(profile ClientProfile) (*JA4Result, error)
ComputeJA4FromSpec(spec tls.ClientHelloSpec) (*JA4Result, error)

// JA4S 服务端指纹 🆕
ComputeJA4S(data ServerHelloData) (*JA4SResult, error)
ComputeJA4SFromBytes(serverHelloBytes []byte) (*JA4SResult, error)
MatchJA4S(hash1, hash2 string) bool
NewJA4SAnalyzer() *JA4SAnalyzer

// 被动识别 🆕
RecognizeFromUserAgent(userAgent string) *RecognitionResult
NewPassiveRecognizer() *PassiveRecognizer

// 异常检测 🆕
NewAnomalyDetector() *AnomalyDetector

// 矛盾检测 🆕
NewContradictionDetector() *ContradictionDetector

// 噪声注入 🆕
NewNoiseInjector(config NoiseConfig) *NoiseInjector
GenerateBrowserNoiseProfile() *BrowserNoiseProfile
```plaintext

### 数据结构

```go
type FingerprintResult struct {
    Profile       ClientProfile  // TLS 指纹配置
    UserAgent     string         // 对应的 User-Agent
    HelloClientID string         // Client Hello ID
    Headers       *HTTPHeaders   // 标准 HTTP 请求头
}

type JA3Result struct {
    Hash                      string       // JA3 MD5 哈希
    RawString                 string       // JA3 原始字符串
    TLSVersion                uint16       // TLS 版本
    CipherSuites              []uint16     // 密码套件（已过滤GREASE）
    Extensions                []uint16     // 扩展（已过滤GREASE）
    EllipticCurves            []CurveID    // 椭圆曲线
    EllipticCurvePointFormats []uint8      // 椭圆曲线点格式
}

type JA4Result struct {
    Hash      string  // JA4 完整哈希
    RawString string  // JA4_r 原始字符串
    JA4A      string  // JA4_a 部分
    JA4B      string  // JA4_b 部分（密码套件）
    JA4C      string  // JA4_c 部分（扩展+签名算法）
}

type RecognitionResult struct {
    Browser        BrowserType     // 检测到的浏览器类型
    OS             OperatingSystem // 检测到的操作系统
    BrowserVersion string          // 浏览器版本
    Confidence     float64         // 置信度（0.0-1.0）
    IsMobile       bool            // 是否为移动设备
    IsBot          bool            // 是否疑似机器人
}

type NoiseConfig struct {
    Intensity    float64 // 噪声强度（0.0-1.0）
    EnableCanvas bool    // 启用Canvas噪声
    EnableAudio  bool    // 启用Audio噪声
    EnableWebGL  bool    // 启用WebGL噪声
    EnableFont   bool    // 启用Font噪声
    EnableScreen bool    // 启用Screen噪声
}

type BrowserNoiseProfile struct {
    Canvas  *CanvasNoise
    Audio   *AudioNoise
    WebGL   *WebGLNoise
    Font    *FontNoise
    Screen  *ScreenNoise
}
```plaintext

### 操作系统

```go
OSWindows10, OSWindows11           // Windows
OSMacOS13, OSMacOS14, OSMacOS15    // macOS
OSLinux, OSLinuxUbuntu, OSLinuxDebian // Linux
```plaintext

### 浏览器类型

```go
BrowserChrome, BrowserFirefox, BrowserSafari, BrowserOpera, BrowserEdge
```plaintext

## 性能

```text
GetRandomFingerprint:     1374 ns/op    1779 B/op   11 allocs
GetUserAgentByProfileName: 149 ns/op     134 B/op    2 allocs
GenerateHeaders:           244 ns/op     304 B/op    4 allocs
ComputeJA3ByProfileName:  ~500 ns/op
ComputeJA4ByProfileName:  ~600 ns/op
RandomLanguage:             16 ns/op       0 B/op    0 allocs ⭐
RandomOS:                   15 ns/op       0 B/op    0 allocs ⭐

并发性能: 5-6 倍提升
线程安全: 100% 验证通过
```plaintext

## 项目结构

```plaintext
/
├── examples/         # 示例代码
├── internal/utils/   # 内部工具
├── profiles/         # 指纹配置（含Edge新增指纹）
├── test/             # 测试文件
├── types.go          # 类型定义
├── headers.go        # HTTP Headers（含Edge支持）
├── useragent.go      # User-Agent 生成（含Edge模板）
├── useragent_helper.go # User-Agent 辅助函数
├── random.go         # 随机指纹
├── ja3.go            # JA3 指纹计算 🆕
├── ja4.go            # JA4 指纹计算 🆕
├── defense.go        # 异常检测/矛盾检测/被动识别 🆕
├── noise.go          # 噪声注入（主动防护） 🆕
└── README.md
```plaintext

## 开发路线图

### 当前进度：P0 ✅ | P1 ✅ | P2 ✅

| 阶段 | 核心目标 | 预期周期 | 状态 |
| ------ | --------- | --------- | ------ |
| **P0** | 特征提取统一层 + 规则配置化 | 1-2 周 | ✅ 完成 |
| **P1** | JA4S/H + HTTP/2 签名 + QUIC + 综合评分 + UA-CH 协商 | 2-4 周 | ✅ 完成 |
| **P2** | 行为信号分析 + 风险评分体系优化 | 4-8 周 | ✅ 完成 |

**详细计划** → [缺口分析与分期规划](docs/1-analysis/02-fingerprint-gap-analysis-2026-02-28.md)

---

## 更新日志

### v2.1.0-rc1 (P0 阶段 - 2026-02-28) 🆕
- ✅ **统一 Feature Extractor 接口** - 可扩展的特征提取框架
- ✅ **JSON 规则配置引擎** - `internal/config/rules.json` 完全可配置（统一入口为 `internal/extension`）
- ✅ **特征配置加载器** - 支持自动查找和热更新
- ✅ **兼容性适配层** - `LegacyFeatureAdapter` 零成本迁移
- ✅ **特征向量 & 风险评分** - 完整的特征记录与去重
- ✅ **开发文档完善** - [快速指南](docs/2-guides/developer/03-feature-extractor-quickstart.md) + [实施指南](docs/5-process/development/04-p0-feature-extractor-guide.md)

### v2.0.0 (2026-02-28)
- ✅ 新增 Edge 浏览器指纹（5个版本：99/101/120/131/133）
- ✅ 新增 JA3 指纹生成（MD5哈希，支持所有内置指纹）
- ✅ 新增 JA4 指纹生成（SHA256哈希，支持排序和原始顺序版本）
- ✅ 新增被动识别模块（从HTTP请求头识别浏览器/OS）
- ✅ 新增异常检测模块（无头浏览器/机器人检测/数据熵分析）
- ✅ 新增矛盾检测模块（OS/Platform/UA/Screen分辨率一致性验证）
- ✅ 新增噪声注入模块（Canvas/Audio/WebGL/Font噪声参数生成）
- ✅ Edge 浏览器 HTTP Headers 完整支持
- ✅ 指纹总数从66增加到71

### v1.0.2 (2025-12-13)
- ✅ 全面代码重构和优化
- ✅ 创建统一的工具函数包（internal/utils）
- ✅ 优化性能：字符串操作提升 3-5 倍，并发性能提升 5-6 倍
- ✅ 新增完整的集成测试套件（100% 通过率）
- ✅ 14 个基准测试，并发安全验证
- ✅ 简化文档结构

### v1.0.1 (2024)
- 功能增强和 bug 修复

### v1.0.0 (2024)
- 初始版本发布

## 示例

查看 `examples/` 目录获取更多示例：
- `examples/basic/` - 基础使用
- `examples/simple/` - 简单示例
- `examples/random/` - 随机指纹
- `examples/headers/` - Headers 使用
- `examples/useragent/` - User-Agent 生成

## 测试

```bash
# 运行所有测试
go test ./test -v

# 运行基准测试
go test ./test -bench=. -benchmem

# 运行示例
go run examples/random/main.go
```plaintext

## 依赖

- `github.com/bogdanfinn/utls` - TLS 指纹核心库
- `github.com/bogdanfinn/fhttp` - HTTP/2 支持

## 许可证

BSD 3-Clause License。原始代码来自 [bogdanfinn/tls-client](https://github.com/bogdanfinn/tls-client)。

## 相关项目

- [fingerprint-rust](https://github.com/vistone/fingerprint-rust) - Rust 版本高性能指纹库
- [quic](https://github.com/vistone/quic) - QUIC 连接池
- [netconnpool](https://github.com/vistone/netconnpool) - 网络连接池
- [domaindns](https://github.com/vistone/domaindns) - 域名 DNS 解析
- [localippool](https://github.com/vistone/localippool) - 本地 IP 池管理

