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
go test ./modules/... -race
go fmt ./...
golangci-lint run
```

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

- [ ] `go test ./modules/... -race` ✓
- [ ] `go fmt ./...` ✓
- [ ] `golangci-lint run` ✓
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
