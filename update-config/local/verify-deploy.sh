#!/usr/bin/env bash
set -euo pipefail

source "$(dirname "$0")/env.sh"

echo 验证上传结果，
expected_count=20
actual_count=$(find "$CONFIG_DEPOT_DATA_DIR/config" -maxdepth 1 -type f | wc -l | tr -d ' ')
if [[ "$actual_count" != "$expected_count" ]]; then
    echo 配置文件数量异常: "$actual_count"
    find "$CONFIG_DEPOT_DATA_DIR/config" -maxdepth 1 -type f -print
    exit 6
fi

for name in clash clash-basic clash-noban clash-ban quan v2ray ssr singbox; do
    if [[ ! -s "$CONFIG_DEPOT_DATA_DIR/config/$name" ]]; then
        echo 配置文件为空: "$name"
        exit 7
    fi
done

secret=$(tr -d '\r\n' <"$CONFIG_DEPOT_DATA_DIR/sub_secret")
curl -s -L --fail -o "$WORK_DIR/sub-check-clash" "http://$CONFIG_DEPOT_ADDR/sub?target=clash&url=$secret"
if ! cmp -s "$CONFIG_DEPOT_DATA_DIR/config/clash" "$WORK_DIR/sub-check-clash"; then
    echo /sub缓存读取结果和config/clash不一致
    exit 8
fi

echo 端到端验收通过，
