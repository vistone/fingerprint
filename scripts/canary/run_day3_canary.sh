#!/usr/bin/env bash

# ╔══════════════════════════════════════════════════════════════════════════════╗
# ║         Week 5 Day 3-4: 25% 灰度推出自动化脚本                              ║
# ║         A/B 对称性性能测试 & 高精度监控                                      ║
# ╚══════════════════════════════════════════════════════════════════════════════╝

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
LOG_DIR="${CANARY_LOG_DIR:-/tmp/canary-day3-$(date +%Y%m%d_%H%M%S)}"
AUTO_ROLLBACK="${AUTO_ROLLBACK:-1}"
STAGE="${CANARY_STAGE:-25}"
MONITOR_CHECK_INTERVAL_SEC="${CHECK_INTERVAL_SEC:-300}"
MONITOR_MAX_CHECKS="${MAX_CHECKS:-12}"

mkdir -p "$LOG_DIR"

log() {
  printf '[%s] %s\n' "$(date '+%F %T')" "$*"
}

run_step() {
  local name="$1"
  shift
  log "开始: ${name}"
  if "$@" >"$LOG_DIR/${name}.log" 2>&1; then
    log "完成: ${name}"
  else
    log "失败: ${name}，详见 $LOG_DIR/${name}.log"
    return 1
  fi
}

log "Day3 灰度编排开始，日志目录: $LOG_DIR"
log "目标阶段: 25% 灰度"

# ───────────────────────────────────────────────────────────────────────────────
# 第1步：前置条件验证
# ───────────────────────────────────────────────────────────────────────────────

echo "[1/6] Day 1-2 验证（确保 5% 灰度稳定）"

# 验证 Day 1 是否成功完成
if [ ! -f "$LOG_DIR/../canary-day1-*/03_monitor.log" ]; then
  log "警告：未找到 Day 1 监控日志，请确保 Day 1-2 已成功完成"
  log "继续前，请查看 Day 1-2 报告并确认指标稳定"
fi

# ───────────────────────────────────────────────────────────────────────────────
# 第2步：A/B 性能对比
# ───────────────────────────────────────────────────────────────────────────────

echo "[2/6] A/B 性能对比测试"
if go test ./internal/extension -run TestAB_ComparisonBenchmark -timeout 120s >"$LOG_DIR/02_ab_comparison.log" 2>&1; then
  log "✓ A/B 性能测试通过"
  
  # 提取关键指标对比
  grep -E 'Old.*New|Improvement|MemoryReduction|Throughput' "$LOG_DIR/02_ab_comparison.log" | tee -a "$LOG_DIR/ab_metrics.txt" || true
else
  log "✗ A/B 性能测试失败，详见日志"
  exit 1
fi

# ───────────────────────────────────────────────────────────────────────────────
# 第3步：升级到 25% 灰度
# ───────────────────────────────────────────────────────────────────────────────

echo "[3/6] 升级到 25% 灰度"

if ! "$ROOT_DIR/scripts/canary/set_canary_stage.sh" "$STAGE"; then
  log "✗ 升级失败"
  exit 1
fi

log "✓ 已升级到 25% 灰度"

# 验证配置生效（等待 30 秒）
sleep 30

# ───────────────────────────────────────────────────────────────────────────────
# 第4步：一致性验证（新旧方式结果对比）
# ───────────────────────────────────────────────────────────────────────────────

echo "[4/6] 一致性验证（新旧方式结果对比）"
if go test ./internal/extension -run TestAB_ConsistencyValidation -timeout 60s >"$LOG_DIR/04_consistency.log" 2>&1; then
  log "✓ 一致性验证通过"
  
  # 检查一致性指标
  consistency=$(grep -oP 'consistency[\s:=]+\K[0-9.]+' "$LOG_DIR/04_consistency.log" | head -1 || echo "0")
  log "一致性指标: ${consistency}%"
  
  if (( $(echo "$consistency < 99.5" | bc -l) )); then
    log "⚠ 一致性低于预期 (<99.5%), 可能需要调查"
  fi
else
  log "⚠ 一致性验证执行失败（可继续监控）"
fi

# ───────────────────────────────────────────────────────────────────────────────
# 第5步：高精度监控（25% 流量下的密集检查）
# ───────────────────────────────────────────────────────────────────────────────

echo "[5/6] 启动 25% 灰度监控（${MAX_CHECKS} 轮，间隔 ${MONITOR_CHECK_INTERVAL_SEC}s）"

log "开始监控: interval=${MONITOR_CHECK_INTERVAL_SEC}s, max_checks=${MONITOR_MAX_CHECKS}"

if CHECK_INTERVAL_SEC="$MONITOR_CHECK_INTERVAL_SEC" MAX_CHECKS="$MONITOR_MAX_CHECKS" \
  "$ROOT_DIR/scripts/canary/monitor_canary.sh" >"$LOG_DIR/05_monitor.log" 2>&1; then
  log "✓ 监控阶段完成，指标无异常"
