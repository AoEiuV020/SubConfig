#!/usr/bin/env bash
set -euo pipefail

UPLOAD_SECRET_FILE="${UPLOAD_SECRET_FILE:-upload_secret}"
ENCRYPTED_CONFIG_FILE="${ENCRYPTED_CONFIG_FILE:-config.tar.gz.aes}"
DECRYPT_WORK_DIR="${DECRYPT_WORK_DIR:-update-config/tmp/decrypt}"
UPLOAD_IV_FILE="${UPLOAD_IV_FILE:-$DECRYPT_WORK_DIR/upload.iv}"
CIPHERTEXT_FILE="${CIPHERTEXT_FILE:-$DECRYPT_WORK_DIR/config.tar.gz.ciphertext}"
DECRYPTED_ARCHIVE_FILE="${DECRYPTED_ARCHIVE_FILE:-$DECRYPT_WORK_DIR/config-decrypt.tar.gz}"
DECRYPTED_CONFIG_DIR="${DECRYPTED_CONFIG_DIR:-$DECRYPT_WORK_DIR/config-decrypt}"

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

key=$(tr -d '\r\n' <"$UPLOAD_SECRET_FILE")
ensure_parent "$UPLOAD_IV_FILE"
ensure_parent "$CIPHERTEXT_FILE"
ensure_parent "$DECRYPTED_ARCHIVE_FILE"
head -c 16 "$ENCRYPTED_CONFIG_FILE" >"$UPLOAD_IV_FILE"
tail -c +17 "$ENCRYPTED_CONFIG_FILE" >"$CIPHERTEXT_FILE"
key_hex=$(printf '%s' "$key" | base64_decode | od -A n -v -t x1 | tr -d ' \n')
iv_hex=$(od -A n -v -t x1 "$UPLOAD_IV_FILE" | tr -d ' \n')
openssl enc -d -aes-256-cbc -K "$key_hex" -iv "$iv_hex" -nosalt <"$CIPHERTEXT_FILE" >"$DECRYPTED_ARCHIVE_FILE"
mkdir -p "$DECRYPTED_CONFIG_DIR"
tar -zxf "$DECRYPTED_ARCHIVE_FILE" -C "$DECRYPTED_CONFIG_DIR"
