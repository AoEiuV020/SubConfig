#!/usr/bin/env bash
set -euo pipefail

UPLOAD_TOKEN_FILE="${UPLOAD_TOKEN_FILE:-upload_token}"
DEPLOY_URL_FILE="${DEPLOY_URL_FILE:-deploy_url}"
ENCRYPTED_CONFIG_FILE="${ENCRYPTED_CONFIG_FILE:-config.tar.gz.aes}"
DEPLOY_OUTPUT_FILE="${DEPLOY_OUTPUT_FILE:-deploy_config.out}"

ensure_parent() {
    local path="$1"
    local parent
    parent=$(dirname "$path")
    if [[ "$parent" != "." ]]; then
        mkdir -p "$parent"
    fi
}

echo 发布配置文件压缩包，
if [[ -r "$UPLOAD_TOKEN_FILE" ]]; then
    token=$(tr -d '\r\n' <"$UPLOAD_TOKEN_FILE")
else
    token=$(head -c 32 /dev/urandom | od -A n -v -t x1 | tr -d ' \n')
    echo generate random token at "$UPLOAD_TOKEN_FILE"
    ensure_parent "$UPLOAD_TOKEN_FILE"
    printf '%s\n' "$token" >"$UPLOAD_TOKEN_FILE"
fi

code=$(curl -s -o "$DEPLOY_OUTPUT_FILE" -w '%{http_code}' "$(cat "$DEPLOY_URL_FILE")" -F "token=$token" -F "file=@$ENCRYPTED_CONFIG_FILE;filename=config.tar.gz.aes")
if [[ "$code" != 200 ]]; then
    echo 上传配置异常，
    cat "$DEPLOY_OUTPUT_FILE"
    exit 4
fi
