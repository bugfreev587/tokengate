# TokenGate Public Onboarding

This document defines the first-run experience for a public TokenGate user. It is written as product guidance first and implementation QA second.

## Target User

The initial public user is a developer or indie product builder who wants a stable AI API key without directly managing every upstream account, subscription, quota, and billing rule.

## First Five Minutes

The first successful session should feel like this:

1. User lands on the public homepage and understands that TokenGate is an AI API gateway with metered usage.
2. User signs up or signs in.
3. User opens **API Keys** and creates a key.
4. User copies one curl example from the dashboard or docs.
5. User receives a successful model response and sees the usage record.

If a user cannot complete those five steps without asking support, onboarding is not ready.

## Required Public Pages

These pages should be visible or discoverable before public launch:

- Landing page with one-line positioning, supported models/providers, and clear call to action.
- Pricing page explaining balance, tokens, image/video units, and plan limits.
- Quickstart page with API key creation and copy-paste curl examples.
- Model pricing page showing per-1M-token text pricing and non-text unit pricing.
- Account settings page for email, password, MFA, and connected identities.
- Billing page for balance, top-ups, invoices/orders, and current plan.
- Usage page with request history, model, token count, cost, status, and timestamp.
- Support/contact page or clear contact block in settings and error states.

## API Quickstart Contract

Every public quickstart must include:

- A backend base URL, not the Vercel frontend URL.
- A bearer API key example.
- One OpenAI-compatible request to `/v1/chat/completions`.
- One Anthropic-compatible request to `/v1/messages`.
- The expected success indicators in the UI: `Last Used`, **Usage**, dashboard totals, and balance deduction.
- A short troubleshooting section for `401`, `403`, `404`, `405`, and upstream account errors.

Reference implementation: [TOKENGATE_QUICKSTART.md](TOKENGATE_QUICKSTART.md).

## Billing Language

Use these product terms consistently:

- **Tokens** are provider/model usage units for text models.
- **Image units** are image-generation units, usually per image or provider-native output unit.
- **Video units** are video-generation units, usually per job, second, or provider-native unit.
- **Balance** is the money-like account value that usage settles against.
- **Included usage** is plan allowance that may be consumed before prepaid balance.

Avoid calling the public wallet "credits" unless the pricing model intentionally defines credits as a separate abstraction. For TokenGate V1, the safer public language is **balance + metered usage**.

## Empty States

Empty states should teach the next action:

- No API keys: show a create-key action and the first curl example.
- No usage: explain that usage appears after a successful request.
- No balance: explain whether trial, included usage, top-up, or admin grant is needed.
- No payment methods: tell the user payments are not available yet or contact support.
- No upstream account available: tell the admin which provider/group needs configuration.

## Error Messaging

Public errors should avoid leaking internal implementation details but still be actionable:

| Error | User-Facing Meaning | Next Action |
| --- | --- | --- |
| `401` | API key missing or invalid | Create or rotate an API key |
| `403` | Key is valid but not allowed | Check account status, group, or plan |
| `404` | Endpoint or route is wrong | Use backend API base URL and documented path |
| `405` | Browser/frontend URL used as API URL | Point SDK/curl to the backend domain |
| `429` | Rate or quota limit reached | Wait, upgrade, or contact support |
| Upstream unavailable | Provider account has an issue | Try another model later or contact support |

## Admin Launch Checklist

Before inviting public users, an admin should verify:

- `FRONTEND_URL` and `CORS_ALLOWED_ORIGINS` match the production Vercel domain.
- `VITE_API_BASE_URL` points to the Railway backend and includes `/api/v1`.
- Public settings show TokenGate branding and a real support contact.
- At least one Claude-compatible and one OpenAI-compatible account test succeeds.
- At least one API key request succeeds for each enabled provider family.
- Usage records show model, status, token count, and cost after successful requests.
- Payment test mode succeeds, including webhook confirmation and balance update.
- Password reset email succeeds from production.
- Database backup and restore have been rehearsed.

## Definition Of Done

Public onboarding is ready when a new user can sign up, create an API key, send a request, understand the charge, and know how to get support without direct founder intervention.
