# Canary 脚本使用说明

用于 Week 5 Day 1-2 的 5% 灰度执行。

## 环境变量

- `KUBECTL_NAMESPACE`（默认 `production`）
- `CANARY_CONFIGMAP`（默认 `pipeline-config`）
- `CANARY_DEPLOYMENT`（默认 `processing-engine`）
- `CANARY_METRICS_BASE_URL`（默认 `http://localhost:8080`）

## Day 1 执行顺序

### 方式 A：一键编排（推荐）

```bash
./scripts/canary/run_day1_canary.sh
```

可选参数：

```bash
AUTO_ROLLBACK=1 CHECK_INTERVAL_SEC=300 MAX_CHECKS=12 ./scripts/canary/run_day1_canary.sh
```

### 方式 B：手动分步

1. 预检

```bash
./scripts/canary/precheck_day1.sh
```

2. 启用 5% 灰度

```bash
./scripts/canary/set_canary_stage.sh 5
```

3. 监控（示例：每5分钟检查一次，共12次）

```bash
CHECK_INTERVAL_SEC=300 MAX_CHECKS=12 ./scripts/canary/monitor_canary.sh
```

4. 触发严重阈值时回滚

```bash
./scripts/canary/rollback_canary.sh
```

## Makefile 快捷入口

```bash
make canary-day1
make canary-stage STAGE=25
make canary-rollback
```

## Day 2 决策

- 指标稳定：继续 5% 到完整 48h，再升级 25%
- 指标异常：保持 5% 并排查，必要时执行回滚

## 阶段切换

```bash
./scripts/canary/set_canary_stage.sh 25
./scripts/canary/set_canary_stage.sh 50
./scripts/canary/set_canary_stage.sh 100
./scripts/canary/set_canary_stage.sh off
```
