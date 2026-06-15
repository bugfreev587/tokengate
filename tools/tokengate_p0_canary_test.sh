#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CANARY="$ROOT_DIR/tools/tokengate_p0_canary.sh"

fail() {
  printf 'FAIL %s\n' "$1" >&2
  exit 1
}

assert_file_contains() {
  local file="$1"
  local needle="$2"
  if ! grep -Fq -- "$needle" "$file"; then
    printf 'File did not contain expected text: %s\n' "$needle" >&2
    sed -n '1,80p' "$file" >&2 || true
    exit 1
  fi
}

assert_file_not_contains() {
  local file="$1"
  local needle="$2"
  if grep -Fq -- "$needle" "$file"; then
    printf 'File contained forbidden text: %s\n' "$needle" >&2
    sed -n '1,80p' "$file" >&2 || true
    exit 1
  fi
}

write_fake_suite() {
  local path="$1"
  local exit_code="$2"
  local marker="$3"
  cat > "$path" <<EOF
#!/usr/bin/env bash
set -euo pipefail
printf 'fake suite marker: %s\\n' "$marker"
printf 'backend=%s model=%s\\n' "\${TOKENGATE_BASE_URL:-}" "\${TOKENGATE_OPENAI_MODEL:-}"
printf 'api_key=%s\\n' "\${TOKENGATE_API_KEY:-}"
if [[ "$exit_code" -ne 0 ]]; then
  printf 'FAIL chat_completions HTTP 500\\n' >&2
fi
exit "$exit_code"
EOF
  chmod +x "$path"
}

write_fake_curl() {
  local path="$1"
  cat > "$path" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
payload=""
url=""
while [[ "$#" -gt 0 ]]; do
  case "$1" in
    -d|--data|--data-raw)
      payload="$2"
      shift 2
      ;;
    http://*|https://*)
      url="$1"
      shift
      ;;
    *)
      shift
      ;;
  esac
done
printf '%s\n' "$url" > "${WEBHOOK_URL_FILE:?}"
printf '%s\n' "$payload" > "${WEBHOOK_PAYLOAD_FILE:?}"
printf 'ok\n'
EOF
  chmod +x "$path"
}

run_success_without_alert() {
  local tmp_dir="$1"
  local suite="$tmp_dir/suite_success.sh"
  local curl_bin="$tmp_dir/fake_curl.sh"
  local state_dir="$tmp_dir/state_success"
  local payload_file="$tmp_dir/success_payload.json"
  local url_file="$tmp_dir/success_url.txt"

  write_fake_suite "$suite" 0 "success"
  write_fake_curl "$curl_bin"

  WEBHOOK_PAYLOAD_FILE="$payload_file" \
  WEBHOOK_URL_FILE="$url_file" \
  TOKENGATE_BASE_URL="https://backend.example" \
  TOKENGATE_API_KEY="sk-test-secret" \
  TOKENGATE_OPENAI_MODEL="gpt-5.4" \
  TOKENGATE_P0_CANARY_SUITE="$suite" \
  TOKENGATE_P0_CANARY_CURL="$curl_bin" \
  TOKENGATE_P0_CANARY_STATE_DIR="$state_dir" \
  "$CANARY"

  assert_file_contains "$state_dir/latest.json" '"status":"ok"'
  assert_file_contains "$state_dir/latest.log" 'fake suite marker: success'
  assert_file_not_contains "$state_dir/latest.log" 'sk-test-secret'
  if [[ -e "$payload_file" ]]; then
    fail "success run sent a webhook even though notify_on defaults to failure"
  fi
}

run_failure_with_alert() {
  local tmp_dir="$1"
  local suite="$tmp_dir/suite_failure.sh"
  local curl_bin="$tmp_dir/fake_curl.sh"
  local state_dir="$tmp_dir/state_failure"
  local payload_file="$tmp_dir/failure_payload.json"
  local url_file="$tmp_dir/failure_url.txt"

  write_fake_suite "$suite" 1 "failure"
  write_fake_curl "$curl_bin"

  set +e
  WEBHOOK_PAYLOAD_FILE="$payload_file" \
  WEBHOOK_URL_FILE="$url_file" \
  TOKENGATE_BASE_URL="https://backend.example" \
  TOKENGATE_API_KEY="tg_test-key:super-secret_123" \
  TOKENGATE_OPENAI_MODEL="gpt-5.4" \
  TOKENGATE_P0_CANARY_SUITE="$suite" \
  TOKENGATE_P0_CANARY_CURL="$curl_bin" \
  TOKENGATE_P0_CANARY_STATE_DIR="$state_dir" \
  TOKENGATE_P0_CANARY_NOTIFY_WEBHOOK_URL="https://hooks.example/p0" \
  "$CANARY"
  status="$?"
  set -e

  if [[ "$status" -eq 0 ]]; then
    fail "failure run exited 0"
  fi
  assert_file_contains "$state_dir/latest.json" '"status":"failed"'
  assert_file_contains "$payload_file" '"status":"failed"'
  assert_file_contains "$payload_file" 'FAIL chat_completions HTTP 500'
  assert_file_not_contains "$state_dir/latest.log" 'tg_test-key:super-secret_123'
  assert_file_not_contains "$payload_file" 'tg_test-key:super-secret_123'
  assert_file_contains "$url_file" 'https://hooks.example/p0'
}

