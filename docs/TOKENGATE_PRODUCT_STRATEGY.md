# TokenGate Product Strategy

## 1. Product Definition

TokenGate should not be positioned as "another model relay."

It should be positioned as:

**A subscription-native AI API gateway for builders and small AI products.**

Its core value is not raw forwarding. Its value is combining these layers into one service:

- subscription plans
- API key issuance
- usage metering
- credit and quota deduction
- model routing
- unified billing records
- team-friendly admin operations

For your own use case, the first customer is the product that turns a product link into text, images, and ad videos. That is a strong starting point because it creates real demand for:

- multi-model orchestration
- provider switching
- cost control
- unified usage accounting
- packaged plans for end users

## 2. Target Users

### Phase 1: Self-serve internal use

Primary user:

- your own AI content generation product

Goal:

- prove that subscription-backed API access is easier to operate than direct per-provider integration

### Phase 2: Friendly external adopters

Primary users:

- indie hackers
- AI wrapper builders
- small agencies
- teams shipping AI features into an existing app

These users do not want to build:

- API key management
- quota deduction
- pricing tables
- subscription resets
- payment and top-up workflows
- model/provider failover

### Phase 3: Public platform

Primary users:

- developers who want a unified AI gateway with built-in commercial controls

At this stage, the product becomes:

- "Stripe for AI usage access" for small software teams

## 3. Recommended Positioning

### What TokenGate is

- a hosted subscription-to-API platform
- a commercial layer on top of LLM, image, and video APIs
- a control plane for selling AI usage as plans, credits, and keys

### What TokenGate is not

- not a generic consumer chatbot
- not just an OpenAI-compatible proxy
- not an enterprise integration suite on day one
- not a marketplace for every possible provider at launch

## 4. Product Architecture Decision

The imported `sub2api` codebase is a strong foundation for TokenGate.

Why it fits:

- it already has gateway logic
- it already has API keys, users, subscriptions, usage logs, and payments
- it already uses the stack you want: Go backend, Vue frontend, Postgres, Redis
- it already supports admin and end-user surfaces

Why it cannot be shipped unchanged:

- it is optimized for a broad relay platform, not your narrow initial product
- it contains many advanced and region-specific capabilities that will slow down your first public launch
- the current information architecture is too wide for a clean mainstream product narrative

So the right move is:

**Use `sub2api` as the core platform, but aggressively simplify the first TokenGate release.**

## 5. What To Keep

These are already aligned with TokenGate and should stay in the product core.

### Backend core

- gateway request handling in [backend/internal/handler/endpoint.go](/Users/xiaoboyu/token-gate/backend/internal/handler/endpoint.go:1)
- API key management in `backend/internal/service/api_key*.go`
- subscription logic in `backend/internal/service/subscription*.go`
- usage billing and logging in `backend/internal/service/billing*.go` and `backend/internal/service/usage*.go`
- pricing resources in `backend/resources/model-pricing`
- payment order and fulfillment services in `backend/internal/service/payment*.go`
- account and channel abstractions for upstream providers

### Data model

- `subscription_plans` in [backend/ent/schema/subscription_plan.go](/Users/xiaoboyu/token-gate/backend/ent/schema/subscription_plan.go:1)
- `usage_logs` in [backend/ent/schema/usage_log.go](/Users/xiaoboyu/token-gate/backend/ent/schema/usage_log.go:1)
- `api_keys`, `users`, `user_subscriptions`, `payment_orders`, `settings`

### Frontend surfaces

- auth
- user dashboard
- keys
- usage
- subscriptions
- payment flow
- admin dashboard
- admin users
- admin subscriptions
- admin usage
- admin settings

## 6. What To Remove From V1 Scope

These features may be valuable later, but they should not define TokenGate V1.

### De-prioritize from product narrative

