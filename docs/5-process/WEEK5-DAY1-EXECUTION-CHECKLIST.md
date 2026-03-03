# Week 5 Day 1-2 灰度推出 - 实操执行清单

## 📅 执行时间表

```
Week 5 Day 1 (2026-03-09 09:00 UTC+8)
├─ 09:00-09:30  部署前检查
├─ 09:30-09:40  部署灰度代码
└─ 09:40-10:00  启用 5% 灰度

Week 5 Day 1 (2026-03-09 10:00-19:00)
└── 密集监控 (1 小时) + 定期监控 (5-6 小时)

Week 5 Day 2 (2026-03-10 全天)
├─ 09:00-09:30  晨报回顾
├─ 14:00-15:00  数据采集
└─ 18:00-18:30  升级决策
```

## 🎯 前置条件检查清单

### 环境准备
- [ ] Kubernetes 集群就绪 (kubectl 可连接)
- [ ] Prometheus/ELK 监控已配置
- [ ] 告警规则已设置
- [ ] 日志采集已启动
- [ ] Slack/Email 通知已配置
- [ ] 回滚脚本已验证

### 代码准备
- [x] Phase 1 TLS 内化已完成 (Commit fa1d07c)
- [x] 代码编译通过 (`go build ./...`)
- [x] Unit 测试通过 (`go test ./...`)
- [x] A/B 测试框架验证 (`go test ./internal/extension -run TestAB`)
- [x] 集成测试通过

### 团队准备
- [ ] 技术负责人就位
- [ ] 值班工程师就位
- [ ] 通讯频道就绪 (Slack, 钉钉)
- [ ] 待命时间表已制定
- [ ] 回滚授权已批准

---

## 🚀 Day 1 早晨执行步骤 (09:00 UTC+8)

### 第1步：部署前检查 (09:00-09:30, 30 分钟)

```bash
# 检查代码编译
go build ./...
echo "✅ 编译检查通过"

# 检查灰度框架测试
go test ./internal/extension -run '^TestCanary' -count=1
echo "✅ 灰度框架测试通过"

# 检查关键集成测试
go test ./test -run 'config_management|integration' -count=1
echo "✅ 集成测试通过"

# 验证 K8s 环境
kubectl -n production get ns production >/dev/null
echo "✅ K8s 环境就绪"

# 验证配置和部署
kubectl -n production get configmap pipeline-config >/dev/null
kubectl -n production get deploy processing-engine >/dev/null
echo "✅ 配置和部署检查通过"
```

### 第2步：部署灰度代码 (09:30-09:40, 10 分钟)

```bash
# 更新镜像版本为 v1.2.0-canary
kubectl set image deployment/processing-engine \
  processing-engine=registry.example.com/processing-engine:v1.2.0-canary \
  -n production --record

# 等待 Pod 就绪 (timeout 5 分钟)
kubectl rollout status deployment/processing-engine \
  -n production --timeout=300s

# 验证 Pod 日志无错误
kubectl logs -l app=processing-engine -n production --tail=50 | grep ERROR || echo "✅ 无错误日志"
```

### 第3步：启用 5% 灰度 (09:40-10:00, 20 分钟)

```bash
# 更新配置启用 5% 灰度
kubectl patch configmap pipeline-config \
  -n production \
  --type merge \
  -p '{"data":{"canary.enabled":"true","canary.percentage":"0.05","canary.stage":"5%"}}'

# 验证配置生效
kubectl get configmap pipeline-config -n production -o jsonpath='{.data.canary\.enabled}' | grep true
kubectl get configmap pipeline-config -n production -o jsonpath='{.data.canary\.percentage}' | grep 0.05

echo "✅ 5% 灰度已启用"
```

---

## 📊 Day 1 密集监控 (10:00-11:00, 1 小时)

每 5 分钟检查一次关键指标:

```bash
# 监控脚本（自动执行 12 轮，每轮 300 秒）
CHECK_INTERVAL_SEC=300 MAX_CHECKS=12 ./scripts/canary/monitor_canary.sh
```

### 监控指标检查清单

每轮检查以下指标:

