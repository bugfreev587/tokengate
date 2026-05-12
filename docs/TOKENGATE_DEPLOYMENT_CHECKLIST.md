# TokenGate First Deployment Checklist

This checklist is for the first real-environment verification of TokenGate as a public-facing product.

## Goal

The goal of this deployment is not a public launch. It is a controlled verification that the product works end to end on:

- `frontend` on Vercel
- `backend` on Railway
- PostgreSQL on Railway
- Redis on Railway

## Recommended Push Point

Push now when you want to verify:

- TokenGate branding and landing page messaging
- Login and registration
- Dashboard, API keys, usage, plans, and billing
- Admin settings information architecture
- Public pricing and balance language

## Frontend Build Modes

The frontend now supports two build targets:

- `npm run build:standalone`
  Use this for Vercel or any standalone static hosting.
- `npm run build:embedded`
  Use this when the backend should serve the built frontend bundle.

The default `npm run build` still works, but for deployment it is better to choose the explicit target.

## Vercel Setup

Use the `frontend` directory as the Vercel project root.

Recommended settings:

- Framework preset: `Vite`
- Root directory: `frontend`
- Install command: `npm install`
- Build command: `npm run build:standalone`
- Output directory: `dist`

Required environment variables:

- `VITE_API_BASE_URL=https://<your-railway-backend-domain>/api/v1`

Optional environment variables:

- `VITE_BUILD_TARGET=standalone`

## Railway Setup

Use the repository root for the backend service.

Before deploying, make sure Railway provides:

- one web service for the Go backend
- one PostgreSQL instance
- one Redis instance

Recommended backend build path:

- Use the repository root `Dockerfile`
- Let Railway build the backend container from source
- Keep `AUTO_SETUP=true` for the first deployment so the service can initialize itself from environment variables instead of waiting for the setup wizard

Minimum backend environment variables to verify the first deployment:

- `SERVER_PORT`
- `AUTO_SETUP=true`
- `RUN_MODE=standard`
- `JWT_SECRET`
- `TOTP_ENCRYPTION_KEY`
- `ADMIN_EMAIL`
- `ADMIN_PASSWORD`
- `DATABASE_URL`
- `REDIS_URL`

If you want the first deployment to behave like the public product, keep `RUN_MODE=standard`.

Notes:

- The backend now supports Railway-style `DATABASE_URL` and `REDIS_URL` directly.
- If you prefer more explicit configuration, the split `DATABASE_*` and `REDIS_*` variables still work too.
- `ADMIN_PASSWORD` can be left empty, but for the first public-style verification it is better to set it explicitly.

## First Verification Path

After deployment, verify these flows in order:

1. Open the landing page and confirm TokenGate branding appears correctly.
2. Open login and registration pages and confirm they point to the Railway API.
3. Register or sign in with the admin account.
4. Open user dashboard, API keys, usage, plans, and billing.
5. Open admin settings and confirm the new tab order and guide cards render correctly.
6. Confirm public settings such as site name and subtitle flow through to the frontend.

## Not A Launch Blocker

These are still acceptable for the first verification deployment:

- Vite chunk-size warnings
- existing dynamic import warnings during frontend build

They should be cleaned up later, but they do not block the first online test.
