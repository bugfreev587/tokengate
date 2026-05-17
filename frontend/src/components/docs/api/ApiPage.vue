<template>
  <div class="min-h-screen bg-white text-slate-950">
    <header class="sticky top-0 z-30 border-b border-slate-200 bg-white/90 backdrop-blur-xl">
      <div class="mx-auto flex h-16 max-w-[1500px] items-center justify-between gap-4 px-4 sm:px-6">
        <div class="flex items-center gap-4">
          <button
            type="button"
            class="inline-flex rounded-xl border border-slate-200 p-2 text-slate-600 lg:hidden"
            aria-label="Open API navigation"
            @click="mobileSidebarOpen = true"
          >
            <span class="h-4 w-4 font-mono text-sm leading-4">=</span>
          </button>

          <RouterLink to="/home" class="flex items-center gap-3">
            <span class="flex h-9 w-9 items-center justify-center overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
              <img :src="siteLogo || '/logo.png'" alt="TokenGate" class="h-full w-full object-contain" />
            </span>
            <span class="text-base font-semibold tracking-tight">{{ siteName }}</span>
          </RouterLink>
        </div>

        <nav class="hidden items-center gap-6 text-sm font-medium text-slate-600 xl:flex">
          <RouterLink to="/home" class="hover:text-slate-950">Overview</RouterLink>
          <RouterLink to="/docs" class="text-slate-950">API Reference</RouterLink>
          <RouterLink to="/pricing" class="hover:text-slate-950">Pricing</RouterLink>
          <RouterLink to="/support" class="hover:text-slate-950">Resources</RouterLink>
        </nav>

        <div class="flex items-center gap-3">
          <div class="hidden min-w-[220px] items-center gap-2 rounded-xl border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-500 xl:flex">
            <span class="font-mono text-xs">/</span>
            <span class="flex-1">Search docs</span>
            <span class="rounded-md border border-slate-200 bg-white px-1.5 py-0.5 font-mono text-[10px] text-slate-400">Cmd K</span>
          </div>
          <RouterLink
            to="/login"
            class="rounded-xl bg-slate-950 px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-slate-800"
          >
            Dashboard
          </RouterLink>
        </div>
      </div>
    </header>

    <div class="mx-auto grid max-w-[1500px] gap-8 px-4 py-8 sm:px-6 lg:grid-cols-[280px_minmax(0,1fr)] xl:grid-cols-[280px_minmax(0,1fr)_420px]">
      <ApiSidebar
        :groups="resolvedSidebarGroups"
        :mobile-open="mobileSidebarOpen"
        @close="mobileSidebarOpen = false"
        @select="selectSection"
      />

      <main class="min-w-0">
        <EndpointHeader :endpoint="endpoint" />
        <AuthSection :auth="endpoint.auth" />
        <ResponseSchema :responses="endpoint.responses" />

        <section class="tg-api-section">
          <h2 class="tg-api-section-title">Related endpoints</h2>
          <div class="grid gap-3 sm:grid-cols-2">
            <a
              v-for="link in relatedLinks"
              :key="`${link.href}-${link.title}`"
              :href="link.href"
              class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm transition hover:border-slate-300 hover:shadow-md"
              @click.prevent="selectSection(link.href)"
            >
              <div class="flex items-center gap-2">
                <span
                  v-if="link.method"
                  class="rounded-md px-1.5 py-0.5 font-mono text-[10px] font-bold"
                  :class="methodClass(link.method)"
                >
                  {{ link.method }}
                </span>
                <span class="text-sm font-semibold text-slate-950">{{ link.title }}</span>
              </div>
            </a>
          </div>
        </section>
      </main>

      <aside class="min-w-0 lg:col-start-2 xl:sticky xl:top-24 xl:col-start-auto xl:self-start">
        <CodeTabs title="Request example" :examples="endpoint.examples" />
        <section class="mt-4 overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
          <div class="border-b border-slate-200 px-4 py-3">
            <p class="text-xs font-semibold uppercase tracking-[0.16em] text-slate-400">Response examples</p>
          </div>
          <div class="flex border-b border-slate-200 bg-slate-50 p-1">
            <button
              v-for="response in endpoint.responses"
              :key="response.status"
              type="button"
              class="rounded-lg px-3 py-1.5 font-mono text-xs font-bold transition"
              :class="response.status === activeResponseStatus ? 'bg-white text-slate-950 shadow-sm' : 'text-slate-500 hover:text-slate-950'"
              @click="activeResponseStatus = response.status"
            >
              {{ response.status }}
            </button>
          </div>
          <pre class="max-h-[420px] overflow-auto bg-slate-950 p-5 text-sm leading-7 text-slate-100"><code>{{ activeResponseExample }}</code></pre>
        </section>
      </aside>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import ApiSidebar from './ApiSidebar.vue'
