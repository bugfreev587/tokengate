#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORKFLOW="$ROOT_DIR/.github/workflows/p0-regression-monitor.yml"

fail() {
  printf 'FAIL %s\n' "$1" >&2
  exit 1
}

assert_file_contains() {
  local file="$1"
  local needle="$2"
  if ! grep -Fq -- "$needle" "$file"; then
    printf 'File did not contain expected text: %s\n' "$needle" >&2
    sed -n '1,180p' "$file" >&2 || true
    exit 1
  fi
}

if [[ ! -f "$WORKFLOW" ]]; then
  fail "P0 regression monitor workflow is missing: $WORKFLOW"
fi

assert_file_contains "$WORKFLOW" 'name: TokenGate P0 Regression Monitor'
assert_file_contains "$WORKFLOW" 'workflow_dispatch:'
assert_file_contains "$WORKFLOW" 'TOKENGATE_BASE_URL: ${{ vars.TOKENGATE_REGRESSION_BASE_URL }}'
assert_file_contains "$WORKFLOW" 'TOKENGATE_API_KEY: ${{ secrets.TOKENGATE_REGRESSION_API_KEY }}'
assert_file_contains "$WORKFLOW" 'TOKENGATE_OPENAI_MODEL: ${{ vars.TOKENGATE_REGRESSION_OPENAI_MODEL || '\''gpt-5.4'\'' }}'
assert_file_contains "$WORKFLOW" 'TOKENGATE_P0_CANARY_NOTIFY_WEBHOOK_URL: ${{ secrets.REGRESSION_ALERT_WEBHOOK_URL }}'
assert_file_contains "$WORKFLOW" 'TOKENGATE_P0_CANARY_NOTIFY_ON: never'
assert_file_contains "$WORKFLOW" 'tools/tokengate_p0_canary.sh'
assert_file_contains "$WORKFLOW" 'actions/upload-artifact'
assert_file_contains "$WORKFLOW" 'Notify Discord on failure'
assert_file_contains "$WORKFLOW" 'TokenGate P0 regression monitor failed.'
assert_file_contains "$WORKFLOW" '--data-binary @/tmp/tokengate-discord-payload.json'

if ! grep -Eq 'cron: ["'\''][0-9]+ \* \* \* \*["'\'']' "$WORKFLOW"; then
  fail "P0 regression monitor workflow must run once per hour"
fi

printf 'PASS tokengate_regression_workflow_test\n'
