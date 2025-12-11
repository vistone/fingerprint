# fingerprint

[![Go Reference](https://pkg.go.dev/badge/github.com/vistone/fingerprint.svg)](https://pkg.go.dev/github.com/vistone/fingerprint)
[![License](https://img.shields.io/badge/License-BSD_3--Clause-blue.svg)](https://opensource.org/licenses/BSD-3-Clause)
[![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)](https://github.com/vistone/fingerprint/releases/tag/v1.0.0)

一个独立的浏览器 TLS 指纹库，从 [tls-client](https://github.com/bogdanfinn/tls-client) 迁移而来。

## 简介

`fingerprint` 提供了丰富的浏览器 TLS 指纹配置，支持主流浏览器（Chrome、Firefox、Safari、Opera）以及移动端和自定义配置。这些指纹配置可以用于模拟真实浏览器的 TLS 握手行为，帮助绕过基于 TLS 指纹的反爬虫检测。

## 特性

- ✅ **真实浏览器指纹**：只包含市场上真实存在的浏览器指纹（Chrome、Firefox、Safari、Opera）
- ✅ **主流浏览器支持**：Chrome、Firefox、Safari、Opera 等
- ✅ **移动端支持**：iOS、Android 移动端指纹
- ✅ **HTTP/2 支持**：包含完整的 HTTP/2 配置
- ✅ **User-Agent 匹配**：自动为每个指纹生成匹配的 User-Agent
- ✅ **标准 HTTP Headers**：自动生成完整的标准 HTTP 请求头
- ✅ **全球语言支持**：支持 30+ 种语言的 Accept-Language 头
- ✅ **操作系统随机化**：支持随机选择操作系统，增强真实性
- ✅ **一次调用全部获取**：一次调用同时获得指纹、User-Agent 和 Headers
- ✅ **独立库**：不依赖 tls-client 的其他部分

## 安装

```bash
go get github.com/vistone/fingerprint
```

## 快速开始

### 最简单的方式（推荐）⭐

**只需一行代码，随机获取指纹和对应的 User-Agent：**

```go
package main

import (
    "fmt"
    "log"
    "github.com/vistone/fingerprint"
)

func main() {
    // 随机获取一个指纹和完整的 HTTP Headers
    result, err := fingerprint.GetRandomFingerprint()
    if err != nil {
        log.Fatal(err)
    }
    
    // result.Profile 是 TLS 指纹配置
    // result.Headers 包含完整的 HTTP Headers（包括 User-Agent 和 Accept-Language）
    // result.Name 是指纹名称
    
    fmt.Printf("指纹名称: %s\n", result.Name)
    
    // 使用指纹进行 TLS 握手
    spec, err := result.Profile.GetClientHelloSpec()
    if err != nil {
        log.Fatal(err)
    }
    // 使用 spec 进行 TLS 连接...
    
    // 使用完整的 Headers（包含 User-Agent、Accept-Language 和所有标准 HTTP 头）
    headers := result.Headers.ToMap()
    // 在 HTTP 请求中使用这些 headers...
    // 注意：Headers 中已经包含了 User-Agent 和 Accept-Language，无需单独使用 result.UserAgent
}
```

**就是这么简单！** 一次调用，同时获得：
- ✅ 随机选择的浏览器指纹（只包含真实浏览器）
- ✅ 完整的标准 HTTP Headers（包含 User-Agent、Accept-Language 和所有标准头，操作系统也是随机的）
- ✅ 完整的 TLS 配置

**重要提示**：`result.Headers` 已经包含了 `User-Agent` 和 `Accept-Language`，直接使用 `result.Headers.ToMap()` 即可，无需单独访问 `result.UserAgent`。

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/vistone/fingerprint"
)

func main() {
    // 获取默认指纹（Chrome 133）
    profile := fingerprint.DefaultClientProfile
    
    // 通过名称获取指纹
    profile, ok := fingerprint.MappedTLSClients["chrome_133"]
    if !ok {
        fmt.Println("指纹不存在")
        return
    }
    
    // 获取 Client Hello 规范
    spec, err := profile.GetClientHelloSpec()
    if err != nil {
        fmt.Printf("错误: %v\n", err)
        return
    }
    
    fmt.Printf("指纹: %s\n", profile.GetClientHelloStr())
    fmt.Printf("密码套件数量: %d\n", len(spec.CipherSuites))
}
```

### User-Agent 功能

选择浏览器指纹时，必须使用对应的 User-Agent。操作系统可以随机选择：

```go
package main

import (
    "fmt"
    "github.com/vistone/fingerprint"
)

func main() {
    // 方式 1: 根据 profile 名称获取 User-Agent（操作系统随机）
    ua, err := fingerprint.GetUserAgentByProfileName("chrome_133")
    if err != nil {
        fmt.Printf("错误: %v\n", err)
        return
    }
    fmt.Println("User-Agent:", ua)
    
    // 方式 2: 指定操作系统获取 User-Agent
    ua, err = fingerprint.GetUserAgentByProfileNameWithOS(
        "chrome_133",
        fingerprint.OSWindows10,
    )
    if err != nil {
        fmt.Printf("错误: %v\n", err)
        return
    }
    fmt.Println("User-Agent (Windows):", ua)
    
    // 方式 3: 从 ClientProfile 对象获取 User-Agent
    profile := fingerprint.MappedTLSClients["firefox_135"]
    ua, err = fingerprint.GetUserAgentFromProfile(profile)
    if err != nil {
        fmt.Printf("错误: %v\n", err)
        return
    }
    fmt.Println("User-Agent:", ua)
    
    // 方式 4: 随机选择操作系统
    randomOS := fingerprint.RandomOS()
    ua, err = fingerprint.GetUserAgentByProfileNameWithOS("chrome_120", randomOS)
    fmt.Printf("随机 OS (%s): %s\n", randomOS, ua)
}
```

### 使用特定浏览器指纹

```go
import (
    "github.com/vistone/fingerprint"
    "github.com/vistone/fingerprint/profiles"
)

// 直接使用 profiles 子包中的指纹
chrome133 := profiles.Chrome_133
firefox135 := profiles.Firefox_135
safari16 := profiles.Safari_16_0
```

### 列出所有可用指纹

```go
import "github.com/vistone/fingerprint"

func listAllProfiles() {
    for name, profile := range fingerprint.MappedTLSClients {
        fmt.Printf("%s: %s\n", name, profile.GetClientHelloStr())
    }
}
```

## 支持的指纹

### Chrome 系列
- Chrome 103, 104, 105, 106, 107, 108, 109, 110, 111, 112
- Chrome 116_PSK, 116_PSK_PQ
- Chrome 117, 120, 124
- Chrome 130_PSK, 131, 131_PSK
- Chrome 133, 133_PSK

### Firefox 系列
- Firefox 102, 104, 105, 106, 108, 110, 117, 120, 123, 132, 133, 135

### Safari 系列
- Safari 15.6.1, 16.0
- Safari iPad 15.6
- Safari iOS 15.5, 15.6, 16.0, 17.0, 18.0, 18.5

### Opera 系列
- Opera 89, 90, 91

**注意**：当前 `MappedTLSClients` 中包含 44 个主流浏览器指纹。代码中还定义了其他移动端和自定义指纹（如 Zalando、Nike、MMS、Mesh、Confirmed、OkHttp4、Cloudflare Custom），但这些指纹需要直接从 `profiles` 子包中导入使用，未包含在 `MappedTLSClients` 映射表中。

## API 参考

### ClientProfile

`ClientProfile` 表示一个客户端 TLS 指纹配置，包含以下方法：

- `GetClientHelloSpec() (tls.ClientHelloSpec, error)` - 获取 TLS Client Hello 规范
- `GetClientHelloStr() string` - 获取 Client Hello 字符串标识
- `GetSettings() map[http2.SettingID]uint32` - 获取 HTTP/2 设置
- `GetSettingsOrder() []http2.SettingID` - 获取 HTTP/2 设置顺序
- `GetConnectionFlow() uint32` - 获取连接流 ID
- `GetPseudoHeaderOrder() []string` - 获取 HTTP/2 伪头部顺序
- `GetHeaderPriority() *http2.PriorityParam` - 获取 HTTP/2 头部优先级参数
- `GetClientHelloId() tls.ClientHelloID` - 获取 Client Hello ID
- `GetPriorities() []http2.Priority` - 获取 HTTP/2 优先级列表

### 随机指纹 API（推荐）

- `GetRandomFingerprint() (*FingerprintResult, error)` - 随机获取一个指纹和对应的 User-Agent（操作系统随机）
- `GetRandomFingerprintWithOS(os OperatingSystem) (*FingerprintResult, error)` - 随机获取一个指纹和对应的 User-Agent，并指定操作系统
- `GetRandomFingerprintByBrowser(browserType string) (*FingerprintResult, error)` - 根据浏览器类型随机获取指纹和 User-Agent
- `GetRandomFingerprintByBrowserWithOS(browserType string, os OperatingSystem) (*FingerprintResult, error)` - 根据浏览器类型随机获取指纹和 User-Agent，并指定操作系统

`FingerprintResult` 结构：
```go
type FingerprintResult struct {
    Profile   ClientProfile  // TLS 指纹配置
    UserAgent string         // 对应的 User-Agent（已包含在 Headers 中，建议直接使用 Headers）
    Name      string         // 指纹名称
    Headers   *HTTPHeaders   // 标准 HTTP 请求头（包含 User-Agent、Accept-Language 和所有标准头）
}
```

**注意**：`Headers` 已经包含了 `User-Agent` 和 `Accept-Language`，建议直接使用 `result.Headers.ToMap()` 获取所有 HTTP 头，而不是单独访问 `result.UserAgent`。

### HTTP Headers API

`HTTPHeaders` 包含完整的标准浏览器请求头：
- `Accept` - 内容类型接受
- `Accept-Language` - 语言偏好（支持 30+ 种全球语言，随机选择）
- `Accept-Encoding` - 编码支持
- `User-Agent` - 用户代理（已自动匹配指纹）
- `Sec-Fetch-*` - 安全获取头（Chrome/Safari/Opera）
- `Sec-CH-UA-*` - 客户端提示头（Chrome/Opera）
- `Upgrade-Insecure-Requests` - 升级不安全请求

**方法：**
- `ToMap() map[string]string` - 将 HTTPHeaders 转换为 map，直接用于 HTTP 请求

**函数：**
- `GenerateHeaders(browserType BrowserType, userAgent string, isMobile bool) *HTTPHeaders` - 根据浏览器类型和 User-Agent 生成标准 HTTP headers
- `RandomLanguage() string` - 随机选择一个语言（从 30+ 种全球语言中选择）

**重要**：`Headers` 已经包含了 `User-Agent` 和 `Accept-Language`，无需单独使用 `FingerprintResult.UserAgent` 字段。

### User-Agent API

**主要函数（推荐使用）：**
- `GetUserAgentByProfileName(profileName string) (string, error)` - 根据 profile 名称获取 User-Agent（操作系统随机）
- `GetUserAgentByProfileNameWithOS(profileName string, os OperatingSystem) (string, error)` - 根据 profile 名称和指定操作系统获取 User-Agent
- `GetUserAgentFromProfile(profile ClientProfile) (string, error)` - 从 ClientProfile 对象获取 User-Agent
- `GetUserAgentFromProfileWithOS(profile ClientProfile, os OperatingSystem) (string, error)` - 从 ClientProfile 对象获取 User-Agent，并指定操作系统
- `RandomOS() OperatingSystem` - 随机选择一个操作系统

**辅助函数：**
- `GetUserAgentForProfile(profileName string) (string, error)` - 为指定的 ClientProfile 获取 User-Agent（内部使用）
- `GetUserAgentForProfileWithOS(profileName string, os OperatingSystem) (string, error)` - 为指定的 ClientProfile 和操作系统获取 User-Agent（内部使用）
- `GetUserAgentForMappedProfile(profileName string) (string, error)` - 从 MappedTLSClients 中获取指定名称的 profile 的 User-Agent
- `GetUserAgentForMappedProfileWithOS(profileName string, os OperatingSystem) (string, error)` - 从 MappedTLSClients 中获取指定名称的 profile 的 User-Agent，并指定操作系统

### 支持的操作系统

- `OSWindows10` - Windows 10
- `OSWindows11` - Windows 11
- `OSMacOS13` - macOS 13
- `OSMacOS14` - macOS 14
- `OSMacOS15` - macOS 15
- `OSLinux` - Linux
- `OSLinuxUbuntu` - Ubuntu Linux
- `OSLinuxDebian` - Debian Linux

### 全局变量和常量

**变量：**
- `DefaultClientProfile` - 默认客户端指纹配置（Chrome 133）
- `MappedTLSClients` - 所有可用的 TLS 客户端指纹映射表（map[string]ClientProfile）
- `Languages` - 全球语言列表（30+ 种语言，用于随机选择 Accept-Language）
- `OperatingSystems` - 操作系统列表（用于随机选择）
- `NewClientProfile` - 创建新的客户端指纹配置的函数

**类型：**
- `BrowserType` - 浏览器类型（chrome, firefox, safari, opera, edge）
- `OperatingSystem` - 操作系统类型（字符串类型）

**BrowserType 常量：**
- `BrowserChrome` - Chrome 浏览器
- `BrowserFirefox` - Firefox 浏览器
- `BrowserSafari` - Safari 浏览器
- `BrowserOpera` - Opera 浏览器
- `BrowserEdge` - Edge 浏览器

## 依赖

- `github.com/bogdanfinn/utls` - TLS 指纹核心库
- `github.com/bogdanfinn/fhttp` - HTTP/2 支持

## 许可证

本项目基于 BSD 3-Clause 许可证。原始代码来自 [bogdanfinn/tls-client](https://github.com/bogdanfinn/tls-client)。

## 致谢

本库的指纹配置来源于 [bogdanfinn/tls-client](https://github.com/bogdanfinn/tls-client) 项目，感谢原作者的贡献。

## 相关项目

- [quic](https://github.com/vistone/quic) - QUIC 连接池
- [netconnpool](https://github.com/vistone/netconnpool) - 网络连接池
- [domaindns](https://github.com/vistone/domaindns) - 域名 DNS 解析
- [localippool](https://github.com/vistone/localippool) - 本地 IP 池管理

