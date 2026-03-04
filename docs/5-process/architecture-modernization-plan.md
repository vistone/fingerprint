# 架构现代化改造计划

## 执行摘要

基于项目当前的 `ProcessingEngine`（switch-case）、`BehaviorAnalyzer`（聚合统计）、以及现有的扩展机制，提出5个前瞻性架构改造建议。该计划分3个阶段，预期在6-12个月内逐步实施。

### 执行进展（2026-03-03）

- Week 5 Day 1-2 的灰度执行脚本已就绪（预检、阶段切换、监控、回滚）。
- 脚本目录：`scripts/canary/`
    - `precheck_day1.sh`：上线前自动预检（编译、测试、K8s配置检查）
    - `set_canary_stage.sh`：灰度阶段切换（5/25/50/100/off）
    - `monitor_canary.sh`：周期监控并在严重阈值触发时返回非零状态
    - `rollback_canary.sh`：快速回滚并采集诊断数据
- 操作说明：`scripts/canary/readme.md`
- 下一执行窗口：Week 5 Day 1（2026-03-09）按 5% 灰度流程落地。
- 一键编排入口：`scripts/canary/run_day1_canary.sh`（预检→切换→监控→自动回滚）。

**📦 并行进行：包结构重构计划 (Week 5-8)**

新增完整的包结构重构计划，目标将混乱的包结构改造为清晰的 3 层架构：
- **Phase 1 TLS**: Week 5-6，低风险，可立即启动
- **Phase 2 HTTP**: Week 6-7，中等风险
- **Phase 3 pkg化**: Week 7-8，高风险但带来长期收益

文档与工具已完全就绪：
- 🚀 [快速启动指南（5分钟入门）](restructuring-quickstart.md)
- 📖 [完整重构规划](package-restructuring-plan.md)
- 📋 [Phase 1 import 变更清单](phase1-import-mapping.md)
- 🔧 自动迁移脚本：`scripts/phase1_tls_migration.sh`

与灰度推出（Week 5-7）完全并行，无任何冲突。

---

## 第1部分：建议评估矩阵

| 建议 | 优先级 | 复杂度 | 收益 | 当前支持度 | 风险等级 |
| ------ | ------- | ------- | ------ | ----------- | --------- |
| **1. 插件化架构** | 中 | 高 | 高 | 25% | 中 |
| **2. 流水线模式** | 高 | 中 | 高 | 30% | 低 |
| **3. 事件溯源** | 高 | 高 | 极高 | 10% | 中 |
| **4. WASM 支持** | 低 | 极高 | 中 | 0% | 高 |
| **5. 可观测性** | 高 | 低 | 高 | 20% | 极低 |

---

## 第2部分：详细方案

### 1️⃣ 插件化架构（Plugin System）

#### 🎯 现状分析
- **当前扩展机制**：在 `internal/extension/` 中使用编译时注册
  ```go
  // 现状：紧耦合的组件加载
  registry.Register("ja3", JA3ParserComponent{})
  registry.Register("ja4", JA4ParserComponent{})
  ```
- **问题**：新增解析器需要重新编译和部署整个项目

#### 📋 实现方案

**阶段1：定义插件接口（2-3周）**

```go
// internal/plugin/plugin.go
package plugin

import "context"

// Plugin 所有插件必须实现的基础接口
type Plugin interface {
    // GetMetadata 返回插件元数据
    GetMetadata() PluginMetadata
    
    // Initialize 初始化插件（加载配置、启动资源）
    Initialize(ctx context.Context, cfg map[string]interface{}) error
    
    // Execute 执行插件的核心逻辑
    Execute(ctx context.Context, input interface{}) (interface{}, error)
    
    // Shutdown 清理资源
    Shutdown(ctx context.Context) error
}

// PluginMetadata 插件元数据
type PluginMetadata struct {
    Name        string
    Version     string
    Author      string
    Description string
    InputType   string
    OutputType  string
    Requires    []string // 依赖的其他插件
}

// ParserPlugin 指纹解析器插件
type ParserPlugin interface {
    Plugin
    Parse(ctx context.Context, data []byte) (ParseResult, error)
}

// AnalyzerPlugin 分析器插件
type AnalyzerPlugin interface {
    Plugin
    Analyze(ctx context.Context, parsed ParseResult) (AnalysisResult, error)
}
```plaintext

**阶段2：插件管理器（3-4周）**

```go
// internal/plugin/manager.go
package plugin

import (
    "plugin"
    "sync"
)

// Manager 管理所有加载的插件
type Manager struct {
    mu          sync.RWMutex
    plugins     map[string]Plugin
    pluginPath  string
    pluginDeps  map[string][]string // 追踪依赖关系
}

// LoadPlugin 动态加载插件（.so 文件）
func (m *Manager) LoadPlugin(ctx context.Context, path string, cfg map[string]interface{}) error {
    // 使用 Go 的 plugin 包加载动态库
    // 支持热加载和版本管理
    
    p, err := plugin.Open(path)
    if err != nil {
        return fmt.Errorf("failed to load plugin: %w", err)
    }
    
    // 加载指定的符号（必须是 NewPlugin 函数）
    symNewPlugin, err := p.Lookup("NewPlugin")
    if err != nil {
        return fmt.Errorf("plugin missing NewPlugin symbol: %w", err)
    }
    
    newPluginFunc := symNewPlugin.(func() Plugin)
    instance := newPluginFunc()
    
    // 初始化插件
    if err := instance.Initialize(ctx, cfg); err != nil {
        return fmt.Errorf("plugin initialization failed: %w", err)
    }
    
    m.mu.Lock()
    m.plugins[instance.GetMetadata().Name] = instance
    m.mu.Unlock()
    
    return nil
}

// GetPlugin 获取已加载的插件
func (m *Manager) GetPlugin(name string) (Plugin, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    plugin, exists := m.plugins[name]
    if !exists {
        return nil, fmt.Errorf("plugin %s not found", name)
    }
    return plugin, nil
}

// ReloadPlugin 热加载（不中断服务）
func (m *Manager) ReloadPlugin(ctx context.Context, name string, path string, cfg map[string]interface{}) error {
    // 1. 验证新插件
    // 2. 初始化新实例
    // 3. 原子性替换（使用 atomic.CompareAndSwap）
    // 4. 关闭旧实例
    // 5. 回滚机制：如果失败，恢复旧实例
    
    // 实现详见下面的 "热加载演进" 部分
    return nil
}
```plaintext

