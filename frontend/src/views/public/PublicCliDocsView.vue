<template>
  <DocsLayout
    :site-name="siteName"
    :site-logo="siteLogo"
    @open-sidebar="mobileSidebarOpen = true"
  >
    <DocsSidebar
      :groups="sidebarGroups"
      :mobile-open="mobileSidebarOpen"
      @close="mobileSidebarOpen = false"
      @select="selectSidebarItem"
    />

    <main class="min-w-0">
      <section class="border-b border-gray-200 pb-8">
        <p class="text-xs font-semibold uppercase tracking-[0.2em] text-gray-400">Guide</p>
        <h1 class="mt-3 max-w-3xl text-3xl font-semibold tracking-tight text-gray-950 sm:text-4xl">
          Use Claude Code CLI and Codex CLI with TokenGate
        </h1>
        <p class="mt-4 max-w-3xl text-base leading-7 text-gray-600">
          Point your local coding agent at TokenGate, authenticate with a TokenGate API key, and let TokenGate route the request to the upstream account group you selected.
        </p>

        <div class="mt-6 grid gap-3 sm:grid-cols-3">
          <div class="border-t border-gray-200 pt-4">
            <p class="text-sm font-semibold text-gray-950">1. Create an API key</p>
            <p class="mt-2 text-sm leading-6 text-gray-600">Open Dashboard / API Keys and create a key for the group your CLI should use.</p>
          </div>
          <div class="border-t border-gray-200 pt-4">
            <p class="text-sm font-semibold text-gray-950">2. Use the right base URL</p>
            <p class="mt-2 text-sm leading-6 text-gray-600">Claude Code uses the backend root. Codex CLI uses the OpenAI-compatible <code class="tg-inline-code">/v1</code> base URL.</p>
          </div>
          <div class="border-t border-gray-200 pt-4">
            <p class="text-sm font-semibold text-gray-950">3. Start the CLI</p>
            <p class="mt-2 text-sm leading-6 text-gray-600">Run the CLI from your project after the environment variables or config file are in place.</p>
          </div>
        </div>
      </section>

      <section id="base-url-rules" class="tg-doc-section">
        <h2 class="tg-doc-heading">Base URL rules</h2>
        <div class="mt-4 overflow-hidden rounded-xl border border-gray-200">
          <div class="grid border-b border-gray-200 bg-gray-50 px-4 py-3 text-xs font-semibold uppercase tracking-[0.14em] text-gray-500 sm:grid-cols-[170px_minmax(0,1fr)]">
            <span>Client</span>
            <span>TokenGate base URL</span>
          </div>
          <div class="grid gap-2 border-b border-gray-200 px-4 py-4 text-sm sm:grid-cols-[170px_minmax(0,1fr)]">
            <span class="font-semibold text-gray-950">Claude Code CLI</span>
            <code class="tg-code-line">https://api.tokengate.to</code>
          </div>
          <div class="grid gap-2 px-4 py-4 text-sm sm:grid-cols-[170px_minmax(0,1fr)]">
            <span class="font-semibold text-gray-950">Codex CLI</span>
            <code class="tg-code-line">https://api.tokengate.to/v1</code>
          </div>
        </div>
        <p class="mt-4 text-sm leading-7 text-gray-600">
          Use <code class="tg-inline-code">api.tokengate.to</code> for CLI and SDK traffic. The web app domains are for dashboard/docs pages only. Replace <code class="tg-inline-code">&lt;tokengate-api-key&gt;</code> with a key created in TokenGate.
        </p>
      </section>

      <section id="claude-code" class="tg-doc-section">
        <div class="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <p class="text-xs font-semibold uppercase tracking-[0.18em] text-orange-600">Anthropic compatible</p>
            <h2 class="tg-doc-heading">Claude Code CLI</h2>
          </div>
          <span class="w-fit rounded-full bg-orange-50 px-3 py-1 text-xs font-semibold text-orange-700">Uses /v1/messages</span>
        </div>
        <p class="mt-4 text-sm leading-7 text-gray-600">
          Claude Code reads Anthropic-compatible environment variables. Use the TokenGate backend root as <code class="tg-inline-code">ANTHROPIC_BASE_URL</code>; do not append <code class="tg-inline-code">/v1</code>.
        </p>
        <p class="mt-3 text-sm leading-7 text-gray-600">
          After routing Claude Code through TokenGate, add the
          <RouterLink to="/docs/cli/statusline" class="font-semibold text-gray-950 underline decoration-gray-300 underline-offset-4 hover:decoration-gray-950">
            Claude Code statusline
          </RouterLink>
          to show live spend and budget progress in the terminal.
        </p>

        <CodePanel
          id="claude-shell"
          title="Terminal"
          :code="claudeShell"
          :copied-id="copiedSnippetId"
          @copy="copySnippet"
        />

        <CodePanel
          id="claude-settings"
          title="Persistent settings"
          subtitle="~/.claude/settings.json"
          :code="claudeSettings"
          :copied-id="copiedSnippetId"
          @copy="copySnippet"
        />
      </section>

      <section id="codex-cli" class="tg-doc-section">
        <div class="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <p class="text-xs font-semibold uppercase tracking-[0.18em] text-emerald-600">OpenAI Responses compatible</p>
            <h2 class="tg-doc-heading">Codex CLI</h2>
          </div>
          <span class="w-fit rounded-full bg-emerald-50 px-3 py-1 text-xs font-semibold text-emerald-700">Uses /v1/responses</span>
        </div>
        <p class="mt-4 text-sm leading-7 text-gray-600">
          Codex CLI can use a custom model provider. Store the TokenGate API key in <code class="tg-inline-code">TOKENGATE_API_KEY</code>, then point the provider at TokenGate's <code class="tg-inline-code">/v1</code> endpoint.
        </p>

        <CodePanel
          id="codex-env"
          title="Environment"
          :code="codexEnv"
          :copied-id="copiedSnippetId"
          @copy="copySnippet"
        />

        <CodePanel
          id="codex-config"
          title="Codex config"
          subtitle="~/.codex/config.toml"
          :code="codexConfig"
          :copied-id="copiedSnippetId"
          @copy="copySnippet"
        />
      </section>

      <section id="smoke-test" class="tg-doc-section">
        <h2 class="tg-doc-heading">Smoke test</h2>
        <p class="mt-4 text-sm leading-7 text-gray-600">
          Before opening the coding agent, verify that the key can reach TokenGate and list the models available to its group.
        </p>
        <CodePanel
          id="smoke-test-curl"
          title="curl"
          :code="smokeTest"
          :copied-id="copiedSnippetId"
          @copy="copySnippet"
        />
      </section>

      <section id="troubleshooting" class="tg-doc-section">
        <h2 class="tg-doc-heading">Troubleshooting</h2>
        <div class="mt-4 divide-y divide-gray-200 rounded-xl border border-gray-200">
          <div v-for="item in troubleshooting" :key="item.title" class="p-4">
            <p class="text-sm font-semibold text-gray-950">{{ item.title }}</p>
            <p class="mt-2 text-sm leading-6 text-gray-600">{{ item.body }}</p>
          </div>
        </div>
      </section>
    </main>

    <aside class="min-w-0 lg:sticky lg:top-20 lg:self-start">
      <div class="rounded-xl border border-gray-200 bg-white p-4">
        <p class="text-sm font-semibold text-gray-950">Setup checklist</p>
        <ol class="mt-3 space-y-3 text-sm leading-6 text-gray-600">
          <li>1. Create or select a TokenGate API key.</li>
          <li>2. Confirm the key belongs to a group with the model you want.</li>
          <li>3. Use backend root for Claude Code.</li>
          <li>4. Use backend <code class="tg-inline-code">/v1</code> for Codex CLI.</li>
        </ol>
      </div>

      <div class="mt-4 rounded-xl border border-gray-200 bg-gray-50 p-4">
        <p class="text-sm font-semibold text-gray-950">Where to get exact values</p>
        <p class="mt-2 text-sm leading-6 text-gray-600">
          In Dashboard / API Keys, use the key row's Use Key action. It fills the same templates with your actual TokenGate base URL and API key.
        </p>
        <RouterLink to="/login" class="mt-4 inline-flex rounded-lg bg-gray-950 px-3.5 py-2 text-sm font-semibold text-white transition hover:bg-gray-800">
          Open Dashboard
        </RouterLink>
      </div>
    </aside>
  </DocsLayout>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onBeforeUnmount, onMounted, ref } from 'vue'