- affiliate system
- announcement system
- backup UI
- channel monitor UI
- proxy management UI
- redeem code complexity
- extensive ops center
- broad OAuth identity matrix
- WeChat/LinuxDo-specific onboarding unless you need them immediately
- highly specialized provider-specific behaviors that only exist for niche relay scenarios

### Keep in code for now, but hide in V1 UI

You do not need to delete these immediately. The lower-risk move is:

- keep backend support
- remove or hide navigation entry points
- avoid documenting them publicly

That gets you to a cleaner launch much faster.

## 7. V1 Product Shape

TokenGate V1 should feel like a focused commercial developer tool.

### Public site promise

"Sell AI API access with subscriptions, credits, and unified model routing."

### V1 user journey

1. User lands on the site.
2. User sees plans and supported model categories.
3. User signs up.
4. User buys a subscription or tops up credits.
5. User creates an API key.
6. User calls a unified endpoint.
7. User views usage, cost, and remaining balance.

### V1 admin journey

1. Admin configures providers.
2. Admin defines plans.
3. Admin sets model pricing and access rules.
4. Admin monitors usage and payment success.
5. Admin supports users through a small, clear operations console.

## 8. V1 Information Architecture

### Public / marketing

- Home
- Pricing
- Docs
- Sign in
- Sign up

### User app

- Dashboard
- API Keys
- Usage
- Billing
- Subscriptions
- Profile

### Admin app

- Overview
- Users
- API Keys
- Providers
- Plans
- Usage
- Payments
- Settings

This is much narrower than the current frontend route graph in [frontend/src/router/index.ts](/Users/xiaoboyu/token-gate/frontend/src/router/index.ts:1), and that is intentional.

## 9. Business Model Recommendation

TokenGate should support two billing modes from the start:

- subscription plans
- prepaid credits

That gives you flexibility.

## 9.1 Credits vs Tokens

This needs a strict product definition before launch.

### Recommendation

**Credits should not be presented as tokens.**

They should be presented as:

- an internal wallet unit
- a billing balance
- a prepaid spend balance

Tokens are model-specific usage units. Credits are platform billing units.

That distinction matters because:

- different models have very different token prices
- image and video workloads are not cleanly expressed as text tokens
- public users understand money and token prices better than abstract credits
- credits can easily feel opaque or manipulative if they are the only visible pricing layer

### Public-facing rule

For public pricing, TokenGate should primarily display:

- price per 1M input tokens
- price per 1M output tokens
- price per image or per video job when token pricing is not the right abstraction

### Internal rule

Internally, you can still keep a wallet-like ledger:

- user balance in USD
- or user balance in platform credits pegged to USD

But if you keep credits internally, the peg must be simple and stable:

- `1 credit = $1` is acceptable
- `1000 credits = $1` is acceptable

What you should avoid is:

- "1 credit" meaning an unclear amount of model usage
- separate hidden exchange rates by model
- forcing users to mentally convert credits into true API pricing

## 9.2 Recommended Public Pricing Model

For a public developer-facing product, the best default is:

- token models priced per 1M tokens
- image models priced per image
- video models priced per generation job, second, or output package

This matches how sophisticated API buyers already think.

It is also aligned with the current codebase, which already stores per-token pricing in the backend and displays scaled per-million pricing in the frontend.

Relevant implementation signals:

- backend pricing stores cost as per-token USD values in [backend/internal/service/billing_service.go](/Users/xiaoboyu/token-gate/backend/internal/service/billing_service.go:1)
- pricing data sources are per-token in [backend/internal/service/pricing_service.go](/Users/xiaoboyu/token-gate/backend/internal/service/pricing_service.go:1)
- frontend formatting already scales token pricing to per-million display in [frontend/src/utils/pricing.ts](/Users/xiaoboyu/token-gate/frontend/src/utils/pricing.ts:1)

## 9.3 Recommended Commercial Presentation

The cleanest public model is:

- plans include monthly wallet balance or included spend
- usage is metered transparently by model pricing
- overages continue automatically if balance is available

