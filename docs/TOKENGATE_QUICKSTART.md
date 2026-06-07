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

## 2. Anthropic-Compatible Request

Use `/v1/messages` for Claude-compatible requests.

```bash
curl "$TOKENGATE_BASE_URL/v1/messages" \
  -H "Authorization: Bearer $TOKENGATE_API_KEY" \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude-haiku-4-5-20251001",
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

## 3. OpenAI-Compatible Request

Use `/v1/chat/completions` for OpenAI-compatible chat requests.

```bash
curl "$TOKENGATE_BASE_URL/v1/chat/completions" \
  -H "Authorization: Bearer $TOKENGATE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-5.4",
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
  model: "gpt-5.4",
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
TOKENGATE_OPENAI_MODEL="gpt-5.4" \
tools/tokengate_p0_compatibility_suite.sh
```

This verifies `/v1/models`, non-streaming chat, streaming chat with usage, and `/v1/responses`.