import DocsLayout from '@/components/docs/api/DocsLayout.vue'
import DocsSidebar from '@/components/docs/api/DocsSidebar.vue'
import { tokenGateApiSidebarGroups, type ApiSidebarGroup } from '@/config/apiReference'
import { useAppStore } from '@/stores/app'

const CodePanel = defineComponent({
  props: {
    id: { type: String, required: true },
    title: { type: String, required: true },
    subtitle: { type: String, default: '' },
    code: { type: String, required: true },
    copiedId: { type: String, default: undefined },
  },
  emits: ['copy'],
  setup(props, { emit }) {
    return () => h('div', { class: 'mt-5 overflow-hidden rounded-xl border border-gray-200 bg-white' }, [
      h('div', { class: 'flex items-center justify-between gap-3 border-b border-gray-200 bg-gray-50 px-4 py-3' }, [
        h('div', [
          h('p', { class: 'text-sm font-semibold text-gray-950' }, props.title),
          props.subtitle ? h('p', { class: 'mt-1 font-mono text-xs text-gray-500' }, props.subtitle) : null,
        ]),
        h(
          'button',
          {
            type: 'button',
            class: 'rounded-md border border-gray-200 bg-white px-2.5 py-1.5 text-xs font-semibold text-gray-600 transition hover:border-gray-300 hover:text-gray-950',
            onClick: () => emit('copy', props.id, props.code),
          },
          props.copiedId === props.id ? 'Copied' : 'Copy',
        ),
      ]),
      h('pre', { class: 'overflow-x-auto bg-[#0f172a] p-4 text-[13px] leading-6 text-gray-100' }, [
        h('code', props.code),
      ]),
    ])
  },
})

