#!/usr/bin/env bash
set -euo pipefail

STAGE="${1:-}"
NAMESPACE="${KUBECTL_NAMESPACE:-production}"
CONFIGMAP="${CANARY_CONFIGMAP:-pipeline-config}"

if [[ -z "$STAGE" ]]; then
  echo "用法: $0 <5|25|50|100|off>"
  exit 1
fi

case "$STAGE" in
  5)
    ENABLED="true"; PERCENTAGE="0.05"; LABEL="5%" ;;
  25)
    ENABLED="true"; PERCENTAGE="0.25"; LABEL="25%" ;;
  50)
    ENABLED="true"; PERCENTAGE="0.50"; LABEL="50%" ;;
  100)
    ENABLED="true"; PERCENTAGE="1.00"; LABEL="100%" ;;
  off)
    ENABLED="false"; PERCENTAGE="0.00"; LABEL="off" ;;
  *)
    echo "不支持的阶段: $STAGE（仅支持 5/25/50/100/off）"
    exit 1 ;;
esac

echo "设置灰度阶段: $LABEL"
kubectl -n "$NAMESPACE" patch configmap "$CONFIGMAP" --type merge -p \
  "{\"data\":{\"canary.enabled\":\"$ENABLED\",\"canary.percentage\":\"$PERCENTAGE\",\"canary.stage\":\"$LABEL\"}}"

echo "当前灰度配置:"
kubectl -n "$NAMESPACE" get configmap "$CONFIGMAP" -o jsonpath='{.data.canary\.enabled}{"\n"}{.data.canary\.percentage}{"\n"}{.data.canary\.stage}{"\n"}'

echo "✅ 阶段切换完成"
