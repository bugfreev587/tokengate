# TokenGate Dual-Mode Homepage Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the existing TokenGate landing page with a responsive, bilingual homepage that presents Usage-based and BYO as equal capacity choices and routes commercial decisions to Pricing.

**Architecture:** Keep `HomeView.vue` as the settings/auth/theme orchestration layer and preserve the custom-home override. Move the default landing experience into focused Vue components under `components/home`, backed by one explicit `home` locale contract in English and Chinese. Use existing Vue Router, Tailwind CSS 3, shared icons, locale controls, and theme state without adding dependencies.

**Tech Stack:** Vue 3 Composition API, TypeScript, Vue Router 4, vue-i18n 9, Tailwind CSS 3, Vitest, Vue Test Utils, Vite.

---

## File Structure

- Create `frontend/src/components/home/HomePublicHeader.vue`: brand, public navigation, locale/theme controls, and auth-aware action.
- Create `frontend/src/components/home/HomeCapacityHero.vue`: shared hero copy and the two equal Pricing-linked capacity panels.
- Create `frontend/src/components/home/HomeSharedCapabilities.vue`: common TokenGate control-layer capabilities.
- Create `frontend/src/components/home/HomeHowItWorks.vue`: three-step path and lightweight supported-client strip.
- Create `frontend/src/components/home/HomeClosingCta.vue`: final Pricing action.
- Create `frontend/src/components/home/HomePublicFooter.vue`: public footer navigation.
- Create `frontend/src/components/home/__tests__/HomeCapacityHero.spec.ts`: equal-mode and Pricing-link contract.
- Create `frontend/src/i18n/__tests__/homepageLocales.spec.ts`: English/Chinese homepage terminology contract.
- Modify `frontend/src/views/HomeView.vue`: reduce it to page orchestration plus custom-home override.
- Modify `frontend/src/views/__tests__/HomeView.spec.ts`: default-page, auth, Docs, and custom-home behavior.
- Modify `frontend/src/i18n/locales/en.ts`: replace the old homepage narrative with dual-mode English copy.
- Modify `frontend/src/i18n/locales/zh.ts`: replace the old homepage narrative with natural Chinese copy.

### Task 1: Lock the bilingual homepage content contract

**Files:**
- Create: `frontend/src/i18n/__tests__/homepageLocales.spec.ts`
- Modify: `frontend/src/i18n/locales/en.ts`
- Modify: `frontend/src/i18n/locales/zh.ts`

- [ ] **Step 1: Write the failing locale contract test**

```ts
import { describe, expect, it } from 'vitest'

import en from '@/i18n/locales/en'
import zh from '@/i18n/locales/zh'

describe('homepage locale contract', () => {
  it('presents Usage-based and BYO as equal English capacity modes', () => {
    expect(en.home.hero.eyebrow).toBe('One gateway. Two ways to run.')
    expect(en.home.modes.usageBased.label).toBe('Usage-based')
    expect(en.home.modes.byo.label).toBe('BYO')
    expect(en.home.modes.byo.points.balance).toContain('No TokenGate prepaid balance deduction')
    expect(en.home.closing.button).toBe('Compare modes on Pricing')
  })

  it('uses natural Chinese terminology for both modes', () => {
    expect(zh.home.modes.usageBased.label).toBe('按量付费')
    expect(zh.home.modes.byo.label).toBe('BYO')
    expect(zh.home.modes.byo.description).toContain('自己的 AI 服务账号')
    expect(zh.home.modes.byo.points.balance).toContain('不会扣除 TokenGate 预付余额')
  })
})
```

- [ ] **Step 2: Run the test and verify the new contract is absent**

Run: `cd frontend && pnpm test:run src/i18n/__tests__/homepageLocales.spec.ts`

Expected: FAIL because `home.hero`, `home.modes`, and `home.closing` do not yet expose the new keys.

- [ ] **Step 3: Add the complete English and Chinese homepage locale trees**

Add matching structures to `en.ts` and `zh.ts` with these top-level groups:

