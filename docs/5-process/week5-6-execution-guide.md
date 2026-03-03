# Week 5-6 灰度推出实施指南 🚀

**从 PoC 到生产环境的 ProcessWithPipeline 灰度推出**

---

## 📅 时间表概览

```
Week 5 (2026-03-09 ~ 2026-03-15)
├─ Day 1-2 (03/09-03/10): 5% 灰度   - 基础验证
├─ Day 3-4 (03/11-03/12): 25% 灰度  - A/B 测试
├─ Day 5-6 (03/13-03/14): 50% 灰度  - 对称性验证
└─ Day 7   (03/15):        100% 灰度 - 全量切换

Week 6 (2026-03-16 ~ 2026-03-22)
├─ Day 1-2 (03/16-03/17): 继续监控
├─ Day 3-4 (03/18-03/19): 性能优化
└─ Day 5   (03/20):        项目总结
```

---

## 🎯 Week 5 详细执行步骤

### Day 1-2: 5% 灰度 (基础功能验证)

#### 🚀 一键执行入口（推荐）

```bash
# 默认：预检 -> 5%切换 -> 监控(12轮) -> 异常自动回滚
./scripts/canary/run_day1_canary.sh

# 可选：覆盖监控参数
AUTO_ROLLBACK=1 CHECK_INTERVAL_SEC=300 MAX_CHECKS=12 ./scripts/canary/run_day1_canary.sh
```

#### ✅ Day 1 早晨 (09:00 UTC+8)

**部署前检查** (需要 30 分钟):
- [ ] 检查代码编译: `go build ./...`
- [ ] 运行所有测试: `go test ./... -timeout 30s`
- [ ] 确认灰度框架正常: `go test ./internal/extension -v -run TestCanary`
- [ ] 检查监控告警规则已配置 (Prometheus)
- [ ] 验证回滚脚本可用
- [ ] 确认团队待命、通讯畅通
- [ ] 数据备份 (可选，但推荐)
- [ ] 确认灰度配置中 `Enabled: false` (初始禁用)
- [ ] Mock Kafka 消费者或测试数据源启动
- [ ] 日志收集器启动 (ELK 或中心日志)

**部署灰度代码** (需要 10 分钟):
```bash
# 1. 部署灰度框架代码
kubectl set image deployment/processing-engine \
  processing-engine=registry.example.com/processing-engine:v1.2.0-canary \
  --record --namespace=production

# 2. 验证 Pod 启动
kubectl get pods -l app=processing-engine -w

# 3. 检查日志（无错误）
kubectl logs -f deployment/processing-engine --tail=50
```

**启用 5% 灰度** (需要 5 分钟):
```bash
# 更新配置，启用 5% 灰度
kubectl patch configmap pipeline-config --type merge \
  -p '{"data":{"canary.enabled":"true","canary.percentage":"0.05"}}'

# 验证配置生效
curl http://localhost:8080/config/canary
# 期望输出: {"enabled": true, "percentage": 0.05, "stage": "5%"}
```

**关键检查点 (1 小时后)**:
```bash
# 1. 检查流量分配
curl http://localhost:8080/metrics/canary
# 期望: 约 5% 流量进入新方式

# 2. 检查错误率
curl http://localhost:8080/metrics/error-rate
# 期望: < 0.5% (< 1% 警告阈值)

# 3. 检查延迟
curl http://localhost:8080/metrics/p99-latency
# 期望: < 150ms (< 150ms 告警阈值)

# 4. 检查缓存命中率
curl http://localhost:8080/metrics/cache-hitrate
# 期望: > 50%

# 如果任何指标异常，立即回滚：
kubectl patch configmap pipeline-config --type merge \
  -p '{"data":{"canary.enabled":"false"}}'
```

**继续监控 (6 小时续)** （13:00-19:00):
```bash
# 每 30 分钟检查一次关键指标
watch -n 1800 'curl -s http://localhost:8080/metrics/canary | jq'
```

