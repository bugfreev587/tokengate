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
        <p class="text-xs font-semibold uppercase tracking-[0.2em] text-gray-400">CLI Setup</p>
        <h1 class="mt-3 max-w-3xl text-3xl font-semibold tracking-tight text-gray-950 sm:text-4xl">
          Claude Code statusline
        </h1>
        <p class="mt-4 max-w-3xl text-base leading-7 text-gray-600">
          Add a TokenGate-aware status line to Claude Code so each session shows model, project, context usage, today's spend, and budget progress without leaving the terminal.
        </p>

        <div class="mt-6 grid gap-3 sm:grid-cols-3">
          <div class="border-t border-gray-200 pt-4">
            <p class="text-sm font-semibold text-gray-950">Default mode</p>
            <p class="mt-2 text-sm leading-6 text-gray-600">TokenGate mode is selected when no mode is configured.</p>
          </div>
          <div class="border-t border-gray-200 pt-4">
            <p class="text-sm font-semibold text-gray-950">Switch in command</p>
            <p class="mt-2 text-sm leading-6 text-gray-600">Use <code class="tg-inline-code">--mode tokengate</code> or <code class="tg-inline-code">--mode claude</code>.</p>
          </div>
          <div class="border-t border-gray-200 pt-4">
            <p class="text-sm font-semibold text-gray-950">Quiet fallback</p>
            <p class="mt-2 text-sm leading-6 text-gray-600">If TokenGate cannot be reached, Claude Code still gets a usable status line.</p>
          </div>
        </div>
      </section>

      <section id="install" class="tg-doc-section">
        <h2 class="tg-doc-heading">Install</h2>
        <p class="mt-4 text-sm leading-7 text-gray-600">
          Download the script into Claude Code's config directory and make it executable.
        </p>
        <CodePanel
          id="install-script"
          title="Terminal"
          :code="installScript"
          :copied-id="copiedSnippetId"
          @copy="copySnippet"
        />
      </section>

      <section id="configure" class="tg-doc-section">
        <h2 class="tg-doc-heading">Configure Claude Code</h2>
        <p class="mt-4 text-sm leading-7 text-gray-600">
          Add the command to <code class="tg-inline-code">~/.claude/settings.json</code>. The default mode is <code class="tg-inline-code">tokengate</code>.
        </p>
        <CodePanel
          id="settings-default"
          title="Default TokenGate mode"
          subtitle="~/.claude/settings.json"
          :code="settingsDefault"
          :copied-id="copiedSnippetId"
          @copy="copySnippet"
        />
      </section>

      <section id="mode-selection" class="tg-doc-section">
        <h2 class="tg-doc-heading">Mode selection</h2>
        <p class="mt-4 text-sm leading-7 text-gray-600">
          Command arguments have the highest priority, then <code class="tg-inline-code">TOKENGATE_STATUSLINE_MODE</code>, then the default <code class="tg-inline-code">tokengate</code>.
        </p>
        <CodePanel
          id="settings-mode-command"
          title="Choose mode in the command"
          :code="settingsModeCommand"
          :copied-id="copiedSnippetId"
          @copy="copySnippet"
        />
        <CodePanel
          id="settings-mode-env"
          title="Choose mode with env"
          :code="settingsModeEnv"
          :copied-id="copiedSnippetId"
          @copy="copySnippet"
        />
      </section>

      <section id="tokengate-mode" class="tg-doc-section">
        <h2 class="tg-doc-heading">TokenGate mode</h2>
        <p class="mt-4 text-sm leading-7 text-gray-600">
          TokenGate mode is designed for Claude Code sessions routed through <code class="tg-inline-code">https://api.tokengate.to</code>. It highlights TokenGate spend and budget progress.
        </p>
        <div class="mt-5 overflow-hidden rounded-xl border border-gray-200">
          <div class="border-b border-gray-200 bg-gray-50 px-4 py-3 text-sm font-semibold text-gray-950">
            Example output
          </div>
          <pre class="overflow-x-auto bg-[#0f172a] p-4 text-[13px] leading-6 text-gray-100"><code>{{ tokengateExample }}</code></pre>
        </div>
      </section>

      <section id="claude-mode" class="tg-doc-section">
        <h2 class="tg-doc-heading">Claude mode</h2>
        <p class="mt-4 text-sm leading-7 text-gray-600">
          Claude mode preserves the original Claude Code-oriented view, including context remaining and Claude OAuth usage windows when available.
        </p>
        <CodePanel
          id="claude-mode-command"
          title="Claude compatibility mode"
          :code="claudeModeCommand"
          :copied-id="copiedSnippetId"
          @copy="copySnippet"
        />
      </section>

      <section id="environment" class="tg-doc-section">
        <h2 class="tg-doc-heading">Environment variables</h2>
        <div class="mt-4 overflow-hidden rounded-xl border border-gray-200">
          <div class="grid border-b border-gray-200 bg-gray-50 px-4 py-3 text-xs font-semibold uppercase tracking-[0.14em] text-gray-500 sm:grid-cols-[250px_minmax(0,1fr)]">
            <span>Name</span>
            <span>Purpose</span>
          </div>
          <div v-for="item in environmentRows" :key="item.name" class="grid gap-2 border-b border-gray-200 px-4 py-4 text-sm last:border-b-0 sm:grid-cols-[250px_minmax(0,1fr)]">
            <code class="tg-code-line">{{ item.name }}</code>
            <p class="leading-6 text-gray-600">{{ item.body }}</p>
          </div>
        </div>
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
        <p class="text-sm font-semibold text-gray-950">Mode priority</p>
        <ol class="mt-3 space-y-3 text-sm leading-6 text-gray-600">
          <li>1. Command argument: <code class="tg-inline-code">--mode tokengate</code></li>
          <li>2. Settings env: <code class="tg-inline-code">TOKENGATE_STATUSLINE_MODE</code></li>
          <li>3. Default: <code class="tg-inline-code">tokengate</code></li>
        </ol>
      </div>

      <div class="mt-4 rounded-xl border border-gray-200 bg-gray-50 p-4">
        <p class="text-sm font-semibold text-gray-950">Need base CLI setup?</p>
        <p class="mt-2 text-sm leading-6 text-gray-600">
          Configure Claude Code to route through TokenGate before expecting budget data in the status line.
        </p>
        <RouterLink to="/docs/cli" class="mt-4 inline-flex rounded-lg bg-gray-950 px-3.5 py-2 text-sm font-semibold text-white transition hover:bg-gray-800">
          Open CLI Setup
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
    href: item.href.startsWith('#') ? `/docs/cli/statusline${item.href}` : item.href,
    active: item.href === '/docs/cli/statusline',
  })),
})))

