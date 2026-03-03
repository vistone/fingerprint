# 📚 包结构重构计划 - 完整导航索引

> **开始日期**: 2026-03-03  
> **完成时间**: ~12 周 (Week 5-8 灰度推出期间并行执行)  
> **影响范围**: 全量重构 | 3 个阶段 | ~50+ 个文件  
> **完成标志**: pkg/ API 暴露 + 内部结构清晰  

---

## 🗂 文档导航 (Quick Navigation)

### 📍 你的角色是...

#### 🔧 一线开发者 / 想快速理解
**→ [重构快速启动指南](RESTRUCTURING_QUICKSTART.md)**
- 5 分钟速读
- 立即开始 Phase 1
- 常见问题 FAQ
- 脚本使用示例

#### 🏗️ 架构师 / 想了解全貌
**→ [包结构重构完整计划](PACKAGE_RESTRUCTURING_PLAN.md)**
- 详细的 3 阶段规划
- 风险评估
- 自动化工具集
- 回滚策略

#### 🔍 详细执行者 / 关注 import 变更
**→ [Phase 1 Import 变更清单](PHASE1_IMPORT_MAPPING.md)**
- 完整的文件清单
- import 映射表
- 验证步骤
- 逐个手工操作指南

---

## 📊 重构概览

### 当前状态 (Current)
```
混乱的包分布:
├── tls/        (ja3/ ja4/ ja4s/ utils/ ech/)
├── http/       (headers/ useragent/ clienthints/ ...)
└── internal/   (utils/ errors/ tlsutil/ 混杂)
              ↓
           问题:
           • 工具包分散
           • 责任边界不清
           • 新人难以理解
```

### 目标状态 (Target)
```
清晰的分层:
├── tls/
│   ├── ja3/ja4/ja4s/  (公开算法)
│   └── internal/       (TLS工具私有)
├── http/
│   ├── headers/clienthints/useragent/  (公开 API)
│   └── internal/       (HTTP工具私有)
├── internal/
│   ├── utils/          (精选全局工具)
│   ├── errors/         (错误体系)
│   ├── cache/          (新: 统一缓存)
│   └── httputil/       (新: HTTP工具)
└── pkg/
    ├── fingerprint/    (新: 主 API)
    ├── profiling/      (新: 性能工具 API)
    └── telemetry/      (新: 遥测 API)
```

---

## 🔄 三阶段执行路线图

### Phase 1: TLS 层内化 ⭐ 低风险
| 周期 | 工作内容 | 脚本 | 预期时间 |
|------|---------|------|---------|
| Week 5-6 | 创建 tls/internal/ | ✅ phase1_tls_migration.sh | 2-3 天 |
| | 迁移 tlsutil → tls/internal/utils | | |
| | 更新所有 import | | |
| | 验证测试通过 | | |

**验证方式**: `go test ./tls/... && go build ./...`

---

### Phase 2: HTTP 层内化 ⭐⭐ 中等风险
| 周期 | 工作内容 | 脚本 | 预期时间 |
|------|---------|------|---------|
| Week 6-7 | 创建 http/internal/ | ⏳ 规划中 | 3-4 天 |
| | 迁移 headers/useragent → http/internal/ | | |
| | 提取 public API facades | | |
| | 验证 clienthints 兼容性 | | |

**验证方式**: `go test ./http/... && go test ./test/http*`

---

### Phase 3: 公共 API + 工具整合 ⭐⭐⭐ 高风险
| 周期 | 工作内容 | 脚本 | 预期时间 |
|------|---------|------|---------|
| Week 7-8 | 新建 pkg/ 暴露公开 API | ⏳ 规划中 | 3-5 天 |
| | 整合 internal/ 工具包 | | |
| | 设计稳定版本契约 | | |
| | 更新所有文档和示例 | | |

**验证方式**: 外部项目仅导入 pkg/ 可用

---

## 🚀 现在就开始

