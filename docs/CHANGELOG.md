# Changelog

This project follows the [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) specification.

## [Unreleased]

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
