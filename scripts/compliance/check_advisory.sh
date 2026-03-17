#!/usr/bin/env bash

set -euo pipefail

warn() {
  echo "::warning::$1"
  echo "WARNING: $1"
}

echo "WARNING: advisory checks started"

high_risk_modules=(gateway ml agent waf)
for module in "${high_risk_modules[@]}"; do
  set +e
  output=$(go test -cover "./modules/${module}/..." 2>&1)
  status=$?
  set -e

  total=$(echo "${output}" | sed -n 's/.*coverage: \([0-9.]\+%\) of statements.*/\1/p' | tail -n 1)
  if [ -n "${total}" ]; then
    echo "WARNING: modules/${module} coverage ${total}"
    continue
  fi

  if [ "${status}" -ne 0 ]; then
    warn "coverage collection failed for modules/${module}: $(echo "${output}" | tail -n 1)"
    continue
  fi

  warn "coverage summary unavailable for modules/${module}"
done

for module in core profiles ml tls; do
  if ! bench_output=$(go test -bench=. -benchmem -run=^$ -benchtime=1x "./modules/${module}/..." 2>&1); then
    warn "benchmark smoke failed for modules/${module}: $(echo "${bench_output}" | tail -n 1)"
  else
    echo "WARNING: benchmark smoke completed for modules/${module}"
  fi
done

if [ -f "CODE_ANALYSIS_REPORT.md" ]; then
  warn "architecture optimizations from CODE_ANALYSIS_REPORT.md are tracked as backlog and are not blocking in this round"
fi

echo "WARNING: advisory checks completed"
