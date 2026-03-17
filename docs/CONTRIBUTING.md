# Contributing Guide

Thank you for contributing code! Please follow these guidelines.

## Required Reading

- **[Developer Guide](./DEVELOPER_GUIDE.md)** - Version control rules, coding standards, module development
- **[Development Guide](./DEVELOPER_GUIDE.md)** - Complete reference

## Development Workflow

### 1. Clone the Repository

```bash
git clone https://github.com/vistone/fingerprint.git
cd fingerprint && go work sync
git checkout -b feature/your-feature
```

### 2. Development and Testing

```bash
# 1) Format check only (no auto-fix)
gofmt -s -l .

# 2) Lint
golangci-lint run

# 3) Test
go test ./... -race
```

### 2.1 Compliance Gates (Mixed Policy)

- `BLOCKING` (CI fail): `go test ./...`, `golangci-lint run`, English-only comment check, file/function/parameter threshold checks.
- `WARNING` (CI warn): coverage trend, benchmark regression signals, architecture optimizations from analysis reports.
- First-round enforcement scope: `modules/**/*.go` (non-test files) for threshold checks; oversized legacy test files are reported as advisory debt.
- CI output must clearly label `BLOCKING` and `WARNING`.

### 3. Commit Message Format

```
<type>(<scope>): <subject>

<body>
```

**Type:** `feat` / `fix` / `docs` / `refactor` / `test` / `chore`

**Example:** `feat(profiles): add Chrome 140 support`

### 4. Version Control (Mandatory)

See [Developer Guide - Version Control Rules](./DEVELOPER_GUIDE.md#version-control-rules-mandatory)

**7-Step Process:**
1. Code changes + tests pass
2. Update CHANGELOG.md
3. Update version number (minor +1)
4. Create version commit
5. Create version Tags
6. Push to GitHub
7. Verify compliance

**Must follow!** Violations result in commit rejection or rollback.

## Code Standards

### File Length Control

**Strictly enforced!** Code files must be short and single-responsibility.

| Metric | Limit | Note |
|--------|-------|------|
| Lines per file | ≤ 500 | Must split if exceeded |
| Lines per function | ≤ 80 | Must refactor if exceeded |
| Function parameters | ≤ 5 | Use a struct instead |

**Splitting guidelines:**

- Split by responsibility: each file should cover one logical domain
- Clear naming: split file names should reflect content (e.g., `handler_analysis.go`, `handler_defense.go`)
- Keep package consistency: split files stay in the same package, sharing package-level vars and types
- Split proactively: if new code pushes a file over 500 lines, split in the same PR

```bash
# Check for oversized files
find . -name "*.go" -not -path "*/vendor/*" | xargs awk 'END{if(NR>500) print FILENAME": "NR" lines"}' | sort -t: -k2 -rn
```

### Error Handling

```go
if err != nil {
    return fmt.Errorf("failed to get profile %q: %w", id, err)
}

var ErrInvalidProfile = errors.New("invalid profile")
```

### Concurrency Safety

```go
type Registry struct {
    mu       sync.RWMutex
    profiles map[string]Profile
}

func (r *Registry) Get(id string) (Profile, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    p, ok := r.profiles[id]
    return p, ok
}
```

### Logging

```go
import "log/slog"

slog.Debug("request", "profile", name)
```

## Testing

### Unit Tests

```go
func TestGet(t *testing.T) {
    tests := []struct {
        name string
        id   string
        ok   bool
    }{
        {"exist", "chrome_140", true},
        {"not", "unknown", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, ok := Get(tt.id)
            if ok != tt.ok {
                t.Errorf("want %v, got %v", tt.ok, ok)
            }
        })
    }
}
```

### Coverage

```bash
go test -coverprofile=coverage.out ./modules/...
go tool cover -html=coverage.out
```

## Module Development

### Create New Module

```bash
mkdir modules/newmod
cd modules/newmod

cat > go.mod << 'EOF'
module github.com/vistone/fingerprint/modules/newmod
go 1.25.7
require github.com/vistone/fingerprint/modules/core v1.0.11
replace github.com/vistone/fingerprint/modules/core => ../core
EOF

cd ../..
echo "./modules/newmod" >> go.work
go work sync
```

### Dependency Rules

- No circular dependencies
- core is a zero-dependency module
- Other modules can depend on core

## Pre-Submit Checklist

Before submitting:

- [ ] `gofmt -s -l .` (check only, no auto-fix) ✓
- [ ] `golangci-lint run` ✓
- [ ] `go test ./... -race` ✓
- [ ] All .go files ≤ 500 lines ✓
- [ ] All functions ≤ 80 lines ✓
- [ ] CHANGELOG.md updated
- [ ] Version number updated (sed)
- [ ] Version commit created
- [ ] Tags created (main project + 17 modules)
- [ ] Version audit passed

## Documentation

- [Developer Guide](./DEVELOPER_GUIDE.md)
- [Version Management](./VERSION_MANAGEMENT.md)
- [API](./API.md)
- [Architecture](./ARCHITECTURE.md)
- [Changelog](./CHANGELOG.md)

## Thank You

Thanks for contributing!
