# Changelog

This project follows the [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) specification.

## [Unreleased]

## [v1.0.23] - 2026-03-14

### Added

- **Closed-loop learning architecture**: Gateway integrates Crawler and WAF into an adversarial training loop
  - `modules/ml/feedback_types.go`: Shared `CrawlerFeedback` and `WAFDetectionFeedback` types for cross-module ML communication
  - `modules/crawler/crawler_ml.go`: `CrawlerMLAdapter` — bridges crawler to ML service with UCB1-based adaptive profile selection and online learning feedback
  - `modules/waf/waf_learning.go`: `LearningPipeline` — bridges WAF detection results to ML for continuous model improvement
  - `modules/gateway/gateway_closedloop.go`: `ClosedLoopController` — orchestrates Crawler → ML ← WAF adversarial training cycles

- **Crawler ML integration**
  - Adaptive profile selection using UCB1 (Upper Confidence Bound) exploration/exploitation balance
  - Crawl results automatically fed back to ML `OnlineLearner` for drift detection and model evolution
  - `SetMLService()` on `Crawler` to inject ML service from gateway

- **WAF ML inference and learning**
  - Filled ML inference placeholder in `waf_analyze.go` — WAF now runs ML forgery detection during request analysis
  - WAF detection results (risk scores, detection layers, block decisions) fed to ML for online learning
  - Learning pipeline statistics tracking

- **Gateway orchestration**
  - `ClosedLoopConfig` and `ClosedLoopController` for adversarial training cycle management
  - `SetCrawler()` / `SetWAF()` on `Gateway` to inject subsystems into the closed loop
  - Periodic adversarial training: generate fingerprints → validate against detection → feed back to ML → model evolution

### Fixed

- **modules/waf**: Fix `Block()` call signature — was passing `time.Duration` where `string` reason was expected
- **modules/crawler**: Fix `initProfilePool()` to use `profiles.GetAll()` correctly (value vs pointer types)
- **modules/crawler**: Fix `BrowserType` field reference (was `Browser`)

## [v1.0.22] - 2026-03-14

### Fixed

- **modules/client**: Add missing `go.mod` file, fixing full workspace build failure
- **go.work**: Add `modules/client` to workspace

### Refactored

- **File splits (over 500 lines)**
  - `scripts/verify_fingerprint_packet.go` (558→330 lines) → extracted `verify_fingerprint_report.go` (248 lines)
  - `modules/gateway/gateway_scanner.go` (505→327 lines) → extracted `gateway_fetch.go` (197 lines)

- **Function splits (over 80 lines)**
  - `modules/frontend/sdk.go` `GenerateJSCore()` (360 lines) → extracted into `sdk_js_core.go` with 8 sub-functions
  - `modules/gateway/gateway_analyze.go` `Analyze()` (209→75 lines) → extracted `analyzeNetworkLayer`, `runPluginPipeline`, `enrichWithMLValidation`
  - `modules/client/client_proxy.go` `ExecuteProxyRequest()` (209→66 lines) → extracted into `client_proxy_helpers.go` with `normalizeProxyInput`, `newProxyResult`, `buildHTTPRequest`, `retryTransientErrors`, `buildProxyResponse`

- **Parameter optimization (over 5 parameters)**
  - `modules/ml/pipeline_training_detectors.go` `recordMetric()` 7→1 param, reusing `TrainingMetrics` struct
  - `modules/internal/security/audit.go` `Log()` 6→3 params, `LogWithContext()` 7→4 params, introduced `AuditLogEntry` struct
  - `modules/profiles/legacy/profiles.go` `NewClientProfile()` 7→1 param, introduced `ClientProfileParams` struct
  - `modules/gateway/headless_fetcher.go` `fetchHTMLWithHeadlessBrowser()` 6→2 params, introduced `headlessFetchOptions` struct
  - `modules/gateway/headless_fetcher.go` `fetchScriptBody()` 6→2 params, introduced `scriptFetchParams` struct

## [v1.0.21] - 2026-03-14

### Added

- **Active Crawler Module** (`modules/crawler/`)
  - New active crawler for testing anti-crawling capabilities and training detection models
  - Profile rotation with 4 strategies: Random, Rotate, Sticky, Adaptive
  - Smart proxy pool management with health checking
  - Stealth mode with human behavior simulation (mouse paths, scroll patterns, typing)
  - Auto-detection of blocking responses (403, CAPTCHA, rate limiting)
  - Data feedback loop for ML training with batch collection
  - Comprehensive metrics collection for performance monitoring
  - Split into 5 focused files per 500-line policy:
    - `crawler_config.go`: Configuration types and constants
    - `crawler_core.go`: Core Crawler struct and lifecycle methods
    - `crawler_profile.go`: Profile pool management and rotation
    - `crawler_proxy.go`: Proxy pool management
    - `crawler_worker.go`: Worker execution and request processing

