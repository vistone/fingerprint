# Fingerprint 架构文档

本文档描述 fingerprint 库的 Go Workspace 架构、模块划分和设计决策。

## 架构概览

```plaintext
┌─────────────────────────────────────────────────────────────────┐
│                     Go Workspace (go.work)                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │              modules/fingerprint (Facade)                 │   │
│  │         统一 API 入口，整合所有子模块功能                  │   │
│  └──────────────────────────────────────────────────────────┘   │
│                              │                                   │
│  ┌─────────────┬─────────────┼─────────────┬─────────────┐      │
│  │             │             │             │             │      │
│  ▼             ▼             ▼             ▼             ▼      │
│  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐  │
│  │ core │  │profiles│  │ tls  │  │ http │  │  ml  │  │defense│  │
│  └──────┘  └──────┘  └──────┘  └──────┘  └──────┘  └──────┘  │
│  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐  │
│  │frontend│  │gateway │  │generator│  │network│  │internal│  │config │  │
│  └──────┘  └──────┘  └──────┘  └──────┘  └──────┘  └──────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Go Workspace 结构

```plaintext
github.com/vistone/fingerprint/
├── go.work                     # Workspace 定义
├── modules/                    # 所有模块
│   ├── fingerprint/            # 主 facade 模块
│   ├── core/                   # 核心类型和接口
│   │   └── types/              # 共享类型定义
│   ├── profiles/               # 指纹配置管理
│   │   ├── builtin.go          # 内置指纹
│   │   ├── chrome.go           # Chrome 指纹
│   │   ├── firefox.go          # Firefox 指纹
│   │   ├── safari.go           # Safari 指纹
│   │   ├── edge.go             # Edge 指纹
│   │   ├── opera.go            # Opera 指纹
│   │   ├── brave.go            # Brave 指纹
│   │   ├── mobile.go           # 移动端指纹
│   │   └── legacy/             # 遗留兼容代码
│   ├── tls/                    # TLS 指纹分析
│   │   ├── ja3.go              # JA3 实现
│   │   └── legacy/             # JA4, JA4S, ECH 等
│   ├── http/                   # HTTP 指纹分析
│   │   └── legacy/             # HTTP/2, JA4H, Client Hints
│   ├── ml/                     # ML 分类器
│   ├── defense/                # 安全防护
│   │   └── legacy/             # 行为分析、风险评估
│   ├── frontend/               # 前端 SDK
│   ├── gateway/                # API 网关
│   ├── generator/              # 指纹生成器
│   ├── network/                # 网络层分析
│   ├── internal/               # 内部工具
│   ├── config/                 # 配置管理
│   └── plugin/                 # 插件系统
├── cmd/                        # 应用程序入口
├── examples/                   # 示例代码
└── test/                       # 集成测试
```

## 模块详解

### 1. Core 模块 (`modules/core`)

**职责**: 提供所有模块共享的基础类型和接口。

```go
// 核心类型
package core

type BrowserType string
type OperatingSystem string
type ClientProfile struct { ... }
type HTTPHeaders struct { ... }
type TLSExtension struct { ... }
```

**零依赖原则**: core 模块不依赖任何其他内部模块。

### 2. Profiles 模块 (`modules/profiles`)

**职责**: 管理 TLS/HTTP 客户端指纹配置。

```plaintext
modules/profiles/
├── profile.go              # ClientProfile 定义和注册表
├── builtin.go              # Chrome 133, Firefox 133, Safari 18
├── chrome.go               # Chrome 115-140 系列
├── firefox.go              # Firefox 115-135 系列
├── safari.go               # Safari 16-18 系列
├── edge.go                 # Edge 115-130 系列
├── opera.go                # Opera 100-110 系列
├── brave.go                # Brave 1.60-1.72 系列
├── mobile.go               # iOS/Android 28个移动指纹
├── comprehensive.go        # 程序化生成配置
├── extended.go             # 扩展配置集合
└── legacy/                 # 遗留兼容代码
    ├── profiles.go
    ├── contributed_browser_profiles.go
    └── ...
```

**导入示例**:
```go
import "github.com/vistone/fingerprint/modules/profiles"

// 获取指纹
profile := profiles.Get("chrome_133")

// 获取所有 Chrome 指纹
chromeProfiles := profiles.GetByBrowser(core.BrowserChrome)
```

### 3. TLS 模块 (`modules/tls`)

**职责**: TLS 指纹分析和 JA3/JA4 计算。

```go
import "github.com/vistone/fingerprint/modules/tls"

// JA3 指纹
ja3 := tls.CalculateJA3(clientHello)

// JA4 指纹 (legacy)
ja4 := tls.CalculateJA4(clientHello)
```

### 4. HTTP 模块 (`modules/http`)

**职责**: HTTP/2 和 HTTP 头部分析。

```go
import "github.com/vistone/fingerprint/modules/http"

// HTTP/2 帧分析
frames := http.ParseHTTP2Frames(data)

// JA4H 指纹 (legacy)
ja4h := http.CalculateJA4H(headers)
```

### 5. ML 模块 (`modules/ml`)

**职责**: 机器学习分类器。

```go
import "github.com/vistone/fingerprint/modules/ml"

