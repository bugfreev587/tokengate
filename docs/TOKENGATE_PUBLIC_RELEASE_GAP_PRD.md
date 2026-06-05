# TokenGate Public Release Gap PRD

Date: 2026-06-05

Status: Draft for owner review

## 1. Executive Summary

TokenGate is not ready for broad public self-serve release yet.

The current production deployment is healthy enough for controlled validation:

- Frontend routes on `https://tokengate-psi.vercel.app` return the SPA shell for `/home`, `/docs`, `/pricing`, `/support`, `/login`, `/dashboard`, `/usage`, `/admin/accounts`, and `/admin/launch-readiness`.
- Backend `GET https://tokengate-production.up.railway.app/health` returns `{"status":"ok"}`.
- Backend public settings are reachable at `GET /api/v1/settings/public`.
- Backend OpenAI-compatible route `POST /v1/chat/completions` exists and returns a correct `401 API_KEY_REQUIRED` response when no API key is supplied.
- Frontend typecheck, production build, and selected launch/payment/settings tests passed locally.

However, public self-serve release still has hard blockers:

- Security scans are red in GitHub Actions and locally.
- Frontend lint is red on the public API reference page.
- Backend unit tests are red locally on macOS path normalization.
- Password reset and payment are disabled in live public settings.
- Email verification is disabled for self-serve signup.
- Legal agreement documents are present but empty.
- OpenAI/ChatGPT usage has not been proven end-to-end with a live TokenGate API key.
- Official app/API domains are not finalized; current app and API still use Vercel/Railway service domains, while `tokengate.to` points to a separate Vercel project/repo.
- Release automation, CLA, Docker, and some docs still carry Sub2API branding and package names.

Release recommendation: keep TokenGate in private beta or invite-only beta until all P0 blockers in this PRD are closed and verified.

## 2. Product Goal

Enable a public developer or indie SaaS builder to sign up, create a TokenGate API key, call Claude-compatible and ChatGPT/OpenAI-compatible APIs, see usage and cost, recover their account, understand pricing/legal terms, and pay or receive entitlement without founder intervention.

## 3. Non-Goals

- Building every advanced admin analytics panel before public release.
- Removing every internal `sub2api` module import before public release.
- Enabling every provider family in V1.
- Launching video/image generation publicly unless payment, pricing, and provider routes are verified separately.

## 4. Target Release Profiles

### 4.1 Invite-Only Public Beta

Acceptable if:

- Self-serve registration can be disabled or gated by invite.
- Payment can remain disabled if users are manually granted balance/plan access.
- Support is explicitly founder-led.
- Claude and ChatGPT routes are both verified with real API key usage and billing records.

### 4.2 Self-Serve Public Release

Required if:

- `registration_enabled=true`.
- Email verification is enabled or an equivalent anti-abuse control is live.
- Password reset is enabled and verified.
- Payment or plan entitlement is enabled and verified, or public copy clearly says access is approval-based.
- Legal terms are non-empty and accepted during signup.

## 5. Audit Evidence

### 5.1 Production Readiness Probe

Command:

```bash
TOKENGATE_FRONTEND_URL=https://tokengate-psi.vercel.app \
TOKENGATE_BACKEND_URL=https://tokengate-production.up.railway.app \
TOKENGATE_LAUNCH_PROFILE=public \
TOKENGATE_SIGNUP_MODE=self_serve \
TOKENGATE_REQUIRE_PAYMENT=1 \
TOKENGATE_EXPECTED_CONTACT_INFO=bugfreev587@gmail.com \
TOKENGATE_RUN_API_SMOKE=0 \
tools/tokengate_launch_readiness.sh
```

Observed result:

- Frontend route refresh checks: pass.
- Backend public settings: pass.
- CORS preflight: pass.
- `password_reset_enabled=false`: fail.
- `payment_enabled=false`: fail.
- `email_verify_enabled=false`: warn.
- The script also reported `registration_enabled=false`, but direct settings inspection shows `registration_enabled=true`; see GAP-P1-003.

Direct public settings excerpt:

