#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${TOKENGATE_BASE_URL:-}"
API_KEY="${TOKENGATE_API_KEY:-}"
REQUIRE_CLAUDE="${TOKENGATE_REQUIRE_CLAUDE_MODELS:-1}"
REQUIRE_OPENAI="${TOKENGATE_REQUIRE_OPENAI_MODELS:-0}"

if [[ -z "$BASE_URL" || -z "$API_KEY" ]]; then
  cat >&2 <<'USAGE'
Usage:
  TOKENGATE_BASE_URL="https://your-backend-domain" \
  TOKENGATE_API_KEY="sk-..." \
  tools/tokengate_model_visibility.sh

Optional:
  TOKENGATE_REQUIRE_CLAUDE_MODELS=1|0
  TOKENGATE_REQUIRE_OPENAI_MODELS=1|0
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

out="$tmp_dir/models.json"
err="$tmp_dir/models.err"
status="$(curl -sS -o "$out" -w '%{http_code}' \
  --connect-timeout 20 \
  --max-time 60 \
  "$BASE_URL/v1/models" \
  -H "Authorization: Bearer $API_KEY" 2>"$err" || true)"

if [[ "$status" == "000" && -s "$err" ]]; then
  sed -n '1,3p' "$err" >&2
fi

if [[ "$status" -lt 200 || "$status" -ge 300 ]]; then
  echo "FAIL models HTTP $status" >&2
  sed -n '1,40p' "$out" >&2 || true
  exit 1
fi

echo "PASS models HTTP $status"

count_matches() {
  local pattern="$1"
  local flags="${2:-}"
  if [[ "$flags" == "extended" ]]; then
    { grep -Eo "$pattern" "$out" 2>/dev/null || true; } | wc -l | tr -d ' '
  else
    { grep -o "$pattern" "$out" 2>/dev/null || true; } | wc -l | tr -d ' '
  fi
}

model_count="$(count_matches '"id":"[^"]*"')"
claude_count="$(count_matches '"id":"claude[^"]*"')"
openai_count="$(count_matches '"id":"(gpt|o[0-9]|text-|dall-e|computer-use)[^"]*"' extended)"

printf 'Models visible: %s total, %s Claude, %s OpenAI-compatible\n' "$model_count" "$claude_count" "$openai_count"
grep -o '"id":"[^"]*"' "$out" | sed 's/"id":"//; s/"$//' | sed -n '1,40p'

failures=0
case "$REQUIRE_CLAUDE" in
  1|true|yes)
    if [[ "$claude_count" -gt 0 ]]; then
      echo "PASS Claude models visible"
    else
      echo "FAIL Claude models are required but none are visible" >&2
      failures=$((failures + 1))
    fi
    ;;
esac

case "$REQUIRE_OPENAI" in
  1|true|yes)
    if [[ "$openai_count" -gt 0 ]]; then
      echo "PASS OpenAI-compatible models visible"
    else
      echo "FAIL OpenAI-compatible models are required but none are visible" >&2
      failures=$((failures + 1))
    fi
    ;;
esac

if [[ "$failures" -gt 0 ]]; then
  exit 1
fi
