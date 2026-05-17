<template>
  <DocsLayout
    :site-name="siteName"
    :site-logo="siteLogo"
    @open-sidebar="mobileSidebarOpen = true"
  >
    <DocsSidebar
      :groups="resolvedSidebarGroups"
      :mobile-open="mobileSidebarOpen"
      @close="mobileSidebarOpen = false"
      @select="selectSection"
    />

    <main class="min-w-0">
      <EndpointHeader :endpoint="endpoint" />
      <AuthorizationSection :auth="endpoint.auth" />
      <ParametersSection :parameters="endpoint.parameters" />
      <ResponseSchema :responses="endpoint.responses" />

      <section class="tg-api-section">
        <h2 class="tg-api-section-title">Related endpoints</h2>
        <div class="mt-4 grid gap-3 sm:grid-cols-2">
          <a
            v-for="link in relatedLinks"
            :key="`${link.href}-${link.title}`"
            :href="link.href"
            class="rounded-xl border border-gray-200 bg-white p-4 transition hover:border-gray-300 hover:bg-gray-50"
            @click.prevent="selectSection(link.href)"
          >
            <div class="flex items-center gap-2">
              <MethodBadge v-if="link.method" :method="link.method" />
              <span class="text-sm font-semibold text-gray-950">{{ link.title }}</span>
            </div>
          </a>
        </div>
      </section>
    </main>

    <aside class="min-w-0 lg:sticky lg:top-20 lg:self-start">
      <CodeExampleTabs title="Request example" :examples="endpoint.examples" />
      <section class="mt-4 overflow-hidden rounded-xl border border-gray-200 bg-white">
        <div class="flex items-center justify-between gap-3 border-b border-gray-200 px-3 py-2.5">
          <p class="text-[11px] font-semibold uppercase tracking-[0.16em] text-gray-400">Response examples</p>
          <StatusCodeTabs v-model="activeResponseStatus" :statuses="responseStatuses" />
        </div>
        <pre class="max-h-[420px] overflow-auto bg-[#0f172a] p-4 text-[13px] leading-6 text-gray-100"><code>{{ activeResponseExample }}</code></pre>
      </section>
    </aside>
  </DocsLayout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import AuthorizationSection from './AuthorizationSection.vue'
import CodeExampleTabs from './CodeExampleTabs.vue'
import DocsLayout from './DocsLayout.vue'
import DocsSidebar from './DocsSidebar.vue'
import EndpointHeader from './EndpointHeader.vue'
import MethodBadge from './MethodBadge.vue'
import ParametersSection from './ParametersSection.vue'
import ResponseSchema from './ResponseSchema.vue'
import StatusCodeTabs from './StatusCodeTabs.vue'
import type { ApiEndpointConfig, ApiSidebarGroup, ApiSidebarItem } from '@/config/apiReference'

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
const endpointAliases: Record<string, string> = {
  endpoint: 'chat-completions',
}

const endpointList = computed(() => props.endpoints?.length ? props.endpoints : props.endpoint ? [props.endpoint] : [])
const endpoint = computed(() => endpointList.value.find((item) => item.id === activeEndpointId.value) ?? endpointList.value[0])
const endpointIds = computed(() => new Set(endpointList.value.map((item) => item.id)))
const responseStatuses = computed(() => endpoint.value?.responses.map((response) => response.status) ?? [])

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
  return resolvedSidebarGroups.value
    .flatMap((group) => group.items)
    .filter((item) => !item.active && endpointIds.value.has(item.href.replace(/^#/, '')))
    .slice(0, 4)
})

function normalizeEndpointId(id: string) {
  return endpointAliases[id] ?? id
}

function syncEndpointFromHash() {
  const hashId = normalizeEndpointId(window.location.hash.replace(/^#/, ''))
  activeEndpointId.value = endpointIds.value.has(hashId) ? hashId : endpointList.value[0]?.id ?? ''
}

function selectSection(href: string) {
  const id = normalizeEndpointId(href.replace(/^#/, ''))
  if (!endpointIds.value.has(id)) {
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
</script>

<style>
.tg-api-section {
  @apply border-b border-gray-200 py-7;
}

.tg-api-section-title {
  @apply text-xl font-semibold tracking-tight text-gray-950;
}
</style>
