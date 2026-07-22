<template>
  <section aria-labelledby="home-flow-title" class="py-20 sm:py-24 lg:py-28">
    <div class="mx-auto max-w-7xl px-5 sm:px-8 lg:px-10">
      <div class="grid gap-8 lg:grid-cols-2 lg:items-end lg:gap-16">
        <div>
          <p class="text-[11px] font-bold uppercase tracking-[0.18em] text-emerald-700 dark:text-emerald-300">
            {{ t('home.flow.eyebrow') }}
          </p>
          <h2
            id="home-flow-title"
            class="mt-5 max-w-xl text-3xl font-bold leading-[1.05] tracking-[-0.045em] text-gray-950 dark:text-white sm:text-4xl"
          >
            {{ t('home.flow.title') }}
          </h2>
        </div>
        <p class="max-w-xl text-sm leading-7 text-gray-600 dark:text-dark-300 lg:justify-self-end">
          {{ t('home.flow.description') }}
        </p>
      </div>

      <ol class="mt-12 grid border-t border-stone-200 md:grid-cols-3 dark:border-dark-800">
        <li
          v-for="(step, index) in steps"
          :key="step.key"
          class="border-b border-stone-200 py-7 md:border-b-0 md:px-7 md:first:pl-0 md:last:pr-0 md:[&+li]:border-l dark:border-dark-800"
        >
          <span class="font-mono text-[11px] text-emerald-700 dark:text-emerald-300">
            {{ String(index + 1).padStart(2, '0') }}
          </span>
          <h3 class="mt-8 text-lg font-bold tracking-tight text-gray-950 dark:text-white">
            {{ step.title }}
          </h3>
          <p class="mt-3 max-w-sm text-sm leading-6 text-gray-500 dark:text-dark-400">
            {{ step.description }}
          </p>
        </li>
      </ol>
    </div>

    <div class="mt-16 border-y border-stone-200 bg-stone-100/70 dark:border-dark-800 dark:bg-dark-900/55">
      <div
        class="mx-auto flex max-w-7xl flex-wrap items-center justify-center gap-x-8 gap-y-3 px-5 py-5 text-xs text-gray-500 sm:px-8 lg:px-10 dark:text-dark-400"
      >
        <strong class="text-gray-800 dark:text-dark-100">{{ t('home.clients.worksWith') }}</strong>
        <span v-for="client in clients" :key="client.key">{{ client.label }}</span>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const stepKeys = ['choose', 'key', 'connect'] as const
const clientKeys = ['claude', 'codex', 'openai', 'anthropic', 'custom'] as const

const steps = computed(() =>
  stepKeys.map((key) => ({
    key,
    title: t(`home.flow.${key}.title`),
    description: t(`home.flow.${key}.description`)
  }))
)

const clients = computed(() =>
  clientKeys.map((key) => ({
    key,
    label: t(`home.clients.${key}`)
  }))
)
</script>
