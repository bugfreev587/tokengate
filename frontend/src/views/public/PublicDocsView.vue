<template>
  <div class="docs-page min-h-screen bg-[#07111f] text-slate-100">
    <header class="sticky top-0 z-30 border-b border-white/10 bg-[#07111f]/88 backdrop-blur-xl">
      <div class="mx-auto flex max-w-7xl items-center justify-between gap-4 px-5 py-4">
        <RouterLink to="/home" class="flex items-center gap-3">
          <span class="flex h-10 w-10 items-center justify-center overflow-hidden rounded-xl bg-white shadow-sm ring-1 ring-white/15">
            <img :src="siteLogo || '/logo.png'" alt="TokenGate" class="h-full w-full object-contain" />
          </span>
          <span class="text-base font-semibold">{{ siteName }}</span>
        </RouterLink>

        <div class="flex items-center gap-3">
          <RouterLink to="/home" class="docs-top-link">Home</RouterLink>
          <RouterLink to="/pricing" class="docs-top-link">Pricing</RouterLink>
          <RouterLink to="/support" class="docs-top-link">Support</RouterLink>
          <RouterLink to="/login" class="rounded-full bg-white px-4 py-2 text-sm font-semibold text-slate-950 transition hover:bg-cyan-100">
            Sign in
          </RouterLink>
        </div>
      </div>
    </header>

    <main class="mx-auto grid max-w-7xl gap-8 px-5 py-8 lg:grid-cols-[260px_minmax(0,1fr)]">
      <aside class="hidden lg:block">
        <nav class="sticky top-24 rounded-3xl border border-white/10 bg-white/[0.04] p-4 shadow-2xl shadow-black/20">
          <p class="px-3 text-xs font-bold uppercase tracking-[0.22em] text-cyan-300">Docs</p>
          <div class="mt-4 space-y-5">
            <section v-for="section in navSections" :key="section.title">
              <p class="px-3 text-xs font-semibold uppercase tracking-[0.16em] text-slate-500">{{ section.title }}</p>
              <div class="mt-2 space-y-1">
                <a
                  v-for="item in section.items"
                  :key="item.href"
                  :href="item.href"
                  class="block rounded-xl px-3 py-2 text-sm font-medium text-slate-300 transition hover:bg-white/10 hover:text-white"
                >
                  {{ item.label }}
                </a>
              </div>
            </section>
          </div>
        </nav>
      </aside>

      <div class="min-w-0">
        <section id="overview" class="docs-hero overflow-hidden rounded-[2rem] border border-cyan-300/20 bg-slate-950 p-6 shadow-2xl shadow-cyan-950/30 sm:p-8 lg:p-10">
          <div class="grid gap-8 xl:grid-cols-[1fr_0.86fr] xl:items-center">
            <div>
              <p class="text-sm font-bold uppercase tracking-[0.28em] text-cyan-300">TokenGate Docs</p>
              <h1 class="mt-4 max-w-3xl text-4xl font-black tracking-tight text-white sm:text-5xl">
                Ship AI API access behind one customer key.
              </h1>
              <p class="mt-5 max-w-2xl text-base leading-8 text-slate-300">
                TokenGate is an OpenAI-compatible and Anthropic-compatible gateway for packaging subscribed upstream accounts into a user-facing API product with balance, usage, groups, and model routing.
              </p>
              <div class="mt-8 flex flex-col gap-3 sm:flex-row">
                <RouterLink to="/register" class="docs-primary-button">Create account</RouterLink>
                <RouterLink to="/keys" class="docs-secondary-button">Create API key</RouterLink>
                <a href="#api-reference" class="docs-secondary-button">API Reference</a>
              </div>
            </div>

            <div class="rounded-3xl border border-white/10 bg-black/70 p-5">
              <div class="mb-4 flex items-center justify-between">
                <div class="flex items-center gap-2">
                  <span class="h-3 w-3 rounded-full bg-red-400"></span>
                  <span class="h-3 w-3 rounded-full bg-amber-300"></span>
                  <span class="h-3 w-3 rounded-full bg-emerald-400"></span>
                </div>
                <span class="rounded-full bg-cyan-400/10 px-3 py-1 text-xs font-bold text-cyan-200">OpenAI compatible</span>
              </div>
              <pre class="overflow-x-auto text-sm leading-7 text-slate-100"><code>{{ quickstartCurl }}</code></pre>
            </div>
          </div>
        </section>

        <section id="quickstart" class="docs-section">
          <div class="docs-section-heading">
            <p class="docs-eyebrow">Quickstart</p>
            <h2>Five minute customer setup</h2>
            <p>Follow this exact path for a regular user. Admin-only setup lives in the dashboard; customers only need balance, group access, and an API key.</p>
          </div>

          <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
            <article v-for="step in quickstartSteps" :key="step.title" class="docs-card">
              <div class="docs-step-number">{{ step.number }}</div>
              <h3>{{ step.title }}</h3>
              <p>{{ step.body }}</p>
            </article>
          </div>
        </section>

        <section id="core-concepts" class="docs-section">
          <div class="docs-section-heading">
            <p class="docs-eyebrow">Core concepts</p>
            <h2>How TokenGate decides what a request can use</h2>
            <p>The mental model is intentionally small: users hold keys, keys bind to groups, groups route to upstream accounts, and usage settles against balance.</p>
          </div>

          <div class="grid gap-4 lg:grid-cols-2">
            <article v-for="concept in concepts" :key="concept.title" class="docs-card">
              <p class="font-mono text-xs font-bold uppercase tracking-[0.18em] text-cyan-300">{{ concept.kicker }}</p>
              <h3 class="mt-3">{{ concept.title }}</h3>
              <p>{{ concept.body }}</p>
            </article>
          </div>
        </section>

        <section id="base-url" class="docs-section">
          <div class="docs-section-heading">
            <p class="docs-eyebrow">Base URL and auth</p>
            <h2>Send model traffic to Railway, not Vercel</h2>
            <p>The Vercel domain serves the web dashboard. SDKs, curl, and production apps must call the Railway backend.</p>
          </div>

          <div class="grid gap-4 lg:grid-cols-[0.9fr_1.1fr]">
            <div class="docs-card">
              <h3>Production API base URL</h3>
              <div class="mt-4 rounded-2xl border border-cyan-300/20 bg-slate-950 p-4 font-mono text-sm text-cyan-100">
                {{ apiBaseUrl }}
              </div>
              <p class="mt-4">Use <code>/v1</code> paths directly after this host. For OpenAI SDKs, set <code>baseURL</code> to <code>{{ apiBaseUrl }}/v1</code>.</p>
            </div>

            <div class="docs-card">
              <h3>Required headers</h3>
              <div class="mt-4 overflow-hidden rounded-2xl border border-white/10">
                <table class="docs-table">
                  <thead>
                    <tr>
                      <th>Header</th>
                      <th>Value</th>
                      <th>When</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="header in authHeaders" :key="header.name">
                      <td><code>{{ header.name }}</code></td>
                      <td><code>{{ header.value }}</code></td>
                      <td>{{ header.when }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </section>

        <section id="sdk-examples" class="docs-section">
          <div class="docs-section-heading">
            <p class="docs-eyebrow">SDK examples</p>
            <h2>Use existing OpenAI and Anthropic clients</h2>
            <p>TokenGate is designed to fit into the client libraries users already know. Replace the API key and base URL, then keep normal request shapes.</p>
          </div>

          <div class="grid gap-4 xl:grid-cols-2">
            <article v-for="snippet in sdkSnippets" :key="snippet.title" class="docs-code-panel">
              <div class="docs-code-panel-header">
                <span>{{ snippet.title }}</span>
                <span>{{ snippet.badge }}</span>
              </div>
              <pre><code>{{ snippet.code }}</code></pre>
            </article>
          </div>
        </section>

        <section id="api-reference" class="docs-section">
          <div class="docs-section-heading">
            <p class="docs-eyebrow">API Reference</p>
            <h2>Gateway endpoints</h2>
            <p>These are the customer-facing endpoints a regular user calls with a TokenGate API key. Admin endpoints are intentionally excluded.</p>
          </div>

          <div class="space-y-4">
            <article v-for="endpoint in endpoints" :key="`${endpoint.method} ${endpoint.path}`" :id="endpoint.id" class="docs-endpoint-card">
              <div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
                <div class="min-w-0">
                  <div class="flex flex-wrap items-center gap-3">
                    <span class="docs-method" :class="methodClass(endpoint.method)">{{ endpoint.method }}</span>
                    <code class="break-all font-mono text-sm font-bold text-slate-100">{{ endpoint.path }}</code>
                  </div>
                  <h3 class="mt-4">{{ endpoint.title }}</h3>
                  <p class="mt-2 text-sm leading-7 text-slate-300">{{ endpoint.description }}</p>
                </div>
                <span class="shrink-0 rounded-full border border-white/10 bg-white/5 px-3 py-1 text-xs font-bold text-slate-300">{{ endpoint.compatibility }}</span>
              </div>

              <div class="mt-5 grid gap-4 xl:grid-cols-[0.9fr_1.1fr]">
                <div class="rounded-2xl border border-white/10 bg-white/[0.03] p-4">
                  <p class="text-xs font-bold uppercase tracking-[0.18em] text-slate-500">Important fields</p>
                  <dl class="mt-3 space-y-3">
                    <div v-for="field in endpoint.fields" :key="field.name">
                      <dt class="font-mono text-sm font-semibold text-cyan-200">{{ field.name }}</dt>
                      <dd class="mt-1 text-sm leading-6 text-slate-300">{{ field.description }}</dd>
                    </div>
                  </dl>
                </div>
                <div class="docs-code-panel">
                  <div class="docs-code-panel-header">
                    <span>Example</span>
                    <span>{{ endpoint.exampleLabel }}</span>
                  </div>
                  <pre><code>{{ endpoint.example }}</code></pre>
                </div>
              </div>
            </article>
          </div>
        </section>

        <section id="errors" class="docs-section">
          <div class="docs-section-heading">
            <p class="docs-eyebrow">Errors</p>
            <h2>How to debug failed requests</h2>
            <p>Most production issues are wrong host, wrong key, missing group access, no balance, or an upstream account problem.</p>
          </div>

          <div class="overflow-hidden rounded-3xl border border-white/10 bg-white/[0.04]">
            <table class="docs-table">
              <thead>
                <tr>
                  <th>Status</th>
                  <th>Meaning</th>
                  <th>Next action</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="error in errors" :key="error.status">
                  <td><code>{{ error.status }}</code></td>
                  <td>{{ error.meaning }}</td>
                  <td>{{ error.action }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        <section id="verification" class="docs-section">
          <div class="docs-section-heading">
            <p class="docs-eyebrow">Verification</p>
            <h2>Production smoke checklist</h2>
            <p>Use this after changing groups, upstream accounts, payment settings, or deployment environment variables.</p>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <article class="docs-card">
              <h3>OpenAI group</h3>
              <ul class="docs-checklist">
                <li><code>GET /v1/models</code> returns GPT and image models.</li>
                <li><code>POST /v1/chat/completions</code> returns text.</li>
                <li><code>POST /v1/responses</code> returns text.</li>
                <li><code>POST /v1/images/generations</code> returns image data if image generation is enabled for the group.</li>
              </ul>
            </article>

            <article class="docs-card">
              <h3>Claude group</h3>
              <ul class="docs-checklist">
                <li><code>GET /v1/models</code> returns Claude models.</li>
                <li><code>POST /v1/messages</code> returns Anthropic-compatible text.</li>
                <li><code>POST /v1/chat/completions</code> returns OpenAI-compatible text.</li>
                <li><code>POST /v1/messages/count_tokens</code> may depend on upstream account support.</li>
              </ul>
            </article>
          </div>
        </section>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores/app'

type Method = 'GET' | 'POST'

interface EndpointField {
  name: string
  description: string
}

interface EndpointDoc {
  id: string
  method: Method
  path: string
  title: string
  description: string
  compatibility: string
  exampleLabel: string
  fields: EndpointField[]
  example: string
}

const appStore = useAppStore()

const apiBaseUrl = 'https://tokengate-production.up.railway.app'
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'TokenGate')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')

const navSections = [
  {
    title: 'Start',
    items: [
      { label: 'Overview', href: '#overview' },
      { label: 'Quickstart', href: '#quickstart' },
      { label: 'Base URL and auth', href: '#base-url' },
    ],
  },
  {
    title: 'Guide',
    items: [
      { label: 'Core concepts', href: '#core-concepts' },
      { label: 'SDK examples', href: '#sdk-examples' },
      { label: 'Errors', href: '#errors' },
      { label: 'Verification', href: '#verification' },
    ],
  },
  {
    title: 'API Reference',
    items: [
      { label: 'List models', href: '#list-models' },
      { label: 'Chat completions', href: '#chat-completions' },
      { label: 'Responses', href: '#responses' },
      { label: 'Images', href: '#images' },
      { label: 'Messages', href: '#messages' },
      { label: 'Count tokens', href: '#count-tokens' },
    ],
  },
]

const quickstartCurl = `export TOKENGATE_API_KEY="sk-..."

curl "${apiBaseUrl}/v1/chat/completions" \\
  -H "Authorization: Bearer $TOKENGATE_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "gpt-5.2-chat-latest",
    "messages": [
      { "role": "user", "content": "Say hi in one short sentence." }
    ]
  }'`

const quickstartSteps = [
  {
    number: '1',
    title: 'Get access',
    body: 'Sign up, then make sure the admin has added balance and enabled the groups you need.',
  },
  {
    number: '2',
    title: 'Create a key',
    body: 'Open API Keys, create a TokenGate API key, and bind it to the OpenAI or Claude group you want to use.',
  },
  {
    number: '3',
    title: 'Send a request',
    body: 'Call the Railway backend using standard OpenAI-compatible or Anthropic-compatible request shapes.',
  },
  {
    number: '4',
    title: 'Check usage',
    body: 'Open Usage to confirm model, endpoint, tokens, status, and balance deduction.',
  },
]

const concepts = [
  {
    kicker: 'User',
    title: 'Customer account',
    body: 'A regular user owns API keys, balance, usage history, and billing state. Admins configure upstream capacity and permissions.',
  },
  {
    kicker: 'Key',
    title: 'TokenGate API key',
    body: 'The bearer credential used by apps. A key usually belongs to one group, which determines model routing and billing rules.',
  },
  {
    kicker: 'Group',
    title: 'Access and billing layer',
    body: 'Groups package model access, upstream accounts, RPM, image permission, and rate multipliers. Public groups are available to everyone; exclusive groups require per-user authorization.',
  },
  {
    kicker: 'Balance',
    title: 'Settlement value',
    body: 'Tokens, image units, and future video units are usage units. Balance is the money-like value successful usage settles against.',
  },
]

const authHeaders = [
  { name: 'Authorization', value: 'Bearer sk-...', when: 'All gateway requests' },
  { name: 'Content-Type', value: 'application/json', when: 'All JSON requests' },
  { name: 'anthropic-version', value: '2023-06-01', when: 'Anthropic /v1/messages clients' },
]

const sdkSnippets = [
  {
    title: 'OpenAI SDK',
    badge: 'Chat and Responses',
    code: `import OpenAI from "openai";

const client = new OpenAI({
  apiKey: process.env.TOKENGATE_API_KEY,
  baseURL: "${apiBaseUrl}/v1",
});

const response = await client.chat.completions.create({
  model: "gpt-5.2-chat-latest",
  messages: [{ role: "user", content: "Say hi." }],
});

console.log(response.choices[0]?.message?.content);`,
  },
  {
    title: 'Anthropic-compatible cURL',
    badge: 'Claude messages',
    code: `curl "${apiBaseUrl}/v1/messages" \\
  -H "Authorization: Bearer $TOKENGATE_API_KEY" \\
  -H "Content-Type: application/json" \\
  -H "anthropic-version: 2023-06-01" \\
  -d '{
    "model": "claude-sonnet-4-6",
    "max_tokens": 64,
    "messages": [
      { "role": "user", "content": "Say hi." }
    ]
  }'`,
  },
]

const endpoints: EndpointDoc[] = [
  {
    id: 'list-models',
    method: 'GET',
    path: '/v1/models',
    title: 'List available models',
    description: 'Returns the models available through the API key group. OpenAI keys return GPT/image models; Claude keys return Claude models.',
    compatibility: 'OpenAI style',
    exampleLabel: 'cURL',
    fields: [
      { name: 'data[].id', description: 'Provider model identifier to use in later requests.' },
      { name: 'data[].object', description: 'OpenAI-compatible object type.' },
    ],
    example: `curl "${apiBaseUrl}/v1/models" \\
  -H "Authorization: Bearer $TOKENGATE_API_KEY"`,
  },
  {
    id: 'chat-completions',
    method: 'POST',
    path: '/v1/chat/completions',
    title: 'Create chat completion',
    description: 'OpenAI-compatible chat endpoint. It works for OpenAI groups and also for Claude groups when the group supports chat-compatible dispatch.',
    compatibility: 'OpenAI compatible',
    exampleLabel: 'cURL',
    fields: [
      { name: 'model', description: 'Model ID returned by /v1/models.' },
      { name: 'messages', description: 'Array of role/content messages.' },
      { name: 'max_tokens', description: 'Optional response token cap.' },
    ],
    example: quickstartCurl,
  },
  {
    id: 'responses',
    method: 'POST',
    path: '/v1/responses',
    title: 'Create response',
    description: 'OpenAI Responses endpoint for modern OpenAI clients. Use this for GPT-5 class models and tools-oriented clients.',
    compatibility: 'OpenAI Responses',
    exampleLabel: 'cURL',
    fields: [
      { name: 'model', description: 'OpenAI model ID, for example gpt-5.2.' },
      { name: 'input', description: 'String or structured input accepted by the upstream Responses API.' },
      { name: 'max_output_tokens', description: 'Optional response token cap.' },
    ],
    example: `curl "${apiBaseUrl}/v1/responses" \\
  -H "Authorization: Bearer $TOKENGATE_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "gpt-5.2",
    "input": "Reply with exactly: pong",
    "max_output_tokens": 64
  }'`,
  },
  {
    id: 'images',
    method: 'POST',
    path: '/v1/images/generations',
    title: 'Generate image',
    description: 'OpenAI-compatible image generation. The target group must have image generation enabled.',
    compatibility: 'OpenAI Images',
    exampleLabel: 'cURL',
    fields: [
      { name: 'model', description: 'Image model ID, for example gpt-image-2.' },
      { name: 'prompt', description: 'Text prompt for the image.' },
      { name: 'size', description: 'Provider-supported output size.' },
    ],
    example: `curl "${apiBaseUrl}/v1/images/generations" \\
  -H "Authorization: Bearer $TOKENGATE_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "gpt-image-2",
    "prompt": "A tiny blue dot icon on a white background.",
    "size": "1024x1024",
    "n": 1
  }'`,
  },
  {
    id: 'messages',
    method: 'POST',
    path: '/v1/messages',
    title: 'Create message',
    description: 'Anthropic-compatible messages endpoint for Claude groups. OpenAI groups return 403 unless messages dispatch is explicitly enabled.',
    compatibility: 'Anthropic compatible',
    exampleLabel: 'cURL',
    fields: [
      { name: 'model', description: 'Claude model ID returned by /v1/models.' },
      { name: 'max_tokens', description: 'Required for Anthropic-compatible requests.' },
      { name: 'messages', description: 'Array of Anthropic role/content messages.' },
    ],
    example: sdkSnippets[1].code,
  },
  {
    id: 'count-tokens',
    method: 'POST',
    path: '/v1/messages/count_tokens',
    title: 'Count message tokens',
    description: 'Anthropic-compatible token counting endpoint. Availability depends on the selected upstream account type.',
    compatibility: 'Anthropic compatible',
    exampleLabel: 'cURL',
    fields: [
      { name: 'model', description: 'Claude model ID.' },
      { name: 'messages', description: 'Messages to estimate.' },
    ],
    example: `curl "${apiBaseUrl}/v1/messages/count_tokens" \\
  -H "Authorization: Bearer $TOKENGATE_API_KEY" \\
  -H "Content-Type: application/json" \\
  -H "anthropic-version: 2023-06-01" \\
  -d '{
    "model": "claude-sonnet-4-6",
    "messages": [
      { "role": "user", "content": "Count this." }
    ]
  }'`,
  },
]

const errors = [
  { status: '401', meaning: 'Missing, invalid, disabled, or deleted API key.', action: 'Create or rotate the key in API Keys.' },
  { status: '403', meaning: 'The key is valid but not allowed to use the requested group, model, endpoint, or feature.', action: 'Check group access, image permission, balance, and endpoint compatibility.' },
  { status: '404', meaning: 'Wrong path or non-existent endpoint.', action: 'Use the documented /v1 path on the Railway backend.' },
  { status: '405', meaning: 'Usually an API request sent to the Vercel frontend.', action: `Use ${apiBaseUrl}, not the frontend app domain.` },
  { status: '429', meaning: 'Rate limit, concurrency limit, quota, or balance limit reached.', action: 'Wait, increase limits, add balance, or use another group.' },
  { status: '5xx', meaning: 'Upstream provider or gateway runtime failure.', action: 'Retry with request ID, check Usage, then contact support.' },
]

const methodClass = (method: Method) => {
  return method === 'GET' ? 'docs-method-get' : 'docs-method-post'
}

onMounted(() => {
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>

<style scoped>
.docs-page {
  background:
    radial-gradient(circle at 18% 0%, rgb(20 184 166 / 0.2), transparent 28rem),
    radial-gradient(circle at 82% 8%, rgb(14 165 233 / 0.18), transparent 26rem),
    linear-gradient(180deg, #07111f 0%, #0b1322 46%, #07111f 100%);
}

.docs-top-link {
  @apply hidden text-sm font-semibold text-slate-300 transition hover:text-white sm:inline-flex;
}

.docs-hero {
  position: relative;
}

.docs-hero::before {
  position: absolute;
  inset: 0;
  pointer-events: none;
  content: "";
  background-image:
    linear-gradient(rgb(255 255 255 / 0.04) 1px, transparent 1px),
    linear-gradient(90deg, rgb(255 255 255 / 0.04) 1px, transparent 1px);
  background-size: 36px 36px;
  mask-image: linear-gradient(180deg, black, transparent 85%);
}

.docs-hero > * {
  position: relative;
}

.docs-primary-button {
  @apply inline-flex items-center justify-center rounded-full bg-cyan-300 px-5 py-3 text-sm font-bold text-slate-950 shadow-lg shadow-cyan-950/40 transition hover:bg-cyan-200;
}

.docs-secondary-button {
  @apply inline-flex items-center justify-center rounded-full border border-white/15 bg-white/5 px-5 py-3 text-sm font-bold text-slate-100 transition hover:border-cyan-300/50 hover:bg-cyan-300/10;
}

.docs-section {
  @apply mt-10 scroll-mt-24;
}

.docs-section-heading {
  @apply mb-5 max-w-3xl;
}

.docs-section-heading h2 {
  @apply mt-2 text-3xl font-black tracking-tight text-white;
}

.docs-section-heading p:not(.docs-eyebrow) {
  @apply mt-3 text-sm leading-7 text-slate-300;
}

.docs-eyebrow {
  @apply text-xs font-bold uppercase tracking-[0.24em] text-cyan-300;
}

.docs-card {
  @apply rounded-3xl border border-white/10 bg-white/[0.045] p-5 shadow-xl shadow-black/10;
}

.docs-card h3 {
  @apply text-lg font-bold text-white;
}

.docs-card p,
.docs-card li {
  @apply mt-2 text-sm leading-7 text-slate-300;
}

.docs-step-number {
  @apply mb-4 flex h-9 w-9 items-center justify-center rounded-full bg-cyan-300 text-sm font-black text-slate-950;
}

.docs-code-panel {
  @apply min-w-0 overflow-hidden rounded-3xl border border-white/10 bg-black/70;
}

.docs-code-panel-header {
  @apply flex items-center justify-between gap-4 border-b border-white/10 bg-white/[0.04] px-4 py-3 text-xs font-bold uppercase tracking-[0.14em] text-slate-400;
}

.docs-code-panel pre {
  @apply overflow-x-auto p-4 text-sm leading-7 text-slate-100;
}

.docs-endpoint-card {
  @apply scroll-mt-24 rounded-3xl border border-white/10 bg-white/[0.045] p-5 shadow-xl shadow-black/10;
}

.docs-endpoint-card h3 {
  @apply text-xl font-bold text-white;
}

.docs-method {
  @apply inline-flex rounded-lg px-2.5 py-1 font-mono text-xs font-black tracking-[0.08em];
}

.docs-method-get {
  @apply bg-emerald-400/15 text-emerald-300;
}

.docs-method-post {
  @apply bg-blue-400/15 text-blue-300;
}

.docs-table {
  @apply w-full divide-y divide-white/10 text-left text-sm;
}

.docs-table thead {
  @apply bg-white/[0.04] text-xs uppercase tracking-[0.14em] text-slate-500;
}

.docs-table th,
.docs-table td {
  @apply px-4 py-3 align-top;
}

.docs-table tbody {
  @apply divide-y divide-white/10 text-slate-300;
}

.docs-table code,
.docs-card code,
.docs-checklist code {
  @apply rounded-md border border-white/10 bg-slate-950 px-1.5 py-0.5 font-mono text-cyan-200;
}

.docs-checklist {
  @apply mt-4 space-y-3;
}

.docs-checklist li {
  @apply list-none;
}

.docs-checklist li::before {
  content: "";
  display: inline-block;
  width: 0.42rem;
  height: 0.72rem;
  margin-right: 0.75rem;
  border: solid #67e8f9;
  border-width: 0 2px 2px 0;
  transform: rotate(45deg);
}
</style>