run_failure_with_discord_alert() {
  local tmp_dir="$1"
  local suite="$tmp_dir/suite_discord.sh"
  local curl_bin="$tmp_dir/fake_curl.sh"
  local state_dir="$tmp_dir/state_discord"
  local payload_file="$tmp_dir/discord_payload.json"
  local url_file="$tmp_dir/discord_url.txt"

  write_fake_suite "$suite" 1 "discord"
  write_fake_curl "$curl_bin"

  set +e
  WEBHOOK_PAYLOAD_FILE="$payload_file" \
  WEBHOOK_URL_FILE="$url_file" \
  GITHUB_EVENT_NAME="schedule" \
  GITHUB_REF_NAME="main" \
  GITHUB_REPOSITORY="tokengate/token-gate" \
  GITHUB_RUN_ID="123456" \
  GITHUB_SERVER_URL="https://github.com" \
  TOKENGATE_BASE_URL="https://backend.example" \
  TOKENGATE_API_KEY="sk-test-secret" \
  TOKENGATE_OPENAI_MODEL="gpt-5.4" \
  TOKENGATE_P0_CANARY_SUITE="$suite" \
  TOKENGATE_P0_CANARY_CURL="$curl_bin" \
  TOKENGATE_P0_CANARY_STATE_DIR="$state_dir" \
  TOKENGATE_P0_CANARY_NOTIFY_WEBHOOK_URL="https://discord.com/api/webhooks/test/webhook" \
  "$CANARY"
  status="$?"
  set -e

  if [[ "$status" -eq 0 ]]; then
    fail "discord failure run exited 0"
  fi
  assert_file_contains "$payload_file" '"content":'
  assert_file_contains "$payload_file" 'TokenGate P0 canary failed.'
  assert_file_contains "$payload_file" 'Repository: tokengate/token-gate'
  assert_file_contains "$payload_file" 'Run: https://github.com/tokengate/token-gate/actions/runs/123456'
  assert_file_not_contains "$state_dir/latest.log" 'sk-test-secret'
  assert_file_not_contains "$payload_file" 'sk-test-secret'
  assert_file_contains "$url_file" 'https://discord.com/api/webhooks/test/webhook'
}

run_success_with_always_alert() {
  local tmp_dir="$1"
  local suite="$tmp_dir/suite_always.sh"
  local curl_bin="$tmp_dir/fake_curl.sh"
  local state_dir="$tmp_dir/state_always"
  local payload_file="$tmp_dir/always_payload.json"
  local url_file="$tmp_dir/always_url.txt"

  write_fake_suite "$suite" 0 "always"
  write_fake_curl "$curl_bin"

  WEBHOOK_PAYLOAD_FILE="$payload_file" \
  WEBHOOK_URL_FILE="$url_file" \
  TOKENGATE_BASE_URL="https://backend.example" \
  TOKENGATE_API_KEY="sk-test-secret" \
  TOKENGATE_OPENAI_MODEL="gpt-5.4" \
  TOKENGATE_P0_CANARY_SUITE="$suite" \
  TOKENGATE_P0_CANARY_CURL="$curl_bin" \
  TOKENGATE_P0_CANARY_STATE_DIR="$state_dir" \
  TOKENGATE_P0_CANARY_NOTIFY_WEBHOOK_URL="https://hooks.example/p0" \
  TOKENGATE_P0_CANARY_NOTIFY_ON="always" \
  "$CANARY"

  assert_file_contains "$payload_file" '"status":"ok"'
  assert_file_contains "$payload_file" 'fake suite marker: always'
  assert_file_not_contains "$payload_file" 'sk-test-secret'
}

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

run_success_without_alert "$tmp_dir"
run_failure_with_alert "$tmp_dir"
run_failure_with_discord_alert "$tmp_dir"
run_success_with_always_alert "$tmp_dir"

printf 'PASS tokengate_p0_canary_test\n'
