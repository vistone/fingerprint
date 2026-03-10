# 变更日志

此项目遵循 [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) 规范。

## [Unreleased]

## [v1.0.7] - 2026-03-10

### Added

- **CONTRIBUTING.md** - 完整的贡献指南和工作流
  - Fork/Clone 和分支设置
  - 开发和测试要求
  - 7 步强制版本控制工作流
  - 代码文数和最佳实践
  - Pull Request 流程

- **SECURITY.md** - 安全政策和漏洞报告流程
  - 安全漏洞报告指南
  - 响应时间预期
  - 用户和贡献者最佳实践
  - 依赖安全指导
  - TLS/密码学相关安全说明
  - 版本支持矩阵

## [v1.0.6] - 2026-03-10

### Added

- **完整的版本控制开发规则** - 确保 Git 规范遵守
  - 新增 `docs/DEVELOPER_GUIDE.md` 版本控制规则部分：强制性的 7 步发布工作流
  - 新增 `COMMIT_CHECKLIST.md`：快速参考卡片供开发者提交前检查
  - 明确定义"不能乱来"的规则：顺序为 CHANGELOG → 版本号增加 → Tag → Push
  - 列出所有 18 个模块的 tag 创建清单
  - 定义严禁的行为及其后果（提交拒绝/回滚）

### Changed

- **README.md 文档链接**：添加版本控制规则快速链接
- **DEVELOPER_GUIDE.md 发布流程**：扩展详细步骤、快速参考脚本、故障排查

### Fixed

- 之前 v1.0.5 版本控制问题已在上一版本完全修正
  - 解决 2 个未标记提交（已纳入 v1.0.5）
  - CHANGELOG 已更新为 [v1.0.5] - 2026-03-10
  - 所有 go.mod 版本已统一为 v1.0.5

## [v1.0.5] - 2026-03-10

### Added

- **多语言国际化支持 (i18n: English/Chinese Language Support)** (2026-03-13)
  - 新增 `modules/gateway/web/static/js/i18n.js`：完整的前端国际化框架（1000+ 行，500+ 翻译键）
  - i18n 模块特性：
    - 双语言字典：English (en) 和简体中文 (zh) 完全对等
    - 核心函数：`t(key)` 翻译、`setLang(lang)` 语言切换、`applyToDOM()` DOM 自动翻译
    - localStorage 持久化：用户语言偏好跨会话保存（`fp-lang` 键）
    - 占位符支持：{n} 参数替换（用于动态内容）
  - HTML 标记完整更新（模板：data-i18n 属性）：
    - 全部 14 个页面导航项、仪表板统计、表格标头、表单标签、按钮、模态框
    - 共 223 个 i18n 属性标记
  - JavaScript 集成（app.js）：
    - 7 个核心函数已转换：`updateRecentClassifications()` / `renderProfileList()` / `updateRequests()` / `renderLogs()` / `startLogStream()` / `updateSystemStatus()` / `saveConfig()`
    - 页面加载时自动调用 `I18N.applyToDOM()` 进行全 DOM 翻译
  - UI 语言切换器：
    - 按钮位置：页面头部右侧（标题栏）
    - 功能：`onclick="I18N.toggle()"` 即时切换中英文
    - 样式：新增 `.lang-switch-btn` CSS 类（含 hover/active 交互效果）
  - 结果：前端从混乱的中英混用状态完全转变为统一可切换的双语界面

### Fixed

- **修复 Profiles 页面 View 弹窗样式缺失** (2026-03-12)
  - 补全 `.modal-overlay` / `.modal` / `.modal-header` / `.modal-body` / `.modal-footer` / `.modal-close` CSS 样式
  - 补全 `.profile-detail-section` / `.profile-detail-grid` / `.profile-detail-item` / `.profile-detail-label` / `.profile-detail-value` / `.profile-detail-tags` CSS 样式
  - 修复 View 按钮点击后详情内容以内联方式显示而非弹窗覆盖层的回归问题

### Added