const installScript = `mkdir -p ~/.claude
curl -o ~/.claude/tokengate-statusline.sh \\
  https://raw.githubusercontent.com/bugfreev587/tokengate/main/statusline/tokengate-statusline.sh
chmod +x ~/.claude/tokengate-statusline.sh`

const settingsDefault = `{
  "statusLine": {
    "type": "command",
    "command": "sh ~/.claude/tokengate-statusline.sh"
  }
}`

const settingsModeCommand = `{
  "statusLine": {
    "type": "command",
    "command": "sh ~/.claude/tokengate-statusline.sh --mode tokengate"
  }
}

{
  "statusLine": {
    "type": "command",
    "command": "sh ~/.claude/tokengate-statusline.sh --mode claude"
  }
}`

const settingsModeEnv = `{
  "env": {
    "TOKENGATE_STATUSLINE_MODE": "tokengate",
    "ANTHROPIC_BASE_URL": "https://api.tokengate.to",
    "ANTHROPIC_AUTH_TOKEN": "<tokengate-api-key>"
  },
  "statusLine": {
    "type": "command",
    "command": "sh ~/.claude/tokengate-statusline.sh"
  }
}`

const tokengateExample = `Claude Sonnet | token-gate@main | ctx 42k/200k 21% | $1.28 today | month ●●●○○○ $38/$100 38% | day ●○○○○○ $4/$20 20%
Claude Sonnet | token-gate@main | ctx 42k/200k 21% | TokenGate unavailable`

const claudeModeCommand = `{
  "statusLine": {
    "type": "command",
    "command": "sh ~/.claude/tokengate-statusline.sh --mode claude"
  }
}`

const environmentRows = [
  {
    name: 'TOKENGATE_STATUSLINE_MODE',
    body: 'Optional mode selector. Use tokengate or claude. Command arguments override this value.',
  },
  {
    name: 'ANTHROPIC_BASE_URL',
    body: 'TokenGate gateway root, for example https://api.tokengate.to.',
  },
  {
    name: 'TOKENGATE_API_KEY',
    body: 'Preferred TokenGate API key for statusline data. The script also reads ANTHROPIC_AUTH_TOKEN and ANTHROPIC_API_KEY.',
  },
  {
    name: 'TOKENGATE_STATUSLINE_POLL',
    body: 'Cache TTL in seconds between TokenGate polls. Default is 60.',
  },
  {
    name: 'TOKENGATE_STATUSLINE_BARS',
    body: 'Number of progress dots in budget bars. Default is 6.',
  },
]

const troubleshooting = [
  {
    title: 'TokenGate unavailable',
    body: 'Check ANTHROPIC_BASE_URL, the TokenGate API key, and whether your gateway exposes /v1/statusline or /v1/usage.',
  },
  {
    title: 'Mode does not change',
    body: 'A --mode argument in the statusLine command overrides TOKENGATE_STATUSLINE_MODE from settings env.',
  },
  {
    title: 'jq is missing',
    body: 'Install jq locally. The script uses jq to parse Claude Code status JSON and TokenGate usage responses.',
  },
  {
    title: 'The line is too long',
    body: 'Set COLUMNS to your terminal width or lower TOKENGATE_STATUSLINE_BARS to reduce the budget bar length.',
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
