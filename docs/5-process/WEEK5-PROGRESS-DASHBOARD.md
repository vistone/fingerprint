# Week 5-6 灰度推出 与 Phase 2-3 重构 - 综合进度仪表板

**更新时间**: 2026-03-03 10:00 UTC+8  
**状态**: 🚀 Week 5 Day 1-2 灰度推出已准备完毕，可立即启动

---

## 📊 全景进度图

```
Week 5-6 任务总览:
═══════════════════════════════════════════════════════════════════════════════

灰度推出（Week 5）
├─ Day 1-2  (03/09-03/10): 5%   灰度  ••••••••□ 准备完毕 (📋 清单已生成)
├─ Day 3-4  (03/11-03/12): 25%  灰度  □□□□□□□□□ 脚本已准备
├─ Day 5-6  (03/13-03/14): 50%  灰度  □□□□□□□□□ 计划中
└─ Day 7    (03/15):       100% 灰度  □□□□□□□□□ 计划中

包结构重构（并行）
├─ Phase 1 TLS 内化 (✅ 完成) ██████████ 
│  └─ Commit fa1d07c，已推送 origin/main
├─ Phase 2 HTTP 内化 (⏳ 待启动) □□□□□□□□□
│  └─ 计划: Week 6 启动，可与灰度推出 Day 3+ 并行
└─ Phase 3 pkg 公开化 (📋 规划完毕) □□□□□□□□□
   └─ 计划: Week 7 启动

═══════════════════════════════════════════════════════════════════════════════
```

---

## 🎯 Week 5 Day 1-2 灰度推出 - 实施状态

### ✅ 已完成的准备工作

| 项目 | 完成度 | 说明 |
|------|--------|------|
| **文档** | ✅ 100% | Week5-6执行指南、Day1执行清单已生成 |
| **脚本** | ✅ 100% | run_day1/day3_canary.sh, precheck, monitor, rollback 脚本就绪 |
| **代码** | ✅ 100% | Phase 1 完成，灰度框架验证通过 |
| **A/B框架** | ✅ 100% | 灰度性能对比框架已实现 |
| **测试基线** | ✅ 100% | benchmark_ab_test.go 可执行 |

### 📋 Day 1-2 执行检查清单

```
□ 环境准备
  □ K8s 集群就绪
  □ Prometheus/ELK 监控配置
  □ Slack/Email 告警配置
  □ 值班团队待命

□ 代码准备
  □ go build ./... 通过
  □ go test ./... 通过
  □灰度框架测试通过

□ 执行
  □ 部署前检查 (09:00-09:30)
  □ 部署灰度代码 (09:30-09:40)
  □ 启用 5% 灰度 (09:40-10:00)
  □ 密集监控 1 小时 (10:00-11:00)
  □ 定期监控 6 小时 (13:00-19:00)
  □ Day 2 晨报 (09:00 下天)
  □ 升级决策 (18:00)
```

### 🚀 启动 Day 1-2 灰度推出的方式

#### 方式 A：一键启动（生产环境推荐）

```bash
# 在 K8s 生产环境中执行
./scripts/canary/run_day1_canary.sh

# 或带参数覆盖
AUTO_ROLLBACK=1 CHECK_INTERVAL_SEC=300 MAX_CHECKS=12 \
  ./scripts/canary/run_day1_canary.sh
```

#### 方式 B：Makefile 快捷命令（如果已配置）

```bash
# 启动 Day 1-2 灰度
make canary-day1

# Day 3 升级到 25%
make canary-stage STAGE=25

# 紧急回滚
make canary-rollback
```

#### 方式 C：手动分步执行

```bash
# 1. 部署前检查
./scripts/canary/precheck_day1.sh

# 2. 启用 5% 灰度
./scripts/canary/set_canary_stage.sh 5

# 3. 监控（12 轮，每轮 5 分钟）
CHECK_INTERVAL_SEC=300 MAX_CHECKS=12 ./scripts/canary/monitor_canary.sh

# 4. 若需回滚
./scripts/canary/rollback_canary.sh
```

### 📊 Day 1-2 成功标准

