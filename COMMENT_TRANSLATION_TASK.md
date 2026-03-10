# Comment Translation Task - English Only Policy

## Issue Summary

✅ **Rule Established**: All code comments MUST be in English (enforced in DEVELOPER_GUIDE.md)

⚠️ **Current State**: Codebase has 7,378 Chinese comment lines in 233 Go files

## Scope

| Priority | Category | Count | Files | Status |
|----------|----------|-------|-------|--------|
| **P0** | Core modules | 2,100+ | agent, core, gateway, ml, defense | ❌ Needs translation |
| **P1** | Client/TLS | 1,500+ | client, tls, http | ❌ Needs translation |
| **P2** | Config/Profiles | 1,200+ | config, profiles, frontend | ❌ Needs translation |
| **P3** | Testing/Utils | 1,000+ | *_test.go files, internal, generator | ❌ Needs translation |
| **P4** | Examples/Tools | 600+ | cmd/*, examples/* | ❌ Needs translation |

**Total Effort**: ~7,378 lines of comments to translate

## Translation Rule

### BEFORE (❌ NOT ALLOWED)
```go
// Package client 提供完整的浏览器指纹模拟客户端
// 从 TCP/IP 层到 TLS 层到 HTTP 层的全栈模拟

// 超时常量定义
const (
    TimeoutDialConnect = 10 * time.Second  // 建立连接超时
)

// 创建智能传输层（支持 HTTP/2 → HTTP/1.1 回退）
transport := NewSmartTransport()
```

### AFTER (✅ REQUIRED)
```go
// Package client provides a complete browser fingerprint simulation client
// Full-stack simulation from TCP/IP to TLS to HTTP layer

// Universal timeout constants
const (
    TimeoutDialConnect = 10 * time.Second  // TCP connection establishment timeout
)

// Create smart transport layer (supports HTTP/2 → HTTP/1.1 fallback)
transport := NewSmartTransport()
```

## Translation Tools

### 1. Automated Scanner (Ready)
```bash
# Scan for Chinese comments
python3 scripts/translate_comments.py --scan

# Show files needing translation
python3 scripts/translate_comments.py --scan | grep "lines):"
```

### 2. Auto-Translate Tool (Limited capability)
```bash
# Attempt auto-translation (requires manual review)
python3 scripts/translate_comments.py --fix

# Translate specific file
python3 scripts/translate_comments.py --file modules/core/core.go
```

## Implementation Strategy

⚠️ **UPDATE**: Auto-translation has quality issues with complex technical comments. 
Using **hybrid approach**: Manual translation for critical files + careful English style.

### Phase 1: Core Modules (P0) - Week 1-2 (PRIORITY)
- [ ] modules/agent/agent.go - 80 lines (Core OADA architecture)
- [ ] modules/agent/knowledge.go - 79 lines (Knowledge base)
- [ ] modules/client/client.go - 60 lines (Client implementation)
- [ ] modules/core/pool.go - 77 lines (Connection pool)
- [ ] modules/gateway/gateway.go - 144 lines (Gateway logic)

**Method**: 
Support from:
1. Use ChatGPT/Claude for translating specific comment blocks
2. Developer manually applies translation to code
3. Code review for accuracy and English quality
4. Commit with proper message

### Phase 2: Client & TLS (P1) - Week 2
- [ ] modules/client/* - 90 lines
- [ ] modules/tls/* - 250 lines
- [ ] modules/http/* - 300 lines

### Phase 3: Config & Profiles (P2) - Week 3
- [ ] modules/config/* - 100 lines
- [ ] modules/profiles/* - 400 lines
- [ ] modules/frontend/* - 120 lines

### Phase 4: Tests & Utils (P3) - Week 4
- [ ] *_test.go files - 800 lines
- [ ] modules/internal/* - 150 lines
- [ ] modules/generator/* - 100 lines

### Phase 5: Examples & Tools (P4) - Week 5
- [ ] cmd/* - 300 lines
- [ ] examples/* - 200 lines

## Validation Checklist

After translation, verify:
- [ ] No Chinese characters in comments (`// `, `/* */`, `* `)
- [ ] All comments are grammatically correct English
- [ ] Technical accuracy is preserved
- [ ] Code functionality unchanged
- [ ] Tests still pass: `go test ./modules/... -v`
- [ ] Build succeeds: `go build ./modules/fingerprint`

## Commands for Each Phase

### Before starting phase:
```bash
cd /media/stone/data1/fingerprint
git checkout -b translate/phase-N

