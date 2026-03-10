# API Documentation

This document describes the public API usage of the fingerprint library (Go Workspace version).

## Quick Start

### Installation

```bash
go get github.com/vistone/fingerprint/modules/fingerprint
```

### Basic Usage

```go
package main

import (
    "github.com/vistone/fingerprint/modules/profiles"
    "github.com/vistone/fingerprint/modules/tls"
)

func main() {
    // Get fingerprint
    profile, _ := profiles.Get("chrome_133")
    
    // Use TLS module
    ja3 := tls.CalculateJA3(clientHello)
}
```

## Import Path Reference

| Old Path (Deprecated) | New Path |
|----------------------|----------|
| `github.com/vistone/fingerprint` | `github.com/vistone/fingerprint/modules/fingerprint` |
| `github.com/vistone/fingerprint/profiles` | `github.com/vistone/fingerprint/modules/profiles` |
| `github.com/vistone/fingerprint/tls/ja3` | `github.com/vistone/fingerprint/modules/tls` |
| `github.com/vistone/fingerprint/http/ja4h` | `github.com/vistone/fingerprint/modules/http/legacy/ja4h` |
| `github.com/vistone/fingerprint/internal/config` | `github.com/vistone/fingerprint/modules/config` |
| `github.com/vistone/fingerprint/internal/tcpip` | `github.com/vistone/fingerprint/modules/internal/tcpip` |
| `github.com/vistone/fingerprint/types` | `github.com/vistone/fingerprint/modules/core/types` |

## Core Module API

### Basic Types

```go
import "github.com/vistone/fingerprint/modules/core"

// Browser types
const (
    BrowserChrome  = core.BrowserChrome
    BrowserFirefox = core.BrowserFirefox
    BrowserSafari  = core.BrowserSafari
    BrowserEdge    = core.BrowserEdge
    BrowserOpera   = core.BrowserOpera
    BrowserBrave   = core.BrowserBrave
)

// Operating system types
const (
    OSWindows10   = core.OSWindows10
    OSWindows11   = core.OSWindows11
    OSMacOS13     = core.OSMacOS13
    OSMacOS14     = core.OSMacOS14
    OSMacOS15     = core.OSMacOS15
    OSLinux       = core.OSLinux
    OSiOS         = core.OSiOS
    OSAndroid     = core.OSAndroid
)

// Error codes
const (
    CodeOK              = 0
    CodeInvalidInput    = 400
    CodeNotFound        = 404
    CodeInternalError   = 500
    CodeTimeout         = 503
)
```

## Profiles Module API

### Get Fingerprint

```go
import "github.com/vistone/fingerprint/modules/profiles"

// Get by ID
profile, ok := profiles.Get("chrome_133")
if !ok {
    log.Fatal("profile not found")
}

// Get random fingerprint
profile := profiles.GetRandom()

// Get by browser type
chromeProfiles := profiles.GetByBrowser(core.BrowserChrome)

// Get all profiles
allProfiles := profiles.GetAll()

// Count profiles
count := profiles.Count()
```

### Profile Structure

```go
type Profile struct {
    ID              string
    BrowserType     string
    Version         string
    
    // TLS fingerprint
    TLSVersion      string
    CipherSuites    []uint16
    Extensions      []uint16
    
    // HTTP/2 settings
    Settings        []http.Setting
    SettingsOrder   []http.SettingID
    
    // TCP/IP
    TTL             int
    WindowSize      int
    
    // Metadata
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

## TLS Module API

### JA3 Fingerprint

```go
import "github.com/vistone/fingerprint/modules/tls"

// Calculate JA3 from ClientHello
ja3String := tls.CalculateJA3(clientHello)

// Parse TLS handshake
analyzer := tls.NewAnalyzer(conn)
fingerprint := analyzer.Analyze()

// Get JA3 hash
ja3Hash := tls.CalculateJA3Hash(clientHello)
```

### Supported Algorithms

- JA3: MD5 hash of TLS fingerprint
- JA4: Compact TLS fingerprint representation
- JA4S: Server-side JA4 variant

## HTTP Module API

### HTTP/2 Signature

```go
import "github.com/vistone/fingerprint/modules/http"

// Analyze HTTP/2 connection
sig := http.AnalyzeHTTP2(conn)

// Get settings order
settings := http.GetSettingsOrder()

// Validate pseudo-header order
err := http.ValidatePseudoHeaderOrder(headers)
```

## Gateway Module API

### Profile Manager

```go
import "github.com/vistone/fingerprint/modules/gateway"

// Create manager
manager := gateway.NewProfileManager()

// Load profiles
err := manager.LoadProfiles(profileDir)

// Get profile
profile, err := manager.GetProfile("chrome_133")

// List profiles
profiles := manager.ListProfiles()

// Reload
err := manager.ReloadAll()
```

### Analysis

```go
// Analyze request
result := gateway.Analyze(request, profile)

// Get risk score
score := result.RiskScore

// Get threat type
threat := result.ThreatType
```

## Generator Module API

### Random Fingerprint

```go
import "github.com/vistone/fingerprint/modules/generator"

// Generate random fingerprint
profile := generator.GenerateRandom()

// Generate by browser type
chromeProfile := generator.GenerateByBrowser(core.BrowserChrome)

// Generate with constraints
profile := generator.GenerateWithConstraints(opts)
```

## ML Module API

### Classification

```go
import "github.com/vistone/fingerprint/modules/ml"

