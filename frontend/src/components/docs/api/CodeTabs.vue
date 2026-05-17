<template>
  <section class="overflow-hidden rounded-2xl border border-slate-200 bg-slate-950 shadow-sm">
    <div class="flex items-center justify-between border-b border-white/10 bg-slate-900 px-4 py-3">
      <p class="text-xs font-semibold uppercase tracking-[0.16em] text-slate-400">{{ title }}</p>
      <div class="flex overflow-x-auto rounded-lg border border-white/10 bg-slate-950 p-1">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          type="button"
          class="rounded-md px-2.5 py-1 text-xs font-semibold transition"
          :class="tab.key === activeTab ? 'bg-white text-slate-950' : 'text-slate-400 hover:text-white'"
          @click="activeTab = tab.key"
        >
          {{ tab.label }}
        </button>
      </div>
    </div>

    <pre class="max-h-[520px] overflow-auto p-5 text-sm leading-7 text-slate-100"><code>{{ activeCode }}</code></pre>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { ApiExamples } from '@/config/apiReference'

type TabKey = keyof ApiExamples

const props = defineProps<{
  title: string
  examples: ApiExamples
}>()

const tabs: { key: TabKey; label: string }[] = [
  { key: 'curl', label: 'cURL' },
  { key: 'node', label: 'Node.js' },
  { key: 'python', label: 'Python' },
  { key: 'go', label: 'Go' },
  { key: 'java', label: 'Java' },
]

const activeTab = ref<TabKey>('curl')
const activeCode = computed(() => props.examples[activeTab.value])
</script>