- **WAF Module** (`modules/waf/`)
  - 5-layer protection architecture:
    - L1 Network: IP reputation, geo-detection, ASN analysis
    - L2 TLS: JA3/JA4 fingerprint blacklisting, TLS anomaly detection
    - L3 HTTP: Header analysis, User-Agent validation, cookie anomaly detection
    - L4 Behavior: Request rate analysis, path traversal detection, session anomaly
    - L5 Device: Device fingerprinting, consistency checks
  - Token bucket rate limiter with burst support
  - Configurable IP/Path/CIDR blocklists and whitelists
  - Middleware for HTTP handlers with risk scoring
  - 3 operation modes: Protection, Monitoring, Learning
  - Integration with ML module for threat classification
  - Integration with Agent module for autonomous decisions

- **Integration Examples** (`examples/`)
  - `crawler/`: Active crawler usage examples
  - `waf/`: WAF middleware integration examples
  - `crawler-waf-integration/`: Complete testing loop demonstration

- **Deployment Configurations**
  - `Dockerfile.crawler`: Container image for crawler service
  - `Dockerfile.waf`: Container image for WAF service
  - `docker-compose.crawler-waf.yml`: Full stack with Redis, Prometheus, Grafana
  - `deploy/kubernetes/crawler-deployment.yaml`: K8s deployment with HPA
  - `deploy/kubernetes/waf-deployment.yaml`: K8s deployment with ConfigMap and HPA

- **Documentation** (`docs/CRAWLER_INTEGRATION.md`)
  - Architecture overview with data flow diagram
  - Component descriptions for Crawler, WAF, Agent, and ML modules
  - Quick start guides for different use cases
  - Configuration reference and best practices

### Changed

- **Code Quality**: All crawler and WAF module comments converted to English per development guidelines
- **Example Code**: All example files updated with English comments only

## [v1.0.20] - 2026-03-14

### Changed

- **Code Structure Refactor**: Split 34 oversized source files (>500 lines) across 15 modules into smaller, focused files
  - profiles/legacy: 7 files → 17 files
  - profiles: 3 files → 6 files
  - ml: 4 files → 11 files
  - gateway: 4 files → 15 files
  - agent: 2 files → 4 files
  - core, client, generator, frontend, network/tcp: each split into 2 files
  - internal/extension, internal/tcpip, internal/features: split into smaller files
  - http/legacy/ja4h, tls/legacy/ja4s, http/legacy/websocket: each split into 2 files
  - All files now comply with 500-line maximum policy
  - Zero API changes, all tests and race detection pass

### Added

- **File Length CI Check**: New `file-length-check` job in CI workflow (`.github/workflows/lint.yml`)
- **Copilot Instructions**: `.github/copilot-instructions.md` with file length coding rules
- **Linter Config**: `.golangci.yml` funlen configuration for function length enforcement
- **Contributing Docs**: Updated `CONTRIBUTING.md` and `CONTRIBUTING_zh-cn.md` with file length policy

## [v1.0.19] - 2026-03-13

### Added

- **Network Module Integration** (`modules/gateway/gateway.go`)
  - Integrated JA4T transport fingerprinting (ja4t package) into Gateway Analyze pipeline
  - Integrated TCP/IP stream analysis (tcp package) with OS detection, VPN/proxy/NAT identification
  - New `TCPRequestData` and `TCPPackets` fields in AnalyzeRequest for network-layer input
  - New `JA4TInfo` and `NetworkAnalysisResult` in AnalyzeResponse for transport-layer results
  - TCP features (window size, MSS, TTL, DF) merged into FeatureVector for ML classification

- **Plugin Pipeline Integration** (`modules/gateway/gateway.go`)
  - Plugin Manager initialized in NewGateway with optional config-based loading
  - Plugin analyzers and validators executed in Analyze pipeline, results in PluginResults field
  - New `GetPluginManager()` accessor on Gateway
  - New `PluginConfigPath` configuration field with `FP_PLUGIN_CONFIG_PATH` env var

- **Risk Threshold Enforcement** (`modules/gateway/gateway.go`)
  - `RiskBlocked` field in AnalyzeResponse, set when risk score exceeds configured threshold
  - `RiskThreshold` config now actively evaluated during Analyze

- **ML Drift Auto-Evolution** (`modules/ml/service.go`)
  - `Feedback()` now checks `OnlineLearner.DriftDetected()` and triggers auto `Evolve()` on drift
  - New `Learner()` accessor on MLService for external drift monitoring
  - `MLClassifierPath` config wired into MLService model store path

