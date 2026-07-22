<template>
  <header class="border-b border-stone-200/80 bg-stone-50/90 backdrop-blur dark:border-dark-800 dark:bg-dark-950/90">
    <nav
      :aria-label="t('home.nav.product')"
      class="mx-auto flex h-[4.5rem] max-w-7xl items-center justify-between gap-4 px-5 sm:px-8 lg:px-10"
    >
      <RouterLink
        to="/home"
        class="flex min-w-0 items-center gap-3 rounded-lg focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-600"
      >
        <span
          class="grid h-9 w-9 shrink-0 place-items-center overflow-hidden rounded-xl border border-stone-200 bg-white shadow-[inset_0_1px_0_rgba(255,255,255,0.7)] dark:border-dark-700 dark:bg-dark-900"
        >
          <img :src="siteLogo || '/logo.png'" :alt="siteName" class="h-full w-full object-contain" />
        </span>
        <span class="hidden truncate text-sm font-bold tracking-tight text-gray-950 sm:block dark:text-white">
          {{ siteName }}
        </span>
      </RouterLink>

      <div class="hidden items-center gap-7 text-sm text-gray-600 md:flex dark:text-dark-300">
        <a
          href="#capacity-modes"
          class="transition-colors hover:text-gray-950 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-600 dark:hover:text-white"
        >
          {{ t('home.nav.product') }}
        </a>
        <RouterLink
          to="/pricing"
          class="transition-colors hover:text-gray-950 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-600 dark:hover:text-white"
        >
          {{ t('home.nav.pricing') }}
        </RouterLink>
        <RouterLink
          v-if="!docUrlExternal"
          :to="docUrl"
          class="transition-colors hover:text-gray-950 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-600 dark:hover:text-white"
        >
          {{ t('home.nav.docs') }}
        </RouterLink>
        <a
          v-else
          :href="docUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="transition-colors hover:text-gray-950 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-600 dark:hover:text-white"
        >
          {{ t('home.nav.docs') }}
        </a>
      </div>

      <div class="flex shrink-0 items-center gap-1.5 sm:gap-2">
        <LocaleSwitcher />

        <button
          type="button"
          class="grid h-9 w-9 place-items-center rounded-full text-gray-500 transition-colors hover:bg-stone-200/70 hover:text-gray-950 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-600 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white"
          :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          :aria-label="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          @click="emit('toggleTheme')"
        >
          <Icon v-if="isDark" name="sun" size="sm" :stroke-width="1.5" />
          <Icon v-else name="moon" size="sm" :stroke-width="1.5" />
        </button>

        <RouterLink
          v-if="isAuthenticated"
          :to="dashboardPath"
          class="inline-flex items-center gap-2 rounded-full bg-gray-900 py-2 pl-2 pr-3 text-xs font-semibold text-white transition duration-300 ease-out hover:-translate-y-0.5 hover:bg-gray-800 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-600 focus-visible:ring-offset-2 active:translate-y-0 dark:bg-white dark:text-gray-950 dark:hover:bg-gray-100 dark:focus-visible:ring-offset-dark-950"
        >
          <span
            v-if="userInitial"
            class="grid h-5 w-5 place-items-center rounded-full bg-emerald-700 text-[10px] text-white dark:bg-emerald-600"
          >
            {{ userInitial }}
          </span>
          {{ t('home.dashboard') }}
        </RouterLink>
        <RouterLink
          v-else
          to="/login"
          class="inline-flex items-center rounded-full bg-gray-900 px-4 py-2 text-xs font-semibold text-white transition duration-300 ease-out hover:-translate-y-0.5 hover:bg-gray-800 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-600 focus-visible:ring-offset-2 active:translate-y-0 dark:bg-white dark:text-gray-950 dark:hover:bg-gray-100 dark:focus-visible:ring-offset-dark-950"
        >
          {{ t('home.login') }}
        </RouterLink>
      </div>
    </nav>
  </header>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  siteName: string
  siteLogo: string
  docUrl: string
  docUrlExternal: boolean
  isDark: boolean
  isAuthenticated: boolean
  dashboardPath: string
  userInitial: string
}>()

const emit = defineEmits<{
  toggleTheme: []
}>()

const { t } = useI18n()
</script>
