# Fingerprint 架构文档

本文档描述 fingerprint 库的整体架构、模块划分和设计决策。

## 架构概览

```
┌─────────────────────────────────────────────────────────────────┐
│                         Application Layer                        │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │   Profiles   │  │   Config     │  │      Generator       │  │
│  │   (pkg)      │  │   (pkg)      │  │      (pkg)           │  │
│  └──────┬───────┘  └──────┬───────┘  └──────────┬───────────┘  │
├─────────┼─────────────────┼────────────────────┼────────────────┤
│         │                 │                    │                │
│  ┌──────▼─────────────────▼────────────────────▼──────────┐    │
│  │                    Core Library                        │    │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐  │    │
│  │  │   TLS    │ │  HTTP/2  │ │  TCP/IP  │ │ Behavior │  │    │
│  │  │  (pkg)   │ │  (pkg)   │ │  (pkg)   │ │  (pkg)   │  │    │
│  │  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘  │    │
│  └───────┼────────────┼────────────┼────────────┼────────┘    │
├──────────┼────────────┼────────────┼────────────┼─────────────┤
│          │            │            │            │             │
│  ┌───────▼────────────▼────────────▼────────────▼──────────┐  │
│  │                   Internal Packages                      │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐   │  │
│  │  │ Extension│ │ Pipeline │ │  Utils   │ │  Cache   │   │  │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘   │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## 模块详解

### 1. Profiles 模块 (`profiles/`)

负责管理 TLS/HTTP 客户端指纹配置。

```
profiles/
├── client_profile.go          # ClientProfile 定义
├── internal_browser_profiles.go   # 内置浏览器配置
├── contributed_browser_profiles.go # 社区贡献配置
├── custom_profiles.go         # 自定义配置
└── specs/                     # YAML 规格目录
    ├── chrome_120.yaml
    ├── firefox_121.yaml
    └── ...
```

**关键类型**:
- `ClientProfile`: 完整的客户端配置
- `ClientHelloSpec`: TLS 握手参数
- `Settings`: HTTP/2 帧设置
- `PseudoHeaderOrder`: 伪头部顺序

### 2. TLS 模块 (`tls/`)

TLS 指纹分析和 JA3/JA4 计算。

```
tls/
├── ja3/              # JA3 指纹计算
├── ja4/              # JA4 指纹计算
├── ja4s/             # JA4S 指纹计算
├── ech/              # Encrypted Client Hello
└── internal/utils/   # 内部工具
```

**功能**:
- JA3 字符串解析和生成
- JA4 指纹计算
- TLS ClientHello 分析
- ECH 支持（实验中）

### 3. HTTP 模块 (`http/`)

HTTP/2 和 HTTP 头部分析。

```
http/
├── http2/            # HTTP/2 帧分析
├── ja4h/             # JA4H HTTP 指纹
├── headers/          # 头部分析
├── useragent/        # User-Agent 解析
├── clienthints/      # Client Hints 处理
└── policy/           # HTTP 策略
```

**功能**:
- HTTP/2 帧结构分析
- SETTINGS、WINDOW_UPDATE、PRIORITY 帧解析
- HTTP 头部指纹 (JA4H)
- Client Hints 处理

### 4. TCP/IP 模块 (`internal/tcpip/`)

TCP/IP 层指纹分析。

**功能**:
- TCP 窗口大小分析
- TCP 选项解析
- TTL/IPID 分析
- OS 指纹识别

### 5. Config 模块 (`internal/config/`)

配置管理和动态更新。

**关键类型**:
- `ManagedConfig`: 可热重载的配置
- `FeatureExtractionConfig`: 特征提取配置

**特性**:
- JSON/YAML 配置解析
- 配置热重载（通过文件监听）
- 配置克隆（优化版本）

### 6. Internal 模块

#### 6.1 Extension 模块 (`internal/extension/`)

TLS 扩展解析。

支持的扩展类型:
- `server_name` (0)
- `supported_groups` (10)
- `ec_point_formats` (11)
- `signature_algorithms` (13)
- `application_layer_protocol_negotiation` (16)
- `supported_versions` (43)
- `psk_key_exchange_modes` (45)
- `key_share` (51)
- `grease_placeholder` ( grease )

#### 6.2 Pipeline 模块 (`internal/pipeline/`)

处理管道和中间件。

```go
type Processor interface {
    Process(ctx context.Context, data interface{}) (interface{}, error)
}

