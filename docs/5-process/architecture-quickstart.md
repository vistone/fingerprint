# 架构现代化快速启动指南

> 本指南提供立即可执行的步骤，无需等待完整的改造计划

## 📋 执行优先级（基于复杂度和收益）

```plaintext
🔴 High Priority (立即开始，1-2个月)
├── 可观测性集成 (P1)
├── 流水线模式 PoC (P2)
└── 事件溯源框架设计 (P3)

🟡 Medium Priority (2-3个月)
├── 事件溯源完整实现
├── 插件化架构框架
└── 现有代码迁移

🟢 Low Priority (可选，3-6个月)
└── WASM 规则引擎
```plaintext

---

## Week 1-2：可观测性集成（你现在可以做的）

### 目标
- 为核心路径添加 Prometheus 指标
- 集成 OpenTelemetry 追踪
- 实现结构化日志

### 依赖包

```bash
go get github.com/prometheus/client_golang@latest
go get go.opentelemetry.io/otel@latest
go get go.opentelemetry.io/otel/sdk/trace@latest
go get go.opentelemetry.io/otel/exporters/jaeger@latest
go get go.uber.org/zap@latest
```plaintext

### 步骤1：在 main.go 中初始化

```go
package main

import (
    "github.com/vistone/fingerprint/internal/observability"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "net/http"
)

func main() {
    // 初始化追踪
    observability.InitTracer("fingerprint")
    
    // 初始化日志
    logger := observability.NewLogger()
    
    // 初始化指标
    metrics := observability.NewFingerprintMetrics()
    
    // 暴露 Prometheus 指标
    go http.ListenAndServe(":8080", http.HandlerFunc(promhttp.ServeHTTP))
    
    // 你的业务代码
    // ...
}
```plaintext

### 步骤2：为现有代码添加指标

```go
// 在 BehaviorAnalyzer.Analyze() 中

func (ba *BehaviorAnalyzer) Analyze(ctx context.Context, clientIP string) error {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        ba.metrics.AnalysisDuration.
            WithLabelValues("behavior").
            Observe(duration.Seconds())
    }()
    
    // 你的分析逻辑
    // ...
}
```plaintext

### 步骤3：验证

```bash
# 构建
go build ./...

# 启动服务
./fingerprint

# 查看指标（另一个终端）
curl http://localhost:8080/metrics | grep fingerprint_
```plaintext

### 预期输出
```plaintext
# HELP fingerprint_analysis_duration_seconds Time spent analyzing behavior
# TYPE fingerprint_analysis_duration_seconds histogram
fingerprint_analysis_duration_seconds_bucket{le="0.01",...} 5
fingerprint_analysis_duration_seconds_bucket{le="0.05",...} 12
```plaintext

---

## Week 3-4：流水线框架 PoC

### 目标
- 理解 Pipeline 框架的工作原理
- 创建 PoC 验证可行性
- 测试中间件链

### 步骤1：创建首个流水线

```go
package main

import (
    "context"
    "fmt"
    "github.com/vistone/fingerprint/internal/pipeline"
    "go.opentelemetry.io/otel"
)

func main() {
    tracer := otel.Tracer("fingerprint")
    
    // 创建流水线
    p := pipeline.NewPipeline(tracer).
        AddStage(&ParseJA3Stage{}). // 来自 pipeline/examples_test.go
        AddStage(&AnalyzeJA3Stage{}).
        AddMiddleware(pipeline.NewLoggingMiddleware(yourLogger)).
        AddMiddleware(pipeline.NewMetricsMiddleware(yourMetrics))
    
    // 执行
    tlsData := []byte{/* TLS ClientHello */}
    result, err := p.Execute(context.Background(), tlsData)
    if err != nil {
        fmt.Printf("Pipeline failed: %v\n", err)
    }
}
```plaintext

### 步骤2：对比 vs 当前 ProcessingEngine

| 维度 | 当前 switch-case | Pipeline |
| ------ | ----------------- | ---------- |
| 新增步骤 | 修改 switch | 新建 Stage 类 |
| 中间件 | 需要手写 | 内置支持 |
| 可测试性 | 较差 | 良好 |
| 代码复杂度 | 30 行 | 100-150 行 |

### 步骤3：性能基准

```bash
go test -bench=BenchmarkPipeline -benchmem ./internal/pipeline/...
```plaintext

**预期结果**：
- 单阶段执行：< 1% 开销
- 三阶段 + 3中间件：2-5% 开销

---

## Week 5-6：事件溯源框架设计

