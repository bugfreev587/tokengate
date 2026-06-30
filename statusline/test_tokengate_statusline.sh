#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SCRIPT="$ROOT/statusline/tokengate-statusline.sh"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

strip_ansi() {
  sed -E $'s/\x1B\\[[0-9;]*[mK]//g'
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"
  if [[ "$haystack" != *"$needle"* ]]; then
    printf 'FAIL: %s\nExpected to contain: %s\nActual: %s\n' "$label" "$needle" "$haystack" >&2
    exit 1
  fi
}

write_input() {
  cat > "$TMP/input.json" <<JSON
{
  "model": { "display_name": "Claude Sonnet" },
  "context_window": {
    "context_window_size": 200000,
    "current_usage": {
      "input_tokens": 40000,
      "cache_creation_input_tokens": 1000,
      "cache_read_input_tokens": 1000
    }
  },
  "cost": 0.1234,
  "cwd": "$ROOT",
  "version": "2.1.0"
}
JSON
}

write_statusline_cache() {
  mkdir -p "$TMP/cache"
  cat > "$TMP/cache/tokengate-statusline-cache.json" <<'JSON'
{
  "ok": true,
  "billing_mode": "API_USAGE",
  "cost": { "today": "1.28" },
  "budgets": {
    "monthly": { "used": "38", "limit": "100", "percent": 38 },
    "daily": { "used": "4", "limit": "20", "percent": 20 }
  }
}
JSON
}

write_input
write_statusline_cache

default_output="$(
  TOKENGATE_STATUSLINE_CACHE_DIR="$TMP/cache" \
  TOKENGATE_API_KEY="tg_test" \
  ANTHROPIC_BASE_URL="https://api.tokengate.to" \
  COLUMNS=240 \
  sh "$SCRIPT" < "$TMP/input.json" | strip_ansi
)"
assert_contains "$default_output" "Claude Sonnet" "default mode includes model"
assert_contains "$default_output" "token-gate@" "default mode includes project and branch"
assert_contains "$default_output" "ctx 42k/200k 21%" "default mode includes context"
assert_contains "$default_output" "\$1.28 today" "default mode includes today cost"
assert_contains "$default_output" "month" "default mode includes month budget"
assert_contains "$default_output" "\$38/\$100 38%" "default mode includes month budget amount"
assert_contains "$default_output" "day" "default mode includes day budget"
assert_contains "$default_output" "\$4/\$20 20%" "default mode includes day budget amount"

env_claude_output="$(
  TOKENGATE_STATUSLINE_MODE="claude" \
  TOKENGATE_STATUSLINE_CACHE_DIR="$TMP/cache" \
  TOKENGATE_API_KEY="tg_test" \
  ANTHROPIC_BASE_URL="https://api.tokengate.to" \
  COLUMNS=240 \
  sh "$SCRIPT" < "$TMP/input.json" | strip_ansi
)"
assert_contains "$env_claude_output" "21% used" "env mode selects Claude renderer"
assert_contains "$env_claude_output" "79% remain" "env mode keeps Claude remain segment"

arg_claude_output="$(
  TOKENGATE_STATUSLINE_MODE="tokengate" \
  TOKENGATE_STATUSLINE_CACHE_DIR="$TMP/cache" \
  TOKENGATE_API_KEY="tg_test" \
  ANTHROPIC_BASE_URL="https://api.tokengate.to" \
  COLUMNS=240 \
  sh "$SCRIPT" --mode claude < "$TMP/input.json" | strip_ansi
)"
assert_contains "$arg_claude_output" "21% used" "argument mode overrides env mode"

unavailable_output="$(
  TOKENGATE_STATUSLINE_CACHE_DIR="$TMP/empty-cache" \
  TOKENGATE_API_KEY="tg_test" \
  ANTHROPIC_BASE_URL="http://127.0.0.1:9" \
  COLUMNS=240 \
  sh "$SCRIPT" < "$TMP/input.json" | strip_ansi
)"
assert_contains "$unavailable_output" "TokenGate unavailable" "tokengate mode degrades"

empty_output="$(sh "$SCRIPT" </dev/null | strip_ansi)"
assert_contains "$empty_output" "Claude" "empty stdin fallback"

printf 'statusline tests passed\n'