#### ✅ Day 2 全天 (24 小时监控)

**晨报** (09:00):
- 回顾过去 24 小时的数据
- 检查是否有告警
- 验证错误率、延迟、缓存指标

**检查清单**:
- [ ] 错误率稳定在 < 1%
- [ ] 延迟无明显增加 (P99 < 150ms)
- [ ] 缓存命中率 > 50%
- [ ] 新旧方式结果一致性 100%
- [ ] 无内存泄漏迹象
- [ ] 无网络问题

**数据收集** (14:00):
```bash
# 导出指标数据用于分析
kubectl logs deployment/processing-engine --since=48h \
  | grep "canary\|metrics" > /tmp/canary-day1-2.log

# 导出 Prometheus 数据
curl 'http://prometheus:9090/api/v1/query_range' \
  -d 'query=canary_requests_total&start=<start>&end=<end>&step=60s' \
  > /tmp/canary-metrics.json
```

**决策** (18:00):
```
IF 所有指标正常:
  ✅ 决定升级到 Day 3 的 25% 灰度
ELSE IF 有轻微警告 (缓存低等):
  ⚠️  予以注意，但继续观察
ELSE IF 有严重问题:
  ❌ 立即回滚，调查根本原因
```

### Day 3-4: 25% 灰度 (A/B 性能测试)

#### ✅ Day 3 早晨 (09:00 UTC+8)

**前置条件检查**:
- [ ] Day 1-2 的数据全部正常
- [ ] 无待处理的告警
- [ ] 团队已准备好

**升级到 25% 灰度**:
```bash
kubectl patch configmap pipeline-config --type merge \
  -p '{"data":{"canary.percentage":"0.25","canary.stage":"25%"}}'

# 验证
curl http://localhost:8080/config/canary
```

**A/B 性能对标** (Day 3 全天):

在这个阶段，我们需要收集足够数据进行 A/B 测试：

```bash
# 每 2 小时收集一份快照
05:00: curl http://localhost:8080/metrics/ab-snapshot > /tmp/ab-snap-05.json
07:00: curl http://localhost:8080/metrics/ab-snapshot > /tmp/ab-snap-07.json
09:00: curl http://localhost:8080/metrics/ab-snapshot > /tmp/ab-snap-09.json
...
23:00: curl http://localhost:8080/metrics/ab-snapshot > /tmp/ab-snap-23.json

# 快照应该包含:
# {
#   "new_method": {"avg_latency": "12.5ms", "p99": "45.2ms", "error_rate": 0.0005},
#   "old_method": {"avg_latency": "10.8ms", "p99": "40.1ms", "error_rate": 0.0003},
#   "cache_hit_rate": 0.942,
#   "total_requests": 125000
# }
```

**性能分析** (Day 3 晚间 18:00):
```
对标指标:
┌─────────────────────┬──────────┬──────────┬───────────┐
│ 指标                │ 新方式   │ 旧方式   │ 对等价    │
├─────────────────────┼──────────┼──────────┼───────────┤
│ P50 延迟            │ 12.5ms   │ 10.8ms   │ ✅ 允许   │
│ P99 延迟            │ 45.2ms   │ 40.1ms   │ ✅ 允许   │
│ 错误率              │ 0.05%    │ 0.03%    │ ✅ 正常   │
│ 缓存命中率          │ 94.2%    │ N/A      │ ✅ 优秀   │
│ 成功率              │ 99.95%   │ 99.97%   │ ✅ 接近   │
└─────────────────────┴──────────┴──────────┴───────────┘

结果一致性: 99.8% (使用 1000 对样本)
统计显著性: p < 0.05 ✅ (T 检验)

建议: ✅ 可继续推出
```

#### ✅ Day 4 全天 (继续 A/B 测试)

重复 Day 3 的监控，确保数据稳定。

**数据最终确认** (18:00):
```bash
# 收集最终的 A/B 报告
./scripts/ab-test-report.sh > /tmp/ab-final-report.md

# 报告应该包含:
# - 总样本量
# - 置信区间
# - 建议（可升级/需调查/立即回滚）
```

