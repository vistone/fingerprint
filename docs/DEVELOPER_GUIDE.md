# Developer Guide

This guide is for developers who want to contribute code or deeply understand the internal implementation of the fingerprint library.

## Development Environment Setup

### Prerequisites

- Go 1.25+
- Make
- Git
- golangci-lint (optional, for local linting)

### Clone and Build

```bash
# Clone repository
git clone https://github.com/vistone/fingerprint.git
cd fingerprint

# Sync workspace
go work sync

# Build specific module
go build ./modules/core
go build ./modules/profiles
go build ./modules/fingerprint

# Run tests
go test ./modules/...

# Run specific module tests
go test ./modules/profiles/... -v

# Run tests with race detector
go test -race ./modules/core/...

# Generate coverage report
go test -coverprofile=coverage.out ./modules/...
go tool cover -html=coverage.out -o coverage.html
```

### Makefile Commands

```bash
make help          # Show all available commands
make test          # Run all tests
make test-race     # Run tests with race detector
make coverage      # Generate coverage report
make lint          # Run linter
make fmt           # Format code
make clean         # Clean build artifacts
```

## Go Workspace Structure

```plaintext
github.com/vistone/fingerprint/
├── go.work                     # Workspace definition
├── modules/                    # All modules
│   ├── core/                   # Core types (zero dependencies)
│   ├── profiles/               # Fingerprint profiles
│   ├── tls/                    # TLS analysis
│   ├── http/                   # HTTP analysis
│   ├── ml/                     # ML classifier
│   ├── defense/                # Security defense
│   ├── frontend/               # Frontend SDK
│   ├── gateway/                # API Gateway
│   ├── generator/              # Fingerprint generator
│   ├── network/                # Network layer
│   ├── internal/               # Internal utilities
│   ├── config/                 # Configuration management
│   ├── plugin/                 # Plugin system
│   └── fingerprint/            # Facade module
├── cmd/                        # Application entry points
├── examples/                   # Example code
└── test/                       # Integration tests
```

## Module Development Standards

### Directory Convention

Each module follows this structure:

```plaintext
modules/<module>/
├── go.mod                  # Module definition
├── *.go                    # Public API
├── *_test.go               # Unit tests
└── legacy/                 # Legacy compatibility code (optional)
    └── *.go
```

### Module Dependency Rules

```plaintext
# Allowed dependency direction
core (zero dependencies)
    ▲
    ├── profiles ──▶ core
    ├── tls ───────▶ core
    ├── http ──────▶ core
    ├── ml ────────▶ core
    └── ...

# Circular dependencies are forbidden
A ──▶ B ──▶ A  ❌
```

### Creating New Module

```bash
# 1. Create directory
mkdir modules/mynewmodule
cd modules/mynewmodule

# 2. Initialize go.mod
cat > go.mod << 'EOF'
module github.com/vistone/fingerprint/modules/mynewmodule

go 1.25.7

require github.com/vistone/fingerprint/modules/core v0.0.0

replace github.com/vistone/fingerprint/modules/core => ../core
EOF

# 3. Create main.go
cat > mynewmodule.go << 'EOF'
package mynewmodule

import "github.com/vistone/fingerprint/modules/core"

// Public API
EOF

# 4. Add to workspace
cd ../..
echo "./modules/mynewmodule" >> go.work

# 5. Sync
go work sync
```

## Coding Standards

### Comments Must Be English ⚠️

**MANDATORY RULE**: All code comments, documentation, and commit messages **MUST be written in English**. Do NOT mix Chinese and English.

```go
// ✅ CORRECT: English only
// Calculate the checksum of the profile
func calculateChecksum(data []byte) string {
    // ...
}

// ❌ WRONG: Never use Chinese comments
// 计算profile的校验和  // FORBIDDEN!
func calculateChecksum(data []byte) string {
    // ...
}

// ❌ WRONG: Never mix languages
// 获取 profile，返回 error // FORBIDDEN!
func getProfile(id string) error {
    // ...
}
```

**Consequences**:
- Code review will reject mixed-language comments
- Pull requests must pass English-only comment check
- CI will fail on Chinese comments in code files

**Exception**: Chinese comments are allowed ONLY in Chinese documentation files (`*_zh-cn.md`)

### Error Handling

```go
// Always check errors
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// Define error sentinel values
var ErrInvalidProfile = errors.New("invalid profile")

// Create custom error types
type ValidationError struct {
    Field string
    Reason string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error: %s - %s", e.Field, e.Reason)
}
```

### Logging

```go
import "github.com/vistone/fingerprint/modules/internal/logger"

// Use structured logging
logger.Debug("request processed", "profile", name, "duration_ms", elapsed)
logger.Info("profile created", "id", id, "browser", browser)
logger.Warn("cache miss", "key", key)
logger.Error("database error", "err", err)
```

### Concurrency Safety

