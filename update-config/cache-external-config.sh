#!/usr/bin/env bash
set -euo pipefail

SUBCONVERTER_DIR="${SUBCONVERTER_DIR:-subconverter}"
SUBCONVERTER_BASE_PATH="${SUBCONVERTER_BASE_PATH:-_SubConfig}"
ACL4SSR_BASE_PATH="${ACL4SSR_BASE_PATH:-_ACL4SSR}"
SUBCONFIG_DIR="${SUBCONFIG_DIR:-$SUBCONVERTER_DIR/$SUBCONVERTER_BASE_PATH}"
ACL4SSR_DIR="${ACL4SSR_DIR:-$SUBCONVERTER_DIR/$ACL4SSR_BASE_PATH}"
ACL4SSR_ARCHIVE="${ACL4SSR_ARCHIVE:-ACL4SSR.tar.gz}"
ACL4SSR_ARCHIVE_URL="${ACL4SSR_ARCHIVE_URL:-https://github.com/ACL4SSR/ACL4SSR/archive/refs/heads/master.tar.gz}"
ACL4SSR_EXTRACTED_DIR="${ACL4SSR_EXTRACTED_DIR:-ACL4SSR-master}"
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

echo 下载ACL4SSR，用的比较多的一个规则仓库，
curl -s -L -o "$ACL4SSR_ARCHIVE" "$ACL4SSR_ARCHIVE_URL"
acl_extract_parent=$(dirname "$ACL4SSR_DIR")
rm -rf "$acl_extract_parent/$ACL4SSR_EXTRACTED_DIR"
tar -zxf "$ACL4SSR_ARCHIVE" -C "$acl_extract_parent"
rm -rf "$ACL4SSR_DIR"
mv "$acl_extract_parent/$ACL4SSR_EXTRACTED_DIR" "$ACL4SSR_DIR"

echo 替换配置文件, 包含以上仓库的地址，改成本地地址以加速，
replace_url "https://github.com/$SUBCONFIG_REPOSITORY/raw/$branch" "$SUBCONVERTER_BASE_PATH"
replace_url "https://github.com/ACL4SSR/ACL4SSR/raw/master" "$ACL4SSR_BASE_PATH"
