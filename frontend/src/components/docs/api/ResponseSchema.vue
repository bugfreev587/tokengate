<template>
  <section class="tg-api-section">
    <div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <h2 class="tg-api-section-title">Response Body</h2>
        <p class="mt-2 text-sm leading-7 text-slate-600">
          Schema fields for the selected HTTP status code.
        </p>
      </div>

      <div class="flex rounded-xl border border-slate-200 bg-slate-50 p-1">
        <button
          v-for="response in responses"
          :key="response.status"
          type="button"
          class="rounded-lg px-3 py-1.5 font-mono text-xs font-bold transition"
          :class="response.status === activeStatus ? 'bg-white text-slate-950 shadow-sm' : 'text-slate-500 hover:text-slate-950'"
          @click="activeStatus = response.status"
        >
          {{ response.status }}
        </button>
      </div>
    </div>

    <div class="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
      <div class="border-b border-slate-200 bg-slate-50 px-5 py-4">
        <div class="flex flex-wrap items-center gap-3">
          <span class="rounded-lg bg-white px-2.5 py-1 font-mono text-xs font-bold text-slate-800 ring-1 ring-slate-200">
            {{ activeResponse?.status }}
          </span>
          <span class="text-sm font-semibold text-slate-950">{{ activeResponse?.label }}</span>
        </div>
        <p class="mt-2 text-sm leading-6 text-slate-600">{{ activeResponse?.description }}</p>
      </div>

      <div class="divide-y divide-slate-200">
        <div
          v-for="field in activeResponse?.fields"
          :key="field.name"
          class="grid gap-3 px-5 py-4 md:grid-cols-[minmax(180px,0.7fr)_minmax(130px,0.35fr)_1fr]"
        >
          <div>
            <code class="font-mono text-sm font-semibold text-slate-950">{{ field.name }}</code>
            <span v-if="field.required" class="ml-2 text-xs font-semibold text-rose-500">required</span>
          </div>
          <code class="font-mono text-sm text-slate-500">{{ field.type }}</code>
          <p class="text-sm leading-6 text-slate-600">{{ field.description }}</p>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { ApiResponse } from '@/config/apiReference'

const props = defineProps<{
  responses: ApiResponse[]
}>()

const activeStatus = ref(props.responses[0]?.status ?? 200)
const activeResponse = computed(() => props.responses.find((response) => response.status === activeStatus.value) ?? props.responses[0])
</script>
