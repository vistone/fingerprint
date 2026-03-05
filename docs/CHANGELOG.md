# 变更日志

此项目遵循 [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) 规范。

## [Unreleased]

### Changed

- **Go Workspace 架构重构** (2026-03-05)
  - 将单模块项目重构为 Go Workspace 多模块架构
  - 根目录移除 `go.mod`，使用 `go.work` 管理 14 个模块
  - 模块结构：
    - `modules/core` - 核心类型（零依赖）
    - `modules/profiles` - 指纹配置（187+ 指纹）
    - `modules/tls` - TLS 指纹分析
    - `modules/http` - HTTP 指纹分析
    - `modules/ml` - ML 分类器
    - `modules/defense` - 安全防护
    - `modules/frontend` - 前端 SDK
    - `modules/gateway` - API 网关
    - `modules/generator` - 指纹生成器
    - `modules/network` - 网络层分析
    - `modules/internal` - 内部工具
    - `modules/config` - 配置管理
    - `modules/plugin` - 插件系统
    - `modules/fingerprint` - Facade 统一入口
  - 迁移根目录代码到对应模块：
    - `types/` → `modules/core/types/`
    - `profiles/` → `modules/profiles/legacy/`
    - `http/` → `modules/http/legacy/`
    - `tls/` → `modules/tls/legacy/`
    - `security/` → `modules/defense/legacy/`
    - `generator/` → `modules/generator/`
    - `network/` → `modules/network/`
    - `internal/` → `modules/internal/`
    - `config/` → `modules/config/`
    - `plugin/` → `modules/plugin/`
  - 更新所有导入路径为新的模块路径
  - 文档对齐：
    - 更新 `docs/ARCHITECTURE.md` - Go Workspace 架构说明
    - 更新 `docs/API.md` - 新模块导入路径
    - 更新 `docs/DEVELOPER_GUIDE.md` - 开发环境设置

### Added

- **WebSocket 指纹异常检测** (2026-03-04)
  - 新增 `http/websocket/detector.go` 完整的异常检测系统
  - 支持 12 种异常类型：InvalidHandshake, SuspiciousHeaders, KnownBotUA 等
  - 风险评分算法（0-100 分，5 级风险分类）
  - 修复 `calculateHeaderOrderMatch` 算法 bug（map 不存在键时错误匹配）
  - 测试覆盖：87.8%（17 个测试用例）

- **测试覆盖全面提升 Phase 6** (2026-03-04)
  - **`tls/internal/utils`**: 0% → **100%**
    - GREASE 值处理、类型转换完整测试
  - **`types`**: 0% → **100%**
    - 常量、结构体、HTTPHeaders 方法测试
  - **`internal/config`**: 30.4% → **67.9%**
    - ConfigCenter 加载、更新、回滚、自动重载测试
    - 修复 2 个 bug：RegisterListener 切片长度错误、Rollback 未检查 loaded
  - **`internal/extension`**: 14% → **34.4%**
    - 容器、防御系统、配置加载测试（373 个子用例）
  - **`internal/tcpip`**: 30.2% → **53.2%**
    - NetworkBehaviorAnalyzer 完整测试
  - **`tls/ja4s`**: 35% → **92%**
    - JA4S 指纹分析、异常检测测试
  - **`internal/features`**: 55.6% → **94.9%**
    - 遗留特征适配器完整测试（1129 行）
  - **`tls/ja3`**: 57.3% → **86.6%**
    - 错误处理、便捷函数测试
  - **`config`**: 0% → **100%**
    - 配置中心桥接层测试
  - **`http/clienthints`**: 0% → **94%**
    - Client Hints 生命周期管理、协商策略测试（36 个用例）
  - **`generator`**: 0% → **100%**
    - 生成器错误处理测试

- **性能优化 Phase 7** (2026-03-04)
  - **`security/behavior.AnalyzeTemporalPattern`**: 10.3x 速度提升，98% 内存减少
    - 优化前：47.2 µs/op，51 KB/op，17 allocs/op
    - 优化后：4.6 µs/op，976 B/op，2 allocs/op
  - **`generator/random.GetRandomFingerprintByBrowser`**: 4.2x 速度提升，55% 内存减少
    - 优化前：21.2 µs/op，2 KB/op，26 allocs/op
    - 优化后：5.0 µs/op，903 B/op，15 allocs/op
  - 预计算浏览器类型索引，使用 `sync.Once` 延迟初始化
  - 单次遍历筛选请求并计算间隔，零分配统计计算
  - 新增 19 个基准测试覆盖关键路径

- **安全审计与加固 Phase 8** (2026-03-04)
  - 依赖漏洞扫描：修复 `cloudflare/circl` v1.6.1 → **v1.6.3**
  - 代码安全扫描：`math/rand` 使用符合非加密场景安全实践
  - 敏感信息检测：无硬编码密钥，配置文件使用环境变量
  - 识别 11 个待修复的标准库漏洞（需升级 Go 至 1.25.7+）

