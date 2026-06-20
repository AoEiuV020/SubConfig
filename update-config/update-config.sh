#!/usr/bin/env bash
set -euo pipefail

SUBSCRIBE_FILE="${SUBSCRIBE_FILE:-subscribe}"
SUBSCRIBE_JSON_FILE="${SUBSCRIBE_JSON_FILE:-subscribe.json}"
SUBCONVERTER_DIR="${SUBCONVERTER_DIR:-subconverter}"
SUBCONVERTER_SUB_DIR="${SUBCONVERTER_SUB_DIR:-$SUBCONVERTER_DIR/sub}"
CONFIG_OUTPUT_DIR="${CONFIG_OUTPUT_DIR:-config}"
DEFAULT_CONFIG="${DEFAULT_CONFIG:-_SubConfig/subconverter.ini}"
SUBCONVERTER_URL="${SUBCONVERTER_URL:-http://127.0.0.1:25500/sub}"
SUBCONVERTER_PARAMS="${SUBCONVERTER_PARAMS:-emoji=true&list=false&udp=false&tfo=false&scv=false&fdn=false&sort=false&new_name=true}"
CONFIG_SUFFIXES="${CONFIG_SUFFIXES:---default--}"
CONFIG_TARGETS="${CONFIG_TARGETS:-clash quan v2ray ssr singbox}"

if [[ "$CONFIG_SUFFIXES" == "--default--" ]]; then
    config_suffixes=("" "-basic" "-noban" "-ban")
else
    read -r -a config_suffixes <<<"$CONFIG_SUFFIXES"
fi
read -r -a config_targets <<<"$CONFIG_TARGETS"

mkdir -p "$SUBCONVERTER_SUB_DIR"
echo 先整理出订阅数组，
jq -srR 'split("\n") | map(select(length > 0))' "$SUBSCRIBE_FILE" >"$SUBSCRIBE_JSON_FILE"

echo 再下载节点列表，方便后面复用避免重复请求机场订阅，
jq -r 'to_entries[] | select(.value | startswith("http") and (startswith("https://t.me/") | not)) | [.key, .value] | @tsv' "$SUBSCRIBE_JSON_FILE" |
while IFS=$'\t' read -r index subscribe_url; do
    output_file="$SUBCONVERTER_SUB_DIR/$index"
    echo 下载订阅: "$index"
    curl -s -L --fail -o "$output_file" "$subscribe_url"
    bytes=$(wc -c <"$output_file" | tr -d ' ')
    echo 订阅缓存完成: "$index" "$bytes" 字节
done

echo 订阅转成本地请求，其他链接保留，
url=$(jq -r 'to_entries | map(if (.value | startswith("http") and (startswith("https://t.me/") | not)) then (.key | tostring | "sub/" + .) else .value end) | join("|") | @uri' "$SUBSCRIBE_JSON_FILE")

echo 拼接自己需要的配置请求，
mkdir -p "$CONFIG_OUTPUT_DIR"
for suffix in "${config_suffixes[@]}"; do
    for target in "${config_targets[@]}"; do
        config_path="${DEFAULT_CONFIG/%.ini/${suffix}.ini}"
        config=$(printf '%s' "$config_path" | jq -rR @uri)
        output_file="$CONFIG_OUTPUT_DIR/$target$suffix"
        code=$(curl -s -L -o "$output_file" -w '%{http_code}' "$SUBCONVERTER_URL?target=$target&url=$url&config=$config&$SUBCONVERTER_PARAMS")
        if [[ "$code" != 200 || ! -s "$output_file" ]]; then
            echo 订阅转换异常，
            wc -c "$output_file" || true
            cat "$output_file" || true
            exit 1
        fi
    done
done