### Day 5-6: 50% 灰度 (对称性验证)

#### ✅ Day 5 早晨 (09:00 UTC+8)

**前置条件检查**:
- [ ] Day 3-4 的 A/B 测试通过
- [ ] 所有指标对等
- [ ] 结果一致性 > 99%

**升级到 50% 灰度**:
```bash
kubectl patch configmap pipeline-config --type merge \
  -p '{"data":{"canary.percentage":"0.50","canary.stage":"50%"}}'
```

**对称性验证** (核心验证):

对称性是指：新旧方式处理相同请求时，结果应该完全一致。

```bash
# 随机抽取 1000 对请求
# 让新旧方式分别处理相同请求，对比结果

./scripts/symmetry-validation.sh \
  --samples 1000 \
  --output /tmp/symmetry-report.json

# 期望输出:
# {
#   "total_pairs": 1000,
#   "consistent_pairs": 995,
#   "consistency_rate": 0.995,
#   "issues": [
#     {"request_id": "req-123", "diff": "..."}
#   ]
# }

# 验收标准:
# - 一致性率 > 99.5% ✅
# - 差异都是预期的(缓存时间戳、随机数等) ✅
```

**48 小时长期稳定性测试** (Day 5-6):
```bash
# 开启持续监控，24/7 运行
nohup ./scripts/continuous-monitor.sh \
  --interval 15m \
  --duration 48h \
  --output /tmp/stability-report.log &

# 监控项:
# - 错误率趋势
# - 延迟分位数趋势
# - 内存占用趋势
# - GC 行为
# - 缓存命中率稳定性
```

**Day 6 决策** (18:00):
```
检查清单:
- [ ] 一致性率 > 99.5%
- [ ] 错误率始终 < 1%
- [ ] 延迟无明显增加
- [ ] 缓存工作正常
- [ ] 无内存泄漏
- [ ] 无 GC 问题

IF 全部通过:
  ✅ 准备 Day 7 全量切换
ELSE:
  ❌ 修复问题后重新尝试 50% 灰度
```

### Day 7: 100% 灰度 (全量切换)

#### ✅ Day 7 早晨 (07:00 UTC+8)

**全量切换前检查** (需要 1 小时):
- [ ] Day 5-6 的对称性验证通过
- [ ] 所有测试通过
- [ ] 版本标签已打: `v1.2.0-ready-for-production`
- [ ] 回滚计划已准备
- [ ] 通讯渠道已打开

**执行全量切换** (08:00):
```bash
# 方法 1: 立即 100% 切换
kubectl patch configmap pipeline-config --type merge \
  -p '{"data":{"canary.enabled":"false"}}'
# 这时所有请求都会使用新方式

# 或者
# 方法 2: 渐进式切换 (更保险)
kubectl patch configmap pipeline-config --type merge \
  -p '{"data":{"canary.percentage":"0.75"}}'  # 75%
# 等待 15 分钟后
kubectl patch configmap pipeline-config --type merge \
  -p '{"data":{"canary.percentage":"1.00"}}'  # 100%
```

**关键检查点** (切换后每 15 分钟):
- 15 分钟 (08:15):
  ```bash
  # 错误率是否有跳升?
  curl http://localhost:8080/metrics/error-rate
  # 期望: < 0.5%
  
  # 延迟是否有跳升?
  curl http://localhost:8080/metrics/latency
  # 期望: P99 < 150ms
  ```

- 1 小时 (09:00):
  ```bash
  # 所有指标稳定吗?
  curl http://localhost:8080/metrics/summary
  ```

- 6 小时后 (14:00):
  ```bash
  # 长期稳定吗?
  curl http://localhost:8080/metrics/summary
  ```

**监控和优化** (Day 7 全天):
```
08:00 - 08:45: 密集监控 (每 5 分钟)
09:00 - 18:00: 常规监控 (每 30 分钟)
18:00 - 次日:  宽松监控 (每 2 小时)
```