# Get list of files for phase N
python3 scripts/translate_comments.py --scan | grep "modules/<module>" | head -20
```

### For each file:
```bash
# 1. Review current Chinese comments
cat modules/<module>/<file>.go | grep "//" | grep -E "[\u4e00-\u9fff]"

# 2. Translate (manual editing recommended for quality)
# Edit file in VS Code

# 3. Verify no Chinese remains
grep -n "[\u4e00-\u9fff]" modules/<module>/<file>.go || echo "✅ Clean"

# 4. Test affected code
go test ./modules/<module>/... -v
```

### After phase completion:
```bash
# Verify entire phase
python3 scripts/translate_comments.py --scan | grep "modules/<module>"

# Commit phase
git add modules/<module>/*.go
git commit -m "refactor(comments): translate Chinese comments to English in modules/<module>"

# Create PR and merge
git push origin translate/phase-N
```

## File Lists by Priority

### P0 Files (Critical)
```
modules/agent/agent.go (80 lines with Chinese)
modules/agent/knowledge.go (79 lines)
modules/agent/strategy.go (35 lines)
modules/agent/behavior.go (29 lines)
modules/agent/memory.go (15 lines)
modules/agent/anomaly.go (43 lines)
modules/agent/agent_test.go (22 lines)

modules/core/logger.go (varies)
modules/core/pool.go (varies)
modules/core/types.go (varies)
modules/core/constants.go (varies)

modules/gateway/gateway.go (200+ lines)
modules/gateway/cache.go (varies)
modules/gateway/breaker.go (varies)

modules/ml/classifier.go (150+ lines)
modules/ml/extractor.go (varies)

modules/defense/detector.go (180+ lines)
```

### Automation Notes

The `translate_comments.py` script provides:
- ✅ Scanning for Chinese comments
- ⚠️ Basic pattern-based translation (unreliable for complex comments)
- ❌ AI/API-based translation (not implemented - requires API key)

**Recommendation**: Use auto-translator as a starting point, then manually review and correct each comment for accuracy and English grammar.

## Quality Standards

✅ **Good English Comments**:
```go
// Validate that TLS cipher suites match browser fingerprint
// Check if client hello contains expected TLS extensions
// Calculate risk score based on anomalies detection
// Memory cleanup runs every 5 minutes to remove stale entries
// Implements the OADA (Observe-Analyze-Decide-Act) decision loop
```

❌ **Bad English Comments** (to avoid):
```go
// Validate TLS cipher suites with browser fingerprint (awkward)
// Check the client hello contain expected TLS extensions (grammatically wrong)
// Calculate risk based on anomaly (incomplete sentence)
// Memory cleanup runs in 5 minutes intervals (incorrect)
// Implements OADA process loop (unclear acronym usage)
```

## Timeline

| Phase | Week | Files | Comments | Status |
|-------|------|-------|----------|--------|
| P0 (Core) | 1 | 30+ | 800+ | ⏳ Ready to start |
| P1 (Client/TLS) | 2 | 50+ | 1,500+ | ⏳ Waiting for P0 |
| P2 (Config) | 3 | 40+ | 1,200+ | ⏳ Waiting for P1 |
| P3 (Tests) | 4 | 100+ | 1,000+ | ⏳ Waiting for P2 |
| P4 (Examples) | 5 | 20+ | 600+ | ⏳ Waiting for P3 |

**Total Effort**: ~3-4 weeks for manual review approach, 1 week if using AI translation API

## Next Steps

1. ✅ Document the rule: DEVELOPER_GUIDE.md updated
2. ✅ Create scanner tool: translate_comments.py ready
3. ⏳ Start Phase 1: Agent module translation
4. ⏳ Review and validate translations
5. ⏳ Update CI to reject Chinese comments
6. ⏳ Complete all 5 phases
7. ⏳ Final validation and merge all changes

## Support

For questions or blockers:
- Check script usage: `python3 scripts/translate_comments.py --help`
- Use VS Code search-replace: `Ctrl+H` with regex enabled
- Reference [DEVELOPER_GUIDE.md](./docs/DEVELOPER_GUIDE.md) for comment standards
