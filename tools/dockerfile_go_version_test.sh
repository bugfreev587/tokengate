#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DOCKERFILE="$ROOT_DIR/Dockerfile"
GO_MOD="$ROOT_DIR/backend/go.mod"

fail() {
  printf 'FAIL %s\n' "$1" >&2
  exit 1
}

docker_go_version="$(
  awk '
    /^ARG GOLANG_IMAGE=golang:/ {
      value=$0
      sub(/^ARG GOLANG_IMAGE=golang:/, "", value)
      sub(/-.*/, "", value)
      print value
      exit
    }
  ' "$DOCKERFILE"
)"

go_mod_version="$(
  awk '
    /^go / {
      print $2
      exit
    }
  ' "$GO_MOD"
)"

if [[ -z "$docker_go_version" ]]; then
  fail "Dockerfile GOLANG_IMAGE version was not found"
fi

if [[ -z "$go_mod_version" ]]; then
  fail "backend/go.mod go version was not found"
fi

if [[ "$docker_go_version" != "$go_mod_version" ]]; then
  fail "Dockerfile Go image version ($docker_go_version) must match backend/go.mod ($go_mod_version)"
fi

printf 'PASS dockerfile_go_version_test go=%s\n' "$go_mod_version"