- **前端深度集成：全模块功能可视化 (Deep Frontend Integration: Full Module Visualization)** (2026-03-11)
  - 新增 `modules/gateway/web/handler_advanced.go`：18 个高级 API 端点，覆盖全部核心模块
  - Gateway 新增 6 个访问器方法：`GetClassifier()`、`GetExtractor()`、`GetRiskEngine()`、`GetSDK()`、`GetInjector()`、`GetProfileManager()`
  - **分析引擎页面 (Analyze Engine)**
    - `POST /api/admin/analyze/profile` — 完整分析管线：ML 分类 → 风险评估 → 威胁检测 → JA3/JA4 指纹 → Agent 决策
  - **ML 引擎页面 (ML Engine)**
    - `GET /api/admin/ml/info` — 三层分层分类器架构（Protocol→Family→Version）、18 种特征类型、置信度阈值
    - `POST /api/admin/ml/extract` — 指定 Profile 特征向量提取
    - `POST /api/admin/ml/classify` — 单 Profile ML 分类（含三层置信度分数）
    - `GET /api/admin/ml/batch` — 全部 Profile 批量分类，统计高置信度比例
  - **防御系统页面 (Defense System)**
    - `GET /api/admin/defense/rules` — 4 条检测规则 + Agent 策略列表 + 风险等级定义
    - `POST /api/admin/defense/detect` — 威胁检测 + 风险评估 + 防御建议
  - **反检测引擎页面 (Anti-Detection Engine)**
    - `GET /api/admin/antidetect/status` — 反检测配置状态 + 6 种 JS 反检测代码生成器 + 可用 Profile 列表
    - `POST /api/admin/antidetect/preview` — 指定 Profile 的反检测 JS 代码预览
    - `POST /api/admin/antidetect/inject` — HTML 注入测试
    - `GET /api/admin/antidetect/sdk` — SDK JavaScript 完整预览
  - **插件系统页面 (Plugin System)**
    - `GET /api/admin/plugins/info` — 插件注册表 + 4 种插件类型（Analyzer/Transformer/Exporter/Validator）+ 三阶段扩展架构（Parser→Analyzer→Handler）+ 注册 API
  - **指纹工具页面 (Fingerprint Tools)**
    - `POST /api/admin/tools/ja3` — JA3/JA4/JA4H 指纹计算器
    - `POST /api/admin/tools/validate` — Profile 完整性校验（结构/Header/TCP-IP）
    - `POST /api/admin/tools/compare` — Profile 相似度对比（字段级 diff + 相似度分数）
  - 前端 SPA 新增 6 个交互式页面，含导航分区（基础功能 / 高级功能）
  - 新增 ~130 行 CSS（page-header/badge/kvtable/risk-bar/code-block/feature-grid/arch-pipeline/plugin-type-card）
  - 新增 15 个 API 客户端方法（api.js）+ ~600 行页面逻辑（app.js）

- **前端全功能集成 (Frontend Full-Feature Integration)** (2026-03-10)
  - 新增 `modules/gateway/web/logbuffer.go`：实时日志捕获与 SSE 推送系统
    - 环形缓冲区（500 条）存储结构化日志条目（时间戳/级别/消息/来源）
    - `InitLogCapture()` 劫持 Go `log` 标准输出，实时写入缓冲区 + stderr
    - SSE 订阅者模式：fan-out 推送新日志到所有连接的浏览器客户端
  - Gateway 新增公开方法：`GetAgent()`、`GetConfig()`（读锁）、`UpdateConfig()`（写锁回调）
  - 新增 4 个 API 端点：
    - `GET /api/admin/agent/status` — Agent 实时状态（启用/观察数/会话数/策略数）
    - `GET /api/admin/agent/knowledge` — 完整知识库浏览（7 大浏览器家族/版本/TLS/H2 细节）
    - `GET /api/admin/agent/strategies` — 活跃策略列表（动作/威胁类型/来源标记）
    - `GET /api/admin/logs/stream` — SSE 实时日志推流
  - 重写 `handleLogs`：从真实 LogBuffer 读取，支持按日志级别过滤
  - 重写 `handleConfig` GET：读取真实 GatewayConfig 全部 7 个配置区（server/rateLimit/cache/ml/antiDetect/scanner/agent）
  - 重写 `handleConfig` POST：通过 `applyConfigUpdate()` + `UpdateConfig()` 线程安全热更新运行时配置
  - 重写 `handleStats`：动态 systemStatus、实时 Agent 统计
  - 前端新增 Agent 页面（🤖）：状态卡片、活跃策略表格、知识库浏览器（按浏览器家族展开版本细节）
  - 前端 Config 页面扩展为 7 个配置卡：Server / Rate Limiting / Cache / ML / Anti-Detection / Scanner / Agent
  - 前端 Logs 页面重写：真实日志展示、级别过滤、SSE 实时推流开关、自动滚动
  - Dashboard System Status 改为动态渲染（API Server / ML / Cache / Anti-Detection / Agent / Scanner）
  - `cmd/gateway/main.go` 启动时写入结构化日志（版本/端口/Anti-Detection/Agent/Scanner 状态）

