# API 文档

本文档描述 fingerprint 库的公共 API 使用方式（Go Workspace 版本）。

## 快速开始

### 安装

```bash
go get github.com/vistone/fingerprint/modules/fingerprint
```

### 基本使用

```go
package main

import (
    "github.com/vistone/fingerprint/modules/profiles"
    "github.com/vistone/fingerprint/modules/tls"
)

func main() {
    // 获取指纹
    profile, _ := profiles.Get("chrome_133")
    
    // 使用 TLS 模块
    ja3 := tls.CalculateJA3(clientHello)
}
```

## 导入路径对照表

| 旧路径 (废弃) | 新路径 |
|--------------|--------|
| `github.com/vistone/fingerprint` | `github.com/vistone/fingerprint/modules/fingerprint` |
| `github.com/vistone/fingerprint/profiles` | `github.com/vistone/fingerprint/modules/profiles` |
| `github.com/vistone/fingerprint/tls/ja3` | `github.com/vistone/fingerprint/modules/tls` |
| `github.com/vistone/fingerprint/http/ja4h` | `github.com/vistone/fingerprint/modules/http/legacy/ja4h` |
| `github.com/vistone/fingerprint/internal/config` | `github.com/vistone/fingerprint/modules/config` |
| `github.com/vistone/fingerprint/internal/tcpip` | `github.com/vistone/fingerprint/modules/internal/tcpip` |
| `github.com/vistone/fingerprint/types` | `github.com/vistone/fingerprint/modules/core/types` |

## Core 模块 API

### 基础类型

```go
import "github.com/vistone/fingerprint/modules/core"

// 浏览器类型
const (
    BrowserChrome  = core.BrowserChrome
    BrowserFirefox = core.BrowserFirefox
    BrowserSafari  = core.BrowserSafari
    BrowserEdge    = core.BrowserEdge
    BrowserOpera   = core.BrowserOpera
    BrowserBrave   = core.BrowserBrave
)

// 操作系统类型
const (
    OSWindows10   = core.OSWindows10
    OSWindows11   = core.OSWindows11
    OSMacOS13     = core.OSMacOS13
    OSMacOS14     = core.OSMacOS14
    OSMacOS15     = core.OSMacOS15
    OSLinux       = core.OSLinux
    OSiOS         = core.OSiOS
    OSAndroid     = core.OSAndroid
)
```

## Profiles 模块 API

### 获取指纹

```go
import "github.com/vistone/fingerprint/modules/profiles"

// 通过 ID 获取
profile, ok := profiles.Get("chrome_133")
if !ok {
    log.Fatal("profile not found")
}

// 获取随机指纹
profile := profiles.GetRandom()

// 按浏览器类型获取
chromeProfiles := profiles.GetByBrowser(core.BrowserChrome)

// 获取所有指纹
allProfiles := profiles.GetAll()
```

### 预定义指纹变量

```go
// Chrome 系列
profiles.Chrome115, profiles.Chrome116, ... profiles.Chrome140
profiles.Chrome131, profiles.Chrome133  // 内置

// Firefox 系列  
profiles.Firefox115, ... profiles.Firefox135

// Safari 系列
profiles.Safari16_0, ... profiles.Safari18_2

// Edge 系列
profiles.Edge115, ... profiles.Edge130

// Opera 系列
profiles.Opera100, ... profiles.Opera110

// Brave 系列
profiles.Brave1_60, ... profiles.Brave1_72

// 移动端
profiles.IOSSafari16, profiles.IOSSafari17, profiles.IOSSafari18
profiles.AndroidChrome115, ... profiles.AndroidChrome131
profiles.AndroidFirefox115, ... profiles.AndroidFirefox130
```

### 注册自定义指纹

```go
myProfile := profiles.ClientProfile{
    ID:             "my_custom",
    Name:           "My Custom Profile",
    BrowserType:    core.BrowserChrome,
    BrowserVersion: "120.0",
    OS:             core.OSWindows11,
    OSVersion:      "10.0.22631",
    TLSVersion:     0x0303,
    CipherSuites:   []uint16{0x1301, 0x1302, 0x1303},
    Headers: &core.HTTPHeaders{
        Accept:         "text/html,*/*",
        AcceptLanguage: "en-US,en;q=0.9",
    },
}

profiles.Register(myProfile)
```

## Fingerprint Facade API

### 统一入口

```go
import "github.com/vistone/fingerprint/modules/fingerprint"

// 获取随机指纹
profile := fingerprint.GetRandom()

// 按浏览器获取
chrome := fingerprint.GetRandomByBrowser(fingerprint.BrowserChrome)

// 分析请求
analyzer := fingerprint.NewAnalyzer()
result := analyzer.Analyze(request)
```

