# TokenGate Deployment Notes

This directory contains deployment templates inherited from the Sub2API foundation and adapted for TokenGate. The current product deployment target is:

| Layer | Platform | Path |
| --- | --- | --- |
| Backend API | Railway | repository root |
| Frontend dashboard | Vercel | `frontend` |
| Database | Railway Postgres | managed service |
| Cache | Railway Redis | managed service |

For the live TokenGate service, use the root `Dockerfile` on Railway and the `frontend` sub-folder on Vercel. Docker Compose and systemd files in this directory are kept for self-hosting and migration compatibility.

## Railway Backend

Create these Railway components:

- TokenGate backend web service
- Railway Postgres
- Railway Redis

Deploy the backend service from the repository root. Railway should build the root `Dockerfile`.

Required backend environment variables:

```bash
DATABASE_URL="${{Postgres.DATABASE_URL}}"
REDIS_URL="${{Redis.REDIS_URL}}"
JWT_SECRET="replace_with_openssl_rand_hex_32"
TOTP_ENCRYPTION_KEY="replace_with_openssl_rand_hex_32"
RUN_MODE="standard"
SERVER_MODE="release"
LOG_SERVICE_NAME="tokengate"
LOG_ENV="production"
FRONTEND_URL="https://your-vercel-domain"
CORS_ALLOWED_ORIGINS="https://your-vercel-domain"
```

Generate secrets with:

```bash
openssl rand -hex 32
```

Recommended first deployment settings:

- Keep `AUTO_SETUP=true` for the first deploy so the backend can initialize database schema and admin credentials from environment variables.
- Set `ADMIN_EMAIL` and `ADMIN_PASSWORD` explicitly for the first production-style verification.
- Lock `CORS_ALLOWED_ORIGINS` to the Vercel frontend origin. Do not use `*` in production.

## Vercel Frontend

Use the `frontend` directory as the Vercel project root.

Recommended Vercel settings:

| Setting | Value |
| --- | --- |
| Framework preset | Vite |
| Root directory | `frontend` |
| Install command | `npm install` |
| Build command | `npm run build:standalone` |
| Output directory | `dist` |

Required frontend environment variable:

```bash
VITE_API_BASE_URL="https://your-railway-backend-domain/api/v1"
```

`VITE_API_BASE_URL` must point to the Railway backend API prefix and include `/api/v1`. Do not point it to the Vercel frontend domain.

## Environment Validation

Run the checker before deployment or after changing environment variables:

```bash
TOKENGATE_BACKEND_URL="https://your-railway-backend-domain" \
TOKENGATE_FRONTEND_URL="https://your-vercel-domain" \
DATABASE_URL="postgresql://example" \
REDIS_URL="redis://example" \
JWT_SECRET="$(openssl rand -hex 32)" \
TOTP_ENCRYPTION_KEY="$(openssl rand -hex 32)" \
FRONTEND_URL="https://your-vercel-domain" \
VITE_API_BASE_URL="https://your-railway-backend-domain/api/v1" \
tools/check_tokengate_env.sh
```

## Production Smoke Test

After deployment, create a TokenGate API key and run:

```bash
TOKENGATE_BASE_URL="https://your-railway-backend-domain" \
TOKENGATE_API_KEY="tg_live_or_sub2api_key" \
TOKENGATE_RUN_CLAUDE=1 \
TOKENGATE_RUN_OPENAI=1 \
tools/tokengate_smoke_test.sh
```

Set `TOKENGATE_RUN_OPENAI=0` or `TOKENGATE_RUN_CLAUDE=0` while a provider is not configured yet.

## Self-Hosting Notes

The files below are still useful for self-hosted deployments:

- `docker-compose.yml`: all-in-one Docker Compose with named volumes
- `docker-compose.local.yml`: Docker Compose with local data directories
- `docker-compose.standalone.yml`: app-only Compose file for external Postgres/Redis
- `config.example.yaml`: full backend configuration reference
- `sub2api.service`: legacy-compatible systemd unit
- `install.sh`: legacy-compatible binary installer

Some filenames still contain `sub2api` for runtime compatibility with the upstream project. Do not rename those files casually unless you also update the corresponding binary names, systemd service names, Docker entrypoints, and release automation.
