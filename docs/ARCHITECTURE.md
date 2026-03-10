# Architecture Documentation

This document describes the Go Workspace architecture, module structure, and design decisions of the fingerprint library.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                     Go Workspace (go.work)                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │              modules/fingerprint (Facade)                 │   │
│  │         Unified API entry, integrates all submodules       │   │
│  └──────────────────────────────────────────────────────────┘   │
│                              │                                   │
│  ┌─────────────┬─────────────┼─────────────┬─────────────┐      │
│  │             │             │             │             │      │
│  ▼             ▼             ▼             ▼             ▼      │
│  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐  │
│  │ core │  │profiles│  │ tls │  │ http │  │  ml  │  │defense│  │
│  └──────┘  └──────┘  └──────┘  └──────┘  └──────┘  └──────┘  │
│  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐  │
│  │frontend│  │gateway │  │generator│  │network│  │internal│  │config │  │
│  └──────┘  └──────┘  └──────┘  └──────┘  └──────┘  └──────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Module Structure

```
github.com/vistone/fingerprint/
├── go.work                     # Workspace definition
├── modules/                    # All modules
│   ├── core/                   # Core types (zero dependencies)
│   ├── profiles/               # Browser fingerprint configurations
│   ├── tls/                    # TLS fingerprint analysis
│   ├── http/                   # HTTP fingerprint analysis
│   ├── ml/                     # ML classifier
│   ├── defense/                # Security protection
│   ├── frontend/               # Frontend SDK
│   ├── gateway/                # API gateway
│   ├── generator/              # Fingerprint generator
│   ├── network/                # Network layer
│   ├── internal/               # Internal utilities
│   ├── config/                 # Configuration management
│   ├── plugin/                 # Plugin system
│   └── fingerprint/            # Facade entry point
├── cmd/                        # Application entry points
├── examples/                   # Example code
└── docs/                       # Documentation
```

## Module Details

### Core Module (`modules/core`)

**Responsibility**: Provides basic types and interfaces shared by all modules.

```go
package core

type BrowserType string
type OperatingSystem string
type ClientProfile struct { ... }
type HTTPHeaders struct { ... }
type TLSExtension struct { ... }
```

**Zero Dependency Principle**: The core module has no external dependencies.

### Profiles Module (`modules/profiles`)

**Responsibility**: Manages TLS/HTTP client fingerprint configurations.

- **200+ Browser Fingerprints**: Chrome, Firefox, Safari, Edge, Opera, Brave
- Organized by browser type in separate files
- `legacy/` subdirectory for backward compatibility

### TLS Module (`modules/tls`)

**Responsibility**: TLS fingerprint analysis and JA3/JA4 calculation.

```go
import "github.com/vistone/fingerprint/modules/tls"

// JA3 fingerprint
ja3 := tls.CalculateJA3(clientHello)
```

### HTTP Module (`modules/http`)

**Responsibility**: HTTP/2 and HTTP header analysis.

### ML Module (`modules/ml`)

**Responsibility**: Machine learning classifier.

### Defense Module (`modules/defense`)

**Responsibility**: Security protection and anomaly detection.

## Usage Patterns

### Pattern 1: Using Facade Module (Recommended)

```go
import "github.com/vistone/fingerprint/modules/fingerprint"

// Get random fingerprint
profile := fingerprint.GetRandom()

// Get fingerprint by browser
chrome := fingerprint.GetByBrowser(fingerprint.BrowserChrome)
```

### Pattern 2: Direct Submodule Import

```go
import (
    "github.com/vistone/fingerprint/modules/core"
    "github.com/vistone/fingerprint/modules/profiles"
)

profile, _ := profiles.Get("chrome_133")
chromeProfiles := profiles.GetByBrowser(core.BrowserChrome)
```

## Fingerprint Coverage

| Browser | Version Range | Count |
|---------|--------------|-------|
| Chrome | 115-144 | 64 |
| Firefox | 115-140 | 43 |
| Safari | 16-18 | 48 |
| Edge | 115-134 | 23 |
| Opera | 100-110 | 12 |
| Brave | 1.60-1.72 | 7 |
| Mobile | iOS/Android | 3 |
| **Total** | | **200** |

## Design Principles

1. **Modular Design**: Each module has clear boundaries and can be developed independently
2. **Zero Dependency Core**: The `core` module has no external dependencies
3. **Progressive Complexity**: From basic types to advanced features
4. **Backward Compatibility**: Legacy code kept in `legacy/` subdirectories

## Documentation

- [Architecture](./ARCHITECTURE.md) - Architecture overview (Chinese)
- [API Documentation](./API.md) - Complete API reference
- [Developer Guide](./DEVELOPER_GUIDE.md) - Development and contribution guide
- [Changelog](./CHANGELOG.md) - Version history

## License

BSD 3-Clause License
