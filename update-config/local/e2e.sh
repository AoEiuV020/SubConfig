#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="${ROOT_DIR:-$PWD}"
export ROOT_DIR

source "$(dirname "$0")/env.sh"

cleanup() {
    "$ROOT_DIR/update-config/local/stop-services.sh"
}
trap cleanup EXIT

"$ROOT_DIR/update-config/local/prepare-workspace.sh"
"$ROOT_DIR/update-config/local/start-config-depot.sh"
"$ROOT_DIR/update-config/run-subconverter.sh"
"$ROOT_DIR/update-config/local/prepare-subconfig-copy.sh"
"$ROOT_DIR/update-config/cache-external-config.sh"
"$ROOT_DIR/update-config/update-config.sh"
"$ROOT_DIR/update-config/compress-config.sh"
"$ROOT_DIR/update-config/deploy-config.sh"
"$ROOT_DIR/update-config/local/verify-deploy.sh"
