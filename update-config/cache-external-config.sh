#!/usr/bin/env bash
set -euo pipefail

SUBCONVERTER_DIR="${SUBCONVERTER_DIR:-subconverter}"
SUBCONVERTER_BASE_PATH="${SUBCONVERTER_BASE_PATH:-_SubConfig}"
ACL4SSR_BASE_PATH="${ACL4SSR_BASE_PATH:-_ACL4SSR}"
META_RULES_DAT_BASE_PATH="${META_RULES_DAT_BASE_PATH:-_meta-rules-dat}"
SUBCONFIG_DIR="${SUBCONFIG_DIR:-$SUBCONVERTER_DIR/$SUBCONVERTER_BASE_PATH}"
ACL4SSR_DIR="${ACL4SSR_DIR:-$SUBCONVERTER_DIR/$ACL4SSR_BASE_PATH}"
META_RULES_DAT_DIR="${META_RULES_DAT_DIR:-$SUBCONVERTER_DIR/$META_RULES_DAT_BASE_PATH}"
ACL4SSR_ARCHIVE="${ACL4SSR_ARCHIVE:-ACL4SSR.tar.gz}"
ACL4SSR_ARCHIVE_URL="${ACL4SSR_ARCHIVE_URL:-https://github.com/ACL4SSR/ACL4SSR/archive/refs/heads/master.tar.gz}"
ACL4SSR_EXTRACTED_DIR="${ACL4SSR_EXTRACTED_DIR:-ACL4SSR-master}"
META_RULES_DAT_ARCHIVE="${META_RULES_DAT_ARCHIVE:-meta-rules-dat.tar.gz}"
META_RULES_DAT_ARCHIVE_URL="${META_RULES_DAT_ARCHIVE_URL:-https://github.com/MetaCubeX/meta-rules-dat/archive/refs/heads/meta.tar.gz}"
META_RULES_DAT_EXTRACTED_DIR="${META_RULES_DAT_EXTRACTED_DIR:-meta-rules-dat-meta}"
SUBCONFIG_REPOSITORY="${SUBCONFIG_REPOSITORY:-${GITHUB_REPOSITORY:-AoEiuV020/SubConfig}}"
default_ref="${GITHUB_REF:-refs/heads/main}"
branch="${SUBCONFIG_BRANCH:-${default_ref#refs/heads/}}"

sed_in_place() {
    local expression="$1"
    local file="$2"
    if sed --version >/dev/null 2>&1; then
        sed -i "$expression" "$file"
    else
        sed -i '' "$expression" "$file"
    fi
}

replace_url() {
    local from="$1"
    local to="$2"
    local escaped_from
    escaped_from=$(printf '%s' "$from" | sed 's/\//\\\//g')
    for file in "$SUBCONFIG_DIR"/*.*; do
        sed_in_place "s/$escaped_from/$to/g" "$file"
    done
}

cache_repository() {
    local name="$1"
    local archive="$2"
    local archive_url="$3"
    local target_dir="$4"
    local extracted_dir="$5"
    local extract_parent

    echo "下载${name}规则仓库"
    curl -s -L -o "$archive" "$archive_url"
    extract_parent=$(dirname "$target_dir")
    rm -rf "$extract_parent/$extracted_dir"
    tar -zxf "$archive" -C "$extract_parent"
    rm -rf "$target_dir"
    mv "$extract_parent/$extracted_dir" "$target_dir"
}

cache_repository "ACL4SSR" "$ACL4SSR_ARCHIVE" "$ACL4SSR_ARCHIVE_URL" "$ACL4SSR_DIR" "$ACL4SSR_EXTRACTED_DIR"
cache_repository "meta-rules-dat" "$META_RULES_DAT_ARCHIVE" "$META_RULES_DAT_ARCHIVE_URL" "$META_RULES_DAT_DIR" "$META_RULES_DAT_EXTRACTED_DIR"

echo 替换配置文件, 包含以上仓库的地址，改成本地地址以加速，
replace_url "https://github.com/$SUBCONFIG_REPOSITORY/raw/$branch" "$SUBCONVERTER_BASE_PATH"
replace_url "https://github.com/ACL4SSR/ACL4SSR/raw/master" "$ACL4SSR_BASE_PATH"
replace_url "https://github.com/MetaCubeX/meta-rules-dat/raw/refs/heads/meta" "$META_RULES_DAT_BASE_PATH"
