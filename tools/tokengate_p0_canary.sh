#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

BASE_URL="${TOKENGATE_BASE_URL:-}"
API_KEY="${TOKENGATE_API_KEY:-}"
OPENAI_MODEL="${TOKENGATE_OPENAI_MODEL:-gpt-5.4}"
CANARY_NAME="${TOKENGATE_P0_CANARY_NAME:-tokengate-p0-canary}"
STATE_DIR="${TOKENGATE_P0_CANARY_STATE_DIR:-${TMPDIR:-/tmp}/tokengate-p0-canary}"
SUITE="${TOKENGATE_P0_CANARY_SUITE:-$ROOT_DIR/tools/tokengate_p0_compatibility_suite.sh}"
CURL_BIN="${TOKENGATE_P0_CANARY_CURL:-curl}"
WEBHOOK_URL="${TOKENGATE_P0_CANARY_NOTIFY_WEBHOOK_URL:-}"
NOTIFY_ON="${TOKENGATE_P0_CANARY_NOTIFY_ON:-failure}"
INTERVAL_SECONDS="${TOKENGATE_P0_CANARY_INTERVAL_SECONDS:-0}"
MAX_RUNS="${TOKENGATE_P0_CANARY_MAX_RUNS:-0}"
OUTPUT_LINES="${TOKENGATE_P0_CANARY_OUTPUT_LINES:-80}"

usage() {
  cat <<'USAGE'
Usage:
  TOKENGATE_BASE_URL="https://your-backend-domain" \
  TOKENGATE_API_KEY="sk-..." \
  tools/tokengate_p0_canary.sh

Optional:
  TOKENGATE_OPENAI_MODEL="gpt-5.4"
  TOKENGATE_P0_CANARY_NAME="tokengate-p0-canary"
  TOKENGATE_P0_CANARY_STATE_DIR="/tmp/tokengate-p0-canary"
  TOKENGATE_P0_CANARY_NOTIFY_WEBHOOK_URL="https://hooks.example/..."
  TOKENGATE_P0_CANARY_NOTIFY_ON=failure|always|never
  TOKENGATE_P0_CANARY_INTERVAL_SECONDS=300
  TOKENGATE_P0_CANARY_MAX_RUNS=0
  TOKENGATE_P0_CANARY_OUTPUT_LINES=80

By default the canary runs once. Set TOKENGATE_P0_CANARY_INTERVAL_SECONDS to a
positive integer to loop. In loop mode, TOKENGATE_P0_CANARY_MAX_RUNS=0 means
run forever.
USAGE
}

fail_config() {
  printf 'FAIL %s\n' "$1" >&2
  usage >&2
  exit 2
}

is_nonnegative_integer() {
  case "${1:-}" in
    ''|*[!0-9]*) return 1 ;;
    *) return 0 ;;
  esac
}

timestamp_utc() {
  date -u '+%Y-%m-%dT%H:%M:%SZ'
}

json_escape() {
  local value="$1"
  value="${value//\\/\\\\}"
  value="${value//\"/\\\"}"
  value="${value//$'\r'/}"
  value="${value//$'\n'/\\n}"
  value="${value//$'\t'/\\t}"
  printf '%s' "$value"
}

redact_file() {
  local in_file="$1"
  local out_file="$2"
  sed -E \
    -e 's/sk-[A-Za-z0-9_-]+/sk-<redacted>/g' \
    -e 's/tg_[A-Za-z0-9_-]+:[A-Za-z0-9_-]+/tg_<redacted>/g' \
    "$in_file" > "$out_file"
}

write_status_json() {
  local path="$1"
  local status="$2"
  local exit_code="$3"
  local started_at="$4"
  local finished_at="$5"
  local log_path="$6"
  local excerpt="$7"
  local escaped_name
  local escaped_backend
  local escaped_model
  local escaped_log_path
  local escaped_excerpt

  escaped_name="$(json_escape "$CANARY_NAME")"
  escaped_backend="$(json_escape "$BASE_URL")"
  escaped_model="$(json_escape "$OPENAI_MODEL")"
  escaped_log_path="$(json_escape "$log_path")"
  escaped_excerpt="$(json_escape "$excerpt")"

  cat > "$path" <<EOF
{"name":"$escaped_name","status":"$status","severity":"P0","backend_url":"$escaped_backend","model":"$escaped_model","exit_code":$exit_code,"started_at":"$started_at","finished_at":"$finished_at","log_path":"$escaped_log_path","output_excerpt":"$escaped_excerpt"}
EOF
}

is_discord_webhook_url() {
  local url="$1"
  [[ "$url" == *"discord.com/api/webhooks/"* || "$url" == *"discordapp.com/api/webhooks/"* ]]
}

github_run_url() {
  local server_url="${GITHUB_SERVER_URL:-https://github.com}"
  local repository="${GITHUB_REPOSITORY:-}"
  local run_id="${GITHUB_RUN_ID:-}"

  if [[ -z "$repository" || -z "$run_id" ]]; then
    printf 'unknown'
    return
  fi

  printf '%s/%s/actions/runs/%s' "${server_url%/}" "$repository" "$run_id"
}

write_discord_payload() {
  local path="$1"
  local status="$2"
  local exit_code="$3"
  local run_url
  local message
  local escaped_message

  run_url="$(github_run_url)"
  message=$(
    cat <<EOF
TokenGate P0 canary $status.
Repository: ${GITHUB_REPOSITORY:-unknown}
Branch: ${GITHUB_REF_NAME:-unknown}
Trigger: ${GITHUB_EVENT_NAME:-manual}
Backend: $BASE_URL
Model: $OPENAI_MODEL
Exit code: $exit_code
Run: $run_url
EOF
  )
  escaped_message="$(json_escape "$message")"

  printf '{"content":"%s"}\n' "$escaped_message" > "$path"
}