## TLS 模块 API

### JA3 指纹

```go
import "github.com/vistone/fingerprint/modules/tls"

// 计算 JA3
ja3 := tls.CalculateJA3(clientHello)

// 计算 JA3 (legacy)
ja3, err := tls.CalculateJA3Legacy(clientHello)
```

### JA4/JA4S (Legacy)

```go
import "github.com/vistone/fingerprint/modules/tls/legacy/ja4"
import "github.com/vistone/fingerprint/modules/tls/legacy/ja4s"

// JA4
fp, err := ja4.CalculateJA4(clientHello)

// JA4S
result, err := ja4s.ComputeJA4S(serverHelloData)
```

## HTTP 模块 API

### JA4H (Legacy)

```go
import "github.com/vistone/fingerprint/modules/http/legacy/ja4h"

// 计算 JA4H
fp, err := ja4h.CalculateJA4H(req)
```

### Client Hints (Legacy)

```go
import "github.com/vistone/fingerprint/modules/http/legacy/clienthints"

// 解析 Client Hints
hints, err := clienthints.ParseClientHints(headers)
```

## ML 模块 API

### 分类器

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

## Defense 模块 API

### 异常检测

```go
import "github.com/vistone/fingerprint/modules/defense"

// 创建检测器
detector := defense.NewDetector()

// 分析
score := detector.Analyze(fingerprint)
```

## Gateway 模块 API

### API 网关

```go
import "github.com/vistone/fingerprint/modules/gateway"

// 创建网关
gw := gateway.New(gateway.Config{
    RateLimit: 1000,
})

// 启动
gw.Start(":8080")
```

## Generator 模块 API

### 指纹生成

```go
import "github.com/vistone/fingerprint/modules/generator"

// 生成随机指纹
profile := generator.GenerateRandom()

// 生成特定浏览器指纹
profile := generator.GenerateForBrowser(core.BrowserChrome)
```

## Network 模块 API

### TCP/IP 分析

```go
import "github.com/vistone/fingerprint/modules/network"

// 分析 TCP
behavior := network.AnalyzeTCP(packet)

// 分析 QUIC
quicInfo := network.AnalyzeQUIC(packet)
```

## Internal 模块 API

### 工具函数

```go
import "github.com/vistone/fingerprint/modules/internal/utils"

// 随机选择
choice := utils.RandomChoice(items)
```

### 指标

```go
import "github.com/vistone/fingerprint/modules/internal/metrics"

// 记录指标
metrics.RecordFingerprintGeneration("Chrome", "Windows", nil)
```

## Config 模块 API

### 配置管理

```go
import "github.com/vistone/fingerprint/modules/config"

// 加载配置
cfg := config.Load("config.yaml")

// 监听变更
ch := make(chan config.Change)
config.Watch(ch)
```

## 完整示例

### 使用 Facade 模块

```go
package main

import (
    "log"
    "github.com/vistone/fingerprint/modules/fingerprint"
)

func main() {
    // 获取随机 Chrome 指纹
    profile := fingerprint.GetRandomByBrowser(fingerprint.BrowserChrome)
    log.Printf("Selected: %s", profile.Name)
}
```

### 直接使用子模块

```go
package main

import (
    "log"
    "github.com/vistone/fingerprint/modules/core"
    "github.com/vistone/fingerprint/modules/profiles"
    "github.com/vistone/fingerprint/modules/tls"
)

func main() {
    // 获取指纹
    profile, ok := profiles.Get("chrome_133")
    if !ok {
        log.Fatal("not found")
    }
    
    // 计算 JA3
    ja3 := tls.CalculateJA3(clientHello)
    log.Printf("JA3: %s", ja3)
}
```

## 错误处理

```go
import (
    "errors"
    "github.com/vistone/fingerprint/modules/profiles"
)

profile, err := profiles.Get("unknown")
if err != nil {
    if errors.Is(err, profiles.ErrNotFound) {
        // 使用默认指纹
        profile = profiles.Chrome133
    }
}
```

## 性能优化

### Profile 缓存

```go
var cache = sync.Map{}

func GetCached(id string) (profiles.ClientProfile, bool) {
    if v, ok := cache.Load(id); ok {
        return v.(profiles.ClientProfile), true
    }
    
    profile, ok := profiles.Get(id)
    if ok {
        cache.Store(id, profile)
    }
    return profile, ok
}
```

## 版本兼容性

- **v0.x.x**: 当前版本，Go Workspace 架构
- **legacy 包**: 向后兼容的旧代码
- **核心 API**: 稳定，不会破坏性变更
