<template>
  <section class="tg-api-section">
    <h2 class="tg-api-section-title">Parameters</h2>
    <p class="mt-2 text-sm leading-6 text-gray-600">
      Request fields accepted by this endpoint. Unsupported fields may be ignored or forwarded depending on the upstream provider.
    </p>

    <div class="mt-4 overflow-hidden rounded-xl border border-gray-200 bg-white">
      <div v-if="parameters.length" class="divide-y divide-gray-200">
        <div
          v-for="parameter in parameters"
          :key="`${parameter.location}-${parameter.name}`"
          class="grid gap-3 px-4 py-4 md:grid-cols-[minmax(160px,0.55fr)_110px_110px_1fr]"
        >
          <div>
            <code class="font-mono text-sm font-semibold text-gray-950">{{ parameter.name }}</code>
            <span v-if="parameter.required" class="ml-2 text-xs font-semibold text-rose-600">required</span>
          </div>
          <span class="text-xs font-semibold uppercase tracking-[0.12em] text-gray-400">{{ parameter.location }}</span>
          <code class="font-mono text-sm text-gray-500">{{ parameter.type }}</code>
          <p class="text-sm leading-6 text-gray-600">{{ parameter.description }}</p>
        </div>
      </div>
      <div v-else class="px-4 py-5 text-sm text-gray-600">
        This endpoint does not accept query or body parameters.
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import type { ApiParameter } from '@/config/apiReference'

defineProps<{
  parameters: ApiParameter[]
}>()
</script>
