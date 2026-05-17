<template>
  <span
    class="inline-flex items-center rounded-md px-2 py-0.5 font-mono text-[11px] font-bold leading-5 ring-1"
    :class="methodClass"
  >
    {{ method }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ApiMethod } from '@/config/apiReference'

const props = defineProps<{
  method: ApiMethod
}>()

const methodClass = computed(() => {
  const classes: Record<ApiMethod, string> = {
    GET: 'bg-emerald-50 text-emerald-700 ring-emerald-200',
    POST: 'bg-blue-50 text-blue-700 ring-blue-200',
    PUT: 'bg-violet-50 text-violet-700 ring-violet-200',
    PATCH: 'bg-amber-50 text-amber-700 ring-amber-200',
    DELETE: 'bg-rose-50 text-rose-700 ring-rose-200',
  }
  return classes[props.method]
})
</script>
