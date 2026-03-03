# 变更日志

此项目遵循 [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) 规范。

## [Unreleased]

### Added

- **TCP/IP 模块测试覆盖** (2026-03-02)
  - 新增 `internal/tcpip/analyzer_test.go`，包含 11 个测试函数
  - 覆盖 OS 数据库构建、签名计算、TTL 分析、窗口大小分析等
  - 包含 3 个基准测试函数
  - 标记 `ExtractTCPOptions` 为待实现（TODO）

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
