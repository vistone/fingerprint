#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
LOG_DIR="${CANARY_LOG_DIR:-/tmp/canary-day1-$(date +%Y%m%d_%H%M%S)}"
AUTO_ROLLBACK="${AUTO_ROLLBACK:-1}"
STAGE="${CANARY_STAGE:-5}"
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

log "Day1 灰度编排开始，日志目录: $LOG_DIR"

run_step "01_precheck" "$ROOT_DIR/scripts/canary/precheck_day1.sh"
run_step "02_set_stage_${STAGE}" "$ROOT_DIR/scripts/canary/set_canary_stage.sh" "$STAGE"

log "开始监控: interval=${MONITOR_CHECK_INTERVAL_SEC}s, max_checks=${MONITOR_MAX_CHECKS}"
if CHECK_INTERVAL_SEC="$MONITOR_CHECK_INTERVAL_SEC" MAX_CHECKS="$MONITOR_MAX_CHECKS" \
  "$ROOT_DIR/scripts/canary/monitor_canary.sh" >"$LOG_DIR/03_monitor.log" 2>&1; then
  log "监控阶段完成，未触发严重阈值"
else
  monitor_exit=$?
  log "监控阶段异常退出，code=${monitor_exit}"
  if [[ "$AUTO_ROLLBACK" == "1" ]]; then
    log "触发自动回滚"
    if "$ROOT_DIR/scripts/canary/rollback_canary.sh" >"$LOG_DIR/04_rollback.log" 2>&1; then
      log "自动回滚完成"
    else
      log "自动回滚失败，请人工介入。日志: $LOG_DIR/04_rollback.log"
      exit 3
    fi
  else
    log "AUTO_ROLLBACK=0，未执行自动回滚"
  fi
  exit 2
fi

log "Day1 灰度编排成功完成"
log "输出日志: $LOG_DIR"