```json
{
  "registration_enabled": true,
  "email_verify_enabled": false,
  "password_reset_enabled": false,
  "turnstile_enabled": false,
  "site_name": "TokenGate",
  "contact_info": "bugfreev587@gmail.com",
  "payment_enabled": false,
  "login_agreement_documents": [
    {"id": "terms", "title": "服务条款", "content_md": ""},
    {"id": "usage-policy", "title": "使用政策", "content_md": ""},
    {"id": "supported-regions", "title": "支持的国家和地区", "content_md": ""},
    {"id": "service-specific-terms", "title": "服务特定条款", "content_md": ""}
  ]
}
```

### 5.2 API Probe

Command:

```bash
curl -i -X POST https://tokengate-production.up.railway.app/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{"model":"gpt-5.2-chat-latest","messages":[{"role":"user","content":"hello"}]}'
```

Observed result:

```json
{
  "code": "API_KEY_REQUIRED",
  "message": "API key is required in Authorization header (Bearer scheme), x-api-key header, or x-goog-api-key header"
}
```

Interpretation: the route is reachable and protected, but no live ChatGPT/OpenAI success path was verified because no production `TOKENGATE_API_KEY` was available in the local environment.

### 5.3 Vercel Deployment

Connected Vercel project:

- Project: `tokengate`
- Latest production deployment: `READY`
- Production URL: `tokengate-28r8y280a-xiaobo-yus-projects.vercel.app`
- Aliases: `tokengate-psi.vercel.app`, `tokengate-xiaobo-yus-projects.vercel.app`, `tokengate-git-main-xiaobo-yus-projects.vercel.app`
- Current custom app domain: none found in this Vercel project.

Separate Vercel project:

- `tokengate.to` exists and returns a marketing page titled `TokenGate - Control Your AI Spend`.
- That project is linked to another repo history (`burnrate.ai` in deployment metadata), not the current `tokengate` app repo.

### 5.4 Local Verification

Passed:

```bash
GOCACHE=/private/tmp/tokengate-gocache CGO_ENABLED=0 go build -ldflags='-s -w' -trimpath -o /private/tmp/tokengate-server ./cmd/server
COREPACK_ENABLE_PROJECT_SPEC=0 corepack pnpm exec vue-tsc --noEmit
COREPACK_ENABLE_PROJECT_SPEC=0 corepack pnpm exec vite build
COREPACK_ENABLE_PROJECT_SPEC=0 corepack pnpm exec vitest run \
  src/router/__tests__/launch-readiness-route.spec.ts \
  src/api/__tests__/payment.spec.ts \
  src/views/admin/__tests__/SettingsView.spec.ts
```

Failed:

