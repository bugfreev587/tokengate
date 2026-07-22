<template>
  <section
    id="capacity-modes"
    aria-labelledby="home-hero-title"
    class="mx-auto grid max-w-7xl gap-12 px-5 pb-16 pt-14 sm:px-8 sm:pb-20 sm:pt-20 lg:grid-cols-[0.78fr_1.22fr] lg:items-end lg:gap-16 lg:px-10 lg:pb-24 lg:pt-28"
  >
    <div class="home-reveal">
      <p
        class="flex items-center gap-3 text-[11px] font-bold uppercase tracking-[0.18em] text-emerald-700 dark:text-emerald-300"
      >
        <span class="h-1.5 w-1.5 rounded-full bg-emerald-600 ring-4 ring-emerald-600/10"></span>
        {{ t('home.hero.eyebrow') }}
      </p>

      <h1
        id="home-hero-title"
        class="mt-7 max-w-xl text-4xl font-bold leading-[0.98] tracking-[-0.055em] text-gray-950 dark:text-white sm:text-5xl lg:text-6xl"
      >
        {{ t('home.hero.title') }}
      </h1>

      <p v-if="siteSubtitle" class="mt-6 max-w-xl text-sm font-semibold text-gray-800 dark:text-dark-100">
        {{ siteSubtitle }}
      </p>
      <p class="mt-3 max-w-xl text-base leading-8 text-gray-600 dark:text-dark-300">
        {{ t('home.hero.description') }}
      </p>

      <div class="mt-8 flex flex-col gap-3 sm:flex-row">
        <RouterLink
          to="/pricing"
          class="inline-flex items-center justify-center gap-2 rounded-full bg-gray-900 px-5 py-3 text-sm font-semibold text-white transition duration-300 ease-out hover:-translate-y-0.5 hover:bg-gray-800 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-600 focus-visible:ring-offset-2 active:translate-y-0 dark:bg-white dark:text-gray-950 dark:hover:bg-gray-100 dark:focus-visible:ring-offset-dark-950"
        >
          {{ t('home.hero.explorePricing') }}
          <span aria-hidden="true">&rarr;</span>
        </RouterLink>

        <RouterLink
          v-if="!docUrlExternal"
          :to="docUrl"
          class="inline-flex items-center justify-center rounded-full border border-stone-300 bg-white/70 px-5 py-3 text-sm font-semibold text-gray-800 transition duration-300 ease-out hover:-translate-y-0.5 hover:border-stone-400 hover:bg-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-600 focus-visible:ring-offset-2 active:translate-y-0 dark:border-dark-700 dark:bg-dark-900/60 dark:text-dark-100 dark:hover:border-dark-600 dark:hover:bg-dark-900 dark:focus-visible:ring-offset-dark-950"
        >
          {{ t('home.hero.readDocs') }}
        </RouterLink>
        <a
          v-else
          :href="docUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="inline-flex items-center justify-center rounded-full border border-stone-300 bg-white/70 px-5 py-3 text-sm font-semibold text-gray-800 transition duration-300 ease-out hover:-translate-y-0.5 hover:border-stone-400 hover:bg-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-600 focus-visible:ring-offset-2 active:translate-y-0 dark:border-dark-700 dark:bg-dark-900/60 dark:text-dark-100 dark:hover:border-dark-600 dark:hover:bg-dark-900 dark:focus-visible:ring-offset-dark-950"
        >
          {{ t('home.hero.readDocs') }}
        </a>
      </div>
    </div>

    <div
      class="home-reveal home-reveal-delayed rounded-[1.75rem] border border-stone-200 bg-stone-200/65 p-2 shadow-[0_24px_70px_-38px_rgba(26,51,43,0.36)] dark:border-dark-700 dark:bg-dark-900/80"
    >
      <div class="flex flex-col gap-1 px-3 py-3 sm:flex-row sm:items-center sm:justify-between sm:px-4">
        <p class="text-xs font-semibold text-gray-800 dark:text-dark-100">
          {{ t('home.modes.caption') }}
        </p>
        <p class="text-[11px] text-gray-500 dark:text-dark-400">
          {{ t('home.modes.equalHint') }}
        </p>
      </div>

      <div class="grid gap-2 md:grid-cols-2">
        <RouterLink
          to="/pricing"
          data-mode="usage-based"
          class="group relative flex min-h-[23rem] flex-col overflow-hidden rounded-[1.35rem] border border-stone-200 bg-white p-6 text-gray-950 transition duration-300 ease-out hover:-translate-y-1 hover:border-stone-300 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-600 focus-visible:ring-offset-2 active:translate-y-0 dark:border-dark-700 dark:bg-dark-800 dark:text-white dark:focus-visible:ring-offset-dark-900 sm:p-7"
        >
          <div class="flex items-start justify-between gap-4">
            <div>
              <p class="font-mono text-[10px] uppercase tracking-[0.14em] text-gray-400 dark:text-dark-400">
                {{ t('home.modes.usageBased.index') }}
              </p>
              <p class="mt-3 text-xs font-bold uppercase tracking-[0.15em] text-emerald-700 dark:text-emerald-300">
                {{ t('home.modes.usageBased.label') }}
              </p>
            </div>
            <span
              aria-hidden="true"
              class="grid h-9 w-9 place-items-center rounded-full border border-stone-200 text-sm transition duration-300 group-hover:-translate-y-0.5 group-hover:translate-x-0.5 dark:border-dark-600"
            >
              &nearr;
            </span>
          </div>

          <h2 class="mt-10 text-3xl font-bold leading-none tracking-[-0.045em]">
            {{ t('home.modes.usageBased.title') }}
          </h2>
          <p class="mt-4 text-sm leading-6 text-gray-600 dark:text-dark-300">
            {{ t('home.modes.usageBased.description') }}
          </p>

          <ul class="mt-auto space-y-3 border-t border-stone-200 pt-5 text-sm dark:border-dark-700">
            <li v-for="point in usagePoints" :key="point" class="flex items-start gap-3">
              <span
                aria-hidden="true"
                class="mt-0.5 grid h-4 w-4 shrink-0 place-items-center rounded-full bg-emerald-100 text-[10px] font-bold text-emerald-700 dark:bg-emerald-900/35 dark:text-emerald-300"
              >
                &#10003;
              </span>
              {{ point }}
            </li>
          </ul>
          <span class="sr-only">{{ t('home.modes.usageBased.cta') }}</span>
        </RouterLink>

        <RouterLink
          to="/pricing"
          data-mode="byo"
          class="group relative flex min-h-[23rem] flex-col overflow-hidden rounded-[1.35rem] border border-[#23362f] bg-[#23362f] p-6 text-stone-50 transition duration-300 ease-out hover:-translate-y-1 hover:border-[#315044] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-300 focus-visible:ring-offset-2 active:translate-y-0 dark:border-emerald-900/50 dark:bg-emerald-950/70 dark:focus-visible:ring-offset-dark-900 sm:p-7"
        >
          <div class="flex items-start justify-between gap-4">
            <div>
              <p class="font-mono text-[10px] uppercase tracking-[0.14em] text-stone-400">
                {{ t('home.modes.byo.index') }}
              </p>
              <p class="mt-3 text-xs font-bold uppercase tracking-[0.15em] text-emerald-200">
                {{ t('home.modes.byo.label') }}
              </p>
            </div>
            <span
              aria-hidden="true"
              class="grid h-9 w-9 place-items-center rounded-full border border-white/15 text-sm transition duration-300 group-hover:-translate-y-0.5 group-hover:translate-x-0.5"
            >
              &nearr;
            </span>
          </div>

          <h2 class="mt-10 text-3xl font-bold leading-none tracking-[-0.045em]">
            {{ t('home.modes.byo.title') }}
          </h2>
          <p class="mt-4 text-sm leading-6 text-stone-300">
            {{ t('home.modes.byo.description') }}
          </p>

          <ul class="mt-auto space-y-3 border-t border-white/10 pt-5 text-sm text-stone-100">
            <li v-for="point in byoPoints" :key="point" class="flex items-start gap-3">
              <span
                aria-hidden="true"
                class="mt-0.5 grid h-4 w-4 shrink-0 place-items-center rounded-full bg-emerald-800/70 text-[10px] font-bold text-emerald-200"
              >
                &#10003;
              </span>
              {{ point }}
            </li>
          </ul>
          <p class="mt-4 text-[11px] leading-5 text-stone-400">
            {{ t('home.modes.byo.providerNote') }}
          </p>
          <span class="sr-only">{{ t('home.modes.byo.cta') }}</span>
        </RouterLink>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

defineProps<{
  siteSubtitle: string
  docUrl: string
  docUrlExternal: boolean
}>()

const { t } = useI18n()

const usagePoints = computed(() => [
  t('home.modes.usageBased.points.balance'),
  t('home.modes.usageBased.points.capacity'),
  t('home.modes.usageBased.points.commitment')
])

const byoPoints = computed(() => [
  t('home.modes.byo.points.account'),
  t('home.modes.byo.points.gateway'),
  t('home.modes.byo.points.balance')
])
</script>

<style scoped>
.home-reveal {
  animation: home-rise 600ms cubic-bezier(0.16, 1, 0.3, 1) both;
}

.home-reveal-delayed {
  animation-delay: 100ms;
}

@keyframes home-rise {
  from {
    opacity: 0;
    transform: translateY(14px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .home-reveal,
  .home-reveal-delayed {
    animation: none;
  }
}
</style>
