#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${KUBECTL_NAMESPACE:-production}"
CONFIGMAP="${CANARY_CONFIGMAP:-pipeline-config}"
DEPLOYMENT="${CANARY_DEPLOYMENT:-processing-engine}"
RUN_FULL_TESTS="${RUN_FULL_TESTS:-0}"

echo "[1/8] Go 编译检查"
go build ./...

echo "[2/8] 灰度框架测试"
go test ./internal/extension -run '^TestCanary' -count=1

echo "[3/8] 关键集成测试"
go test ./test -run 'config_management|integration' -count=1

if [[ "$RUN_FULL_TESTS" == "1" ]]; then
  echo "[4/8] 全量测试"
  go test ./... -count=1
else
  echo "[4/8] 跳过全量测试（设置 RUN_FULL_TESTS=1 可开启）"
fi

echo "[5/8] Kubernetes 连通性"
kubectl -n "$NAMESPACE" get ns "$NAMESPACE" >/dev/null

echo "[6/8] 配置检查"
kubectl -n "$NAMESPACE" get configmap "$CONFIGMAP" >/dev/null

echo "[7/8] 部署检查"
kubectl -n "$NAMESPACE" get deploy "$DEPLOYMENT" >/dev/null

echo "[8/8] 初始灰度状态（建议为 disabled 或 <=5%）"
kubectl -n "$NAMESPACE" get configmap "$CONFIGMAP" -o jsonpath='{.data.canary\.enabled}{"\n"}{.data.canary\.percentage}{"\n"}{.data.canary\.stage}{"\n"}' || true

echo "✅ Day1 预检通过"