**阶段3：从组件到插件的迁移（4-6周）**

```go
// plugins/ja3_parser.go - 示例：JA3 解析器插件
package main

import (
    "context"
    "github.com/vistone/fingerprint/internal/plugin"
)

// JA3ParserPlugin JA3 指纹解析器插件
type JA3ParserPlugin struct {
    config map[string]interface{}
}

func NewPlugin() plugin.Plugin {
    return &JA3ParserPlugin{}
}

func (p *JA3ParserPlugin) GetMetadata() plugin.PluginMetadata {
    return plugin.PluginMetadata{
        Name:        "ja3-parser",
        Version:     "1.0.0",
        Author:      "fingerprint team",
        Description: "JA3 TLS fingerprint parser",
        InputType:   "[]byte",
        OutputType:  "ParseResult",
        Requires:    []string{},
    }
}

func (p *JA3ParserPlugin) Initialize(ctx context.Context, cfg map[string]interface{}) error {
    p.config = cfg
    // 加载规则、初始化资源
    return nil
}

func (p *JA3ParserPlugin) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    data := input.([]byte)
    return p.Parse(ctx, data)
}

func (p *JA3ParserPlugin) Parse(ctx context.Context, data []byte) (interface{}, error) {
    // JA3 解析逻辑（从现有代码迁移）
    return nil, nil
}

func (p *JA3ParserPlugin) Shutdown(ctx context.Context) error {
    // 清理资源
    return nil
}
```plaintext

#### 🔄 热加载演进

```go
// 支持不中断服务的插件更新
type VersionedPlugin struct {
    Current  Plugin
    Staging  Plugin          // 新版本（待验证）
    Config   map[string]interface{}
    mu       sync.RWMutex
}

func (vp *VersionedPlugin) Promote(ctx context.Context) error {
    vp.mu.Lock()
    defer vp.mu.Unlock()
    
    if vp.Staging == nil {
        return fmt.Errorf("no staged version")
    }
    
    // 验证新版本：运行单元测试和集成测试
    if err := vp.ValidateStaging(ctx); err != nil {
        return fmt.Errorf("staging validation failed: %w", err)
    }
    
    // 金丝雀部署：让部分流量用新版本
    vp.Canary = vp.Staging
    
    // 监控：如果错误率上升 > 5%，自动回滚
    if err := vp.Monitor(ctx); err != nil {
        vp.Rollback()
        return err
    }
    
    // 全量切换
    oldCurrent := vp.Current
    vp.Current = vp.Staging
    vp.Staging = nil
    
    // 异步关闭旧版本
    go func() {
        if err := oldCurrent.Shutdown(context.Background()); err != nil {
            log.Printf("warning: old plugin shutdown failed: %v", err)
        }
    }()
    
    return nil
}
```plaintext

#### 📊 收益评估
- ✅ **可观测架构**：无需重新编译可加载新指纹解析器
- ✅ **故障隔离**：插件崩溃不影响主进程
- ✅ **灰度部署**：支持金丝雀部署和蓝绿部署
- ⚠️ **学习曲线**：开发者需学习插件开发框架

---

### 2️⃣ 流水线模式（Pipeline Pattern）

#### 🎯 现状分析
- **当前代码**（ProcessingEngine.Process）：
  ```go
  for _, step := range steps {
      switch step {
      case "parse": ...
      case "analyze": ...
      case "transform": ...
      default: return error
      }
  }
  ```
- **问题**：
  - 新增步骤需要修改 switch-case
  - 步骤间的依赖关系隐含（如 "transform" 依赖 "parse"）
  - 难以添加中间件、日志、性能监控

#### 📋 实现方案

**设计原则：链式责任模式 + 中间件支持**

