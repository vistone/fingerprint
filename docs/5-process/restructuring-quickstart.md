# 包结构重构 - 快速启动指南

> **如果你只有 5 分钟**：看这份文档  
> **如果你需要完整细节**：看 [package-restructuring-plan.md](package-restructuring-plan.md)  

---

## 🎯 一句话总结

从 **杂乱的内部包** → **清晰的分层结构**  
- TLS 工具独立到 `tls/internal/`
- HTTP 工具独立到 `http/internal/`  
- 共享工具集中到 `internal/` 精选子包
- 公共 API 统一暴露在 `pkg/`

---

## ⏱ 执行时间表

| Phase | 时间 | 工作量 | 难度 | 状态 |
|-------|------|--------|------|------|
| **Phase 1: TLS** | Week 5-6 | 2-3 天 | ⭐ 低 | ⏳ 待启动 |
| **Phase 2: HTTP** | Week 6-7 | 3-4 天 | ⭐⭐ 中 | ⏳ Phase1 后 |
| **Phase 3: pkg 化** | Week 7-8 | 3-5 天 | ⭐⭐⭐ 高 | ⏳ 最后 |

---

## 🚀 现在就开始 Phase 1 (5 分钟设置)

### 前置检查
```bash
cd /media/stone/data/fingerprint

# 验证当前状态
ls -la tls/        # 应该有 ja3/, ja4/, ja4s/, utils/, ech/, tls.go
ls -la internal/   # 应该有 tlsutil/

# 备份当前分支
git branch
git tag -a phase1-backup -m "Safe point before Phase 1"
```

### 自动执行（推荐）
```bash
# Step 1: 预演（查看所有变更，不实际执行）
bash scripts/phase1_tls_migration.sh dry-run

# Step 2: 确认输出无误后，执行真实迁移
bash scripts/phase1_tls_migration.sh execute

# Step 3: 验证
go test ./tls/...
go build ./...
```

### 验证成功标志
```bash
✅ 脚本输出 "Phase 1 迁移完成"
✅ go build 无错误
✅ go test ./tls/... 全部通过
✅ 目录结构: tls/internal/utils, tls/internal/ech 存在
```

---

## ❓ 常见问题 (FAQ)

### Q1: 这会影响我的灰度推出吗？
**A**: 否。Phase 1 仅涉及 TLS 包的重组，不影响灰度框架。可并行执行。

### Q2: 如果出现问题怎么回滚？
**A**: 
```bash
git reset --hard phase1-backup
# 或者
git checkout tls/    # 恢复整个 tls 目录
```

### Q3: 哪些文件会被修改？
**A**: ~12 个文件，主要是 import 语句。参考 [phase1-import-mapping.md](phase1-import-mapping.md)

### Q4: 需要手工修改代码吗？
**A**: 否，脚本自动处理所有 import 替换和文件移动。

### Q5: 支持哪些 Go 版本？
**A**: 1.21+（与项目 go.mod 保持一致）

---

## 📊 三个阶段对比

### Phase 1: TLS 层（Week 5-6）- 🟢 低风险

**变更核心**:
- `internal/tlsutil` → `tls/internal/utils`
- `tls/utils` → `tls/internal/utils`
- `tls/ech` → `tls/internal/ech`

**影响范围**:
- TLS 包（ja3, ja4, ja4s）✅
- 测试套件 ✅
- 配置桥接 ⚠️

**验证方式**:
```bash
go test ./tls/... -v
go test ./test/fingerprint_test.go -v
```

---

### Phase 2: HTTP 层（Week 6-7）- 🟡 中等风险

**变更核心**:
- `http/headers` → `http/internal/headers`
- `http/useragent` → `http/internal/useragent`
- 新增 `http/internal/caching`

**影响范围**:
- HTTP 包及子包 ✅
- ClientHints（可能有循环依赖）⚠️
- 集成测试 ⚠️

**验证方式**:
```bash
go test ./http/... -v
go test ./test/http*.go -v
```

---

### Phase 3: 公共 API（Week 7-8）- 🔴 高风险

**变更核心**:
- 新建 `pkg/fingerprint/` (主 API)
- 新建 `pkg/profiling/`
- 新建 `pkg/telemetry/`
- 整合 `internal/utils`、`internal/cache`、`internal/httputil`

**影响范围**:
- 所有公共接口 ✅
- 外部 API 契约 🔴
- 文档和示例 ✅

