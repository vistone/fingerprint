# 变更日志

此项目遵循 [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) 规范。

## [Unreleased]

## [v1.0.18] - 2026-03-13

### 新增

- **MLService Web API 端点** (`modules/gateway/web/handler.go`、`handler_advanced.go`)
  - 新增 9 个 REST API 端点 `/api/admin/ml/service/`，用于中央 ML 服务管理
  - 端点：stats、health、infer、validate、generate、evolve、train、training-status、feedback
  - 异步训练支持，含状态追踪（阶段、进度、错误报告）
  - MLService 统计信息集成至仪表盘 `/api/admin/stats` 和 `/api/admin/ml/info`

- **前端 MLService 仪表盘** (`modules/gateway/web/static/`)
  - 仪表盘新增 ML Service 统计卡片（推理次数、反馈次数、进化次数、漂移状态）
  - `api.js` 新增 MLService API 客户端方法（全部 9 个端点）
  - `index.html` 新增 ML Service 区域，实时显示统计数据
  - 系统状态新增 MLService 启用/就绪状态

- **GPU 训练 Docker 支持** (`Dockerfile`、`docker-compose.yml`)
  - Dockerfile 切换为 NVIDIA CUDA 12.6 运行时，含 Python 3 和 PyTorch
  - docker-compose.yml 新增 NVIDIA runtime、GPU 环境变量和 models 卷挂载
  - 新增 `training/gpu_train.py` GPU 训练脚本

- **MLService 网关集成** (`cmd/gateway/main.go`、`modules/gateway/gateway.go`)
  - 新增 `FP_ML_SERVICE_ENABLED` 环境变量，运行时切换 ML 服务
  - 网关启动时输出 MLService 启用日志

### 变更

- **代码格式化** (`modules/ml/evolution.go`、`modules/generator/generator.go`、`modules/fingerprint/ml_api.go`、`modules/gateway/gateway.go`)
  - 统一对齐 ML 和网关模块的结构体字段标签

## [v1.0.17] - 2026-03-12

### 新增

- **中央 ML 服务** (`modules/ml/service.go`)
  - `MLService` 单例：项目级 AI 大脑，统一推理、验证、生成与进化
  - `ServiceConfig` 可配置模型存储路径、漂移阈值、伪造/一致性阈值
  - 接口：`Infer()`、`InferFromFeatures()`、`InferBatch()`、`Validate()`、`ValidateFeatures()`
  - `Feedback()` 持续学习、`Generate()` ML 驱动指纹生成
  - `Evolve()` 自动模型进化、`Train()` 基于注册表重训练
  - `Stats()` 公开推理/反馈/进化计数器

- **在线学习系统** (`modules/ml/learner.go`)
  - `OnlineLearner` 环形缓冲反馈样本 + 自动漂移检测
  - `BrowserDistribution` 浏览器分布追踪器，KL 散度计算
  - 通过峰值与近期准确率对比实现自动漂移检测

- **配置进化引擎** (`modules/ml/evolution.go`)
  - `ProfileEvolutionEngine` 逐配置统计：命中次数、伪造率 EMA、置信度 EMA
  - `CheckHealth()` 返回 `EvolutionHealthReport`（过时/伪造/漂移配置）
  - `ShouldEvolve()` 决策逻辑、`TopStaleProfiles()`、`TopForgeryProfiles()`
  - `SnapshotDistribution()` 跨所有配置的浏览器分布快照

- **ML 驱动指纹生成器** (`modules/generator/generator.go`)
  - `SmartGenerator`：ML 验证生成 + 质量评分 + 缓存
  - 质量分 = 40% × (1-伪造概率) + 30% × 一致性 + 30% × 置信度
  - `GenerateBatch()`、`GenerateForBrowser()`、`GenerateForOS()` 便捷方法
  - `RankProfiles()` 配置质量排名、`FindSimilarProfiles()` 嵌入距离搜索
  - `GenerateFeatureVector()`、`GenerateEmbedding()` 原始 ML 数据访问

- **ML 门面 API** (`modules/fingerprint/ml_api.go`)
  - `MLFacade`：统一入口，封装 MLService + SmartGenerator + EvolutionEngine
  - `MLAnalyze()`、`MLAnalyzeWithBehavior()`、`MLAnalyzeBatch()` ML 增强分析
  - `MLGenerate()`、`MLGenerateRandom()`、`MLGenerateBatch()` ML 驱动生成
  - `MLValidate()`、`MLValidateAll()` 注册表级验证
  - `MLFeedback()`、`MLEvolve()`、`MLTrain()`、`MLCheckHealth()` 生命周期管理
  - `MLFindSimilar()`、`MLEmbedding()`、`MLStats()` 高级查询