```go
// internal/pipeline/pipeline.go
package pipeline

import (
    "context"
    "fmt"
    "time"
)

// Stage 流水线的单个阶段
type Stage interface {
    // GetName 返回阶段名称
    GetName() string
    
    // GetDependencies 返回此阶段依赖的前置阶段
    GetDependencies() []string
    
    // Execute 执行阶段逻辑
    Execute(ctx context.Context, data *StageData) error
}

// Middleware 中间件（装饰器）
type Middleware interface {
    // Process 在 Stage 执行前后运行
    Process(ctx context.Context, stageName string, data *StageData, next ExecutionFunc) error
}

type ExecutionFunc func(ctx context.Context, data *StageData) error

// StageData 流水线数据
type StageData struct {
    Input       interface{}
    Output      interface{}
    Context     map[string]interface{}
    ExecutedAt  time.Time
    Duration    time.Duration
    Error       error
}

// Pipeline 流水线
type Pipeline struct {
    stages      []Stage
    middlewares []Middleware
    stageIndex  map[string]int
}

// NewPipeline 创建新流水线
func NewPipeline() *Pipeline {
    return &Pipeline{
        stages:     []Stage{},
        stageIndex: make(map[string]int),
    }
}

// AddStage 添加阶段（自动处理依赖）
func (p *Pipeline) AddStage(stage Stage) *Pipeline {
    p.stages = append(p.stages, stage)
    p.stageIndex[stage.GetName()] = len(p.stages) - 1
    return p
}

// AddMiddleware 添加中间件
func (p *Pipeline) AddMiddleware(mw Middleware) *Pipeline {
    p.middlewares = append(p.middlewares, mw)
    return p
}

// Validate 验证流水线的有效性
func (p *Pipeline) Validate() error {
    for _, stage := range p.stages {
        for _, dep := range stage.GetDependencies() {
            if _, exists := p.stageIndex[dep]; !exists {
                return fmt.Errorf("stage %s depends on %s, but %s not found", 
                    stage.GetName(), dep, dep)
            }
        }
    }
    return nil
}

// Execute 执行流水线
func (p *Pipeline) Execute(ctx context.Context, input interface{}) (*StageData, error) {
    if err := p.Validate(); err != nil {
        return nil, fmt.Errorf("pipeline validation failed: %w", err)
    }
    
    data := &StageData{
        Input:   input,
        Context: make(map[string]interface{}),
    }
    
    for _, stage := range p.stages {
        stageName := stage.GetName()
        
        // 检查依赖
        if err := p.checkDependencies(stageName, data); err != nil {
            return data, fmt.Errorf("dependency check failed for stage %s: %w", stageName, err)
        }
        
        // 用中间件包装执行
        startTime := time.Now()
        err := p.executeWithMiddleware(ctx, stage, data)
        data.Duration = time.Since(startTime)
        data.ExecutedAt = time.Now()
        
        if err != nil {
            data.Error = err
            return data, fmt.Errorf("stage %s failed: %w", stageName, err)
        }
    }
    
    return data, nil
}

// executeWithMiddleware 使用中间件执行阶段
func (p *Pipeline) executeWithMiddleware(ctx context.Context, stage Stage, data *StageData) error {
    var handler ExecutionFunc = func(ctx context.Context, d *StageData) error {
        return stage.Execute(ctx, d)
    }
    
    // 反向构建中间件链
    for i := len(p.middlewares) - 1; i >= 0; i-- {
        mw := p.middlewares[i]
        nextHandler := handler
        handler = func(ctx context.Context, d *StageData) error {
            return mw.Process(ctx, stage.GetName(), d, nextHandler)
        }
    }
    
    return handler(ctx, data)
}

// checkDependencies 检查依赖的前置阶段是否已执行
func (p *Pipeline) checkDependencies(stageName string, data *StageData) error {
    stageIdx := p.stageIndex[stageName]
    stage := p.stages[stageIdx]
    
    for _, dep := range stage.GetDependencies() {
        depIdx := p.stageIndex[dep]
        if depIdx >= stageIdx {
            return fmt.Errorf("circular dependency detected: %s depends on %s", stageName, dep)
        }
    }
    
    return nil
}
```plaintext

**迁移现有 ProcessingEngine 步骤**

```go
// internal/pipeline/stages.go

// ParseStage JA3/JA4 解析阶段
type ParseStage struct {
    parsers map[string]FingerprintParser
}

func (ps *ParseStage) GetName() string {
    return "parse"
}

func (ps *ParseStage) GetDependencies() []string {
    return []string{} // 无依赖
}

func (ps *ParseStage) Execute(ctx context.Context, data *StageData) error {
    rawData := data.Input.([]byte)
    // 使用多个解析器
    result := make(map[string]interface{})
    for name, parser := range ps.parsers {
        parsed, err := parser.Parse(ctx, rawData)
        if err != nil {
            return fmt.Errorf("parser %s failed: %w", name, err)
        }
        result[name] = parsed
    }
    data.Output = result
    return nil
}

// AnalyzeStage 分析阶段
type AnalyzeStage struct {
    analyzer Analyzer
}

func (as *AnalyzeStage) GetName() string {
    return "analyze"
}

func (as *AnalyzeStage) GetDependencies() []string {
    return []string{"parse"} // 依赖 parse 阶段
}

func (as *AnalyzeStage) Execute(ctx context.Context, data *StageData) error {
    parsed := data.Output.(map[string]interface{})
    analysis, err := as.analyzer.Analyze(ctx, parsed)
    if err != nil {
        return fmt.Errorf("analysis failed: %w", err)
    }
    data.Output = analysis
    return nil
}

// TransformStage 转换阶段
type TransformStage struct {
    transformer Transformer
}

func (ts *TransformStage) GetName() string {
    return "transform"
}

func (ts *TransformStage) GetDependencies() []string {
    return []string{"parse", "analyze"}
}

func (ts *TransformStage) Execute(ctx context.Context, data *StageData) error {
    analysis := data.Output
    transformed, err := ts.transformer.Transform(ctx, analysis)
    if err != nil {
        return fmt.Errorf("transform failed: %w", err)
    }
    data.Output = transformed
    return nil
}
```plaintext

**使用示例（新代码）**

```go
// 创建流水线
pipeline := pipeline.NewPipeline().
    AddStage(&pipeline.ParseStage{parsers: parsersMap}).
    AddStage(&pipeline.AnalyzeStage{analyzer: analyzer}).
    AddStage(&pipeline.TransformStage{transformer: transformer}).
    AddMiddleware(&LoggingMiddleware{}).
    AddMiddleware(&MetricsMiddleware{}).
    AddMiddleware(&TimeoutMiddleware{timeout: 5 * time.Second})

// 执行
result, err := pipeline.Execute(ctx, rawTLSData)
if err != nil {
    log.Printf("Pipeline failed: %v", err)
}
```plaintext

