# Version Management Strategy

## 📋 Overview

This document defines the version control workflow for the Fingerprint project, ensuring all modules are versioned consistently and traceably.

## 🔄 Versioning Scheme

Uses **Semantic Versioning (Semantic Versioning)**: `MAJOR.MINOR.PATCH`

### Version Number Meaning

| Position | Name | Change Rule | Example |
|----------|------|------------|---------|
| Major | Major version | Breaking changes, incompatible updates | v2.0.0 |
| Minor | Minor version | **Incremented on each GitHub commit** | v1.0.7 → v1.0.8 |
| Patch | Patch version | Internal build/hotfix only | v1.0.8-1 |

### Current Version

- **Main project**: v1.0.11
- **All modules**: v1.0.11 (unified)

## 📤 Commit & Tagging Workflow

### Step 1: Commit Code

```bash
cd /media/stone/data1/fingerprint

# Add all changes
git add -A

# Write meaningful commit message
# Format: feat/fix: [module] feature description
git commit -m "feat: [frontend] Add i18n support for English/Chinese language switching"
```

### Step 2: Create Main Project Tag

```bash
# Tag main project (format: v<version>)
git tag -a v1.0.8 -m "Release v1.0.8"
```

### Step 3: Create Module Tags

**Only create tags for modified modules**

```bash
# Module tag format: modules/<module>/v<version>
git tag -a modules/core/v1.0.8 -m "Release modules/core v1.0.8"
git tag -a modules/profiles/v1.0.8 -m "Release modules/profiles v1.0.8"
git tag -a modules/gateway/v1.0.8 -m "Release modules/gateway v1.0.8"
git tag -a modules/ml/v1.0.8 -m "Release modules/ml v1.0.8"

# ... repeat for all modified modules
```

### Step 4: Push to GitHub

```bash
# Push commits and tags to main branch
git push origin main
git push origin --tags
```

### Step 5: Verify Version Consistency

```bash
# 1. Verify all go.mod files have correct version
grep "require github.com/vistone/fingerprint/modules" \
    modules/*/go.mod | grep v1.0.8

# 2. Verify tags exist
git tag | grep v1.0.8

# 3. Verify CHANGELOG is updated
grep "## \[v1.0.8\]" docs/CHANGELOG.md

# 4. Verify all modules are versioned
find . -name "go.mod" -type f -exec grep "^go 1" {} \;
```

## 📝 CHANGELOG Management

### Update CHANGELOG Before Release

1. **Add new section** at the top:
   ```markdown
   ## [v1.0.8] - YYYY-MM-DD

   ### Added
   - New feature description

   ### Changed
   - Modified feature description

   ### Fixed
   - Bug fixes
   ```

2. **Move unreleased changes** from `[Unreleased]` section

3. **Keep historical versions** for reference

