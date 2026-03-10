# Changelog

This project follows the [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) specification.

## [Unreleased]

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