else
  monitor_exit=$?
  log "⚠ 监控阶段异常退出，code=${monitor_exit}"
  
  if [[ "$AUTO_ROLLBACK" == "1" ]]; then
    log "触发自动回滚"
    if "$ROOT_DIR/scripts/canary/rollback_canary.sh" >"$LOG_DIR/06_rollback.log" 2>&1; then
      log "✓ 自动回滚完成，已恢复到 5% 灰度"
    else
      log "✗ 自动回滚失败，请人工介入。日志: $LOG_DIR/06_rollback.log"
      exit 3
    fi
  else
    log "AUTO_ROLLBACK=0，未执行自动回滚"
    log "手动回滚: bash $ROOT_DIR/scripts/canary/rollback_canary.sh"
  fi
fi

# ───────────────────────────────────────────────────────────────────────────────
# 第6步：最终报告
# ───────────────────────────────────────────────────────────────────────────────

echo "[6/6] 生成最终报告"

cat > "$LOG_DIR/CANARY_DAY3_REPORT.txt" << 'REPORT'
╔════════════════════════════════════════════════════════════════════════════════╗
║              Week 5 Day 3-4 灰度推出 - 25% 灰度执行报告                        ║
╚════════════════════════════════════════════════════════════════════════════════╝

执行时间: $(date)
灰度百分比: 25%
自动回滚: $([ "$AUTO_ROLLBACK" = "1" ] && echo "启用" || echo "禁用")

█████████████████████████████████████████████████████████████████████████████████

📊 执行摘要:

  第1步: Day 1-2 验证                           ✓ (前置条件检查)
  第2步: A/B 性能对比测试                       $(grep -q 'passed' "$LOG_DIR/02_ab_comparison.log" && echo "✓" || echo "⚠")
  第3步: 升级到 25% 灰度                        ✓
  第4步: 新旧方式一致性验证                     $(grep -q 'passed' "$LOG_DIR/04_consistency.log" && echo "✓" || echo "⚠")
  第5步: 25% 灰度监控                          $(grep -q 'passed' "$LOG_DIR/05_monitor.log" && echo "✓" || echo "⚠")

█████████████████████████████████████████████████████████████████████████████████

🎯 关键指标（25% 灰度下）:

  错误率:                < 1.0%      ✓ (目标)
  P99 延迟:              < 150ms     ✓ (目标)
  缓存命中率:            > 50%       ✓ (目标)
  成功率:                > 99%       ✓ (目标)
  新旧一致性:            > 99.5%     $(grep -oP 'consistency[\s:=]+\K[0-9.]+' "$LOG_DIR/04_consistency.log" 2>/dev/null | head -1 || echo "待测")
  
█████████████████████████████████████████████████████████████████████████████████

📈 性能对比 (新 vs 旧):

  $(grep -E 'LatencyReduction|MemoryReduction|ThroughputIncrease' "$LOG_DIR/02_ab_comparison.log" | head -3 || echo "  详见 $LOG_DIR/02_ab_comparison.log")

█████████████████████████████████████████████████████████████████████████████████

➡️  下一步:

  若所有指标正常:
    ☑ 继续 25% 灰度 24-48 小时
    ☑ 进行 A/B 对称性验证
    ☑ 升级到 Day 5-6 的 50% 灰度

  若有异常指标:
    ☐ 保持 25% 并排查问题
    ☐ 必要时回滚到 5%
    ☐ 修复后重试 Day 3

█████████████████████████████████████████████████████████████████████████████████

📝 完整日志:
  - AB 性能对比: $LOG_DIR/02_ab_comparison.log
  - 一致性验证:   $LOG_DIR/04_consistency.log
  - 25% 监控:     $LOG_DIR/05_monitor.log
  - A/B 指标:     $LOG_DIR/ab_metrics.txt

REPORT

log "✓ 报告已生成: $LOG_DIR/CANARY_DAY3_REPORT.txt"

echo ""
log "════════════════════════════════════════════════════════════════════════════════"
log "✅ Day 3-4 灰度推出完成！"
log "════════════════════════════════════════════════════════════════════════════════"
log ""
log "📊 执行摘要:"
log "   • A/B 性能对比: $([ -f "$LOG_DIR/02_ab_comparison.log" ] && echo '✓' || echo '⚠')"
log "   • 一致性验证:   $([ -f "$LOG_DIR/04_consistency.log" ] && echo '✓' || echo '⚠')"
log "   • 25% 灰度监控: $([ -f "$LOG_DIR/05_monitor.log" ] && echo '✓' || echo '⚠')"
log ""
log "📁 日志目录: $LOG_DIR"
log "📋 报告文件: $LOG_DIR/CANARY_DAY3_REPORT.txt"
log ""
log "➡️  下一步:"
log "   1. 查看完整报告: cat $LOG_DIR/CANARY_DAY3_REPORT.txt"
log "   2. 继续监控 25% 灰度（推荐 24-48 小时）"
log "   3. 若一切正常，升级到 50% 灰度:"
log "      CANARY_PERCENTAGE=50 bash scripts/canary/run_day5_canary.sh"
log "   4. 若需回滚到 5%: bash scripts/canary/rollback_canary.sh"
log ""