4. **Format**: Follow [Keep a Changelog](https://keepachangelog.com/) specification

## 🔑 Key Rules

### MUST DO ✅

- [ ] Update CHANGELOG before every release
- [ ] Increment MINOR version in all go.mod files
- [ ] Create tags for every release
- [ ] Push tags immediately after commits
- [ ] Verify version consistency before release
- [ ] Document changes in CHANGELOG

### NEVER DO ❌

- [ ] Create tags without version bumps
- [ ] Bump version without updating CHANGELOG
- [ ] Create tags without commits
- [ ] Manually edit CHANGELOG without version update
- [ ] Skip module tags for releases
- [ ] Create tags out of order

## ⚙️ Version Sync Across Modules

### All Modules Stay in Sync

```plaintext
Main project version = All module versions

v1.0.8 (main)
    ├── modules/core/v1.0.8
    ├── modules/profiles/v1.0.8
    ├── modules/tls/v1.0.8
    ├── modules/http/v1.0.8
    ├── modules/ml/v1.0.8
    ├── modules/gateway/v1.0.8
    └── ... (all 18 modules)
```

### Rational

- Simplifies dependency management
- Clear version history
- Easy to track which features are in which version
- Prevents version mismatch bugs

## 📊 Version History

```
v1.0.0 → v1.0.1 → v1.0.2 → ... → v1.0.8
```

Each version increment represents a set of changes committed to main branch.

## 🚀 Release Process

### Complete Release Checklist

```bash
# 1. Create feature branch
git checkout -b feature/your-feature

# 2. Make changes
# ... code changes here ...
git add -A
git commit -m "feat: [module] description"

# 3. Update CHANGELOG.md
# ... edit CHANGELOG.md ...

# 4. Update version in all go.mod files
find . -name "go.mod" -type f | while read file; do
    sed -i 's/v1\.0\.7/v1.0.8/g' "$file"
done

# 5. Create version commit
git add docs/CHANGELOG.md $(find . -name "go.mod" -type f)
git commit -m "chore: bump version to v1.0.8"

# 6. Create main project tag
git tag -a v1.0.8 -m "Release v1.0.8"

# 7. Create module tags (18 modules)
for module in core profiles tls http ml defense frontend gateway \
              generator network internal config plugin fingerprint \
              agent errors kit metrics scanner client tracer; do
    git tag -a modules/$module/v1.0.8 -m "Release modules/$module v1.0.8"
done

# 8. Push everything
git push origin main --tags

# 9. Verify
git tag | grep v1.0.8 | wc -l  # Should show 19 tags
```

## 🔍 Verification Commands

### Check Current Version

```bash
# Check repository version
git describe --tags

# Check go.mod versions
grep "require github.com/vistone/fingerprint" go.mod

# Check all module versions
find . -name "go.mod" -type f -exec grep "^module" {} \;
```

### Check Release Status

```bash
# List all tags
git tag -l | grep v1.0.8

# Show tag details
git show v1.0.8

# Compare versions
git log v1.0.7..v1.0.8 --oneline
```

### Validate Version Consistency

```bash
# Verify all modules have matching version
for file in $(find . -name "go.mod" -type f); do
    version=$(grep "require github.com/vistone/fingerprint" "$file" | grep -o "v[0-9.]*$")
    echo "$file: $version"
done

# Verify CHANGELOG matches version
grep "^\## \\[v1.0.8\\]" docs/CHANGELOG.md
```

## 📌 Module List (18 Modules)

1. core - Core types (zero dependencies)
2. profiles - Fingerprint profiles
3. tls - TLS fingerprint analysis
4. http - HTTP fingerprint analysis
5. ml - Machine learning classifier
6. defense - Security defense system
7. frontend - Frontend SDK
8. gateway - API Gateway
9. generator - Fingerprint generator
10. network - Network layer analysis
11. internal - Internal utilities
12. config - Configuration management
13. plugin - Plugin system
14. agent - Autonomous security agent
15. errors - Error handling
16. kit - Utility toolkit
17. metrics - Prometheus metrics
18. scanner - Security scanner

## 🎯 GitHub Release Notes

When creating a GitHub Release:

1. **Title**: `Release v1.0.8`
2. **Target**: Select tag `v1.0.8`
3. **Description**: Copy from CHANGELOG.md
4. **Binary**: Attach compiled binaries

```markdown
# Release v1.0.8

## Changes

- Simplify CONTRIBUTING.md and SECURITY.md documentation
- Update version control rules
- Enhance version management consistency

See [CHANGELOG.md](./docs/CHANGELOG.md) for full details.
```

## 🔗 Related Documents

- [Developer Guide](./DEVELOPER_GUIDE.md) - Development workflow
- [Contributing Guide](./CONTRIBUTING.md) - Contributing guidelines
- [Changelog](./CHANGELOG.md) - Version history
- [Architecture](./ARCHITECTURE.md) - Project architecture

## ❓ FAQ

**Q: How often should we release?**
A: With each meaningful commit to main branch.

**Q: Should patch version be used?**
A: Only for internal builds and hotfixes outside release cycle.

**Q: What if a module hasn't changed?**
A: Still create a tag to keep versions in sync.

**Q: How do I revert a version?**
A: Create a new version with revert commits, don't delete tags.

**Q: Can I skip a version number?**
A: No, versions must be sequential: v1.0.7 → v1.0.8 → v1.0.9