```bash
COREPACK_ENABLE_PROJECT_SPEC=0 corepack pnpm run lint:check
GOCACHE=/private/tmp/tokengate-gocache go test -tags=unit ./...
COREPACK_ENABLE_PROJECT_SPEC=0 corepack pnpm audit --prod --audit-level=high
GOCACHE=/private/tmp/tokengate-gocache GOMODCACHE=/private/tmp/tokengate-gomodcache \
  go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

### 5.5 GitHub Actions

Recent main branch runs:

- `Security Scan`, scheduled 2026-06-01: failure.
  - `backend-security`: failed at `govulncheck`.
  - `frontend-security`: failed at audit exception check.
- `CI`, push 2026-05-17: failure.
  - Backend unit/integration and golangci-lint jobs passed.
  - Frontend job failed at ESLint `vue/no-dupe-keys` in `frontend/src/components/docs/api/ApiReferencePage.vue`.

## 6. Gap Register

### GAP-P0-001: Backend Security Scan Is Red

Priority: P0

Area: Security, backend runtime

Evidence:

- Local `govulncheck ./...` found reachable vulnerabilities:
  - `GO-2026-5039` in standard library `net/textproto`, fixed in Go 1.26.4.
  - `GO-2026-5037` in standard library `crypto/x509`, fixed in Go 1.26.4.
  - `GO-2026-5026` in `golang.org/x/net/idna`, fixed in `golang.org/x/net` v0.55.0.
- GitHub Actions `Security Scan` backend job failed at `Run govulncheck`.
- `backend/go.mod` currently declares `go 1.26.3`.

Impact:

- Public launch would ship with known reachable vulnerabilities in request, SMTP, TLS, and IDNA-related call paths.
- Security scan cannot be used as a public release gate until it is green.

Requirements:

- Upgrade Go toolchain to 1.26.4 or newer across `backend/go.mod`, CI, release workflows, local docs, Docker build images, and deployment images.
- Upgrade `golang.org/x/net` to v0.55.0 or newer.
- Re-run `govulncheck ./...` in CI and locally.

Acceptance Criteria:

- `govulncheck ./...` exits 0.
- GitHub Actions `Security Scan / backend-security` is green on `main`.
- Release workflow uses the same fixed Go version.

### GAP-P0-002: Frontend Security Scan Is Red

Priority: P0

Area: Security, frontend dependencies

Evidence:

- `pnpm audit --prod --audit-level=high` found 3 high vulnerabilities:
  - `xlsx`: `GHSA-4r6h-8v6p-xvw6`.
  - `xlsx`: `GHSA-5pgg-2g8v-p4x9`.
  - `js-cookie`: `GHSA-qjx8-664m-686j`.
- `tools/check_pnpm_audit_exceptions.py` fails because `js-cookie` is missing from `.github/audit-exceptions.yml`.
- Existing audit exceptions have placeholder owner `security@your-domain`.

Impact:

- Frontend security gate is red.
- A new high advisory is not risk-accepted or remediated.
- Existing exceptions do not have a real accountable owner.

Requirements:

- Upgrade the dependency chain that pulls `js-cookie <=3.0.5`, or add a time-bound exception with real owner, exploitability assessment, and mitigation.
- Reassess `xlsx`; replace SheetJS if no patched npm package is available and admin export can be implemented through a safer library or server-generated CSV/XLSX.
- Replace placeholder security owner with a real TokenGate owner alias.

Acceptance Criteria:

- `pnpm audit --prod --audit-level=high` has no unhandled high/critical advisory.
- `tools/check_pnpm_audit_exceptions.py` exits 0.
- GitHub Actions `Security Scan / frontend-security` is green on `main`.

### GAP-P0-003: Main CI Is Not Green

Priority: P0

Area: Release gate, CI

Evidence:

- Latest inspected `CI` run on main failed in frontend lint.
- Local `pnpm run lint:check` reproduces:

```text
frontend/src/components/docs/api/ApiReferencePage.vue
86:7 error Duplicate key 'endpoint'. May cause name collision in script or template tag vue/no-dupe-keys
```

- Local backend unit tests fail on `TestResolvePageImagePath` because macOS resolves temp paths as `/private/var/...` while the test expects `/var/...`.

Impact:

- Public release cannot be cut from a red CI baseline.
- The failing frontend file is part of public API docs, which are central to onboarding.

Requirements:

- Rename the computed `endpoint` or the prop in `ApiReferencePage.vue` to remove the Vue duplicate key.
- Normalize symlinked temp paths or compare cleaned/evaluated paths in `page_handler_test.go`.
- Run the full CI-equivalent suite after fixes.

Acceptance Criteria:

- `pnpm run lint:check` exits 0.
- `go test -tags=unit ./...` exits 0 on macOS and CI Linux.
- GitHub Actions `CI` is green on `main`.

### GAP-P0-004: Password Reset Is Disabled In Production

Priority: P0 for self-serve public release

Area: Auth, support load, account recovery

Evidence:

- Live public settings return `password_reset_enabled=false`.
- Existing private beta checklist explicitly says password reset email must arrive and reset link must open the Vercel frontend.

Impact:

- Public users cannot recover accounts without founder intervention.
- Self-serve launch is not supportable.

Requirements:

- Configure SMTP in production.
- Enable password reset only after SMTP test succeeds.
- Verify unknown email reset requests return generic success.
- Verify reset link points to the final frontend domain, not Railway.

Acceptance Criteria:

- `password_reset_enabled=true` in public settings for self-serve release.
- Reset email arrives in a controlled production test.
- Password can be changed through the reset flow.

### GAP-P0-005: Payment Is Disabled And Payment Webhooks Are Not Public-Verified

Priority: P0 if paid self-serve; P1 if invite-only beta with manual grants

Area: Billing, monetization, entitlement

Evidence:

- Live public settings return `payment_enabled=false`.
- Existing launch docs require payment test order and webhook verification before public launch.
- Payment docs target Stripe + Alipay + WeChat Pay, but no current live test evidence was found in this audit.

Impact:

- A public user cannot self-serve paid top-up or subscription purchase.
- If payment copy is public while payment is disabled, users may hit a dead-end.
- If payment is enabled without verified webhook behavior, balance/subscription crediting risk is high.

Requirements:

- Decide public V1 payment model:
  - Self-serve paid launch: enable payment providers in test mode first, then live.
  - Invite-only launch: keep payment disabled and clearly state that access is manually approved.
- Verify provider order creation, webhook receipt, successful crediting, failed/cancelled non-crediting, and admin order visibility.
- Confirm public pricing matches admin plan configuration.

Acceptance Criteria:

- For self-serve: `payment_enabled=true`, at least one provider verified, webhook verified, balance/subscription updates proven.
- For invite-only: public UI/copy states payment is disabled and directs users to support.

### GAP-P0-006: ChatGPT/OpenAI Route Is Not Proven End-To-End In Production

Priority: P0

Area: Gateway, Citeloop integration, public API contract

Evidence:

- Route exists and returns `401 API_KEY_REQUIRED` without a key.
- No `TOKENGATE_API_KEY` was present locally for a real smoke.
- `TOKENGATE_LAUNCH_ROADMAP.md` still records that OpenAI-compatible `/v1/chat/completions` returned 404 for `gpt-4.1-mini` in prior production smoke notes.
- Model visibility script can require OpenAI models, but it was not run with a production key in this audit.

Impact:

- Public docs promise ChatGPT/OpenAI-compatible access, but production success is not verified.
- CiteLoop or other customers cannot be considered ready to rely on ChatGPT through TokenGate until the success path is proven.

Requirements:

- Create or reuse a production TokenGate API key in an OpenAI-enabled user group.
- Assign at least one healthy OpenAI upstream account to that group.
- Confirm `TOKENGATE_REQUIRE_OPENAI_MODELS=1 tools/tokengate_model_visibility.sh` passes.
- Send a real `/v1/chat/completions` request using the documented default model.
- Confirm usage log, API key last-used, token totals, cost, and balance/plan deduction.

Acceptance Criteria:

- `POST /v1/chat/completions` returns HTTP 200 for a production key.
- `/v1/models` shows at least one OpenAI-compatible model for that key.
- `/v1/usage` includes the successful request with model, tokens, cost, and timestamp.
- Dashboard totals and balance update after refresh.

### GAP-P0-007: Legal Terms Are Empty For Public Signup

Priority: P0 for self-serve public release

Area: Legal, compliance, onboarding

Evidence:

- Public settings include four legal/login agreement documents, all with `content_md=""`.
- `login_agreement_enabled=false`.

Impact:

- Public users can sign up without accepting meaningful terms, usage policy, supported regions, or service-specific terms.
- Payment and API usage without terms creates avoidable legal/support risk.

Requirements:

- Draft and approve public legal content:
  - Terms of Service.
  - Acceptable Use / Usage Policy.
  - Supported Regions.
  - Service-specific AI/API terms.
  - Privacy policy if personal data, email, payments, logs, or analytics are stored.
- Enable login/signup agreement flow if self-serve registration is open.
- Link legal documents from footer, signup, support, and billing.

Acceptance Criteria:

- Legal document pages render non-empty content.
- Signup requires agreement acceptance when self-serve is enabled.
- Agreement revision/date is visible and stored.

### GAP-P1-001: Email Verification And Anti-Abuse Controls Are Disabled

Priority: P1; P0 if open signup is broad public

Area: Auth, abuse prevention

Evidence:

- `email_verify_enabled=false`.
- `turnstile_enabled=false`.
- `registration_enabled=true`.
- `risk_control_enabled=false`.

Impact:

- Self-serve signup can create unverified accounts.
- Gateway abuse can be attempted before operational controls are proven.

Requirements:

- Decide whether public signup is self-serve or invite-only.
- For self-serve, enable email verification and at least one bot/abuse control.
- Verify registration email delivery and resend behavior.
- Review API key creation limits and new-user default group/plan limits.

Acceptance Criteria:

- New public users must verify email before API key usage, or a documented alternative control is active.
- Abuse controls are tested with rate-limit and bot-challenge scenarios.

### GAP-P1-002: Official Public Domains Are Fragmented

Priority: P1

Area: Deployment, brand, customer trust

Evidence:

- Current app project aliases are `tokengate-psi.vercel.app` and related Vercel domains.
- Backend API is `tokengate-production.up.railway.app`.
- `tokengate.to` and `app.tokengate.to` return a separate marketing page from a separate Vercel project/repo.
- Existing roadmap still lists official production domain and API domain as decision gates.

Impact:

- Customers may confuse frontend dashboard URL, marketing URL, and API base URL.
- Docs repeatedly warn not to use the Vercel frontend domain as API base URL, which becomes easier to get wrong when domains are inconsistent.
- Public trust is weaker without branded app/API domains.

Requirements:

- Choose final domain layout, for example:
  - Marketing: `https://tokengate.to`
  - App/dashboard/docs: `https://app.tokengate.to`
  - API: `https://api.tokengate.to`
