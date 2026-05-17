<template>
  <section class="tg-api-section">
    <div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <h2 class="tg-api-section-title">Response Body</h2>
        <p class="mt-2 text-sm leading-6 text-gray-600">
          Response schema for the selected status code.
        </p>
      </div>

      <StatusCodeTabs v-model="activeStatus" :statuses="statuses" />
    </div>

    <div class="overflow-hidden rounded-xl border border-gray-200 bg-white">
      <div class="border-b border-gray-200 bg-gray-50 px-4 py-3">
        <div class="flex flex-wrap items-center gap-3">
          <span class="rounded-md bg-white px-2 py-0.5 font-mono text-xs font-bold text-gray-800 ring-1 ring-gray-200">
            {{ activeResponse?.status }}
          </span>
          <span class="text-sm font-semibold text-gray-950">{{ activeResponse?.label }}</span>
        </div>
        <p class="mt-2 text-sm leading-6 text-gray-600">{{ activeResponse?.description }}</p>
      </div>

      <div class="divide-y divide-gray-200">
        <div
          v-for="field in activeResponse?.fields"
          :key="field.name"
          class="grid gap-3 px-4 py-4 md:grid-cols-[minmax(180px,0.6fr)_120px_1fr]"
        >
          <div>
            <code class="font-mono text-sm font-semibold text-gray-950">{{ field.name }}</code>
            <span v-if="field.required" class="ml-2 text-xs font-semibold text-rose-600">required</span>
          </div>
          <code class="font-mono text-sm text-gray-500">{{ field.type }}</code>
          <p class="text-sm leading-6 text-gray-600">{{ field.description }}</p>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import StatusCodeTabs from './StatusCodeTabs.vue'
import type { ApiResponse } from '@/config/apiReference'

const props = defineProps<{
  responses: ApiResponse[]
}>()

const activeStatus = ref(props.responses[0]?.status ?? 200)
const statuses = computed(() => props.responses.map((response) => response.status))
const activeResponse = computed(() => props.responses.find((response) => response.status === activeStatus.value) ?? props.responses[0])

watch(() => props.responses, (responses) => {
  activeStatus.value = responses[0]?.status ?? 200
})
</script>
