#!/usr/bin/env bash
set -euo pipefail

source "$(dirname "$0")/env.sh"

if [[ -f "$SUBCONVERTER_PID_FILE" ]]; then
    kill "$(cat "$SUBCONVERTER_PID_FILE")" >/dev/null 2>&1 || true
fi

if [[ -f "$CONFIG_DEPOT_PID_FILE" ]]; then
    kill "$(cat "$CONFIG_DEPOT_PID_FILE")" >/dev/null 2>&1 || true
fi