### 目标
- 设计 WAL 数据库表结构
- 实现基础事件存储
- 验证性能

### 步骤1：数据库表设计

```sql
-- internal/eventsourcing/schema.sql

CREATE TABLE events (
    id TEXT PRIMARY KEY,
    aggregate_id TEXT NOT NULL,
    type TEXT NOT NULL,
    timestamp DATETIME NOT NULL,
    version INT NOT NULL,
    payload JSON NOT NULL,
    metadata JSON,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_aggregate_timestamp (aggregate_id, timestamp),
    INDEX idx_type (type)
);

CREATE TABLE snapshots (
    aggregate_id TEXT PRIMARY KEY,
    version INT NOT NULL,
    state JSON NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```plaintext

### 步骤2：定义核心事件类型

```go
// internal/eventsourcing/events.go

const (
    EventTypeRequestReceived = "request.received"
    EventTypeAnomalyDetected = "anomaly.detected"
    EventTypeRiskChanged     = "risk.changed"
    EventTypeAnalysisCompleted = "analysis.completed"
)

type RequestReceivedEvent struct {
    ClientIP    string
    SNI         string
    TLSVersion  string
    Timestamp   time.Time
    Headers     map[string]string
}

type AnomalyDetectedEvent struct {
    ClientIP      string
    AnomalyType   string
    RiskScore     float64
    Timestamp     time.Time
}
```plaintext

### 步骤3：实现基础 EventStore

```go
// internal/eventsourcing/store.go

type EventStore interface {
    Append(ctx context.Context, event Event) error
    GetEvents(ctx context.Context, aggregateID string, start, end time.Time) ([]Event, error)
    Snapshot(ctx context.Context, aggregateID string, state interface{}) error
    GetSnapshot(ctx context.Context, aggregateID string) (interface{}, error)
}

// SQLite 实现（轻量)
type SQLiteEventStore struct {
    db *sql.DB
}

func NewSQLiteEventStore(dbPath string) (*SQLiteEventStore, error) {
    db, err := sql.Open("sqlite3", dbPath)
    return &SQLiteEventStore{db: db}, err
}
```plaintext

### 步骤4：性能测试

```go
// internal/eventsourcing/store_test.go

func BenchmarkEventAppend(b *testing.B) {
    store := NewSQLiteEventStore(":memory:")
    defer store.Close()
    
    event := Event{
        ID:          uuid.New().String(),
        Type:        EventTypeRequestReceived,
        AggregateID: "192.168.1.1",
        Timestamp:   time.Now(),
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        store.Append(context.Background(), event)
    }
}

func BenchmarkEventReplay(b *testing.B) {
    store := NewSQLiteEventStore(":memory:")
    // ... 插入 10000 个事件
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        store.GetEvents(context.Background(), "192.168.1.1", startTime, endTime)
    }
}
```plaintext

**期望性能**：
- Append: > 1000 ops/sec
- Replay (10K records): < 100ms

---

## Week 7-8：集成到业务代码

### 为 BehaviorAnalyzer 启用事件溯源

```go
// internal/security/behavior/analyzer_v2.go

type BehaviorAnalyzerV2 struct {
    eventStore eventsourcing.EventStore
    metrics    *observability.BehaviorAnalysisMetrics
    logger     *observability.Logger
}

// RecordRequest 记录请求事件
func (ba *BehaviorAnalyzerV2) RecordRequest(ctx context.Context, req *RequestInfo) error {
    event := eventsourcing.RequestReceivedEvent{
        ClientIP:    req.ClientIP,
        SNI:         req.SNI,
        TLSVersion:  req.TLSVersion,
        Timestamp:   time.Now(),
    }
    
    // 异步写入，避免阻塞
    go func() {
        if err := ba.eventStore.Append(context.Background(), event); err != nil {
            ba.logger.Error("failed to record event", "error", err)
        }
    }()
    
    return nil
}

// GetAnalysis 从事件存储重建状态
func (ba *BehaviorAnalyzerV2) GetAnalysis(ctx context.Context, clientIP string, fromTime time.Time) (*BehaviorState, error) {
    events, err := ba.eventStore.GetEvents(ctx, clientIP, fromTime, time.Now())
    if err != nil {
        return nil, err
    }
    
    // 从事件投影到行为状态
    state := &BehaviorState{}
    for _, event := range events {
        ba.applyEvent(state, event)
    }
    
    return state, nil
}
```plaintext

---

## 第二阶段：插件化架构（可以等待）

### 何时开始
- 当有 3+ 个新的指纹类型需要添加
- 当部署频率超过每周一次

