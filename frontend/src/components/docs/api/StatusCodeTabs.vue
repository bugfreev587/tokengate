<template>
  <div class="inline-flex rounded-lg border border-gray-200 bg-gray-50 p-1">
    <button
      v-for="status in statuses"
      :key="status"
      type="button"
      class="rounded-md px-3 py-1.5 font-mono text-xs font-bold transition"
      :class="status === modelValue ? 'bg-white text-gray-950 shadow-sm' : 'text-gray-500 hover:text-gray-950'"
      @click="$emit('update:modelValue', status)"
    >
      {{ status }}
    </button>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  statuses: number[]
  modelValue: number
}>()

defineEmits<{
  'update:modelValue': [status: number]
}>()
</script>
