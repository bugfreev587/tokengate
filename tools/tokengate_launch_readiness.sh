#!/usr/bin/env bash
set -euo pipefail

FRONTEND_URL="${TOKENGATE_FRONTEND_URL:-}"
BACKEND_URL="${TOKENGATE_BACKEND_URL:-}"
API_KEY="${TOKENGATE_API_KEY:-}"
RUN_API_SMOKE="${TOKENGATE_RUN_API_SMOKE:-auto}"
LAUNCH_PROFILE="${TOKENGATE_LAUNCH_PROFILE:-private}"
SIGNUP_MODE="${TOKENGATE_SIGNUP_MODE:-auto}"
REQUIRE_PAYMENT="${TOKENGATE_REQUIRE_PAYMENT:-auto}"
EXPECTED_CONTACT_INFO="${TOKENGATE_EXPECTED_CONTACT_INFO:-}"

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
  TOKENGATE_LAUNCH_PROFILE=private|public
  TOKENGATE_SIGNUP_MODE=auto|invite|self_serve
  TOKENGATE_REQUIRE_PAYMENT=auto|1|0
  TOKENGATE_EXPECTED_CONTACT_INFO="support@example.com"
  TOKENGATE_FRONTEND_ROUTES="/home /docs /pricing /support /login /dashboard /usage /admin/accounts /admin/launch-readiness"

Checks:
  - frontend SPA routes survive refresh
  - backend /api/v1/settings/public is reachable
  - CORS preflight allows the frontend origin
  - public settings match the intended private/public launch profile
  - optional P0 gateway compatibility smoke via tools/tokengate_p0_compatibility_suite.sh
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

json_string_value() {
  local key="$1"
  local file="$2"
  sed -n "s/.*\"$key\":\"\\([^\"]*\\)\".*/\\1/p" "$file" | head -1
}

json_bool_value() {
  local key="$1"
  local file="$2"
  sed -n "s/.*\"$key\":\\(true\\|false\\).*/\\1/p" "$file" | head -1
}

profile_fail_or_warn() {
  local message="$1"
  if [[ "$LAUNCH_PROFILE" == "public" ]]; then
    fail "$message"
  else
    warn "$message"
  fi
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

require_url "TOKENGATE_FRONTEND_URL" "$FRONTEND_URL" || true
require_url "TOKENGATE_BACKEND_URL" "$BACKEND_URL" || true

case "$LAUNCH_PROFILE" in
  private|public) ;;
  *)
    fail "TOKENGATE_LAUNCH_PROFILE must be private or public"
    ;;
esac

case "$SIGNUP_MODE" in
  auto|invite|self_serve) ;;
  *)
    fail "TOKENGATE_SIGNUP_MODE must be auto, invite, or self_serve"
    ;;
esac

case "$REQUIRE_PAYMENT" in
  auto|1|true|yes|0|false|no) ;;
  *)
    fail "TOKENGATE_REQUIRE_PAYMENT must be auto, 1, or 0"
    ;;
esac

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
routes="${TOKENGATE_FRONTEND_ROUTES:-/home /docs /pricing /support /login /dashboard /usage /admin/accounts /admin/launch-readiness}"
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

  printf '\nPublic settings launch gate checks\n'
  site_name="$(json_string_value "site_name" "$settings_out")"
  contact_info="$(json_string_value "contact_info" "$settings_out")"
  registration_enabled="$(json_bool_value "registration_enabled" "$settings_out")"
  password_reset_enabled="$(json_bool_value "password_reset_enabled" "$settings_out")"
  payment_enabled="$(json_bool_value "payment_enabled" "$settings_out")"
  email_verify_enabled="$(json_bool_value "email_verify_enabled" "$settings_out")"

  if [[ -n "$site_name" ]]; then
    pass "site_name is configured: $site_name"
  else
    fail "site_name is empty"
  fi

  if [[ -n "$contact_info" ]]; then
    pass "contact_info is configured"
    if [[ -n "$EXPECTED_CONTACT_INFO" ]]; then
      if [[ "$contact_info" == "$EXPECTED_CONTACT_INFO" ]]; then
        pass "contact_info matches expected value"
      else
        profile_fail_or_warn "contact_info is '$contact_info', expected '$EXPECTED_CONTACT_INFO'"
      fi
    fi
  else
    if [[ -n "$EXPECTED_CONTACT_INFO" ]]; then
      profile_fail_or_warn "contact_info is empty; expected '$EXPECTED_CONTACT_INFO'"
    else
      profile_fail_or_warn "contact_info is empty; /support has no real support channel"
    fi
  fi

  if [[ "$LAUNCH_PROFILE" == "public" ]]; then
    if [[ "$password_reset_enabled" == "true" ]]; then
      pass "password reset is enabled"
    else
      fail "password_reset_enabled is false; public users cannot recover accounts"
    fi

    effective_signup_mode="$SIGNUP_MODE"
    if [[ "$effective_signup_mode" == "auto" ]]; then
      effective_signup_mode="self_serve"
    fi
    if [[ "$effective_signup_mode" == "self_serve" ]]; then
      if [[ "$registration_enabled" == "true" ]]; then
        pass "self-serve registration is enabled"
      else
        fail "registration_enabled is false but TOKENGATE_SIGNUP_MODE=self_serve"
      fi
      if [[ "$email_verify_enabled" == "true" ]]; then
        pass "email verification is enabled"
      else
        warn "email_verify_enabled is false for self-serve signup"
      fi
    else
      pass "signup mode is invite; public self-serve registration is not required"
    fi
  else
    if [[ "$registration_enabled" == "true" ]]; then
      warn "registration_enabled is true during private launch profile"
    else
      pass "registration is closed for private launch profile"
    fi
    if [[ "$password_reset_enabled" == "true" ]]; then
      pass "password reset is enabled"
    else
      warn "password_reset_enabled is false; acceptable for private beta but must be verified before public launch"
    fi
  fi

  effective_require_payment="$REQUIRE_PAYMENT"
  if [[ "$effective_require_payment" == "auto" ]]; then
    if [[ "$LAUNCH_PROFILE" == "public" ]]; then
      effective_require_payment="1"
    else
      effective_require_payment="0"
    fi
  fi
  case "$effective_require_payment" in
    1|true|yes)
      if [[ "$payment_enabled" == "true" ]]; then
        pass "payment is enabled"
      else
        fail "payment_enabled is false but payment is required"
      fi
      ;;
    *)
      if [[ "$payment_enabled" == "true" ]]; then
        pass "payment is enabled"
      else
        pass "payment is disabled and not required for this launch profile"
      fi
      ;;
  esac
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
    TOKENGATE_BASE_URL="$BACKEND_URL" TOKENGATE_API_KEY="$API_KEY" tools/tokengate_p0_compatibility_suite.sh || failures=$((failures + 1))
  fi
fi

printf '\nSummary: %d failure(s), %d warning(s)\n' "$failures" "$warnings"
if [[ "$failures" -gt 0 ]]; then
  exit 1
fi
