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
tools/tokengate_launch_readiness.sh
```

This verifies public pages plus protected SPA entry routes such as `/dashboard`, `/usage`, `/admin/accounts`, and `/admin/launch-readiness`.

Use `TOKENGATE_LAUNCH_PROFILE=private` for invite-only beta checks. Use `TOKENGATE_LAUNCH_PROFILE=public` before public self-serve launch; public mode fails if support contact, password reset, registration, or required payment settings are missing.

For gateway-only checks, run:

```bash
TOKENGATE_BASE_URL=https://<backend-domain> \
TOKENGATE_API_KEY=sk-... \
bash tools/tokengate_smoke_test.sh
```

Expected:

- Claude-compatible request returns `2xx`
- OpenAI-compatible request returns `2xx`
- API key `Last Used` changes
- **Usage** records appear after refresh
- dashboard totals update after refresh

If a provider is not configured yet:

```bash
TOKENGATE_RUN_OPENAI=0 bash tools/tokengate_smoke_test.sh
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