- Connect current app Vercel project to the chosen app domain.
- Connect Railway backend to the chosen API domain.
- Update `FRONTEND_URL`, `CORS_ALLOWED_ORIGINS`, `VITE_API_BASE_URL`, password reset links, docs examples, smoke scripts, and support copy.

Acceptance Criteria:

- Public docs and quickstart use the final API domain.
- Dashboard and auth flows use the final app domain.
- `tools/tokengate_launch_readiness.sh` passes against final domains.

### GAP-P1-003: Public Readiness Script Has A Cross-Platform Boolean Parser Bug

Priority: P1

Area: QA tooling, release gate

Evidence:

- Direct public settings show `registration_enabled=true`.
- On macOS, `tools/tokengate_launch_readiness.sh` reported `registration_enabled=false`.
- The parser uses basic `sed` alternation in `json_bool_value`, which is not portable across BSD/GNU sed.

Impact:

- Release gate can produce false blockers or false confidence depending on host.
- Public readiness checks become less trustworthy.

Requirements:

- Replace ad hoc JSON parsing with `jq`, `node`, `python3`, or a POSIX-safe parser.
- Add a fixture test for settings JSON with true/false values.

Acceptance Criteria:

- The script reports `registration_enabled=true` for the current production JSON.
- The script behaves the same on macOS and Linux.

