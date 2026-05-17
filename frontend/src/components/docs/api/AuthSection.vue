<template>
  <section class="tg-api-section">
    <h2 class="tg-api-section-title">Authorization</h2>

    <div class="rounded-2xl border border-slate-200 bg-white shadow-sm">
      <div class="grid gap-4 border-b border-slate-200 p-5 sm:grid-cols-[180px_1fr]">
        <div>
          <p class="text-sm font-semibold text-slate-950">
            {{ auth.header }}<span v-if="auth.required" class="text-rose-500">*</span>
          </p>
          <p class="mt-1 text-xs text-slate-500">In header</p>
        </div>
        <div>
          <code class="inline-flex rounded-lg border border-slate-200 bg-slate-50 px-2.5 py-1 font-mono text-sm text-slate-800">
            Bearer &lt;token&gt;
          </code>
          <p class="mt-3 text-sm leading-7 text-slate-600">
            {{ auth.description }}
          </p>
        </div>
      </div>
      <div class="grid gap-3 p-5 sm:grid-cols-3">
        <div>
          <p class="text-xs font-medium uppercase tracking-[0.14em] text-slate-400">Type</p>
          <p class="mt-1 text-sm font-medium text-slate-800">{{ auth.type }}</p>
        </div>
        <div>
          <p class="text-xs font-medium uppercase tracking-[0.14em] text-slate-400">Required</p>
          <p class="mt-1 text-sm font-medium text-slate-800">{{ auth.required ? 'Yes' : 'No' }}</p>
        </div>
        <div>
          <p class="text-xs font-medium uppercase tracking-[0.14em] text-slate-400">Header</p>
          <p class="mt-1 font-mono text-sm font-medium text-slate-800">{{ auth.header }}</p>
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