- **TLS/HTTP/跨层 ML 验证器** (`modules/ml/tls_validator.go`)
  - `TLSValidator`：ML 驱动的 TLS ClientHello 配置验证
  - `HTTPValidator`：ML 驱动的 HTTP 头部/设置验证
  - `CrossLayerValidator`：跨层一致性检查（TLS + HTTP 浏览器一致性）

### 变更

- **网关 ML 集成** (`modules/gateway/gateway.go`)
  - Gateway 结构体新增可选 `MLService` 字段
  - `GatewayConfig` 新增 `MLServiceEnabled` / `MLServiceConfig` 配置项
  - `Analyze()` 启用 MLService 时自动附加 `MLValidation` 结果
  - 伪造检测结果自动追加至 `DefenseHints`
  - 新增 `GetMLService()` 访问器

## [v1.0.16] - 2026-03-11

### 新增

- **国际浏览器指纹配置** (`modules/profiles/international.go`)
  - 23 个新配置: Yandex (4 变体), Vivaldi (3), QQ 浏览器 (2), UC 浏览器 (2), Naver Whale (2), 小米浏览器 (1), 华为浏览器 (1), OPPO 浏览器 (1), Tor 浏览器 (1), DuckDuckGo (1), 360 安全浏览器 (1), 搜狗浏览器 (1), 百度浏览器 (1), Arc (2)
  - Chromium 内核国际浏览器映射至 `BrowserChrome`，Tor 映射至 `BrowserFirefox`，DuckDuckGo/UC iOS 映射至 `BrowserSafari`

- **机器人与自动化工具指纹** (`modules/profiles/bots.go`)
  - 17 个新配置用于伪造检测训练
  - 无头浏览器: Puppeteer Headless (120, 124), Puppeteer Stealth (120)
  - Playwright: Chromium 120, Firefox 121, WebKit 17.4
  - Selenium: ChromeDriver 120, GeckoDriver 121
  - HTTP 客户端: cURL/OpenSSL, Go net/http, Python requests, Node.js Axios, Scrapy
  - 反检测浏览器: Multilogin Mimic 10, GoLogin Chrome 124, Dolphin Anty
  - 无服务器: Cloudflare Worker
  - 每个配置标记 `bot_type`, `tool`, `forgery` 元数据供 ML 标注

- **指纹采集工具** (`cmd/collector/main.go`)
  - 内置知识库: 15 个 JA3 哈希, 6 个 TLS 指纹, 14 个机器人/无头浏览器条目, 20 个国际浏览器条目
  - 采集 55 条指纹至 `training/collected/fingerprints.json`

- **PyTorch GPU 训练流水线** (`training/train_pytorch.py`)
  - 6 个 PyTorch 模型类: EncoderNet, ClassifierNet, ForgeryDetectorNet, ForgeryTypeNet, ThreatAssessorNet, ActionNet
  - 4 阶段训练: 编码器 (triplet loss), 分类器 (交叉熵), 伪造检测器 (BCE+CE), 威胁评估器
  - RTX 2080 SUPER GPU 加速，CUDA 12.4
  - 权重导出为 JSON 供 Go 推理加载

- **Go 推理验证工具** (`cmd/validate/main.go`)
  - 加载 PyTorch 训练权重至 Go ML 流水线
  - 逐配置推理: 浏览器分类、伪造检测、威胁评估
  - 嵌入质量分析（同族 vs 跨族余弦相似度）

- **训练特征导出工具** (`cmd/export_features/main.go`)
  - 将所有注册配置导出为 30 维特征向量供 PyTorch 训练

### 变更

- **扩展训练数据集**: 200 → 240 个配置，覆盖国际浏览器和机器人
- **重训练模型**: 整体准确率 55% → 58.3%，Chrome 50% → 65.3%，Safari 62.5% → 70.6%
- **`BatchNormLayer.Params()`**: 返回 4 个参数（gamma, beta, runMean, runVar）以正确加载 PyTorch 权重

## [v1.0.15] - 2026-03-12

### 新增

