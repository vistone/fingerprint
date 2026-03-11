# 版本管理策略 / Version Management Strategy

## 📋 概述 / Overview

本文档规定 Fingerprint 项目的版本控制流程，确保所有模块版本统一、可追溯。

## 🔄 版本号规范 / Versioning Scheme

采用 **语义化版本 (Semantic Versioning)**：`MAJOR.MINOR.PATCH`

### 版本号含义

| 位置 | 名称 | 变化规则 | 示例 |
|------|------|---------|------|
| Major | 主版本号 | 重大功能变化、不兼容更新 | v2.0.0 |
| Minor | 次版本号 | **每次 GitHub 提交时递增** | v1.0.10 → v1.0.11 |
| Patch | 修补版本号 | 仅内部构建/热修复 | v1.0.10-1 |

### 当前版本

- **主项目**: v1.0.11
- **所有模块**: v1.0.11（统一）

## 📤 提交与 Tag 工作流 / Commit & Tagging Workflow

### Step 1: 提交代码

```bash
cd /media/stone/data1/fingerprint

# 添加所有改动
git add -A

# 编写有意义的提交信息
# 格式: feat/fix: [模块] 功能描述
git commit -m "feat: [frontend] Add i18n support for English/Chinese language switching"
```

### Step 2: 创建主项目 Tag

```bash
# 为主项目打 tag（格式: v版本号）
git tag -a v1.0.11 -m "Release v1.0.11"
```

### Step 3: 创建模块 Tags

**仅需要为有改动的模块创建 tag**

```bash
# 格式: modules/[模块名]/v版本号

# 例如：只有 frontend/gateway 有改动
git tag -a modules/frontend/v1.0.11 -m "Release modules/frontend v1.0.11"
git tag -a modules/gateway/v1.0.11 -m "Release modules/gateway v1.0.11"
```

### Step 4: 推送到 GitHub

```bash
# 推送主分支
git push origin main

# 推送所有 tags
git push origin --tags
```

### Step 5: 验证

```bash
# 列出所有 v1.0.11 tags
git tag -l | grep v1.0.11

# 查看 tag 详情
git show v1.0.11
```

## 📝 更新 CHANGELOG 规范

每次发布必须更新 `docs/CHANGELOG.md`：

```markdown
## [v1.0.11] - 2026-03-10

### Added
- **功能标题**: 功能描述
  - 更新项 A
  - 更新项 B

### Fixed
- **Bug 标题**: Bug 修复描述

### Changed
- 改进项
```

**规则**：
- 日期格式：YYYY-MM-DD
- 分类：Added / Changed / Fixed / Removed / Deprecated / Security
- 描述：中文 (description) 或 English (description)

## 🔍 版本一致性检查

### 检查本地版本

```bash
# 检查所有 go.mod 中的模块版本
grep "github.com/vistone/fingerprint/modules" modules/*/go.mod | grep -o "v[0-9.]*" | sort | uniq

# 应该全部输出: v1.0.11（或当前版本）
```

### 检查 GitHub Tags

```bash
# 列出所有 tags
git tag -l | head -50

# 列出特定版本的所有 tags
git tag -l | grep "v1.0.11"
```

## 🚀 版本升级步骤

### 场景 1：优先修复 (Patch 升级)

```bash
# v1.0.10 → v1.0.10-1
git commit -m "fix: [core] Fix critical bug in XYZ"
```

### 场景 2：新功能发布 (Minor 升级)

```bash
# v1.0.10 → v1.0.11（下一个版本）
git commit -m "feat: [frontend] Add i18n multi-language support"

# 更新 go.mod 中所有模块版本为 v1.0.11
sed -i 's/v1.0.10/v1.0.11/g' go.mod modules/*/go.mod

git add -A
git commit -m "chore: Bump version to v1.0.11"
```

### 场景 3：重大改变 (Major 升级)

```bash
# v1.0.11 → v2.0.0
# 修改 go.mod 中所有版本为 v2.0.0
# 更新 CHANGELOG 中的 [Unreleased] 为 [v2.0.0]
# 执行完整的提交和 tag 流程
```

## ⚠️ 常见陷阱 / Common Pitfalls

| 问题 | 原因 | 解决 |
|------|------|------|
| Docker 构建时显示 v0.0.0 | GitHub 上没有对应的 tag | 创建并推送 tags 到 GitHub |
| 模块版本不一致 | go.mod 文件未统一更新 | 使用 sed 批量更新所有文件 |
| tag 推送失败 | 未使用 `git push origin --tags` | 显式推送 tags：`git push origin [tag-name]` |
| replace 指令被忽略 | Docker 构建中的 go.work 配置 | 在 Dockerfile 后复制 go.work 和完整源代码 |

## 📌 快速参考 / Quick Reference

```bash
# 完整的发布流程（3 分钟）
git add -A
git commit -m "feat: [模块] 功能描述"

# 创建 main tag
git tag -a vX.Y.Z -m "Release vX.Y.Z"

# 创建受影响模块的 tags
git tag -a modules/frontend/vX.Y.Z -m "Release modules/frontend vX.Y.Z"
git tag -a modules/gateway/vX.Y.Z -m "Release modules/gateway vX.Y.Z"

# 推送
git push origin main
git push origin --tags

# 验证
git tag -l | grep vX.Y.Z
```

---

**最后更新**: 2026-03-10
**负责人**: DevOps Team
**生效日期**: 2026-03-10
