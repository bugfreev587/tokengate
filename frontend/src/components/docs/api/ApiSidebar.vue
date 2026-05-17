<template>
  <aside
    class="api-sidebar"
    :class="mobileOpen ? 'fixed inset-0 z-40 block bg-white/96 p-4 backdrop-blur lg:static lg:z-auto lg:block lg:bg-transparent lg:p-0' : 'hidden lg:block'"
  >
    <div class="mb-4 flex items-center justify-between lg:hidden">
      <p class="text-sm font-semibold text-slate-950">API navigation</p>
      <button
        type="button"
        class="rounded-full border border-slate-200 px-3 py-1.5 text-sm font-medium text-slate-700"
        @click="$emit('close')"
      >
        Close
      </button>
    </div>

    <nav class="sticky top-24 max-h-[calc(100vh-7rem)] overflow-y-auto pr-2">
      <div v-for="group in groups" :key="group.title" class="mb-7">
        <p class="mb-2 px-2 text-[11px] font-semibold uppercase tracking-[0.16em] text-slate-400">
          {{ group.title }}
        </p>
        <a
          v-for="item in group.items"
          :key="`${group.title}-${item.title}`"
          :href="item.href"
          class="flex items-center justify-between gap-3 rounded-xl px-2.5 py-2 text-sm transition"
          :class="item.active ? 'bg-slate-100 text-slate-950' : 'text-slate-600 hover:bg-slate-50 hover:text-slate-950'"
          @click="$emit('close')"
        >
          <span class="truncate">{{ item.title }}</span>
          <span
            v-if="item.method"
            class="shrink-0 rounded-md px-1.5 py-0.5 font-mono text-[10px] font-bold"
            :class="methodClass(item.method)"
          >
            {{ item.method }}
          </span>
        </a>
      </div>
    </nav>
  </aside>
</template>

<script setup lang="ts">
import type { ApiMethod, ApiSidebarGroup } from '@/config/apiReference'

defineProps<{
  groups: ApiSidebarGroup[]
  mobileOpen: boolean
}>()

defineEmits<{
  close: []
}>()

function methodClass(method: ApiMethod) {
  const classes: Record<ApiMethod, string> = {
    GET: 'bg-emerald-50 text-emerald-700 ring-1 ring-emerald-200',
    POST: 'bg-blue-50 text-blue-700 ring-1 ring-blue-200',
    PATCH: 'bg-amber-50 text-amber-700 ring-1 ring-amber-200',
    DELETE: 'bg-rose-50 text-rose-700 ring-1 ring-rose-200',
  }
  return classes[method]
}
</script>