#### 📊 对比分析

| 维度 | switch-case | Pipeline |
| ------ | ----------- | ---------- |
| **代码行数** | 20-30 | 100-150（首次） |
| **新增步骤** | 修改 switch | 新建 Stage 类 |
| **依赖管理** | 隐含 | 显式 |
| **中间件支持** | 无 | 有 |
| **可测试性** | 中 | 高 |
| **性能开销** | <1% | ~2-5% |

---

### 3️⃣ 事件溯源（Event Sourcing）

#### 🎯 现状分析
- **当前 BehaviorAnalyzer**：
  ```go
  type BehaviorAnalyzer struct {
      requestHistory *ring.Ring    // 循环缓冲区
      temporalPatterns map[string]*TemporalPattern
      // ... 存储聚合后的统计
  }
  ```
- **问题**：
  - 聚合后无法重放历史数据
  - 算法更新无法回溯验证
  - 调试困难：不知道具体哪些请求组成了统计数据

#### 📋 实现方案

**核心思想：Write-Ahead Log (WAL) + Event Store**

```go
// internal/eventsourcing/event.go
package eventsourcing

import (
    "time"
)

// EventTypes
const (
    EventTypeRequestReceived = "request.received"
    EventTypeRiskDetected    = "risk.detected"
    EventTypeAnomalyDetected = "anomaly.detected"
    EventTypeBehaviorChanged = "behavior.changed"
)

// Event 基础事件结构
type Event struct {
    // 事件ID（UUID）
    ID string
    
    // 事件类型
    Type string
    
    // 发生时间（精确到微秒）
    Timestamp time.Time
    
    // 聚合根ID（如 ClientIP）
    AggregateID string
    
    // 事件版本（用于重放时的schema演进）
    Version int
    
    // 事件有效负载（JSON）
    Payload map[string]interface{}
    
    // 元数据
    Metadata map[string]string
}

// RequestReceivedEvent 请求接收事件
type RequestReceivedEvent struct {
    Event
    
    ClientIP      string
    SNI           string
    TLSVersion    string
    CipherSuite   string
    Extensions    []string
    UserAgent     string
    Headers       map[string]string
    Timestamp     time.Time
}

// AnomalyDetectedEvent 异常检测事件
type AnomalyDetectedEvent struct {
    Event
    
    AnomalyType   string  // "pattern_change", "timing_anomaly", etc
    RiskScore     float64
    Details       string
    RecommendedAction string
}
```plaintext

**事件存储接口**

```go
// internal/eventsourcing/store.go
package eventsourcing

import (
    "context"
)

// EventStore 事件存储（可多种实现：PostgreSQL, SQLite, Redis Stream）
type EventStore interface {
    // Append 追加事件到日志
    Append(ctx context.Context, event Event) error
    
    // GetEvents 获取指定时间范围内的事件
    GetEvents(ctx context.Context, aggregateID string, start, end time.Time) ([]Event, error)
    
    // GetEventsSince 从指定事件ID之后获取事件（用于增量读）
    GetEventsSince(ctx context.Context, aggregateID string, lastEventID string) ([]Event, error)
    
    // ReplayEvents 从头重放所有事件（用于重建状态）
    ReplayEvents(ctx context.Context, aggregateID string) ([]Event, error)
    
    // Snapshot 创建快照（优化重放性能）
    Snapshot(ctx context.Context, aggregateID string, state interface{}) error
    
    // GetSnapshot 获取最新快照
    GetSnapshot(ctx context.Context, aggregateID string) (interface{}, error)
}

// WALStore Write-Ahead Log 实现
type WALStore struct {
    db      *sql.DB // PostgreSQL 或 SQLite
    bufferSize int
    buffer  []*Event
    mu      sync.RWMutex
}

func NewWALStore(dbPath string) (*WALStore, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }
    
    // 创建表
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS events (
            id TEXT PRIMARY KEY,
            aggregate_id TEXT NOT NULL,
            type TEXT NOT NULL,
            timestamp DATETIME NOT NULL,
            version INT NOT NULL,
            payload JSON NOT NULL,
            metadata JSON,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            INDEX idx_aggregate_timestamp (aggregate_id, timestamp)
        )
    `)
    
    return &WALStore{
        db:         db,
        bufferSize: 1000,
        buffer:     make([]*Event, 0),
    }, err
}

// Append 批量写入（降低写入延迟）
func (w *WALStore) Append(ctx context.Context, event Event) error {
    w.mu.Lock()
    w.buffer = append(w.buffer, &event)
    shouldFlush := len(w.buffer) >= w.bufferSize
    w.mu.Unlock()
    
    if shouldFlush {
        return w.Flush(ctx)
    }
    
    return nil
}

// Flush 刷新缓冲区到磁盘
func (w *WALStore) Flush(ctx context.Context) error {
    w.mu.Lock()
    events := make([]*Event, len(w.buffer))
    copy(events, w.buffer)
    w.buffer = w.buffer[:0]
    w.mu.Unlock()
    
    if len(events) == 0 {
        return nil
    }
    
    // 批量插入
    tx, err := w.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    for _, ev := range events {
        _, err := tx.ExecContext(ctx,
            "INSERT INTO events (id, aggregate_id, type, timestamp, version, payload) VALUES (?, ?, ?, ?, ?, ?)",
            ev.ID, ev.AggregateID, ev.Type, ev.Timestamp, ev.Version, 
            mustMarshalJSON(ev.Payload))
        if err != nil {
            return err
        }
    }
    
    return tx.Commit().Error
}
```plaintext

**事件投影（从事件重建状态）**

```go
// internal/eventsourcing/projections.go
package eventsourcing