Example:

- Starter: $29/month, includes $30 platform balance
- Growth: $99/month, includes $110 platform balance
- Pay as you go beyond included balance at published model rates

This is better than selling "10 million tokens included" across all models because the economics of:

- GPT-class text models
- image generation models
- ad video generation models

are too different.

### Why not package everything purely as tokens

If you define the whole platform in tokens:

- video becomes awkward
- image pricing becomes artificial
- users may think all tokens are comparable across providers
- margin control gets harder when providers change pricing

So the better abstraction is:

- publish token pricing where token pricing is native
- publish per-image or per-video pricing where token pricing is not native
- settle everything against a simple wallet balance underneath

### Recommended commercial model

Plan A:

- Starter subscription includes monthly included balance
- Growth subscription includes more included balance and higher rate limits
- pay-as-you-go overage after included balance is exhausted

This model is better than "unlimited" because your upstream costs are variable.

### Internal accounting rule

All usage should reduce a single internal billing ledger, even if the public-facing UI talks about:

- monthly included balance
- top-ups
- bonus balance

That keeps billing explainable and operationally safe.

## 10. Infrastructure Recommendation

Your selected stack fits the product well.

### Deployment

- backend to Railway
- Postgres on Railway
- Redis on Railway
- frontend to Vercel
- domain on `tokengate`

### Recommended environment split

- production
- staging

### Recommended domain layout

- `tokengate.yourdomain.com` or `app.yourdomain.com` for the app
- `api.yourdomain.com` for public API
- `www.yourdomain.com` for marketing/docs if you want separation later

For V1, you can also simplify to one app domain plus `/api`.

## 11. Key Product Risks

### 1. Overbuilding before narrowing the product

The biggest risk is treating the current imported codebase as the product itself.

It is the platform foundation, not the final V1 shape.

### 2. Weak pricing semantics

If plans, credits, and true upstream cost are not aligned, you will create support burden and margin risk.

### 3. Too many payment and auth options too early

Every extra payment rail and identity provider increases testing, compliance, and support work.

### 4. Too many models/providers on day one

Your first version should support only the model surfaces you actually need:

- text
- image
- video

## 12. Recommended Scope Cuts For Launch

Launch with:

- email login
- one or two OAuth options at most
- Stripe first if public/global
- one China payment path only if you truly need it immediately
- OpenAI-compatible API surface first
- usage dashboard
- subscription purchase
- credit top-up
- admin pricing and provider management

Delay:

- affiliate/referral
- complex promo systems
- embedded external systems
- advanced monitoring center
- multi-region payment sprawl
- long tail provider integrations

## 13. Product Roadmap

### Phase 0: Foundation audit

- import `sub2api`
- understand module boundaries
- define TokenGate V1 scope

### Phase 1: Internal production use

- configure your own providers
- define your own plans
- run your product traffic through TokenGate
- validate usage accounting and payment flows

### Phase 2: Simplified public beta

- narrow frontend navigation
- tighten onboarding
- publish public pricing and docs
- expose stable developer API endpoints

### Phase 3: Expansion

- teams and workspaces
- better analytics
- more provider types
- webhooks
- overage automation
- enterprise controls

## 14. Immediate Build Recommendation

The next implementation pass should focus on simplification, not expansion.

Priority order:

1. define TokenGate V1 navigation and product copy
2. hide non-V1 modules in the frontend
3. standardize a small set of public API endpoints
4. review payment configuration for your target market
5. review subscription and credit semantics end to end
6. prepare Railway and Vercel deployment config for your domain

## 15. Final Recommendation

Yes, this should become a standalone public-facing service.

But the winning move is not to build more from scratch. The winning move is to:

- keep the strong `sub2api` core
- remove product noise
- sharpen the commercial story
- launch a much narrower V1

TokenGate can become a real business if it is introduced as a clean product for selling AI access, not as a kitchen-sink relay console.
