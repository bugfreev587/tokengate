#!/usr/bin/env bash
set -euo pipefail

FRONTEND_URL="${TOKENGATE_FRONTEND_URL:-}"
BACKEND_URL="${TOKENGATE_BACKEND_URL:-}"
API_KEY="${TOKENGATE_API_KEY:-}"
EXPECTED_CONTACT_INFO="${TOKENGATE_EXPECTED_CONTACT_INFO:-}"

if [[ -z "$FRONTEND_URL" || -z "$BACKEND_URL" || -z "$API_KEY" ]]; then
  cat >&2 <<'USAGE'
Usage:
  TOKENGATE_FRONTEND_URL="https://your-frontend-domain" \
  TOKENGATE_BACKEND_URL="https://your-backend-domain" \
  TOKENGATE_API_KEY="sk-..." \
  tools/tokengate_private_beta_acceptance.sh

Optional:
  TOKENGATE_EXPECTED_CONTACT_INFO="support@example.com"
  TOKENGATE_CLAUDE_MODEL="claude-haiku-4-5-20251001"
USAGE
  exit 2
fi

BACKEND_URL="${BACKEND_URL%/}"
if [[ "$BACKEND_URL" == */api/v1 ]]; then
  echo "WARN TOKENGATE_BACKEND_URL ended with /api/v1; stripping it for gateway endpoints." >&2
  BACKEND_URL="${BACKEND_URL%/api/v1}"
fi

echo
echo "== Launch readiness =="
TOKENGATE_FRONTEND_URL="$FRONTEND_URL" \
TOKENGATE_BACKEND_URL="$BACKEND_URL" \
TOKENGATE_LAUNCH_PROFILE=private \
TOKENGATE_EXPECTED_CONTACT_INFO="$EXPECTED_CONTACT_INFO" \
TOKENGATE_RUN_API_SMOKE=0 \
tools/tokengate_launch_readiness.sh

echo
echo "== Model visibility =="
TOKENGATE_BASE_URL="$BACKEND_URL" \
TOKENGATE_API_KEY="$API_KEY" \
TOKENGATE_REQUIRE_CLAUDE_MODELS=1 \
TOKENGATE_REQUIRE_OPENAI_MODELS=0 \
tools/tokengate_model_visibility.sh

echo
echo "== Claude gateway smoke =="
TOKENGATE_BASE_URL="$BACKEND_URL" \
TOKENGATE_API_KEY="$API_KEY" \
TOKENGATE_RUN_OPENAI=0 \
bash tools/tokengate_smoke_test.sh

echo
echo "== API key usage =="
curl -sS "$BACKEND_URL/v1/usage" \
  -H "Authorization: Bearer $API_KEY" \
  | sed -n '1,80p'

echo
echo "Private beta acceptance smoke completed."