send_webhook() {
  local payload_file="$1"
  local status="${2:-unknown}"
  local exit_code="${3:-0}"
  local send_payload_file="$payload_file"

  if [[ -z "$WEBHOOK_URL" ]]; then
    return 0
  fi
  if ! command -v "$CURL_BIN" >/dev/null 2>&1; then
    printf 'WARN webhook skipped because curl command is not available: %s\n' "$CURL_BIN" >&2
    return 1
  fi

  if is_discord_webhook_url "$WEBHOOK_URL"; then
    send_payload_file="${payload_file%.json}.discord.json"
    write_discord_payload "$send_payload_file" "$status" "$exit_code"
  fi

  "$CURL_BIN" -sS \
    --connect-timeout 20 \
    --max-time 60 \
    -X POST "$WEBHOOK_URL" \
    -H "Content-Type: application/json" \
    -d "$(sed -n '1,$p' "$send_payload_file")" >/dev/null
}

should_notify() {
  local status="$1"
  case "$NOTIFY_ON" in
    never) return 1 ;;
    always) return 0 ;;
    failure)
      [[ "$status" != "ok" ]]
      ;;
  esac
}

run_once() {
  local run_id
  local started_at
  local finished_at
  local raw_log
  local log_file
  local latest_log
  local latest_json
  local payload_file
  local status
  local suite_exit
  local webhook_exit=0
  local excerpt

  run_id="$(date -u '+%Y%m%dT%H%M%SZ')"
  started_at="$(timestamp_utc)"
  raw_log="$STATE_DIR/$run_id.raw.log"
  log_file="$STATE_DIR/$run_id.log"
  latest_log="$STATE_DIR/latest.log"
  latest_json="$STATE_DIR/latest.json"
  payload_file="$STATE_DIR/$run_id.payload.json"

  set +e
  TOKENGATE_BASE_URL="$BASE_URL" \
  TOKENGATE_API_KEY="$API_KEY" \
  TOKENGATE_OPENAI_MODEL="$OPENAI_MODEL" \
  "$SUITE" > "$raw_log" 2>&1
  suite_exit="$?"
  set -e

  redact_file "$raw_log" "$log_file"
  cp "$log_file" "$latest_log"
  rm -f "$raw_log"

  if [[ "$suite_exit" -eq 0 ]]; then
    status="ok"
  else
    status="failed"
  fi
  finished_at="$(timestamp_utc)"
  excerpt="$(tail -n "$OUTPUT_LINES" "$log_file" || true)"
  write_status_json "$latest_json" "$status" "$suite_exit" "$started_at" "$finished_at" "$log_file" "$excerpt"
  write_status_json "$payload_file" "$status" "$suite_exit" "$started_at" "$finished_at" "$log_file" "$excerpt"

  if should_notify "$status"; then
    if send_webhook "$payload_file" "$status" "$suite_exit"; then
      printf 'PASS webhook notification sent status=%s\n' "$status"
    else
      webhook_exit=1
      printf 'FAIL webhook notification failed status=%s\n' "$status" >&2
    fi
  fi

  if [[ "$status" == "ok" && "$webhook_exit" -eq 0 ]]; then
    printf 'PASS %s status=ok model=%s log=%s\n' "$CANARY_NAME" "$OPENAI_MODEL" "$log_file"
    return 0
  fi

  printf 'FAIL %s status=%s exit=%s log=%s\n' "$CANARY_NAME" "$status" "$suite_exit" "$log_file" >&2
  return 1
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ -z "$BASE_URL" ]]; then
  fail_config "TOKENGATE_BASE_URL is required"
fi
if [[ "$BASE_URL" != http://* && "$BASE_URL" != https://* ]]; then
  fail_config "TOKENGATE_BASE_URL must start with http:// or https://"
fi
if [[ -z "$API_KEY" ]]; then
  fail_config "TOKENGATE_API_KEY is required"
fi
if [[ ! -x "$SUITE" ]]; then
  fail_config "P0 suite is not executable: $SUITE"
fi
case "$NOTIFY_ON" in
  failure|always|never) ;;
  *) fail_config "TOKENGATE_P0_CANARY_NOTIFY_ON must be failure, always, or never" ;;
esac
if ! is_nonnegative_integer "$INTERVAL_SECONDS"; then
  fail_config "TOKENGATE_P0_CANARY_INTERVAL_SECONDS must be a non-negative integer"
fi
if ! is_nonnegative_integer "$MAX_RUNS"; then
  fail_config "TOKENGATE_P0_CANARY_MAX_RUNS must be a non-negative integer"
fi
if ! is_nonnegative_integer "$OUTPUT_LINES" || [[ "$OUTPUT_LINES" -eq 0 ]]; then
  fail_config "TOKENGATE_P0_CANARY_OUTPUT_LINES must be a positive integer"
fi

BASE_URL="${BASE_URL%/}"
mkdir -p "$STATE_DIR"

run_count=0
last_exit=0
while true; do
  set +e
  run_once
  last_exit="$?"
  set -e

  run_count=$((run_count + 1))
  if [[ "$INTERVAL_SECONDS" -eq 0 ]]; then
    exit "$last_exit"
  fi
  if [[ "$MAX_RUNS" -gt 0 && "$run_count" -ge "$MAX_RUNS" ]]; then
    exit "$last_exit"
  fi
  sleep "$INTERVAL_SECONDS"
done
