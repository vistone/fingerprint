# 提交前检查清单 (Pre-Commit Checklist)

> **本文件是快速参考。完整规则见 [docs/DEVELOPER_GUIDE.md](./docs/DEVELOPER_GUIDE.md#-版本控制规则强制执行)**

## ⚠️ 核心规则：不能乱来

每次提交都必须：
1. ✅ 更新 CHANGELOG.md
2. ✅ 更新所有 go.mod 版本号
3. ✅ 创建版本 tag
4. ✅ 推送到 GitHub

**违反规则 = 提交拒绝或回滚**

---

## 提交前 5 分钟检查

### ① 代码质量检查
- [ ] 所有测试通过: `go test ./modules/... -race`
- [ ] 代码格式化: `go fmt ./...`
- [ ] Lint 通过: `golangci-lint run`
- [ ] 没有 TODO/FIXME 遗漏

### ② 版本号准备 (版本 = v1.0.5 → v1.0.6 示例)

当前版本可以通过以下方式查看：

```bash
grep "module github.com/vistone/fingerprint$" go.mod | tail -1
# 看 require github.com/vistone/fingerprint/modules/* 后面的版本
```

**更新 CHANGELOG.md:**
```markdown
将这个：
## [Unreleased]

改为这个（现在的日期）：
## [v1.0.6] - 2026-03-10

然后添加内容：
### Added
### Fixed
### Changed
```

- [ ] 已在 `docs/CHANGELOG.md` 中更新版本号和日期

**更新版本号:**

```bash
# 查看当前版本
grep -o "v[0-9]\.[0-9]\.[0-9]" go.mod | head -1

# 假设当前版本是 v1.0.5，新版本是 v1.0.6
# 执行：
sed -i 's/v1\.0\.5/v1.0.6/g' go.mod modules/*/go.mod

# 验证：
grep "v1.0.6" go.mod modules/*/go.mod | wc -l
# 应该显示 ~20（1 主项目 + 19 个模块）
```

- [ ] 已运行 sed 命令更新所有版本号
- [ ] 已验证约 20 个文件包含新版本号

### ③ 提交
```bash
# 提交 CHANGELOG 和版本号更新
git add docs/CHANGELOG.md go.mod modules/*/go.mod
git commit -m "chore: Release v1.0.6"

# 创建特定的 tag
git tag -a v1.0.6 -m "Release v1.0.6"

# 创建所有模块的 tag
for module in agent client config core defense errors fingerprint frontend gateway generator http internal kit metrics ml network plugin profiles tls; do
    [ -f "modules/$module/go.mod" ] && git tag -a modules/$module/v1.0.6 -m "Release modules/$module v1.0.6" || true
done
```

- [ ] 已提交版本更新 (commit hash: ______)
- [ ] 已创建主项目 tag: v1.0.6
- [ ] 已创建所有模块 tags (共 18 个)

### ④ 推送到 GitHub

```bash
git push origin main
git push origin --tags
```

- [ ] 已推送 main 分支
- [ ] 已推送所有 tags

### ⑤ 最终验证

```bash
# 查看 GitHub 上是否显示新的 tag
git tag -l | grep v1.0.6

# 查看最新提交
git log -1 --oneline
# 应该显示: xxxxx (HEAD -> main, tag: v1.0.6) chore: Release v1.0.6
```

- [ ] GitHub 上显示新 tags
- [ ] GitHub 上显示新提交
- [ ] CHANGELOG 在 GitHub 上已更新

---

## 常见错误

| ❌ 错误 | ✅ 纠正方法 |
|--------|-----------|
| 忘记更新 CHANGELOG | `git revert HEAD && git push` 后重新提交 |
| 版本号更新不一致 | 检查 `grep v1.0.6 go.mod modules/*/go.mod \| wc -l` 是否 = ~20 |
| 忘记创建 tags | `git tag -a v1.0.6...` 补创后 `git push origin --tags` 推送 |
| 只有部分 tags | 检查所有 18 个模块是否都有 tag（`git tag -l \| wc -l`） |
| 本地和 GitHub 不同步 | `git pull origin main` 后 `git push origin main --tags` |

---

## 快速命令参考

### 查看当前版本
```bash
grep -o "v[0-9]\.[0-9]\.[0-9]" go.mod | head -1
```

### 检查版本一致性
```bash
# 所有 go.mod 中的版本应该都一样
grep "github.com/vistone/fingerprint/modules" modules/*/go.mod | grep -o "v[0-9.]*" | sort | uniq -c
# 输出应该是一行: ~18 v1.0.X （都是同一版本）
```

### 列出所有 tags
```bash
git tag -l | sort -V | tail -20  # 最新的 20 个
```

### 查看最新提交
```bash
git log -1 --oneline --all
```

### 检查 CHANGELOG 格式
```bash
head -5 docs/CHANGELOG.md
# 应该显示: ## [v1.0.6] - 2026-03-10
```

---

## 必读文档

- **完整规则**: [docs/DEVELOPER_GUIDE.md - 版本控制规则](./docs/DEVELOPER_GUIDE.md#-版本控制规则强制执行)
- **版本管理**: [docs/VERSION_MANAGEMENT.md](./docs/VERSION_MANAGEMENT.md)
- **提交规范**: [docs/DEVELOPER_GUIDE.md - 提交规范](./docs/DEVELOPER_GUIDE.md#提交规范)

---

## 联系人

如有版本控制问题，请参考上述文档或联系项目维护人员。

**记住：按规则提交，不能乱来！** 🔒
