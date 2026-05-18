#!/usr/bin/env bash
# Lists all external container images referenced by this Promise.
# With a registry argument, shows the remapped images and offers to update all files.
#
# Usage:
#   list-images.sh                   # list original image references
#   list-images.sh NEW_REGISTRY      # remap registry and offer to update files
#
# Example:
#   list-images.sh my-registry.example.com
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

collect_images() {
  {
    # Standard Kubernetes 'image:' fields in manifests (handles both 'image:' and '- image:')
    find "$REPO_ROOT" -not -path "*/.github/*" \( -name "*.yaml" -o -name "*.yml" \) \
      -exec grep -hE '^\s+(- )?image:[[:space:]]+' {} \; \
      | sed -E 's/.*image:[[:space:]]+"?([^"[:space:]#]+)"?.*/\1/'

    # Zalando operator image config fields (docker_image, logical_backup_docker_image, connection_pooler_image)
    find "$REPO_ROOT" -not -path "*/.github/*" \( -name "*.yaml" -o -name "*.yml" \) \
      -exec grep -hE '^\s+(docker_image|logical_backup_docker_image|connection_pooler_image):[[:space:]]+' {} \; \
      | sed -E 's/.*:[[:space:]]+"?([^"[:space:]#]+)"?.*/\1/'

    # CRD schema default values that are image references (registry/path:tag format)
    find "$REPO_ROOT" -not -path "*/.github/*" \( -name "*.yaml" -o -name "*.yml" \) \
      -exec grep -hE '^\s+default:[[:space:]]+"[a-z][^"]*/[^"]+:[^"]+"' {} \; \
      | sed -E 's/.*default:[[:space:]]+"([^"]+)".*/\1/'

    # Dockerfile FROM references (skips shell variable placeholders like $BUILDPLATFORM)
    find "$REPO_ROOT" -not -path "*/.github/*" -name "Dockerfile" \
      -exec grep -hE '^FROM ' {} \; \
      | sed -E 's/FROM (--platform=[^ ]+ )?([^ ]+).*/\2/' \
      | grep -v '^\$'
  } | sort -u
}

# Replace the registry portion of an image reference.
# Images without an explicit registry (e.g. 'golang:1.26-alpine') are treated as
# unqualified and the new registry is simply prepended.
replace_registry() {
  local new_registry="$1"
  local image="$2"
  local first="${image%%/*}"

  if [[ "$image" == */* ]] && [[ "$first" == *"."* || "$first" == *":"* || "$first" == "localhost" ]]; then
    echo "${new_registry}/${image#*/}"
  else
    echo "${new_registry}/${image}"
  fi
}

# Escapes dots in a string so it can be used as a literal sed pattern.
escape_sed_pattern() {
  printf '%s' "$1" | sed 's/\./\\./g'
}

update_files() {
  local old_image="$1"
  local new_image="$2"
  local pattern
  pattern="$(escape_sed_pattern "$old_image")"

  while IFS= read -r file; do
    if grep -qF "$old_image" "$file" 2>/dev/null; then
      if [[ "$(uname)" == "Darwin" ]]; then
        sed -i '' "s|${pattern}|${new_image}|g" "$file"
      else
        sed -i "s|${pattern}|${new_image}|g" "$file"
      fi
    fi
  done < <(find "$REPO_ROOT" -not -path "*/.github/*" \
    \( -name "*.yaml" -o -name "*.yml" -o -name "Dockerfile" \))
}

images="$(collect_images)"

if [[ $# -eq 0 ]]; then
  echo "$images"
  exit 0
fi

new_registry="${1%/}"

declare -a old_images=()
declare -a new_images=()
while IFS= read -r image; do
  old_images+=("$image")
  new_images+=("$(replace_registry "$new_registry" "$image")")
done <<< "$images"

echo "Images needed to mirror:"
printf '  %s\n' "${old_images[@]}"

echo ""
echo "New paths to be updated on files:"
printf '  %s\n' "${new_images[@]}"

echo ""
read -rp "Update files? (y/N) " answer
if [[ "${answer,,}" != "y" ]]; then
  echo "Aborted."
  exit 0
fi

for i in "${!old_images[@]}"; do
  update_files "${old_images[$i]}" "${new_images[$i]}"
done

echo "Done."
