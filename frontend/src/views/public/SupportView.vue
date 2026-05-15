<template>
  <div class="min-h-screen bg-slate-50 text-slate-950 dark:bg-dark-950 dark:text-white">
    <header class="border-b border-slate-200 bg-white/90 backdrop-blur dark:border-dark-800 dark:bg-dark-900/90">
      <div class="mx-auto flex max-w-5xl items-center justify-between gap-4 px-5 py-4">
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
          <RouterLink to="/login" class="rounded-full bg-slate-950 px-4 py-2 text-sm font-semibold text-white transition hover:bg-slate-800 dark:bg-white dark:text-slate-950 dark:hover:bg-slate-200">
            {{ copy.nav.signIn }}
          </RouterLink>
        </div>
      </div>
    </header>

    <main class="mx-auto max-w-5xl px-5 py-12 sm:py-16">
      <section class="rounded-[2rem] border border-slate-200 bg-white p-7 shadow-sm dark:border-dark-800 dark:bg-dark-900 sm:p-10">
        <p class="text-sm font-semibold uppercase tracking-[0.24em] text-primary-600 dark:text-primary-400">
          {{ copy.eyebrow }}
        </p>
        <h1 class="mt-4 max-w-3xl text-4xl font-black tracking-tight sm:text-5xl">
          {{ copy.title }}
        </h1>
        <p class="mt-5 max-w-3xl text-base leading-8 text-slate-600 dark:text-dark-300">
          {{ copy.subtitle }}
        </p>

        <div class="mt-8 rounded-3xl bg-slate-50 p-6 dark:bg-dark-800">
          <p class="text-sm font-semibold text-slate-500 dark:text-dark-300">{{ copy.primaryContact }}</p>
          <a
            v-if="contactHref"
            :href="contactHref"
            target="_blank"
            rel="noopener noreferrer"
            class="mt-3 inline-flex break-all text-xl font-bold text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300"
          >
            {{ contactInfo }}
          </a>
          <p v-else class="mt-3 break-words text-xl font-bold text-slate-900 dark:text-white">
            {{ contactInfo || copy.noContact }}
          </p>
          <p class="mt-3 text-sm leading-6 text-slate-500 dark:text-dark-300">
            {{ contactInfo ? copy.contactHint : copy.noContactHint }}
          </p>
        </div>
      </section>

      <section class="mt-8 grid gap-4 md:grid-cols-3">
        <article v-for="item in copy.beforeContact" :key="item.title" class="rounded-2xl border border-slate-200 bg-white p-6 dark:border-dark-800 dark:bg-dark-900">
          <p class="text-sm font-semibold text-primary-600 dark:text-primary-400">{{ item.kicker }}</p>
          <h2 class="mt-3 text-xl font-bold">{{ item.title }}</h2>
          <p class="mt-3 text-sm leading-7 text-slate-600 dark:text-dark-300">{{ item.body }}</p>
        </article>
      </section>

      <section class="mt-8 rounded-3xl border border-slate-200 bg-white p-6 dark:border-dark-800 dark:bg-dark-900">
        <h2 class="text-2xl font-bold">{{ copy.template.title }}</h2>
        <p class="mt-3 text-sm leading-7 text-slate-600 dark:text-dark-300">{{ copy.template.body }}</p>
        <pre class="mt-5 overflow-x-auto rounded-2xl bg-slate-950 p-5 text-sm leading-7 text-slate-100"><code>{{ copy.template.text }}</code></pre>
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
const contactInfo = computed(() => (appStore.cachedPublicSettings?.contact_info || '').trim())
const contactHref = computed(() => {
  const value = contactInfo.value
  if (!value) return ''
  if (/^https?:\/\//i.test(value)) return value
  if (/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) return `mailto:${value}`
  return ''
})

const enCopy = {
  nav: { docs: 'Docs', signIn: 'Sign in' },
  eyebrow: 'Support',
  title: 'Need help with TokenGate?',
  subtitle: 'Send the smallest useful reproduction and we can debug faster: endpoint, model, request time, request ID if available, and the error shown by your client.',
  primaryContact: 'Primary contact',
  noContact: 'Support contact is not configured yet.',
  contactHint: 'Use this channel for billing, account access, API errors, and production incident reports.',
  noContactHint: 'Admins can configure this in Admin -> Settings -> Site -> Contact Info.',
  beforeContact: [
    { kicker: 'Step 1', title: 'Check API host', body: 'Gateway requests must go to the Railway backend domain, not the Vercel frontend domain.' },
    { kicker: 'Step 2', title: 'Check key status', body: 'Confirm the API key is active, belongs to the right user or group, and has enough balance or plan access.' },
    { kicker: 'Step 3', title: 'Attach evidence', body: 'Include model, timestamp, HTTP status, response body, and whether the request created a Usage record.' },
  ],
  template: {
    title: 'Useful support template',
    body: 'Copy this shape when reporting an API issue. It avoids the painful back-and-forth and gets us to the bug faster.',
    text: `Issue:
Environment: production / local
Endpoint:
Model:
HTTP status:
Approx request time:
TokenGate request ID, if available:
Did Usage update? yes / no
Expected result:
Actual result:`,
  },
}

const zhCopy = {
  nav: { docs: '文档', signIn: '登录' },
  eyebrow: '支持',
  title: '需要 TokenGate 支持？',
  subtitle: '请尽量提供一个最小可复现信息：端点、模型、请求时间、request ID（如果有），以及客户端看到的错误。',
  primaryContact: '主要联系方式',
  noContact: '暂未配置支持联系方式。',
  contactHint: '这个渠道可用于计费、账号访问、API 错误和生产事故反馈。',
  noContactHint: '管理员可在 Admin -> Settings -> Site -> Contact Info 中配置。',
  beforeContact: [
    { kicker: '第一步', title: '检查 API 域名', body: '网关请求必须发到 Railway 后端域名，而不是 Vercel 前端域名。' },
    { kicker: '第二步', title: '检查 Key 状态', body: '确认 API key 处于 active，属于正确用户或分组，并且有足够余额或套餐权限。' },
    { kicker: '第三步', title: '附上证据', body: '请包含模型、时间、HTTP status、响应 body，以及是否生成了 Usage 记录。' },
  ],
  template: {
    title: '推荐的支持模板',
    body: '反馈 API 问题时可以复制这个结构，减少来回沟通，更快定位问题。',
    text: `问题:
环境: production / local
端点:
模型:
HTTP status:
大致请求时间:
TokenGate request ID（如果有）:
Usage 是否更新: yes / no
预期结果:
实际结果:`,
  },
}

const copy = computed(() => locale.value.startsWith('zh') ? zhCopy : enCopy)

onMounted(() => {
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>
