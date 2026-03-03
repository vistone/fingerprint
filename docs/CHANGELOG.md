# 变更日志

此项目遵循 [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) 规范。

## [Unreleased]

### Added

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

- **修复 internal/errors 包** (2026-03-02)
  - 添加缺失的哨兵错误定义：`ErrProfileNotFound`, `ErrInvalidFingerprint`, `ErrClientHelloSpecNotImplemented`
  - 确保与 profiles 包的兼容性

### Changed

- **优化 ConfigCenter.copyConfig** (2026-03-02)
  - 从 JSON 序列化改为调用 `config.Clone()`
  - 提高性能，减少内存分配
  - 更好的类型安全

## [1.0.0] - 2026-03-01

### Added
- 初始版本发布
- TLS 指纹识别（JA3/JA4/JA4S）
- HTTP/2 签名分析
- 浏览器指纹配置管理
- 行为分析模块
