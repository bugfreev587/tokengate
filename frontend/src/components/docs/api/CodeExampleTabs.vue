<template>
  <section class="overflow-hidden rounded-xl border border-gray-200 bg-[#0f172a] shadow-sm">
    <div class="flex items-center justify-between gap-3 border-b border-white/10 bg-[#111827] px-3 py-2.5">
      <p class="text-[11px] font-semibold uppercase tracking-[0.16em] text-gray-400">{{ title }}</p>
      <button
        type="button"
        class="rounded-md border border-white/10 px-2.5 py-1 text-xs font-semibold text-gray-300 transition hover:bg-white/10 hover:text-white"
        @click="copyActiveCode"
      >
        {{ copied ? 'Copied' : 'Copy' }}
      </button>
    </div>

    <div class="flex gap-1 overflow-x-auto border-b border-white/10 bg-[#0b1120] px-2 py-2">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        type="button"
        class="rounded-md px-2.5 py-1.5 text-xs font-semibold transition"
        :class="tab.key === activeTab ? 'bg-white text-gray-950' : 'text-gray-400 hover:bg-white/10 hover:text-white'"
        @click="activeTab = tab.key"
      >
        {{ tab.label }}
      </button>
    </div>

    <pre class="max-h-[500px] overflow-auto p-4 text-[13px] leading-6 text-gray-100"><code>{{ activeCode }}</code></pre>
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
const copied = ref(false)
const activeCode = computed(() => props.examples[activeTab.value])

async function copyActiveCode() {
  await navigator.clipboard.writeText(activeCode.value)
  copied.value = true
  window.setTimeout(() => {
    copied.value = false
  }, 1400)
}
</script>