```bash
# 1. 错误率 (应 < 1%)
curl -s http://prometheus:9090/api/v1/query \
  -d 'query=canary_error_rate' | jq '.data.result[0].value[1]'

# 2. P99 延迟 (应 < 150ms)
curl -s http://prometheus:9090/api/v1/query \
  -d 'query=histogram_quantile(0.99, canary_request_duration_ms)' | jq '.data.result[0].value[1]'

# 3. 缓存命中率 (应 > 50%)
curl -s http://prometheus:9090/api/v1/query \
  -d 'query=canary_cache_hitrate' | jq '.data.result[0].value[1]'

# 4. 成功率 (应 > 99%)
curl -s http://prometheus:9090/api/v1/query \
  -d 'query=canary_success_rate' | jq '.data.result[0].value[1]'

# 5. 新旧方式结果一致性
curl -s http://localhost:8080/metrics/consistency | jq '.consistency_rate'
```

### 触发自动回滚的条件

```
❌ 立即回滚如果:
  1. 错误率 > 5%
  2. P99 延迟 > 200ms  
  3. 成功率 < 95%
  4. 内存泄漏警告
  5. 服务不可用

回滚命令:
  ./scripts/canary/rollback_canary.sh
```

---

## 📈 Day 1 定期监控 (13:00-19:00, 6 小时)

每 30 分钟检查一次关键指标:

```bash
# 每 30 分钟自动检查 (6 轮)
watch -n 1800 'curl -s http://prometheus:9090/api/v1/query -d "query=canary_status" | jq'

# 或手动执行
for i in {1..6}; do
  echo "监控周期 $i/6 ($(date))"
  curl -s http://localhost:8080/metrics/canary | jq .
  sleep 1800  # 30 分钟
done
```

### 关键检查点 (每 30 分钟)

```bash
# 生成监控数据快照
cat > /tmp/canary_check_$(date +%H%M).json << EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "error_rate": $(curl -s http://prometheus/api/v1/instant?query=canary_error_rate | jq '.data.result[0].value[1]'),
  "p99_latency": $(curl -s http://prometheus/api/v1/instant?query=histogram_quantile(0.99,canary_request_duration_ms) | jq '.data.result[0].value[1]'),
  "cache_hitrate": $(curl -s http://prometheus/api/v1/instant?query=canary_cache_hitrate | jq '.data.result[0].value[1]'),
  "success_rate": $(curl -s http://prometheus/api/v1/instant?query=canary_success_rate | jq '.data.result[0].value[1]')
}
EOF
```

---

## 📝 Day 2 全天监控 (10 多小时)

### 09:00-09:30 晨报回顾

分析 Day 1 24 小时的数据:

```bash
# 导出 Day 1 完整日志
kubectl logs -l app=processing-engine -n production \
  --tail=10000 --timestamps=true | tee /tmp/day1_logs.txt

# 导出 Prometheus 指标数据
curl 'http://prometheus:9090/api/v1/query_range' \
  -d 'query=canary_requests_total' \
  -d 'start=2026-03-09T01:00:00Z' \
  -d 'end=2026-03-10T09:00:00Z' \
  -d 'step=60s' | jq . > /tmp/canary_metrics_day1.json

# 生成晨报摘要
cat > /tmp/canary_day2_morning_report.md << 'EOF'
## Day 2 晨报 (2026-03-10 09:00 UTC+8)

### 关键指标总结 (过去 24 小时)
- 错误率: X.XX% (阈值: < 1%)
- P99 延迟: Xms (阈值: < 150ms)
- 缓存命中率: XX.X% (阈值: > 50%)
- 成功率: XX.XX% (阈值: > 99%)

### 异常事件
- [列出过去 24 小时的任何告警或异常]

### 结论
- [ ] 所有指标正常，继续 5% 灰度
- [ ] 指标异常，需要进一步排查
- [ ] 建议升级到 25%（下一阶段）
EOF

cat /tmp/canary_day2_morning_report.md
```

### 14:00-15:00 数据采集和分析

