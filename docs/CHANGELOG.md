# 变更日志

此项目中的所有重要变更都将记录在此文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)，
本项目遵循 [语义化版本](https://semver.org/spec/v2.0.0.html)。

## [未发布]

### 未发布 - 新增

- **模块化重构第二阶段完成（至 JA4S）**：核心 TLS 指纹算法实现已下沉到专用子模块
  - ✅ **QUIC 签名模块**：`network/quic/signature.go` - 完整实现迁移，根包转为兼容转发层
  - ✅ **TCP/IP 指纹模块**：`network/tcp/fingerprint.go` - 完整实现迁移，根包转为兼容别名层
  - ✅ **JA3 TLS 指纹模块**：`tls/ja3/ja3.go` (280 行) - 完整实现迁移，支持 GREASE 过滤和多配置匹配
    - 使用懒初始化 (`ensureJA3Initialized`) 避免循环依赖
    - 根包提供 5 个转发函数：ComputeJA3FromSpec, ComputeJA3FromProfile, ComputeJA3ByProfileName, MatchJA3, FindProfileByJA3
    - ✓ 所有 JA3 相关单元测试通过
  - ✅ **JA4 TLS 签名模块**：`tls/ja4/ja4.go` (400+ 行) - 完整实现迁移，包含 JA4 和 JA4_r 变体
    - 支持 SNI 提取、ALPN 处理、签名算法排序
    - 根包提供类型别名和转发函数
    - ✓ JA4 计算测试通过
  - ✅ **JA4S 服务器指纹模块**：`tls/ja4s/ja4s.go` (387 行) - 完整实现迁移，包含异常检测与分析
    - JA4SAnalyzer 类型、ServerHello 解析、弱密码检测
    - 根包提供 3 个类型别名和转发函数
    - ✓ JA4S 基础功能和弱密码检测测试通过
  - ⏭️ **Random 和 Noise 生成器**：已评估但未迁移（循环依赖复杂度高）
    - 这两个模块有深层根包依赖（MappedTLSClients, UserAgent 生成，Headers 生成等）
    - 现有架构下迁移会引入循环导入问题，需要更深层的重构，推迟至后续专项优化
  
  迁移模式说明：
  - **完整实现迁移**：将完整代码复制到子模块，仅改变 package 声明
  - **类型兼容层**：根包使用 `type Alias = submodule.Type` 的类型别名（JA4S）或结构体转发（JA3/JA4）
  - **懒初始化**：通过 sync.Once 在首次使用时初始化跨包依赖，避免 init() 时的循环依赖问题
  - **向后兼容**：所有根包 API 保持不变，现有代码无需修改，自动使用新子模块实现
  
- **模块化重构（安全落地第一阶段）**：新增模块目录骨架与 HTTP/TLS 包装层（不迁移旧实现）
  - 新增模块目录：`tls/`、`http/`、`network/`、`security/`、`generator/`、`config/`、`plugin/`、`internal/errors/`
  - 注：HTTP headers/useragent 模块暂保留为桥接层，因现有根包结构复杂且 ClientProfile、UserAgentGenerator 等已有类型分布于多个文件（profiles.go、useragent_helper.go），迁移成本高且风险大，待后续专项重构
  - 兼容性说明：根包现有 API 保持不变，新子包通过桥接调用现有实现

- **监控和可观测性系统**：完整的 Prometheus/OpenTelemetry 兼容监控框架
  - **内部模块**：
    - metrics：指标收集和导出（Counter、Gauge、Histogram）
    - tracing：分布式追踪（TraceID、Span、标签系统）
    - monitor：健康检查和自愈机制
  - **指标类型**：
    - Counter：累计计数器（请求数、错误数）
    - Gauge：实时值仪表（CPU、内存、连接数）
    - Histogram：分布统计（延迟、响应时间）
  - **性能指标**：
    - PerformanceMetrics：QPS、错误率、延迟分布（P50/P95/P99）
    - 资源监控：CPU 使用率、内存使用量、活跃连接数
    - 缓存统计：缓存命中、数据包处理计数
  - **分布式追踪**：
    - OpenTelemetry 风格的 Trace 和 Span
    - 自动时间测量和标签支持
    - 完整的追踪链关联和查询
  - **健康检查**：
    - 多检查器注册和执行
    - 三级状态：HEALTHY / DEGRADED / UNHEALTHY
    - 自动故障检测和恢复机制
  - **Prometheus 导出**：
    - 标准 Prometheus 文本格式
    - 全局指标注册表
    - `/metrics` 端点集成支持
  - **示例和文档**：
    - 完整的监控演示程序（examples/monitoring/main.go）
    - 详细的使用文档和最佳实践（docs/monitoring/README.md）
    - HTTP 中间件集成示例

- **TCP/IP 层指纹识别系统**：完整的 TCP/IP 网络层特征分析
  - NetworkBehaviorAnalyzer：RTT 分析、序列号模式检测、IP ID 分析、时间模式识别
  - DeviceFingerprintingEngine：设备轮廓注册、操作系统匹配、风险评估
  - TCPIPAnalyzer：数据包分析器，支持添加和流式处理
  - PacketParser：数据包解析工具，包括校验和计算和签名生成
  - OS 数据库：包含 10+ 操作系统特征（Windows、Linux、macOS、iOS、Android）
  - 网络行为分析：RTT 统计（最小/最大/平均）、网络类型分类(本地/国内/区域/国际)
  - 设备指纹识别：支持移动设备、台式机、服务器、IoT、路由器等设备类型
  - 异常检测：非标准 TTL、异常 MSS、小窗口、可疑 TCP 选项识别
  - VPN/代理/NAT 检测框架：信号库和检测规则
  - 风险指标计算：机器人、扫描器、欺骗检测的综合风险评分
- **高级 TCP/IP 工具库**：constants.go、packet_parser.go、behavior_analyzer.go
  - TCP 常量定义：标志位、选项类型、TTL 推荐值、IP 标志位
  - 网络设备签名：FritzBox、Cisco、Juniper、HP、QNAP 等家族设备特征
  - 数据包解析：IP 头解析、TCP 头解析、标志位提取、检验和计算
  - 签名格式化：TTL、MSS、窗口大小和 TCP 选项的规范化
  - 私有 IP 检测：IsPrivateIP() 和 IsReservedIP() 验证
  - 初始 TTL 推断：基于路由跳转数推断原始 TTL 值
- **完整的演示示例**：4 个实际运行的示例展示新功能
  - 网络行为分析演示：RTT 测量、序列号模式、时间分布、行为特征
  - 设备指纹识别引擎演示：多设备轮廓匹配、OS 识别、风险评估
  - 数据包解析演示：数据包详情、签名生成、IP 验证、初始 TTL 推断
  - 风险评估演示：多维度威胁检测、风险评分、威胁等级分类

### 未发布 - 新增（之前的内容）

- **统一错误处理系统**：46 个错误代码，8 个类别，5 级严重性分类
  - ErrorCode 枚举：注册表、验证、解析、分析、配置、插件、系统、安全错误码
  - Error 结构体：支持错误链、上下文键值对、时间戳记录
  - ErrorSeverity 分类：Info、Warning、Error、Critical、Fatal
  - 错误恢复方法：IsRecoverable()、IsFatal() 自动判断
- **输入验证框架**：多层验证机制防止恶意输入
  - DefaultValidator：数据、元数据、配置全面验证
  - InputSanitizer：字符串和字节数据清理
  - ComprehensiveValidation：完整验证流程
- **防御和资源保护系统**：生产环境资源限制和监控
  - RateLimiter：令牌桶算法，防止请求溅射
  - ResourceMonitor：内存分配追踪、超时检查、并发监控
  - SecurityAuditor：安全事件记录和审计日志
  - RequestGuard：集成防御守卫
  - DefensePolicy：可配置防御策略预设
- **日志系统**：SimpleLogger 提供 5 级日志（Debug、Info、Warn、Error、Fatal）
- **错误恢复机制**：SafeExecute 和 SafeExecuteWithRecovery 实现 panic 捕获和恢复
- **综合风险评分系统**：8 维度风险评估与威胁等级判定
- **风险评分引擎**：可配置权重、阈值、严格模式、置信度计算
- **8 维度风险评估**：TLS 指纹、服务端行为、HTTP/2 签名、HTTP 请求头、QUIC 签名、Client Hints 不一致性、ECH 影响、行为异常
- **威胁等级分类**：safe (≤0.2)、low (≤0.4)、medium (≤0.6)、high (≤0.8)、critical (>0.8)
- **上下文感知评分**：IP 信誉、地理位置风险、历史行为、已知客户端识别
- **自动建议生成**：基于威胁等级的分层防御策略建议
- **风险因素详情**：类型、严重程度、描述、证据、权重、置信度
- ECH 分析：Encrypted Client Hello 检测与影响评估
- ECH 类型识别：GREASE、Outer、Inner ClientHello 类型检测
- ECH 可见字段签名：当 SNI 被加密时的替代指纹方法
- ECH 异常检测：配置错误、协议异常、不完整的 ClientHello 结构
- ECH 影响评估：SNI 可见性、受影响方法、可用替代方法
- ECH 替代策略建议：基于影响等级的多层防御策略建议
- JA4S 指纹分析：TLS ServerHello 指纹计算与异常检测
- HTTP/2 帧签名分析：帧序列特征识别与客户端匹配
- JA4H 指纹分析：HTTP 请求头特征与浏览器行为识别
- QUIC 签名分析：QUIC Initial 包指纹、传输参数、帧序列特征识别
- HTTP/3 支持：基于 QUIC v1/v2 的 HTTP/3 流量识别
- QUIC 异常检测：草稿版本、缺失 CRYPTO 帧、可疑限制值、无效连接 ID
- Client Hints 完整支持：低熵/高熵提示、Accept-CH 协商、跨源委托、Permissions-Policy
- JA4H Client Hints 不一致性检测：UA 与 CH 平台匹配、移动设备识别、架构位宽验证
- ClientProfile 元数据字段：浏览器类型/版本、操作系统/架构/位宽、移动设备型号
- 统一特征提取框架：可扩展的特征提取接口
- JSON 规则配置引擎：阈值与规则完全可配置化
- 兼容适配层：现有代码零成本迁移
- 文档管理规范：明确禁止创建报告/总结文档，统一使用 CHANGELOG
- 完整的 Go 开发标准和风格指南
- Go Vet composites 警告处理
- GitHub Actions CI/CD 工作流
- Makefile 开发工具
- 增强的文档系统

### 未发布 - 变更

- **文档治理收敛**：将历史“总结/报告”类文档标注为归档状态，仅用于历史对照
  - `PROJECT_SUMMARY.md`
  - `docs/COMPLETION_REPORT.md`
  - `docs/2-guides/developer/02-error-fixes-summary.md`
  - `docs/sdk/IMPLEMENTATION_SUMMARY.md`
  - 新增统一指引：后续变更统一记录到 `docs/CHANGELOG.md`

- **架构改进：依赖注入模式** (`internal/extension/container.go`)
  - 集中式依赖注入容器管理组件生命周期
  - 支持工厂模式注册自定义组件
  - 单例缓存提高性能
  - 线程安全的组件访问 (sync.RWMutex)
  - 类型安全的访问器方法 (GetLogger, GetValidator, GetRequestGuard 等)
  - Initialize() 支持关键组件预加载
  - Reset() 支持测试环境重置

- **架构改进：集中式配置管理** (`internal/extension/config.go`)
  - 统一的 Config 结构体聚合所有模块配置
  - 三种环境预设 (Development、Testing、Production) 
  - Environment 环境区分，每个环境有不同的严格程度
  - 环境变量覆盖支持 (FINGERPRINT_* 前缀)
  - LoggerConfig、ValidatorConfig、AuditConfig、DefensePolicy 统一管理
  - Config.Validate() 在使用前检查配置有效性
  - NewConfig(env) 快速创建环境特定配置
  - NewConfigFromEnv() 从环境变量完整初始化

- **架构改进：小而精的接口设计** (`internal/extension/component.go`)  
  - 遵循 Go 隐式接口设计和单一职责原则
  - Component: 所有可治理组件的基础接口
  - Configurable: 运行时配置支持
  - Auditable: 审计事件记录
  - Closeable: 兼容 io.Closer 的资源清理
  - Initializable: 延迟初始化支持
  - Identifiable: 唯一标识符
  - 接口设计遵循 Go 最佳实践，避免"胖接口"
  - 通过接口组合实现复杂功能

- 改进的项目结构和组织
- 增强的文档和 API 注释
- 最新的 Go 依赖
- **代码规范性改进**：提取重复代码，统一错误处理模式
  - RequestGuard.logAndReturnError()：统一处理审计日志记录和错误返回
  - 详细的包级和类型级注释：所有公开接口都有完整的使用说明
  - 一致的错误包装：使用 NewError、NewErrorWithCause 统一错误创建
  - 遵守 golangci-lint 配置：所有代码通过 go vet 检查

### 未发布 - 修复

- 修复 Go Vet composites 警告
- 代码格式化标准化

---

## [2.0.1] - 2026-02-28

### 2.0.1 - 新增

- Edge 浏览器指纹配置
- JA3/JA4 指纹计算方法
- 被动浏览器识别
- 异常检测
- 噪声注入功能
- 完整的错误修复和文档

### 2.0.1 - 变更

- 改进的代码质量和标准
- 增强的项目文档
- 改进的错误处理

### 2.0.1 - 修复

- 解决 Go Vet 警告
- 代码格式化修复

---

## [2.0.0] - 2026-02-01

### 2.0.0 - 新增

- 完整的浏览器 TLS 指纹库
- 支持 70+ 浏览器指纹配置
- User-Agent 生成
- HTTP Headers 生成
- JA3 指纹计算和匹配
- 被动浏览器识别
- 异常检测
- 矛盾检测
- 综合测试套件（90%+ 覆盖率）

### 2.0.0 - 变更

- 指纹匹配引擎主要重写
- 零分配函数优化性能
- 改进的 API 设计

### 2.0.0 - 修复

- 指纹验证的安全问题
- 匹配算法的性能瓶颈

---

## [1.0.2] - 2025-12-15

### 1.0.2 - 新增

- Safari 浏览器指纹支持
- 边界情况的额外测试
- 指纹匹配的性能优化

### 1.0.2 - 修复

- 移动设备 User-Agent 解析错误
- 指纹缓存内存泄漏
- Go 1.24 兼容性问题

---

## [1.0.1] - 2025-11-20

### 1.0.1 - 新增

- Firefox 浏览器指纹支持
- Opera 浏览器支持
- 改进的文档

### 1.0.1 - 修复

- Chrome User-Agent 生成问题
- 请求头排序问题
- TLS 版本协商修复

---

## [1.0.0] - 2025-10-01

### 1.0.0 - 新增

- 指纹库初始版本
- 核心 TLS 指纹匹配
- User-Agent 生成
- HTTP Headers 生成
- Chrome 和 Edge 浏览器支持
- 基础指纹数据库

### 1.0.0 - 特性

- 零分配随机选择
- 高性能指纹匹配
- 线程安全的并发访问
- 完整的错误处理
- 完整的测试覆盖

---

## 支持

- 📖 [文档](./README.md)
- 🐛 [问题追踪](https://github.com/vistone/fingerprint/issues)
- 💬 [讨论](https://github.com/vistone/fingerprint/discussions)
- 📧 [邮件支持](mailto:support@example.com)

---

## 版本控制

本项目遵循 [语义化版本](https://semver.org/)：

- **主版本号** - 不兼容的 API 变更
- **次版本号** - 兼容的功能增加
- **修订号** - 兼容的错误修复

---

## 链接

- [GitHub 仓库](https://github.com/vistone/fingerprint)
- [包文档](https://pkg.go.dev/github.com/vistone/fingerprint)
- [贡献指南](./CONTRIBUTING.md)
- [安全政策](./SECURITY.md)
- [许可证](../LICENSE)
