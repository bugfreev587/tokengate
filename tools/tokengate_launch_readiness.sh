#!/usr/bin/env bash
set -euo pipefail

FRONTEND_URL="${TOKENGATE_FRONTEND_URL:-}"
BACKEND_URL="${TOKENGATE_BACKEND_URL:-}"
API_KEY="${TOKENGATE_API_KEY:-}"
RUN_API_SMOKE="${TOKENGATE_RUN_API_SMOKE:-auto}"

failures=0
warnings=0

usage() {
  cat <<'EOF'
Usage:
  TOKENGATE_FRONTEND_URL="https://your-frontend-domain" \
  TOKENGATE_BACKEND_URL="https://your-backend-domain" \
  tools/tokengate_launch_readiness.sh

Optional:
  TOKENGATE_API_KEY="sk-..."
  TOKENGATE_RUN_API_SMOKE=auto|1|0
  TOKENGATE_PUBLIC_ROUTES="/home /docs /pricing /support /login"

Checks:
  - public frontend SPA routes survive refresh
  - backend /api/v1/settings/public is reachable
  - CORS preflight allows the frontend origin
  - optional Claude/OpenAI gateway smoke via tools/tokengate_smoke_test.sh
EOF
}

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

require_url() {
  local name="$1"
  local value="$2"
  if [[ -z "$value" ]]; then
    fail "$name is required"
    return 1
  fi
  if [[ "$value" != http://* && "$value" != https://* ]]; then
    fail "$name must start with http:// or https://"
    return 1
  fi
}

http_status() {
  local url="$1"
  local out="$2"
  local err="${out}.err"
  local status
  status="$(curl -sS -o "$out" -w '%{http_code}' --connect-timeout 20 --max-time 60 "$url" 2>"$err" || true)"
  if [[ "$status" == "000" && -s "$err" ]]; then
    sed -n '1,3p' "$err" >&2
  fi
  printf '%s' "$status"
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

require_url "TOKENGATE_FRONTEND_URL" "$FRONTEND_URL" || true
require_url "TOKENGATE_BACKEND_URL" "$BACKEND_URL" || true

if [[ "$failures" -gt 0 ]]; then
  usage >&2
  exit 1
fi

FRONTEND_URL="${FRONTEND_URL%/}"
BACKEND_URL="${BACKEND_URL%/}"
if [[ "$BACKEND_URL" == */api/v1 ]]; then
  warn "TOKENGATE_BACKEND_URL ended with /api/v1; stripping it for readiness checks"
  BACKEND_URL="${BACKEND_URL%/api/v1}"
fi

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

printf '\nFrontend route refresh checks\n'
routes="${TOKENGATE_PUBLIC_ROUTES:-/home /docs /pricing /support /login}"
for route in $routes; do
  out="$tmp_dir/frontend_${route//\//_}.html"
  status="$(http_status "$FRONTEND_URL$route" "$out")"
  if [[ "$status" -ge 200 && "$status" -lt 300 ]] && grep -q '<div id="app"' "$out"; then
    pass "frontend$route HTTP $status"
  else
    fail "frontend$route HTTP $status did not return the SPA shell"
    if [[ -f "$out" ]]; then
      sed -n '1,12p' "$out" >&2 || true
    fi
  fi
done

printf '\nBackend public settings check\n'
settings_out="$tmp_dir/public_settings.json"
settings_status="$(http_status "$BACKEND_URL/api/v1/settings/public" "$settings_out")"
if [[ "$settings_status" -ge 200 && "$settings_status" -lt 300 ]]; then
  pass "backend public settings HTTP $settings_status"
  if grep -q '"site_name"' "$settings_out"; then
    pass "public settings includes site_name"
  else
    warn "public settings response did not visibly include site_name"
  fi
else
  fail "backend public settings HTTP $settings_status"
  if [[ -f "$settings_out" ]]; then
    sed -n '1,20p' "$settings_out" >&2 || true
  fi
fi

printf '\nCORS preflight check\n'
cors_out="$tmp_dir/cors_headers.txt"
cors_err="$tmp_dir/cors.err"
cors_status="$(
  curl -sS -D "$cors_out" -o /dev/null -w '%{http_code}' \
    --connect-timeout 20 \
    --max-time 60 \
    -X OPTIONS "$BACKEND_URL/api/v1/settings/public" \
    -H "Origin: $FRONTEND_URL" \
    -H "Access-Control-Request-Method: GET" \
    -H "Access-Control-Request-Headers: authorization,content-type" 2>"$cors_err" || true
)"
if [[ "$cors_status" == "000" && -s "$cors_err" ]]; then
  sed -n '1,3p' "$cors_err" >&2
fi
if [[ "$cors_status" -ge 200 && "$cors_status" -lt 300 ]]; then
  pass "CORS preflight HTTP $cors_status"
else
  fail "CORS preflight HTTP $cors_status"
fi

if grep -iq "^access-control-allow-origin: $FRONTEND_URL" "$cors_out"; then
  pass "CORS allows frontend origin"
else
  fail "CORS allow-origin did not match $FRONTEND_URL"
  sed -n '1,30p' "$cors_out" >&2 || true
fi

printf '\nGateway smoke check\n'
should_run_smoke=0
case "$RUN_API_SMOKE" in
  1|true|yes)
    should_run_smoke=1
    ;;
  0|false|no)
    should_run_smoke=0
    ;;
  auto)
    if [[ -n "$API_KEY" ]]; then
      should_run_smoke=1
    else
      warn "Skipping gateway smoke because TOKENGATE_API_KEY is not set"
    fi
    ;;
  *)
    fail "TOKENGATE_RUN_API_SMOKE must be auto, 1, or 0"
    ;;
esac

if [[ "$should_run_smoke" == "1" ]]; then
  if [[ -z "$API_KEY" ]]; then
    fail "TOKENGATE_API_KEY is required when TOKENGATE_RUN_API_SMOKE=1"
  else
    TOKENGATE_BASE_URL="$BACKEND_URL" TOKENGATE_API_KEY="$API_KEY" bash tools/tokengate_smoke_test.sh || failures=$((failures + 1))
  fi
fi

printf '\nSummary: %d failure(s), %d warning(s)\n' "$failures" "$warnings"
if [[ "$failures" -gt 0 ]]; then
  exit 1
fi