- **Environment Variable Wiring** (`cmd/gateway/main.go`)
  - `FP_ML_CLASSIFIER_PATH`: ML classifier model path
  - `FP_ML_TRAINING_DATA`: ML training data path
  - `FP_RISK_THRESHOLD`: Risk score threshold (float)
  - `FP_TRUSTED_PROXIES`: Comma-separated trusted proxy IP list
  - `FP_PLUGIN_CONFIG_PATH`: Plugin configuration file path
  - Startup logging for plugin, risk threshold configuration

### Changed

- **Gateway go.mod** (`modules/gateway/go.mod`)
  - Added dependency on `modules/network` for ja4t and tcp packages

## [v1.0.18] - 2026-03-13

### Added

- **MLService Web API Endpoints** (`modules/gateway/web/handler.go`, `handler_advanced.go`)
  - 9 new REST API endpoints under `/api/admin/ml/service/` for central ML service management
  - Endpoints: stats, health, infer, validate, generate, evolve, train, training-status, feedback
  - Async training support with status tracking (phase, progress, error reporting)
  - MLService stats integration into dashboard `/api/admin/stats` and `/api/admin/ml/info`

- **Frontend MLService Dashboard** (`modules/gateway/web/static/`)
  - ML Service stats cards on dashboard (infer count, feedback count, evolution count, drift status)
  - MLService API client methods in `api.js` for all 9 new endpoints
  - ML Service section in `index.html` with real-time stats display
  - System status now shows MLService enabled/ready state

- **GPU Training Docker Support** (`Dockerfile`, `docker-compose.yml`)
  - Dockerfile switched to NVIDIA CUDA 12.6 runtime with Python 3 and PyTorch
  - docker-compose.yml adds NVIDIA runtime, GPU environment variables, and models volume mount
  - New `training/gpu_train.py` GPU training script

- **MLService Gateway Integration** (`cmd/gateway/main.go`, `modules/gateway/gateway.go`)
  - `FP_ML_SERVICE_ENABLED` environment variable for runtime ML service toggle
  - MLService enabled logging on gateway startup

### Changed

- **Code Formatting** (`modules/ml/evolution.go`, `modules/generator/generator.go`, `modules/fingerprint/ml_api.go`, `modules/gateway/gateway.go`)
  - Aligned struct field tags for consistent Go formatting across ML and gateway modules

## [v1.0.17] - 2026-03-12

### Added

- **Central ML Service** (`modules/ml/service.go`)
  - `MLService` singleton: project-wide AI brain for inference, validation, generation and evolution
  - `ServiceConfig` with model store path, drift threshold, forgery/consistency thresholds
  - APIs: `Infer()`, `InferFromFeatures()`, `InferBatch()`, `Validate()`, `ValidateFeatures()`
  - `Feedback()` for continuous learning, `Generate()` for ML-guided fingerprint creation
  - `Evolve()` for automated model evolution, `Train()` for retraining from registry
  - `Stats()` exposes inference/feedback/evolve counters

- **Online Learning System** (`modules/ml/learner.go`)
  - `OnlineLearner` with ring buffer for feedback samples and drift detection
  - `BrowserDistribution` tracker with KL divergence computation for distribution monitoring
  - Automatic drift detection via peak-vs-recent accuracy comparison

- **Profile Evolution Engine** (`modules/ml/evolution.go`)
  - `ProfileEvolutionEngine` with per-profile hit count, forgery rate EMA, confidence EMA
  - `CheckHealth()` returns `EvolutionHealthReport` with stale/forgery/drift profiles
  - `ShouldEvolve()` decision logic, `TopStaleProfiles()`, `TopForgeryProfiles()`
  - `SnapshotDistribution()` for browser distribution tracking across all profiles

- **ML-Driven Fingerprint Generator** (`modules/generator/generator.go`)
  - `SmartGenerator`: ML-validated generation with quality scoring and caching
  - Quality score: 40% (1-forgeryProb) + 30% consistency + 30% confidence
  - `GenerateBatch()`, `GenerateForBrowser()`, `GenerateForOS()` convenience methods
  - `RankProfiles()` for profile quality ranking, `FindSimilarProfiles()` via embedding distance
  - `GenerateFeatureVector()`, `GenerateEmbedding()` for raw ML data access

