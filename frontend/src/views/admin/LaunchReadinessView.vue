<template>
  <div class="space-y-6">
    <section class="rounded-2xl border border-gray-200 bg-white p-6 shadow-sm dark:border-dark-700 dark:bg-dark-800">
      <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <p class="text-xs font-semibold uppercase tracking-[0.22em] text-primary-600 dark:text-primary-400">
            {{ copy.eyebrow }}
          </p>
          <h1 class="mt-3 text-2xl font-bold text-gray-900 dark:text-white">{{ copy.title }}</h1>
          <p class="mt-2 max-w-3xl text-sm leading-7 text-gray-600 dark:text-dark-300">{{ copy.description }}</p>
        </div>
        <button type="button" class="btn btn-secondary" :disabled="loading" @click="refresh">
          {{ loading ? copy.refreshing : copy.refresh }}
        </button>
      </div>
    </section>

    <section class="grid gap-4 lg:grid-cols-2">
      <article class="rounded-2xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <div class="flex items-center justify-between gap-3">
          <div>
            <h2 class="text-lg font-bold text-gray-900 dark:text-white">{{ copy.private.title }}</h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-300">{{ copy.private.description }}</p>
          </div>
          <span :class="badgeClass(privateSummary.level)">{{ summaryText(privateSummary) }}</span>
        </div>
        <div class="mt-5 space-y-3">
          <CheckRow v-for="item in privateChecks" :key="item.key" :item="item" />
        </div>
      </article>

      <article class="rounded-2xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <div class="flex items-center justify-between gap-3">
          <div>
            <h2 class="text-lg font-bold text-gray-900 dark:text-white">{{ copy.public.title }}</h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-300">{{ copy.public.description }}</p>
          </div>
          <span :class="badgeClass(publicSummary.level)">{{ summaryText(publicSummary) }}</span>
        </div>
        <div class="mt-5 space-y-3">
          <CheckRow v-for="item in publicChecks" :key="item.key" :item="item" />
        </div>
      </article>
    </section>

    <section class="grid gap-4 xl:grid-cols-[1.1fr_0.9fr]">
      <article class="rounded-2xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <h2 class="text-lg font-bold text-gray-900 dark:text-white">{{ copy.actions.title }}</h2>
        <div class="mt-5 grid gap-3 sm:grid-cols-2">
          <RouterLink v-for="action in copy.actions.items" :key="action.to" :to="action.to" class="rounded-xl border border-gray-200 p-4 transition hover:border-primary-300 hover:bg-primary-50/60 dark:border-dark-700 dark:hover:border-primary-700 dark:hover:bg-primary-900/10">
            <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ action.title }}</p>
            <p class="mt-1 text-xs leading-5 text-gray-500 dark:text-dark-300">{{ action.body }}</p>
          </RouterLink>
        </div>
      </article>

      <article class="rounded-2xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <div class="flex items-center justify-between gap-3">
          <h2 class="text-lg font-bold text-gray-900 dark:text-white">{{ copy.command.title }}</h2>
          <button type="button" class="text-sm font-semibold text-primary-600 hover:text-primary-700 dark:text-primary-400" @click="copyCommand">
            {{ copy.command.copy }}
          </button>
        </div>
        <p class="mt-2 text-sm leading-6 text-gray-500 dark:text-dark-300">{{ copy.command.description }}</p>
        <pre class="mt-4 overflow-x-auto rounded-xl bg-gray-950 p-4 text-xs leading-6 text-gray-100"><code>{{ readinessCommand }}</code></pre>
      </article>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'

type Level = 'pass' | 'warn' | 'fail'

interface CheckItem {
  key: string
  level: Level
  title: string
  body: string
}

const CheckRow = defineComponent({
  props: {
    item: {
      type: Object as () => CheckItem,
      required: true,
    },
  },
  setup(props) {
    return () => h('div', { class: 'flex gap-3 rounded-xl bg-gray-50 p-4 dark:bg-dark-900/60' }, [
      h('span', { class: statusDotClass(props.item.level) }),
      h('div', { class: 'min-w-0' }, [
        h('p', { class: 'text-sm font-semibold text-gray-900 dark:text-white' }, props.item.title),
        h('p', { class: 'mt-1 text-xs leading-5 text-gray-500 dark:text-dark-300' }, props.item.body),
      ]),
    ])
  },
})

const { locale } = useI18n()
const appStore = useAppStore()
const loading = ref(false)

const settings = computed(() => appStore.cachedPublicSettings)

const apiBaseUrl = (import.meta.env.VITE_API_BASE_URL || '').replace(/\/api\/v1\/?$/, '')
const frontendUrl = typeof window !== 'undefined' ? window.location.origin : 'https://<frontend-domain>'

const readinessCommand = computed(() => `TOKENGATE_FRONTEND_URL="${frontendUrl}" \\
TOKENGATE_BACKEND_URL="${apiBaseUrl || 'https://<backend-domain>'}" \\
TOKENGATE_LAUNCH_PROFILE=private \\
TOKENGATE_RUN_API_SMOKE=0 \\
tools/tokengate_launch_readiness.sh`)

