# TokenGate Operations Runbook

This runbook covers the operational checks required before public launch.

For the consolidated Railway/Vercel variable list, see [TOKENGATE_PRODUCTION_ENV_CHECKLIST.md](TOKENGATE_PRODUCTION_ENV_CHECKLIST.md).

## 1. Environment Check

Run this before every production deploy or after changing Railway/Vercel variables:

```bash
AUTO_SETUP=true \
RUN_MODE=standard \
SERVER_PORT=8080 \
JWT_SECRET=<secret> \
TOTP_ENCRYPTION_KEY=<secret> \
ADMIN_EMAIL=<email> \
ADMIN_PASSWORD=<secret> \
DATABASE_URL=<postgres-url> \
REDIS_URL=<redis-url> \
FRONTEND_URL=https://<frontend-domain> \
CORS_ALLOWED_ORIGINS=https://<frontend-domain> \
VITE_API_BASE_URL=https://<backend-domain>/api/v1 \
VITE_BUILD_TARGET=standalone \
bash tools/check_tokengate_env.sh all
```

## 2. Production Smoke Test

Run this after every frontend/backend deploy:

```bash
TOKENGATE_FRONTEND_URL=https://<frontend-domain> \
TOKENGATE_BACKEND_URL=https://<backend-domain> \
TOKENGATE_API_KEY=sk-... \
TOKENGATE_EXPECTED_CONTACT_INFO=bugfreev587@gmail.com \
tools/tokengate_private_beta_acceptance.sh
```

This verifies public pages plus protected SPA entry routes such as `/dashboard`, `/usage`, `/admin/accounts`, and `/admin/launch-readiness`. It also checks CORS, Claude model visibility, Claude gateway smoke, and API-key `/v1/usage` metering.

Use `TOKENGATE_LAUNCH_PROFILE=private` for invite-only beta checks. Use `TOKENGATE_LAUNCH_PROFILE=public` before public self-serve launch; public mode fails if support contact, password reset, registration, or required payment settings are missing.

For the P0 OpenAI-compatible release gate, run:

```bash
TOKENGATE_BASE_URL=https://<backend-domain> \
TOKENGATE_API_KEY=sk-... \
TOKENGATE_OPENAI_MODEL=gpt-5.4 \
tools/tokengate_p0_compatibility_suite.sh
```

Expected:

- `/v1/models` returns `2xx`
- the configured OpenAI-compatible model is visible
- `/v1/chat/completions` returns `2xx`
- streaming `/v1/chat/completions` emits SSE data, `[DONE]`, and usage
- `/v1/responses` returns `2xx`
- API key `Last Used` changes
- **Usage** records appear after refresh
- dashboard totals update after refresh

## P0 Production Canary

After launch, create a dedicated TokenGate API key for canary traffic. Put it in
the same user group and account routing path that normal OpenAI-compatible
customers should use. Name the key clearly, for example
`p0-production-canary`, and rotate it whenever the upstream account or group
assignment changes.

Run one canary check:

```bash
TOKENGATE_BASE_URL=https://<backend-domain> \
TOKENGATE_API_KEY=sk-... \
TOKENGATE_OPENAI_MODEL=gpt-5.4 \
TOKENGATE_P0_CANARY_STATE_DIR=/var/tmp/tokengate-p0-canary \
tools/tokengate_p0_canary.sh
```

Run it every five minutes from a long-running scheduler:

```bash
TOKENGATE_BASE_URL=https://<backend-domain> \
TOKENGATE_API_KEY=sk-... \
TOKENGATE_P0_CANARY_NOTIFY_WEBHOOK_URL=https://<alert-webhook> \
TOKENGATE_P0_CANARY_INTERVAL_SECONDS=300 \
tools/tokengate_p0_canary.sh
```

For GitHub Actions, Railway Cron, or another external scheduler, leave
`TOKENGATE_P0_CANARY_INTERVAL_SECONDS` unset and schedule the command externally.
The latest result is written to `latest.json` and `latest.log` inside
`TOKENGATE_P0_CANARY_STATE_DIR`.

