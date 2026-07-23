<template>
  <footer class="border-t border-stone-200 bg-stone-50 dark:border-dark-800 dark:bg-dark-950">
    <div
      class="mx-auto flex max-w-7xl flex-col gap-6 px-5 py-9 text-sm text-gray-500 sm:px-8 md:flex-row md:items-center md:justify-between lg:px-10 dark:text-dark-400"
    >
      <p>&copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}</p>

      <nav :aria-label="t('home.nav.product')" class="flex flex-wrap items-center gap-x-6 gap-y-3">
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
        <RouterLink
          to="/support"
          class="transition-colors hover:text-gray-950 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-600 dark:hover:text-white"
        >
          {{ t('home.nav.support') }}
        </RouterLink>
        <RouterLink
          to="/legal/privacy"
          class="transition-colors hover:text-gray-950 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-600 dark:hover:text-white"
        >
          {{ t('home.nav.privacy') }}
        </RouterLink>
      </nav>
    </div>
  </footer>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

defineProps<{
  siteName: string
  currentYear: number
  docUrl: string
  docUrlExternal: boolean
}>()

const { t } = useI18n()
</script>
