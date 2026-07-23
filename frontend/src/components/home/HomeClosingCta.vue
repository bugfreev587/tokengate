<template>
  <section class="bg-[#23362f] text-stone-50 dark:bg-emerald-950/70">
    <div
      class="mx-auto flex max-w-7xl flex-col items-start justify-between gap-8 px-5 py-14 sm:px-8 sm:py-16 md:flex-row md:items-center lg:px-10"
    >
      <h2 class="max-w-2xl text-3xl font-bold leading-[1.05] tracking-[-0.045em] sm:text-4xl">
        {{ t('home.closing.title') }}
      </h2>
      <RouterLink
        to="/pricing"
        class="inline-flex shrink-0 items-center gap-2 rounded-full bg-stone-50 px-5 py-3 text-sm font-semibold text-[#23362f] transition duration-300 ease-out hover:-translate-y-0.5 hover:bg-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-300 focus-visible:ring-offset-2 focus-visible:ring-offset-[#23362f] active:translate-y-0"
      >
        {{ t('home.closing.button') }}
        <span aria-hidden="true">&rarr;</span>
      </RouterLink>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>
