# 架构重新设计提案

本文档提出 fingerprint 库的架构优化方案，旨在提高可维护性、可扩展性和性能。

## 当前架构问题

### 1. Profile 管理
- **问题**: Profile 定义分散在多个 Go 文件中，难以维护
- **影响**: 添加新 Profile 需要修改代码并重新编译
- **现状**: 90+ Profile 分布在 6+ 文件中

### 2. 配置管理
- **问题**: 缺乏统一的配置系统
- **影响**: 无法热重载配置，必须重启服务
- **现状**: 硬编码配置为主

### 3. 扩展性
- **问题**: 难以添加新的指纹算法
- **影响**: 新算法需要修改核心代码
- **现状**: JA3/JA4 硬编码在各自包中

### 4. 测试覆盖
- **问题**: 部分模块缺乏测试
- **影响**: 难以保证代码质量
- **现状**: 整体覆盖率约 60%

## 目标架构

```plaintext
┌─────────────────────────────────────────────────────────────────┐
│                        API Layer                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │ REST API     │  │ gRPC API     │  │ WebSocket API        │  │
│  └──────────────┘  └──────────────┘  └──────────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│                      Service Layer                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │ Fingerprint  │  │ Profile      │  │ Analytics            │  │
│  │ Service      │  │ Service      │  │ Service              │  │
│  └──────────────┘  └──────────────┘  └──────────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│                      Core Engine                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │ Parser       │  │ Analyzer     │  │ Generator            │  │
│  │ Registry     │  │ Registry     │  │ Registry             │  │
│  └──────────────┘  └──────────────┘  └──────────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│                      Plugin System                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │ TLS Plugins  │  │ HTTP Plugins │  │ TCP/IP Plugins       │  │
│  └──────────────┘  └──────────────┘  └──────────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│                      Storage Layer                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │ Profile      │  │ Config       │  │ Metrics              │  │
│  │ Store        │  │ Store        │  │ Store                │  │
│  └──────────────┘  └──────────────┘  └──────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```plaintext

## 核心改进

### 1. 插件化架构

#### Parser 接口

```go
// types/parser.go
package types

type Parser interface {
    // 解析器名称
    Name() string
    
    // 支持的 MIME 类型
    MimeTypes() []string
    
    // 解析数据
    Parse(ctx context.Context, data []byte) (ParsedData, error)
    
    // 版本信息
    Version() string
}

// 解析器注册表
type ParserRegistry struct {
    parsers map[string]Parser
    mu      sync.RWMutex
}

func (r *ParserRegistry) Register(p Parser) error
func (r *ParserRegistry) Get(name string) (Parser, bool)
func (r *ParserRegistry) List() []Parser
```plaintext

#### Analyzer 接口

```go
// types/analyzer.go
package types

type Analyzer interface {
    Name() string
    
    // 分析数据，生成指纹
    Analyze(ctx context.Context, data ParsedData) (Fingerprint, error)
    
    // 返回此分析器生成的指纹类型
    FingerprintType() string
    
    // 置信度评分 (0-1)
    Confidence() float64
}
```plaintext

#### Generator 接口

```go
// types/generator.go
package types

type Generator interface {
    Name() string
    
    // 生成客户端配置
    Generate(ctx context.Context, profile Profile) (ClientConfig, error)
    
    // 支持的协议版本
    SupportedVersions() []string
}
```plaintext

### 2. Profile 存储系统

#### 存储接口

```go
// types/storage.go
package types

type ProfileStore interface {
    // 基础 CRUD
    Get(ctx context.Context, name string) (Profile, error)
    List(ctx context.Context, filter Filter) ([]Profile, error)
    Create(ctx context.Context, profile Profile) error
    Update(ctx context.Context, profile Profile) error
    Delete(ctx context.Context, name string) error
    
    // 高级功能
    Watch(ctx context.Context, ch chan<- ProfileEvent) error
    Search(ctx context.Context, query string) ([]Profile, error)
    Match(ctx context.Context, criteria MatchCriteria) (Profile, error)
}

// 实现选项：
// - FileStore: 本地文件系统
// - EtcdStore: 分布式键值存储
// - DatabaseStore: SQL/NoSQL 数据库
// - CacheStore: 带缓存的包装器
```plaintext

#### Profile 结构

