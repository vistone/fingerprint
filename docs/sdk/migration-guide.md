# 迁移指南：从硬编码到模块化架构

## 执行摘要

本指南帮助您将现有的硬编码扩展迁移到新的模块化扩展架构。

**迁移好处**：
- ✅ 支持第三方扩展，无需修改核心代码
- ✅ 更好的代码组织和可维护性
- ✅ 便于测试和性能优化
- ✅ 为社区贡献奠定基础

## 完全向后兼容

新架构**完全向后兼容**现有代码：
- 旧的 `ech_analysis.go` 继续工作
- 新代码可逐步集成
- 支持混合使用新旧组件

## 配置收敛迁移（新增）

从 2026-02-28 起，配置入口已收敛到 `internal/extension`：

- 统一入口：`extension.NewUnifiedConfigFromEnv()`
- 统一规则模型：`extension.RulesConfig`
- 兼容层（仅历史，不建议新代码）：
    - `extension.LoadRulesConfig(path)`
    - `extension.LoadRulesConfigByFilename(filename)`
    - `extension.DefaultRulesConfig()`

### 迁移前后对比

```go
// 迁移前（双轨，仅用于历史对照，不建议新项目复制）
import cfg "github.com/vistone/fingerprint/internal/config"

rules, err := cfg.LoadRulesConfigByFilename("rules.json")

// 迁移后（单入口）
import "github.com/vistone/fingerprint/internal/extension"

appCfg := extension.NewUnifiedConfigFromEnv()
rules := appCfg.Rules
_ = rules
```

### 环境变量

- `FINGERPRINT_ENV`: `development|testing|production`
- `FINGERPRINT_RULES_PATH`: 规则文件路径（优先级最高）
- `FINGERPRINT_RULES_FILE`: 规则文件名（默认 `rules.json`）

### 兼容性说明

`internal/config/loader.go` 已保留为兼容转发层（类型别名 + 函数转发），旧代码可继续运行；
建议新代码直接使用 `internal/extension` 的统一入口，避免再次形成双轨。

## 迁移路线图

```
阶段 1: 准备            2 周
├─ 理解新架构
├─ 设置开发环境
└─ 编写测试

阶段 2: 迁移核心        2 周
├─ 注册扩展元数据
├─ 实现解析器
└─ 实现分析器

阶段 3: 集成处理        1 周
├─ 创建处理器
├─ 配置处理引擎
└─ 集成拦截器

阶段 4: 验证与优化      1 周
├─ 单元测试
├─ 集成测试
└─ 性能优化
```

## 第一步：理解新架构

### 核心概念

| 概念 | 说明 | 责任 |
|------|------|------|
| **注册表** | 管理所有扩展和组件 | 集中管理 |
| **解析器** | 原始字节 → 结构化数据 | 数据转换 |
| **分析器** | 数据分析 → 风险评分 | 逻辑评估 |
| **处理器** | 事件驱动的中间件 | 异步处理 |
| **引擎** | 协调完整流程 | 流程编排 |

### 对比：旧 vs 新

```go
// ===== 旧方式（硬编码） =====
const (
    ExtensionECH = 0xfe0d
)

func analyzeECH(data []byte) {
    if findECH(data) == 0xfe0d {
        // 硬编码的处理逻辑
    }
}

// ===== 新方式（模块化） =====
// 1. 注册
RegisterExtension(&ExtensionMetadata{
    Type: 0xfe0d,
    // ...
})

// 2. 部署
RegisterParserBuilder(0xfe0d, func() (Parser, error) {
    return NewECHParser(), nil
})

// 3. 使用
parser, _ := GetParser(0xfe0d)
result, _ := parser.Parse(data, ctx)
```

## 第二步：准备迁移

### 环境检查

```bash
# 验证 Go 版本
go version  # 要求 1.16+

# 检查依赖
go list -m all | grep fingerprint

# 查看现有扩展
grep -r "0xfe0d\|0xfd00" .
```

### 备份现有代码

```bash
# 创建备份分支
git checkout -b ech-migration-backup
git push origin ech-migration-backup

# 返回主分支
git checkout main
```

## 第三步：迁移 ECH 分析器

### 步骤 1：注册扩展元数据

**位置**：`internal/extension/constants.go`（已提供）

扩展元数据定义了扩展的静态信息：

```go
RegisterExtension(&ExtensionMetadata{
    Type:                  ExtensionEncryptedClientHello,
    Name:                  "Encrypted Client Hello",
    Description:           "Encodes ClientHello to reduce cleartext information",
    RFC:                   "RFC 9180",
    IANANumber:            0xfe0d,
    Category:              CategoryEncryption,
    IsExperimental:        false,
    CompatibleTLSVersions: []uint16{0x0304},
})
```

