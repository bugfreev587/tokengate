<template>
  <section
    aria-labelledby="shared-capabilities-title"
    class="mx-auto grid max-w-7xl gap-8 border-y border-stone-200 px-5 py-10 sm:px-8 lg:grid-cols-[0.7fr_2fr] lg:gap-14 lg:px-10 lg:py-12 dark:border-dark-800"
  >
    <h2 id="shared-capabilities-title" class="text-base font-bold tracking-tight text-gray-950 dark:text-white">
      {{ t('home.shared.title') }}
    </h2>

    <dl class="grid gap-x-8 gap-y-8 sm:grid-cols-2 lg:grid-cols-4">
      <div v-for="capability in capabilities" :key="capability.key">
        <dt class="text-sm font-semibold text-gray-900 dark:text-dark-100">
          {{ capability.title }}
        </dt>
        <dd class="mt-2 text-xs leading-5 text-gray-500 dark:text-dark-400">
          {{ capability.description }}
        </dd>
      </div>
    </dl>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const capabilityKeys = ['apiKeys', 'compatibleApis', 'usageLogs', 'explicitRouting'] as const
const capabilities = computed(() =>
  capabilityKeys.map((key) => ({
    key,
    title: t(`home.shared.${key}.title`),
    description: t(`home.shared.${key}.description`)
  }))
)
</script>
