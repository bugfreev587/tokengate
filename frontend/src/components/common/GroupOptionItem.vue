<template>
  <div class="flex min-w-0 flex-1 items-start justify-between gap-3">
    <!-- Left: name + description -->
    <div
      class="flex min-w-0 flex-1 flex-col items-start"
      :title="description || undefined"
    >
      <!-- Row 1: platform badge (name bold) -->
      <GroupBadge
        :name="name"
        :platform="platform"
        :subscription-type="subscriptionType"
        :capacity-source="capacitySource"
        :byo-enabled="byoEnabled"
        :byo-disabled-reason="byoDisabledReason"
        :show-rate="false"
        class="groupOptionItemBadge"
      />
      <!-- Row 2: description with top spacing -->
      <span
        v-if="description"
        class="mt-1.5 w-full text-left text-xs leading-relaxed text-gray-500 dark:text-gray-400 line-clamp-2"
      >
        {{ description }}
      </span>
      <span
        v-if="isConnectedAccountCapacity && !isBYODisabled"
        class="mt-1.5 w-full text-left text-xs leading-relaxed text-emerald-700 dark:text-emerald-300"
      >
        {{ t('keys.byoGroupDescription') }}
      </span>
      <span
        v-if="isBYODisabled"
        class="mt-1.5 w-full text-left text-xs font-medium leading-relaxed text-red-600 dark:text-red-300"
      >
        {{ byoDisabledHint }}
      </span>
    </div>

    <!-- Right: rate pill + checkmark (vertically centered to first row) -->
    <div class="flex shrink-0 items-center gap-2 pt-0.5">
      <span
        v-if="isConnectedAccountCapacity"
        :class="[
          'inline-flex items-center whitespace-nowrap rounded-md px-3 py-1 text-xs font-semibold',
          isBYODisabled
            ? 'bg-red-50 text-red-700 dark:bg-red-900/20 dark:text-red-300'
            : 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/20 dark:text-emerald-300'
        ]"
      >
        {{ isBYODisabled ? t('keys.byoGroupDisabledLabel') : t('keys.byoGroupLabel') }}
      </span>
      <!-- Rate pill (platform color) -->
      <span v-else-if="rateMultiplier !== undefined" :class="['inline-flex items-center whitespace-nowrap rounded-full px-3 py-1 text-xs font-semibold', ratePillClass]">
        <template v-if="hasCustomRate">
          <span class="mr-1 line-through opacity-50">{{ rateMultiplier }}x</span>
          <span class="font-bold">{{ userRateMultiplier }}x</span>
        </template>
        <template v-else>
          {{ rateMultiplier }}x 倍率
        </template>
      </span>
      <!-- Checkmark -->
      <svg
        v-if="showCheckmark && selected"
        class="h-4 w-4 shrink-0 text-primary-600 dark:text-primary-400"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
        stroke-width="2"
      >
        <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
      </svg>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import GroupBadge from './GroupBadge.vue'
import type { SubscriptionType, GroupPlatform, GroupCapacitySource } from '@/types'

interface Props {
  name: string
  platform: GroupPlatform
  subscriptionType?: SubscriptionType
  rateMultiplier?: number
  userRateMultiplier?: number | null
  description?: string | null
  selected?: boolean
  showCheckmark?: boolean
  capacitySource?: GroupCapacitySource
  byoEnabled?: boolean | null
  byoDisabledReason?: string | null
}

const props = withDefaults(defineProps<Props>(), {
  subscriptionType: 'standard',
  selected: false,
  showCheckmark: true,
  userRateMultiplier: null,
  capacitySource: 'tokengate'
})

const { t } = useI18n()
const isConnectedAccountCapacity = computed(() => props.capacitySource === 'connected_account')
const isBYODisabled = computed(() => isConnectedAccountCapacity.value && props.byoEnabled === false)
const byoDisabledHint = computed(() => {
  switch (props.byoDisabledReason) {
    case 'subscription_inactive':
      return t('keys.byoGroupDisabled.subscriptionInactive')
    case 'account_missing':
      return t('keys.byoGroupDisabled.accountMissing')
    default:
      return t('keys.byoGroupDisabled.accountDisabled')
  }
})

// Whether user has a custom rate different from default
const hasCustomRate = computed(() => {
  return (
    !isConnectedAccountCapacity.value &&
    props.userRateMultiplier !== null &&
    props.userRateMultiplier !== undefined &&
    props.rateMultiplier !== undefined &&
    props.userRateMultiplier !== props.rateMultiplier
  )
})

// Rate pill color matches platform badge color
const ratePillClass = computed(() => {
  switch (props.platform) {
    case 'anthropic':
      return 'bg-amber-50 text-amber-700 dark:bg-amber-900/20 dark:text-amber-400'
    case 'openai':
      return 'bg-green-50 text-green-700 dark:bg-green-900/20 dark:text-green-400'
    case 'gemini':
      return 'bg-sky-50 text-sky-700 dark:bg-sky-900/20 dark:text-sky-400'
    default: // antigravity and others
      return 'bg-violet-50 text-violet-700 dark:bg-violet-900/20 dark:text-violet-400'
  }
})
</script>

<style scoped>
/* Bold the group name inside GroupBadge when used in dropdown option */
.groupOptionItemBadge :deep(span.truncate) {
  font-weight: 600;
}
</style>
