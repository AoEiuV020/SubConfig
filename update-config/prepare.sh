#!/usr/bin/env bash
set -euo pipefail

UPLOAD_SECRET_FILE="${UPLOAD_SECRET_FILE:-upload_secret}"
UPLOAD_TOKEN_FILE="${UPLOAD_TOKEN_FILE:-upload_token}"
DEPLOY_URL_FILE="${DEPLOY_URL_FILE:-deploy_url}"
SUBSCRIBE_FILE="${SUBSCRIBE_FILE:-subscribe}"

append_github_env() {
    local name="$1"
    local value="$2"
    if [[ -n "${GITHUB_ENV:-}" ]]; then
        printf '%s=%s\n' "$name" "$value" >>"$GITHUB_ENV"
    fi
}

write_optional_file() {
    local value="$1"
    local path="$2"
    local parent
    if [[ -z "$value" ]]; then
        return
    fi
    parent=$(dirname "$path")
    if [[ "$parent" != "." ]]; then
        mkdir -p "$parent"
    fi
    printf '%s\n' "$value" >"$path"
}

write_optional_file "${UPLOAD_SECRET:-}" "$UPLOAD_SECRET_FILE"
write_optional_file "${UPLOAD_TOKEN:-}" "$UPLOAD_TOKEN_FILE"
write_optional_file "${DEPLOY_URL:-}" "$DEPLOY_URL_FILE"
write_optional_file "${SUBSCRIBE:-}" "$SUBSCRIBE_FILE"

if [[ -r "$UPLOAD_SECRET_FILE" && -r "$UPLOAD_TOKEN_FILE" && -r "$DEPLOY_URL_FILE" && -r "$SUBSCRIBE_FILE" ]]; then
    echo 发布到指定地址，
    append_github_env deploy true
fi

if [[ ! -r "$SUBSCRIBE_FILE" ]]; then
    echo 上传到artifact,
    append_github_env artifact true
fi

if [[ ! -r "$SUBSCRIBE_FILE" ]]; then
    echo 没有节点，生成一个示例，
    printf '%s\n' 'tg://http?server=1.2.3.4&port=233&user=user&pass=pass&remarks=Example' >"$SUBSCRIBE_FILE"
fi