```
✅ Day 1 检查通过标准:
  • 错误率 < 1.0%
  • P99 延迟 < 150ms
  • 缓存命中率 > 50%
  • 成功率 > 99%
  • 无内存泄漏

✅ Day 2 进度标准:
  • 24 小时内指标稳定
  • 无持续告警
  • 新旧方式结果一致性 > 99%
  • 建议升级到 Day 3 的 25%
```

---

## 📈 Week 5 Day 3-7 灰度推出规划

### Day 3-4: 25% 灰度（A/B 对称性测试）

```bash
# 启动 Day 3-4
./scripts/canary/run_day3_canary.sh

# 或一键启动
CANARY_PERCENTAGE=25 bash scripts/canary/run_day3_canary.sh
```

**关键验证**:
- A/B 性能对比（新 vs 旧）
- 新旧方式结果一致性验证
- 25% 流量下的延迟和缓存性能

**预期结果**: 如果性能对比无差异，进入 Day 5

### Day 5-6: 50% 灰度（对称性验证）

```bash
CANARY_PERCENTAGE=50 bash scripts/canary/run_day5_canary.sh
```

**关键验证**:
- 继续 A/B 对称性验证
- 多地域部署验证
- 性能恒定性测试

### Day 7: 100% 灰度（全量切换）

```bash
./scripts/canary/set_canary_stage.sh 100

# 继续监控 24 小时
./scripts/canary/monitor_canary.sh
```

**切换后验证**:
- 全量流量路由到新的 ProcessWithPipeline
- 旧的方式完全禁用
- 继续 24 小时监控

---

## 🔄 Phase 2 HTTP 层内化 - 并行计划

### ⏱️ 启动时间表

```
最早启动: Week 6 中期 (03/17 左右)
可并行: 与灰度推出 Day 3+ 同步进行
前提条件: Day 1-2 灰度推出验证通过

预计耗时: 3-4 天（规划和自动化脚本已完毕）
```

### 📋 Phase 2 工作内容

参考: [PACKAGE_RESTRUCTURING_PLAN.md](./PACKAGE_RESTRUCTURING_PLAN.md) Phase 2 章节

```
├─ http/ 包内化
│  ├─ http/internal/useragent/   (新建，从 http/useragent 迁移)
│  ├─ http/internal/headers/     (新建，从 http/headers 迁移)
│  ├─ http/internal/clienthints/ (新建，从 http/clienthints 迁移)
│  └─ http/internal/policy/      (新建，内化流量策略)
├─ internal/contrib 清理
│  └─ 删除过期的 contribution 工具
└─ 验证和文档
   ├─ Import 清理
   ├─ 包访问级别验证
   └─ 生成 Phase 2 执行报告
```

### 🚀 启动 Phase 2 的预期命令

```bash
# Day 1: dry-run 预检
bash scripts/phase2_http_migration.sh dry-run

# Day 2-3: 执行迁移
bash scripts/phase2_http_migration.sh execute

# Day 4: 合并到 main
git checkout main
git merge restructure/phase2 --no-edit
git push origin main
```

### ✅ Phase 2 成功条件

```
✓ http/internal/ 新目录创建，4 个包迁移完成
✓ 旧 http/ 子目录删除
✓ 0 个 import 遗留引用
✓ 所有测试通过
✓ 并入 main 分支，推送远程
```

---

## 🎯 关键时间节点

### 本周 (Week 5: 03/03-03/09)

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
2026-03-09 (周日) 09:00 UTC+8
  ▶ 灰度推出 Day 1 启动
     - 部署前检查
     - 5% 灰度启用
     - 密集监控

2026-03-10 (周一) 全天
  ▶ 灰度推出 Day 2 继续
     - 晨报回顾
     - 数据采集
     - 升级决策 -> Day 3
```

### 下周 (Week 6: 03/10-03/16)

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
2026-03-11 ~ 2026-03-12 (3-4⊆)
  ▶ 灰度推出 Day 3-4
     - 25% 灰度推出
     - A/B 对称性测试

2026-03-13 ~ 2026-03-14 (5-6)
  ▶ 灰度推出 Day 5-6
     - 50% 灰度推出
     - 对称性继续验证
  
  ⚡ 可在此时启动 Phase 2
     - 不会与灰度推出冲突

2026-03-15 (日)
  ▶ 灰度推出 Day 7
     - 100% 灰度切换（全量）
     - 继续 24h 监控
```