```go
type Registry struct {
    mu       sync.RWMutex
    profiles map[string]Profile
}

// Use RWMutex for read-heavy workloads
func (r *Registry) Get(id string) (Profile, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    p, ok := r.profiles[id]
    return p, ok
}

// Use sync.Once for lazy initialization
var singleton *Instance
var once sync.Once

func GetInstance() *Instance {
    once.Do(func() {
        singleton = &Instance{}
    })
    return singleton
}
```

### Testing

```go
// Unit test template
func TestFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:  "success case",
            input: "valid input",
            want:  "expected output",
        },
        {
            name:    "error case",
            input:   "invalid input",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Function(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("got %q, want %q", got, tt.want)
            }
        })
    }
}

// Benchmark template
func BenchmarkFunction(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Function("input")
    }
}
```

### Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./modules/...

# View HTML report
go tool cover -html=coverage.out

# Check coverage for specific package
go tool cover -func=coverage.out | grep modules/core
```

## Version Control Rules (Mandatory)

### Git Workflow

1. **Create feature branch**
   ```bash
   git checkout -b feature/feature-name
   ```

2. **Make changes and test**
   ```bash
   go test ./modules/... -race
   go fmt ./...
   golangci-lint run
   ```

3. **Commit with conventional message**
   ```bash
   git commit -m "feat(module): description"
   git commit -m "fix(module): description"
   git commit -m "docs: description"
   ```

4. **Push and create PR**
   ```bash
   git push origin feature/feature-name
   ```

### Release Process (7-Step Mandatory Workflow)

**Step 1: Code changes**
- Implement feature
- Add tests
- All tests pass: `go test -race ./modules/...`

**Step 2: Update CHANGELOG**
- Add entry to `docs/CHANGELOG.md` under `[Unreleased]`
- Use format: `### Added`, `### Changed`, `### Fixed`, `### Security`

**Step 3: Bump version**
- Update all `go.mod` files version number
- Increment minor version: `v1.0.7` → `v1.0.8`

```bash
# Update all go.mod files
find . -name "go.mod" -type f | while read file; do
    sed -i 's/v1\.0\.7/v1.0.8/g' "$file"
done
```

**Step 4: Create version commit**
```bash
git add docs/CHANGELOG.md $(find . -name "go.mod" -type f)
git commit -m "chore: bump version to v1.0.8"
```

**Step 5: Create tags**
```bash
# Main project tag
git tag -a v1.0.8 -m "Release v1.0.8"

# Module tags (18 modules)
git tag -a modules/core/v1.0.8 -m "Release modules/core v1.0.8"
git tag -a modules/profiles/v1.0.8 -m "Release modules/profiles v1.0.8"
# ... repeat for all modules
```

**Step 6: Push to GitHub**
```bash
git push origin main
git push origin --tags
```

**Step 7: Verify compliance**
```bash
# Verify tags
git tag | grep v1.0.8

# Verify CHANGELOG
grep "v1.0.8" docs/CHANGELOG.md

# Verify go.mod versions
grep -h "require" $(find . -name "go.mod" -type f) | grep v1.0.8
```

### Rules That Must Be Followed

✅ **MUST DO:**
- Update CHANGELOG for every release
- Increment version number for every release
- Create tags for every release
- Push all tags and commits together

❌ **NEVER DO:**
- Tag without version bump
- Version bump without CHANGELOG update
- Tag without commit
- Manually edit CHANGELOG without version update
- Create tags out of order

### Violations and Consequences

- **Breaking rules**: Commit rejection / Rollback required
- **Unauthorized release**: PR rejection, version rollback
- **Data corruption**: Code review blocker until fixed

## Development Lifecycle

### 1. Planning
- Create issue or use existing issue
- Specify scope and requirements
- Link with related issues

### 2. Implementation
- Create feature branch from `main`
- Follow coding standards
- Add comprehensive tests
- Ensure no race conditions

### 3. Testing
- Unit tests: `go test ./modules/... -v`
- Race testing: `go test -race ./modules/...`
- Coverage target: 80%+

### 4. Code Review
- Push to GitHub
- Request review from maintainers
- Address review comments
- Ensure CI passes

### 5. Release
- Follow 7-step mandatory workflow
- Update version and CHANGELOG
- Create and push tags
- Verify all deployments

## Common Tasks

### Running Tests
```bash
# All tests
go test ./modules/...

# Specific module
go test ./modules/core/...

# With verbose output
go test ./modules/... -v

# With race detector
go test -race ./modules/...

# With coverage
go test -coverprofile=coverage.out ./modules/...
```

### Formatting Code
```bash
# Format all files
go fmt ./modules/...

# Fix imports
go mod tidy

# Check style
golangci-lint run
```

### Debugging
```bash
# Run with debug output
GODEBUG=gctrace=1 go run ./cmd/gateway

# Use delve debugger
go install github.com/go-delve/delve/cmd/dlv@latest
dlv debug ./cmd/gateway
```

## References

- [Go Modules Documentation](https://go.dev/wiki/Modules)
- [Keep a Changelog](https://keepachangelog.com/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Semantic Versioning](https://semver.org/)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