type Pipeline struct {
    processors []Processor
}
```

#### 6.3 Cache 模块 (`internal/cache/`)

缓存实现。

```go
type Cache interface {
    Get(key string) (interface{}, bool)
    Set(key string, value interface{}, ttl time.Duration)
    Delete(key string)
}
```

## 数据流

### 指纹生成流程

```
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
│  Client  │───▶│   TLS    │───▶│   HTTP   │───▶│  Output  │
│  Request │    │  Layer   │    │  Layer   │    │          │
└──────────┘    └────┬─────┘    └────┬─────┘    └──────────┘
                     │               │
                     ▼               ▼
              ┌──────────┐    ┌──────────┐
              │  JA3/4   │    │  JA4H    │
              │Fingerprint│   │Fingerprint│
              └──────────┘    └──────────┘
```

### Profile 选择流程

```
┌──────────────┐
│ User Request │
└──────┬───────┘
       │
       ▼
┌──────────────┐     ┌──────────────┐
│ Parse UA/    │────▶│ Match        │
│ Client Hints │     │ Profile      │
└──────────────┘     └──────┬───────┘
                            │
              ┌─────────────┼─────────────┐
              ▼             ▼             ▼
        ┌─────────┐   ┌─────────┐   ┌─────────┐
        │ Chrome  │   │ Firefox │   │ Safari  │
        └────┬────┘   └────┬────┘   └────┬────┘
             │             │             │
             └─────────────┴─────────────┘
                           │
                           ▼
                    ┌──────────────┐
                    │ Return Spec  │
                    └──────────────┘
```

## 设计原则

### 1. 模块化设计

- 每个模块有明确的职责边界
- 模块间通过接口通信
- 支持独立测试和替换

### 2. 可扩展性

```go
// 处理器接口允许自定义扩展
type FingerprintProcessor interface {
    Process(ctx context.Context, req *Request) (*Fingerprint, error)
}

// 插件系统
func RegisterPlugin(name string, plugin Plugin) error
```

### 3. 性能优化

| 组件 | 优化策略 | 效果 |
|------|----------|------|
| Config Clone | 手写 Clone 替代 JSON | 5-10x 提升 |
| Profile Cache | LRU 缓存 | 避免重复生成 |
| Extension Parse | 预分配切片 | 减少内存分配 |

### 4. 线程安全

```go
// 使用 sync.RWMutex 保护共享状态
type Manager struct {
    mu      sync.RWMutex
    configs map[string]*Config
}

func (m *Manager) Get(name string) (*Config, bool) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    c, ok := m.configs[name]
    return c, ok
}
```

## 依赖关系

```
profiles ──┬──▶ tls
           ├──▶ http
           └──▶ internal/extension

tls ───────▶ internal/utils

http ──────▶ tls (optional)
           ▶ internal/extension

internal/config ──▶ (no external deps in pkg)

internal/tcpip ───▶ (self-contained)
```

## 性能特征

| 操作 | 时间复杂度 | 空间复杂度 | 说明 |
|------|-----------|-----------|------|
| Profile 选择 | O(1) | O(1) | 哈希表查找 |
| JA3 解析 | O(n) | O(n) | n = 字符串长度 |
| HTTP/2 分析 | O(m) | O(m) | m = 帧数量 |
| Config Clone | O(1) | O(1) | 固定大小结构 |

## 扩展点

### 1. 自定义 Profile 源

```go
type ProfileSource interface {
    Load() ([]ClientProfile, error)
    Watch(changes chan<- ProfileChange) error
}

// 实现：文件、数据库、配置中心等
```

### 2. 自定义分析器

```go
type Analyzer interface {
    Name() string
    Analyze(ctx context.Context, data []byte) (Analysis, error)
}

// 注册：RegisterAnalyzer("custom", &CustomAnalyzer{})
```

### 3. 输出格式化

```go
type Formatter interface {
    Format(fp *Fingerprint) ([]byte, error)
}

// JSON, YAML, Protobuf 等实现
```

## 未来架构演进

### 阶段 1: 配置外部化（当前）
- 将 Profile 从代码迁移到 YAML
- 支持热重载

### 阶段 2: 插件化（计划中）
- WASM 插件支持
- 动态加载分析器

### 阶段 3: 分布式（长期）
- 指纹数据共享
- 分布式缓存
