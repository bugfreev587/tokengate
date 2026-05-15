#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${TOKENGATE_BASE_URL:-}"
API_KEY="${TOKENGATE_API_KEY:-}"
CLAUDE_MODEL="${TOKENGATE_CLAUDE_MODEL:-claude-haiku-4-5-20251001}"
OPENAI_MODEL="${TOKENGATE_OPENAI_MODEL:-gpt-4.1-mini}"
RUN_CLAUDE="${TOKENGATE_RUN_CLAUDE:-1}"
RUN_OPENAI="${TOKENGATE_RUN_OPENAI:-1}"

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
tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

request() {
  local name="$1"
  local url="$2"
  local body="$3"
  local extra_header="${4:-}"
  local out="$tmp_dir/${name}.json"
  local status

  if [[ -n "$extra_header" ]]; then
    status="$(curl -sS -o "$out" -w '%{http_code}' \
      -X POST "$url" \
      -H "Authorization: Bearer $API_KEY" \
      -H "Content-Type: application/json" \
      -H "$extra_header" \
      -d "$body")"
  else
    status="$(curl -sS -o "$out" -w '%{http_code}' \
      -X POST "$url" \
      -H "Authorization: Bearer $API_KEY" \
      -H "Content-Type: application/json" \
      -d "$body")"
  fi

  if [[ "$status" -lt 200 || "$status" -ge 300 ]]; then
    echo "FAIL $name HTTP $status" >&2
    sed -n '1,40p' "$out" >&2
    return 1
  fi

  echo "PASS $name HTTP $status"
  sed -n '1,8p' "$out"
}

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