const enCopy = {
  eyebrow: 'Launch readiness',
  title: 'Production launch checklist',
  description: 'This page turns public settings into an operator-friendly launch gate. It does not replace live API, SMTP, payment, or backup drills, but it keeps the obvious blockers visible.',
  refresh: 'Refresh settings',
  refreshing: 'Refreshing...',
  private: { title: 'Private beta gate', description: 'Safe enough for controlled self-use and invited users.' },
  public: { title: 'Public self-serve gate', description: 'Required before opening signup and paid usage broadly.' },
  states: { ready: 'Ready', warning: 'Needs attention', blocked: 'Blocked' },
  checks: {
    siteName: ['Branding configured', 'The public site has a visible brand name.'],
    supportOk: ['Support contact configured', 'Users can find a real contact path from /support.'],
    supportMissing: ['Support contact missing', 'Set Contact Info in Admin -> Settings -> Site before inviting users.'],
    registrationClosed: ['Registration closed', 'Good for private beta or invite-only validation.'],
    registrationRequiredMissing: ['Self-serve registration disabled', 'Enable registration before public self-serve launch, or keep launch mode invite-only.'],
    registrationOpen: ['Registration open', 'Appropriate only when email, abuse controls, and support are ready.'],
    passwordResetOk: ['Password reset enabled', 'Users can recover accounts without founder intervention.'],
    passwordResetMissing: ['Password reset disabled', 'Acceptable for private beta, but not for public self-serve launch.'],
    emailVerifyOk: ['Email verification enabled', 'Self-serve signup has a basic account ownership check.'],
    emailVerifyMissing: ['Email verification disabled', 'Enable this before public self-serve signup.'],
    paymentOk: ['Payment enabled', 'Paid top-up and plan purchase paths are visible.'],
    paymentMissing: ['Payment disabled', 'Fine for invite-only beta; public paid launch needs test-mode webhook verification first.'],
    paymentRequiredMissing: ['Payment disabled', 'Enable payments and verify test-mode webhooks before a public paid launch.'],
  },
  actions: {
    title: 'Where to fix blockers',
    items: [
      { to: '/admin/settings', title: 'System settings', body: 'Configure support contact, SMTP, registration, password reset, and public site settings.' },
      { to: '/admin/accounts', title: 'Provider accounts', body: 'Verify Claude/OpenAI account tests and provider health.' },
      { to: '/admin/orders/plans', title: 'Plans and payment', body: 'Define public plans and payment behavior after choosing the provider.' },
      { to: '/admin/usage', title: 'Usage records', body: 'Confirm real requests create auditable usage and balance changes.' },
    ],
  },
  command: { title: 'CLI readiness command', description: 'Run this from the repo after deploy. Add TOKENGATE_API_KEY when you want live gateway smoke.', copy: 'Copy' },
}

const zhCopy = {
  eyebrow: '上线就绪',
  title: '生产上线检查表',
  description: '这个页面把 public settings 转成运营视角的 launch gate。它不能替代真实 API、SMTP、支付和备份演练，但会把明显 blocker 放在眼前。',
  refresh: '刷新设置',
  refreshing: '刷新中...',
  private: { title: 'Private beta gate', description: '适合自用和邀请制小范围用户验证。' },
  public: { title: 'Public self-serve gate', description: '开放注册和付费使用前必须通过。' },
  states: { ready: 'Ready', warning: '需关注', blocked: 'Blocked' },
  checks: {
    siteName: ['品牌已配置', '公开站点有可见品牌名称。'],
    supportOk: ['支持联系方式已配置', '用户可以在 /support 找到真实联系方式。'],
    supportMissing: ['缺少支持联系方式', '邀请用户前，请在 Admin -> Settings -> Site 设置 Contact Info。'],
    registrationClosed: ['注册已关闭', '适合 private beta 或邀请制验证。'],
    registrationRequiredMissing: ['自助注册未开启', '公开 self-serve launch 前需要开启注册，或继续保持邀请制。'],
    registrationOpen: ['注册已开启', '只有在邮箱、风控和支持流程准备好后才适合开启。'],
    passwordResetOk: ['密码重置已开启', '用户可以自助恢复账号。'],
    passwordResetMissing: ['密码重置未开启', 'private beta 可接受，但 public self-serve launch 不可接受。'],
    emailVerifyOk: ['邮箱验证已开启', '自助注册具备基础账号归属验证。'],
    emailVerifyMissing: ['邮箱验证未开启', '公开自助注册前建议开启。'],
    paymentOk: ['支付已开启', '付费充值和套餐购买路径可见。'],
    paymentMissing: ['支付未开启', '邀请制 beta 可以接受；公开付费前必须先验证 test webhook。'],
    paymentRequiredMissing: ['支付未开启', '公开付费上线前，需要开启支付并验证 test-mode webhook。'],
  },
  actions: {
    title: '去哪里修复 blockers',
    items: [
      { to: '/admin/settings', title: '系统设置', body: '配置联系方式、SMTP、注册、密码重置和公开站点设置。' },
      { to: '/admin/accounts', title: 'Provider 账号', body: '验证 Claude/OpenAI account test 和 provider 健康状态。' },
      { to: '/admin/orders/plans', title: '套餐与支付', body: '选定支付 provider 后，定义公开套餐和支付行为。' },
      { to: '/admin/usage', title: '用量记录', body: '确认真实请求生成可审计 usage 和余额变化。' },
    ],
  },
  command: { title: 'CLI readiness command', description: '部署后在 repo 里运行。需要 live gateway smoke 时，加上 TOKENGATE_API_KEY。', copy: '复制' },
}

