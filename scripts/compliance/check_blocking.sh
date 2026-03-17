#!/usr/bin/env bash

set -euo pipefail

echo "BLOCKING: gofmt check (no auto-fix)"
if [ -n "$(gofmt -s -l .)" ]; then
  echo "BLOCKING: gofmt found unformatted files"
  gofmt -s -d .
  exit 1
fi

echo "BLOCKING: code standards checks"
go run ./cmd/compliance \
  --root modules \
  --checks comments,file-length,func-length,params \
  --max-file-lines 500 \
  --max-func-lines 80 \
  --max-params 5

echo "BLOCKING: golangci-lint"
golangci-lint run

echo "BLOCKING: go test ./... (root + all modules)"
./scripts/compliance/run_all_tests.sh