- **代码生成工具（中期优先级）** (2026-03-03)
  - 创建 `cmd/profilegen/` 代码生成工具，从 YAML 生成类型安全的 Go 代码
  - 消除 `nolint:composites` 警告，实现零警告代码
  - 支持 17 种 TLS 扩展类型的代码生成
  - 创建示例配置 `profiles/specs/chrome_133.yaml`
  - 完整文档：`cmd/profilegen/README.md`
  - 迁移计划：YAML → 代码生成 → 替换手工代码

- **Prometheus 指标接入（近期优先级）** (2026-03-03)
  - 新增 `internal/metrics/metrics.go`，集成 Prometheus 客户端
  - 指纹生成指标：`fingerprint_generation_total`, `fingerprint_generation_duration_ms`
  - 缓存指标：`fingerprint_profile_cache_hit_total`, `fingerprint_profile_cache_miss_total`
  - 连接指标：`fingerprint_active_connections`
  - 行为分析指标：`fingerprint_behavior_signals_total`
  - HTTP/2 分析指标：`fingerprint_http2_analysis_duration_ms`
  - 在 `generator/random/random.go` 中集成指标采集
  - 在 `security/behavior/analysis.go` 中集成行为信号指标
  - 在 `http/http2/signature.go` 中集成 HTTP/2 分析耗时指标
  - 新增 `internal/metrics/server.go` HTTP 服务器暴露 `/metrics` 端点
  - 新增 Grafana 仪表板配置 `docs/monitoring/grafana-dashboard.json`
  - 文档：`internal/metrics/README.md`, `docs/monitoring/README.md`

- **测试覆盖提升（中期优先级）** (2026-03-03)
  - **profiles 包**: 新增 `profiles/profiles_test.go`，验证 90+ 指纹配置
    - 测试所有 profile 的基本有效性（GetClientHelloStr, GetSettings 等）
    - 按浏览器类型分组测试（Chrome, Firefox, Safari, Edge）
    - 验证配置一致性（settings 与 settingsOrder）
    - 包含 3 个基准测试
  - **security/behavior**: 新增 `security/behavior/analysis_test.go`
    - 覆盖分析器创建、请求添加、时序模式分析
    - 测试信号生成和风险评分计算
    - 包含 2 个基准测试
  - **TCP/IP 模块**: 已完成 `internal/tcpip/analyzer_test.go`
    - 11 个测试函数，覆盖所有公共函数
    - 3 个基准测试

- **配置深拷贝优化** (2026-03-02)
  - 新增 `internal/config/clone.go`，实现所有配置类型的手写深拷贝
  - 替换原有的 JSON 序列化深拷贝方案，性能提升 5-10 倍
  - 新增 `Clone()` 方法到所有配置类型：
    - `ManagedConfig.Clone()`
    - `BehaviorAnalysisConfig.Clone()`
    - `RiskScoringConfig.Clone()` / `RiskWeights.Clone()`
    - `FeatureExtractionConfig.Clone()`
    - `QUICConfig.Clone()`
    - `TLSConfig.Clone()`
    - `GlobalConfig.Clone()`
    - `ConfigMetadata.Clone()`
  - 新增完整测试覆盖 `internal/config/clone_test.go`

### Fixed

- **修复 `calculateHeaderOrderMatch` 算法 bug** (2026-03-04)
  - 原代码访问 map 不存在的键时返回 0，被误认为位置匹配
  - 修复：使用 comma ok 模式检查头部是否存在

- **修复 internal/errors 包** (2026-03-02)
  - 添加缺失的哨兵错误定义：`ErrProfileNotFound`, `ErrInvalidFingerprint`, `ErrClientHelloSpecNotImplemented`
  - 确保与 profiles 包的兼容性

### Changed

- **升级 Go 版本至 1.25.7** (2026-03-04)
  - 修复 11 个标准库安全漏洞
  - GO-2026-4340: crypto/tls 握手消息处理错误
  - GO-2026-4337: TLS 会话恢复问题
  - GO-2025-4175: x509 DNS 名称约束验证错误
  - GO-2025-4155: 证书验证资源消耗问题
  - GO-2025-4013: DSA 公钥验证 panic
  - GO-2025-4012: Cookie 解析内存耗尽
  - GO-2025-4011: ASN1 解析内存耗尽
  - GO-2025-4010: IPv6 URL 验证不足
  - GO-2025-4009: PEM 解析复杂度问题
  - GO-2025-4008: ALPN 协商信息泄露
  - GO-2025-4007: 名称约束检查复杂度
  - 更新 go.mod: `go 1.24.1` → `go 1.25.7`

- **优化 ConfigCenter.copyConfig** (2026-03-02)
  - 从 JSON 序列化改为调用 `config.Clone()`
  - 提高性能，减少内存分配
  - 更好的类型安全

- **升级依赖 `cloudflare/circl`** v1.6.1 → v1.6.3 (2026-03-04)
  - 修复安全漏洞 GO-2026-4550

## [1.0.0] - 2026-03-01

### Added
- 初始版本发布
- TLS 指纹识别（JA3/JA4/JA4S）
- HTTP/2 签名分析
- 浏览器指纹配置管理
- 行为分析模块