### GAP-P1-004: Frontend Package Manager Is Not Reproducibly Pinned

Priority: P1

Area: Build reproducibility

Evidence:

- `frontend/package.json` has no `packageManager` field.
- `frontend` contains both `pnpm-lock.yaml` and `package-lock.json`.
- Direct `pnpm` is not on the local PATH.
- Corepack attempted to write `/Users/xiaoboyu/package.json` unless `COREPACK_ENABLE_PROJECT_SPEC=0` was set.
- CI uses pnpm 9; local corepack resolved pnpm 10.26.0.

Impact:

- Local, CI, and Vercel builds can drift.
- Corepack may behave unexpectedly outside the repo root.

Requirements:

- Add a `packageManager` field pinned to the intended pnpm version.
- Remove the unused lockfile after choosing pnpm or npm.
- Align CI, Vercel install command, and local docs.

Acceptance Criteria:

- `corepack pnpm --version` works from `frontend/` without writing outside the repo.
- CI and Vercel use the same package manager major version.

### GAP-P1-005: Admin Launch Readiness Does Not Cover Live Provider Smoke Evidence

Priority: P1

Area: Admin UX, operations

Evidence:

- Admin launch readiness page checks public settings.
- It does not itself prove current Claude/OpenAI account test success, model visibility, usage logging, payment webhook status, SMTP status, or backup recency.

Impact:

- Admin may see readiness items pass while the most important production dependencies are unverified.

Requirements:

- Extend admin readiness to include:
  - last successful Claude account test.
  - last successful OpenAI account test.
  - last successful `/v1/messages` and `/v1/chat/completions` smoke.
  - last SMTP test and password reset test.
  - last payment test order and webhook confirmation.
  - last backup timestamp and restore drill timestamp.