Default alert behavior:

- `TOKENGATE_P0_CANARY_NOTIFY_ON=failure` sends a webhook only when the P0 suite
  fails.
- `TOKENGATE_P0_CANARY_NOTIFY_ON=always` sends success and failure status.
- `TOKENGATE_P0_CANARY_NOTIFY_ON=never` disables webhook delivery.

Triage failed canary runs in this order:

- Open `latest.log` and identify the first failed P0 check.
- If `/v1/models` fails, check backend availability, API key status, and group
  model visibility.
- If Chat Completions or Responses fails with `401` or `403`, re-test the
  upstream OpenAI account authorization and group assignment.
- If the model is visible but rejected by upstream, switch the canary model to
  the backend's supported P0 model and update public docs before launch.
- If only streaming fails, check SSE buffering, proxy timeouts, and
  `stream_options.include_usage=true` behavior.

The canary redacts TokenGate keys from logs and webhook payloads, but webhook
destinations should still be treated as operationally sensitive.

The older provider smoke remains useful for narrower Claude/OpenAI checks:

```bash
TOKENGATE_BASE_URL=https://<backend-domain> \
TOKENGATE_API_KEY=sk-... \
bash tools/tokengate_smoke_test.sh
```

## 3. CORS Check

Preflight should return `204` for the production frontend origin:

```bash
curl -i -X OPTIONS "https://<backend-domain>/api/v1/admin/accounts/1/test" \
  -H "Origin: https://<frontend-domain>" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: authorization,content-type"
```

Expected:

- `HTTP/2 204` or `HTTP/1.1 204`
- `Access-Control-Allow-Origin: https://<frontend-domain>`
- `Access-Control-Allow-Credentials: true`

If this fails, verify:

```env
CORS_ALLOWED_ORIGINS=https://<frontend-domain>
```

Do not use `*` in production.

## 3.1 Support Contact Check

The current production support contact should be:

```env
TOKENGATE_SUPPORT_CONTACT=bugfreev587@gmail.com
```

Admin -> Settings -> Contact Info takes precedence. If Contact Info is empty, the backend falls back to `TOKENGATE_SUPPORT_CONTACT`, then `TOKENGATE_CONTACT_INFO`.

Expected:

- `/api/v1/settings/public` returns `contact_info: "bugfreev587@gmail.com"`.
- `/support` shows the same contact.
- `tools/tokengate_launch_readiness.sh` passes with `TOKENGATE_EXPECTED_CONTACT_INFO=bugfreev587@gmail.com`.

## 4. SMTP Check

In the admin UI:

1. Open **Admin -> Settings -> Email**.
2. Configure SMTP host, port, username, password, sender email, sender name, and TLS.
3. Click **Test SMTP Connection**.
4. Send a test email to your own address.

Backend endpoints:

- `POST /api/v1/admin/settings/test-smtp`
- `POST /api/v1/admin/settings/send-test-email`

Both require admin authentication.

Expected:

- SMTP connection test succeeds
- test email arrives
- sender name is `TokenGate` or your configured brand

## 5. Password Reset Check

Prerequisites:

- SMTP is configured and test email succeeds
- email verification is enabled
- password reset is enabled
- frontend URL points to the Vercel frontend domain

In the user flow:

1. Open **Forgot Password**.
2. Submit an existing user email.
3. Confirm the reset email arrives.
4. Open the reset link.
5. Set a new password.
6. Sign in with the new password.

Expected:

- unknown emails still return the same generic success message
- reset token is one-time use
- reset link uses the frontend domain, not the Railway backend domain

## 6. Payment Test Mode Check

TokenGate V1 should support Stripe, Alipay, and WeChat Pay. Use Stripe first for international test-mode validation, then enable Alipay and WeChat Pay after the merchant credentials and callback paths are verified.