```go
// types/profile.go
package types

type Profile struct {
    Metadata ProfileMetadata `json:"metadata" yaml:"metadata"`
    TLS      *TLSConfig      `json:"tls,omitempty" yaml:"tls,omitempty"`
    HTTP2    *HTTP2Config    `json:"http2,omitempty" yaml:"http2,omitempty"`
    HTTP1    *HTTP1Config    `json:"http1,omitempty" yaml:"http1,omitempty"`
    Behavior *BehaviorConfig `json:"behavior,omitempty" yaml:"behavior,omitempty"`
    
    // 内部字段
    raw     []byte    // 原始 YAML/JSON
    cached  any       // 缓存的解析结果
    mu      sync.RWMutex
}

type ProfileMetadata struct {
    Name        string            `json:"name" yaml:"name"`
    Version     string            `json:"version" yaml:"version"`
    Category    ProfileCategory   `json:"category" yaml:"category"`
    Browser     BrowserInfo       `json:"browser" yaml:"browser"`
    OS          OSInfo            `json:"os" yaml:"os"`
    Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
    Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
    CreatedAt   time.Time         `json:"created_at" yaml:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at" yaml:"updated_at"`
}
```plaintext

### 3. 配置管理系统

#### 配置接口

```go
// types/config.go
package types

type ConfigManager interface {
    // 获取配置
    Get(path string) (Value, error)
    GetString(path string) (string, error)
    GetInt(path string) (int, error)
    GetBool(path string) (bool, error)
    GetDuration(path string) (time.Duration, error)
    
    // 设置配置
    Set(path string, value any) error
    
    // 订阅变更
    Watch(path string, callback func(Event)) (Subscription, error)
    
    // 重载配置
    Reload(ctx context.Context) error
    
    // 验证配置
    Validate() error
}

// 配置源
type ConfigSource interface {
    Load() (map[string]any, error)
    Watch(changes chan<- ConfigChange) error
}
```plaintext

#### 配置实现

```go
// config/manager.go
package config

type Manager struct {
    sources   []ConfigSource
    values    atomic.Value // map[string]any
    watchers  map[string][]*Watcher
    mu        sync.RWMutex
    schema    *Schema
}

func (m *Manager) Load(ctx context.Context) error {
    merged := make(map[string]any)
    
    for _, source := range m.sources {
        data, err := source.Load()
        if err != nil {
            return fmt.Errorf("load from %s: %w", source.Name(), err)
        }
        mergeConfig(merged, data)
    }
    
    if err := m.schema.Validate(merged); err != nil {
        return fmt.Errorf("validate: %w", err)
    }
    
    m.values.Store(merged)
    return nil
}
```plaintext

### 4. 流水线引擎

```go
// pipeline/engine.go
package pipeline

type Engine struct {
    stages []Stage
    mu     sync.RWMutex
}

type Stage struct {
    Name      string
    Processor Processor
    Parallel  bool
    ContinueOnError bool
    Timeout   time.Duration
}

type Processor interface {
    Process(ctx context.Context, input any) (any, error)
}

func (e *Engine) Execute(ctx context.Context, input any) (any, error) {
    current := input
    
    for _, stage := range e.stages {
        if stage.Timeout > 0 {
            var cancel context.CancelFunc
            ctx, cancel = context.WithTimeout(ctx, stage.Timeout)
            defer cancel()
        }
        
        output, err := stage.Processor.Process(ctx, current)
        if err != nil {
            if !stage.ContinueOnError {
                return nil, fmt.Errorf("stage %s: %w", stage.Name, err)
            }
            continue
        }
        
        current = output
    }
    
    return current, nil
}
```plaintext

## 模块重构计划

### Phase 1: 基础设施 (2 周)

1. **创建类型包** (`types/`)
   - 定义核心接口
   - 定义 Profile 结构
   - 定义配置类型

2. **实现配置管理**
   - ConfigManager 实现
   - 文件/环境变量源
   - 热重载支持

3. **实现存储接口**
   - FileStore 实现
   - Profile 缓存

### Phase 2: 核心重构 (3 周)

1. **解析器注册表**
   ```go
   // tls/parser.go
   type TLSParser struct{}
   
   func (p *TLSParser) Parse(ctx context.Context, data []byte) (types.ParsedData, error)
   
   // 注册
   func init() {
       registry.RegisterParser(&TLSParser{})
   }
   ```

2. **分析器注册表**
   ```go
   // ja3/analyzer.go
   type JA3Analyzer struct{}
   
   func (a *JA3Analyzer) Analyze(ctx context.Context, data types.ParsedData) (types.Fingerprint, error)
   
   // 注册
   func init() {
       registry.RegisterAnalyzer(&JA3Analyzer{})
   }
   ```

3. **生成器注册表**
   ```go
   // profiles/generator.go
   type ProfileGenerator struct{}
   
   func (g *ProfileGenerator) Generate(ctx context.Context, profile types.Profile) (types.ClientConfig, error)
   ```

### Phase 3: 服务层 (2 周)

1. **Fingerprint Service**
   ```go
   type FingerprintService struct {
       parsers   *ParserRegistry
       analyzers *AnalyzerRegistry
       store     ProfileStore
   }
   
   func (s *FingerprintService) Fingerprint(ctx context.Context, req *FingerprintRequest) (*FingerprintResponse, error)
   ```

2. **Profile Service**
   ```go
   type ProfileService struct {
       store     ProfileStore
       generator *ProfileGenerator
   }
   
   func (s *ProfileService) GetProfile(ctx context.Context, name string) (*Profile, error)
   func (s *ProfileService) MatchProfile(ctx context.Context, hints *ClientHints) (*Profile, error)
   ```

3. **Analytics Service**
   ```go
   type AnalyticsService struct {
       metrics   MetricsCollector
       storage   TimeSeriesStorage
   }
   
   func (s *AnalyticsService) GetStats(ctx context.Context, req *StatsRequest) (*StatsResponse, error)
   ```

### Phase 4: API 层 (1 周)

1. **HTTP API**
   ```go
   type HTTPServer struct {
       fingerprintSvc *FingerprintService
       profileSvc     *ProfileService
       analyticsSvc   *AnalyticsService
   }
   
   func (s *HTTPServer) RegisterRoutes(r *mux.Router)
   ```

2. **gRPC API** (可选)

3. **WebSocket API** (实时指纹分析)

## 迁移策略

### 向后兼容

```go
// profiles/deprecated.go
// 保留旧 API 作为包装器

