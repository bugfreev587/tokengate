<template>
  <div class="min-h-screen bg-white text-gray-950">
    <header class="sticky top-0 z-30 border-b border-gray-200 bg-white/95 backdrop-blur">
      <div class="mx-auto flex h-14 max-w-[1500px] items-center justify-between gap-4 px-4 sm:px-6">
        <div class="flex items-center gap-3">
          <button
            type="button"
            class="inline-flex rounded-lg border border-gray-200 p-2 text-gray-600 lg:hidden"
            aria-label="Open API navigation"
            @click="$emit('open-sidebar')"
          >
            <span class="h-4 w-4 font-mono text-sm leading-4">=</span>
          </button>

          <RouterLink to="/home" class="flex items-center gap-2.5">
            <span class="flex h-8 w-8 items-center justify-center overflow-hidden rounded-lg border border-gray-200 bg-white shadow-sm">
              <img :src="siteLogo || '/logo.png'" alt="TokenGate" class="h-full w-full object-contain" />
            </span>
            <span class="text-sm font-semibold tracking-tight">{{ siteName }}</span>
          </RouterLink>
        </div>

        <nav class="hidden items-center gap-6 text-sm font-medium text-gray-600 xl:flex">
          <RouterLink to="/home" class="hover:text-gray-950">Overview</RouterLink>
          <RouterLink to="/docs" class="text-gray-950">API Reference</RouterLink>
          <RouterLink to="/pricing" class="hover:text-gray-950">Pricing</RouterLink>
          <RouterLink to="/support" class="hover:text-gray-950">Resources</RouterLink>
        </nav>

        <div class="flex items-center gap-3">
          <div class="hidden min-w-[220px] items-center gap-2 rounded-lg border border-gray-200 bg-gray-50 px-3 py-1.5 text-sm text-gray-500 xl:flex">
            <span class="font-mono text-xs">/</span>
            <span class="flex-1">Search docs</span>
            <span class="rounded border border-gray-200 bg-white px-1.5 py-0.5 font-mono text-[10px] text-gray-400">Cmd K</span>
          </div>
          <RouterLink
            to="/login"
            class="rounded-lg bg-gray-950 px-3.5 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-gray-800"
          >
            Dashboard
          </RouterLink>
        </div>
      </div>
    </header>

    <div class="mx-auto grid max-w-[1500px] gap-6 px-4 py-7 sm:px-6 lg:grid-cols-[240px_minmax(0,1fr)_340px] xl:grid-cols-[270px_minmax(0,1fr)_430px] xl:gap-8">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  siteName: string
  siteLogo?: string
}>()

defineEmits<{
  'open-sidebar': []
}>()
</script>