// Create classifier
clf := ml.NewClassifier()

// Load model
err := clf.LoadModel(modelPath)

// Classify
result := clf.Classify(features)

// Get confidence scores
scores := result.Scores
```

## Defense Module API

### Risk Scoring

```go
import "github.com/vistone/fingerprint/modules/defense"

// Create analyzer
analyzer := defense.NewRiskAnalyzer()

// Analyze request
risk := analyzer.Analyze(request, profile)

// Get risk level
level := risk.Level

// Get risk reasons
reasons := risk.Reasons
```

### Detection Rules

```go
// Check for anomalies
anomalies := defense.DetectAnomalies(fingerprint)

// Check for known bots
isBot := defense.IsKnownBot(ua)

// Check headless browser indicators
isHeadless := defense.IsHeadlessBrowser(fingerprint)
```

## Agent Module API

### Decision Making

```go
import "github.com/vistone/fingerprint/modules/agent"

// Create agent
agent := agent.NewAgent(config)

// Process request
decision := agent.Process(request, profile)

// Get decision
action := decision.Action // Allow, Monitor, Challenge, Throttle, Block

// Get threat classification
threat := decision.ThreatType
```

## Config Module API

### Configuration Management

```go
import "github.com/vistone/fingerprint/modules/config"

// Load configuration
cfg, err := config.Load("config.yaml")

// Update configuration
err = cfg.Update(newConfig)

// Get specific setting
value := cfg.Get("key")

// Register listener
cfg.RegisterListener(func(old, new interface{}) {
    // Handle configuration change
})
```

## Internal Module API

### Logger

```go
import "github.com/vistone/fingerprint/modules/internal/logger"

// Structured logging
logger.Debug("message", "key", value)
logger.Info("message", "key", value)
logger.Warn("message", "key", value)
logger.Error("message", "err", err)
```

### Connection Pool

```go
// Create connection pool
pool := internal.NewConnPool(opts)

// Get connection
conn, err := pool.Get()

// Return connection
pool.Put(conn)

// Close pool
pool.Close()
```

## Error Handling

### Error Codes

```go
const (
    CodeOK                 = 0      // Success
    CodeBadRequest         = 400    // Invalid input
    CodeUnauthorized       = 401    // Authentication failed
    CodeForbidden          = 403    // Access denied
    CodeNotFound           = 404    // Resource not found
    CodeConflict           = 409    // Resource conflict
    CodeInternalError      = 500    // Internal server error
    CodeServiceUnavailable = 503    // Service unavailable
)
```

### Error Types

```go
// Custom error types
var (
    ErrProfileNotFound    = errors.New("profile not found")
    ErrInvalidInput       = errors.New("invalid input")
    ErrConnectionFailed   = errors.New("connection failed")
    ErrTimeoutError       = errors.New("request timeout")
)

// Detailed error information
type DetailedError struct {
    Code    int
    Message string
    Details map[string]interface{}
}
```

## Examples

### Get Browser Fingerprint

```go
package main

import (
    "fmt"
    "github.com/vistone/fingerprint/modules/profiles"
    "github.com/vistone/fingerprint/modules/core"
)

func main() {
    // Get Chrome 133 profile
    profile, ok := profiles.Get("chrome_133")
    if !ok {
        panic("profile not found")
    }
    
    fmt.Printf("Profile: %+v\n", profile)
}
```

### Analyze TLS Handshake

```go
package main

import (
    "fmt"
    "github.com/vistone/fingerprint/modules/tls"
)

func main() {
    // Create TLS analyzer
    analyzer := tls.NewAnalyzer(conn)
    
    // Analyze connection
    fingerprint := analyzer.Analyze()
    
    // Print JA3
    fmt.Println("JA3:", fingerprint.JA3)
}
```

### Risk Assessment

```go
package main

import (
    "fmt"
    "github.com/vistone/fingerprint/modules/gateway"
    "github.com/vistone/fingerprint/modules/defense"
)

func main() {
    // Create analyzer
    analyzer := defense.NewRiskAnalyzer()
    
    // Assess risk
    risk := analyzer.Analyze(request, profile)
    
    // Check risk level
    if risk.Score > 70 {
        fmt.Println("High risk:", risk.Reasons)
    }
}
```

## Migration Guide

### From v1.0.7 to v1.0.8

1. Update imports
   ```go
   // Old
   import "github.com/vistone/fingerprint"
   
   // New
   import fp "github.com/vistone/fingerprint/modules/fingerprint"
   ```

2. Update function calls
   ```go
   // Old
   fingerprint.GetProfile()
   
   // New
   fp.GetProfile()
   ```

3. Run tests
   ```bash
   go test ./modules/...
   ```

## FAQ

**Q: Which module should I import?**
A: Start with `modules/fingerprint` for the facade API, or import specific modules like `modules/profiles`, `modules/tls`, etc.

**Q: How do I handle errors?**
A: Check error types and use error sentinel values. Use `errors.Is()` for error comparisons.

**Q: How do I integrate with my application?**
A: Use the Gateway module's `Analyze()` method to get risk scores and threat classifications.

**Q: Is the API thread-safe?**
A: Yes, all public APIs are thread-safe. Use sync.RWMutex for concurrent access.

## References

- [Developer Guide](./DEVELOPER_GUIDE.md)
- [Architecture](./ARCHITECTURE.md)
- [Changelog](./CHANGELOG.md)
