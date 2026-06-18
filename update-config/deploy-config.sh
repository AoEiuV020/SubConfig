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
if [[ ! -r "$DEPLOY_URL_FILE" ]]; then
    echo 上传地址文件不可读：$DEPLOY_URL_FILE
    exit 4
fi
deploy_url=$(tr -d '\r\n' <"$DEPLOY_URL_FILE")
if [[ -z "$deploy_url" ]]; then
    echo 上传地址为空：$DEPLOY_URL_FILE
    exit 4
fi
if [[ ! -r "$ENCRYPTED_CONFIG_FILE" ]]; then
    echo 配置压缩包不可读：$ENCRYPTED_CONFIG_FILE
    exit 4
fi
if [[ ! -s "$ENCRYPTED_CONFIG_FILE" ]]; then
    echo 配置压缩包为空：$ENCRYPTED_CONFIG_FILE
    exit 4
fi

if [[ -r "$UPLOAD_TOKEN_FILE" ]]; then
    token=$(tr -d '\r\n' <"$UPLOAD_TOKEN_FILE")
else
    token=$(head -c 32 /dev/urandom | od -A n -v -t x1 | tr -d ' \n')
    echo 生成随机上传令牌：$UPLOAD_TOKEN_FILE
    ensure_parent "$UPLOAD_TOKEN_FILE"
    printf '%s\n' "$token" >"$UPLOAD_TOKEN_FILE"
fi
if [[ -z "$token" ]]; then
    echo 上传令牌为空：$UPLOAD_TOKEN_FILE
    exit 4
fi

upload_bytes=$(wc -c <"$ENCRYPTED_CONFIG_FILE" | tr -d ' ')
echo 上传地址：$deploy_url
echo 上传文件：$ENCRYPTED_CONFIG_FILE
echo 上传文件大小：$upload_bytes 字节
echo 上传令牌长度：${#token}

set +e
code=$(curl --silent --show-error --location --connect-timeout 15 --max-time 120 -o "$DEPLOY_OUTPUT_FILE" -w '%{http_code}' "$deploy_url" -F "token=$token" -F "file=@$ENCRYPTED_CONFIG_FILE;filename=config.tar.gz.aes")
curl_status=$?
set -e
if (( curl_status != 0 )); then
    echo 上传请求失败：curl退出码 $curl_status
    echo 服务端响应内容：
    cat "$DEPLOY_OUTPUT_FILE" || true
    exit 4
fi
echo 上传HTTP状态：$code
if [[ "$code" != 200 ]]; then
    echo 上传配置异常，
    echo 服务端响应内容：
    cat "$DEPLOY_OUTPUT_FILE"
    exit 4
fi
echo 上传配置完成，
cat "$DEPLOY_OUTPUT_FILE"