- **ML Facade API** (`modules/fingerprint/ml_api.go`)
  - `MLFacade`: unified entry point wrapping MLService + SmartGenerator + EvolutionEngine
  - `MLAnalyze()`, `MLAnalyzeWithBehavior()`, `MLAnalyzeBatch()` for ML-powered analysis
  - `MLGenerate()`, `MLGenerateRandom()`, `MLGenerateBatch()` for ML-guided generation
  - `MLValidate()`, `MLValidateAll()` for registry-wide validation
  - `MLFeedback()`, `MLEvolve()`, `MLTrain()`, `MLCheckHealth()` for lifecycle management
  - `MLFindSimilar()`, `MLEmbedding()`, `MLStats()` for advanced queries

- **TLS/HTTP/Cross-Layer ML Validators** (`modules/ml/tls_validator.go`)
  - `TLSValidator`: ML-guided TLS ClientHello configuration validation
  - `HTTPValidator`: ML-guided HTTP header/settings validation
  - `CrossLayerValidator`: cross-layer consistency checking (TLS + HTTP browser agreement)

### Changed

- **Gateway ML Integration** (`modules/gateway/gateway.go`)
  - Added optional `MLService` field to `Gateway` struct
  - `MLServiceEnabled` / `MLServiceConfig` configuration in `GatewayConfig`
  - `Analyze()` now enriches response with `MLValidation` when MLService is enabled
  - Forgery detection results automatically appended to `DefenseHints`
  - `GetMLService()` accessor for external consumers

## [v1.0.16] - 2026-03-11

### Added

- **International Browser Profiles** (`modules/profiles/international.go`)
  - 23 new profiles: Yandex (4 variants), Vivaldi (3), QQ Browser (2), UC Browser (2), Naver Whale (2), Mi Browser (1), Huawei Browser (1), OPPO Browser (1), Tor Browser (1), DuckDuckGo (1), 360 Safe Browser (1), Sogou Browser (1), Baidu Browser (1), Arc (2)
  - Maps Chromium-based international browsers to `BrowserChrome`, Tor to `BrowserFirefox`, DuckDuckGo/UC iOS to `BrowserSafari`

- **Bot & Automation Tool Profiles** (`modules/profiles/bots.go`)
  - 17 new profiles for forgery detection training
  - Headless browsers: Puppeteer Headless (120, 124), Puppeteer Stealth (120)
  - Playwright: Chromium 120, Firefox 121, WebKit 17.4
  - Selenium: ChromeDriver 120, GeckoDriver 121
  - HTTP clients: cURL/OpenSSL, Go net/http, Python requests, Node.js Axios, Scrapy
  - Anti-detect browsers: Multilogin Mimic 10, GoLogin Chrome 124, Dolphin Anty
  - Serverless: Cloudflare Worker
  - Each profile tagged with `bot_type`, `tool`, `forgery` metadata for ML labeling

- **Fingerprint Collector Tool** (`cmd/collector/main.go`)
  - Built-in knowledge bases: 15 JA3 hashes, 6 TLS fingerprints, 14 bot/headless entries, 20 international browser entries
  - Collected 55 fingerprints to `training/collected/fingerprints.json`

- **PyTorch GPU Training Pipeline** (`training/train_pytorch.py`)
  - 6 PyTorch model classes: EncoderNet, ClassifierNet, ForgeryDetectorNet, ForgeryTypeNet, ThreatAssessorNet, ActionNet
  - 4-phase training: encoder (triplet loss), classifier (cross-entropy), forgery detector (BCE+CE), threat assessor
  - RTX 2080 SUPER GPU acceleration with CUDA 12.4
  - Weight export to JSON for Go inference loading

- **Go Inference Validation** (`cmd/validate/main.go`)
  - Loads PyTorch-trained weights into Go ML pipeline
  - Per-profile inference with browser classification, forgery detection, threat assessment
  - Embedding quality analysis (same-family vs cross-family cosine similarity)

- **Training Feature Export** (`cmd/export_features/main.go`)
  - Exports all registered profiles as 30-dim feature vectors for PyTorch training

### Changed

- **Expanded training dataset**: 200 → 240 profiles with international and bot coverage
- **Retrained model**: overall accuracy 55% → 58.3%, Chrome 50% → 65.3%, Safari 62.5% → 70.6%
- **`BatchNormLayer.Params()`**: returns 4 parameters (gamma, beta, runMean, runVar) for correct PyTorch weight loading

## [v1.0.15] - 2026-03-12

### Added

- **Persistent Versioned Model Store** (`modules/ml/store.go`)
  - `ModelStore` manages versioned model snapshots on disk with auto-pruning
  - `StoreConfig` with configurable base directory and max version count (default 10)
  - `ModelManifest` index file tracks all versions with metadata (parent version, epochs, loss, accuracy)
  - `Save()` / `Load()` / `LoadLatest()` / `ListVersions()` for full model lifecycle management
  - Automatic cleanup of oldest versions when exceeding `MaxVersions`
  - Store persistence: manifest survives process restarts, all versions recoverable

