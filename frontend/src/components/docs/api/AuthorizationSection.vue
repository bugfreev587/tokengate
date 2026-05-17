<template>
  <section class="tg-api-section">
    <h2 class="tg-api-section-title">Authorization</h2>
    <p class="mt-2 text-sm leading-6 text-gray-600">
      TokenGate uses API keys from the customer dashboard. Send the key as a bearer token.
    </p>

    <div class="mt-4 overflow-hidden rounded-xl border border-gray-200 bg-white">
      <div class="grid gap-4 border-b border-gray-200 px-4 py-4 md:grid-cols-[170px_1fr]">
        <div>
          <p class="font-mono text-sm font-semibold text-gray-950">{{ auth.header }}</p>
          <p class="mt-1 text-xs text-gray-500">Header</p>
        </div>
        <div>
          <code class="rounded-md border border-gray-200 bg-gray-50 px-2.5 py-1 font-mono text-sm text-gray-800">
            Bearer &lt;token&gt;
          </code>
          <p class="mt-3 text-sm leading-6 text-gray-600">{{ auth.description }}</p>
        </div>
      </div>
      <div class="grid gap-4 px-4 py-4 text-sm md:grid-cols-3">
        <div>
          <p class="text-xs font-semibold uppercase tracking-[0.14em] text-gray-400">Type</p>
          <p class="mt-1 font-medium text-gray-900">{{ auth.type }}</p>
        </div>
        <div>
          <p class="text-xs font-semibold uppercase tracking-[0.14em] text-gray-400">Required</p>
          <p class="mt-1 font-medium text-gray-900">{{ auth.required ? 'Yes' : 'No' }}</p>
        </div>
        <div>
          <p class="text-xs font-semibold uppercase tracking-[0.14em] text-gray-400">Scope</p>
          <p class="mt-1 font-medium text-gray-900">Customer API key</p>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import type { ApiAuthConfig } from '@/config/apiReference'

defineProps<{
  auth: ApiAuthConfig
}>()
</script>
