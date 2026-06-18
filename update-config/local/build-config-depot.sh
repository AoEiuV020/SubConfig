#!/usr/bin/env bash
set -euo pipefail

source "$(dirname "$0")/env.sh"

mkdir -p "$(dirname "$CONFIG_DEPOT_BINARY")"
echo 编译config-depot服务，
(
    cd "$CONFIG_DEPOT_SOURCE_DIR"
    go build -o "$CONFIG_DEPOT_BINARY" ./cmd/config-depot
)
