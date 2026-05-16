# TokenGate Private Beta Acceptance Checklist

This checklist defines when TokenGate is ready for a controlled private beta with real users.

The goal is not perfection. The goal is a safe, supportable, measurable first release.

## 1. Deployment

- Production frontend is deployed on Vercel and returns `200` for `/`, `/home`, `/login`, `/register`, and `/key-usage`.
- Production backend is deployed on Railway and returns `200` for `/health`.
- Frontend `VITE_API_BASE_URL` points to the Railway backend API prefix and includes `/api/v1`.
- Backend `FRONTEND_URL` points to the Vercel frontend.
- Backend `CORS_ALLOWED_ORIGINS` includes only approved frontend origins.
- `DATABASE_URL` uses `${{Postgres.DATABASE_URL}}`.
- `REDIS_URL` uses `${{Redis.REDIS_URL}}`.
- `JWT_SECRET` and `TOTP_ENCRYPTION_KEY` are fixed production secrets generated with `openssl rand -hex 32`.

## 2. Admin Setup

- Admin can log in from the production frontend.
- Admin dashboard loads without console-level fatal errors.
- Admin can create, edit, enable, disable, and test upstream accounts.
- Admin can create at least one active user group.
- Admin can create at least one user-facing subscription plan.
- Admin can create at least one API key for a test user.

## 3. Gateway E2E

- Claude-compatible request succeeds through TokenGate with a TokenGate API key.
- OpenAI-compatible request succeeds through TokenGate with a TokenGate API key.
- Failed upstream requests surface actionable errors in the dashboard/logs.
- Usage logs are written after successful requests.
- Dashboard usage totals update after test traffic.
- API key last-used metadata updates after test traffic.

Recommended command:

```bash
TOKENGATE_FRONTEND_URL="https://<vercel-frontend-domain>" \
TOKENGATE_BACKEND_URL="https://<railway-backend-domain>" \
TOKENGATE_API_KEY="<token-gate-api-key>" \
TOKENGATE_EXPECTED_CONTACT_INFO="bugfreev587@gmail.com" \
tools/tokengate_private_beta_acceptance.sh
```

## 4. User Onboarding

- Registration page works, or registration is intentionally disabled with an alternative invite path.
- New user can log in and reach the user dashboard.
- New user can view available channels/plans.
- New user can create an API key if their group/plan allows it.
- `/key-usage` can query usage for a valid API key.
- Public docs link points to TokenGate docs.
- Contact/support information is visible in public settings or onboarding copy.

## 5. Email

- SMTP settings are configured in Admin -> Settings -> Email.
- SMTP connection test succeeds.
- Test email arrives in the target inbox.
- Password reset email arrives and reset link opens the Vercel frontend.
- Unknown password reset email requests return a generic success response.

## 6. Payment

- Payment is disabled for private testing unless a provider has been verified.
- If payment is enabled, provider credentials are test-mode credentials.
- Stripe or the selected provider can create a test checkout/order.
- Payment webhook endpoint receives provider events.
- Successful payment credits the correct user balance or subscription entitlement.
- Failed/cancelled payment does not credit balance.
- Admin can inspect payment order state.

## 7. Safety And Operations

- Railway logs do not show recurring CORS warnings.
- Backend logs do not expose secrets, provider tokens, or API keys.
- Rate limits are enabled for auth-sensitive endpoints.
- Backup process is documented with `tools/tokengate_backup_database.sh` and at least one backup export has been tested.
- Restore process is documented and rehearsed against a disposable staging database before public launch.
- Operations runbook is linked from the README.
- Known limitations are documented for beta users.

## 8. Private Beta Exit Criteria

TokenGate can leave private beta when:

- At least 5 real users complete onboarding without manual database intervention.
- At least 100 gateway requests complete successfully across supported providers.
- Usage accounting matches expected provider/token behavior for sampled requests.
- Payment flow, if enabled, is verified end to end in test mode and then live mode.
- Support/contact process has handled at least one real user issue.
- Backup and restore process has been rehearsed.

## Current Status

As of this checklist, TokenGate is ready for controlled self-use and production validation.

It should not be opened broadly until payment, email, backup/restore, and final public pricing have been verified in the live deployment.