- **全球指纹特征知识库 (Global Fingerprint Knowledge Base)** (2026-03-10)
  - 新增 `modules/agent/knowledge.go`：编码全球真实浏览器指纹蓝图
    - 7 大浏览器家族（Chrome/Firefox/Safari/Edge/Opera/Brave/Samsung）的 TLS、HTTP/2、TCP/IP 特征基线
    - Chrome 115→134、Firefox 115→135、Safari 16→18 等 15+ 版本的精确密码套件/扩展/曲线数据
    - 5 个 OS 家族（Windows/macOS/Linux/iOS/Android）的 TCP/IP 栈签名（TTL/WindowSize/WindowScale）
    - TLS 1.3 标准套件、GREASE 值全集、各浏览器密码套件/扩展数量合理范围
    - HTTP/2 伪头顺序（Chrome: `:method,:authority,:scheme,:path` vs Firefox: `:method,:path,:authority,:scheme` vs Safari: `:method,:scheme,:path,:authority`）
    - 浏览器市场份额估算数据
  - 新增 `modules/agent/anomaly.go`：知识驱动异常检测器
    - 跨层一致性校验：TLS ↔ HTTP/2 ↔ TCP/IP ↔ JS 特征是否指向同一浏览器身份
    - TLS 密码套件/扩展数量范围校验
    - HTTP/2 InitialWindowSize / MaxConcurrentStreams / 伪头顺序校验
    - TCP/IP TTL 段位校验（128段 vs 64段）、窗口大小校验
    - Headless 浏览器 / 自动化工具标记检测
    - ML 分类置信度校验
    - 矛盾信号加权汇总为 SuspicionScore
  - Agent OADA 循环集成知识校验：`Process()` 现在包含 A1(行为分析) + A2(知识校验) 双分析层
  - `Decision` 新增 `KnowledgeMatch` 字段，包含匹配得分、矛盾列表、可疑度
  - 知识校验高可疑时自动提升威胁等级并标记为 `ThreatFingerprintSpoof`
  - 12 个新增单元测试覆盖知识库初始化、GREASE/密码套件验证、TCP/IP/HTTP2 校验、异常检测、端到端集成

- **自主安全智能体 (Autonomous Security Agent)** (2026-03-10)
  - 新增 `modules/agent` 模块，实现 OADA（Observe→Analyze→Decide→Act）决策循环
  - `BehaviorAnalyzer`：基于滑动窗口的客户端行为画像，追踪指纹切换频率、请求模式、ML 分类一致性、风险趋势
  - `StrategyEngine`：自适应策略引擎，5 条内置策略 + 自演化学习能力，支持自动生成/淘汰检测规则
  - `Memory`：智能体记忆系统，按客户端会话存储观测历史，支持过期清理和容量控制
  - 5 级响应动作：Allow / Monitor / Challenge / Throttle / Block
  - 6 类威胁分类：Bot / FingerprintSpoof / SessionAnomaly / BehavioralAnomaly / Evasion
  - 集成到 Gateway.Analyze() 管线，`AnalyzeResponse` 新增 `agent_decision` 字段
  - `GatewayConfig` 新增 `AgentEnabled` / `AgentConfig` 配置，默认启用
  - Gateway 新增 `Close()` 方法用于优雅关闭 Agent 后台协程
  - 10 个单元测试覆盖核心场景

### Security

- **P0: injector 响应体大小限制** (2026-03-09)
  - `modifyResponse()` 使用 `io.LimitReader` 限制最大读取 10MB，防止恶意超大 HTML 响应导致 OOM

- **P0: ProfileRegistry 并发安全** (2026-03-09)
  - `ProfileRegistry` 添加 `sync.RWMutex`，Register/Get/GetAll/GetByBrowser/GetByOS/Count 全部加锁
  - 修复并发 map 读写导致的 panic

- **P0: tracer JA3Hash 算法修正** (2026-03-09)
  - `calculateJA3Hash` 从 `sha256.Sum256` 改为 `md5.Sum`，符合 JA3 标准规范
  - 与 `tls/ja3.go` 的正确 MD5 实现保持一致

- **P0: RateLimiter goroutine 泄漏修复** (2026-03-09)
  - `cleanup()` 增加 `stopCh` 停止信号和 `defer ticker.Stop()`
  - 新增 `Close()` 方法用于优雅关闭后台 goroutine

- **P0: 网关安全加固** (2026-03-09)
  - `HTTPHandler` 增加 `http.MaxBytesReader` 请求体大小限制，防止 DoS
  - 所有 HTTP handler 错误响应不再泄露内部错误信息，统一使用 `writeJSONError`
  - `getClientIP` 仅在 `RemoteAddr` 匹配 `TrustedProxies` 配置时才信任 `X-Forwarded-For`/`X-Real-IP` 头

### Fixed

- **P1: injector 正则预编译** (2026-03-09)
  - `InjectIntoHTML()` 的 `<head>` / `<html>` 匹配正则提升为包级预编译变量，避免每次请求重复编译