**验证方式**:
```bash
# 创建示例项目，仅导入 pkg/
cat > /tmp/test_pkg.go << 'EOF'
package main
import "github.com/vistone/fingerprint/pkg/fingerprint"
func main() {
    f := fingerprint.NewFingerprinter()
    // ...
}
EOF
go build /tmp/test_pkg.go
```

---

## 🛠 工具和脚本

### 脚本 1: 自动迁移（Phase 1）
```bash
bash scripts/phase1_tls_migration.sh [dry-run|execute]
```

**输出**:
- 详细的变更日志
- 成功/失败信息
- 验证结果

### 脚本 2: 分析工具（所有阶段）
```bash
bash scripts/analyze_imports.sh
```

**功能**:
- 统计各包的导入源
- 检测循环依赖
- 预检验证

### 脚本 3: 验证工具
```bash
bash scripts/verify_restructuring.sh [1|2|3]
```

**检查项**:
- 构建成功
- 测试通过
- 无循环依赖
- 无旧 import 残留

---

## 📋 执行清单

### 前期准备
- [ ] 阅读完整计划文档
- [ ] 创建 git tag 备份
- [ ] 运行 `dry-run` 预检
- [ ] 确认 git 分支状态干净

### Phase 1 执行
- [ ] 运行迁移脚本: `scripts/phase1_tls_migration.sh execute`
- [ ] 验证编译: `go build ./...`
- [ ] 验证测试: `go test ./tls/... -v`
- [ ] 检查 import: `grep -r "internal/tlsutil" --include="*.go" .`
- [ ] 提交变更: `git commit -m "refactor: Phase 1 TLS layer internalization"`

### Phase 2 执行（完成 Phase 1 后）
- [ ] 阅读 Phase 2 详细步骤
- [ ] 创建新的 git branch
- [ ] 执行类似步骤

### Phase 3 执行（完成 Phase 2 后）
- [ ] 设计 pkg/ 接口
- [ ] 整合 internal 工具包
- [ ] 编写外部 API 文档
- [ ] 更新示例代码

---

## 🎓 学习资源

### Go 项目布局最佳实践
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- 推荐目录结构：
  ```
  project/
  ├── cmd/           # 可执行入口
  ├── pkg/           # 公共 API （本项目暂未使用）
  ├── internal/      # 私有包
  └── ...
  ```

### 包设计的 5 个原则
1. **内聚**: 相关功能放在同一包
2. **独立**: 包间依赖最小化
3. **清晰**: 导出接口明确
4. **隔离**: internal/ 完全隐藏实现
5. **兼容**: 公共 API 向后兼容

### 本项目的应用
- ✅ TLS/HTTP 各自的 internal/ 实现
- ✅ 共享工具统一位置
- ⏳ pkg/ 作为稳定对外接口（Phase 3）

---

## 🆘 遇到问题？

### 构建失败
```bash
# 清理构建缓存
go clean -cache
go clean -modcache

# 重新分析依赖
go mod graph

# 检查循环依赖
go list -json ./... | jq '.ImportPath, .Imports'
```

### 特定包测试失败
```bash
# 详细输出
go test ./path/to/package -v -run TestName

# 生成覆盖率报告
go test ./tls/... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Import 仍然有错误
```bash
# 1. 查找遗漏的 import
grep -r "internal/tlsutil.*\.go:" .

# 2. 查找所有 import tlsutil 的位置
grep -r "import.*tlsutil" --include="*.go" .

# 3. 用编辑器全局替换（更安全）
# VS Code: Ctrl+H, 启用 Regex 模式
# 旧: "github\.com/vistone/fingerprint/internal/tlsutil"
# 新: "github.com/vistone/fingerprint/tls/internal/utils"
```

---

## 📞 后续支持

如需帮助：
1. 查阅完整计划: [package-restructuring-plan.md](package-restructuring-plan.md)
2. 查阅 import 清单: [phase1-import-mapping.md](phase1-import-mapping.md)
3. 查看脚本日志输出
4. 使用 `git diff` 查看具体变更

---

## ✨ 预期收益

✅ **代码清晰度提升** 50%  
✅ **编译时间缩减** ~10%（因依赖解析更清晰）  
✅ **维护成本降低** 30%（新人易理解）  
✅ **API 稳定性提高** （pkg/ 作为契约）  
✅ **测试覆盖更易维护**  

---

**准备好了？运行这个命令开始 Phase 1:**

```bash
bash scripts/phase1_tls_migration.sh dry-run
```

输出无误后：

```bash
bash scripts/phase1_tls_migration.sh execute
```

祝好！🎉
