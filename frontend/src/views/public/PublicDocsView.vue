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
          <RouterLink to="/home" class="text-sm font-medium text-slate-500 hover:text-slate-900 dark:text-dark-300 dark:hover:text-white">
            {{ copy.nav.home }}
          </RouterLink>
          <RouterLink to="/login" class="rounded-full bg-slate-950 px-4 py-2 text-sm font-semibold text-white transition hover:bg-slate-800 dark:bg-white dark:text-slate-950 dark:hover:bg-slate-200">
            {{ copy.nav.signIn }}
          </RouterLink>
        </div>
      </div>
    </header>

    <main class="mx-auto max-w-6xl px-5 py-10 sm:py-14">
      <section class="grid gap-8 lg:grid-cols-[1.1fr_0.9fr] lg:items-center">
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
            <RouterLink to="/keys" class="btn btn-primary">
              {{ copy.cta.createKey }}
            </RouterLink>
            <a
              href="https://github.com/bugfreev587/tokengate/blob/main/docs/TOKENGATE_QUICKSTART.md"
              target="_blank"
              rel="noopener noreferrer"
              class="btn btn-secondary"
            >
              {{ copy.cta.githubDocs }}
            </a>
          </div>
        </div>

        <div class="overflow-hidden rounded-3xl border border-slate-200 bg-slate-950 p-5 shadow-2xl shadow-primary-900/10 dark:border-dark-700">
          <div class="mb-4 flex items-center gap-2">
            <span class="h-3 w-3 rounded-full bg-red-400"></span>
            <span class="h-3 w-3 rounded-full bg-yellow-400"></span>
            <span class="h-3 w-3 rounded-full bg-green-400"></span>
          </div>
          <pre class="overflow-x-auto text-sm leading-7 text-slate-100"><code>{{ copy.code }}</code></pre>
        </div>
      </section>

      <section class="mt-12 grid gap-4 md:grid-cols-3">
        <article v-for="card in copy.cards" :key="card.title" class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm dark:border-dark-800 dark:bg-dark-900">
          <p class="text-sm font-semibold text-primary-600 dark:text-primary-400">{{ card.kicker }}</p>
          <h2 class="mt-3 text-xl font-bold">{{ card.title }}</h2>
          <p class="mt-3 text-sm leading-7 text-slate-600 dark:text-dark-300">{{ card.body }}</p>
        </article>
      </section>

      <section class="mt-12 grid gap-6 lg:grid-cols-[0.9fr_1.1fr]">
        <div class="rounded-3xl border border-slate-200 bg-white p-6 dark:border-dark-800 dark:bg-dark-900">
          <h2 class="text-2xl font-bold">{{ copy.billing.title }}</h2>
          <p class="mt-3 text-sm leading-7 text-slate-600 dark:text-dark-300">{{ copy.billing.body }}</p>
          <dl class="mt-6 space-y-4">
            <div v-for="item in copy.billing.items" :key="item.term" class="rounded-2xl bg-slate-50 p-4 dark:bg-dark-800">
              <dt class="font-semibold">{{ item.term }}</dt>
              <dd class="mt-1 text-sm leading-6 text-slate-600 dark:text-dark-300">{{ item.definition }}</dd>
            </div>
          </dl>
        </div>

        <div class="rounded-3xl border border-slate-200 bg-white p-6 dark:border-dark-800 dark:bg-dark-900">
          <h2 class="text-2xl font-bold">{{ copy.troubleshooting.title }}</h2>
          <div class="mt-6 overflow-hidden rounded-2xl border border-slate-200 dark:border-dark-700">
            <table class="w-full divide-y divide-slate-200 text-left text-sm dark:divide-dark-700">
              <thead class="bg-slate-50 text-xs uppercase tracking-wider text-slate-500 dark:bg-dark-800 dark:text-dark-300">
                <tr>
                  <th class="px-4 py-3">{{ copy.troubleshooting.error }}</th>
                  <th class="px-4 py-3">{{ copy.troubleshooting.meaning }}</th>
                  <th class="px-4 py-3">{{ copy.troubleshooting.action }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100 dark:divide-dark-800">
                <tr v-for="row in copy.troubleshooting.rows" :key="row.error">
                  <td class="px-4 py-3 font-mono font-semibold text-primary-600 dark:text-primary-400">{{ row.error }}</td>
                  <td class="px-4 py-3 text-slate-600 dark:text-dark-300">{{ row.meaning }}</td>
                  <td class="px-4 py-3 text-slate-600 dark:text-dark-300">{{ row.action }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </section>

      <section class="mt-12 rounded-3xl border border-slate-200 bg-white p-6 dark:border-dark-800 dark:bg-dark-900">
        <h2 class="text-2xl font-bold">{{ copy.faq.title }}</h2>
        <div class="mt-6 grid gap-4 md:grid-cols-2">
          <article v-for="item in copy.faq.items" :key="item.q" class="rounded-2xl bg-slate-50 p-5 dark:bg-dark-800">
            <h3 class="font-semibold">{{ item.q }}</h3>
            <p class="mt-2 text-sm leading-7 text-slate-600 dark:text-dark-300">{{ item.a }}</p>
          </article>
        </div>
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
  nav: { home: 'Home', signIn: 'Sign in' },
  eyebrow: 'Developer docs',
  title: 'Use TokenGate in five minutes',
  subtitle: 'Create an API key, call OpenAI-compatible or Anthropic-compatible endpoints, and verify usage, token cost, and balance changes from the dashboard.',
  cta: { createKey: 'Create an API key', githubDocs: 'Open GitHub docs' },
  code: `export TOKENGATE_BASE_URL="https://your-backend-domain"
export TOKENGATE_API_KEY="sk-..."

curl "$TOKENGATE_BASE_URL/v1/chat/completions" \\
  -H "Authorization: Bearer $TOKENGATE_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "gpt-4.1-mini",
    "messages": [{"role": "user", "content": "Reply with exactly: hello"}]
  }'`,
  cards: [
    { kicker: 'Step 1', title: 'Create a key', body: 'Open API Keys, create a user key, and keep it private. It acts as the bearer token for downstream apps.' },
    { kicker: 'Step 2', title: 'Send one request', body: 'Use the backend Railway domain for gateway endpoints. Do not send API traffic to the Vercel frontend domain.' },
    { kicker: 'Step 3', title: 'Check metering', body: 'After a successful request, Last Used, Usage records, dashboard totals, and balance should update.' },
  ],
  billing: {
    title: 'Billing language',
    body: 'TokenGate separates provider usage units from settlement. This keeps model pricing transparent while users see a simple balance.',
    items: [
      { term: 'Tokens', definition: 'Text model usage units. Public pricing should show input and output prices per 1M tokens.' },
      { term: 'Image units', definition: 'Image generation units, usually per image or provider-native output unit.' },
      { term: 'Video units', definition: 'Video generation units, usually per job, second, or provider-native unit.' },
      { term: 'Balance', definition: 'The account value that successful usage settles against. This is clearer than calling everything credits.' },
    ],
  },
  troubleshooting: {
    title: 'Common API errors',
    error: 'Error',
    meaning: 'Meaning',
    action: 'Action',
    rows: [
      { error: '401', meaning: 'API key missing or invalid.', action: 'Create, rotate, or re-enable the API key.' },
      { error: '403', meaning: 'Key is valid but not allowed.', action: 'Check group, plan, balance, or model access.' },
      { error: '404', meaning: 'Wrong endpoint or host.', action: 'Use the backend base URL and documented gateway path.' },
      { error: '405', meaning: 'Usually hitting the frontend as an API.', action: 'Point SDKs and curl to the Railway backend domain.' },
      { error: '429', meaning: 'Rate, quota, or balance limit reached.', action: 'Wait, top up, or change plan limits.' },
    ],
  },
  faq: {
    title: 'FAQ',
    items: [
      { q: 'Should customers think in credits or tokens?', a: 'For V1, use balance plus transparent metered usage. Tokens are usage units; balance is what pays for them.' },
      { q: 'Can I use OpenAI SDKs?', a: 'Yes. Point the SDK base URL to the TokenGate backend and use the TokenGate API key as the bearer key.' },
      { q: 'Where should VITE_API_BASE_URL point?', a: 'It must point to the backend API prefix and include /api/v1. It should not point to the Vercel frontend.' },
      { q: 'How do I verify production after deploy?', a: 'Run tools/tokengate_smoke_test.sh with a real TokenGate API key, then check Usage and Dashboard.' },
    ],
  },
}

const zhCopy = {
  nav: { home: '首页', signIn: '登录' },
  eyebrow: '开发者文档',
  title: '五分钟接入 TokenGate',
  subtitle: '创建 API 密钥，调用 OpenAI-compatible 或 Anthropic-compatible 端点，并在 Dashboard 里确认用量、Token 成本和余额变化。',
  cta: { createKey: '创建 API 密钥', githubDocs: '打开 GitHub 文档' },
  code: enCopy.code,
  cards: [
    { kicker: '第一步', title: '创建密钥', body: '打开 API Keys，创建一个用户 API key，并妥善保存。它就是下游应用调用 TokenGate 的 bearer token。' },
    { kicker: '第二步', title: '发送请求', body: '网关请求要使用 Railway 后端域名。不要把 API 流量发到 Vercel 前端域名。' },
    { kicker: '第三步', title: '检查计量', body: '请求成功后，Last Used、Usage 记录、Dashboard 汇总和余额都应该更新。' },
  ],
  billing: {
    title: '计费语言',
    body: 'TokenGate 把 provider 用量单位和结算余额分开。这样模型价格透明，用户侧也能保持简单的余额心智。',
    items: [
      { term: 'Tokens', definition: '文本模型用量单位。公开价格建议展示每 1M input/output tokens 的价格。' },
      { term: 'Image units', definition: '图片生成单位，通常按图片数量或 provider 原生输出单位计费。' },
      { term: 'Video units', definition: '视频生成单位，通常按任务、秒数或 provider 原生单位计费。' },
      { term: 'Balance', definition: '成功请求最终扣减的账户余额。V1 里比把所有东西都叫 credits 更清晰。' },
    ],
  },
  troubleshooting: {
    title: '常见 API 错误',
    error: '错误',
    meaning: '含义',
    action: '处理方式',
    rows: [
      { error: '401', meaning: 'API key 缺失或无效。', action: '创建、轮换或重新启用 API key。' },
      { error: '403', meaning: 'Key 有效，但权限不足。', action: '检查分组、套餐、余额或模型访问权限。' },
      { error: '404', meaning: '端点或域名错误。', action: '使用后端 base URL 和文档里的网关路径。' },
      { error: '405', meaning: '通常是把前端域名当 API 调了。', action: '把 SDK/curl 指向 Railway 后端域名。' },
      { error: '429', meaning: '触发限速、额度或余额限制。', action: '等待、充值或调整套餐限制。' },
    ],
  },
  faq: {
    title: 'FAQ',
    items: [
      { q: '用户应该理解 credits 还是 tokens？', a: 'V1 建议使用“余额 + 透明用量”。Tokens 是文本模型用量单位，balance 才是最终结算对象。' },
      { q: '可以用 OpenAI SDK 吗？', a: '可以。把 SDK base URL 指向 TokenGate 后端，并使用 TokenGate API key 作为 bearer key。' },
      { q: 'VITE_API_BASE_URL 应该指向哪里？', a: '必须指向后端 API 前缀，并包含 /api/v1。不要指向 Vercel 前端域名。' },
      { q: '部署后怎么验证生产环境？', a: '用真实 TokenGate API key 运行 tools/tokengate_smoke_test.sh，然后检查 Usage 和 Dashboard。' },
    ],
  },
}

const copy = computed(() => locale.value.startsWith('zh') ? zhCopy : enCopy)

onMounted(() => {
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>