```ts
home: {
  nav: { product: string, pricing: string, docs: string },
  hero: {
    eyebrow: string,
    title: string,
    description: string,
    explorePricing: string,
    readDocs: string
  },
  modes: {
    caption: string,
    equalHint: string,
    usageBased: {
      index: string,
      label: string,
      title: string,
      description: string,
      points: { balance: string, capacity: string, commitment: string },
      cta: string
    },
    byo: {
      index: string,
      label: string,
      title: string,
      description: string,
      points: { account: string, gateway: string, balance: string },
      providerNote: string,
      cta: string
    }
  },
  shared: {
    title: string,
    apiKeys: { title: string, description: string },
    compatibleApis: { title: string, description: string },
    usageLogs: { title: string, description: string },
    explicitRouting: { title: string, description: string }
  },
  flow: {
    eyebrow: string,
    title: string,
    description: string,
    choose: { title: string, description: string },
    key: { title: string, description: string },
    connect: { title: string, description: string }
  },
  clients: { worksWith: string, claude: string, codex: string, openai: string, anthropic: string, custom: string },
  closing: { title: string, button: string },
  footer: { allRightsReserved: string },
  viewDocs: string,
  switchToLight: string,
  switchToDark: string,
  dashboard: string,
  login: string
}
```

Use these approved message anchors:

```text
EN: One gateway. Two ways to run. / Your AI stack, on your terms.
EN Usage-based: Top up. Build. / Prepaid balance / TokenGate-managed capacity
EN BYO: Bring your own access. / No TokenGate prepaid balance deduction
ZH: 一个网关，两种使用方式。 / AI 接入方式，由你决定。
ZH Usage-based: 先充值，再按实际用量使用 TokenGate 管理的模型容量。
ZH BYO: 连接自己的 AI 服务账号，将 TokenGate 用作托管 API 网关。
```

Do not include dollar amounts, minimum top-up amounts, or trial duration.

- [ ] **Step 4: Run the locale test and verify it passes**

Run: `cd frontend && pnpm test:run src/i18n/__tests__/homepageLocales.spec.ts`

Expected: PASS, 2 tests.

- [ ] **Step 5: Commit the locale contract**

```bash
git add frontend/src/i18n/locales/en.ts frontend/src/i18n/locales/zh.ts frontend/src/i18n/__tests__/homepageLocales.spec.ts
git commit -m "feat: define dual-mode homepage copy"
```

### Task 2: Build the equal-capacity hero with TDD

**Files:**
- Create: `frontend/src/components/home/HomeCapacityHero.vue`
- Create: `frontend/src/components/home/__tests__/HomeCapacityHero.spec.ts`

- [ ] **Step 1: Write the failing hero component test**

Mount the component with a `t()` helper backed by `en` and `RouterLinkStub`, then assert the two mode articles and Pricing links:

```ts
const wrapper = mount(HomeCapacityHero, {
  props: { siteSubtitle: 'Subscription-native AI API gateway', docUrl: '/docs', docUrlExternal: false },
  global: { stubs: { RouterLink: RouterLinkStub, Icon: { template: '<span />' } } }
})

expect(wrapper.get('[data-mode="usage-based"]').text()).toContain('Usage-based')
expect(wrapper.get('[data-mode="byo"]').text()).toContain('BYO')
expect(wrapper.get('[data-mode="byo"]').text()).toContain('No TokenGate prepaid balance deduction')

const pricingLinks = wrapper.findAllComponents(RouterLinkStub).filter(link => link.props('to') === '/pricing')
expect(pricingLinks).toHaveLength(3)
expect(wrapper.text()).not.toMatch(/\$19|7-day|minimum top-up/i)
```

- [ ] **Step 2: Run the hero test and verify it fails**

Run: `cd frontend && pnpm test:run src/components/home/__tests__/HomeCapacityHero.spec.ts`

Expected: FAIL because `HomeCapacityHero.vue` does not exist.

- [ ] **Step 3: Implement the hero component**

Use a semantic `<section aria-labelledby="home-hero-title">`, an asymmetric grid, and two equal `<article>` elements. Component props are:

```ts
defineProps<{
  siteSubtitle: string
  docUrl: string
  docUrlExternal: boolean
}>()
```

Required navigation behavior:

```vue
<RouterLink to="/pricing">{{ t('home.hero.explorePricing') }}</RouterLink>
<RouterLink v-if="!docUrlExternal" :to="docUrl">{{ t('home.hero.readDocs') }}</RouterLink>
<a v-else :href="docUrl" target="_blank" rel="noopener noreferrer">{{ t('home.hero.readDocs') }}</a>

<article data-mode="usage-based">...</article>
<article data-mode="byo">...</article>
```

Use equal grid tracks at `md` and above, stack below `md`, add visible `focus-visible` rings, and use only transform/opacity transitions. Add no new animation library.

- [ ] **Step 4: Run the hero test and verify it passes**

Run: `cd frontend && pnpm test:run src/components/home/__tests__/HomeCapacityHero.spec.ts`

Expected: PASS.

- [ ] **Step 5: Commit the hero**

