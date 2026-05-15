# TokenGate

TokenGate is a subscription-native AI API gateway. It turns upstream AI product subscriptions and provider accounts into managed API access with user auth, API keys, usage tracking, billing controls, and an admin dashboard.

This repository is currently optimized for a split production deployment:

| Layer | Platform | Sub-folder |
| --- | --- | --- |
| Backend API | Railway | repository root |
| Frontend dashboard | Vercel | `frontend` |
| Database | Railway Postgres | managed service |
| Cache | Railway Redis | managed service |

TokenGate is built on top of the open-source Sub2API foundation and is being productized as an independent public service.

## What It Does

- Provides OpenAI-compatible and Anthropic-compatible gateway endpoints.
- Lets admins connect upstream provider accounts such as OpenAI and Claude.
- Lets users create TokenGate API keys for downstream apps.
- Tracks usage and cost at token/model level.
- Supports SaaS-style account management, balance checks, plans, payments, and admin operations.
- Includes production deployment docs, smoke tests, and operations runbooks.

## Quick Links

- [Quickstart](docs/TOKENGATE_QUICKSTART.md)
- [Public onboarding](docs/TOKENGATE_PUBLIC_ONBOARDING.md)
- [Billing model](docs/TOKENGATE_BILLING_MODEL.md)
- [Deployment checklist](docs/TOKENGATE_DEPLOYMENT_CHECKLIST.md)
- [Operations runbook](docs/TOKENGATE_OPERATIONS_RUNBOOK.md)
- [Private beta acceptance checklist](docs/TOKENGATE_PRIVATE_BETA_ACCEPTANCE.md)
- [Launch roadmap](docs/TOKENGATE_LAUNCH_ROADMAP.md)
- [Product strategy](docs/TOKENGATE_PRODUCT_STRATEGY.md)

## Deployment

### Backend on Railway

Deploy from the repository root. Railway should build the root `Dockerfile`.

Required Railway components:

- TokenGate backend service
- PostgreSQL service
- Redis service

Core backend environment variables:

```bash
DATABASE_URL="${{Postgres.DATABASE_URL}}"
REDIS_URL="${{Redis.REDIS_URL}}"
JWT_SECRET="replace_with_openssl_rand_hex_32"
TOTP_ENCRYPTION_KEY="replace_with_openssl_rand_hex_32"
RUN_MODE="standard"
SERVER_MODE="release"
LOG_SERVICE_NAME="tokengate"
LOG_ENV="production"
FRONTEND_URL="https://your-frontend-domain"
CORS_ALLOWED_ORIGINS="https://your-frontend-domain"
```

Generate production secrets with:

```bash
openssl rand -hex 32
```

### Frontend on Vercel

Deploy from the `frontend` sub-folder.

Required Vercel environment variable:

```bash
VITE_API_BASE_URL="https://your-railway-backend-domain/api/v1"
```

Do not point `VITE_API_BASE_URL` at the Vercel frontend domain. It must point to the Railway backend API prefix and include `/api/v1`.

## Verification

Run the production environment checker before deploying or after changing env vars:

```bash
TOKENGATE_BACKEND_URL="https://your-railway-backend-domain" \
TOKENGATE_FRONTEND_URL="https://your-frontend-domain" \
DATABASE_URL="postgresql://example" \
REDIS_URL="redis://example" \
JWT_SECRET="$(openssl rand -hex 32)" \
TOTP_ENCRYPTION_KEY="$(openssl rand -hex 32)" \
FRONTEND_URL="https://your-frontend-domain" \
VITE_API_BASE_URL="https://your-railway-backend-domain/api/v1" \
tools/check_tokengate_env.sh
```

Run a live gateway smoke test after creating a TokenGate API key:

```bash
TOKENGATE_BASE_URL="https://your-railway-backend-domain" \
TOKENGATE_API_KEY="tg_live_or_sub2api_key" \
TOKENGATE_RUN_CLAUDE=1 \
TOKENGATE_RUN_OPENAI=1 \
tools/tokengate_smoke_test.sh
```

## Local Development

Backend:

```bash
cd backend
go test ./...
```

Frontend:

```bash
cd frontend
npm install
npm run build:standalone
```

## Current Product Status

TokenGate is usable for private production validation. The next public-launch phase is focused on payment-provider test mode, SMTP/password reset verification, backup/restore rehearsal, final pricing plans, and complete public onboarding docs.