### 1️⃣ 第一步（5 分钟）
阅读 [🚀 快速启动指南](RESTRUCTURING_QUICKSTART.md)

### 2️⃣ 第二步（预检）
```bash
cd /media/stone/data/fingerprint

# 查看 Phase 1 要做什么
bash scripts/phase1_tls_migration.sh dry-run

# 输出应该显示所有要迁移的文件和 import 变更
```

### 3️⃣ 第三步（执行 Phase 1）
```bash
# 创建备份分支
git checkout -b restructure/phase1

# 执行迁移
bash scripts/phase1_tls_migration.sh execute

# 验证
go build ./...
go test ./tls/... -v
```

### 4️⃣ 第四步（提交）
```bash
git add -A
git commit -m "refactor: Phase 1 TLS layer internalization"
git push origin restructure/phase1
# 然后发起 PR，获得审查
```

---

## 📋 各文档详细对比

| 文档 | 长度 | 受众 | 用途 |
|------|------|------|------|
| [RESTRUCTURING_QUICKSTART](RESTRUCTURING_QUICKSTART.md) | 📄 5 页 | 🔧 开发者 | ⚡ 快速了解和启动 |
| [PACKAGE_RESTRUCTURING_PLAN](PACKAGE_RESTRUCTURING_PLAN.md) | 📖 20+ 页 | 🏗️ 架构师 | 📋 深入规划和验证 |
| [PHASE1_IMPORT_MAPPING](PHASE1_IMPORT_MAPPING.md) | 📄 8 页 | 🔍 执行者 | 🔗 import 变更清单 |
| **此文档** | 📄 3 页 | 👥 所有人 | 🗺️ 导航和总览 |

---

## 🎯 关键决策点

### 问题 1: 是否立即开始 Phase 1？
- ✅ **推荐**: 是。与灰度推出并行，不冲突。
- ⏳ **替代**: 等待灰度完成（降低同期风险）

### 问题 2: 是否需要 api 兼容性考量？
- ⚠️ **Phase 1-2**: 内部重构，兼容性强
- 🔴 **Phase 3**: pkg/ 暴露需要慎重设计

### 问题 3: 如何处理外部依赖（如果有）？
- 当前: fingerprint 是独立项目，无外部依赖
- 未来: 若有外部用户，需在 Phase 3 建立稳定 API

---

## ⚙️ 项目配置

### go.mod 要求
```go
go 1.21  // 支持类似的包组织
```

### 工具链
- Go 1.21+
- Git (分支管理)
- Bash (自动化脚本)
- Make (可选，用于快捷命令)

---

## 📊 进度跟踪

### 完成情况
```
Week 1-2 ██████████ 可观测性 (完成)
Week 3-4 ██████████ Pipeline框架 (完成)
Week 5   ██████░░░░ 灰度推出 Day 1-2 (进行中)
Week 5   ░░░░░░░░░░ 包结构重构 Phase 1 (待启动)  ← 你在这里
Week 6   ░░░░░░░░░░ 包结构重构 Phase 2 (待启动)
Week 7   ░░░░░░░░░░ 包结构重构 Phase 3 (待启动)
```

### Gantt 视图
```
灰度推出 (Week 5-7):     |████████████████|████√
包结构 (Week 5-8):       ░░░░|████████|████████|████
并行进行，无阻塞 ✅
```

---

## 🔐 风险管理

| 风险 | 影响 | 缓解 | 优先级 |
|------|------|------|--------|
| 编译失败 | 高 | 每步验证 go build | 🔴 P0 |
| 测试失败 | 高 | 逐包测试 | 🔴 P0 |
| 循环依赖 | 中 | 自动化检测 | 🟡 P1 |
| 文档过时 | 低 | 更新文档 | 🟢 P2 |

---

## 💡 最佳实践

### ✅ DO (推荐做法)
- ✅ 每个 Phase 前备份
- ✅ 使用自动化脚本
- ✅ 频繁提交小改动
- ✅ 每步验证测试
- ✅ 文档同步更新