```bash
git add frontend/src/components/home/HomeCapacityHero.vue frontend/src/components/home/__tests__/HomeCapacityHero.spec.ts
git commit -m "feat: add dual-capacity homepage hero"
```

### Task 3: Build shared capabilities, flow, closing CTA, and footer

**Files:**
- Create: `frontend/src/components/home/HomeSharedCapabilities.vue`
- Create: `frontend/src/components/home/HomeHowItWorks.vue`
- Create: `frontend/src/components/home/HomeClosingCta.vue`
- Create: `frontend/src/components/home/HomePublicFooter.vue`
- Create: `frontend/src/components/home/__tests__/HomeSupportingSections.spec.ts`

- [ ] **Step 1: Write the failing supporting-section component test**

Mount the four supporting components with the same `en`-backed `t()` mock and assert:

```ts
expect(shared.text()).toContain('One TokenGate control layer')
expect(shared.text()).toContain('Explicit routing')
expect(flow.text()).toContain('Three steps from capacity to API')
expect(flow.text()).toContain('Claude Code')
expect(flow.text()).toContain('Codex CLI')
expect(closing.getComponent(RouterLinkStub).props('to')).toBe('/pricing')
expect(footer.text()).toContain('Pricing')
expect(footer.text()).toContain('Support')
expect(footer.get('a[href="/legal/privacy"]').exists()).toBe(true)
```

The footer mount props are:

```ts
{ siteName: 'TokenGate', currentYear: 2026, docUrl: '/docs', docUrlExternal: false }
```

- [ ] **Step 2: Run the supporting-section test and verify it fails**

Run: `cd frontend && pnpm test:run src/components/home/__tests__/HomeSupportingSections.spec.ts`

Expected: FAIL because the four components do not exist.

- [ ] **Step 3: Implement the supporting components**

`HomeSharedCapabilities.vue` renders a bordered definition-list style grid from these locale keys:

```ts
const capabilityKeys = ['apiKeys', 'compatibleApis', 'usageLogs', 'explicitRouting'] as const
```

`HomeHowItWorks.vue` renders ordered steps and a client strip from:

```ts
const stepKeys = ['choose', 'key', 'connect'] as const
const clientKeys = ['claude', 'codex', 'openai', 'anthropic', 'custom'] as const
```

`HomeClosingCta.vue` contains one `RouterLink` to `/pricing`.

`HomePublicFooter.vue` accepts `siteName`, `currentYear`, `docUrl`, and `docUrlExternal`; it links to `/pricing`, configured Docs, `/support`, and `/legal/privacy`.

- [ ] **Step 4: Run the supporting-section test and verify it passes**

Run: `cd frontend && pnpm test:run src/components/home/__tests__/HomeSupportingSections.spec.ts`

Expected: PASS.

- [ ] **Step 5: Commit the supporting components**

```bash
git add frontend/src/components/home/HomeSharedCapabilities.vue frontend/src/components/home/HomeHowItWorks.vue frontend/src/components/home/HomeClosingCta.vue frontend/src/components/home/HomePublicFooter.vue frontend/src/components/home/__tests__/HomeSupportingSections.spec.ts
git commit -m "feat: add homepage product journey"
```

The old CLI-code assertions are replaced in Task 4 with:

```ts
expect(text).toContain('One TokenGate control layer')
expect(text).toContain('Explicit routing')
expect(text).toContain('Three steps from capacity to API')
expect(text).toContain('Claude Code')
expect(text).toContain('Codex CLI')
expect(text).not.toContain('ANTHROPIC_BASE_URL')
expect(text).not.toContain('[model_providers.tokengate]')
```

### Task 4: Rewire HomeView while preserving dynamic behavior

**Files:**
- Create: `frontend/src/components/home/HomePublicHeader.vue`
- Modify: `frontend/src/views/HomeView.vue`
- Modify: `frontend/src/views/__tests__/HomeView.spec.ts`

- [ ] **Step 1: Expand the failing orchestration tests**

Use mutable mock state for `useAuthStore` and `useAppStore`, reset it in `beforeEach`, and cover:

```ts
it('renders both capacity modes and three Pricing routes')
it('uses the configured internal Docs route')
it('uses a secure external Docs link when configured')
it('shows Login when signed out and Dashboard when signed in')
it('renders configured home HTML instead of the default homepage')
it('renders configured home URL in an iframe instead of the default homepage')
```

For the custom HTML test:

```ts
appState.cachedPublicSettings = { home_content: '<main data-custom-home>Custom home</main>' }
const wrapper = mountHome()
expect(wrapper.find('[data-custom-home]').exists()).toBe(true)
expect(wrapper.find('[data-mode="usage-based"]').exists()).toBe(false)
```

