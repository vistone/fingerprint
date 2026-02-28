# Contributing to Fingerprint

Thank you for your interest in contributing to the Fingerprint project! This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct. Please be respectful and constructive in all interactions.

## Getting Started

### Prerequisites

- Go 1.24 or later (we test on 1.24 and 1.25)
- Git
- Make (optional, but recommended)

### Development Environment Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/vistone/fingerprint.git
   cd fingerprint
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Install development tools** (optional)
   ```bash
   make install-tools
   ```

4. **Verify setup**
   ```bash
   go test ./...
   ```

## Development Workflow

### 1. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
# or for bug fixes
git checkout -b fix/your-bug-fix
```

Branch naming convention:
- `feature/` for new features
- `fix/` for bug fixes
- `docs/` for documentation changes
- `refactor/` for code refactoring
- `perf/` for performance improvements

### 2. Make Changes

Follow our development standards:

1. **Code Quality**
   - Run `make format` to format code
   - Run `make lint` to check code quality
   - Run `make test` to verify tests pass
   - Ensure `go vet ./...` passes

2. **Documentation**
   - Add/update comments for public APIs
   - Update CHANGELOG.md with your changes
   - Update relevant documentation files

3. **Tests**
   - Write tests for new functionality
   - Ensure all tests pass
   - Aim for 80%+ code coverage

### 3. Commit Changes

Write clear commit messages following this format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type** (required):
- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring without feature changes
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `chore`: Other changes that don't affect code

**Scope** (optional):
- `profiles`: Changes to browser profiles
- `ja3`: JA3 fingerprinting changes
- `ja4`: JA4 fingerprinting changes
- `defense`: Anomaly/contradiction detection
- `api`: Public API changes

**Subject** (required):
- Use imperative mood ("add" not "added")
- Don't capitalize first letter
- No period at the end
- Limit to 50 characters

**Examples**:
```
feat(profiles): add Safari 18 browser profile
fix(ja3): handle GREASE values correctly
docs: update API documentation
test(defense): add anomaly detection tests
```

### 4. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

Then create a Pull Request on GitHub with:

1. **Clear title** following commit message format
2. **Description** explaining:
   - What changes were made
   - Why the changes were made
   - Any related issues (use `Closes #123`)
3. **Screenshots/examples** if applicable
4. **Checklist** (automatically provided by PR template)

## Code Review Process

1. **Automated Checks**
   - GitHub Actions will run tests, linting, and coverage
   - All checks must pass

2. **Manual Review**
   - At least one maintainer will review the code
   - May request changes or ask questions
   - Be responsive to feedback

3. **Approval and Merge**
   - Once approved and checks pass, PR will be merged
   - Maintainer will handle merging

## Quality Standards

### Code Quality Requirements

- ✅ All tests pass (`go test ./...`)
- ✅ No go vet warnings (`go vet ./...`)
- ✅ Code is formatted (`gofmt -s -w .`)
- ✅ Linting passes (`make lint`)
- ✅ No security issues (`gosec ./...`)
- ✅ Test coverage maintained or improved

### Documentation Requirements

- ✅ Public functions/types have godoc comments
- ✅ Complex logic has inline comments
- ✅ CHANGELOG.md is updated
- ✅ README.md is updated if needed
- ✅ Examples in docs/examples are valid

### Performance Requirements

- ✅ No performance regression
- ✅ Benchmarks pass (`make benchmark`)
- ✅ Memory usage is acceptable
- ✅ Concurrent performance is maintained

## Testing Guidelines

### Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test ./... -v

# Run specific test
go test ./test -run TestSpecificName

# Run with coverage
go test ./... -cover

# Run benchmarks
go test ./test -bench=. -benchmem
```

### Writing Tests

1. **Test file naming**: `*_test.go`
2. **Test function naming**: `TestFunctionName`
3. **Table-driven tests**: Use for multiple scenarios
4. **Test organization**: Group related tests

Example:
```go
func TestGetRandomFingerprint(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"valid", false},
		{"no fingerprints", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetRandomFingerprint()
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error")
			}
			if result == nil && !tt.wantErr {
				t.Errorf("unexpected nil result")
			}
		})
	}
}
```

## Documentation

### Types of Documentation

1. **API Documentation**: Godoc comments in code
2. **User Guides**: `docs/2-guides/`
3. **Reference Docs**: `docs/3-references/`
4. **Development Guides**: `docs/5-process/development/`

### Writing Documentation

1. Use clear, simple English
2. Include code examples
3. Explain why, not just how
4. Keep it up-to-date with code
5. Link to related documentation

## Reporting Issues

### Bug Reports

Include:
- Go version
- Operating system
- Steps to reproduce
- Expected behavior
- Actual behavior
- Code example (if applicable)

### Feature Requests

Include:
- Use case description
- Why it's needed
- Proposed API/interface
- Examples of similar features

## Security

### Reporting Security Issues

⚠️ **Do not open a public issue for security vulnerabilities**

Please report security issues privately to: security@example.com

Include:
- Type of vulnerability
- Location in codebase
- Potential impact
- Suggested fix (if any)

See [SECURITY.md](./SECURITY.md) for more details.

## Additional Resources

- 📖 [Development Guidelines](./docs/5-process/development/README.md)
- 🚀 [Project Setup Guide](./docs/2-guides/01-quick-start.md)
- 📋 [Developer Checklist](./docs/2-guides/developer/01-developer-checklist.md)
- 🔍 [Go Development Rules](./docs/5-process/development/00-go-development-rules.md)

## Questions?

- 📧 Email: support@example.com
- 💬 Discussions: GitHub Discussions
- 📝 Documentation: See docs/README.md

---

## License

By contributing to this project, you agree that your contributions will be licensed under the same license as the project.

Thank you for contributing! 🎉