- **Incremental Model Evolution** (`modules/ml/store.go`)
  - `Evolve()` fine-tunes existing weights with a lower learning rate (default 0.0001) and fewer epochs (default 10)
  - `EvolveConfig` for tuning fine-tuning behavior separately from full training
  - `EvolveAndSave()` convenience: evolve + save new version in one call
  - `TrainAndSave()` convenience: initial full training + save in one call
  - Models evolve continuously from saved weights — never retrain from scratch after initial training

- **Pipeline Store Integration**
  - `LoadFromStore()` / `SaveToStore()` convenience methods on `ModelPipeline`
  - 5 new model store tests: CreateAndLoad, MultipleVersions, EmptyLoadLatest, Persistence, LoadFromStore

### Changed

- **Agent Auto-loads Model from Store** (`modules/agent/agent.go`)
  - `AgentConfig.ModelStorePath` field for specifying model store directory
  - `NewAgent()` automatically loads latest model snapshot on startup if store path is configured
  - Zero-downtime model continuity: model weights persist across process restarts

### Fixed

- **Convert all Chinese comments to English** — strict compliance with DEVELOPER_GUIDE.md mandatory rule
  - `modules/ml/tensor.go`: 35 comment blocks converted
  - `modules/ml/nn.go`: 30 comment blocks converted
  - `modules/ml/models.go`: ~60 comment blocks converted (model docs, struct fields, function docs, inline comments)
  - `modules/ml/pipeline.go`: ~50 comment blocks converted (ASCII art diagram, phase headers, training comments, serialization docs)
  - `modules/ml/pipeline_test.go`: ~20 comment blocks converted (section headers, assertions, test descriptions)
  - `modules/agent/agent.go`: 6 comment blocks converted (Agent doc block, inference pipeline comments)

## [v1.0.14] - 2026-03-11

### Added

- **Pure-Go Tensor Computation Engine** (`modules/ml/tensor.go`)
  - `Tensor` struct with shape-aware operations: MatMul, Add, Sub, MulScalar, MulElem, Transpose, SoftmaxRow, SigmoidApply, ReluApply, Normalize, Argmax, Clamp, Clone
  - `ComputeDevice` interface abstracting hardware backends (CPU now, GPU/CUDA via CGo planned)
  - `cpuDevice` implementation with goroutine-parallel `BatchParallel` and `MatMul`
  - Factory functions: Zeros, Ones, RandN, RandNScaled, FromSlice, NewTensor
  - `SetDevice()` for global device switching

- **Neural Network Layer Library** (`modules/ml/nn.go`)
  - `Layer` interface with Forward/Backward/Params/SetTraining
  - `DenseLayer` with He initialization and full backpropagation
  - Activation layers: ReLU, Sigmoid, Softmax, Dropout (training mode support)
  - `Sequential` model for composable layer stacking
  - `AdamOptimizer` with moment estimation and bias correction
  - Loss functions: CrossEntropyLoss, BinaryCrossEntropyLoss, MSELoss, TripletMarginLoss

- **Four Domain-Specific Neural Network Models** (`modules/ml/models.go`)
  - `FingerprintEncoder`: 30-dim → 128 → 64 → 32-dim L2-normalized embeddings; trained with triplet loss on browser profile similarity
  - `BrowserClassifier`: embedding → 64 → 7-class softmax; identifies Chrome/Firefox/Safari/Edge/Opera/Brave/Samsung families
  - `ForgeryDetector`: dual-head network (40-dim combining fingerprint + cross-layer consistency features); DetectorNet for binary forgery probability, TypeNet for 4-class forgery type (Real/Headless/AntiDetect/Proxy)
  - `ThreatAssessor`: dual-head network (45-dim combining embedding + forgery signals + behavior); ThreatNet for 6-class threat classification, ActionNet for 5-class security action recommendation
  - Feature engineering: `EncodeFingerprint()` extracts 30-dim from TLS/HTTP2/TCP-IP/JS/Behavioral layers; `ComputeCrossLayerFeatures()` detects cross-layer inconsistencies

- **Training Pipeline & Model Serialization** (`modules/ml/pipeline.go`)
  - `ModelPipeline` with end-to-end inference: Infer(), InferFromFeatures(), InferBatch()
  - `trained` flag ensuring model outputs are only used after training or weight loading
  - `NeuralTrainer` with 4-phase training from ProfileRegistry: encoder(triplet) → classifier(cross-entropy) → forgery(binary CE) → threat(cross-entropy)
  - 3x Gaussian noise data augmentation, synthetic forged sample generation, rule-based threat labeling
  - `SaveWeights()`/`LoadWeights()` for JSON-based model serialization

