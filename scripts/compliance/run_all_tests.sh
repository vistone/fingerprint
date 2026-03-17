#!/usr/bin/env bash

set -euo pipefail

echo "Running root go test ./..."
go test ./...

for dir in modules/*/; do
  if [ -f "${dir}/go.mod" ]; then
    echo "Running go test for ${dir}"
    (
      cd "${dir}"
      go test ./...
    )
  fi
done
