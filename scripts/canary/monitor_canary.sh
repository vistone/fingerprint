#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${CANARY_METRICS_BASE_URL:-http://localhost:8080}"
CHECK_INTERVAL_SEC="${CHECK_INTERVAL_SEC:-300}"
MAX_CHECKS="${MAX_CHECKS:-0}"
CRITICAL_ERROR_RATE="${CRITICAL_ERROR_RATE:-0.03}"
CRITICAL_P99_MS="${CRITICAL_P99_MS:-500}"

checks=0

extract_number() {
  grep -Eo '[0-9]+(\.[0-9]+)?' | head -n1
}

echo "开始灰度监控: $BASE_URL, 间隔 ${CHECK_INTERVAL_SEC}s"

while true; do
  checks=$((checks + 1))
  now="$(date '+%F %T')"

  error_rate_raw="$(curl -fsS "$BASE_URL/metrics/error-rate" || echo "0")"
  p99_raw="$(curl -fsS "$BASE_URL/metrics/p99-latency" || echo "0")"
  cache_raw="$(curl -fsS "$BASE_URL/metrics/cache-hitrate" || echo "0")"

  error_rate="$(echo "$error_rate_raw" | extract_number || echo "0")"
  p99_ms="$(echo "$p99_raw" | extract_number || echo "0")"
  cache_hit="$(echo "$cache_raw" | extract_number || echo "0")"

  printf '[%s] error_rate=%s p99_ms=%s cache_hit=%s\n' "$now" "$error_rate" "$p99_ms" "$cache_hit"

  critical_error="$(awk -v a="$error_rate" -v b="$CRITICAL_ERROR_RATE" 'BEGIN{print (a>b)?1:0}')"
  critical_p99="$(awk -v a="$p99_ms" -v b="$CRITICAL_P99_MS" 'BEGIN{print (a>b)?1:0}')"

  if [[ "$critical_error" == "1" || "$critical_p99" == "1" ]]; then
    echo "❌ 触发严重阈值，建议立即执行回滚脚本"
    exit 2
  fi

  if [[ "$MAX_CHECKS" != "0" && "$checks" -ge "$MAX_CHECKS" ]]; then
    echo "✅ 达到监控轮次上限，监控结束"
    exit 0
  fi

  sleep "$CHECK_INTERVAL_SEC"
done