---

## 🛡️ 回滚执行流程

### 自动回滚触发条件

```
监控到以下任何情况，立即回滚:

1. 错误率 > 3% (持续 5 分钟)
   → 立即停止新方式流量
   
2. P99 延迟 > 500ms (持续 10 分钟)
   → 立即停止新方式流量
   
3. OOM / Panic
   → 立即停止新方式流量
   
4. 数据异常 (一致性 < 95%)
   → 立即停止新方式流量
```

### 快速回滚脚本

```bash
#!/bin/bash
# quick-rollback.sh - 快速回滚脚本 (< 5 分钟)

set -e
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
LOG_DIR=/tmp/rollback-$TIMESTAMP

mkdir -p $LOG_DIR

echo "🔴 开始回滚..."

# 1. 立即停止新方式流量 (< 1 分钟)
echo "⏱️  步骤 1: 禁用灰度..."
kubectl patch configmap pipeline-config --type merge \
  -p '{"data":{"canary.enabled":"false"}}' \
  && echo "✅ 灰度已禁用"

# 2. 验证回滚成功 (2 分钟)
echo "⏱️  步骤 2: 验证回滚..."
sleep 30  # 等待配置生效
ERROR_RATE=$(curl -s http://localhost:8080/metrics/error-rate | jq .error_rate)
if (( $(echo "$ERROR_RATE < 0.01" | bc -l) )); then
    echo "✅ 错误率已恢复: $ERROR_RATE"
else
    echo "❌ 错误率未恢复: $ERROR_RATE (预期 < 0.01)"
    exit 1
fi

# 3. 收集诊断数据 (1 分钟)
echo "⏱️  步骤 3: 收集诊断数据..."
kubectl logs deployment/processing-engine --tail=1000 \
  > $LOG_DIR/engine.log
kubectl top pods -l app=processing-engine \
  > $LOG_DIR/resources.txt
curl -s http://localhost:8080/metrics/all \
  > $LOG_DIR/metrics.json

# 4. 发送告警 (< 1 分钟)
echo "⏱️  步骤 4: 发送告警..."
curl -X POST http://alert-service/notify \
  -H "Content-Type: application/json" \
  -d '{
    "title": "灰度推出回滚",
    "level": "critical",
    "stage": "'$(cat /etc/rollback-stage.txt)'",
    "reason": "自动检测到错误率升高",
    "logs": "'$LOG_DIR'"
  }'

# 5. 创建事后分析工单
echo "⏱️  步骤 5: 创建事后分析..."
curl -X POST http://jira/create-ticket \
  -H "Content-Type: application/json" \
  -d '{
    "project": "PIPELINE",
    "type": "Bug",
    "summary": "灰度推出回滚 - "'$TIMESTAMP'",
    "description": "诊断数据: '$LOG_DIR'",
    "reporter": "cicd-automation"
  }'

echo ""
echo "════════════════════════════════════"
echo "✅ 回滚完成 (耗时: ~5 分钟)"
echo "════════════════════════════════════"
echo "诊断数据位置: $LOG_DIR"
echo ""
echo "后续步骤:"
echo "1. 查看日志: cat $LOG_DIR/engine.log"
echo "2. 分析指标: cat $LOG_DIR/metrics.json | jq"
echo "3. 调查根本原因"
echo "4. 修复后再次尝试灰度"
```

---

## 📋 检查清单

### 部署前检查 (✅ 必须全部通过)

Day 1 早晨 09:00:
- [ ] 代码编译成功
- [ ] 所有测试通过
- [ ] 灰度框架测试通过 (TestCanary*)
- [ ] 监控系统就绪
- [ ] Prometheus 告警规则已配置
- [ ] 日志收集器启动
- [ ] 团队通讯就绪
- [ ] 回滚脚本可用
- [ ] 初始灰度配置检查: 只有框架代码，但 enabled=false
- [ ] (可选) 数据备份