- **持久化版本模型库** (`modules/ml/store.go`)
  - `ModelStore` 管理磁盘上的版本化模型快照，自动清理旧版本
  - `StoreConfig` 可配置基础目录和最大版本数（默认 10）
  - `ModelManifest` 索引文件追踪所有版本元数据（父版本、轮次、损失、精度）
  - `Save()` / `Load()` / `LoadLatest()` / `ListVersions()` 完整模型生命周期管理
  - 超过 `MaxVersions` 时自动删除最旧版本
  - 持久化存储：manifest 跨进程重启保留，所有版本可恢复

- **增量模型进化**
  - `Evolve()` 用更低学习率（默认 0.0001）和更少轮次（默认 10）微调现有权重
  - `EvolveConfig` 独立于完整训练配置调整微调行为
  - `EvolveAndSave()` 便捷方法：进化 + 保存新版本一步完成
  - `TrainAndSave()` 便捷方法：初始完整训练 + 保存一步完成
  - 模型从已保存权重持续进化——初始训练后永远不从头重训

- **Pipeline 模型库集成**
  - `LoadFromStore()` / `SaveToStore()` Pipeline 便捷方法
  - 新增 5 个模型库测试

### 变更

- **Agent 自动加载模型**
  - `AgentConfig.ModelStorePath` 指定模型库目录
  - `NewAgent()` 启动时自动加载最新模型快照
  - 模型权重跨进程重启持久化

### 修复

- **所有中文注释转为英文** — 严格遵循 DEVELOPER_GUIDE.md 强制规则
  - 6 个文件约 200 处注释从中文转为英文

## [v1.0.14] - 2026-03-11

### 新增

- **纯 Go 张量计算引擎** (`modules/ml/tensor.go`)
  - `Tensor` 结构体，支持形状感知运算：MatMul、Add、Sub、MulScalar、MulElem、Transpose、SoftmaxRow、SigmoidApply、ReluApply、Normalize、Argmax、Clamp、Clone
  - `ComputeDevice` 接口抽象硬件后端（当前 CPU，未来可通过 CGo 扩展 GPU/CUDA）
  - `cpuDevice` 实现：goroutine 并行 `BatchParallel` 和 `MatMul`
  - 工厂函数：Zeros、Ones、RandN、RandNScaled、FromSlice、NewTensor
  - `SetDevice()` 全局设备切换

- **神经网络层库** (`modules/ml/nn.go`)
  - `Layer` 接口：Forward/Backward/Params/SetTraining
  - `DenseLayer`：He 初始化 + 完整反向传播
  - 激活层：ReLU、Sigmoid、Softmax、Dropout（支持训练/推理模式切换）
  - `Sequential` 模型：可组合的层堆叠
  - `AdamOptimizer`：动量估计 + 偏差校正
  - 损失函数：CrossEntropyLoss、BinaryCrossEntropyLoss、MSELoss、TripletMarginLoss

- **四个领域专用神经网络模型** (`modules/ml/models.go`)
  - `FingerprintEncoder`：30 维 → 128 → 64 → 32 维 L2 归一化嵌入；三元组损失训练浏览器配置文件相似性
  - `BrowserClassifier`：嵌入 → 64 → 7 类 Softmax；识别 Chrome/Firefox/Safari/Edge/Opera/Brave/Samsung 家族
  - `ForgeryDetector`：双头网络（40 维 = 指纹特征 + 跨层一致性特征）；DetectorNet 输出二元伪造概率，TypeNet 输出 4 类伪造类型（真实/无头浏览器/反检测/代理）
  - `ThreatAssessor`：双头网络（45 维 = 嵌入 + 伪造信号 + 行为特征）；ThreatNet 输出 6 类威胁分类，ActionNet 输出 5 类安全动作建议
  - 特征工程：`EncodeFingerprint()` 从 TLS/HTTP2/TCP-IP/JS/行为五层提取 30 维特征；`ComputeCrossLayerFeatures()` 检测跨层不一致性

- **训练管线与模型序列化** (`modules/ml/pipeline.go`)
  - `ModelPipeline`：端到端推理链 Infer()、InferFromFeatures()、InferBatch()
  - `trained` 标志确保仅在训练或权重加载后使用模型输出
  - `NeuralTrainer`：从 ProfileRegistry 进行 4 阶段训练：编码器(三元组) → 分类器(交叉熵) → 伪造检测器(二元交叉熵) → 威胁评估器(交叉熵)
  - 3 倍高斯噪声数据增强、合成伪造样本生成、基于规则的威胁标注
  - `SaveWeights()`/`LoadWeights()`：JSON 格式模型序列化

