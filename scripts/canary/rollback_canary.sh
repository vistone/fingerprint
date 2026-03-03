#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${KUBECTL_NAMESPACE:-production}"
CONFIGMAP="${CANARY_CONFIGMAP:-pipeline-config}"
DEPLOYMENT="${CANARY_DEPLOYMENT:-processing-engine}"
BASE_URL="${CANARY_METRICS_BASE_URL:-http://localhost:8080}"

TS="$(date +%Y%m%d_%H%M%S)"
OUT_DIR="/tmp/canary-rollback-${TS}"
mkdir -p "$OUT_DIR"

echo "[1/5] 禁用灰度流量"
kubectl -n "$NAMESPACE" patch configmap "$CONFIGMAP" --type merge -p \
  '{"data":{"canary.enabled":"false","canary.percentage":"0.00","canary.stage":"off"}}'

echo "[2/5] 验证当前配置"
kubectl -n "$NAMESPACE" get configmap "$CONFIGMAP" -o jsonpath='{.data.canary\.enabled}{"\n"}{.data.canary\.percentage}{"\n"}{.data.canary\.stage}{"\n"}' | tee "$OUT_DIR/canary-config.txt"

echo "[3/5] 收集诊断数据"
kubectl -n "$NAMESPACE" logs deploy/"$DEPLOYMENT" --tail=2000 > "$OUT_DIR/engine.log" || true
kubectl -n "$NAMESPACE" top pods -l app="$DEPLOYMENT" > "$OUT_DIR/resources.txt" || true
curl -fsS "$BASE_URL/metrics/summary" > "$OUT_DIR/metrics-summary.txt" || true

echo "[4/5] 快速健康检查"
curl -fsS "$BASE_URL/metrics/error-rate" | tee "$OUT_DIR/error-rate.txt" || true

echo "[5/5] 回滚完成"
echo "✅ 回滚完成，诊断数据目录: $OUT_DIR"