// 创建分类器
classifier := ml.NewClassifier(ml.Config{
    Threshold: 0.8,
})

// 训练
classifier.Train(dataset)

// 预测
result := classifier.Predict(features)
```

### 6. Defense 模块 (`modules/defense`)

**职责**: 安全防护和风险评估。

```go
import "github.com/vistone/fingerprint/modules/defense"

// 检测异常
detector := defense.NewDetector()
score := detector.Analyze(fingerprint)
```

### 7. Frontend 模块 (`modules/frontend`)

**职责**: JavaScript SDK 和前端集成。

```go
import "github.com/vistone/fingerprint/modules/frontend"

// 生成前端配置
config := frontend.GenerateSDKConfig(profile)
```

### 8. Gateway 模块 (`modules/gateway`)

**职责**: API 网关和流量管理。

```go
import "github.com/vistone/fingerprint/modules/gateway"

// 创建网关
gw := gateway.New(gateway.Config{
    RateLimit: 1000,
})
```

### 9. Generator 模块 (`modules/generator`)

**职责**: 指纹生成和随机化。

```go
import "github.com/vistone/fingerprint/modules/generator"

// 生成随机指纹
profile := generator.GenerateRandom()
```

### 10. Network 模块 (`modules/network`)

**职责**: 网络层分析（TCP/IP, QUIC）。

```go
import "github.com/vistone/fingerprint/modules/network"

// TCP 指纹
tcpFingerprint := network.AnalyzeTCP(packet)
```

### 11. Internal 模块 (`modules/internal`)

**职责**: 内部工具和共享实现。

```go
import "github.com/vistone/fingerprint/modules/internal/utils"
import "github.com/vistone/fingerprint/modules/internal/metrics"
```

### 12. Config 模块 (`modules/config`)

**职责**: 配置管理和热重载。

```go
import "github.com/vistone/fingerprint/modules/config"

// 加载配置
cfg := config.Load("config.yaml")
```

### 13. Plugin 模块 (`modules/plugin`)

**职责**: 插件系统。

```go
import "github.com/vistone/fingerprint/modules/plugin"

// 注册插件
plugin.Register("custom", myPlugin)
```

## 使用方式

### 方式一: 使用 Facade 模块 (推荐)

```go
import "github.com/vistone/fingerprint/modules/fingerprint"

// 获取随机指纹
profile := fingerprint.GetRandom()

// 获取指定浏览器指纹
chrome := fingerprint.GetByBrowser(fingerprint.BrowserChrome)

// 分析指纹
analyzer := fingerprint.NewAnalyzer()
result := analyzer.Analyze(request)
```

### 方式二: 直接导入子模块

```go
import (
    "github.com/vistone/fingerprint/modules/core"
    "github.com/vistone/fingerprint/modules/profiles"
    "github.com/vistone/fingerprint/modules/tls"
)

profile, _ := profiles.Get("chrome_133")
ja3 := tls.CalculateJA3(clientHello)
```

## 依赖关系

```plaintext
core (零依赖)
    ▲
    │
    ├── profiles ──┬──▶ tls
    │              ├──▶ http
    │              └──▶ internal
    │
    ├── tls ──────▶ internal
    │
    ├── http ─────▶ tls (optional)
    │
    ├── ml ───────▶ core
    │
    ├── defense ──┬──▶ ml
    │             └──▶ internal
    │
    ├── frontend ──▶ ml
    │
    ├── gateway ──┬──▶ defense
    │             ├──▶ frontend
    │             └──▶ ml
    │
    └── fingerprint ──▶ 所有模块
```

## 设计原则

### 1. 模块化设计

- 每个模块有独立的 go.mod
- 模块间通过接口通信
- 支持独立测试和版本控制

### 2. 零依赖核心

```plaintext
modules/core/
├── go.mod          # 零外部依赖
├── types.go        # 基础类型
└── interfaces.go   # 共享接口
```

### 3. 渐进式复杂度

```plaintext
Level 1: core        (基础类型)
Level 2: profiles    (指纹配置)
Level 3: tls/http    (协议分析)
Level 4: ml/defense  (高级功能)
Level 5: gateway     (完整系统)
```

### 4. 向后兼容

遗留代码保留在 `legacy/` 子目录中，逐步迁移。

## 性能特征

| 操作 | 时间复杂度 | 说明 |
|------|-----------|------|
| Profile 选择 | O(1) | 哈希表查找 |
| JA3 解析 | O(n) | n = 字符串长度 |
| HTTP/2 分析 | O(m) | m = 帧数量 |
| 指纹注册 | O(1) | 187+ 指纹预注册 |

## 扩展点

### 自定义 Profile

```go
// 创建自定义指纹
myProfile := profiles.ClientProfile{
    ID: "my_custom",
    BrowserType: core.BrowserChrome,
    // ...
}

// 注册
profiles.Register(myProfile)
```

### 自定义分析器

```go
type CustomAnalyzer struct{}

func (a *CustomAnalyzer) Analyze(data []byte) error {
    // 自定义分析逻辑
    return nil
}
```

## 模块版本策略

- **core**: v1.x.x - 稳定 API
- **profiles**: v0.x.x - 频繁更新指纹
- **ml**: v0.x.x - 实验性功能
- **legacy 代码**: 标记为 Deprecated