// BehaviorProjection 从事件流投影出行为分析状态
type BehaviorProjection struct {
    AggregateID string
    State       *BehaviorState
    LastEventID string
    Version     int
}

type BehaviorState struct {
    RequestCount         int64
    AnomalyCount         int64
    RiskScore            float64
    LastRequestTime      time.Time
    DetectedPatterns     map[string]*Pattern
    AnomalyHistory       []Anomaly
}

// Project 将事件流投影到行为状态
func (bp *BehaviorProjection) Project(ctx context.Context, events []Event) error {
    for _, event := range events {
        switch event.Type {
        case EventTypeRequestReceived:
            bp.State.RequestCount++
            bp.State.LastRequestTime = event.Timestamp
            
            // 更新模式
            clientIP := event.Payload["ClientIP"].(string)
            bp.detectPattern(clientIP, event)
            
        case EventTypeAnomalyDetected:
            bp.State.AnomalyCount++
            riskScore := event.Payload["RiskScore"].(float64)
            if riskScore > bp.State.RiskScore {
                bp.State.RiskScore = riskScore
            }
            
            // 记录异常
            bp.State.AnomalyHistory = append(bp.State.AnomalyHistory, Anomaly{
                Type:      event.Payload["AnomalyType"].(string),
                Timestamp: event.Timestamp,
                Score:     riskScore,
            })
        }
        
        bp.LastEventID = event.ID
        bp.Version = event.Version
    }
    
    return nil
}

// 重放历史数据用于算法验证
func (bp *BehaviorProjection) ReplayFrom(ctx context.Context, startTime time.Time) error {
    // 获取快照（如果有）
    snapshot, err := store.GetSnapshot(ctx, bp.AggregateID)
    if err == nil {
        bp.State = snapshot.(*BehaviorState)
    } else {
        bp.State = &BehaviorState{
            DetectedPatterns: make(map[string]*Pattern),
            AnomalyHistory:   make([]Anomaly, 0),
        }
    }
    
    // 从快照之后的事件开始重放
    events, _ := store.GetEvents(ctx, bp.AggregateID, startTime, time.Now())
    return bp.Project(ctx, events)
}
```plaintext

**迁移 BehaviorAnalyzer**

```go
// internal/security/behavior/analyzer_v2.go (新版本)
package behavior

// 新版 BehaviorAnalyzer：基于事件溯源
type BehaviorAnalyzerV2 struct {
    eventStore eventsourcing.EventStore
    projection *eventsourcing.BehaviorProjection
    
    // 用于实时分析（不等待事件落盘）
    inMemoryBuffer []*eventsourcing.Event
    bufferMu       sync.RWMutex
}

// RecordRequest 记录请求事件
func (ba *BehaviorAnalyzerV2) RecordRequest(ctx context.Context, req *RequestInfo) error {
    event := eventsourcing.RequestReceivedEvent{
        Event: eventsourcing.Event{
            ID:          uuid.New().String(),
            Type:        eventsourcing.EventTypeRequestReceived,
            Timestamp:   time.Now(),
            AggregateID: req.ClientIP,
            Version:     1,
            Payload: map[string]interface{}{
                "ClientIP":    req.ClientIP,
                "SNI":         req.SNI,
                "TLSVersion":  req.TLSVersion,
                // ...
            },
        },
        ClientIP: req.ClientIP,
        SNI:      req.SNI,
        // ...
    }
    
    // 异步写入事件存储
    go func() {
        if err := ba.eventStore.Append(context.Background(), event.Event); err != nil {
            log.Printf("error recording event: %v", err)
        }
    }()
    
    // 同时更新内存投影（用于实时查询）
    ba.bufferMu.Lock()
    ba.inMemoryBuffer = append(ba.inMemoryBuffer, &event.Event)
    ba.bufferMu.Unlock()
    
    return nil
}

// GetBehaviorAnalysis 获取行为分析（可选择重放历史数据）
func (ba *BehaviorAnalyzerV2) GetBehaviorAnalysis(ctx context.Context, clientIP string, fromTime time.Time) (*BehaviorState, error) {
    // 选项1：使用内存投影（实时，但可能不完整）
    // return ba.projection.State, nil
    
    // 选项2：重放事件（准确，但较慢）
    projection := &eventsourcing.BehaviorProjection{AggregateID: clientIP}
    if err := projection.ReplayFrom(ctx, fromTime); err != nil {
        return nil, err
    }
    
    return projection.State, nil
}
```plaintext

**算法验证：重放历史数据测试新算法**

```go
// test/algorithm_validation_test.go
func TestNewAnomalyDetectionAlgorithm(t *testing.T) {
    // 从事件存储加载 2 个月的历史数据
    startTime := time.Now().AddDate(0, -2, 0)
    events, _ := eventStore.GetEvents(context.Background(), "192.168.1.100", startTime, time.Now())
    
    // 用新算法重放
    newAnalyzer := behavior.NewAnomalyDetectorV2()
    for _, event := range events {
        if event.Type == eventsourcing.EventTypeRequestReceived {
            req := reconstructRequest(event)
            result := newAnalyzer.Detect(context.Background(), req)
            
            // 对比旧算法的结果
            oldResult := oldAnalyzer.Detect(context.Background(), req)
            
            // 验证一致性或差异可接受
            if diff := compareResults(result, oldResult); diff > threshold {
                t.Logf("algorithm change detected at %v: %v", event.Timestamp, result)
            }
        }
    }
}
```plaintext

#### 📊 收益分析
- ✅ **历史重放**：可用 2 年前的数据测试新算法
- ✅ **调试能力**：追踪具体是哪些请求导致的异常检测
- ✅ **合规性**：完整的审计日志
- ✅ **机器学习**：可训练行为特征模型
- ⚠️ **存储成本**：需要额外的存储（可用压缩和分层存储缓解）

---

### 4️⃣ WASM 支持（WebAssembly）

#### 🎯 现状分析
- **问题**：指纹检测规则（如 JA3 suite order 变化）需要及时更新
- **当前方案**：修改代码 → 编译 → 部署（周期长）

#### 📋 实现方案

**阶段1：规则引擎重构（2-3周）**

```go
// internal/rules/engine.go
package rules