### 两周后 (Week 7: 03/17-03/23)

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
2026-03-17 ~ 2026-03-20
  ▶ Phase 2 HTTP 内化
     - 执行 http/ 包迁移
     - 验证和合并

2026-03-21 ~ 2026-03-23
  ▶ Phase 3 准备（或 Phase 2 完成）
```

---

## 📚 关键文档导览

### 立即需要 (Week 5)

| 文档 | 用途 | 优先级 |
|------|------|--------|
| [WEEK5-DAY1-EXECUTION-CHECKLIST.md](./WEEK5-DAY1-EXECUTION-CHECKLIST.md) | Day 1-2 灰度执行清单 | 🔴 立即 |
| [WEEK5-6-EXECUTION-GUIDE.md](./WEEK5-6-EXECUTION-GUIDE.md) | 完整灰度推出指南 | 🔴 立即 |
| [scripts/canary/README.md](../scripts/canary/README.md) | 灰度脚本使用说明 | 🔴 立即 |

### 后续需要 (Week 6+)

| 文档 | 用途 | 优先级 |
|------|------|--------|
| [PACKAGE_RESTRUCTURING_PLAN.md](./PACKAGE_RESTRUCTURING_PLAN.md) | Phase 2-3 规划 | 🟡 Week 6 |
| [PHASE1_MERGE_AND_PHASE2_PLAN.md](./PHASE1_MERGE_AND_PHASE2_PLAN.md) | Phase 2 详细规划 | 🟡 Week 6 |

### 参考文档 (已有记录)

| 文档 | 说明 |
|------|------|
| [PHASE1_EXECUTION_REPORT.md](./PHASE1_EXECUTION_REPORT.md) | Phase 1 完整执行记录 |
| [RESTRUCTURING_INDEX.md](./RESTRUCTURING_INDEX.md) | 重构项目导航 |

---

## 🎯 当前状态总结

### ✅ 已完成

```
✓ Phase 1 TLS 层内化完成并合并到 main (Commit fa1d07c)
✓ 灰度推出的所有脚本和文档已生成
✓ Day 1-2 执行清单已准备
✓ Phase 2-3 规划文档已完成
```

### 🚀 即将开始 (本周)

```
► Week 5 Day 1-2 灰度推出（5%）
  └─ 预计启动时间: 2026-03-09 09:00 UTC+8
```

### 🔄 进行中 (并行)

```
► 灰度推出流程（Day 1-7）
► Phase 2 规划和准备（Week 6 中启动）
```

### ⏳ 待启动

```
☐ Phase 2 HTTP 内化（Week 6 启动）
☐ Phase 3 pkg 公开化（Week 7 启动）
```

---

## 🆘 遇到问题？

### 灰度推出问题

- **Day 1-2 脚本失败**: 查看 [WEEK5-DAY1-EXECUTION-CHECKLIST.md](./WEEK5-DAY1-EXECUTION-CHECKLIST.md) 的"突发情况处理"
- **指标异常**: 参考 [WEEK5-6-EXECUTION-GUIDE.md](./WEEK5-6-EXECUTION-GUIDE.md) 的回滚步骤
- **脚本权限问题**: 执行 `chmod +x scripts/canary/*.sh`

### Phase 2-3 问题

- **重构规划**: [PACKAGE_RESTRUCTURING_PLAN.md](./PACKAGE_RESTRUCTURING_PLAN.md)
- **自动化脚本**: 参考 Phase 1 中的 `scripts/phase1_tls_migration.sh` 架构

---

## 📞 联系方式

- **灰度推出负责人**: [联系方式]
- **包结构重构负责人**: [联系方式]
- **应急**: #incidents Slack 或 on-call 号码

---

**最后更新**: 2026-03-03 10:00 UTC+8  
**下次更新**: 2026-03-10 (Day 2 完成后)  
**版本**: v1.0 - 初始版本

---

## 📊 历史更新记录

| 日期 | 更新内容 | 版本 |
|------|---------|------|
| 2026-03-03 | 初始版本，Phase 1 完成，Day 1-2 准备完毕 | v1.0 |
| - | - | - |
