#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
project_dir="$(cd -- "$script_dir/.." && pwd)"
repo_root="$(cd -- "$project_dir/.." && pwd)"

image="${IMAGE:-aoeiuv020/subconfig-config-depot}"
platforms="${PLATFORMS:-linux/amd64,linux/arm64}"
builder="${BUILDER:-config-depot-publisher}"
branch="${BRANCH:-$(git -C "$repo_root" rev-parse --abbrev-ref HEAD)}"
sha="${SHA:-$(git -C "$repo_root" rev-parse --short HEAD)}"

if [[ "$branch" == "HEAD" ]]; then
  branch="detached"
fi

safe_branch="$(printf '%s' "$branch" | tr '/:@ ' '----')"

if [[ "$platforms" == *,* ]]; then
  inspect_output="$(docker buildx inspect 2>/dev/null || true)"
  current_driver="$(printf '%s\n' "$inspect_output" | awk '/^Driver:/ { print $2; exit }')"
  if [[ "$current_driver" == "docker" ]]; then
    if ! docker buildx inspect "$builder" >/dev/null 2>&1 && docker buildx inspect multiarch-builder >/dev/null 2>&1; then
      builder="multiarch-builder"
    fi
    if ! docker buildx inspect "$builder" >/dev/null 2>&1; then
      docker buildx create --name "$builder" --driver docker-container --bootstrap >/dev/null
    fi
    docker buildx use "$builder" >/dev/null
  fi
fi

args=(
  buildx
  build
  --tag "$image:latest"
  --tag "$image:$safe_branch"
  --tag "$image:$sha"
  --push
  "$project_dir"
)

if [[ -n "$platforms" ]]; then
  args=(buildx build --platform "$platforms" "${args[@]:2}")
fi

printf '发布镜像：%s\n' "$image"
if [[ -n "$platforms" ]]; then
  printf '平台：%s\n' "$platforms"
else
  printf '平台：当前 Docker builder 默认平台\n'
fi
printf '标签：latest, %s, %s\n' "$safe_branch" "$sha"

docker "${args[@]}"
