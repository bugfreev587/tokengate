#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${TOKENGATE_BASE_URL:-}"
API_KEY="${TOKENGATE_API_KEY:-}"
OPENAI_MODEL="${TOKENGATE_OPENAI_MODEL:-gpt-5.4}"
RUN_MODELS="${TOKENGATE_RUN_P0_MODELS:-1}"
RUN_CHAT="${TOKENGATE_RUN_P0_CHAT:-1}"
RUN_CHAT_STREAM="${TOKENGATE_RUN_P0_CHAT_STREAM:-1}"
RUN_RESPONSES="${TOKENGATE_RUN_P0_RESPONSES:-1}"
RUN_SDK="${TOKENGATE_RUN_P0_SDK:-auto}"
REQUIRE_OPENAI_MODELS="${TOKENGATE_REQUIRE_OPENAI_MODELS:-1}"
REQUIRE_MODEL_VISIBLE="${TOKENGATE_REQUIRE_P0_MODEL_VISIBLE:-1}"
REQUIRE_STREAM_USAGE="${TOKENGATE_REQUIRE_STREAM_USAGE:-1}"

failures=0
warnings=0

usage() {
  cat <<'USAGE'
Usage:
  TOKENGATE_BASE_URL="https://your-backend-domain" \
  TOKENGATE_API_KEY="sk-..." \
  TOKENGATE_OPENAI_MODEL="gpt-5.4" \
  tools/tokengate_p0_compatibility_suite.sh

Optional:
  TOKENGATE_RUN_P0_MODELS=1|0
  TOKENGATE_RUN_P0_CHAT=1|0
  TOKENGATE_RUN_P0_CHAT_STREAM=1|0
  TOKENGATE_RUN_P0_RESPONSES=1|0
  TOKENGATE_RUN_P0_SDK=auto|1|0
  TOKENGATE_REQUIRE_OPENAI_MODELS=1|0
  TOKENGATE_REQUIRE_P0_MODEL_VISIBLE=1|0
  TOKENGATE_REQUIRE_STREAM_USAGE=1|0

Checks:
  - GET /v1/models exposes OpenAI-compatible models
  - POST /v1/chat/completions succeeds
  - POST /v1/chat/completions streaming emits SSE data and [DONE]
  - POST /v1/responses succeeds
  - optional OpenAI Node SDK chat smoke
USAGE
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

is_enabled() {
  case "${1:-}" in
    1|true|yes|on) return 0 ;;
    *) return 1 ;;
  esac
}

print_file_excerpt() {
  local file="$1"
  local lines="${2:-40}"
  if [[ -f "$file" ]]; then
    sed -n "1,${lines}p" "$file" >&2 || true
  fi
}

print_curl_error() {
  local err_file="$1"
  if [[ -s "$err_file" ]]; then
    sed -n '1,3p' "$err_file" >&2
  fi
}

diagnose_status() {
  local status="$1"
  local body_file="${2:-}"
  case "$status" in
    000)
      echo "Hint: curl could not connect. Check DNS, TLS, backend availability, and network access." >&2
      ;;
    401)
      echo "Hint: API key is missing, invalid, disabled, or expired." >&2
      ;;
    403)
      echo "Hint: API key is valid but blocked by group, subscription, balance, content policy, or route permission." >&2
      ;;
    404)
      if [[ -n "$body_file" ]] && grep -q '"message":"model:' "$body_file" 2>/dev/null; then
        echo "Hint: requested model is not routable. Check OpenAI account assignment and model whitelist." >&2
      else
        echo "Hint: endpoint not found. Check that TOKENGATE_BASE_URL is the backend origin." >&2
      fi
      ;;
    405)
      echo "Hint: method not allowed. Requests may be hitting the frontend/static host instead of the backend." >&2
      ;;
    429)
      echo "Hint: rate limit, quota, or balance limit reached." >&2
      ;;
    5*)
      echo "Hint: backend or upstream provider failed. Check runtime logs and provider account health." >&2
      ;;
  esac
}

normalize_base_url() {
  BASE_URL="${BASE_URL%/}"
  if [[ "$BASE_URL" == */api/v1 ]]; then
    warn "TOKENGATE_BASE_URL ended with /api/v1; stripping it for gateway endpoints"
    BASE_URL="${BASE_URL%/api/v1}"
  fi
  if [[ "$BASE_URL" == */v1 ]]; then
    warn "TOKENGATE_BASE_URL ended with /v1; stripping it before adding gateway paths"
    BASE_URL="${BASE_URL%/v1}"
  fi
}