const appStore = useAppStore()
const mobileSidebarOpen = ref(false)
const copiedSnippetId = ref<string>()
let copyTimer: number | undefined

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'TokenGate')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')

const sidebarGroups = computed<ApiSidebarGroup[]>(() => tokenGateApiSidebarGroups.map((group) => ({
  ...group,
  items: group.items.map((item) => ({
    ...item,
    href: item.href.startsWith('#') ? `/docs${item.href}` : item.href,
    active: item.href === '/docs/cli',
  })),
})))

const tokenGateApiBaseUrl = 'https://api.tokengate.to'

const claudeShell = `export ANTHROPIC_BASE_URL="${tokenGateApiBaseUrl}"
export ANTHROPIC_AUTH_TOKEN="<tokengate-api-key>"
export CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1`

const claudeSettings = `{
  "env": {
    "ANTHROPIC_BASE_URL": "${tokenGateApiBaseUrl}",
    "ANTHROPIC_AUTH_TOKEN": "<tokengate-api-key>",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
    "CLAUDE_CODE_ATTRIBUTION_HEADER": "0"
  }
}`

const codexEnv = `export TOKENGATE_API_KEY="<tokengate-api-key>"`

const codexConfig = `model_provider = "tokengate"
model = "gpt-5.4"

[model_providers.tokengate]
name = "TokenGate"
base_url = "${tokenGateApiBaseUrl}/v1"
env_key = "TOKENGATE_API_KEY"
wire_api = "responses"`

const smokeTest = `curl "${tokenGateApiBaseUrl}/v1/models" \\
  -H "Authorization: Bearer <tokengate-api-key>"`

const troubleshooting = [
  {
    title: '401 Unauthorized',
    body: 'Check that the API key is active, copied without extra spaces, and sent as the TokenGate key rather than an upstream provider key.',
  },
  {
    title: 'Model not found',
    body: 'Use a model name enabled for the API key group. The same group controls model access, rate limits, balance, and upstream account selection.',
  },
  {
    title: 'Claude Code cannot count tokens',
    body: 'Claude Code should use an Anthropic-compatible TokenGate group and the backend root URL. Do not use the OpenAI /v1 base URL for Claude Code.',
  },
  {
    title: 'Codex CLI reaches the wrong endpoint',
    body: 'Make sure the provider base_url ends with /v1 and wire_api is set to responses.',
  },
]

function selectSidebarItem(href: string) {
  mobileSidebarOpen.value = false

  if (!href.startsWith('#')) {
    return
  }

  document.querySelector(href)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

async function copySnippet(id: string, code: string) {
  if (!navigator.clipboard) {
    return
  }

  await navigator.clipboard.writeText(code)
  copiedSnippetId.value = id

  if (copyTimer) {
    window.clearTimeout(copyTimer)
  }

  copyTimer = window.setTimeout(() => {
    copiedSnippetId.value = undefined
  }, 1400)
}

onMounted(() => {
  if (!appStore.cachedPublicSettings) {
    appStore.fetchPublicSettings().catch(() => {})
  }
})

onBeforeUnmount(() => {
  if (copyTimer) {
    window.clearTimeout(copyTimer)
  }
})
</script>

<style>
.tg-doc-section {
  @apply border-b border-gray-200 py-8 last:border-b-0;
}

.tg-doc-heading {
  @apply text-2xl font-semibold tracking-tight text-gray-950;
}

.tg-inline-code {
  @apply rounded-md border border-gray-200 bg-gray-50 px-1.5 py-0.5 font-mono text-[0.85em] text-gray-800;
}

.tg-code-line {
  @apply min-w-0 overflow-x-auto rounded-lg bg-gray-50 px-3 py-2 font-mono text-xs leading-6 text-gray-800;
}

.tg-inline-code::selection,
.tg-code-line::selection {
  background-color: #ccfbf1;
  color: #111827;
}
</style>