Acceptance Criteria:

- Admin readiness page shows hard fail for stale or missing provider/payment/email/backup evidence.

### GAP-P1-006: Several Admin Stats Endpoints Still Return Mock Values

Priority: P1

Area: Operations, admin trust

Evidence:

- `DashboardHandler.GetRealtimeMetrics` returns hard-coded zero metrics.
- `ProxyHandler.GetStats` returns hard-coded proxy stats.
- `GroupHandler.GetStats` returns hard-coded group stats.
- `RedeemHandler.GetStats` returns hard-coded redeem stats.

Impact:

- Admin may trust false operational metrics during public launch.
- Support/debugging is weaker when public users report issues.

Requirements:

- Either implement real metrics or hide/label these panels as unavailable.
- Prefer using existing dashboard/usage repositories where possible.

Acceptance Criteria:

- No public/admin launch-critical page displays mock values without a visible label.

### GAP-P1-007: Railway Project Is Not Linked Locally For Deployment/Env Audit

Priority: P1

Area: Deployment operations

Evidence:

- `railway status` returned: `No linked project found. Run railway link to connect to a project`.
- HTTP probes prove backend is online, but this audit could not verify Railway environment variables, services, volume/backups, or deployment config through CLI.

Impact:

- Public release cannot be fully audited from repo tooling.
- Env drift can remain hidden until runtime failures.

Requirements:

- Link the local repo to the correct Railway project or document the authoritative audit path.
- Run a redacted env audit covering required variables in `TOKENGATE_PRODUCTION_ENV_CHECKLIST.md`.
- Capture backup and restore evidence.

Acceptance Criteria:

- A release operator can run one documented command set to verify Railway service, env, domains, and backup readiness without exposing secrets.

### GAP-P1-008: Branding And Release Automation Still Carry Sub2API Residue

Priority: P1 for public brand trust; P2 if runtime compatibility requires keeping internal names

Area: Branding, release, contributor process

Evidence:

- `backend/go.mod` module path is `github.com/Wei-Shaw/sub2api`.
- `.github/workflows/cla.yml` points to `Wei-Shaw/sub2api` CLA.
- `.github/workflows/release.yml` uses DockerHub and package references for `sub2api`, including public release notification copy.
- `README_JA.md` is still a Sub2API README.
- `deploy/*` files still use `sub2api` service names/images; `deploy/README.md` says some names remain for compatibility.

Impact:

- Public release may confuse users about whether TokenGate is an independent service or a Sub2API deployment.
- Release artifacts may be published under the wrong package names or descriptions.

Requirements:

- Decide which names stay for runtime compatibility and which must be public-facing TokenGate.
- Update release notification, Docker image descriptions, CLA links, public docs, and non-English README files.
- If internal module path remains `sub2api`, explicitly document it as an implementation detail.

Acceptance Criteria:

- Public pages, release notes, container descriptions, and contributor docs say TokenGate.
- Any remaining Sub2API references are internal compatibility references with clear rationale.

### GAP-P2-001: Frontend Build Has Performance Warnings

Priority: P2

Area: Frontend performance

Evidence:

- `vite build` succeeds but reports dynamic/static import chunking warnings.
- `AccountsView` chunk is larger than 500 kB after minification.

Impact:

- Admin-heavy pages may load more slowly.
- Not a hard release blocker if public landing/docs are fast, but it affects perceived polish.

Requirements:

- Add manual chunks or split large admin views.
- Validate initial public route bundle size and Lighthouse metrics.

Acceptance Criteria:

- Public pages load within the agreed performance budget.
- Build warnings are either resolved or documented as acceptable for admin-only chunks.

### GAP-P2-002: Public Pricing And Plan Semantics Need Final Owner Sign-Off

Priority: P2 for invite-only; P1 for self-serve paid launch

Area: Pricing, billing copy

Evidence:

- Public `/pricing` route exists.
- `TOKENGATE_LAUNCH_ROADMAP.md` still marks billing and plan semantics as partially complete.
- Payment is disabled in live public settings.

Impact:

