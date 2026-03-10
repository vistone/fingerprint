# Security Policy

## Reporting Vulnerabilities

**Do not** report security vulnerabilities publicly on GitHub Issues.

Please email security@example.com with the following details:

1. Vulnerability description
2. Scope of impact
3. Reproduction steps
4. Suggested fix

### Response Timeline

- Critical vulnerabilities: Fixed within 1-2 weeks
- Medium vulnerabilities: Fixed within 2-4 weeks
- Low vulnerabilities: Fixed in next release

## Best Practices

### Users

- Keep updated: `go get -u github.com/vistone/fingerprint/modules/fingerprint`
- Subscribe to releases: https://github.com/vistone/fingerprint/releases
- Pin versions explicitly: `require ... v1.0.8`
- Audit dependencies: `go list -u -m all`

### Contributors

- All code changes require peer review
- New code must include unit tests
- Never hardcode secrets or sensitive information
- Remove debug logging before commit

## Known Issues

Current version (v1.0.8) has no known unresolved security issues.

## Dependencies

Primary dependencies:

| Package | Purpose | Status |
|---------|---------|--------|
| Go stdlib | Core | ✅ |
| Internal modules | Core + profiles | ✅ |

### Check for Vulnerabilities

```bash
# List all dependencies
go list -m all

# Check for known vulnerabilities (requires installation)
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

## Version Support

| Version | Status | Security Updates |
|---------|--------|-----------------|
| v1.0.8+ | Current | ✅ |
| v1.0.7 | Maintained | ✅ |
| v1.0.6 and earlier | Deprecated | ❌ |

## TLS Fingerprinting Library

This library is used for **generating and analyzing** TLS fingerprints, not for cryptographic operations.

### Features

- ✅ Does not handle or store private keys
- ✅ Does not implement cryptographic algorithms
- ✅ Only reads public values from TLS handshake

### Valid Use Cases

- ✅ Browser identification and traffic analysis
- ✅ Security monitoring and anomaly detection
- ❌ Evading security protections
- ❌ Malicious purposes

## Compliance

- Go coding standards
- BSD 3-Clause License
- Does not collect or store user data

## Audit

This project has not undergone third-party security audit. Please conduct your own assessment for sensitive applications.

## Resources

- [OWASP - Transport Layer Protection](https://owasp.org/www-community/attacks/Manipulator-in-the-middle_attack)
- [RFC 5246 - TLS 1.2](https://tools.ietf.org/html/rfc5246)
- [RFC 8446 - TLS 1.3](https://tools.ietf.org/html/rfc8446)

---

**最后更新：2026-03-10**  
**当前版本：v1.0.7** ✅
