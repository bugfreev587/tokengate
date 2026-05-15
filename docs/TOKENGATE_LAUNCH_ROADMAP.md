# TokenGate Launch Roadmap

This roadmap tracks the work needed to move TokenGate from the current online V1 into a public-ready service.

## Current Stage

TokenGate is currently in **online private beta**.

Completed:

- Vercel standalone frontend deployment
- Railway backend deployment
- Railway Postgres and Redis integration
- API key creation
- Claude gateway request path
- usage logging for a successful model request
- first pass TokenGate branding and product narrative
- split frontend/backend API base URL fixes
- production smoke test script
- deployment environment checker
- operations runbook
- private beta acceptance checklist
- public onboarding contract
- public payment documentation cleanup
- API Keys and Usage onboarding empty states
- production smoke test preflight and HTTP diagnostics
- public `/docs` developer onboarding page
- backend-mode access guard for public docs

Current launch readiness estimate: **87%**

## Phase 1: Core Runtime Verification

Goal: prove that the product works repeatedly in production, not only once.

Required before moving on:

- Claude account test succeeds from the admin UI
- OpenAI account test succeeds from the admin UI
- Claude API key request succeeds and creates a usage log
- OpenAI API key request succeeds and creates a usage log
- `Last Used`, dashboard totals, and usage records update after successful requests
- failed upstream requests produce actionable errors and do not create misleading charges
- Vercel routes such as `/home`, `/dashboard`, `/admin/accounts`, and `/usage` survive refresh

Status: **in progress**

## Phase 2: Billing And Plan Semantics

Goal: make pricing understandable, auditable, and margin-safe.

Required:

- define default public plan names and limits
- define included balance rules
- define pay-as-you-go behavior after included balance is exhausted
- define text pricing as per-1M input/output tokens
- define image pricing as per image or provider-native output unit
- define video pricing as per job, second, or provider-native unit
- make user billing pages explain balance, included usage, and usage charges consistently
- make admin plan creation match the public pricing model
- document the deduction order for included balance, bonus balance, and prepaid balance

Reference: [TOKENGATE_BILLING_MODEL.md](TOKENGATE_BILLING_MODEL.md)

Status: **partially complete**

## Phase 3: Operational Readiness

Goal: make TokenGate safe to run for real users.

Reference: [TOKENGATE_PRIVATE_BETA_ACCEPTANCE.md](TOKENGATE_PRIVATE_BETA_ACCEPTANCE.md)

Required:

- production domain connected to Vercel
- production backend domain connected to Railway or a stable API subdomain
- CORS locked to production frontend domains
- email sending configured and verified
- password reset verified
- payment provider selected and verified in test mode
- payment webhook verified end to end
- basic abuse controls configured
- rate limits reviewed for public signup and API key usage
- admin alerting path for failed upstream accounts
- database backup and restore drill documented
- Railway and Vercel environment variables documented in one final checklist
- Railway and Vercel environment variables validated with `tools/check_tokengate_env.sh`
- production operations checks documented in [TOKENGATE_OPERATIONS_RUNBOOK.md](TOKENGATE_OPERATIONS_RUNBOOK.md)

Status: **early**

## Phase 4: Public Onboarding

Goal: let a new developer understand and use TokenGate without hand-holding.

Required:

- public docs page or hosted documentation
- quickstart with API key creation
- OpenAI-compatible request example
- Anthropic-compatible request example
- model pricing page
- FAQ for credits vs tokens vs balance
- onboarding checklist after signup
- admin checklist after first deployment
- support/contact route
- in-app empty states that guide users back to API key creation and quickstart docs

Reference: [TOKENGATE_PUBLIC_ONBOARDING.md](TOKENGATE_PUBLIC_ONBOARDING.md)

Status: **mostly complete**

## Phase 5: Product Differentiation

Goal: make TokenGate more than a rebranded gateway.

Candidates:

- provider routing policy UI
- fallback rules by model, group, or error type
- margin-aware pricing guardrails
- product-level usage analytics
- hosted subscription backend for small AI SaaS products
- team/workspace support
- developer docs for integrating external product frontends

Status: **not started**

## Decision Gates

These items need owner decisions before public launch:

- official production domain and API domain
- default currency
- default payment provider
- first public plan lineup
- whether V1 is self-serve public signup or invite-only beta
- whether to support China payment rails in V1
- whether to expose video generation pricing in V1 or keep it internal first

## Next Execution Order

1. Finish account test verification for Claude and OpenAI.
2. Verify successful OpenAI usage logging.
3. Make backend default branding fully TokenGate.
4. Lock down public environment configuration.
5. Define the first plan and billing model in admin settings.
6. Verify email, password reset, and payment test mode.
7. Run the private beta acceptance checklist before inviting users.
