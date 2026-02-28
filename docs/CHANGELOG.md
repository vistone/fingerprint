# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive Go development standards and style guidelines
- Go Vet composites warnings handling with //nolint directives
- GitHub Actions CI/CD workflows for automated testing and linting
- Makefile for common development tasks
- Enhanced documentation system with 5-tier classification

### Changed
- Improved project structure and organization
- Enhanced documentation with complete API comments
- Updated Go dependencies to latest versions

### Fixed
- Fixed Go Vet composites warnings from external utls library
- Code formatting standardization with gofmt

---

## [2.0.1] - 2026-02-28

### Added
- Edge browser fingerprint profiles
- JA3/JA4 fingerprint calculation methods
- Passive browser recognition from HTTP headers
- Anomaly detection for automated tools
- Noise injection capabilities
- Comprehensive error fixes and documentation

### Changed
- Improved code quality and standards
- Enhanced project documentation
- Better error handling throughout codebase

### Fixed
- Resolved Go Vet warnings with proper struct field initialization
- Fixed code formatting issues

---

## [2.0.0] - 2026-02-01

### Added
- Complete browser TLS fingerprint library
- Support for 70+ browser fingerprint profiles
- User-Agent generation for different browsers and OS combinations
- HTTP Headers generation matching browser specifications
- JA3 fingerprint calculation and matching
- Passive browser recognition capabilities
- Anomaly detection for suspicious fingerprints
- Contradiction detection for inconsistent attributes
- Comprehensive test suite with 90%+ coverage

### Changed
- Major rewrite of fingerprint matching engine
- Improved performance with zero-allocation functions
- Enhanced API design for better usability

### Fixed
- Various security issues in fingerprint validation
- Performance bottlenecks in matching algorithms

---

## [1.0.2] - 2025-12-15

### Added
- Support for Safari browser fingerprints
- Additional test cases for edge cases
- Performance optimization for fingerprint matching

### Fixed
- Bug in User-Agent parsing for mobile devices
- Memory leak in fingerprint caching
- Compatibility issues with Go 1.24

---

## [1.0.1] - 2025-11-20

### Added
- Firefox browser fingerprint support
- Opera browser support
- Improved documentation

### Fixed
- Chrome User-Agent generation issues
- Header ordering problems
- TLS version negotiation fixes

---

## [1.0.0] - 2025-10-01

### Added
- Initial release of fingerprint library
- Core TLS fingerprint matching
- User-Agent generation
- HTTP Headers generation
- Chrome and Edge browser support
- Basic fingerprint database

### Features
- Zero-allocation random selection
- High-performance fingerprint matching
- Thread-safe concurrent access
- Comprehensive error handling
- Full test coverage

---

## Support

- 📖 [Documentation](./docs/README.md)
- 🐛 [Issue Tracker](../../issues)
- 💬 [Discussions](../../discussions)
- 📧 [Email Support](mailto:support@example.com)

---

## Versioning

This project follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version when you make incompatible API changes
- **MINOR** version when you add functionality in a backwards-compatible manner
- **PATCH** version when you make backwards-compatible bug fixes

---

## Links

- [GitHub Repository](../../)
- [Package Documentation](https://pkg.go.dev/github.com/vistone/fingerprint)
- [Contributing Guidelines](./CONTRIBUTING.md)
- [Security Policy](./SECURITY.md)
- [License](./LICENSE)