import AuthSection from './AuthSection.vue'
import CodeTabs from './CodeTabs.vue'
import EndpointHeader from './EndpointHeader.vue'
import ResponseSchema from './ResponseSchema.vue'
import type { ApiEndpointConfig, ApiMethod, ApiSidebarGroup, ApiSidebarItem } from '@/config/apiReference'

const props = defineProps<{
  endpoint?: ApiEndpointConfig
  endpoints?: ApiEndpointConfig[]
  sidebarGroups: ApiSidebarGroup[]
  siteName: string
  siteLogo?: string
}>()

const mobileSidebarOpen = ref(false)
const activeEndpointId = ref('')
const activeResponseStatus = ref(200)

const endpointList = computed(() => props.endpoints?.length ? props.endpoints : props.endpoint ? [props.endpoint] : [])
const endpoint = computed(() => {
  return endpointList.value.find((item) => item.id === activeEndpointId.value) ?? endpointList.value[0]
})
const endpointIds = computed(() => new Set(endpointList.value.map((item) => item.id)))

const resolvedSidebarGroups = computed<ApiSidebarGroup[]>(() => {
  return props.sidebarGroups.map((group) => ({
    ...group,
    items: group.items.map((item) => ({
      ...item,
      active: item.href.replace(/^#/, '') === endpoint.value?.id,
    })),
  }))
})

const activeResponseExample = computed(() => {
  return endpoint.value?.responses.find((response) => response.status === activeResponseStatus.value)?.example ?? ''
})

const relatedLinks = computed<ApiSidebarItem[]>(() => {
  return resolvedSidebarGroups.value.flatMap((group) => group.items).filter((item) => !item.active && endpointIds.value.has(item.href.replace(/^#/, ''))).slice(0, 4)
})

function getHashEndpointId() {
  return window.location.hash.replace(/^#/, '')
}

function syncEndpointFromHash() {
  const hashId = getHashEndpointId()
  activeEndpointId.value = endpointIds.value.has(hashId) ? hashId : endpointList.value[0]?.id ?? ''
}

function selectSection(href: string) {
  const id = href.replace(/^#/, '')
  if (!endpointIds.value.has(id)) {
    window.location.hash = id
    mobileSidebarOpen.value = false
    return
  }

  activeEndpointId.value = id
  window.history.replaceState(null, '', `${window.location.pathname}#${id}`)
  window.scrollTo({ top: 0, behavior: 'smooth' })
  mobileSidebarOpen.value = false
}

watch(endpoint, (nextEndpoint) => {
  activeResponseStatus.value = nextEndpoint?.responses[0]?.status ?? 200
})

onMounted(() => {
  syncEndpointFromHash()
  window.addEventListener('hashchange', syncEndpointFromHash)
})

onBeforeUnmount(() => {
  window.removeEventListener('hashchange', syncEndpointFromHash)
})

function methodClass(method: ApiMethod) {
  const classes: Record<ApiMethod, string> = {
    GET: 'bg-emerald-50 text-emerald-700 ring-1 ring-emerald-200',
    POST: 'bg-blue-50 text-blue-700 ring-1 ring-blue-200',
    PATCH: 'bg-amber-50 text-amber-700 ring-1 ring-amber-200',
    DELETE: 'bg-rose-50 text-rose-700 ring-1 ring-rose-200',
  }
  return classes[method]
}
</script>

<style>
.tg-api-section {
  @apply border-b border-slate-200 py-8;
}

.tg-api-section-title {
  @apply text-2xl font-semibold tracking-tight text-slate-950;
}
</style>
