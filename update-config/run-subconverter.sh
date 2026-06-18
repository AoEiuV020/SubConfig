#!/usr/bin/env bash
set -euo pipefail

RELEASE_FILE="${SUBCONVERTER_RELEASE_FILE:-release}"
RELEASE_API="${SUBCONVERTER_RELEASE_API:-https://api.github.com/repos/MetaCubeX/subconverter/releases/latest}"
ASSET_NAME="${SUBCONVERTER_ASSET_NAME:-subconverter_linux64.tar.gz}"
ASSET_FILE="${SUBCONVERTER_ASSET_FILE:-$ASSET_NAME}"
SUBCONVERTER_DIR="${SUBCONVERTER_DIR:-subconverter}"
SUBCONVERTER_LOG_FILE="${SUBCONVERTER_LOG_FILE:-../subconverter.log}"
SUBCONVERTER_PID_FILE="${SUBCONVERTER_PID_FILE:-../subconverter.pid}"
SUBCONVERTER_HEALTH_URL="${SUBCONVERTER_HEALTH_URL:-http://127.0.0.1:25500/version}"
SUBCONVERTER_BASE_PATH="${SUBCONVERTER_BASE_PATH:-_SubConfig}"

sed_in_place() {
    local expression="$1"
    local file="$2"
    if sed --version >/dev/null 2>&1; then
        sed -i "$expression" "$file"
    else
        sed -i '' "$expression" "$file"
    fi
}

curl_headers=()
if [[ -n "${GITHUB_TOKEN:-}" ]]; then
    curl_headers=(-H "Authorization: Bearer $GITHUB_TOKEN")
fi

echo 下载subconverter,
code=$(curl -s -L -o "$RELEASE_FILE" -w '%{http_code}' "${curl_headers[@]}" "$RELEASE_API")
if [[ "$code" != 200 ]]; then
    echo api请求异常，
    cat "$RELEASE_FILE"
    exit 3
fi

download_url=$(jq -r --arg name "$ASSET_NAME" '.assets[] | select(.name == $name).browser_download_url' "$RELEASE_FILE")
if [[ -z "$download_url" || "$download_url" == "null" ]]; then
    echo 未找到subconverter发布包: "$ASSET_NAME"
    exit 3
fi
curl -s -L -o "$ASSET_FILE" "$download_url"
rm -rf "$SUBCONVERTER_DIR"
mkdir -p "$(dirname "$SUBCONVERTER_DIR")"
tar -zxf "$ASSET_FILE" -C "$(dirname "$SUBCONVERTER_DIR")"

cd "$SUBCONVERTER_DIR"
echo 更改base_path以便支持缓存base配置文件，
mv pref.example.ini pref.ini
mv pref.example.toml pref.toml
mv pref.example.yml pref.yml
sed_in_place "s/^base_path=.*/base_path=$SUBCONVERTER_BASE_PATH/" pref.ini
sed_in_place "s/^base_path = \".*\"/base_path = \"$SUBCONVERTER_BASE_PATH\"/" pref.toml
sed_in_place "s/base_path: .*/base_path: $SUBCONVERTER_BASE_PATH/" pref.yml

echo 运行subconverter
./subconverter >"$SUBCONVERTER_LOG_FILE" 2>&1 &
echo "$!" >"$SUBCONVERTER_PID_FILE"

for _ in $(seq 1 30); do
    if curl -s -o /dev/null "$SUBCONVERTER_HEALTH_URL"; then
        exit 0
    fi
    sleep 1
done

echo subconverter启动超时，
cat "$SUBCONVERTER_LOG_FILE"
exit 5