import (
    "github.com/wasmerio/wasmer-go"
)

// RuleEngine WASM 规则引擎
type RuleEngine struct {
    module   *wasmer.Module
    instance *wasmer.Instance
    exports  *wasmer.Exports
}

// LoadWASMRule 加载 WASM 规则模块
func (re *RuleEngine) LoadWASMRule(wasmPath string) error {
    wasmBytes, err := os.ReadFile(wasmPath)
    if err != nil {
        return err
    }
    
    // 编译 WASM 模块
    engine := wasmer.NewEngine()
    module, err := wasmer.NewModule(engine, wasmBytes)
    if err != nil {
        return fmt.Errorf("invalid WASM module: %w", err)
    }
    
    // 创建实例
    store := wasmer.NewStore(engine)
    instance, err := wasmer.NewInstance(module, wasmer.NewImportObject())
    if err != nil {
        return fmt.Errorf("failed to instantiate WASM: %w", err)
    }
    
    re.module = module
    re.instance = instance
    re.exports = instance.Exports
    
    return nil
}

// EvaluateRule 执行规则（调用 WASM 函数）
func (re *RuleEngine) EvaluateRule(input interface{}) (bool, error) {
    // 将输入转换为 WASM 参数（通常是 JSON or protobuf）
    inputJSON, _ := json.Marshal(input)
    
    // 调用 WASM 导出函数
    evaluateFunc := re.exports.GetFunction("evaluate")
    if evaluateFunc == nil {
        return false, fmt.Errorf("WASM module missing 'evaluate' function")
    }
    
    // 执行
    result, err := evaluateFunc(inputJSON)
    if err != nil {
        return false, err
    }
    
    return result.(bool), nil
}
```plaintext

**阶段2：规则定义 DSL（1-2周）**

```python
# rules/ja3_suite_detection.wasm.rs (用 Rust 编写，编译为 WASM)
// 检测 JA3 组件变化规则

#[no_mangle]
pub extern "C" fn evaluate(input_ptr: *const u8, input_len: usize) -> i32 {
    let input = unsafe {
        std::slice::from_raw_parts(input_ptr, input_len)
    };
    
    let ja3_data: JA3Data = serde_json::from_slice(input).unwrap();
    
    // 检测规则：连续 N 个请求的 tls_version 大幅变化
    let tls_version_stability = calculate_stability(&ja3_data.tls_versions);
    if tls_version_stability < 0.7 {
        return 1; // 异常
    }
    
    // 检测规则：cipher suite 出现未知的组合
    if is_unusual_cipher_combination(&ja3_data.cipher_suites) {
        return 1;
    }
    
    0 // 正常
}
```plaintext

**阶段3：规则管理和热加载（2-3周）**

```go
// internal/rules/manager.go
type RuleManager struct {
    rules map[string]*RuleEngine // 规则名 -> WASM 引擎
    mu    sync.RWMutex
}

// DeployRule 部署新规则（可热加载，无需重启）
func (rm *RuleManager) DeployRule(ctx context.Context, ruleName string, wasmPath string) error {
    engine := &RuleEngine{}
    if err := engine.LoadWASMRule(wasmPath); err != nil {
        return err
    }
    
    // 可选：在沙箱环境测试新规则
    if err := rm.ValidateRule(ctx, ruleName, engine); err != nil {
        return fmt.Errorf("rule validation failed: %w", err)
    }
    
    rm.mu.Lock()
    rm.rules[ruleName] = engine
    rm.mu.Unlock()
    
    log.Printf("rule %s deployed successfully", ruleName)
    return nil
}

// ValidateRule 在测试数据集上验证规则
func (rm *RuleManager) ValidateRule(ctx context.Context, ruleName string, engine *RuleEngine) error {
    // 从事件存储加载历史 1 周的数据
    events, _ := eventStore.GetEvents(ctx, "", time.Now().AddDate(0, 0, -7), time.Now())
    
    var falsePositives, falseNegatives int
    for _, event := range events {
        result, _ := engine.EvaluateRule(event.Payload)
        expectedResult := event.Payload["expected_result"].(bool)
        
        if result != expectedResult {
            if result {
                falsePositives++
            } else {
                falseNegatives++
            }
        }
    }
    
    // 检查准确率
    accuracy := 1.0 - float64(falsePositives+falseNegatives)/float64(len(events))
    if accuracy < 0.95 {
        return fmt.Errorf("rule accuracy too low: %.2f%%", accuracy*100)
    }
    
    return nil
}