For the URL test:

```ts
appState.cachedPublicSettings = { home_content: 'https://example.com/custom-home' }
const wrapper = mountHome()
expect(wrapper.get('iframe').attributes('src')).toBe('https://example.com/custom-home')
```

- [ ] **Step 2: Run the orchestration tests and verify they fail**

Run: `cd frontend && pnpm test:run src/views/__tests__/HomeView.spec.ts`

Expected: FAIL on dual-mode markup and new navigation behavior.

- [ ] **Step 3: Implement `HomePublicHeader.vue`**

Use these props and event:

```ts
defineProps<{
  siteName: string
  siteLogo: string
  docUrl: string
  docUrlExternal: boolean
  isDark: boolean
  isAuthenticated: boolean
  dashboardPath: string
  userInitial: string
}>()

defineEmits<{ toggleTheme: [] }>()
```

The brand links to `/home`; Product links to `#capacity-modes`; Pricing routes to `/pricing`; Docs uses internal/external handling; theme and auth actions retain localized accessible labels.

- [ ] **Step 4: Replace the default `HomeView` template**

Preserve the existing custom-content branch exactly, but use `min-h-[100dvh]` and iframe `h-[100dvh]`. The default branch becomes:

```vue
<div v-else class="min-h-[100dvh] overflow-hidden bg-stone-50 text-gray-950 dark:bg-dark-950 dark:text-white">
  <HomePublicHeader ... @toggle-theme="toggleTheme" />
  <main>
    <HomeCapacityHero :site-subtitle="siteSubtitle" :doc-url="docUrl" :doc-url-external="isDocUrlExternal" />
    <HomeSharedCapabilities />
    <HomeHowItWorks />
    <HomeClosingCta />
  </main>
  <HomePublicFooter ... />
</div>
```

Retain the existing computed values, `onMounted` settings/auth loading, `toggleTheme`, and theme persistence. Remove the old terminal CSS and CLI markup.

- [ ] **Step 5: Run focused tests and typecheck**

Run:

```bash
cd frontend
pnpm test:run src/i18n/__tests__/homepageLocales.spec.ts src/components/home/__tests__/HomeCapacityHero.spec.ts src/views/__tests__/HomeView.spec.ts
pnpm typecheck
```

Expected: all focused tests PASS and `vue-tsc --noEmit` exits 0.

- [ ] **Step 6: Commit the completed homepage**

```bash
git add frontend/src/components/home/HomePublicHeader.vue frontend/src/views/HomeView.vue frontend/src/views/__tests__/HomeView.spec.ts
git commit -m "feat: redesign TokenGate public homepage"
```

### Task 5: Production and browser verification

**Files:**
- Modify only if verification exposes a defect: `frontend/src/components/home/*.vue`, `frontend/src/views/HomeView.vue`, or focused tests.

- [ ] **Step 1: Run the complete frontend verification set**

Run:

```bash
cd frontend
pnpm lint:check
pnpm test:run
pnpm typecheck
pnpm build:standalone
```

Expected: all commands exit 0. Record any unrelated pre-existing failure separately and do not broaden scope without evidence.

- [ ] **Step 2: Start the Vite development server**

Run: `cd frontend && pnpm dev --host 127.0.0.1`

Expected: Vite prints a local URL and stays running.

- [ ] **Step 3: Verify desktop and mobile flows in a browser**

At `/home`, verify:

- 1440×900 light theme: both mode panels are visible in the first viewport and have equal width.
- 1440×900 dark theme: all text and focus states remain legible.
- 390×844 light and dark themes: panels stack, navigation remains usable, and no horizontal overflow occurs.
- Usage-based panel, BYO panel, hero Pricing action, and closing CTA all navigate to `/pricing`.
- Docs uses the configured route.
- Locale switch changes all homepage copy without layout breakage.
- No browser console errors appear during the flow.

- [ ] **Step 4: Patch only verified defects and rerun the affected checks**

For every discovered defect, add or tighten a focused assertion first, reproduce the failure, make the smallest component/style correction, then rerun the focused test and production build.

- [ ] **Step 5: Commit verification fixes if any**

```bash
git add frontend/src/components/home frontend/src/views/HomeView.vue frontend/src/views/__tests__ frontend/src/i18n
git commit -m "fix: polish dual-mode homepage responsiveness"
```

- [ ] **Step 6: Confirm the worktree is clean**

Run: `git status --short`

Expected: no output.
