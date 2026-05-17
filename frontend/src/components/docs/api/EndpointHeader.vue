<template>
  <section id="endpoint" class="border-b border-slate-200 pb-8">
    <div class="mb-6 flex flex-wrap items-center gap-2 text-sm text-slate-500">
      <RouterLink to="/docs" class="hover:text-slate-950">Docs</RouterLink>
      <span>/</span>
      <span>API Reference</span>
      <span>/</span>
      <span class="text-slate-950">{{ endpoint.section }}</span>
    </div>

    <p class="mb-3 text-xs font-semibold uppercase tracking-[0.18em] text-slate-400">
      {{ endpoint.section }}
    </p>
    <h1 class="text-4xl font-semibold tracking-tight text-slate-950 sm:text-5xl">
      {{ endpoint.title }}
    </h1>
    <p class="mt-4 max-w-3xl text-base leading-8 text-slate-600">
      {{ endpoint.description }}
    </p>

    <div class="mt-7 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <p class="mb-3 text-xs font-medium uppercase tracking-[0.14em] text-slate-400">Base URL</p>
      <code class="block overflow-x-auto rounded-xl bg-slate-50 px-4 py-3 font-mono text-sm text-slate-800">
        {{ endpoint.baseUrl }}
      </code>
    </div>

    <div class="mt-4 flex flex-col gap-3 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm sm:flex-row sm:items-center sm:justify-between">
      <div class="flex min-w-0 items-center gap-3">
        <span
          class="rounded-lg px-2.5 py-1 font-mono text-xs font-bold"
          :class="methodClass(endpoint.method)"
        >
          {{ endpoint.method }}
        </span>
        <code class="min-w-0 truncate font-mono text-sm font-semibold text-slate-950">
          {{ endpoint.path }}
        </code>
      </div>
      <button
        type="button"
        class="inline-flex items-center justify-center rounded-xl border border-slate-200 bg-slate-50 px-4 py-2 text-sm font-semibold text-slate-500 transition hover:border-slate-300 hover:bg-white"
        @click="showRunHint = true"
      >
        Run
      </button>
    </div>

    <p v-if="showRunHint" class="mt-3 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800">
      Interactive execution is not enabled yet. Use the code examples on the right to call this endpoint.
    </p>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { ApiEndpointConfig, ApiMethod } from '@/config/apiReference'

defineProps<{
  endpoint: ApiEndpointConfig
}>()

const showRunHint = ref(false)

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
