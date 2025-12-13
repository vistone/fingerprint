# API 参考文档

## 核心类型

### ClientProfile

`ClientProfile` 表示一个客户端 TLS 指纹配置。

#### 方法

- `GetClientHelloSpec() (tls.ClientHelloSpec, error)` - 获取 TLS Client Hello 规范
- `GetClientHelloStr() string` - 获取 Client Hello 字符串标识
- `GetSettings() map[http2.SettingID]uint32` - 获取 HTTP/2 设置
- `GetSettingsOrder() []http2.SettingID` - 获取 HTTP/2 设置顺序
- `GetConnectionFlow() uint32` - 获取连接流 ID
- `GetPseudoHeaderOrder() []string` - 获取 HTTP/2 伪头部顺序
- `GetHeaderPriority() *http2.PriorityParam` - 获取 HTTP/2 头部优先级参数
- `GetClientHelloId() tls.ClientHelloID` - 获取 Client Hello ID
- `GetPriorities() []http2.Priority` - 获取 HTTP/2 优先级列表

### HTTPHeaders

`HTTPHeaders` 包含完整的标准浏览器请求头。

#### 字段

- `Accept` - 内容类型接受
- `AcceptLanguage` - 语言偏好（支持 30+ 种全球语言）
- `AcceptEncoding` - 编码支持
- `UserAgent` - 用户代理
- `SecFetchSite` - Sec-Fetch-Site 头
- `SecFetchMode` - Sec-Fetch-Mode 头
- `SecFetchUser` - Sec-Fetch-User 头
- `SecFetchDest` - Sec-Fetch-Dest 头
- `SecCHUA` - Sec-CH-UA 头
- `SecCHUAMobile` - Sec-CH-UA-Mobile 头
- `SecCHUAPlatform` - Sec-CH-UA-Platform 头
- `UpgradeInsecureRequests` - Upgrade-Insecure-Requests 头
- `Custom` - 用户自定义的 headers

#### 方法

- `ToMap() map[string]string` - 转换为 map，自动合并 Custom headers
- `Set(key, value string)` - 设置单个自定义 header
- `SetHeaders(customHeaders map[string]string)` - 批量设置自定义 headers
- `Clone() *HTTPHeaders` - 克隆对象
- `Merge(customHeaders map[string]string) *HTTPHeaders` - 合并自定义 headers

### FingerprintResult

`FingerprintResult` 包含指纹、User-Agent 和 HTTP Headers。

#### 字段

- `Profile ClientProfile` - TLS 指纹配置
- `UserAgent string` - 对应的 User-Agent
- `HelloClientID string` - Client Hello ID
- `Headers *HTTPHeaders` - 标准 HTTP 请求头

## 随机指纹 API

### GetRandomFingerprint

```go
func GetRandomFingerprint() (*FingerprintResult, error)
```

随机获取一个指纹和对应的 User-Agent（操作系统随机）。

### GetRandomFingerprintWithOS

```go
func GetRandomFingerprintWithOS(os OperatingSystem) (*FingerprintResult, error)
```

随机获取一个指纹，并指定操作系统。

### GetRandomFingerprintByBrowser

```go
func GetRandomFingerprintByBrowser(browserType string) (*FingerprintResult, error)
```

根据浏览器类型随机获取指纹。

### GetRandomFingerprintByBrowserWithOS

```go
func GetRandomFingerprintByBrowserWithOS(browserType string, os OperatingSystem) (*FingerprintResult, error)
```

根据浏览器类型和操作系统随机获取指纹。

## User-Agent API

### GetUserAgentByProfileName

```go
func GetUserAgentByProfileName(profileName string) (string, error)
```

根据 profile 名称获取 User-Agent（操作系统随机）。

### GetUserAgentByProfileNameWithOS

```go
func GetUserAgentByProfileNameWithOS(profileName string, os OperatingSystem) (string, error)
```

根据 profile 名称和指定操作系统获取 User-Agent。

### GetUserAgentFromProfile

```go
func GetUserAgentFromProfile(profile ClientProfile) (string, error)
```

从 ClientProfile 对象获取 User-Agent。

### GetUserAgentFromProfileWithOS

```go
func GetUserAgentFromProfileWithOS(profile ClientProfile, os OperatingSystem) (string, error)
```

从 ClientProfile 对象获取 User-Agent，并指定操作系统。

### RandomOS

```go
func RandomOS() OperatingSystem
```

随机选择一个操作系统。

## HTTP Headers API

### GenerateHeaders

```go
func GenerateHeaders(browserType BrowserType, userAgent string, isMobile bool) *HTTPHeaders
```

根据浏览器类型和 User-Agent 生成标准 HTTP headers。

### RandomLanguage

```go
func RandomLanguage() string
```

随机选择一个语言（从 30+ 种全球语言中选择）。

## 常量

### BrowserType

- `BrowserChrome` - Chrome 浏览器
- `BrowserFirefox` - Firefox 浏览器
- `BrowserSafari` - Safari 浏览器
- `BrowserOpera` - Opera 浏览器
- `BrowserEdge` - Edge 浏览器

### OperatingSystem

- `OSWindows10` - Windows 10
- `OSWindows11` - Windows 11
- `OSMacOS13` - macOS 13
- `OSMacOS14` - macOS 14
- `OSMacOS15` - macOS 15
- `OSLinux` - Linux
- `OSLinuxUbuntu` - Ubuntu Linux
- `OSLinuxDebian` - Debian Linux

## 全局变量

- `DefaultClientProfile` - 默认客户端指纹配置（Chrome 133）
- `MappedTLSClients` - 所有可用的 TLS 客户端指纹映射表
- `Languages` - 全球语言列表（30+ 种语言）
- `OperatingSystems` - 操作系统列表