### 步骤 2：实现解析器

**创建**：`internal/extension/ech_parser.go`（已提供）

```go
type ECHParser struct{}

func (p *ECHParser) Parse(data []byte, parentContext context.Context) (
    ExtensionData, error) {
    // 专注于解析逻辑
    echData := &ECHExtensionData{
        Type:    ExtensionEncryptedClientHello,
        RawData: make([]byte, len(data)),
    }
    copy(echData.RawData, data)
    
    // 解析版本
    if len(data) >= 2 {
        echData.Version = uint16(data[0])<<8 | uint16(data[1])
    }
    
    return echData, nil
}
```

**注册解析器**：

```go
// 在 Init 函数中
RegisterParserBuilder(ExtensionEncryptedClientHello, func() (Parser, error) {
    return NewECHParser(), nil
})
```

### 步骤 3：实现分析器

**创建**：`internal/extension/ech_parser.go`（ECHAnalyzerImpl 部分）

```go
type ECHAnalyzerImpl struct{}

func (a *ECHAnalyzerImpl) Analyze(data ExtensionData,
    config map[string]interface{}) (AnalysisResult, error) {
    
    echData, ok := data.(*ECHExtensionData)
    if !ok {
        return nil, fmt.Errorf("invalid data type")
    }
    
    result := &ECHAnalysisResultImpl{
        ExtType:    ExtensionEncryptedClientHello,
        Anomalies:  []string{},
    }
    
    // 分析逻辑（从旧代码迁移）
    a.detectECHType(echData, result)
    result.RiskScore = a.calculateRiskScore(echData, result)
    a.detectAnomalies(echData, result)
    
    return result, nil
}
```

**注册分析器**：

```go
RegisterAnalyzerBuilder(ExtensionEncryptedClientHello, func() (Analyzer, error) {
    return NewECHAnalyzerImpl(), nil
})
```

### 步骤 4：创建初始化函数

```go
// 在 ech_parser.go 中
func InitializeECHExtension() error {
    // 注册元数据（在 initStandardExtensions 中）
    // 注册解析器
    if err := RegisterParserBuilder(...); err != nil {
        return err
    }
    // 注册分析器
    if err := RegisterAnalyzerBuilder(...); err != nil {
        return err
    }
    return nil
}
```

### 步骤 5：集成到主程序

```go
// 在 fingerprint.go 或 main 中
func init() {
    // 初始化扩展
    if err := extension.InitializeECHExtension(); err != nil {
        log.Fatalf("Failed to init ECH: %v", err)
    }
}
```

## 第四步：更新现有代码

### 更新 ech_analysis.go（可选）

保持现有的 `ECHAnalyzer` 以维持向后兼容性：

```go
// 旧的 ECHAnalyzer 继续存在
type ECHAnalyzer struct {
    // ...（保留现有实现）
}

// 但也支持新的接口
func NewECHAnalyzerCompat() extension.Analyzer {
    return &ECHAnalyzerAdapter{
        legacy: &ECHAnalyzer{},
    }
}

type ECHAnalyzerAdapter struct {
    legacy *ECHAnalyzer
}

func (a *ECHAnalyzerAdapter) Analyze(data extension.ExtensionData,
    config map[string]interface{}) (extension.AnalysisResult, error) {
    // 转换新格式为旧格式
    // 调用旧分析器
    // 转换结果回新格式
}
```

### 更新客户端代码

**旧用法（仍然支持）**：
```go
analyzer := fingerprint.NewECHAnalyzer()
result, _ := analyzer.AnalyzeClientHello(data)
```

**新用法（推荐）**：
```go
// 方法 1：直接使用解析器和分析器
parser, _ := extension.GetParser(extension.ExtensionEncryptedClientHello)
analyzer, _ := extension.GetAnalyzer(extension.ExtensionEncryptedClientHello)

parsedData, _ := parser.Parse(rawData, ctx)
analysisResult, _ := analyzer.Analyze(parsedData, nil)

// 方法 2：使用处理引擎
engine := extension.NewProcessingEngine(nil)
result := engine.Process(&extension.ProcessingRequest{
    ExtensionType: extension.ExtensionEncryptedClientHello,
    RawData:       rawData,
    Steps:         []string{"parse", "analyze"},
})
```

## 第五步：测试迁移

### 单元测试