// EvaluateRules 对输入评估所有规则
func (rm *RuleManager) EvaluateRules(ctx context.Context, input interface{}) (map[string]bool, error) {
    results := make(map[string]bool)
    
    rm.mu.RLock()
    defer rm.mu.RUnlock()
    
    for ruleName, engine := range rm.rules {
        result, err := engine.EvaluateRule(input)
        if err != nil {
            log.Printf("error evaluating rule %s: %v", ruleName, err)
            continue
        }
        results[ruleName] = result
    }
    
    return results, nil
}
```plaintext

#### 📊 对比分析

| 方面 | Go 代码 | WASM 规则 |
| ------ | -------- | --------- |
| **部署周期** | 小时级 | 分钟级 |
| **隔离性** | 与主进程耦合 | 完全隔离 |
| **性能** | 最快 | 5-10% 开销 |
| **规则更新** | 需要重新编译 | 零停机部署 |
| **学习成本** | 低 | 需学习 Rust/WASM |

---

### 5️⃣ 可观测性（Observability）

#### 🎯 现状分析
- **当前 metrics/registry.go**：基础注册表，缺乏实际集成
- **问题**：无法追踪性能瓶颈、无法监控异常

#### 📋 实现方案

**完整的可观测系统三角柱**

```plaintext
┌─────────────────────────────────────┐
│        Application Metrics           │  <- Prometheus
│  (latency, throughput, error rate)   │
├─────────────────────────────────────┤
│      Distributed Tracing             │  <- OpenTelemetry + Jaeger
│  (request flow across components)    │
├─────────────────────────────────────┤
│      Structured Logging              │  <- Zap/logrus
│  (debug, context, request ID)        │
└─────────────────────────────────────┘
```plaintext

**第1步：集成 OpenTelemetry**

```go
// internal/observability/tracer.go
package observability

import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/sdk/trace"
)

// InitTracer 初始化分布式追踪
func InitTracer(serviceName string) error {
    exp, err := jaeger.New(
        jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://localhost:14268/api/traces")),
    )
    if err != nil {
        return err
    }
    
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exp),
    )
    otel.SetTracerProvider(tp)
    
    return nil
}

// GetTracer 获取追踪器
func GetTracer(name string) trace.Tracer {
    return otel.Tracer(name)
}

// Example: 在 ProcessingEngine 中使用
func (e *ProcessingEngine) Process(ctx context.Context, request *ProcessingRequest) *ProcessingResult {
    tracer := GetTracer("fingerprint.processing")
    
    ctx, span := tracer.Start(ctx, "Process")
    defer span.End()
    
    for _, step := range request.Steps {
        _, stepSpan := tracer.Start(ctx, "step."+step)
        
        // ... 执行步骤
        
        // 记录属性
        stepSpan.SetAttributes(
            attribute.String("step.name", step),
            attribute.Int64("step.duration_ms", duration.Milliseconds()),
        )
        stepSpan.End()
    }
    
    return result
}
```plaintext

**第2步：Prometheus 指标**

```go
// internal/observability/metrics.go
package observability

import (
    "github.com/prometheus/client_golang/prometheus"
)

// 定义指标
var (
    // 指纹生成延迟
    fingerprintLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "fingerprint_generation_duration_seconds",
            Help:    "Time spent generating fingerprints",
            Buckets: []float64{.05, .1, .25, .5, 1, 2.5, 5, 10},
        },
        []string{"fingerprint_type"}, // ja3, ja4, ja4s
    )
    
    // 异常检测率
    anomalyDetectionRate = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "anomaly_detection_rate",
            Help: "Percentage of requests flagged as anomalous",
        },
        []string{"anomaly_type"},
    )
    
    // 缓存命中率
    cacheHitRate = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "cache_hit_rate",
            Help: "Cache hit percentage",
        },
        []string{"cache_name"},
    )
    
    // 管道阶段耗时
    pipelineStageLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "pipeline_stage_duration_seconds",
            Help:    "Time spent in each pipeline stage",
            Buckets: []float64{.01, .05, .1, .5, 1},
        },
        []string{"stage_name"},
    )
    
    // 错误率
    errorRate = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "fingerprint_errors_total",
            Help: "Total number of errors",
        },
        []string{"error_type", "component"},
    )
)

func init() {
    prometheus.MustRegister(
        fingerprintLatency,
        anomalyDetectionRate,
        cacheHitRate,
        pipelineStageLatency,
        errorRate,
    )
}

// RecordFingerprint Latency 使用示例
func RecordFingerprintGeneration(ctx context.Context, fpType string, duration time.Duration) {
    fingerprintLatency.WithLabelValues(fpType).Observe(duration.Seconds())
}
```plaintext

**第3步：结构化日志**

```go
// internal/observability/logger.go
package observability

import (
    "context"
    "go.uber.org/zap"
)

// Logger 结构化日志器
var logger *zap.SugaredLogger

func InitLogger() {
    cfg := zap.NewProductionConfig()
    cfg.OutputPaths = []string{"stdout"}
    cfg.ErrorOutputPaths = []string{"stderr"}
    
    l, _ := cfg.Build()
    logger = l.Sugar()
}

// GetLogger 获取上下文绑定的日志器
func GetLogger(ctx context.Context) *zap.SugaredLogger {
    // 如果上下文中有请求ID，自动添加到日志
    if reqID := ctx.Value("request_id"); reqID != nil {
        return logger.With("request_id", reqID)
    }
    return logger
}

