#!/bin/bash

# This script is used to fuzz the package.
# It iterates through all packages, finds fuzz tests, and fuzzes each one.
# It ensures that the package is fuzzed and that it is working correctly.

set -euo pipefail
# Iterate packages, find fuzz tests, and fuzz each one for a short time.
for pkg in $(go list ./...); do
  names=$(go test -list '^Fuzz' "$pkg" | grep '^Fuzz' || true)
  if [ -z "$names" ]; then
    continue
  fi
  for name in $names; do
    echo "Fuzzing $pkg::$name"
    go test "$pkg" -run=^$ -fuzz="$name" -fuzztime=10s
  done
done
