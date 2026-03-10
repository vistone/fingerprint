# fingerprint

[![Go Reference](https://pkg.go.dev/badge/github.com/vistone/fingerprint.svg)](https://pkg.go.dev/github.com/vistone/fingerprint)
[![License](https://img.shields.io/badge/License-BSD_3--Clause-blue.svg)](https://opensource.org/licenses/BSD-3-Clause)
[![Version](https://img.shields.io/badge/version-2.0.0-blue.svg)](https://github.com/vistone/fingerprint/releases/tag/v2.0.0)
[![Go Version](https://img.shields.io/badge/go-1.25.7+-blue.svg)](https://golang.org)

高性能浏览器 TLS 指纹库，提供 150+ 浏览器指纹配置和全面的指纹识别能力。

## 核心特性

- **150+ 真实浏览器指纹** - Chrome、Firefox、Safari、Edge、Opera、Brave 等
- **TLS 指纹分析** - JA3/JA4 指纹生成与识别
- **HTTP/2 签名** - 完整的帧分析和签名匹配
- **机器学习分类** - 内置 ML 检测异常流量
- **安全防护** - 风险评分和异常检测
- **Go Workspace** - 模块化架构，按需导入

## 快速开始

```bash
go get github.com/vistone/fingerprint/modules/fingerprint
```

```go
package main

import (
    "fmt"
    "github.com/vistone/fingerprint/modules/fingerprint"
)

func main() {
    // 获取随机 Chrome 指纹
    profile := fingerprint.GetRandomByBrowser(fingerprint.BrowserChrome)
    fmt.Printf("Selected: %s\n", profile.Name)
}
```

## 模块结构

```
modules/
├── core          # 核心类型 (零依赖)
├── profiles      # 150+ 浏览器指纹
├── tls           # TLS 指纹分析
├── http          # HTTP/2 分析
├── ml            # 机器学习分类器
├── defense       # 安全防护
├── gateway       # API 网关
├── generator     # 指纹生成器
├── fingerprint   # Facade 统一入口
└── ...
```

## 使用方式

### 方式一: Facade 模块 (推荐)

```go
import "github.com/vistone/fingerprint/modules/fingerprint"

profile := fingerprint.GetRandom()
chrome := fingerprint.GetByBrowser(fingerprint.BrowserChrome)
```

### 方式二: 直接导入子模块

```go
import (
    "github.com/vistone/fingerprint/modules/core"
    "github.com/vistone/fingerprint/modules/profiles"
)

profile, _ := profiles.Get("chrome_133")
chromeProfiles := profiles.GetByBrowser(core.BrowserChrome)
```

## 指纹覆盖

| 浏览器 | 版本范围 | 数量 |
|--------|----------|------|
| Chrome | 115-140 | 22+ |
| Firefox | 115-135 | 16+ |
| Safari | 16-18 | 17+ |
| Edge | 115-130 | 13+ |
| Opera | 100-110 | 6+ |
| Brave | 1.60-1.72 | 7+ |
| 移动端 | iOS/Android | 28+ |

## 文档

- [架构说明](./docs/ARCHITECTURE.md) - Go Workspace 架构详解
- [API 文档](./docs/API.md) - 完整 API 参考
- [开发者指南](./docs/DEVELOPER_GUIDE.md) - 开发和贡献指南
  - [⚠️ 版本控制规则](./docs/DEVELOPER_GUIDE.md#-版本控制规则强制执行) - **必读开发规则**
  - [提交前检查清单](./COMMIT_CHECKLIST.md) - 快速参考卡片
- [版本管理策略](./docs/VERSION_MANAGEMENT.md) - 版本号管理详解
- [变更日志](./docs/CHANGELOG.md) - 版本更新记录

## 示例

见 [examples/](./examples/) 目录：

- `basic/` - 基础使用示例
- `advanced/` - 高级功能示例
- `random/` - 随机指纹生成

## 性能

- 指纹选择: O(1) 哈希表查找
- 零分配关键路径
- 并发安全设计

## 许可证

BSD 3-Clause License
