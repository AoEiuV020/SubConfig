#!/usr/bin/env bash
set -euo pipefail

source "$(dirname "$0")/env.sh"

mkdir -p "$WORK_DIR"
echo "http://$CONFIG_DEPOT_ADDR/upload" >"$DEPLOY_URL_FILE"

if [[ ! -x "$CONFIG_DEPOT_BINARY" ]]; then
    "$(dirname "$0")/build-config-depot.sh"
fi

echo 启动config-depot，
CONFIG_DEPOT_ADDR="$CONFIG_DEPOT_ADDR" CONFIG_DEPOT_DATA_DIR="$CONFIG_DEPOT_DATA_DIR" "$CONFIG_DEPOT_BINARY" >"$CONFIG_DEPOT_LOG_FILE" 2>&1 &
echo "$!" >"$CONFIG_DEPOT_PID_FILE"

for _ in $(seq 1 30); do
    if curl -s -o /dev/null --fail "http://$CONFIG_DEPOT_ADDR/healthz"; then
        exit 0
    fi
    sleep 1
done

echo config-depot启动超时，日志如下：
cat "$CONFIG_DEPOT_LOG_FILE" || true
exit 5
