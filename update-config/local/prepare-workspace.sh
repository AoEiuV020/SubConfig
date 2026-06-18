#!/usr/bin/env bash
set -euo pipefail

source "$(dirname "$0")/env.sh"

echo 准备本地验收工作区，
mkdir -p "$WORK_DIR"
rm -rf "$SUBCONVERTER_DIR" "$CONFIG_OUTPUT_DIR" "$WORK_DIR/config-decrypt"
mkdir -p "$WORK_DIR"