### 灰度期间检查 (每 4-24 小时)

- [ ] 错误率正常 (< 1%)
- [ ] 延迟正常 (P99 < 150ms)
- [ ] 缓存命中率 > 50%
- [ ] 内存占用正常
- [ ] GC 时间正常
- [ ] 无告警消息
- [ ] 新旧方式结果一致性 > 99%

### 灰度完成检查

Day 7 18:00:
- [ ] 100% 灰度已切换
- [ ] 所有关键指标验证通过
- [ ] 一致性验证通过 (>99.5%)
- [ ] 性能对标通过
- [ ] 48 小时以上稳定性验证通过
- [ ] 无内存泄漏
- [ ] 无异常告警

---

## 📊 数据收集和分析

### 关键指标导出

```bash
# 每个阶段结束时导出数据
export_canary_data() {
    local stage=$1
    local output_dir=$2
    
    # 导出 Prometheus 数据
    curl 'http://prometheus:9090/api/v1/query_range' \
      -d 'query=canary_requests_total' \
      -d 'start=<start>&end=<end>&step=60s' \
      > $output_dir/requests-$stage.json
    
    # 导出应用日志
    kubectl logs deployment/processing-engine \
      --tail=10000 \
      | grep -E 'canary|error|latency' \
      > $output_dir/logs-$stage.txt
    
    # 导出资源使用
    kubectl top nodes > $output_dir/nodes-$stage.txt
    kubectl top pods > $output_dir/pods-$stage.txt
}

# 使用:
export_canary_data "day1-2" /tmp/canary-analysis/5percent
export_canary_data "day3-4" /tmp/canary-analysis/25percent
export_canary_data "day5-6" /tmp/canary-analysis/50percent
```

### 分析和报告

```bash
# 使用 Python 脚本分析数据
python3 analyze_canary.py \
  --data /tmp/canary-analysis \
  --output /tmp/canary-report.html

# 报告应该包含:
# - 流量分布图
# - 错误率趋势
# - 延迟分位数对比
# - 缓存命中率时序
# - 一致性统计
# - 建议和结论
```

---

## 🎓 常见问题

### Q1: 灰度期间如果发现轻微问题怎么办?

**答**: 根据问题严重程度：
- 缓存命中率低 (50-70%) → 继续观察，不影响推出
- 错误率 0.5-1% → 继续观察 24 小时，如无改善则调查
- 错误率 > 3% → 立即回滚

### Q2: 是否可以跳过某个阶段?

**答**: 不推荐。虽然技术上可以，但：
- 5% → 25% 是必须的 (基础功能验证)
- 25% → 50% 是必须的 (A/B 性能对标)
- 50% → 100% 可以在一致性验证后立即执行

### Q3: 如果 Day 1-2 失败，应该等多久后重新尝试?

**答**: 
1. 如果是代码 Bug → 修复代码后立即重新尝试
2. 如果是配置问题 → 修改配置后立即重新尝试
3. 如果是间歇性问题 → 等待 2-4 小时后重新尝试

### Q4: 生产上有突发流量，是否应该暂停灰度?

**答**: 
- 如果突发流量是预期的 → 继续灰度，监控是否有影响
- 如果突发流量不预期 → 可考虑暂停灰度 (禁用后仍可随时恢复)

---

## 🚀 Week 6 优化和总结

### Day 1-2: 100% 流量监控

```bash
# 继续监控，但放松频率
watch -n 3600 'curl -s http://localhost:8080/metrics/summary | jq'
```

### Day 3-4: 性能优化

根据灰度期间采集的数据做优化：
- 缓存策略优化
- 处理延迟 hot spot
- 内存占用优化

### Day 5: 项目总结

生成最终的灰度推出总结报告：
- 灰度成效评估
- 性能改进数据
- 问题和解决方案总结
- 未来改进方向

---

**最后更新**: 2026-03-02  
**版本**: 1.0.0  
**状态**: 已准备就绪，等待Week 5 Day 1 执行 🚀