http_get() {
  local name="$1"
  local url="$2"
  local out="$3"
  local err="$4"
  curl -sS -o "$out" -w '%{http_code}' \
    --connect-timeout 20 \
    --max-time 60 \
    "$url" \
    -H "Authorization: Bearer $API_KEY" 2>"$err" || true
}

post_json() {
  local url="$1"
  local body="$2"
  local out="$3"
  local err="$4"
  curl -sS -o "$out" -w '%{http_code}' \
    --connect-timeout 20 \
    --max-time 120 \
    -X POST "$url" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -d "$body" 2>"$err" || true
}

count_matches() {
  local file="$1"
  local pattern="$2"
  local flags="${3:-}"
  if [[ "$flags" == "extended" ]]; then
    { grep -Eo "$pattern" "$file" 2>/dev/null || true; } | wc -l | tr -d ' '
  else
    { grep -o "$pattern" "$file" 2>/dev/null || true; } | wc -l | tr -d ' '
  fi
}

check_models() {
  local out="$tmp_dir/models.json"
  local err="$tmp_dir/models.err"
  local status

  status="$(http_get "models" "$BASE_URL/v1/models" "$out" "$err")"
  if [[ "$status" -lt 200 || "$status" -ge 300 ]]; then
    fail "models HTTP $status"
    print_curl_error "$err"
    print_file_excerpt "$out" 40
    diagnose_status "$status" "$out"
    return
  fi

  pass "models HTTP $status"

  local model_count
  local openai_count
  model_count="$(count_matches "$out" '"id":"[^"]*"')"
  openai_count="$(count_matches "$out" '"id":"(gpt|o[0-9]|text-|dall-e|computer-use)[^"]*"' extended)"
  printf 'Models visible: %s total, %s OpenAI-compatible\n' "$model_count" "$openai_count"
  grep -o '"id":"[^"]*"' "$out" | sed 's/"id":"//; s/"$//' | sed -n '1,40p'

  if is_enabled "$REQUIRE_OPENAI_MODELS"; then
    if [[ "$openai_count" -gt 0 ]]; then
      pass "OpenAI-compatible models visible"
    else
      fail "OpenAI-compatible models are required but none are visible"
    fi
  fi

  if is_enabled "$REQUIRE_MODEL_VISIBLE"; then
    if grep -Fq "\"id\":\"$OPENAI_MODEL\"" "$out"; then
      pass "requested model is visible: $OPENAI_MODEL"
    else
      fail "requested model is not visible: $OPENAI_MODEL"
      echo "Hint: set TOKENGATE_OPENAI_MODEL to a visible model or update group/channel model routing." >&2
    fi
  fi
}

check_chat() {
  local out="$tmp_dir/chat.json"
  local err="$tmp_dir/chat.err"
  local status
  local body

  body="{
    \"model\": \"$OPENAI_MODEL\",
    \"max_tokens\": 32,
    \"messages\": [
      {\"role\": \"user\", \"content\": \"Reply with exactly: hello\"}
    ]
  }"

  status="$(post_json "$BASE_URL/v1/chat/completions" "$body" "$out" "$err")"
  if [[ "$status" -lt 200 || "$status" -ge 300 ]]; then
    fail "chat_completions HTTP $status"
    print_curl_error "$err"
    print_file_excerpt "$out" 40
    diagnose_status "$status" "$out"
    return
  fi

  pass "chat_completions HTTP $status"
  if grep -q '"choices"' "$out"; then
    pass "chat_completions response includes choices"
  else
    fail "chat_completions response did not include choices"
    print_file_excerpt "$out" 40
  fi
  sed -n '1,8p' "$out"
}

check_chat_stream() {
  local out="$tmp_dir/chat_stream.sse"
  local err="$tmp_dir/chat_stream.err"
  local status
  local body

  body="{
    \"model\": \"$OPENAI_MODEL\",
    \"max_tokens\": 32,
    \"stream\": true,
    \"stream_options\": {\"include_usage\": true},
    \"messages\": [
      {\"role\": \"user\", \"content\": \"Reply with exactly: hello\"}
    ]
  }"

  status="$(post_json "$BASE_URL/v1/chat/completions" "$body" "$out" "$err")"
  if [[ "$status" -lt 200 || "$status" -ge 300 ]]; then
    fail "chat_completions_stream HTTP $status"
    print_curl_error "$err"
    print_file_excerpt "$out" 60
    diagnose_status "$status" "$out"
    return
  fi

  pass "chat_completions_stream HTTP $status"
  if grep -q '^data: ' "$out"; then
    pass "stream emits SSE data lines"
  else
    fail "stream did not emit SSE data lines"
  fi
  if grep -q '^data: \[DONE\]' "$out"; then
    pass "stream emits [DONE]"
  else
    fail "stream did not emit [DONE]"
  fi
  if is_enabled "$REQUIRE_STREAM_USAGE"; then
    if grep -q '"usage"' "$out"; then
      pass "stream includes usage"
    else
      fail "stream usage chunk is required but was not found"
    fi
  fi
  sed -n '1,12p' "$out"
}