- **P1: ML 置信度计算修正** (2026-03-09)
  - 三层分类器置信度从直接相乘改为加权平均 (0.3/0.3/0.4)，避免指数衰减过度惩罚

- **P1: JS 反检测 configurable 修复** (2026-03-09)
  - `Object.defineProperty` 的 `configurable` 从 `false` 改为 `true`，允许多次注入和运行时更新

- **P1: GetProfile 返回副本** (2026-03-09)
  - `ProfileManager.GetProfile()` 返回浅拷贝，防止外部修改内部状态导致缓存不一致

- **P1: CalculateMD5 实现修正** (2026-03-09)
  - `core.CalculateMD5` 改为使用 `crypto/md5`，之前错误地返回 SHA256 截断值

- **P1: OperatingSystems 随机选择概率修正** (2026-03-09)
  - `OperatingSystems` 切片去除重复值项（Win10/Win11、Linux 发行版 UA 相同），避免随机选择概率偏移
  - 保留别名常量不变，添加注释说明浏览器真实行为

- **P1: 双重错误体系标注** (2026-03-09)
  - `modules/errors` 增加弃用指引注释，新代码应使用 `core.ErrorCode` + `core.CoreError`

### Changed

- **P2: 文档与代码对齐** (2026-03-09)
  - `docs/API.md` 移除不存在的 `CalculateJA3Legacy`、`generator.GenerateRandom`、`network.AnalyzeTCP` 等幻影 API
  - `docs/API.md` 修正 ML、Gateway、Config、ClientHints 模块的 API 签名
  - `README.md` 指纹数量从 "187+" 修正为 "150+"
  - Generator/Network 模块标注为预留接口尚未实现

- **P3: ListProfiles 排序稳定** (2026-03-09)
  - `ProfileManager.ListProfiles()` 返回结果使用 `sort.Strings` 排序，保证顺序稳定

- **P3: CloneProfile ID 冲突检查** (2026-03-09)
  - `ProfileManager.CloneProfile()` 新增 newID 已存在检查，避免静默覆盖

- **代码清理与重构** (2026-03-09)
  - 消除重复代码：提取 `core.RiskLevelFromScore()` 统一风险等级计算，删除 `defense` 包中 2 处重复实现
  - 统一 JA3/JA4 实现：`gateway` 包复用 `tls` 包完整算法，移除简化版实现
  - 优化并发设计：`internal/config` 的 `Load()` 方法将 IO 操作移出锁外，减少锁竞争
  - 明确接口定义：`internal/pipeline` 的 `LoggingMiddleware` 和 `MetricsMiddleware` 使用具体接口替代 `interface{}`
  - 提取公共常量：`core` 包新增时间、大小、阈值等 30+ 个常量，供各模块复用

### Added

- **标准日志接口与适配器** (2026-03-09)
  - 新增 `core.Logger` 接口定义统一日志规范
  - 提供 `slog`、`zap`、`logrus`、标准库 `log` 的适配器实现
  - 新增 `NoOpLogger` 空实现用于测试场景

- **公共常量库** (2026-03-09)
  - 新增 `core/constants.go` 集中管理超时、缓存、限流、风险阈值等常量
  - 时间常量：`DefaultTimeout`, `DefaultDialTimeout`, `DefaultTLSTimeout` 等
  - 大小常量：`MaxRequestBodySize`, `Size1MB`, `Size5MB` 等
  - 风险阈值：`RiskThresholdLow`, `RiskThresholdMedium`, `RiskThresholdHigh`

- **缓存性能测试** (2026-03-09)
  - 新增 `gateway/cache_bench_test.go` 性能测试框架
  - 覆盖 Get/Set/Mixed 场景的基准测试
  - 命中率、淘汰策略、过期、并发安全测试

- **常量验证测试** (2026-03-09)
  - 新增 `core/constants_test.go` 验证常量值和风险计算逻辑
  - 覆盖 `RiskLevelFromScore` 全部分支测试

- **统一错误处理系统** (2026-03-09)
  - `errors` 包扩展错误码体系（SYS/PRF/CFG/NET/PRT/INP/CCH/ML/PLG）
  - 新增 `CodeError` 类型支持错误码、消息、原因和详细信息
  - 提供便捷错误创建函数：`ProfileNotFound`, `InvalidInput`, `Internal`, `Timeout` 等
  - `gateway.ProfileManager` 迁移到新的错误系统

- **Profile 动态管理增强** (2026-03-09)
  - `gateway.ProfileManager` 新增 `ReloadProfile` 单 Profile 重载
  - 新增 `ReloadAll` 全量重载
  - 新增 `GetProfilesByBrowser`/`GetProfilesByOS` 分类查询
  - 新增 `Count` 统计方法
  - 新增 `CloneProfile` Profile 克隆功能

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