- **完整 ML 测试套件** (`modules/ml/pipeline_test.go`)
  - 23 个测试：张量运算、神经网络层、全部 4 个领域模型、特征编码、管线推理（单条/批量/从特征）、模型序列化、NeuralTrainer 数据构建、CPU 设备

### 变更

- **Agent 升级为模型编排器** (`modules/agent/agent.go`)
  - Agent 现持有 `ModelPipeline` 用于神经推理，与现有策略引擎和 DQN 并行工作
  - `Process()` 集成模型管线：伪造检测需置信度 > 0.6 才触发升级，威胁评估需动作置信度 > 0.7 才覆盖决策
  - 管线仅在训练/权重加载后激活（`Trained()` 守卫）
  - 新增辅助函数：extractBehaviorVector、threatActionToAgentAction、forgeryTypeName、threatClassName

## [v1.0.13] - 2026-03-11

### 变更

- **强化学习升级：从表格式 Q-learning 升级为深度 Q 网络 (DQN)** (`modules/agent/reinforcement.go`)
  - 用多层感知器 (MLP) 神经网络替代 Q 表，采用 He 初始化和 ReLU 激活函数
  - 实现完整的 SGD 反向传播，梯度裁剪 (±1.0) 保证训练稳定性
  - 添加经验回放缓冲区（可配置容量，默认 10000），均匀小批量采样
  - 添加目标网络，定期权重同步（默认每 100 步），稳定 TD 目标
  - 离散 4 桶状态替换为连续 8 维状态向量：[风险分数, ML置信度, 一致性, 切换率, 请求率, 风险趋势, 观测数, 唯一指纹比率]
  - `RLConfig` 扩展：`HiddenLayers []int`、`LearningRate`、`ReplayCapacity`、`BatchSize`、`TargetUpdateFreq`、`StateDim`
  - 新增连续状态 API：`SelectActionContinuous()`、`UpdateContinuous()`、`QValueContinuous()`、`BestActionContinuous()`
  - `ExtractStateVector()` 直接从观测生成 8 维归一化特征向量（无离散化信息损失）
  - `RLStats` 扩展：`Steps`、`TrainSteps`、`AvgLoss`、`ReplaySize`、`NetworkLayers`
  - 保持向后兼容 API：`SelectAction()`、`Update()`、`QValue()`、`BestAction()` 内部自动将离散 State 转换为向量
- **更新 Agent 集成** (`modules/agent/agent.go`)
  - `Process()` 改用 `ExtractStateVector` + `SelectActionContinuous` + `QValueContinuous`
  - `ReportReward()` 改用 `ExtractStateVector` + `UpdateContinuous`，支持终止状态标记
  - 洞察文本前缀从 "RL" 更新为 "DQN"
- **重写 DQN 测试套件** (`modules/agent/reinforcement_test.go`)
  - 18 个测试函数覆盖：神经网络前向/反向/复制传播、回放缓冲区、状态提取、DQN 贪心/探索/收敛/损失/目标网络同步、向后兼容、统计信息、Agent 集成
  - 对比训练覆盖全部 5 个动作，确保收敛验证稳健性

## [v1.0.12] - 2026-03-11

### 新增

- **Agent 强化学习** (`modules/agent/reinforcement.go`)
  - 表格式 Q-learning，epsilon-greedy 探索策略用于动作选择
  - `RLConfig` 支持 Alpha、Gamma、EpsilonMax/Min/Decay 可调参数
  - 离散化状态空间（威胁 × 风险 × 一致性 × 切换率分桶）
  - `ComputeReward` 函数评分：真阳/假阳/真阴/假阴
  - 集成到 `Agent.Process()` 实现仅升级 RL 覆盖
- **ML 在线学习** (`modules/ml/online.go`)
  - `OnlineClassifier` 增量质心更新（在线均值公式）
  - `WeightedPartialFit` 支持重要性加权样本更新
  - 概念漂移检测：滑动窗口准确率追踪
  - `OnlineHierarchicalClassifier` 三层级联分类：协议→家族→版本
- **Contextual Bandit 策略选择** (`modules/agent/bandit.go`)
  - LinUCB 算法，每个策略臂维护线性模型进行策略优先级排序
  - 8 维上下文向量：观测 + 行为 + ML 信号
  - `BuildContext` 构造归一化特征向量作为 Bandit 输入
  - `RandomBandit` 基线用于 A/B 测试对照
  - 集成到 `Agent.Process()` 在触发策略中优先选择最优策略
- **外部反馈环路**：`Agent.ReportReward()` 方法同时反馈 RL 和 Bandit 子系统
- 三个子系统的单元测试（`reinforcement_test.go`、`bandit_test.go`、`online_test.go`）