check_responses() {
  local out="$tmp_dir/responses.json"
  local err="$tmp_dir/responses.err"
  local status
  local body

  body="{
    \"model\": \"$OPENAI_MODEL\",
    \"max_output_tokens\": 32,
    \"input\": \"Reply with exactly: hello\"
  }"

  status="$(post_json "$BASE_URL/v1/responses" "$body" "$out" "$err")"
  if [[ "$status" -lt 200 || "$status" -ge 300 ]]; then
    fail "responses HTTP $status"
    print_curl_error "$err"
    print_file_excerpt "$out" 40
    diagnose_status "$status" "$out"
    return
  fi

  pass "responses HTTP $status"
  if grep -Eq '"object":"response"|"output"|"output_text"' "$out"; then
    pass "responses body has OpenAI-compatible response shape"
  else
    fail "responses body did not look like a response object"
    print_file_excerpt "$out" 40
  fi
  sed -n '1,8p' "$out"
}

check_node_sdk() {
  local sdk_probe_err="$tmp_dir/sdk_probe.err"
  case "$RUN_SDK" in
    0|false|no)
      pass "OpenAI Node SDK smoke skipped"
      return
      ;;
    auto|1|true|yes) ;;
    *)
      fail "TOKENGATE_RUN_P0_SDK must be auto, 1, or 0"
      return
      ;;
  esac

  if ! command -v node >/dev/null 2>&1; then
    if [[ "$RUN_SDK" == "auto" ]]; then
      warn "OpenAI Node SDK smoke skipped because node is not available"
    else
      fail "OpenAI Node SDK smoke requires node"
    fi
    return
  fi

  if ! node --input-type=module -e "await import('openai')" >/dev/null 2>"$sdk_probe_err"; then
    if [[ "$RUN_SDK" == "auto" ]]; then
      warn "OpenAI Node SDK smoke skipped because package 'openai' is not installed"
      print_file_excerpt "$sdk_probe_err" 3
    else
      fail "OpenAI Node SDK smoke requires package 'openai'"
      print_file_excerpt "$sdk_probe_err" 8
    fi
    return
  fi

  if TOKENGATE_BASE_URL="$BASE_URL" TOKENGATE_API_KEY="$API_KEY" TOKENGATE_OPENAI_MODEL="$OPENAI_MODEL" node --input-type=module <<'NODE'
import OpenAI from 'openai';

const baseURL = `${process.env.TOKENGATE_BASE_URL.replace(/\/$/, '')}/v1`;
const model = process.env.TOKENGATE_OPENAI_MODEL;
const client = new OpenAI({
  apiKey: process.env.TOKENGATE_API_KEY,
  baseURL,
});

const models = await client.models.list();
if (!models.data?.some((entry) => entry.id === model)) {
  throw new Error(`model ${model} was not returned by client.models.list()`);
}

const completion = await client.chat.completions.create({
  model,
  max_tokens: 32,
  messages: [{ role: 'user', content: 'Reply with exactly: hello' }],
});

if (!completion.choices?.length) {
  throw new Error('chat.completions.create returned no choices');
}

console.log(`PASS OpenAI SDK chat completion with ${model}`);
NODE
  then
    pass "OpenAI Node SDK smoke"
  else
    fail "OpenAI Node SDK smoke failed"
  fi
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ -z "$BASE_URL" || -z "$API_KEY" ]]; then
  usage >&2
  exit 2
fi

normalize_base_url

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

printf '\nTokenGate P0 compatibility suite\n'
printf 'Backend: %s\n' "$BASE_URL"
printf 'Model:   %s\n' "$OPENAI_MODEL"

if is_enabled "$RUN_MODELS"; then
  printf '\nModel visibility\n'
  check_models
fi

if is_enabled "$RUN_CHAT"; then
  printf '\nChat Completions\n'
  check_chat
fi

if is_enabled "$RUN_CHAT_STREAM"; then
  printf '\nChat Completions streaming\n'
  check_chat_stream
fi

if is_enabled "$RUN_RESPONSES"; then
  printf '\nResponses API\n'
  check_responses
fi

printf '\nOpenAI SDK compatibility\n'
check_node_sdk

printf '\nSummary: %d failure(s), %d warning(s)\n' "$failures" "$warnings"
if [[ "$failures" -gt 0 ]]; then
  exit 1
fi
