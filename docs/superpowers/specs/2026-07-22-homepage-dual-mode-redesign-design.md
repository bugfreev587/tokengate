# TokenGate Dual-Mode Homepage Redesign

Date: 2026-07-22

Status: Approved for implementation

## 1. Objective

Redesign the public TokenGate homepage so a new visitor immediately understands
the two equal ways to use the product:

1. **Usage-based**: preload TokenGate balance and consume TokenGate-managed model
   capacity at published usage rates.
2. **BYO**: connect a supported provider account and use TokenGate as the hosted
   API, routing, key-management, and observability layer around capacity the user
   already owns.

The homepage must present both modes with equal visual weight. It must not imply
that BYO is a secondary feature or that Usage-based is a subscription plan.

## 2. Product Decisions

- Rename the current public-facing “Prepay” concept to **Usage-based** in English
  and **按量付费** in Chinese.
- Show Usage-based and BYO side by side above the fold.
- Do not show hard-coded prices, minimum top-up amounts, trial lengths, or plan
  fees on the homepage. Both mode CTAs lead to `/pricing`, which remains the
  source of current commercial details.
- Explain that BYO requests do not deduct TokenGate prepaid balance. The user's
  provider may still charge its own account or consume its quota.
- Do not imply automatic fallback between capacity sources. The shared-capability
  section describes routing as explicit and isolated by capacity source.
- Replace the current long CLI configuration blocks with a compact three-step
  flow. Detailed Claude Code, Codex, and SDK configuration remains in Docs.

## 3. Information Architecture

### 3.1 Header

The public header contains:

- TokenGate logo and site name
- Product anchor
- Pricing link
- Docs link
- Locale switcher
- Theme switcher
- Login button for signed-out users or Dashboard button for authenticated users

Existing settings continue to control the site logo, site name, Docs URL, and
custom homepage content.

### 3.2 Hero

The hero uses an asymmetric two-part layout:

- The left side carries the shared value proposition, a short explanation, and
  Pricing and Docs actions.
- The right side contains two equal capacity-source panels: Usage-based and BYO.

Approved English message direction:

```text
One gateway. Two ways to run.
Your AI stack, on your terms.
```

The final Chinese copy will communicate the same meaning naturally rather
than translating the English headline word for word.

### 3.3 Capacity-Source Panels

Usage-based panel:

- Prepaid balance
- TokenGate-managed upstream capacity
- Pay for actual model usage
- No recurring subscription requirement for Usage-based access

BYO panel:

- User-owned provider account
- TokenGate API keys, compatible endpoints, routing, and usage logs
- No TokenGate prepaid-balance deduction on the BYO request path
- Provider-side charges, quotas, and rate limits still apply

Both panels link to `/pricing`. Their size, position, typography, and interaction
states must communicate equal product status.

### 3.4 Shared Control Layer

A compact section explains the capabilities shared by both capacity sources:

- API key management
- OpenAI-compatible and Anthropic-compatible endpoints
- Usage logs and request visibility
- Explicit capacity routing without silent cross-mode fallback

### 3.5 How It Works

The homepage presents a concise three-step flow:

1. Choose Usage-based or BYO capacity.
2. Create an API key bound to that capacity source.
3. Connect Claude Code, Codex, an SDK, or a custom application using Docs.

The section names supported client categories where useful but does not embed the current
multi-line environment-variable or TOML configuration examples.

### 3.6 Closing CTA and Footer

The closing CTA sends users to `/pricing` to compare the two modes. The footer
retains links to Pricing, Docs, Support, and Legal resources.

## 4. Visual Direction

- Use a warm neutral base, charcoal primary surfaces, and one restrained green
  accent derived from the existing TokenGate palette.
- Avoid decorative purple, neon glows, oversized centered headings, and generic
  three-column feature-card layouts.
- Use an asymmetric desktop composition. Collapse to a strict single-column
  mobile layout below the existing medium breakpoint.
- Stack the two mode panels on mobile without changing their relative emphasis.
- Preserve light and dark themes.
- Use subtle entrance sequencing plus hover, focus, and active feedback. Avoid
  heavy or perpetual motion.
- Respect `prefers-reduced-motion`.

## 5. Component Boundaries

`HomeView.vue` remains the orchestration layer for public settings,
authentication state, theme state, and custom-home override behavior. The
default homepage is split into focused presentational components with these
boundaries:

- public header
- dual capacity-source hero
- shared capability list
- three-step flow
- closing CTA and footer

Components must use existing Vue, Tailwind CSS 3, router, icon, locale, and theme
patterns. The redesign must not add a new UI or animation dependency.

## 6. Data and Navigation Behavior

No backend or schema changes are required.

The page continues to use:

- cached public settings for logo, site name, subtitle, Docs URL, and custom home
  content
- the auth store for Login versus Dashboard behavior
- the existing router for `/pricing`, `/login`, Dashboard, Support, and Legal
  destinations
- the current locale and theme controls

If public settings are missing or fail to load, existing local fallbacks must
still render a complete homepage. Admin-configured `home_content`, whether HTML
or an iframe URL, continues to replace the default homepage exactly as it does
today.

## 7. Accessibility and Responsive Requirements

- Use semantic landmarks and heading order.
- All buttons and links must have visible keyboard focus states.
- Icon-only actions require localized accessible labels.
- Text and controls must meet WCAG AA contrast in both themes.
- The mobile page must not overflow horizontally at 320 CSS pixels.
- Mode meaning cannot depend on color alone.
- Decorative motion must be disabled or reduced for users who request reduced
  motion.

## 8. Testing and Verification

Automated tests must cover:

- English and Chinese mode terminology
- Usage-based and BYO content on the default homepage
- both mode CTAs targeting `/pricing`
- Docs navigation using the configured internal or external Docs URL
- signed-out Login and signed-in Dashboard behavior
- continued custom `home_content` override behavior
- absence of the removed CLI setup blocks from the default homepage

Verification must include:

- the focused homepage unit tests
- frontend typecheck
- frontend production build
- browser review at desktop and mobile widths
- browser review in light and dark themes
- console-error review during the homepage flow

## 9. Scope Boundaries

This work changes only the public homepage, its focused components, localized
copy, and relevant tests. It does not change:

- backend billing or entitlement logic
- BYO admission and routing behavior
- actual pricing values or the pricing-page commercial configuration
- current unrelated BYO and billing work in the original workspace
- deployment configuration or production state

## 10. Acceptance Criteria

The redesign is complete when:

1. A visitor can identify Usage-based and BYO as two equal choices from the first
   viewport.
2. Both paths lead to Pricing without embedding stale commercial values.
3. The distinction between TokenGate-managed capacity and user-owned capacity is
   accurate in English and Chinese.
4. The page communicates the shared TokenGate control layer and a short setup
   journey without exposing full CLI configuration blocks.
5. Existing dynamic site settings, custom homepage override, auth-aware CTA,
   locale switching, and theming remain functional.
6. Automated and browser verification pass at the defined desktop and mobile
   states.
