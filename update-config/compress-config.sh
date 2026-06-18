#!/usr/bin/env bash
set -euo pipefail

UPLOAD_SECRET_FILE="${UPLOAD_SECRET_FILE:-upload_secret}"
CONFIG_OUTPUT_DIR="${CONFIG_OUTPUT_DIR:-config}"
CONFIG_ARCHIVE_FILE="${CONFIG_ARCHIVE_FILE:-config.tar.gz}"
ENCRYPTED_CONFIG_FILE="${ENCRYPTED_CONFIG_FILE:-config.tar.gz.aes}"
COPYFILE_DISABLE=1
export COPYFILE_DISABLE

if [[ "$CONFIG_ARCHIVE_FILE" != /* ]]; then
    CONFIG_ARCHIVE_FILE="$PWD/$CONFIG_ARCHIVE_FILE"
fi
if [[ "$ENCRYPTED_CONFIG_FILE" != /* ]]; then
    ENCRYPTED_CONFIG_FILE="$PWD/$ENCRYPTED_CONFIG_FILE"
fi

ensure_parent() {
    local path="$1"
    local parent
    parent=$(dirname "$path")
    if [[ "$parent" != "." ]]; then
        mkdir -p "$parent"
    fi
}

base64_decode() {
    if printf '' | base64 -d >/dev/null 2>&1; then
        base64 -d
    else
        base64 -D
    fi
}

echo 打包压缩所有生成的配置文件，
(
    cd "$CONFIG_OUTPUT_DIR"
    shopt -s nullglob
    files=(*)
    if (( ${#files[@]} == 0 )); then
        echo 没有可打包的配置文件，
        exit 1
    fi
    tar -zcf "$CONFIG_ARCHIVE_FILE" "${files[@]}"
)

echo 加密压缩包，
# key是base64字符串，iv每次随机生成，转成16进制交给openssl，避免shell变量保存二进制数据，
iv=$(head -c 16 /dev/urandom | base64)
if [[ -r "$UPLOAD_SECRET_FILE" ]]; then
    key=$(tr -d '\r\n' <"$UPLOAD_SECRET_FILE")
else
    key=$(head -c 32 /dev/urandom | base64)
    echo generate random secret at "$UPLOAD_SECRET_FILE"
    ensure_parent "$UPLOAD_SECRET_FILE"
    printf '%s\n' "$key" >"$UPLOAD_SECRET_FILE"
fi
key_hex=$(printf '%s' "$key" | base64_decode | od -A n -v -t x1 | tr -d ' \n')
iv_hex=$(printf '%s' "$iv" | base64_decode | od -A n -v -t x1 | tr -d ' \n')
ensure_parent "$ENCRYPTED_CONFIG_FILE"
printf '%s' "$iv" | base64_decode >"$ENCRYPTED_CONFIG_FILE"
openssl enc -aes-256-cbc -K "$key_hex" -iv "$iv_hex" -nosalt <"$CONFIG_ARCHIVE_FILE" >>"$ENCRYPTED_CONFIG_FILE"