In admin:

1. Open **Admin -> Settings -> Payment**.
2. Enable payment.
3. Add a Stripe provider instance using test credentials.
4. Add Alipay and WeChat Pay provider instances using test credentials or sandbox-capable merchant credentials.
5. Configure visible payment methods so Alipay routes to the selected Alipay source and WeChat Pay routes to the selected WeChat source.
6. Create a small balance top-up order from the user Billing page.
7. Complete the payment in Stripe test mode first, then repeat for Alipay and WeChat Pay.
8. Confirm the order status becomes completed.
9. Confirm user balance increases.

Webhook URL:

```text
https://<backend-domain>/api/v1/payment/webhook/stripe
https://<backend-domain>/api/v1/payment/webhook/alipay
https://<backend-domain>/api/v1/payment/webhook/wxpay
```

Expected Stripe webhook events:

- `payment_intent.succeeded`
- `payment_intent.payment_failed`

Expected China payment behavior:

- Alipay returns a QR/cashier flow appropriate for desktop or mobile.
- WeChat Pay returns Native, H5, or JSAPI flow based on the client environment.
- Webhook signature verification succeeds before balance is credited.

## 7. Provider Account Check

For each upstream provider:

1. Create or connect the provider account.
2. Assign it to the `default` or target user group.
3. Run **Test Account Connection**.
4. Run a real API key request through the backend gateway.
5. Confirm usage and balance changes.

If **Test Account Connection** returns `405`, the frontend is probably calling the Vercel domain instead of the Railway backend. Verify:

```env
VITE_API_BASE_URL=https://<backend-domain>/api/v1
VITE_BUILD_TARGET=standalone
```

### OpenAI Provider Route Check

Before declaring OpenAI ready, confirm the test API key can actually see an OpenAI-compatible model:

```bash
TOKENGATE_BASE_URL=https://<backend-domain> \
TOKENGATE_API_KEY=<tokengate-api-key> \
TOKENGATE_REQUIRE_OPENAI_MODELS=1 \
tools/tokengate_model_visibility.sh
```

Expected:

- Claude models appear when the Claude Code OAuth account is routed correctly.
- OpenAI models such as `gpt-4.1-mini` appear only after an OpenAI account/provider is connected, assigned to the same group/channel as the test key, and allowed by model whitelist.

Then run:

```bash
TOKENGATE_BASE_URL=https://<backend-domain> \
TOKENGATE_API_KEY=<tokengate-api-key> \
TOKENGATE_RUN_CLAUDE=0 \
bash tools/tokengate_smoke_test.sh
```

If this returns `model route not found`, fix provider account health, group/channel routing, or model whitelist before debugging the frontend.

## 8. Database Backup And Restore Drill

Create a database backup before any public launch, pricing change, payment migration, or risky backend deploy:

```bash
DATABASE_URL="postgresql://..." \
TOKENGATE_BACKUP_DIR=backups \
tools/tokengate_backup_database.sh
```

Expected:

- a timestamped `backups/tokengate-*.dump` file is created
- the file size is greater than zero
- the backup file is stored somewhere durable outside the app container

Restore rehearsal should happen against a disposable staging database, never production:

```bash
RESTORE_DATABASE_URL="postgresql://staging-or-empty-db" \
pg_restore --clean --if-exists --no-owner --no-privileges \
  --dbname "$RESTORE_DATABASE_URL" backups/tokengate-YYYYMMDDTHHMMSSZ.dump
```

After restore:

- admin login works in the staging environment
- users, API keys, provider accounts, payment config, and usage records are visible
- no production webhook or email jobs are pointed at the restored staging database

## 9. Launch Blockers

Do not publicly launch while any of these are true:

- frontend points API calls to the Vercel domain
- CORS allows `*`
- SMTP is not configured
- password reset is not verified
- payment webhook is not verified
- only one provider has been tested
- usage logs do not match successful paid requests
- database backup/restore has not been rehearsed