### 初步设计

```go
// internal/plugin/manager.go

type Manager struct {
    plugins map[string]Plugin
    mu      sync.RWMutex
}

// LoadPlugin 动态加载 .so 文件
func (m *Manager) LoadPlugin(path string) error {
    p, err := plugin.Open(path)
    if err != nil {
        return err
    }
    
    sym, _ := p.Lookup("NewPlugin")
    newFn := sym.(func() Plugin)
    instance := newFn()
    
    m.mu.Lock()
    m.plugins[instance.GetMetadata().Name] = instance
    m.mu.Unlock()
    
    return nil
}
```plaintext

---

## 第三阶段：WASM 规则引擎（低优先级）

### 何时开始
- 当检测规则变更频率 > 每周一次
- 当需要跨多个服务共享规则

### 工具链
```bash
go get github.com/wasmerio/wasmer-go
# 需要 Rust 工具链（可选）
rustup install stable
cargo install wasm-pack
```plaintext

---

## 监控检查清单

### 第1-2周（可观测性）
- [ ] Prometheus 指标端点可访问
- [ ] 至少 5 个关键指标已导出
- [ ] Jaeger 或类似追踪系统已连接
- [ ] 日志包含 request_id 或 trace_id

### 第3-4周（流水线 PoC）
- [ ] Pipeline 框架可编译
- [ ] 至少 1 个 PoC 流水线已创建
- [ ] 中间件链式执行正确
- [ ] 性能开销 < 5%

### 第5-6周（事件溯源设计）
- [ ] 事件表结构已设计
- [ ] 基础 EventStore 接口已定义
- [ ] SQLite 实现已完成
- [ ] 性能基准已测试

### 第7-8周（集成）
- [ ] BehaviorAnalyzer V2 已实现
- [ ] 旧系统和新系统并行运行
- [ ] 未来 30 天数据可重放
- [ ] 所有测试通过

---

## 常见问题

### Q1: 这些改造会影响现有 API 吗？
**A**: 不会。所有改造都是向后兼容的。旧的 ProcessingEngine 可以保留，新代码使用 Pipeline。

### Q2: 何时应该迁移现有代码？
**A**: 
- 优先：新增功能使用新框架
- 逐步：当修改现有组件时，考虑迁移
- 安全：保留旧系统 1-2 个季度以防需要回滚

### Q3: 如何验证改造成功？
**A**:
- 可观测性：能看到完整的调用链和指标
- 流水线：新增步骤不需要修改核心代码
- 事件溯源：可用历史数据重放算法

### Q4: 对性能的影响？
**A**:
- 可观测性：1-2% 开销（可配置关闭）
- 流水线：2-5% 开销（权衡代码清晰度）
- 事件溯源：异步写入，无阻塞影响

---

## 参考资源

### 推荐阅读
- [Observability Engineering](https://www.oreilly.com/library/view/observability-engineering/9781492076438/) - O'Reilly
- [Building Microservices](https://samnewman.io/books/building_microservices_2nd_edition/) - Sam Newman
- [Designing Data-Intensive Applications](https://dataintensive.net/) - Martin Kleppmann

### 开源项目参考
- [Uber Jaeger](https://www.jaegertracing.io/) - 分布式追踪
- [Prometheus](https://prometheus.io/) - 指标系统
- [EventStoreDB](https://www.eventstore.com/) - 事件存储
- [Apache Kafka](https://kafka.apache.org/) - 事件流处理

### 相关库
```plaintext
go get github.com/prometheus/client_golang
go get go.opentelemetry.io/otel
go get go.uber.org/zap
go get github.com/google/uuid
```plaintext

---

## 下一步行动

### 立即（今日）
- [ ] 复制 `internal/observability/observability.go` 到你的项目
- [ ] 复制 `internal/pipeline/pipeline.go` 到你的项目
- [ ] 运行编译检查：`go build ./...`

### 本周
- [ ] 添加依赖：`go get -u github.com/prometheus/client_golang`
- [ ] 为 1 个关键函数添加指标
- [ ] 为 1 个关键函数添加追踪 span

### 本月
- [ ] 完整的可观测性集成
- [ ] Pipeline PoC 验证
- [ ] 事件溯源框架设计完成
- [ ] 分享进度给团队

---

## 获取帮助

如有问题，可查阅：
- [docs/5-process/architecture-modernization-plan.md](architecture-modernization-plan.md) - 详细方案
- [internal/observability/](../../internal/observability/) - 完整实现示例
- [internal/pipeline/](../../internal/pipeline/) - 框架和测试用例
