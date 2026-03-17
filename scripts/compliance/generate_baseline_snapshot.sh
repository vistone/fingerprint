#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
REPORT_DIR="${ROOT_DIR}/docs/reports"
REPORT_FILE="${REPORT_DIR}/CODE_COMPLIANCE_BASELINE.md"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

mkdir -p "${REPORT_DIR}"

run_check() {
  local label="$1"
  shift
  local out="${TMP_DIR}/${label}.txt"
  if "$@" >"${out}" 2>&1; then
    echo "PASS" >"${TMP_DIR}/${label}.status"
  else
    echo "FAIL" >"${TMP_DIR}/${label}.status"
  fi
}

cd "${ROOT_DIR}"

run_check "comments" go run ./cmd/compliance --root modules --checks comments --include-tests
run_check "file_length_non_test" go run ./cmd/compliance --root modules --checks file-length
run_check "file_length_tests" go run ./cmd/compliance --root modules --checks file-length --include-tests
run_check "func_length_non_test" go run ./cmd/compliance --root modules --checks func-length
run_check "params_non_test" go run ./cmd/compliance --root modules --checks params
run_check "lint" golangci-lint run
run_check "test" ./scripts/compliance/run_all_tests.sh

{
  echo "# Code Compliance Baseline Snapshot"
  echo
  echo "- Generated at: $(date '+%Y-%m-%d %H:%M:%S %Z')"
  echo "- Source rules: \`docs/DEVELOPER_GUIDE.md\`, \`docs/CONTRIBUTING.md\`"
  echo "- Scope: \`modules/**/*.go\` for blocking checks, first-round standards consistency, no behavior changes"
  echo
  echo "## BLOCKING Severity"
  echo
  echo "### Chinese comments in Go code (includes tests)"
  echo "- Status: $(cat "${TMP_DIR}/comments.status")"
  echo '```text'
  cat "${TMP_DIR}/comments.txt"
  echo '```'
  echo
  echo "### File length > 500 (non-test)"
  echo "- Status: $(cat "${TMP_DIR}/file_length_non_test.status")"
  echo '```text'
  cat "${TMP_DIR}/file_length_non_test.txt"
  echo '```'
  echo
  echo "### File length > 500 (test files, reference)"
  echo "- Status: $(cat "${TMP_DIR}/file_length_tests.status")"
  echo '```text'
  cat "${TMP_DIR}/file_length_tests.txt"
  echo '```'
  echo
  echo "### Function length > 80 (non-test)"
  echo "- Status: $(cat "${TMP_DIR}/func_length_non_test.status")"
  echo '```text'
  cat "${TMP_DIR}/func_length_non_test.txt"
  echo '```'
  echo
  echo "### Function params > 5 (non-test)"
  echo "- Status: $(cat "${TMP_DIR}/params_non_test.status")"
  echo '```text'
  cat "${TMP_DIR}/params_non_test.txt"
  echo '```'
  echo
  echo "### Lint status"
  echo "- Status: $(cat "${TMP_DIR}/lint.status")"
  echo '```text'
  cat "${TMP_DIR}/lint.txt"
  echo '```'
  echo
  echo "### go test ./... (root + all modules)"
  echo "- Status: $(cat "${TMP_DIR}/test.status")"
  echo '```text'
  cat "${TMP_DIR}/test.txt"
  echo '```'
  echo
  echo "## WARNING Severity (Advisory)"
  echo
  echo "### High-risk module coverage snapshot"
  for module in gateway ml agent waf; do
    echo "- modules/${module}:"
    if out=$(go test -cover "./modules/${module}/..." 2>&1); then
      echo '```text'
      echo "${out}" | tail -n 5
      echo '```'
    else
      echo '```text'
      echo "${out}" | tail -n 20
      echo '```'
    fi
  done
  echo
  echo "### Benchmark advisory snapshot"
  for module in core profiles ml tls; do
    echo "- modules/${module}:"
    if out=$(go test -bench=. -benchmem -run=^$ -benchtime=1x "./modules/${module}/..." 2>&1); then
      echo '```text'
      echo "${out}" | tail -n 5
      echo '```'
    else
      echo '```text'
      echo "${out}" | tail -n 20
      echo '```'
    fi
  done
  echo
  echo "### Architecture optimization backlog"
  echo "- Reference: \`CODE_ANALYSIS_REPORT.md\` (tracked as advisory only in this round)"
} >"${REPORT_FILE}"

echo "Baseline snapshot written to ${REPORT_FILE}"