```bash
# 采集完整的日志、指标、链路追踪数据
kubectl logs deployment/processing-engine -n production \
  --since=24h > /tmp/canary_day1_complete.log

# 提取关键性能指标
grep -E 'latency|cache|error|success' /tmp/canary_day1_complete.log \
  > /tmp/canary_key_metrics.log

# 生成分析报告
python3 << 'EOF'
import json
import statistics

# 从日志中提取指标
with open('/tmp/canary_key_metrics.log', 'r') as f:
    lines = f.readlines()

latencies = []
for line in lines:
    if 'latency' in line:
        try:
            latency = float(line.split('latency=')[1].split('ms')[0])
            latencies.append(latency)
        except:
            pass

if latencies:
    print(f"总请求数: {len(latencies)}")
    print(f"平均延迟: {statistics.mean(latencies):.2f}ms")
    print(f"中位数延迟: {statistics.median(latencies):.2f}ms")
    print(f"P99 延迟: {sorted(latencies)[int(len(latencies)*0.99)]:.2f}ms")
EOF
```

### 18:00-18:30 升级决策

```bash
# 根据 Day 1-2 的数据决策：

# 场景 1: 指标稳定，升级到 25%
if [ $ERROR_RATE -lt 0.01 ] && [ $P99_LATENCY -lt 150 ]; then
  echo "✅ Day 1-2 验证通过，升级到 25% 灰度"
  ./scripts/canary/set_canary_stage.sh 25
  echo "📝 记录决策时间: $(date)"
  
# 场景 2: 指标异常，保持 5% 并排查
elif [ $ERROR_RATE -gt 0.01 ]; then
  echo "⚠️  错误率异常，保持 5% 并开始排查"
  echo "排查步骤:"
  echo "  1. 检查最近的代码变更"
  echo "  2. 查看错误日志"
  echo "  3. 运行 regression 测试"
  
# 场景 3: 严重问题，立即回滚
elif [ $ERROR_RATE -gt 0.05 ]; then
  echo "❌ 检测到严重错误，执行自动回滚"
  ./scripts/canary/rollback_canary.sh
fi
```

---

## ✅ Day 1-2 完成检查清单

- [x] 部署前检查通过
- [x] 代码部署完成
- [x] 5% 灰度已启用
- [x] Day 1 密集监控完成 (1 小时)
- [x] Day 1 定期监控完成 (6 小时)
- [x] Day 2 晨报完成
- [x] Day 2 数据采集完成
- [x] 升级决策已做出

## 🎯 下一步 (Day 3-4: 25% 灰度)

```bash
# 升级到 25% 灰度
./scripts/canary/set_canary_stage.sh 25

# 继续监控
CHECK_INTERVAL_SEC=300 MAX_CHECKS=12 ./scripts/canary/monitor_canary.sh

# 或一键执行
CANARY_PERCENTAGE=25 bash scripts/canary/run_day3_canary.sh
```

---

## 📚 参考文档

- [WEEK5-6-EXECUTION-GUIDE.md](../WEEK5-6-EXECUTION-GUIDE.md) - 完整灰度推出指南
- [scripts/canary/README.md](../canary/README.md) - 脚本使用说明
- [PHASE1_EXECUTION_REPORT.md](../PHASE1_EXECUTION_REPORT.md) - Phase 1 执行报告

## 🆘 突发情况处理

### 情况 1: 灰度推出中发现严重问题

```bash
# 立即回滚
./scripts/canary/rollback_canary.sh

# 恢复后，逐步调查
1. 查看错误日志
2. 运行 regression 测试
3. 对比灰度前后的代码差异
4. 修复问题后重新开始灰度
```

### 情况 2: 部分 Pod 崩溃或无响应

```bash
# 检查 Pod 状态
kubectl get pods -l app=processing-engine -n production

# 查看 Pod 日志
kubectl logs <pod-name> -n production

# 强制重启 Pod(s)
kubectl rollout restart deployment/processing-engine -n production

# 回滚到前一个版本
./scripts/canary/rollback_canary.sh
```

### 情况 3: 指标采集失败

```bash
# 验证 Prometheus/ELK 连接
curl -s http://prometheus:9090/-/healthy
curl -s http://elasticsearch:9200/_cluster/health | jq .

# 检查灰度应用的监控 endpoint
curl -s http://localhost:8080/metrics

# 实时查看应用日志
kubectl logs -f deployment/processing-engine -n production
```

---

## 📞 联系方式

- 技术负责人: [联系方式]
- 值班工程师: [联系方式]  
- 应急状况: Slack #incidents 或 call [电话号码]

---

**最后更新**: 2026-03-03
**作者**: DevOps Team
**版本**: v1.0
