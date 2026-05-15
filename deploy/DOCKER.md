# TokenGate Docker Image

TokenGate is a subscription-native AI API gateway. The primary production target for this repository is Railway backend + Vercel frontend, but the backend can also run as a Docker container with external PostgreSQL and Redis.

## Quick Start

```bash
docker run -d \
  --name tokengate \
  -p 8080:8080 \
  -e DATABASE_URL="postgresql://user:pass@host:5432/tokengate" \
  -e REDIS_URL="redis://host:6379" \
  -e JWT_SECRET="$(openssl rand -hex 32)" \
  -e TOTP_ENCRYPTION_KEY="$(openssl rand -hex 32)" \
  -e FRONTEND_URL="https://your-frontend-domain" \
  -e CORS_ALLOWED_ORIGINS="https://your-frontend-domain" \
  ghcr.io/bugfreev587/tokengate:latest
```

If you build locally before publishing an image:

```bash
docker build -t tokengate:local .
docker run -d \
  --name tokengate \
  -p 8080:8080 \
  -e DATABASE_URL="postgresql://user:pass@host:5432/tokengate" \
  -e REDIS_URL="redis://host:6379" \
  -e JWT_SECRET="$(openssl rand -hex 32)" \
  -e TOTP_ENCRYPTION_KEY="$(openssl rand -hex 32)" \
  -e FRONTEND_URL="https://your-frontend-domain" \
  -e CORS_ALLOWED_ORIGINS="https://your-frontend-domain" \
  tokengate:local
```

## Required Environment Variables

| Variable | Description |
| --- | --- |
| `DATABASE_URL` | PostgreSQL connection string |
| `REDIS_URL` | Redis connection string |
| `JWT_SECRET` | Long random secret for JWT signing |
| `TOTP_ENCRYPTION_KEY` | Long random secret for TOTP encryption |
| `FRONTEND_URL` | Public frontend origin |
| `CORS_ALLOWED_ORIGINS` | Comma-separated allowed frontend origins |

Recommended production variables:

| Variable | Recommended |
| --- | --- |
| `RUN_MODE` | `standard` |
| `SERVER_MODE` | `release` |
| `LOG_SERVICE_NAME` | `tokengate` |
| `LOG_ENV` | `production` |
| `AUTO_SETUP` | `true` for the first deploy |
| `ADMIN_EMAIL` | Your admin email |
| `ADMIN_PASSWORD` | Long generated password |

## Notes

- Railway users should prefer the root `Dockerfile` and Railway-managed `DATABASE_URL` / `REDIS_URL`.
- Vercel frontend users must set `VITE_API_BASE_URL=https://your-backend-domain/api/v1`.
- Some internal runtime names still use `sub2api` for compatibility with the upstream foundation.

## Links

- [TokenGate repository](https://github.com/bugfreev587/tokengate)
- [Deployment checklist](../docs/TOKENGATE_DEPLOYMENT_CHECKLIST.md)
- [Operations runbook](../docs/TOKENGATE_OPERATIONS_RUNBOOK.md)