## [v1.0.11] - 2026-03-11

### 移除

- **死代码：内部日志模块** (`modules/internal/logger/`) — 独立日志实现，零外部引用
- **死代码：手写指标模块** (`modules/metrics/`) — Counter/Gauge/Histogram/Summary 类型，零外部引用；保留基于 Prometheus 的 `modules/internal/metrics/`

### 变更

- **统一错误体系** — `modules/errors` 现为规范错误包
  - 新增 `CoreError` 类型、`NewCodedError`、`NewCodedErrorf`、`WrapError`、`WrapErrorf` 函数
  - 新增 `VAL`、`NTF`、`SEC` 错误码族和 10 个新哨兵错误
  - 重写 `modules/core/errors.go` 为薄层再导出（类型别名 + 函数转发）
- **修复根目录 Dockerfile** — 修正 HEALTHCHECK 语法，移除已删除 modules/metrics 和不存在 examples/v3 的引用
- **简化根目录 docker-compose.yml** — 面向开发环境，移除 Redis/Nginx 服务，引用 `deploy/docker/` 作为生产配置
- **翻译中文注释为英文** — `modules/plugin/bridge.go` 和 `modules/internal/plugins/`（types.go、registry.go、basic_plugin.go）
- **记录插件架构** — 添加包级文档说明三个插件子系统（internal/plugin、internal/plugins、internal/extension）

## [v1.0.10] - 2026-03-11

### 移除

- **从 Git 跟踪中移除已编译二进制文件**（约 25MB）
  - 移除 `examples/` 下 5 个误提交的 ELF 二进制文件
  - 将二进制路径添加到 `.gitignore` 防止再次跟踪
- **死代码清理**
  - 删除 `modules/kit/strings.go`（6 个标准库包装函数，零外部引用）
  - 移除 `modules/core/utils.go` 中未使用的 `Min()`、`Max()`、`Clamp()`（Go 1.25 已有内置 `min`/`max`）
  - 移除对应的 `TestMinMax` 和 `TestClamp` 测试函数

### 修复

- **重复测试函数** `TestRiskLevelString` 同时声明在 `core/constants_test.go` 和 `core/types_test.go`
- **未定义函数调用** — 死代码移除后 `modules/kit/useragent.go` 中的断裂引用，替换为 `strings.*` 标准库调用

### 变更

- **移除 52 行注释掉的示例代码** — `modules/internal/observability/observability.go`

## [v1.0.9] - 2026-03-11

### 新增

- **API 密钥认证中间件** (`modules/gateway/auth.go`)
  - 通过 `X-API-Key` 头或 `api_key` 查询参数进行恒定时间 API 密钥验证
  - 可配置跳过路径（health/metrics 端点）
  - JSON 错误响应，正确设置 `Content-Type: application/json`
  - 6 个单元测试覆盖所有认证场景

### 安全

- **反向代理 SSRF 防护** (`modules/gateway/injector.go`)
  - 新增 `validateProxyTarget()` 阻止回环地址、链路本地地址和云元数据端点
  - 支持 `AllowPrivateTarget` 配置用于 Docker/Kubernetes 环境
  - 4 个 SSRF 验证测试

### 变更

- **HTML 注入器性能优化** (`modules/gateway/injector.go`)
  - 将每请求 `regexp.MustCompile` + `QuoteMeta` 替换为 `strings.Index` 匹配 `CustomInjectionPoint`
- **翻译所有剩余中文注释为英文** — `modules/gateway/injector.go`

## [v1.0.8] - 2026-03-10

### Changed

- Simplify CONTRIBUTING.md and SECURITY.md documentation

## [v1.0.7] - 2026-03-10

### Added

- CONTRIBUTING.md - 完整的贡献指南和工作流
- SECURITY.md - 安全政策和漏洞报告流程

## [v1.0.6] - 2026-03-10

### Added

- **完整的版本控制开发规则** - 确保 Git 规范遵守
  - 新增 `docs/DEVELOPER_GUIDE.md` 版本控制规则部分：强制性的 7 步发布工作流
  - 明确定义"不能乱来"的规则：顺序为 CHANGELOG → 版本号增加 → Tag → Push
  - 列出所有模块的版本管理要求
  - 定义严禁的行为及其后果（提交拒绝/回滚）

### Changed

- Updated DEVELOPER_GUIDE.md with detailed release process

### Fixed

- Version control consistency across all modules

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
