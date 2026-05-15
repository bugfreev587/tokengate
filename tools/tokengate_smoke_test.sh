#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${TOKENGATE_BASE_URL:-}"
API_KEY="${TOKENGATE_API_KEY:-}"
CLAUDE_MODEL="${TOKENGATE_CLAUDE_MODEL:-claude-haiku-4-5-20251001}"
OPENAI_MODEL="${TOKENGATE_OPENAI_MODEL:-gpt-4.1-mini}"
RUN_CLAUDE="${TOKENGATE_RUN_CLAUDE:-1}"
RUN_OPENAI="${TOKENGATE_RUN_OPENAI:-1}"
RUN_PREFLIGHT="${TOKENGATE_RUN_PREFLIGHT:-1}"

if [[ -z "$BASE_URL" || -z "$API_KEY" ]]; then
  cat >&2 <<'USAGE'
Usage:
  TOKENGATE_BASE_URL="https://your-backend-domain" \
  TOKENGATE_API_KEY="sk-..." \
  bash tools/tokengate_smoke_test.sh

Optional:
  TOKENGATE_CLAUDE_MODEL="claude-haiku-4-5-20251001"
  TOKENGATE_OPENAI_MODEL="gpt-4.1-mini"
  TOKENGATE_RUN_CLAUDE=0
  TOKENGATE_RUN_OPENAI=0
USAGE
  exit 2
fi

BASE_URL="${BASE_URL%/}"
if [[ "$BASE_URL" == */api/v1 ]]; then
  echo "WARN TOKENGATE_BASE_URL ended with /api/v1; stripping it for gateway endpoints." >&2
  BASE_URL="${BASE_URL%/api/v1}"
fi

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

diagnose_status() {
  local status="$1"
  case "$status" in
    401)
      echo "Hint: API key is missing, invalid, disabled, or expired." >&2
      ;;
    403)
      echo "Hint: API key is valid but not allowed for this group, plan, balance, or route." >&2
      ;;
    404)
      echo "Hint: endpoint not found. Check that TOKENGATE_BASE_URL is the Railway backend origin, not the Vercel frontend URL." >&2
      ;;
    405)
      echo "Hint: method not allowed. This often means requests are hitting the frontend/static host instead of the backend API." >&2
      ;;
    429)
      echo "Hint: rate limit, quota, or balance limit reached." >&2
      ;;
    5*)
      echo "Hint: backend or upstream provider failed. Check Railway logs and provider account health." >&2
      ;;
  esac
}

preflight() {
  local out="$tmp_dir/public_settings.json"
  local status

  status="$(curl -sS -o "$out" -w '%{http_code}' \
    --connect-timeout 20 \
    --max-time 60 \
    "$BASE_URL/api/v1/settings/public")"

  if [[ "$status" -ge 200 && "$status" -lt 300 ]]; then
    echo "PASS public_settings HTTP $status"
    return 0
  fi

  echo "WARN public_settings HTTP $status" >&2
  sed -n '1,20p' "$out" >&2
  diagnose_status "$status"
}

request() {
  local name="$1"
  local url="$2"
  local body="$3"
  local extra_header="${4:-}"
  local out="$tmp_dir/${name}.json"
  local status

  if [[ -n "$extra_header" ]]; then
    status="$(curl -sS -o "$out" -w '%{http_code}' \
      --connect-timeout 20 \
      --max-time 120 \
      -X POST "$url" \
      -H "Authorization: Bearer $API_KEY" \
      -H "Content-Type: application/json" \
      -H "$extra_header" \
      -d "$body")"
  else
    status="$(curl -sS -o "$out" -w '%{http_code}' \
      --connect-timeout 20 \
      --max-time 120 \
      -X POST "$url" \
      -H "Authorization: Bearer $API_KEY" \
      -H "Content-Type: application/json" \
      -d "$body")"
  fi

  if [[ "$status" -lt 200 || "$status" -ge 300 ]]; then
    echo "FAIL $name HTTP $status" >&2
    sed -n '1,40p' "$out" >&2
    diagnose_status "$status"
    return 1
  fi

  echo "PASS $name HTTP $status"
  sed -n '1,8p' "$out"
}

if [[ "$RUN_PREFLIGHT" == "1" ]]; then
  preflight
fi

if [[ "$RUN_CLAUDE" == "1" ]]; then
  request "claude_messages" "$BASE_URL/v1/messages" "{
    \"model\": \"$CLAUDE_MODEL\",
    \"max_tokens\": 32,
    \"messages\": [
      {\"role\": \"user\", \"content\": \"Reply with exactly: hello\"}
    ]
  }" "anthropic-version: 2023-06-01"
fi

if [[ "$RUN_OPENAI" == "1" ]]; then
  request "openai_chat_completions" "$BASE_URL/v1/chat/completions" "{
    \"model\": \"$OPENAI_MODEL\",
    \"messages\": [
      {\"role\": \"user\", \"content\": \"Reply with exactly: hello\"}
    ]
  }"
fi

echo "Smoke test completed. Refresh TokenGate Usage and Dashboard to confirm metering."
