<template>
  <aside
    :class="mobileOpen ? 'fixed inset-0 z-40 block bg-white/95 p-4 backdrop-blur lg:static lg:z-auto lg:block lg:bg-transparent lg:p-0' : 'hidden lg:block'"
  >
    <div class="mb-4 flex items-center justify-between lg:hidden">
      <p class="text-sm font-semibold text-gray-950">API Reference</p>
      <button
        type="button"
        class="rounded-lg border border-gray-200 px-3 py-1.5 text-sm font-medium text-gray-700"
        @click="$emit('close')"
      >
        Close
      </button>
    </div>

    <nav class="sticky top-20 max-h-[calc(100vh-5.5rem)] overflow-y-auto pr-3">
      <div v-for="group in groups" :key="group.title" class="mb-6">
        <p class="mb-2 px-2 text-[11px] font-semibold uppercase tracking-[0.18em] text-gray-400">
          {{ group.title }}
        </p>
        <template
          v-for="item in group.items"
          :key="`${group.title}-${item.title}`"
        >
          <RouterLink
            v-if="isRouteLink(item.href)"
            :to="item.href"
            class="flex items-center justify-between gap-3 rounded-lg px-2.5 py-2 text-sm transition"
            :class="itemClass(item.active)"
            @click="$emit('close')"
          >
            <span class="truncate">{{ item.title }}</span>
            <MethodBadge v-if="item.method" :method="item.method" />
          </RouterLink>
          <a
            v-else
            :href="item.href"
            class="flex items-center justify-between gap-3 rounded-lg px-2.5 py-2 text-sm transition"
            :class="itemClass(item.active)"
            @click.prevent="$emit('select', item.href)"
          >
            <span class="truncate">{{ item.title }}</span>
            <MethodBadge v-if="item.method" :method="item.method" />
          </a>
        </template>
      </div>
    </nav>
  </aside>
</template>

<script setup lang="ts">
import MethodBadge from './MethodBadge.vue'
import type { ApiSidebarGroup } from '@/config/apiReference'

defineProps<{
  groups: ApiSidebarGroup[]
  mobileOpen: boolean
}>()

defineEmits<{
  close: []
  select: [href: string]
}>()

function isRouteLink(href: string) {
  return href.startsWith('/')
}

function itemClass(active?: boolean) {
  return active ? 'bg-gray-100 text-gray-950' : 'text-gray-600 hover:bg-gray-50 hover:text-gray-950'
}
</script>