- **Comprehensive ML Test Suite** (`modules/ml/pipeline_test.go`)
  - 23 tests: tensor ops, NN layers, all 4 domain models, feature encoding, pipeline inference (single/batch/from-features), model serialization, NeuralTrainer data building, CPU device

### Changed

- **Agent as Model Orchestrator** (`modules/agent/agent.go`)
  - Agent now holds `ModelPipeline` for neural inference alongside existing strategy engine and DQN
  - `Process()` integrates model pipeline: forgery detection with confidence thresholds (>0.6), threat assessment with action confidence gating (>0.7)
  - Pipeline only active after training/weight loading (`Trained()` guard)
  - New helpers: extractBehaviorVector, threatActionToAgentAction, forgeryTypeName, threatClassName

## [v1.0.13] - 2026-03-11

### Changed

- **Upgrade Reinforcement Learning from tabular Q-learning to Deep Q-Network (DQN)** (`modules/agent/reinforcement.go`)
  - Replace Q-table with multi-layer perceptron (MLP) neural network using He initialization and ReLU activations
  - Implement full SGD backpropagation with gradient clipping (±1.0) for stable training
  - Add experience replay buffer (configurable capacity, default 10000) with uniform mini-batch sampling
  - Add target network with periodic weight sync (default every 100 steps) for stable TD targets
  - Replace discretized 4-bucket state with continuous 8-dimensional state vector: [risk_score, ml_confidence, consistency, switch_rate, req_rate, risk_trend, obs_count, unique_fp_ratio]
  - `RLConfig` extended: `HiddenLayers []int`, `LearningRate`, `ReplayCapacity`, `BatchSize`, `TargetUpdateFreq`, `StateDim`
  - New continuous API: `SelectActionContinuous()`, `UpdateContinuous()`, `QValueContinuous()`, `BestActionContinuous()`
  - `ExtractStateVector()` generates 8-dim normalized feature vector directly from observations (no discretization loss)
  - `RLStats` extended: `Steps`, `TrainSteps`, `AvgLoss`, `ReplaySize`, `NetworkLayers`
  - Backward-compatible API preserved: `SelectAction()`, `Update()`, `QValue()`, `BestAction()` convert discrete State to vector internally
- **Update Agent integration** (`modules/agent/agent.go`)
  - `Process()` now uses `ExtractStateVector` + `SelectActionContinuous` + `QValueContinuous`
  - `ReportReward()` now uses `ExtractStateVector` + `UpdateContinuous` with terminal flag
  - Insight text updated from "RL" to "DQN" prefix
- **Rewrite DQN test suite** (`modules/agent/reinforcement_test.go`)
  - 18 test functions covering: neural net forward/backward/copy, replay buffer, state extraction, DQN greedy/exploration/convergence/loss/target-sync, backward compatibility, stats, agent integration
  - Contrastive training with all 5 actions for robust convergence verification

## [v1.0.12] - 2026-03-11

### Added

- **Agent Reinforcement Learning** (`modules/agent/reinforcement.go`)
  - Tabular Q-learning with epsilon-greedy exploration for action selection
  - `RLConfig` with tunable Alpha, Gamma, EpsilonMax/Min/Decay parameters
  - Discretized state space (threat × risk × consistency × switch-rate buckets)
  - `ComputeReward` function scoring true/false positive/negative outcomes
  - Integrated into `Agent.Process()` for escalation-only RL override
- **ML Online Learning** (`modules/ml/online.go`)
  - `OnlineClassifier` with incremental centroid updates (online mean formula)
  - `WeightedPartialFit` for importance-weighted sample ingestion
  - Concept drift detection via sliding-window accuracy tracking
  - `OnlineHierarchicalClassifier` with three-layer Protocol→Family→Version cascade
- **Contextual Bandit Strategy Selection** (`modules/agent/bandit.go`)
  - LinUCB algorithm with per-arm linear models for strategy prioritization
  - 8-dimensional context vector from observation + behavior + ML signals
  - `BuildContext` constructs normalized feature vector for bandit input
  - `RandomBandit` baseline for A/B testing comparison
  - Integrated into `Agent.Process()` to prioritize among triggered strategies
- **External feedback loop** via `Agent.ReportReward()` method feeding both RL and Bandit subsystems
- Unit tests for all three subsystems (`reinforcement_test.go`, `bandit_test.go`, `online_test.go`)

## [v1.0.11] - 2026-03-11

### Removed

