# fingerprint

[![Go Reference](https://pkg.go.dev/badge/github.com/vistone/fingerprint.svg)](https://pkg.go.dev/github.com/vistone/fingerprint)
[![License](https://img.shields.io/badge/License-BSD_3--Clause-blue.svg)](https://opensource.org/licenses/BSD-3-Clause)
[![Version](https://img.shields.io/badge/version-2.0.0-blue.svg)](https://github.com/vistone/fingerprint/releases/tag/v2.0.0)

高性能浏览器 TLS 指纹库，提供全面的浏览器指纹识别和模拟能力。

## 核心功能

- ✅ **71+ 真实浏览器指纹** - Chrome、Firefox、Safari、Opera、**Edge** 等准确版本
- ✅ **JA3 指纹生成** - 从 TLS ClientHello 规范计算标准 JA3 哈希（MD5）
- ✅ **JA4+ 指纹生成** - SHA256 基础的新一代 TLS 指纹（排序版和原始顺序版）
- ✅ **被动识别** - 从 HTTP 请求头识别浏览器类型、版本、操作系统
- ✅ **主动防护** - Canvas/Audio/WebGL/Font 噪声注入，防止精确指纹追踪
- ✅ **异常检测** - 检测无头浏览器、机器人和可疑指纹数据
- ✅ **矛盾检测** - 识别 UA/OS/Platform 等属性间的逻辑不一致
- ✅ **HTTP/2 & HTTP/3** - 完整的 HTTP/2 配置，兼容 HTTP/3
- ✅ **User-Agent 匹配** - 自动生成与指纹匹配的 User-Agent
- ✅ **全球语言支持** - 30+ 种语言的 Accept-Language
- ✅ **操作系统随机化** - 随机选择操作系统
- ✅ **高性能** - 零分配的关键操作，并发安全

## 安装

```bash
go get github.com/vistone/fingerprint
```

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
```

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
```

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
```

### JA4 指纹

```go
// 从指纹名称计算 JA4
ja4, err := fingerprint.ComputeJA4ByProfileName("chrome_133")
fmt.Println(ja4.Hash)      // JA4 哈希，如 "t13i1615h3_8daaf6152771_22334254f9f7"
fmt.Println(ja4.JA4A)      // JA4_a 部分，如 "t13i1615h3"
fmt.Println(ja4.RawString) // JA4_r 原始字符串

// 从 ClientProfile 计算 JA4
ja4, err = fingerprint.ComputeJA4FromProfile(result.Profile)
```

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
```

### 异常检测

```go
detector := fingerprint.NewAnomalyDetector()

// 检测无头浏览器
isHeadless := detector.DetectHeadlessBrowser("Mozilla/5.0 HeadlessChrome/120")

// 检测数据异常
isAnomalous := detector.DetectAnomalies([]byte("fingerprint data"))
```

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
```

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
```

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
```

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
```

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
```

### 操作系统

```go
OSWindows10, OSWindows11           // Windows
OSMacOS13, OSMacOS14, OSMacOS15    // macOS
OSLinux, OSLinuxUbuntu, OSLinuxDebian // Linux
```

### 浏览器类型

```go
BrowserChrome, BrowserFirefox, BrowserSafari, BrowserOpera, BrowserEdge
```

## 性能

```
GetRandomFingerprint:     1374 ns/op    1779 B/op   11 allocs
GetUserAgentByProfileName: 149 ns/op     134 B/op    2 allocs
GenerateHeaders:           244 ns/op     304 B/op    4 allocs
ComputeJA3ByProfileName:  ~500 ns/op
ComputeJA4ByProfileName:  ~600 ns/op
RandomLanguage:             16 ns/op       0 B/op    0 allocs ⭐
RandomOS:                   15 ns/op       0 B/op    0 allocs ⭐

并发性能: 5-6 倍提升
线程安全: 100% 验证通过
```

## 项目结构

```
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
```

## 更新日志

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
```

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