### ❌ DON'T (避免做法)
- ❌ 同时执行多个 Phase
- ❌ 手工批量更改 import
- ❌ 跳过测试验证
- ❌ 忽视循环依赖警告
- ❌ 在 public API 上随意改动

---

## 📝 相关文档链接

### 架构相关
- [架构现代化计划](architecture-modernization-plan.md) - 整体愿景
- [模块化架构](../architecture/MODULAR_ARCHITECTURE.md) - 架构细节

### 灰度推出相关
- [Week 5-6 执行指南](WEEK5-6-EXECUTION-GUIDE.md) - 并行工作
- [灰度推出计划](WEEK5-6-ROLLOUT-PLAN.md) - 灰度内容

### 开发相关
- [开发指南](../sdk/DEVELOPER_GUIDE.md) - 开发约定
- [最佳实践](../sdk/BEST_PRACTICES.md) - 代码规范

---

## 🎓 学习资源

### 推荐阅读
1. **包管理**: https://golang.org/doc/effective_go#names
2. **项目布局**: https://github.com/golang-standards/project-layout
3. **接口设计**: https://www.youtube.com/watch?v=5DVV36uqQ70

### 本项目参考
- tls/ 包: 算法实现 (公开) + 工具实现 (私有)
- http/ 包: API 暴露 (公开) + 生成器 (私有)
- pkg/ 包: 稳定版本API (等待 Phase 3)

---

## ✨ 预期收益

### 代码质量
- 📈 结构清晰度: 50% ↑
- 📈 可维护性: 30% ↑
- 📈 新人效率: 40% ↑

### 开发效率
- ⏱️ 编译时间: 10% ↓
- 🔍 查找文件: 5 倍快
- 🧪 测试范围: 更精准

### API 稳定性
- 🔒 契约清晰: 100%
- 🛡️ 向后兼容: 有保障
- 📚 文档易懂: 大幅改善

---

## 🆘 需要帮助？

按顺序尝试:
1. 📖 阅读 [快速启动](RESTRUCTURING_QUICKSTART.md) 的 FAQ 章节
2. 📋 检查 [Phase 1 清单](PHASE1_IMPORT_MAPPING.md) 的特殊情况处理
3. 🔧 查看脚本的 dry-run 输出：
   ```bash
   bash scripts/phase1_tls_migration.sh dry-run | tail -20
   ```
4. 💬 查看完整计划中的[自动化工具](PACKAGE_RESTRUCTURING_PLAN.md#-自动化工具和脚本)部分

---

## 📅 时间表概览

```
2026-03-03 (今天)
├── Phase 1 规划完毕 ✓
├── 脚本编写完毕 ✓
└── 可立即开始 ✓

2026-03-10 ~ 2026-03-17 (Week 5-6)
├── Phase 1 TLS 执行
├── 灰度推出 Day 1-7 并行
└── PR 审查和合并

2026-03-18 ~ 2026-03-24 (Week 6-7)
├── Phase 2 HTTP 执行
├── 灰度推出完成
└── PR 审查和合并

2026-03-25 ~ 2026-03-31 (Week 7-8)
├── Phase 3 pkg API 执行
├── API 稳定性验证
└── 文档更新完毕
```

---

## 🎬 现在就行动

**你准备好了吗？** 

```bash
# 第一步（1 分钟）
cd /media/stone/data/fingerprint
cat docs/5-process/RESTRUCTURING_QUICKSTART.md | head -50

# 第二步（2 分钟）
bash scripts/phase1_tls_migration.sh dry-run | head -30

# 第三步（30 分钟）
bash scripts/phase1_tls_migration.sh execute

# 第四步（5 分钟）
go test ./tls/... -v
go build ./...
```

---

**下一步**: 打开 [⚡ 快速启动指南](RESTRUCTURING_QUICKSTART.md) 立即开始！🚀
