# Fingerprint Go v3.0

[![Go Reference](https://pkg.go.dev/badge/github.com/vistone/fingerprint.svg)](https://pkg.go.dev/github.com/vistone/fingerprint)
[![License](https://img.shields.io/badge/License-BSD_3--Clause-blue.svg)](https://opensource.org/licenses/BSD-3-Clause)
[![Go Version](https://img.shields.io/badge/go-1.25.7+-blue.svg)](https://golang.org)
[![Tests](https://github.com/vistone/fingerprint/actions/workflows/ci.yml/badge.svg)](https://github.com/vistone/fingerprint/actions)

高性能浏览器指纹库 v3.0 - Go Workspace 架构重构版，移植自 Rust 版本。

## 🎯 核心特性

### v3.0 新功能

- ✅ **Go Workspace 架构** - 8 个独立模块，可单独导入
- ✅ **135+ 浏览器指纹** - 覆盖 Chrome, Firefox, Safari, Edge, Opera
- ✅ **ML 三层分类器** - Protocol → Family → Version (95%+ 准确率)
- ✅ **前端指纹 SDK** - Canvas/WebGL/Audio/Font 收集 + 噪声注入
- ✅ **API 网关** - HTTP/REST + gRPC 双协议支持
- ✅ **Rust 数据兼容** - 双向导入/导出训练数据和模型
- ✅ **安全防护系统** - 被动检测 + 主动防护 + 风险评分
- ✅ **完整测试覆盖** - 单元测试 + 基准测试 + CI/CD

### 协议支持

- ✅ **TLS 指纹** - JA3, JA4 完整实现
- ✅ **HTTP 指纹** - JA4H, HTTP/2 Settings/Priority
- ✅ **HTTP/3** - QUIC 初始包指纹

## 📦 模块结构

```
github.com/vistone/fingerprint
├── modules/core       # 核心类型和接口
├── modules/profiles   # 135+ 浏览器指纹配置
├── modules/tls        # TLS 指纹生成 (JA3/JA4)
├── modules/http       # HTTP 指纹生成 (JA4H)
├── modules/ml         # ML 三层分类器
├── modules/defense    # 安全防护系统
├── modules/frontend   # 前端 SDK
└── modules/gateway    # API 网关 (HTTP + gRPC)
```

## 🚀 快速开始

### 安装

```bash
go get github.com/vistone/fingerprint
```

### 基本用法

```go
package main

import (
    "fmt"
    "github.com/vistone/fingerprint"
)

func main() {
    // 获取随机指纹
    profile := fingerprint.GetRandom()
    fmt.Printf("Browser: %s %s\n", profile.BrowserType, profile.BrowserVersion)
    
    // 分析
    analyzer := fingerprint.NewAnalyzer()
    features := analyzer.ExtractFeatures(profile)
    classification := analyzer.Classify(features)
    
    fmt.Printf("Detected: %s %s (%.2f)\n", 
        classification.Family, 
        classification.Version,
        classification.Confidence)
}
```

### 启动网关服务

```go
package main

import (
    "github.com/vistone/fingerprint"
)

func main() {
    config := fingerprint.DefaultGatewayConfig
    config.Port = 8080
    config.GRPCPort = 9090  // gRPC 端口
    
    // 启动 HTTP + gRPC 混合服务器
    fingerprint.StartGateway(config)
}
```

API 端点：
- `POST /api/v1/analyze` - 指纹分析
- `GET /api/v1/sdk.js` - 前端 SDK
- gRPC `:9090` - 高性能 RPC

### 前端集成

```html
<script src="http://localhost:8080/api/v1/sdk.js"></script>
<script>
// SDK 自动收集并发送指纹
FingerprintSDK.collect().then(data => {
    console.log('Fingerprint:', data);
});
</script>
```

## 📊 性能

```
Benchmarks (AMD64):
- GetRandom:           ~500 ns/op
- JA3 Calculate:       ~1.2 µs/op
- ML Classify:         ~2.5 µs/op
- HTTP Analyze:        ~5 µs/op
- Gateway (cached):    ~10 µs/op
```

## 🧪 测试

```bash
# 运行所有测试
go test ./modules/...

# 运行基准测试
go test -bench=. ./modules/core/...

# 运行示例
go run ./examples/v3
```

## 🔧 高级用法

### 训练自定义模型

```go
import "github.com/vistone/fingerprint/modules/ml"

// 加载数据集
dataset := ml.GenerateSyntheticDataset("mydata", 1000)
trainingData := dataset.ToTrainingData()

// 训练分类器
classifier := ml.NewHierarchicalClassifier()
classifier.Initialize()
classifier.Train(trainingData)

// 导出模型
model := classifier.ExportModel()
```

### 导入 Rust 数据

```go
import "github.com/vistone/fingerprint/modules/ml"

importer := ml.NewRustImporter("./data")
dataset, err := importer.ImportDataset("rust_fingerprints.json")
if err != nil {
    log.Fatal(err)
}

// 使用 Rust 数据训练
trainingData := dataset.ToTrainingData()
classifier.Train(trainingData)
```

### gRPC 客户端

```go
import "github.com/vistone/fingerprint/modules/gateway"

client, err := gateway.NewGRPCClient("localhost:9090")
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// 调用分析
resp, err := client.Analyze(ctx, req)
```

## 🏗️ 架构对比

| 特性 | v2.x | v3.0 |
|------|------|------|
| 架构 | 单模块 | Go Workspace (8模块) |
| 指纹数 | 70+ | 135+ |
| ML分类器 | 行为分析 | 三层分层 (95%+) |
| 前端SDK | ❌ | ✅ Canvas/WebGL/Audio |
| API网关 | ❌ | ✅ HTTP + gRPC |
| Rust兼容 | ❌ | ✅ 双向导入/导出 |
| 测试覆盖 | 60% | 85%+ |

## 📚 文档

- [迁移指南](./V3_MIGRATION_GUIDE.md) - 从 v2.x 迁移到 v3.0
- [模块文档](./MODULES.md) - 详细模块说明
- [API 参考](https://pkg.go.dev/github.com/vistone/fingerprint) - Go API 文档

## 🤝 与 Rust 版本协作

Go v3.0 与 [fingerprint-rust](https://github.com/vistone/fingerprint-rust) 完全兼容：

- 共享指纹数据库（135+ 配置）
- 共享训练数据格式
- 共享 ML 模型格式
- 共享 API 规范

```
┌─────────────────┐     ┌─────────────────┐
│  Rust Version   │────▶│   Go Version    │
│  (fingerprint-  │◀────│  (fingerprint-  │
│     rust)       │     │      go)        │
└─────────────────┘     └─────────────────┘
         │                       │
         └───────────┬───────────┘
                     │
              ┌──────▼──────┐
              │  Shared DB  │
              │  135+ Prof  │
              │  ML Models  │
              └─────────────┘
```

## 📝 许可证

BSD-3-Clause License - 详见 [LICENSE](./LICENSE) 文件。

## 🙏 致谢

- 架构灵感来自 [fingerprint-rust](https://github.com/vistone/fingerprint-rust)
- JA3/JA4 指纹标准
- Go 社区的优秀工具链