- Users may not understand included usage, balance deduction, top-up behavior, or plan limits.

Requirements:

- Approve first public plan lineup.
- Confirm admin subscription plan configuration matches public pricing copy.
- Confirm deduction order for included balance, bonus balance, and prepaid balance.

Acceptance Criteria:

- Public pricing page, admin plan settings, and actual billing behavior match one documented model.

## 7. Release Readiness Requirements

### 7.1 P0 Completion Checklist

TokenGate can move to public self-serve release only when:

- Backend `govulncheck ./...` is green.
- Frontend security audit exceptions checker is green.
- GitHub Actions `CI` and `Security Scan` are green on `main`.
- `tools/tokengate_launch_readiness.sh` correctly parses production settings on macOS and Linux.
- Public settings pass the self-serve profile:
  - `registration_enabled=true`
  - `email_verify_enabled=true` or documented equivalent
  - `password_reset_enabled=true`
  - `payment_enabled=true` if users can self-serve paid access
  - `contact_info` set
  - legal documents non-empty
- Claude and ChatGPT/OpenAI model requests both succeed with a production TokenGate API key.
- `/v1/models` exposes required Claude and OpenAI-compatible models for the test key.
- `/v1/usage`, dashboard usage records, API key `Last Used`, and balance/plan totals update after both provider requests.
- SMTP test and password reset are verified in production.
- Payment order and webhook are verified, or launch profile is explicitly invite-only with payment disabled.
- Backup export and restore drill are documented with current production/staging evidence.

### 7.2 Invite-Only Beta Checklist

TokenGate can expand from self-use to invite-only beta when:

- Security scans are green or all exceptions are explicit, current, and owner-approved.
- CI is green.
- Password reset may be disabled only if support has a documented manual recovery process.
- Payment may be disabled only if users are manually provisioned.
- ChatGPT/OpenAI route is verified with real usage logging.
- Legal/support copy clearly says access is invite-only and billing is manual/approval-based.

## 8. Recommended Execution Plan

### Phase 0: Stop-The-Line Security And CI

1. Upgrade Go toolchain and `golang.org/x/net`.
2. Resolve or risk-accept `js-cookie` and refresh audit exceptions with real owner.
3. Fix frontend lint duplicate key.
4. Fix backend path test portability.
5. Re-run local and GitHub CI/security scans.

### Phase 1: Prove Gateway Value Proposition

1. Configure production OpenAI upstream account and group routing.
2. Run Claude and OpenAI smoke with one production API key.
3. Verify usage, token, cost, balance, and last-used updates.
4. Update `TOKENGATE_LAUNCH_ROADMAP.md` with current smoke evidence.

### Phase 2: Decide Release Profile

1. Choose invite-only or self-serve.
2. If self-serve, enable email verification, password reset, legal acceptance, and payment.
3. If invite-only, close self-serve purchase paths and update public copy.

### Phase 3: Domain And Branding

1. Assign `tokengate.to`, `app.tokengate.to`, and `api.tokengate.to` or final equivalents.
2. Connect Vercel/Railway domains.
3. Update CORS, frontend env, docs, reset links, and quickstart examples.
4. Clean public Sub2API residue in release automation and docs.

### Phase 4: Operational Evidence

1. Link Railway project or document an equivalent read-only audit workflow.
2. Verify SMTP, payment, backups, restore drill, and ops alerting.
3. Extend admin readiness page to show fresh evidence timestamps.

## 9. Final Public Release Definition Of Done

Public release is done when a new user can:

1. Land on the final public domain and understand the product.
2. Read pricing, support, legal terms, and quickstart docs.
3. Sign up with verified email and accepted terms.
4. Recover their password without support.
5. Create an API key.
6. Send one Claude-compatible and one ChatGPT/OpenAI-compatible request.
7. See usage, cost, and balance/plan impact in the UI.
8. Pay or receive entitlement through the documented path.
9. Contact support through a visible production support channel.

Engineering release is done when:

1. CI and security scan are green.
2. Public readiness script passes on final domains.
3. Production smoke scripts pass with Claude and OpenAI required.
4. Backup and restore evidence is current.
5. Release artifacts and public docs use TokenGate branding.