const copy = computed(() => locale.value.startsWith('zh') ? zhCopy : enCopy)

function textPair(key: keyof typeof enCopy.checks): [string, string] {
  return copy.value.checks[key] as [string, string]
}

function makeCheck(key: string, level: Level, pairKey: keyof typeof enCopy.checks): CheckItem {
  const [title, body] = textPair(pairKey)
  return { key, level, title, body }
}

const privateChecks = computed<CheckItem[]>(() => {
  const s = settings.value
  return [
    makeCheck('site-name', s?.site_name ? 'pass' : 'fail', 'siteName'),
    makeCheck('support', s?.contact_info ? 'pass' : 'warn', s?.contact_info ? 'supportOk' : 'supportMissing'),
    makeCheck('registration', s?.registration_enabled ? 'warn' : 'pass', s?.registration_enabled ? 'registrationOpen' : 'registrationClosed'),
    makeCheck('password-reset', s?.password_reset_enabled ? 'pass' : 'warn', s?.password_reset_enabled ? 'passwordResetOk' : 'passwordResetMissing'),
    makeCheck('payment', s?.payment_enabled ? 'pass' : 'pass', s?.payment_enabled ? 'paymentOk' : 'paymentMissing'),
  ]
})

const publicChecks = computed<CheckItem[]>(() => {
  const s = settings.value
  return [
    makeCheck('site-name-public', s?.site_name ? 'pass' : 'fail', 'siteName'),
    makeCheck('support-public', s?.contact_info ? 'pass' : 'fail', s?.contact_info ? 'supportOk' : 'supportMissing'),
    makeCheck('password-reset-public', s?.password_reset_enabled ? 'pass' : 'fail', s?.password_reset_enabled ? 'passwordResetOk' : 'passwordResetMissing'),
    makeCheck('registration-public', s?.registration_enabled ? 'pass' : 'fail', s?.registration_enabled ? 'registrationOpen' : 'registrationRequiredMissing'),
    makeCheck('email-public', s?.email_verify_enabled ? 'pass' : 'warn', s?.email_verify_enabled ? 'emailVerifyOk' : 'emailVerifyMissing'),
    makeCheck('payment-public', s?.payment_enabled ? 'pass' : 'fail', s?.payment_enabled ? 'paymentOk' : 'paymentRequiredMissing'),
  ]
})

function summarize(items: CheckItem[]): { level: Level; failures: number; warnings: number } {
  const failures = items.filter(item => item.level === 'fail').length
  const warnings = items.filter(item => item.level === 'warn').length
  return { level: failures > 0 ? 'fail' : warnings > 0 ? 'warn' : 'pass', failures, warnings }
}

const privateSummary = computed(() => summarize(privateChecks.value))
const publicSummary = computed(() => summarize(publicChecks.value))

function summaryText(summary: { level: Level; failures: number; warnings: number }): string {
  if (summary.level === 'fail') return `${copy.value.states.blocked} · ${summary.failures}`
  if (summary.level === 'warn') return `${copy.value.states.warning} · ${summary.warnings}`
  return copy.value.states.ready
}

function badgeClass(level: Level): string {
  const base = 'rounded-full px-3 py-1 text-xs font-bold'
  if (level === 'pass') return `${base} bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300`
  if (level === 'warn') return `${base} bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300`
  return `${base} bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300`
}

function statusDotClass(level: Level): string {
  const base = 'mt-1 h-2.5 w-2.5 flex-shrink-0 rounded-full'
  if (level === 'pass') return `${base} bg-emerald-500`
  if (level === 'warn') return `${base} bg-amber-500`
  return `${base} bg-red-500`
}

async function refresh(): Promise<void> {
  loading.value = true
  try {
    await appStore.fetchPublicSettings(true)
  } finally {
    loading.value = false
  }
}

async function copyCommand(): Promise<void> {
  await navigator.clipboard?.writeText(readinessCommand.value)
  appStore.showSuccess(copy.value.command.copy)
}

onMounted(() => {
  if (!appStore.publicSettingsLoaded) {
    refresh()
  }
})
</script>
