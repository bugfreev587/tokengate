<template>
  <div class="min-h-screen bg-slate-50 text-slate-950 dark:bg-dark-950 dark:text-white">
    <header class="border-b border-slate-200 bg-white/90 backdrop-blur dark:border-dark-800 dark:bg-dark-900/90">
      <div class="mx-auto flex max-w-6xl items-center justify-between gap-4 px-5 py-4">
        <RouterLink to="/home" class="flex items-center gap-3">
          <span class="flex h-10 w-10 items-center justify-center overflow-hidden rounded-xl bg-white shadow-sm ring-1 ring-slate-200 dark:bg-dark-800 dark:ring-dark-700">
            <img :src="siteLogo || '/logo.png'" alt="TokenGate" class="h-full w-full object-contain" />
          </span>
          <span class="text-base font-semibold">{{ siteName }}</span>
        </RouterLink>
        <div class="flex items-center gap-3">
          <RouterLink to="/docs" class="text-sm font-medium text-slate-500 hover:text-slate-900 dark:text-dark-300 dark:hover:text-white">
            {{ copy.nav.docs }}
          </RouterLink>
          <RouterLink to="/support" class="text-sm font-medium text-slate-500 hover:text-slate-900 dark:text-dark-300 dark:hover:text-white">
            {{ copy.nav.support }}
          </RouterLink>
          <RouterLink to="/login" class="rounded-full bg-slate-950 px-4 py-2 text-sm font-semibold text-white transition hover:bg-slate-800 dark:bg-white dark:text-slate-950 dark:hover:bg-slate-200">
            {{ copy.nav.signIn }}
          </RouterLink>
        </div>
      </div>
    </header>

    <main class="mx-auto max-w-6xl px-5 py-10 sm:py-14">
      <section class="grid gap-8 lg:grid-cols-[1fr_0.9fr] lg:items-center">
        <div>
          <p class="text-sm font-semibold uppercase tracking-[0.24em] text-primary-600 dark:text-primary-400">
            {{ copy.eyebrow }}
          </p>
          <h1 class="mt-4 text-4xl font-black tracking-tight sm:text-5xl">
            {{ copy.title }}
          </h1>
          <p class="mt-5 max-w-2xl text-base leading-8 text-slate-600 dark:text-dark-300">
            {{ copy.subtitle }}
          </p>
          <div class="mt-8 flex flex-col gap-3 sm:flex-row">
            <RouterLink to="/login" class="btn btn-primary">{{ copy.cta.start }}</RouterLink>
            <RouterLink to="/docs" class="btn btn-secondary">{{ copy.cta.docs }}</RouterLink>
          </div>
        </div>

        <div class="rounded-[2rem] border border-slate-200 bg-white p-6 shadow-xl shadow-primary-900/10 dark:border-dark-800 dark:bg-dark-900">
          <p class="text-sm font-semibold text-slate-500 dark:text-dark-300">{{ copy.balanceCard.kicker }}</p>
          <p class="mt-3 text-4xl font-black text-slate-950 dark:text-white">{{ copy.balanceCard.title }}</p>
          <p class="mt-3 text-sm leading-7 text-slate-600 dark:text-dark-300">{{ copy.balanceCard.body }}</p>
          <div class="mt-5 rounded-2xl bg-slate-50 p-4 font-mono text-sm leading-7 text-slate-700 dark:bg-dark-800 dark:text-dark-200">
            {{ copy.balanceCard.formula }}
          </div>
        </div>
      </section>

      <section class="mt-12 grid gap-4 md:grid-cols-3">
        <article v-for="item in copy.metering" :key="item.title" class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm dark:border-dark-800 dark:bg-dark-900">
          <p class="text-sm font-semibold text-primary-600 dark:text-primary-400">{{ item.kicker }}</p>
          <h2 class="mt-3 text-xl font-bold">{{ item.title }}</h2>
          <p class="mt-3 text-sm leading-7 text-slate-600 dark:text-dark-300">{{ item.body }}</p>
        </article>
      </section>

      <section class="mt-12 overflow-hidden rounded-3xl border border-slate-200 bg-white dark:border-dark-800 dark:bg-dark-900">
        <div class="border-b border-slate-200 p-6 dark:border-dark-800">
          <h2 class="text-2xl font-bold">{{ copy.rateTable.title }}</h2>
          <p class="mt-2 text-sm leading-7 text-slate-600 dark:text-dark-300">{{ copy.rateTable.body }}</p>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full divide-y divide-slate-200 text-left text-sm dark:divide-dark-800">
            <thead class="bg-slate-50 text-xs uppercase tracking-wider text-slate-500 dark:bg-dark-800 dark:text-dark-300">
              <tr>
                <th class="px-5 py-4">{{ copy.rateTable.columns.category }}</th>
                <th class="px-5 py-4">{{ copy.rateTable.columns.unit }}</th>
                <th class="px-5 py-4">{{ copy.rateTable.columns.example }}</th>
                <th class="px-5 py-4">{{ copy.rateTable.columns.note }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-dark-800">
              <tr v-for="row in copy.rateTable.rows" :key="row.category">
                <td class="px-5 py-4 font-semibold">{{ row.category }}</td>
                <td class="px-5 py-4 text-slate-600 dark:text-dark-300">{{ row.unit }}</td>
                <td class="px-5 py-4 font-mono text-primary-600 dark:text-primary-400">{{ row.example }}</td>
                <td class="px-5 py-4 text-slate-600 dark:text-dark-300">{{ row.note }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section class="mt-12 grid gap-4 lg:grid-cols-2">
        <article v-for="item in copy.principles" :key="item.title" class="rounded-2xl border border-slate-200 bg-white p-6 dark:border-dark-800 dark:bg-dark-900">
          <h2 class="text-xl font-bold">{{ item.title }}</h2>
          <p class="mt-3 text-sm leading-7 text-slate-600 dark:text-dark-300">{{ item.body }}</p>
        </article>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'

const { locale } = useI18n()
const appStore = useAppStore()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'TokenGate')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')

const enCopy = {
  nav: { docs: 'Docs', support: 'Support', signIn: 'Sign in' },
  eyebrow: 'Pricing model',
  title: 'Transparent usage, simple balance',
  subtitle: 'TokenGate should price text by per-1M input/output tokens, media by provider-native units, and settle everything against account balance.',
  cta: { start: 'Start with an API key', docs: 'Read API docs' },
  balanceCard: {
    kicker: 'Recommended V1 language',
    title: 'Balance, not vague credits',
    body: 'Tokens measure model usage. Balance measures what the user has available to spend. Keeping them separate makes invoices, support, and margin checks easier.',
    formula: 'charge = input_tokens * input_rate + output_tokens * output_rate + media_units * media_rate',
  },
  metering: [
    { kicker: 'Text', title: 'Per 1M tokens', body: 'Show input and output token rates separately. This matches how modern model providers price usage.' },
    { kicker: 'Image', title: 'Per image or output unit', body: 'Expose image generation as a separate unit so users do not confuse it with text tokens.' },
    { kicker: 'Video', title: 'Per job, second, or provider unit', body: 'Keep video pricing explicit because provider costs often depend on duration, quality, and resolution.' },
  ],
  rateTable: {
    title: 'Public pricing shape',
    body: 'Exact numbers should be configured in admin after provider costs and target margin are finalized.',
    columns: { category: 'Category', unit: 'Public unit', example: 'Example label', note: 'Product note' },
    rows: [
      { category: 'Text input', unit: '1M input tokens', example: '$X / 1M input', note: 'Usually lower than output tokens.' },
      { category: 'Text output', unit: '1M output tokens', example: '$Y / 1M output', note: 'Usually the main cost driver.' },
      { category: 'Image', unit: 'image or provider unit', example: '$Z / image', note: 'Separate from text model tokens.' },
      { category: 'Video', unit: 'job, second, or provider unit', example: '$N / second', note: 'Publish only after a stable provider path exists.' },
    ],
  },
  principles: [
    { title: 'Included usage', body: 'Plans can include a starting balance. After it is exhausted, either stop requests or move to pay-as-you-go based on the plan.' },
    { title: 'Deduction order', body: 'Use included balance first, then bonus balance, then prepaid balance. Document this order so support can explain every charge.' },
    { title: 'Margin guardrail', body: 'Before opening public signup, each public rate should include provider cost, payment fee, refund risk, and a safety margin.' },
    { title: 'Receipts and support', body: 'Usage records should show model, token counts, media units, charge, and remaining balance after each successful request.' },
  ],
}

const zhCopy = {
  nav: { docs: '文档', support: '支持', signIn: '登录' },
  eyebrow: '定价模型',
  title: '透明用量，简单余额',
  subtitle: 'TokenGate 建议文本按每 1M input/output tokens 定价，媒体按 provider 原生单位定价，最终都从账户余额结算。',
  cta: { start: '创建 API key', docs: '阅读 API 文档' },
  balanceCard: {
    kicker: '推荐的 V1 表达',
    title: '用 Balance，不用模糊 Credits',
    body: 'Tokens 衡量模型用量，Balance 衡量用户可消费余额。把两者分开，发票、客服解释和毛利检查都会更清楚。',
    formula: 'charge = input_tokens * input_rate + output_tokens * output_rate + media_units * media_rate',
  },
  metering: [
    { kicker: '文本', title: '每 1M tokens', body: '分别展示 input 和 output token 单价，这和现代模型 provider 的计价方式一致。' },
    { kicker: '图片', title: '按图片或输出单位', body: '图片生成独立计量，避免用户把它和文本 tokens 混在一起理解。' },
    { kicker: '视频', title: '按任务、秒数或 provider 单位', body: '视频成本通常和时长、质量、分辨率有关，需要单独公开说明。' },
  ],
  rateTable: {
    title: '公开定价形态',
    body: '具体数字应在确认 provider 成本和目标毛利后，由管理员在后台配置。',
    columns: { category: '类别', unit: '公开单位', example: '示例标签', note: '产品说明' },
    rows: [
      { category: '文本输入', unit: '1M input tokens', example: '$X / 1M input', note: '通常低于 output tokens。' },
      { category: '文本输出', unit: '1M output tokens', example: '$Y / 1M output', note: '通常是主要成本来源。' },
      { category: '图片', unit: '图片或 provider 单位', example: '$Z / image', note: '和文本模型 tokens 分开计价。' },
      { category: '视频', unit: '任务、秒数或 provider 单位', example: '$N / second', note: '建议等稳定 provider 路径后再公开。' },
    ],
  },
  principles: [
    { title: '包含用量', body: '套餐可以包含一笔初始余额。用完后，根据套餐规则停止请求或进入 pay-as-you-go。' },
    { title: '扣减顺序', body: '建议先扣 included balance，再扣 bonus balance，最后扣 prepaid balance。这个顺序要能被客服解释清楚。' },
    { title: '毛利护栏', body: '公开注册前，每个公开价格都应覆盖 provider 成本、支付手续费、退款风险和安全毛利。' },
    { title: '账单和支持', body: 'Usage 记录应展示模型、token 数、媒体单位、扣费和请求后的剩余余额。' },
  ],
}

const copy = computed(() => locale.value.startsWith('zh') ? zhCopy : enCopy)

onMounted(() => {
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>