- **Dead internal logger** (`modules/internal/logger/`) — standalone logger with zero external imports
- **Dead hand-rolled metrics module** (`modules/metrics/`) — Counter/Gauge/Histogram/Summary types with zero imports; the Prometheus-based `modules/internal/metrics/` is retained

### Changed

- **Consolidate error system** — `modules/errors` is now the canonical error package
  - Add `CoreError` type, `NewCodedError`, `NewCodedErrorf`, `WrapError`, `WrapErrorf` functions
  - Add `VAL`, `NTF`, `SEC` error code families and 10 new sentinel errors
  - Rewrite `modules/core/errors.go` as thin re-export layer (type aliases + function forwarding)
- **Fix root Dockerfile** — correct HEALTHCHECK syntax, remove references to deleted modules/metrics and non-existent examples/v3
- **Simplify root docker-compose.yml** — dev-oriented config, remove Redis/Nginx services, reference `deploy/docker/` for production
- **Translate Chinese comments to English** in `modules/plugin/bridge.go` and `modules/internal/plugins/` (types.go, registry.go, basic_plugin.go)
- **Document plugin architecture** — add package-level doc explaining the three plugin subsystems (internal/plugin, internal/plugins, internal/extension)

## [v1.0.10] - 2026-03-11

### Removed

- **Compiled binaries from git tracking** (~25MB total)
  - Remove 5 ELF binaries under `examples/` that were accidentally committed
  - Add binary paths to `.gitignore` to prevent future tracking
- **Dead code cleanup**
  - Delete `modules/kit/strings.go` (6 stdlib wrapper functions with zero external references)
  - Remove unused `Min()`, `Max()`, `Clamp()` from `modules/core/utils.go` (Go 1.25 has builtin `min`/`max`)
  - Remove corresponding `TestMinMax` and `TestClamp` test functions

### Fixed

- **Duplicate test function** `TestRiskLevelString` declared in both `core/constants_test.go` and `core/types_test.go`
- **Undefined function calls** in `modules/kit/useragent.go` after dead code removal, replaced with `strings.*` stdlib calls

### Changed

- **Remove 52-line commented-out example code** from `modules/internal/observability/observability.go`

## [v1.0.9] - 2026-03-11

### Added

- **API key authentication middleware** (`modules/gateway/auth.go`)
  - Constant-time API key validation via `X-API-Key` header or `api_key` query parameter
  - Configurable skip paths for health/metrics endpoints
  - JSON error responses with proper `Content-Type: application/json`
  - 6 unit tests covering all authentication scenarios

### Security

- **SSRF protection for reverse proxy** (`modules/gateway/injector.go`)
  - Add `validateProxyTarget()` to block loopback, link-local, and cloud metadata endpoints
  - Support `AllowPrivateTarget` config for Docker/Kubernetes environments
  - 4 SSRF validation tests (loopback rejection, public allow, empty host, allowPrivate)

### Changed

- **HTML injector performance optimization** (`modules/gateway/injector.go`)
  - Replace per-request `regexp.MustCompile` + `QuoteMeta` with `strings.Index` for `CustomInjectionPoint` matching
- **Translate all remaining Chinese comments to English** in `modules/gateway/injector.go`

## [v1.0.8] - 2026-03-10

### Changed

- Simplify CONTRIBUTING.md and SECURITY.md documentation

## [v1.0.7] - 2026-03-10

### Added

- CONTRIBUTING.md - Complete contribution guide and workflow
- SECURITY.md - Security policy and vulnerability reporting process

## [v1.0.6] - 2026-03-10

### Added

- **Complete version control development rules** - ensure Git standards compliance
  - Added version control rules in `docs/DEVELOPER_GUIDE.md`: 7-step mandatory release workflow
  - Clear definition of version control rules: CHANGELOG → Version bump → Tag → Push
  - Version management requirements for all modules
  - Definition of strict rules and consequences (commit rejection/rollback)

### Changed

- Updated DEVELOPER_GUIDE.md with detailed release process

### Fixed

- Version control consistency across all modules

## [v1.0.5] - 2026-03-10

### Added

- **Multi-language internationalization support (i18n: English/Chinese)**
  - Full i18n framework with dual-language dictionary
  - 500+ translation keys for frontend
  - Real-time language switching capability

### Fixed

- Fixed Profiles page modal styling
- Corrected modal overlay and detail section CSS

## [v1.0.4] - 2026-03-11

### Added

- **Deep frontend integration: full module visualization**
  - 18 advanced API endpoints for comprehensive module access
  - Analysis Engine page with complete analysis pipeline
  - ML Engine page with three-tier classification architecture
  - Defense System page with threat detection
  - Anti-Detection Engine page with JS generator
  - Plugin System page with extension architecture
  - Fingerprint Tools page with JA3/JA4/validation and comparison

