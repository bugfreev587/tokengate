# TokenGate Quickstart

This quickstart verifies the basic TokenGate loop:

- create an API key
- send a model request
- confirm usage is recorded
- understand how balance is deducted

## 1. Create An API Key

1. Sign in to TokenGate.
2. Open **API Keys**.
3. Create a new key.
4. Keep the key private. It should be used as a bearer token.

Example:

```bash
export TOKENGATE_API_KEY="sk-..."
export TOKENGATE_BASE_URL="https://tokengate-production.up.railway.app"
```

For the current Railway deployment shape, the base URL should be the backend domain, not the Vercel frontend domain.

Before copying a model ID into a client, list the models visible to this key:

```bash
curl "$TOKENGATE_BASE_URL/v1/models" \
  -H "Authorization: Bearer $TOKENGATE_API_KEY"
```

Model visibility depends on the group assigned to the key. As of 2026-09-21, the current TokenGate IDs are:

| Client family | Current model IDs |
| --- | --- |
| Codex | `gpt-6-astra`, `gpt-5.6` (Sol alias), `gpt-5.6-sol`, `gpt-5.6-terra`, `gpt-5.6-luna` |
| Claude Code | `claude-fable-5-1`, `claude-opus-5`, `claude-sonnet-5`, `claude-haiku-4-5-20251001` |

`gpt-5.4` is no longer available to ChatGPT-sign-in Codex users. `gpt-5.5` remains supported for API-backed accounts, but ChatGPT-sign-in access is scheduled to retire on 2026-10-14. Claude Fable 5.1 requires Claude Code 2.1.257 or newer.

## 2. Claude Code And Anthropic-Compatible Requests

Configure Claude Code for the current shell:

```bash
export ANTHROPIC_BASE_URL="$TOKENGATE_BASE_URL"
export ANTHROPIC_AUTH_TOKEN="$TOKENGATE_API_KEY"
export CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1
claude --model claude-sonnet-5
```

Use `/v1/messages` for Claude-compatible requests.

```bash
curl "$TOKENGATE_BASE_URL/v1/messages" \
  -H "Authorization: Bearer $TOKENGATE_API_KEY" \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude-sonnet-5",
    "max_tokens": 64,
    "messages": [
      {
        "role": "user",
        "content": "Reply with exactly: hello"
      }
    ]
  }'
```

Expected result:

- the request returns a model response
- the API key `Last Used` timestamp updates
- a new row appears in **Usage**
- the cost is deducted from the user balance

## 3. Codex CLI And OpenAI-Compatible Requests

Create `~/.codex/config.toml` (or `%userprofile%\\.codex\\config.toml` on Windows):

```toml
model_provider = "OpenAI"
model = "gpt-5.6-terra"
review_model = "gpt-5.6-terra"
model_reasoning_effort = "xhigh"
disable_response_storage = true
model_context_window = 1050000
model_auto_compact_token_limit = 900000

[model_providers.OpenAI]
name = "OpenAI"
base_url = "https://tokengate-production.up.railway.app/v1"
wire_api = "responses"
requires_openai_auth = true
```

Then create `~/.codex/auth.json` (or `%userprofile%\\.codex\\auth.json`):

```json
{
  "OPENAI_API_KEY": "replace-with-your-TokenGate-key"
}
```

Replace the example base URL if your TokenGate deployment uses a different backend domain, then restart Codex.

Use `/v1/chat/completions` for OpenAI-compatible chat requests.

```bash
curl "$TOKENGATE_BASE_URL/v1/chat/completions" \
  -H "Authorization: Bearer $TOKENGATE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-5.6-terra",
    "messages": [
      {
        "role": "user",
        "content": "Reply with exactly: hello"
      }
    ]
  }'
```

Expected result:

- the request succeeds if an OpenAI upstream account is assigned to the user's group
- usage appears in **Usage**
- dashboard totals update after refresh

OpenAI SDK example:

```js
import OpenAI from "openai";

const client = new OpenAI({
  apiKey: process.env.TOKENGATE_API_KEY,
  baseURL: "https://tokengate-production.up.railway.app/v1",
});

const response = await client.chat.completions.create({
  model: "gpt-5.6-terra",
  messages: [{ role: "user", content: "Say hi in one short sentence." }],
});

console.log(response.choices[0]?.message?.content);
```

Customers should not use the Vercel frontend URL as the SDK base URL. The frontend domain is for the web dashboard only.

## 4. Verify Account Routing

For each provider account:

1. Open **Admin → Accounts**.
2. Confirm the account is active.
3. Assign the account to the target user group.
4. Use **Test Account Connection**.
5. Send a real API key request after the account test succeeds.

If the account test fails with `405`, the frontend is calling the frontend domain instead of the backend API domain. Verify the Vercel environment variable:

```env
VITE_API_BASE_URL=https://your-railway-backend-domain/api/v1
VITE_BUILD_TARGET=standalone
```

## 5. Billing Expectations

TokenGate treats usage units and billing units separately:

- text models should be priced as input/output cost per 1M tokens
- image models should be priced per image or provider-native output unit
- video models should be priced per job, second, or provider-native unit
- all successful usage deducts from a wallet balance or included plan balance

The user-facing rule is simple:

**model usage is metered transparently, then settled against the account balance.**

## 6. First Production Smoke Test

Run this after every production deployment:

1. Open the landing page.
2. Sign in.
3. Create or reuse an API key.
4. Send one Claude-compatible request.
5. Send one OpenAI-compatible request.
6. Confirm `Last Used` updates.
7. Confirm **Usage** records both requests.
8. Confirm dashboard totals update.
9. Confirm balance changed by the expected amount.

You can also run the repository smoke test:

```bash
TOKENGATE_BASE_URL="https://your-backend-domain" \
TOKENGATE_API_KEY="sk-..." \
bash tools/tokengate_smoke_test.sh
```

Set `TOKENGATE_RUN_OPENAI=0` or `TOKENGATE_RUN_CLAUDE=0` while a provider is not configured yet.

Before calling the OpenAI-compatible surface production-ready, run the P0 compatibility suite:

```bash
TOKENGATE_BASE_URL="https://your-backend-domain" \
TOKENGATE_API_KEY="sk-..." \
TOKENGATE_OPENAI_MODEL="gpt-5.6-terra" \
tools/tokengate_p0_compatibility_suite.sh
```

This verifies `/v1/models`, non-streaming chat, streaming chat with usage, and `/v1/responses`.
