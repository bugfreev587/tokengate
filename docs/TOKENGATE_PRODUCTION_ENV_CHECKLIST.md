# TokenGate Production Environment Checklist

Use this checklist before every public-facing deployment or environment change.

## Deployment Layout

| Layer | Platform | Project root |
| --- | --- | --- |
| Backend API | Railway | repository root |
| Frontend | Vercel | `frontend` |
| Database | Railway Postgres | managed component |
| Cache | Railway Redis | managed component |

Railway backend components:

- TokenGate backend web service
- Postgres
- Redis

## Railway Backend Variables

Required:

```env
DATABASE_URL="${{Postgres.DATABASE_URL}}"
REDIS_URL="${{Redis.REDIS_URL}}"
JWT_SECRET="<openssl rand -hex 32>"
TOTP_ENCRYPTION_KEY="<openssl rand -hex 32>"
RUN_MODE="standard"
SERVER_MODE="release"
SERVER_PORT="8080"
FRONTEND_URL="https://<vercel-frontend-domain>"
CORS_ALLOWED_ORIGINS="https://<vercel-frontend-domain>"
LOG_SERVICE_NAME="tokengate"
LOG_ENV="production"
```

Required on first setup or when resetting the bootstrap admin:

```env
AUTO_SETUP="true"
ADMIN_EMAIL="<owner-email>"
ADMIN_PASSWORD="<long-random-password>"
```

Recommended after setup:

- Rotate the bootstrap admin password in the UI.
- Keep `JWT_SECRET` and `TOTP_ENCRYPTION_KEY` stable across deploys.
- Never set `CORS_ALLOWED_ORIGINS="*"` in production.

## Vercel Frontend Settings

Vercel project settings:

| Setting | Value |
| --- | --- |
| Framework preset | Vite |
| Root directory | `frontend` |
| Install command | `npm install` |
| Build command | `npm run build:standalone` |
| Output directory | `dist` |

Required environment variable:

```env
VITE_API_BASE_URL="https://<railway-backend-domain>/api/v1"
```

Rules:

- `VITE_API_BASE_URL` must include `/api/v1`.
- `VITE_API_BASE_URL` must point to Railway, not the Vercel frontend domain.
- After changing it, redeploy the frontend.

## Admin Settings After Deploy

In **Admin -> Settings**:

- Set `Site Name` to `TokenGate` or the final brand.
- Set `Contact Info` so `/support` has a real support channel.
- Leave `Doc URL` empty to use `/docs`, or set an external docs URL intentionally.
- Configure SMTP before enabling email verification or password reset.
- Configure payment providers in test mode before enabling public top-ups.
- Configure public plans only after the first pricing model is approved.

## Verification Commands

Environment check:

```bash
AUTO_SETUP=true \
RUN_MODE=standard \
SERVER_PORT=8080 \
JWT_SECRET="<secret>" \
TOTP_ENCRYPTION_KEY="<secret>" \
ADMIN_EMAIL="<email>" \
ADMIN_PASSWORD="<secret>" \
DATABASE_URL="postgresql://..." \
REDIS_URL="redis://..." \
FRONTEND_URL="https://<frontend-domain>" \
CORS_ALLOWED_ORIGINS="https://<frontend-domain>" \
VITE_API_BASE_URL="https://<backend-domain>/api/v1" \
VITE_BUILD_TARGET=standalone \
bash tools/check_tokengate_env.sh all
```

Smoke test:

```bash
TOKENGATE_BASE_URL="https://<backend-domain>" \
TOKENGATE_API_KEY="sk-..." \
bash tools/tokengate_smoke_test.sh
```

Launch readiness check:

```bash
TOKENGATE_FRONTEND_URL="https://<frontend-domain>" \
TOKENGATE_BACKEND_URL="https://<backend-domain>" \
TOKENGATE_API_KEY="sk-..." \
tools/tokengate_launch_readiness.sh
```

By default this checks `/home`, `/docs`, `/pricing`, `/support`, `/login`, `/dashboard`, `/usage`, and `/admin/accounts` for SPA refresh compatibility.

Backup:

```bash
DATABASE_URL="postgresql://..." \
tools/tokengate_backup_database.sh
```

## Public Launch Blockers

Do not open public signup until all are true:

- Frontend `/home`, `/docs`, `/pricing`, `/support`, `/dashboard`, and `/usage` survive refresh.
- Claude and OpenAI account tests work from the admin UI.
- A real TokenGate API key request creates usage records and updates balance.
- SMTP test email and password reset both work.
- Payment test order and webhook both work, or payment is intentionally disabled for invite-only beta.
- Backup export and restore drill have been rehearsed against staging.
- Support contact is visible and correct.
