# TokenGate BYO AI Accounts PRD

Date: 2026-07-01

Status: Draft for owner review

## 1. Executive Summary

TokenGate should support a second capacity model: users can bring their own AI
provider accounts through OAuth and use TokenGate as a private API gateway for
those accounts.

Today, a user creates a TokenGate API key and selects a TokenGate group. That
request path uses TokenGate-managed upstream capacity and is billed against the
user's balance, subscription included usage, and model pricing.

With BYO AI Accounts, a user can connect their own AI account from My Account >
Accounts. TokenGate automatically creates a private group for that connected
account. When the user creates or edits an API key, they can choose either:

- TokenGate capacity, which uses TokenGate groups and is billed by model usage.
- My connected AI account, which uses the user's private group and does not
  deduct TokenGate model usage charges.

BYO usage should still be authenticated, routed, logged, rate limited where
needed, and visible in Usage. The usage cost charged by TokenGate is 0 for the
model request path. Provider-side charges, quota consumption, bans, or rate
limits remain the user's responsibility. In a later phase, TokenGate can charge a
subscription or management fee for BYO account hosting.

## 2. Product Goals

- Let regular subscribed users connect and use their own AI provider accounts.
- Reduce TokenGate's upstream supply cost for users who already have provider
  subscriptions.
- Preserve TokenGate's value as a unified API key, OAuth, routing, usage log,
  and account health control plane.
- Keep the user-facing concept simple: "TokenGate capacity" vs "my connected
  account", not "admin group routing internals".
- Create a clean path to future paid BYO account management.

## 3. Non-Goals

- Letting users resell or share their connected AI accounts with other users.
- Mixing user-owned accounts into the admin-managed public account pool.
- Building team/shared BYO account ownership in the first release.
- Supporting every credential type on day one; OAuth is the primary V1 path.
- Removing TokenGate usage logs for BYO traffic.
- Letting BYO traffic bypass TokenGate controls. User concurrency, abuse
  controls, API key status, and gateway safeguards still apply.

## 4. Target Users

Primary users:

- Developers who already pay for OpenAI, Gemini, Claude, or related AI accounts.
- Subscribers who want a stable OpenAI-compatible or Claude-compatible endpoint
  without giving up their provider account quota.
- Power users who want TokenGate's API key management, usage visibility, and
  routing while keeping provider spend on their own account.

Secondary users:

- TokenGate operators who want to increase adoption without carrying all
  upstream capacity cost.
- Support/admin users who need clear ownership and fault attribution for BYO
  account issues.

## 5. Product Positioning

User-facing copy should frame BYO as:

```text
Connect your own AI account and use TokenGate as a private API gateway.
TokenGate will not charge model usage for requests routed through your connected
account. Your upstream provider may still charge your provider account or consume
your provider quota.
```

Avoid exposing "group" as the main user concept. In user-facing UI, use:

- Capacity source
- TokenGate capacity
- Connected account
- My AI account
- Usage charge
- Provider quota

Admin and backend docs may continue to use "group" and "account" precisely.

## 6. Recommended Product Approach

Recommended approach: extend the existing account and group model with
user-owned private accounts.

This is not a UI-only feature. The current account and group schema is
admin-global, so BYO requires new ownership, access-control, and billing
invariants before the user-facing page can be safely exposed.

When a user connects an AI provider account:

1. TokenGate stores the OAuth credentials as a user-owned account.
2. TokenGate creates a private group linked to that account.
3. The private group is visible only to that user in API key creation/edit flows.
4. API keys that select the private group route only to that user's connected
   account.
5. Usage logs are created, but TokenGate model usage charge is 0.

V1 decision: create one private group per connected account. This keeps
deletion, disable, health attribution, and scheduler isolation simple. A
per-provider group with multiple user-owned accounts should wait until paid
multi-account BYO routing is intentionally designed.

Alternatives considered:

- Separate user-owned account tables. This creates clean product separation, but
  duplicates routing, testing, account status, and model visibility logic.
- Direct `api_key -> account` binding with no group. This is simpler for one
  account, but it bypasses too much of TokenGate's existing group-based routing,
  model visibility, and compatibility surface.

The private-group approach gives the best balance: users get a simple
"connected account" experience, while engineering can reuse the existing
group/account scheduler and API key model.

## 7. Information Architecture

Add a new user page:

```text
My Account
- API Keys
- Accounts
- Usage
- Plans / Subscriptions
- Profile
```

For admin users, the same Accounts item should appear under the personal My
Account section, not under admin account management.

An admin's personal connected account is still a user-owned BYO account. It must
be scoped to that admin user's owner ID and must never leak into the admin-managed
public account pool.

The user Accounts page should be inspired by the admin Accounts page, but it
must be narrower:

