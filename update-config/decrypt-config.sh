#!/usr/bin/env bash
set -euo pipefail

UPLOAD_SECRET_FILE="${UPLOAD_SECRET_FILE:-upload_secret}"
ENCRYPTED_CONFIG_FILE="${ENCRYPTED_CONFIG_FILE:-config.tar.gz.aes}"
UPLOAD_IV_FILE="${UPLOAD_IV_FILE:-upload.iv}"
CIPHERTEXT_FILE="${CIPHERTEXT_FILE:-config.tar.gz.ciphertext}"
DECRYPTED_ARCHIVE_FILE="${DECRYPTED_ARCHIVE_FILE:-config-decrypt.tar.gz}"
DECRYPTED_CONFIG_DIR="${DECRYPTED_CONFIG_DIR:-config-decrypt}"

base64_decode() {
    if printf '' | base64 -d >/dev/null 2>&1; then
        base64 -d
    else
        base64 -D
    fi
}

key=$(tr -d '\r\n' <"$UPLOAD_SECRET_FILE")
head -c 16 "$ENCRYPTED_CONFIG_FILE" >"$UPLOAD_IV_FILE"
tail -c +17 "$ENCRYPTED_CONFIG_FILE" >"$CIPHERTEXT_FILE"
key_hex=$(printf '%s' "$key" | base64_decode | od -A n -v -t x1 | tr -d ' \n')
iv_hex=$(od -A n -v -t x1 "$UPLOAD_IV_FILE" | tr -d ' \n')
openssl enc -d -aes-256-cbc -K "$key_hex" -iv "$iv_hex" -nosalt <"$CIPHERTEXT_FILE" >"$DECRYPTED_ARCHIVE_FILE"
mkdir -p "$DECRYPTED_CONFIG_DIR"
tar -zxf "$DECRYPTED_ARCHIVE_FILE" -C "$DECRYPTED_CONFIG_DIR"