// Deprecated: Use types.Profile instead
func GetClientHelloSpec() (tls.ClientHelloSpec, error) {
    // 内部调用新的 ProfileStore
    profile, _ := store.Get(context.Background(), "chrome_120")
    return generator.GenerateTLS(profile)
}
```plaintext

### 数据迁移

1. **Profile 迁移**: 使用已有的 `cmd/profilegen/migrate` 工具
2. **配置迁移**: 提供配置文件转换脚本
3. **API 迁移**: 保留旧端点，添加 deprecation header

## 性能优化

### 缓存策略

```go
// cache/lru.go
type LRUCache struct {
    cache *lru.Cache
    stats *CacheStats
}

func (c *LRUCache) Get(key string) (any, bool)
func (c *LRUCache) Set(key string, value any, ttl time.Duration)
```plaintext

### 并行处理

```go
// 并行分析多个算法
func (e *Engine) AnalyzeParallel(ctx context.Context, data ParsedData) ([]Fingerprint, error) {
    var wg sync.WaitGroup
    results := make(chan Fingerprint, len(e.analyzers))
    
    for _, analyzer := range e.analyzers {
        wg.Add(1)
        go func(a Analyzer) {
            defer wg.Done()
            fp, err := a.Analyze(ctx, data)
            if err == nil {
                results <- fp
            }
        }(analyzer)
    }
    
    go func() {
        wg.Wait()
        close(results)
    }()
    
    var fingerprints []Fingerprint
    for fp := range results {
        fingerprints = append(fingerprints, fp)
    }
    
    return fingerprints, nil
}
```plaintext

## 测试策略

### 接口测试

```go
func TestParserRegistry(t *testing.T) {
    registry := NewParserRegistry()
    
    parser := &mockParser{name: "test"}
    require.NoError(t, registry.Register(parser))
    
    got, ok := registry.Get("test")
    require.True(t, ok)
    assert.Equal(t, parser, got)
}
```plaintext

### 集成测试

```go
func TestFingerprintService(t *testing.T) {
    svc := setupTestService(t)
    
    resp, err := svc.Fingerprint(context.Background(), &FingerprintRequest{
        Data: testClientHello,
    })
    
    require.NoError(t, err)
    assert.NotEmpty(t, resp.JA3)
    assert.NotEmpty(t, resp.JA4)
}
```plaintext

## 实施时间表

| 阶段 | 时长 | 主要工作 |
| ------ | ------ | ---------- |
| Phase 0 | 1 周 | 设计评审、接口冻结 |
| Phase 1 | 2 周 | 基础设施、类型定义 |
| Phase 2 | 3 周 | 核心重构、注册表 |
| Phase 3 | 2 周 | 服务层实现 |
| Phase 4 | 1 周 | API 层、文档 |
| Phase 5 | 2 周 | 测试、优化、迁移 |
| **总计** | **11 周** | |

## 风险评估

| 风险 | 可能性 | 影响 | 缓解措施 |
| ------ | -------- | ------ | ---------- |
| 向后兼容破坏 | 中 | 高 | 提供适配层，渐进式迁移 |
| 性能下降 | 中 | 高 | 基准测试，必要时优化 |
| 项目延期 | 高 | 中 | 分阶段实施，MVP 优先 |
| 团队学习成本 | 中 | 低 | 文档、培训、代码审查 |

## 结论

新架构将提供：
1. **更好的可维护性**: 清晰的模块边界
2. **更强的可扩展性**: 插件化设计
3. **更高的灵活性**: 配置驱动
4. **更好的测试性**: 接口隔离
5. **更好的性能**: 缓存和并行化

建议分阶段实施，确保每个阶段都有可用的产出。