- list connected accounts
- connect account
- re-auth account
- test account
- disable/enable scheduling
- delete account
- view status, provider, last used, and basic usage/health hints

It should not expose admin-only controls such as global group assignment,
priority tuning across shared pools, bulk import/export, proxy fleet management,
or data sync tools.

## 8. API Key Creation UX

The API key creation/edit flow should introduce a source selector:

```text
Capacity source

( ) TokenGate capacity
    Uses TokenGate-managed groups. Requests are billed by model usage.

( ) My connected AI account
    Uses one of your connected accounts. TokenGate model usage charge is $0.
```

If the user chooses TokenGate capacity:

- show available TokenGate groups
- require active balance/subscription rules as today
- keep existing usage pricing and deduction behavior

If the user chooses My connected AI account:

- show only the user's private connected-account groups
- show provider, account name, status, and last health check
- do not require TokenGate balance for model usage
- warn that the provider may charge or rate limit the user's own account

If the user has no connected accounts, show an empty state with a Connect account
action.

## 9. Functional Requirements

### 9.1 User Account Management

- A regular authenticated user can list only their own connected AI accounts.
- A regular authenticated user can start an OAuth connection flow for supported
  providers.
- A regular authenticated user can finish the OAuth flow and create a connected
  account.
- TokenGate automatically creates one private group for each connected account.
- The private group is selectable only by the owner in API key creation/edit.
- The connected account is never schedulable by other users or by public
  TokenGate groups.
- Users can rename, re-authenticate, disable, enable, test, and delete their own
  connected accounts.
- Deleting or disabling a connected account must make dependent API keys safe:
  they should stop routing to that account and show a clear action needed state.

### 9.2 Supported Providers

V1 should reuse the providers already supported by the admin OAuth account
system where they can be safely scoped to a user owner.

Recommended V1 order:

1. OpenAI OAuth.
2. Gemini OAuth.
3. Anthropic/Claude OAuth or setup-token flow, if the current OAuth flow can be
   safely exposed to regular users.
4. Antigravity OAuth, if it is a desired user-facing provider.

Each provider can be feature-gated independently.

### 9.3 Billing And Usage

- TokenGate-managed groups continue to deduct usage according to model pricing,
  subscription included usage, and balance.
- BYO private groups must not deduct TokenGate model usage charges.
- BYO must use an explicit capacity marker, such as
  `capacity_source=connected_account`, on the private group, API key, or resolved
  request context. Do not implement BYO billing as a magic `rate_multiplier=0`.
- BYO usage logs must still record model, tokens/units, endpoint, latency,
  status, upstream account, and API key.
- BYO usage records should preserve provider-estimated usage cost in
  `total_cost` where TokenGate can estimate it, and force TokenGate charged usage
  to 0 in `actual_cost`.
- BYO usage records should clearly label source as "Connected account" or
  equivalent.
- BYO capacity must short-circuit pre-request balance, quota, and subscription
  checks for model usage. TokenGate may still apply API key status, user
  concurrency, abuse controls, and future BYO management-plan checks.
- Future management fees should be modeled separately from model usage charges.
  Examples: monthly BYO access plan, connected-account seat limit, or per-account
  hosting fee.

### 9.4 Routing And Isolation

- A user's private group must route only to that user's connected account.
- Another user's API key must never see or select the private group.
- Admin-managed public groups must never schedule user-owned BYO accounts.
- Sticky sessions, failover, model visibility, and account health logic should
  respect user ownership.
- BYO requires defense-in-depth, not UI filtering only. Ownership must be
  enforced at all of these points:
  - group/account schema and repository queries
  - user group list and connected account list
  - API key create/edit group validation
  - API key auth/cache snapshots
  - scheduler account selection
  - fallback and sticky-session paths
- A crafted request or API call using another user's private `group_id` must be
  rejected server-side before routing.
- If a BYO account is unavailable, fail within the user's private source rather
  than falling back to TokenGate capacity unless the user explicitly configured
  such fallback in a later phase.
- V1 must not automatically fall back from BYO to TokenGate capacity. A request
  that starts in BYO has `$0` TokenGate model usage semantics; fallback to
  TokenGate capacity would require explicit user consent and charged billing.

### 9.5 Admin Controls

Admins should be able to:

- enable or disable BYO globally
- enable or disable BYO by provider
- set max connected accounts per user
- see support-safe metadata for user-owned accounts
- disable a user-owned account for abuse or safety
- distinguish public TokenGate accounts from user-owned private accounts

Admins should not see raw OAuth tokens.

Disabling BYO globally or disabling a provider must stop new BYO routing for
affected API keys immediately, with a clear user-facing error.

## 10. Security, Compliance, And Abuse