```go
// tests/ext_ech_test.go
func TestECHExtensionMigration(t *testing.T) {
    // 验证扩展已注册
    metadata, err := extension.GetMetadata(
        extension.ExtensionEncryptedClientHello)
    if err != nil {
        t.Fatalf("ECH not registered: %v", err)
    }
    
    // 验证解析器
    parser, err := extension.GetParser(
        extension.ExtensionEncryptedClientHello)
    if err != nil {
        t.Fatalf("Parser not found: %v", err)
    }
    
    // 验证分析器
    analyzer, err := extension.GetAnalyzer(
        extension.ExtensionEncryptedClientHello)
    if err != nil {
        t.Fatalf("Analyzer not found: %v", err)
    }
    
    // 测试完整流程
    data := []byte{0xfe, 0x0d, 0x01}
    parsed, _ := parser.Parse(data, context.Background())
    result, _ := analyzer.Analyze(parsed, nil)
    
    if result == nil {
        t.Fatal("Analysis result is nil")
    }
}
```

### 集成测试

```go
func TestECHProcessingEngine(t *testing.T) {
    engine := extension.NewProcessingEngine(nil)
    
    request := &extension.ProcessingRequest{
        ExtensionType: extension.ExtensionEncryptedClientHello,
        RawData:       []byte{0xfe, 0x0d, 0x01, 0x02},
        Steps:         []string{"parse", "analyze"},
        Context:       context.Background(),
    }
    
    result := engine.Process(request)
    
    if !result.Success {
        t.Errorf("Processing failed: %s", result.Error)
    }
}
```

### 性能对比

```go
func BenchmarkOldECHAnalyzer(b *testing.B) {
    analyzer := fingerprint.NewECHAnalyzer()
    data := fingerprint.ClientHelloData{
        TLSVersion: 0x0304,
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        analyzer.AnalyzeClientHello(data)
    }
}

func BenchmarkNewECHAnalyzer(b *testing.B) {
    analyzer, _ := extension.GetAnalyzer(
        extension.ExtensionEncryptedClientHello)
    echData := &extension.ECHExtensionData{
        RawData: []byte{0xfe, 0x0d, 0x01},
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        analyzer.Analyze(echData, nil)
    }
}
```

## 第六步：验证和优化

### 检查列表

- [ ] 所有单元测试通过
- [ ] 集成测试通过
- [ ] 性能基准满足要求
- [ ] 代码覆盖率 > 80%
- [ ] 无内存泄漏
- [ ] 文档已更新

### 性能优化

```go
// 启用缓存
config := &extension.EngineConfig{
    EnableCaching: true,
    CacheSize:    1000,
}
engine := extension.NewProcessingEngine(config)

// 启用并发
config.ConcurrentProcessing = true
config.MaxConcurrency = 16
```

### 监控指标

```go
// 收集统计信息
stats := extension.GetRegistryStats()
fmt.Printf("Registered extensions: %d\n", stats["total_extensions"])
fmt.Printf("Registered parsers: %d\n", stats["registered_parsers"])
fmt.Printf("Registered analyzers: %d\n", stats["registered_analyzers"])
```

## 第七步：部署和回滚

### 部署前检查

```bash
# 运行完整测试套件
go test ./...

# 执行代码检查
go vet ./...

# 检查依赖
go mod graph | grep fingerprint

# 构建验证
go build ./...
```

### 灰度部署

```go
// Feature flag for new architecture
const UseNewExtensionArchitecture = false

func AnalyzeECH(...) {
    if UseNewExtensionArchitecture {
        // 使用新架构
        return analyzeECHNew(...)
    } else {
        // 使用旧架构
        return analyzeECHLegacy(...)
    }
}
```

### 回滚计划

```bash
# 如果出现问题
git revert <commit-hash>
git push origin main

# 恢复到备份分支
git checkout ech-migration-backup
```

## 常见问题

### Q1: 新旧代码可以混用吗？
**A**: 是的，完全兼容。可以逐步迁移。

### Q2: 性能会下降吗？
**A**: 新架构性能相同或更优（支持缓存和并发）。

### Q3: 需要修改外部 API 吗？
**A**: 不需要。外部 API 保持不变。

### Q4: 如何处理自定义扩展？
**A**: 遵循开发指南实现新的解析器和分析器。

### Q5: 支持动态加载扩展吗？
**A**: 是的，通过 RegisterExtension 动态注册。

## 下一步

完成迁移后：

1. 📖 阅读 [开发者指南](developer-guide.md)
2. 🔧 实现你的第一个自定义扩展
3. 🤝 对社区项目做出贡献
4. 📢 分享你的想法和建议

## 获取帮助

- 😺 GitHub Issues: https://github.com/vistone/fingerprint/issues
- 💬 Discussions: https://github.com/vistone/fingerprint/discussions
- 📧 Email: contributors@fingerprint.dev

---

**迁移时间线**：平均 **6-8 周**（取决于代码复杂度）

**成功标志**：所有测试通过 ✅ | 零性能下降 ✅ | 文档完整 ✅
