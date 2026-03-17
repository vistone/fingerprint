#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${1:-http://localhost:8080}"
INTERVAL_SEC="${INTERVAL_SEC:-10}"
MAX_CHECKS="${MAX_CHECKS:-30}"

echo "Monitoring ML canary stats from ${BASE_URL}"
echo "interval=${INTERVAL_SEC}s checks=${MAX_CHECKS}"

for ((i=1; i<=MAX_CHECKS; i++)); do
  payload="$(curl -fsS "${BASE_URL}/api/admin/ml/service/stats")"
  now="$(date '+%F %T')"

  mode="$(jq -r '.inferenceMode // "unknown"' <<< "${payload}")"
  canary_enabled="$(jq -r '.canary.enabled // false' <<< "${payload}")"
  canary_rate="$(jq -r '.canary.canaryRate // 0' <<< "${payload}")"
  canary_total="$(jq -r '.canary.totalRequests // 0' <<< "${payload}")"
  canary_routed="$(jq -r '.canary.canaryRouted // 0' <<< "${payload}")"
  canary_fallback="$(jq -r '.canary.fallbackCount // 0' <<< "${payload}")"

  shadow_sampled="$(jq -r '.shadowCompare.sampledCount // 0' <<< "${payload}")"
  shadow_errors="$(jq -r '.shadowCompare.errorCount // 0' <<< "${payload}")"
  browser_agree="$(jq -r '.shadowCompare.browserTop1AgreeRate // 0' <<< "${payload}")"
  action_agree="$(jq -r '.shadowCompare.actionTop1AgreeRate // 0' <<< "${payload}")"
  forgery_delta="$(jq -r '.shadowCompare.avgForgeryProbDelta // 0' <<< "${payload}")"
  threat_delta="$(jq -r '.shadowCompare.avgThreatProbDelta // 0' <<< "${payload}")"

  echo "[${now}] mode=${mode} canary_enabled=${canary_enabled} canary_rate=${canary_rate} total=${canary_total} routed=${canary_routed} fallback=${canary_fallback} shadow_sampled=${shadow_sampled} shadow_errors=${shadow_errors} browser_agree=${browser_agree} action_agree=${action_agree} forgery_delta=${forgery_delta} threat_delta=${threat_delta}"

  if [[ "${i}" -lt "${MAX_CHECKS}" ]]; then
    sleep "${INTERVAL_SEC}"
  fi
done