BYO materially increases TokenGate's security and compliance responsibility
because TokenGate stores and operates end-user provider credentials.

Requirements:

- Add explicit owner fields to user-owned accounts and private groups. The exact
  schema can be refined during implementation, but ownership must be first-class
  and queryable.
- Encrypt BYO OAuth credentials at rest with envelope encryption or an equivalent
  key-management pattern. Plain JSONB token storage is not acceptable for
  user-owned BYO credentials.
- Do not expose raw OAuth tokens to admins, users, logs, exports, or support
  tooling.
- Run provider-specific Terms of Service and OAuth-client policy review before
  enabling each provider in production.
- Treat provider launch as blocked until the relevant ToS/OAuth review is
  complete.
- Add BYO abuse controls for stolen-account and anonymized-access scenarios,
  including anomalous connect patterns, repeated re-auth failures, and sudden
  high-volume BYO traffic.
- Support admin emergency disable for a user-owned account, a provider, or BYO
  globally.

## 11. Error Handling

BYO errors need clear attribution.

Examples:

- "Your connected OpenAI account needs re-authentication."
- "Your connected Gemini account is rate limited by Google."
- "Your connected account does not currently support this model."
- "This API key is linked to a deleted connected account."
- "Connected account routing is disabled by the administrator."

Avoid generic platform phrasing such as "No available accounts" when the cause
is specifically the user's connected account.

If a refresh token expires, is revoked, or fails mid-request, TokenGate should
fail the request with a BYO-specific auth error, mark the connected account as
needs re-authentication, and not fall back to TokenGate capacity.

## 12. Metrics

Track:

- connected accounts by provider
- active BYO API keys
- BYO request volume
- BYO success/error rate by provider
- re-auth required rate
- BYO vs TokenGate capacity usage share
- BYO users who later upgrade to management-fee plans
- support tickets tied to BYO provider errors
- BYO auth-failure rate distinct from provider rate-limit failures
- BYO credential encryption and key-rotation health
- anomalous BYO connect or traffic patterns

## 13. Acceptance Criteria

- A regular user sees Accounts under My Account.
- A regular user can connect a supported AI account through OAuth.
- After connection, a private group is created automatically.
- The user can create an API key using either TokenGate capacity or their
  connected account.
- TokenGate capacity API keys continue to bill by model usage.
- Connected-account API keys create usage logs but charge 0 TokenGate model
  usage.
- The connected account and private group are not visible to other users.
- Public/admin-managed groups do not schedule connected user accounts.
- Account re-auth, disable, delete, and error states are visible and actionable.
- The UI warns that provider-side charges and quota consumption are the user's
  responsibility.
- Server-side API key create/edit rejects cross-user private group IDs.
- Scheduler selection rejects any user-owned account whose owner does not match
  the authenticated API key user.
- Global BYO disable and provider-level disable immediately stop affected BYO key
  routing.
- BYO credentials are encrypted at rest and excluded from logs, exports, and
  support payloads.
- Provider ToS/OAuth policy review is complete before a provider is enabled in
  production.

## 14. Phased Rollout

### Phase 1: BYO Foundation

- Add user-owned account ownership and private group isolation.
- Add explicit `capacity_source=connected_account` or equivalent billing marker.
- Add encrypted storage for BYO credentials.
- Complete ToS/OAuth policy review for the first launch provider.
- Add My Account > Accounts.
- Add API key capacity source selector.
- Support one or two OAuth providers behind feature flags.
- Record BYO usage with 0 TokenGate model charge.
- Keep BYO fallback to TokenGate capacity disabled.

### Phase 2: Provider Expansion And Polish

- Add more supported OAuth providers.
- Repeat ToS/OAuth policy review for each additional provider before production
  enablement.
- Add richer account health checks and re-auth prompts.
- Improve model visibility for connected accounts.
- Add admin support filters for user-owned accounts.

### Phase 3: BYO Management Monetization

- Add subscription or management-fee gating for BYO.
- Add connected-account limits by plan.
- Add optional managed fallback or multi-account BYO routing for paid tiers.
- Add team/shared ownership if there is clear demand.

## 15. Product Decisions And Open Questions

V1 product decisions:

- Create one private group per connected account.
- Do not automatically fall back from BYO to TokenGate capacity.
- Do not require a paid TokenGate subscription for BYO model usage at launch.
  Authentication is enough until BYO management fees are introduced.
- Model BYO with an explicit capacity source or billing marker, not a zero
  multiplier.

Open questions:

- Which provider should be the first production BYO OAuth launch provider after
  ToS and OAuth-client policy review.
- Which envelope-encryption key-management implementation should be used for BYO
  credentials.
- Whether future management fees should be sold as a monthly BYO feature, a
  connected-account limit by plan, or a standalone per-account hosting fee.