### Changed

- 6 interactive SPA pages for advanced functionality
- ~130 new CSS lines for styling
- 15 API client methods + ~600 lines of page logic

## [v1.0.3] - 2026-03-10

### Added

- **Full frontend integration with real-time capabilities**
  - Real-time log capture and SSE push system
  - Agent status visualization
  - Knowledge base browser
  - 7 configuration sections (Server/RateLimit/Cache/ML/AntiDetect/Scanner/Agent)
  - Real logs display with level filtering
  - Dynamic system status rendering

### Changed

- Rewritten log handling with real LogBuffer integration
- Real-time config hot updates with thread-safe callbacks
- Dynamic Agent statistics and status

## [v1.0.2] - 2026-03-10

### Added

- **Global fingerprint knowledge base**
  - Accurate browserfingerprinting blueprints for 7 major families
  - 15+ version specifications with precise cipher suites
  - 5 OS family TCP/IP stack signatures
  - HTTP/2 pseudo-header ordering
  - TLS 1.3 standard suites and GREASE values
  - Market share estimation data

- **Knowledge-driven anomaly detection**
  - Cross-layer consistency validation (TLS ↔ HTTP/2 ↔ TCP/IP ↔ JS)
  - Cipher suite and extension count validation
  - HTTP/2 parameter validation
  - TCP/IP TTL and window size validation
  - Headless browser and automation detection
  - ML classification confidence validation
  - Contradiction signal weighted aggregation

- **Autonomous security agent (OADA decision loop)**
  - Observe: Client behavior profiling with sliding window
  - Analyze: Base behavior analysis + knowledge verification
  - Decide: Adaptive strategy engine with 5 response actions
  - Act: 6 threat classifications and automatic enforcement
  - Integrated with Gateway.Analyze() pipeline

### Security

- P0: Response body size limit in injector module
- P0: Concurrent safety for ProfileRegistry with RWMutex
- P0: JA3Hash algorithm fix (sha256 → md5)
- P0: RateLimiter goroutine leak fix
- P0: Gateway security hardening (request size limit, error handling)

### Fixed

- P1: Regex precompilation in injector
- P1: ML confidence score calculation (weighted average instead of multiplication)
- P1: JS anti-detection configurable property fix
- P1: GetProfile returns copy to prevent external modification
- P1: CalculateMD5 implementation fix (crypto/md5 instead of SHA256)
- P1: OperatingSystems random selection probability fix
- P1: Dual error system documentation

### Changed

- Docs and code alignment with new module structure
- Stable sorting for ListProfiles
- Code optimization and refactoring

## [v1.0.1] - 2026-03-05

### Added

- **Go Workspace architecture redesign**
  - Migrated from single module to 14-module workspace
  - Core modules: core, profiles, tls, http, ml, defense, frontend, gateway
  - Utility modules: generator, network, internal, config, plugin, fingerprint
  - Facade pattern for unified API entry point

- **Standard logging interface**
  - Unified Logger interface definition
  - Adapters for slog, zap, logrus, and stdlib
  - NoOpLogger for testing

- **Comprehensive test coverage and performance optimization**
  - Phase 6: 87.8% coverage with advanced detection system
  - Phase 7: 10.3x performance improvement in behavior analysis
  - Phase 8: Security audit and hardening

- **Profile dynamic management**
  - ReloadProfile and ReloadAll methods
  - GetProfilesByBrowser/GetProfilesByOS categorized queries
  - CloneProfile functionality

- **Code generation tools**
  - profilegen tool for YAML to Go code generation
  - Zero warnings policy

- **Prometheus metrics integration**
  - Fingerprint generation metrics
  - Cache metrics
  - Connection metrics
  - Behavior analysis metrics
  - HTTP/2 analysis metrics
  - Grafana dashboard configuration

### Security

- Dependency vulnerability scanning and patching
- Code security scanning compliance
- Sensitive information detection (no hardcoded secrets)

### Fixed

- calculateHeaderOrderMatch algorithm bug
- internal/errors package missing sentinel errors
- Go version upgrade to 1.25.7 (11 stdlib vulnerability fixes)

### Changed

- ConfigCenter.copyConfig optimization (JSON → Clone method)
- Optimized CloneProfile ID conflict checking

## [v1.0.0] - 2026-03-01

### Added

- Initial release
- TLS fingerprint identification (JA3/JA4/JA4S)
- HTTP/2 signature analysis
- Browser fingerprint profile management
- Behavior analysis module
