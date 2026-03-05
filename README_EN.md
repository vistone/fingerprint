# Fingerprint

[![Go Reference](https://pkg.go.dev/badge/github.com/vistone/fingerprint.svg)](https://pkg.go.dev/github.com/vistone/fingerprint)
[![License](https://img.shields.io/badge/License-BSD_3--Clause-blue.svg)](https://opensource.org/licenses/BSD-3-Clause)
[![Version](https://img.shields.io/badge/version-2.0.0-blue.svg)](https://github.com/vistone/fingerprint/releases/tag/v2.0.0)
[![Go Version](https://img.shields.io/badge/go-1.25.7+-blue.svg)](https://golang.org)

High-performance browser TLS fingerprinting library providing 200+ browser fingerprint configurations and comprehensive fingerprinting capabilities.

## Features

- **200+ Real Browser Fingerprints** - Chrome, Firefox, Safari, Edge, Opera, Brave, etc.
- **TLS Fingerprinting** - JA3/JA4 fingerprint generation and analysis
- **HTTP/2 Signatures** - Complete frame analysis and signature matching
- **Machine Learning** - Built-in ML anomaly detection
- **Security Protection** - Risk scoring and anomaly detection
- **Go Workspace** - Modular architecture, import only what you need

## Quick Start

```bash
go get github.com/vistone/fingerprint/modules/fingerprint
```

```go
package main

import (
    "fmt"
    "github.com/vistone/fingerprint/modules/fingerprint"
)

func main() {
    // Get a random Chrome fingerprint
    profile := fingerprint.GetRandomByBrowser(fingerprint.BrowserChrome)
    fmt.Printf("Selected: %s\n", profile.Name)
}
```

## Module Structure

```
modules/
├── core          # Core types (zero dependencies)
├── profiles      # 200+ browser fingerprints
├── tls           # TLS fingerprint analysis
├── http          # HTTP/2 analysis
├── ml            # ML classifier
├── defense       # Security protection
├── gateway       # API gateway
├── generator     # Fingerprint generator
├── fingerprint   # Facade entry point
└── ...
```

## Usage

### Pattern 1: Facade Module (Recommended)

```go
import "github.com/vistone/fingerprint/modules/fingerprint"

profile := fingerprint.GetRandom()
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

| Browser | Versions | Count |
|---------|----------|-------|
| Chrome | 115-144 | 64 |
| Firefox | 115-140 | 43 |
| Safari | 16-18 | 48 |
| Edge | 115-134 | 23 |
| Opera | 100-110 | 12 |
| Brave | 1.60-1.72 | 7 |
| Mobile | iOS/Android | 3 |
| **Total** | | **200** |

## Documentation

- [Architecture](./docs/ARCHITECTURE_EN.md) - Architecture overview
- [API Documentation](./docs/API.md) - Complete API reference
- [Developer Guide](./docs/DEVELOPER_GUIDE.md) - Development guide
- [Changelog](./docs/CHANGELOG.md) - Version history

## Examples

See [examples/](./examples/) directory:

```bash
cd examples/basic
go run .
```

## Performance

- Fingerprint selection: O(1) hash table lookup
- Zero-allocation critical paths
- Concurrent-safe design

## License

BSD 3-Clause License