// Example: 在业务代码中
func ProcessRequest(ctx context.Context, data []byte) {
    log := GetLogger(ctx)
    
    log.Infow("processing request",
        "source", "127.0.0.1",
        "data_size", len(data),
    )
    
    if err := process(data); err != nil {
        log.Errorw("processing failed",
            "error", err,
            "error_type", reflect.TypeOf(err).String(),
        )
    }
}
```plaintext

**完整示例：整合到 BehaviorAnalyzer**

```go
// internal/security/behavior/analyzer_instrumented.go
type InstrumentedBehaviorAnalyzer struct {
    inner *BehaviorAnalyzer
    
    tracer     trace.Tracer
    logger     *zap.SugaredLogger
    metrics    *AnalysisMetrics
}

type AnalysisMetrics struct {
    analysisLatency prometheus.Histogram
    anomalyCount    prometheus.Counter
    riskScoreHigh   prometheus.Gauge
}

func (iba *InstrumentedBehaviorAnalyzer) Analyze(ctx context.Context, clientIP string) (*AnalysisResult, error) {
    // 追踪
    ctx, span := iba.tracer.Start(ctx, "BehaviorAnalyzer.Analyze")
    defer span.End()
    
    startTime := time.Now()
    log := iba.logger.With("client_ip", clientIP)
    
    log.Infow("starting behavior analysis")
    
    // 执行分析
    result, err := iba.inner.Analyze(ctx, clientIP)
    if err != nil {
        log.Errorw("analysis failed", "error", err)
        return nil, err
    }
    
    // 记录指标
    duration := time.Since(startTime)
    iba.metrics.analysisLatency.Observe(duration.Seconds())
    
    if result.RiskScore > 0.8 {
        iba.metrics.riskScoreHigh.Inc()
    }
    
    if result.AnomaliesDetected > 0 {
        iba.metrics.anomalyCount.Add(float64(result.AnomaliesDetected))
        log.Warnw("anomalies detected", "count", result.AnomaliesDetected)
    }
    
    span.SetAttributes(
        attribute.Float64("risk_score", result.RiskScore),
        attribute.Int("anomalies", int(result.AnomaliesDetected)),
    )
    
    return result, nil
}
```plaintext

---

## 第3部分：改造路线图（12个月）

### Phase 1: 基础可观测性（1-2个月）
- ✅ 集成 OpenTelemetry + Prometheus + 结构化日志
- ✅ 为核心路径添加追踪和指标
- 成本：1-2 个工程师，优先级：**高**

### Phase 2: 流水线模式重构（2-3个月）
- ✅ 设计和实现 Pipeline 框架
- ✅ 迁移现有 ProcessingEngine 步骤
- ✅ 添加中间件支持
- 成本：2-3 个工程师，优先级：**高**

### Phase 3: 事件溯源（3-4个月）
- ✅ WAL 事件存储设计和实现
- ✅ 迁移 BehaviorAnalyzer 到事件驱动
- ✅ 构建历史重放能力
- 成本：2-3 个工程师，优先级：**高**

### Phase 4: 插件化架构（2-3个月）
- ✅ 设计和实现插件框架
- ✅ 迁移现有解析器为插件
- ✅ 热加载和版本管理
- 成本：2-3 个工程师，优先级：**中**

### Phase 5: WASM 规则引擎（可选，2-3个月）
- ✅ Rust 规则引擎开发
- ✅ WASM 编译和加载
- ✅ 规则管理和热部署
- 成本：1-2 个工程师（需 Rust 经验），优先级：**低**

---

## 第4部分：风险和缓解策略

| 风险 | 概率 | 影响 | 缓解策略 |
| ------ | ------ | ------ | --------- |
| **状态迁移复杂性** | 中 | 高 | 写充分的迁移测试，保留旧系统一个版本 |
| **性能下降** | 中 | 中 | 详细的基准测试，可选择关闭追踪 |
| **学习曲线** | 高 | 中 | 提供文档和内部培训，循序渐进的改造 |
| **集成成本** | 中 | 中 | 优先整合到新代码，现有代码逐步迁移 |
| **运维复杂性** | 低 | 高 | 提供 Docker+Kubernetes 配置，自动化部署 |

---

## 第5部分：快速启动（立即可做）

### 现在就可以做的 5 件事：

1. **Prometheus 指标集成**（1 周）
   ```bash
   go get github.com/prometheus/client_golang
   # 为 metrics/registry.go 添加 Expose() 方法
   ```

2. **OpenTelemetry 导入**（2 周）
   ```bash
   go get go.opentelemetry.io/otel
   # 为核心路径添加 span
   ```

3. **结构化日志**（1 周）
   ```bash
   go get go.uber.org/zap
   # 替换 log.Printf 为 zap 日志
   ```

4. **Pipeline 框架 PoC**（2 周）
   ```bash
   # 在 internal/pipeline 创建概念验证
   # 验证依赖解析和中间件链
   ```

5. **Event Sourcing 数据库设计**（1 周）
   ```bash
   # 设计 WAL 表结构
   # 实现基础的 Append 和 Replay 接口
   ```

---

## 附录：参考实现

### 推荐库清单
- **Observability**：OpenTelemetry, Prometheus, Jaeger
- **Pipeline**：自实现（或参考 Uber go-multierr）
- **Event Sourcing**：自实现（或参考 EventStoreDB Go 客户端）
- **Plugin**：Go 原生 plugin 或 HashiCorp go-plugin
- **WASM**：Wasmer Go 或 TinyGo

### 社区项目参考
- CQRS + Event Sourcing：[EventSourcing.NetCore](https://github.com/EventStore/EventSourcing.NetCore)
- Pipeline 中间件：[moby/moby](https://github.com/moby/moby) 的 registry 实现
- 可观测性最佳实践：[Observability Engineering](https://www.oreilly.com/library/view/observability-engineering/9781492076438/)
