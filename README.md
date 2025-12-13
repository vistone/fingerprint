# fingerprint

[![Go Reference](https://pkg.go.dev/badge/github.com/vistone/fingerprint.svg)](https://pkg.go.dev/github.com/vistone/fingerprint)
[![License](https://img.shields.io/badge/License-BSD_3--Clause-blue.svg)](https://opensource.org/licenses/BSD-3-Clause)
[![Version](https://img.shields.io/badge/version-1.0.2-blue.svg)](https://github.com/vistone/fingerprint/releases/tag/v1.0.2)

一个独立的浏览器 TLS 指纹库，从 [tls-client](https://github.com/bogdanfinn/tls-client) 迁移而来。

## 特性

- ✅ **真实浏览器指纹**：66 个真实浏览器指纹（Chrome、Firefox、Safari、Opera）
- ✅ **移动端支持**：iOS、Android 移动端指纹
- ✅ **HTTP/2 & HTTP/3**：完整的 HTTP/2 配置，兼容 HTTP/3
- ✅ **User-Agent 匹配**：自动生成匹配的 User-Agent
- ✅ **标准 HTTP Headers**：完整的标准 HTTP 请求头
- ✅ **全球语言支持**：30+ 种语言的 Accept-Language
- ✅ **操作系统随机化**：随机选择操作系统
- ✅ **高性能**：零分配的关键操作，并发安全
- ✅ **独立库**：不依赖 tls-client 的其他部分

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

// 指定浏览器和操作系统
result, _ := fingerprint.GetRandomFingerprintByBrowserWithOS(
    "firefox",
    fingerprint.OSWindows10,
)
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

### 浏览器指纹（66 个）

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
```

### 数据结构

```go
type FingerprintResult struct {
    Profile       ClientProfile  // TLS 指纹配置
    UserAgent     string         // 对应的 User-Agent
    HelloClientID string         // Client Hello ID
    Headers       *HTTPHeaders   // 标准 HTTP 请求头
}

type HTTPHeaders struct {
    Accept, AcceptLanguage, AcceptEncoding string
    UserAgent string
    SecFetchSite, SecFetchMode, SecFetchUser, SecFetchDest string
    SecCHUA, SecCHUAMobile, SecCHUAPlatform string
    UpgradeInsecureRequests string
    Custom map[string]string  // 自定义 headers
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
RandomLanguage:             16 ns/op       0 B/op    0 allocs ⭐
RandomOS:                   15 ns/op       0 B/op    0 allocs ⭐

并发性能: 5-6 倍提升
线程安全: 100% 验证通过
```

## 项目结构

```
/workspace/
├── bin/              # 编译输出
├── examples/         # 示例代码
├── internal/utils/   # 内部工具
├── profiles/         # 指纹配置
├── test/            # 测试文件
├── types.go         # 类型定义
├── headers.go       # HTTP Headers
├── useragent.go     # User-Agent 生成
├── random.go        # 随机指纹
└── README.md
```

## 更新日志

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

- [quic](https://github.com/vistone/quic) - QUIC 连接池
- [netconnpool](https://github.com/vistone/netconnpool) - 网络连接池
- [domaindns](https://github.com/vistone/domaindns) - 域名 DNS 解析
- [localippool](https://github.com/vistone/localippool) - 本地 IP 池管理
