#!/usr/bin/env bash

set -euo pipefail

echo "Building server..."

go build -o ./bin/wispy-core ../cmd/server/main.go

echo "Server build complete."