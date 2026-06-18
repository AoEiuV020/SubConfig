#!/usr/bin/env bash
set -euo pipefail

source "$(dirname "$0")/env.sh"

echo 准备SubConfig本地副本，
rm -rf "$SUBCONFIG_DIR"
mkdir -p "$SUBCONFIG_DIR"
tar -cf - \
    --exclude .git \
    --exclude tmp \
    --exclude config-depot/data \
    -C "$ROOT_DIR" . | tar -xf - -C "$SUBCONFIG_DIR"
