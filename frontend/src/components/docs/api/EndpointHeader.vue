<template>
  <section id="endpoint" class="pb-8">
    <div class="mb-5 flex flex-wrap items-center gap-2 text-sm text-gray-500">
      <RouterLink to="/docs" class="hover:text-gray-950">Docs</RouterLink>
      <span>/</span>
      <span>API Reference</span>
      <span>/</span>
      <span class="text-gray-950">{{ endpoint.section }}</span>
    </div>

    <p class="mb-2 text-xs font-semibold uppercase tracking-[0.18em] text-gray-400">
      {{ endpoint.section }}
    </p>
    <h1 class="text-3xl font-semibold tracking-tight text-gray-950 sm:text-4xl">
      {{ endpoint.title }}
    </h1>
    <p class="mt-3 max-w-3xl text-[15px] leading-7 text-gray-600">
      {{ endpoint.description }}
    </p>

    <div class="mt-6">
      <p class="mb-2 text-xs font-medium uppercase tracking-[0.14em] text-gray-400">Base URL</p>
      <code class="block overflow-x-auto rounded-lg bg-gray-50 px-3 py-2.5 font-mono text-sm text-gray-800">
        {{ endpoint.baseUrl }}
      </code>
    </div>

    <div class="mt-4 flex flex-col gap-3 rounded-lg bg-gray-50 px-3 py-3 sm:flex-row sm:items-center sm:justify-between">
      <div class="flex min-w-0 items-center gap-3">
        <MethodBadge :method="endpoint.method" />
        <code class="min-w-0 truncate font-mono text-sm font-semibold text-gray-950">
          {{ endpoint.path }}
        </code>
      </div>
      <button
        type="button"
        class="inline-flex items-center justify-center rounded-lg border border-gray-200 bg-gray-50 px-3.5 py-2 text-sm font-semibold text-gray-500 transition hover:border-gray-300 hover:bg-white"
        @click="showRunHint = true"
      >
        Run
      </button>
    </div>

    <p v-if="showRunHint" class="mt-3 rounded-lg border border-amber-200 bg-amber-50 px-3.5 py-2.5 text-sm text-amber-800">
      Interactive execution is not enabled yet. Use the code examples on the right to call this endpoint.
    </p>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import MethodBadge from './MethodBadge.vue'
import type { ApiEndpointConfig } from '@/config/apiReference'

defineProps<{
  endpoint: ApiEndpointConfig
}>()

const showRunHint = ref(false)
</script>
