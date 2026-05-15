#!/usr/bin/env bash
set -euo pipefail

MODE="${1:-all}"
failures=0
warnings=0

pass() {
  printf 'PASS %s\n' "$1"
}

warn() {
  warnings=$((warnings + 1))
  printf 'WARN %s\n' "$1" >&2
}

fail() {
  failures=$((failures + 1))
  printf 'FAIL %s\n' "$1" >&2
}

get_var() {
  local name="$1"
  printf '%s' "${!name-}"
}

require_var() {
  local name="$1"
  local value
  value="$(get_var "$name")"
  if [[ -z "$value" ]]; then
    fail "$name is required"
  else
    pass "$name is set"
  fi
}

require_url() {
  local name="$1"
  local value
  value="$(get_var "$name")"
  if [[ -z "$value" ]]; then
    fail "$name is required"
    return
  fi
  if [[ "$value" =~ ^https?:// || "$value" =~ ^postgres(ql)?:// || "$value" =~ ^redis(s)?:// ]]; then
    pass "$name has a URL-like value"
  else
    fail "$name must be a URL"
  fi
}

require_secret() {
  local name="$1"
  local value
  value="$(get_var "$name")"
  if [[ -z "$value" ]]; then
    fail "$name is required"
    return
  fi
  if [[ "${#value}" -lt 32 ]]; then
    fail "$name should be at least 32 characters"
  else
    pass "$name length looks safe"
  fi
}

check_no_placeholder() {
  local name="$1"
  local value
  value="$(get_var "$name")"
  if [[ "$value" == *"YOUR_"* || "$value" == *"CHANGE_ME"* || "$value" == *"example.com"* ]]; then
    fail "$name still looks like a placeholder"
  fi
}

check_railway() {
  printf '\nRailway backend environment\n'
  require_var "AUTO_SETUP"
  require_var "RUN_MODE"
  require_var "SERVER_PORT"
  require_secret "JWT_SECRET"
  require_secret "TOTP_ENCRYPTION_KEY"
  require_var "ADMIN_EMAIL"
  require_secret "ADMIN_PASSWORD"
  require_url "DATABASE_URL"
  require_url "REDIS_URL"
  require_url "FRONTEND_URL"
  require_var "CORS_ALLOWED_ORIGINS"

  for name in AUTO_SETUP RUN_MODE SERVER_PORT JWT_SECRET TOTP_ENCRYPTION_KEY ADMIN_EMAIL ADMIN_PASSWORD DATABASE_URL REDIS_URL FRONTEND_URL CORS_ALLOWED_ORIGINS; do
    check_no_placeholder "$name"
  done

  if [[ "$(get_var AUTO_SETUP)" != "true" ]]; then
    warn "AUTO_SETUP is not true; first deploys may require the setup wizard"
  fi
  if [[ "$(get_var RUN_MODE)" != "standard" ]]; then
    warn "RUN_MODE is not standard; public product behavior may differ"
  fi
  if [[ "$(get_var CORS_ALLOWED_ORIGINS)" == "*" ]]; then
    fail "CORS_ALLOWED_ORIGINS must not be '*' for production"
  fi
  if [[ "$(get_var CORS_ALLOWED_ORIGINS)" != https://* ]]; then
    warn "CORS_ALLOWED_ORIGINS should usually be an https frontend origin"
  fi
  if [[ "$(get_var FRONTEND_URL)" != https://* ]]; then
    warn "FRONTEND_URL should usually be an https frontend origin"
  fi
}

check_vercel() {
  printf '\nVercel frontend environment\n'
  require_url "VITE_API_BASE_URL"
  check_no_placeholder "VITE_API_BASE_URL"

  if [[ "$(get_var VITE_API_BASE_URL)" != */api/v1 ]]; then
    fail "VITE_API_BASE_URL should end with /api/v1"
  fi
  if [[ "$(get_var VITE_API_BASE_URL)" == *"vercel.app"* ]]; then
    fail "VITE_API_BASE_URL should point to the backend, not the Vercel frontend"
  fi
  if [[ -n "${VITE_BUILD_TARGET-}" && "$(get_var VITE_BUILD_TARGET)" != "standalone" ]]; then
    warn "VITE_BUILD_TARGET is set but not standalone"
  fi
}

case "$MODE" in
  all)
    check_railway
    check_vercel
    ;;
  railway)
    check_railway
    ;;
  vercel)
    check_vercel
    ;;
  *)
    printf 'Usage: %s [all|railway|vercel]\n' "$0" >&2
    exit 2
    ;;
esac

printf '\nSummary: %d failure(s), %d warning(s)\n' "$failures" "$warnings"
if [[ "$failures" -gt 0 ]]; then
  exit 1
fi
