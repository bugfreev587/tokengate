<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-col justify-between gap-4 lg:flex-row lg:items-start">
          <div class="flex flex-1 flex-wrap items-center gap-3">
            <div class="relative w-full sm:w-72">
              <Icon
                name="search"
                size="md"
                class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-gray-500"
              />
              <input
                v-model="searchQuery"
                type="text"
                :placeholder="t('admin.modelPricing.searchPlaceholder', 'Search model or provider')"
                class="input pl-10"
                @input="handleSearch"
              />
            </div>

            <Select
              v-model="filters.provider"
              :options="providerOptions"
              class="w-full sm:w-40"
              @change="reloadFirstPage"
            />

            <Select
              v-model="filters.source"
              :options="sourceOptions"
              class="w-full sm:w-44"
              @change="reloadFirstPage"
            />

            <Select
              v-model="filters.billing_mode"
              :options="billingModeOptions"
              class="w-full sm:w-40"
              @change="reloadFirstPage"
            />
          </div>

          <div class="flex w-full flex-shrink-0 flex-wrap items-center justify-end gap-3 lg:w-auto">
            <button
              @click="loadPricing"
              :disabled="loading"
              class="btn btn-secondary"
              :title="t('common.refresh', 'Refresh')"
            >
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
            <button @click="openCreateDialog" class="btn btn-primary">
              <Icon name="plus" size="md" class="mr-2" />
              {{ t('admin.modelPricing.addOverride', 'Add Override') }}
            </button>
          </div>
        </div>

        <div
          v-if="errorMessage"
          class="mt-3 rounded border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700 dark:border-red-900/60 dark:bg-red-950/30 dark:text-red-300"
        >
          {{ errorMessage }}
        </div>
      </template>

      <template #table>
        <DataTable
          :columns="columns"
          :data="rows"
          :loading="loading"
          :server-side-sort="false"
        >
          <template #cell-model="{ value }">
            <code class="code text-xs">{{ value }}</code>
          </template>

          <template #cell-provider="{ value }">
            <span
              class="inline-flex rounded px-2 py-0.5 text-xs font-medium"
              :class="providerClass(value)"
            >
              {{ value || '-' }}
            </span>
          </template>

          <template #cell-source="{ value }">
            <span
              class="inline-flex rounded-full px-2 py-0.5 text-xs font-medium"
              :class="sourceClass(value)"
            >
              {{ sourceLabel(value) }}
            </span>
          </template>

          <template #cell-billing_mode="{ value }">
            <span class="text-sm text-gray-700 dark:text-gray-300">{{ billingModeLabel(value) }}</span>
          </template>

          <template #cell-input_price="{ value }">
            <span class="font-mono text-xs">{{ formatMTok(value) }}</span>
          </template>

          <template #cell-output_price="{ value }">
            <span class="font-mono text-xs">{{ formatMTok(value) }}</span>
          </template>

          <template #cell-cache_write_price="{ value }">
            <span class="font-mono text-xs">{{ formatMTok(value) }}</span>
          </template>

          <template #cell-cache_read_price="{ value }">
            <span class="font-mono text-xs">{{ formatMTok(value) }}</span>
          </template>

          <template #cell-image_output_price="{ value }">
            <span class="font-mono text-xs">{{ formatMTok(value) }}</span>
          </template>

          <template #cell-per_request_price="{ value }">
            <span class="font-mono text-xs">{{ formatUSD(value) }}</span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center gap-1">
              <button
                @click="openEditDialog(row)"
                class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-primary-600 dark:hover:bg-dark-700 dark:hover:text-primary-400"
              >
                <Icon name="edit" size="sm" />
                <span class="text-xs">{{ t('common.edit', 'Edit') }}</span>
              </button>
              <button
                v-if="row.override"
                @click="clearOverride(row)"
                class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400"
              >
                <Icon name="trash" size="sm" />
                <span class="text-xs">{{ t('admin.modelPricing.clear', 'Clear') }}</span>
              </button>
            </div>
          </template>

          <template #empty>
            <EmptyState
              :title="t('admin.modelPricing.emptyTitle', 'No model prices found')"
              :description="t('admin.modelPricing.emptyDescription', 'Try another search or add a global override for a custom model.')"
              :action-text="t('admin.modelPricing.addOverride', 'Add Override')"
              @action="openCreateDialog"
            />
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>

    <BaseDialog
      :show="showDialog"
      :title="dialogTitle"
      width="wide"
      @close="closeDialog"
    >
      <form class="space-y-4" @submit.prevent="saveOverride">
        <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
          <div class="md:col-span-2">
            <label class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('admin.modelPricing.fields.model', 'Model') }}
            </label>
            <input
              v-model.trim="form.model"
              :disabled="!creatingOverride"
              class="input font-mono text-sm"
              placeholder="claude-sonnet-4.5"
            />
          </div>
          <div>
            <label class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('admin.modelPricing.fields.provider', 'Provider') }}
            </label>
            <input v-model.trim="form.provider" class="input text-sm" placeholder="anthropic" />
          </div>
        </div>

        <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
          <div>
            <label class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('admin.modelPricing.fields.billingMode', 'Billing Mode') }}
            </label>
            <Select v-model="form.billing_mode" :options="editBillingModeOptions" />
          </div>
          <div class="md:col-span-2 rounded border border-gray-200 bg-gray-50 px-3 py-2 text-xs text-gray-500 dark:border-dark-700 dark:bg-dark-800 dark:text-gray-400">
            {{ t('admin.modelPricing.blankUsesFallback', 'Leave a price blank to keep using the fallback value shown as placeholder.') }}
          </div>
        </div>

        <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-5">
          <div>
            <label class="price-label">{{ t('admin.modelPricing.fields.input', 'Input') }} <span>$/1M</span></label>
            <input v-model="form.input_price" type="number" min="0" step="any" class="input mt-1 text-sm" :placeholder="fallbackMTok('input_price')" />
          </div>
          <div>
            <label class="price-label">{{ t('admin.modelPricing.fields.output', 'Output') }} <span>$/1M</span></label>
            <input v-model="form.output_price" type="number" min="0" step="any" class="input mt-1 text-sm" :placeholder="fallbackMTok('output_price')" />
          </div>
          <div>
            <label class="price-label">{{ t('admin.modelPricing.fields.cacheWrite', 'Cache Write') }} <span>$/1M</span></label>
            <input v-model="form.cache_write_price" type="number" min="0" step="any" class="input mt-1 text-sm" :placeholder="fallbackMTok('cache_write_price')" />
          </div>
          <div>
            <label class="price-label">{{ t('admin.modelPricing.fields.cacheRead', 'Cache Read') }} <span>$/1M</span></label>
            <input v-model="form.cache_read_price" type="number" min="0" step="any" class="input mt-1 text-sm" :placeholder="fallbackMTok('cache_read_price')" />
          </div>
          <div>
            <label class="price-label">{{ t('admin.modelPricing.fields.imageOutput', 'Image Output') }} <span>$/1M</span></label>
            <input v-model="form.image_output_price" type="number" min="0" step="any" class="input mt-1 text-sm" :placeholder="fallbackMTok('image_output_price')" />
          </div>
        </div>

        <div class="w-full sm:w-64">
          <label class="price-label">{{ t('admin.modelPricing.fields.perRequest', 'Per Request / Image') }} <span>$</span></label>
          <input v-model="form.per_request_price" type="number" min="0" step="any" class="input mt-1 text-sm" :placeholder="fallbackUSD('per_request_price')" />
        </div>
      </form>

      <template #footer>
        <div class="flex w-full flex-wrap items-center justify-between gap-3">
          <button
            v-if="editingRow?.override"
            type="button"
            class="btn btn-danger"
            :disabled="saving"
            @click="clearOverride(editingRow)"
          >
            {{ t('admin.modelPricing.clearOverride', 'Clear Override') }}
          </button>
          <div v-else></div>
          <div class="flex items-center gap-3">
            <button type="button" class="btn btn-secondary" :disabled="saving" @click="closeDialog">
              {{ t('common.cancel', 'Cancel') }}
            </button>
            <button type="button" class="btn btn-primary" :disabled="saving" @click="saveOverride">
              <Icon v-if="saving" name="refresh" size="sm" class="mr-2 animate-spin" />
              {{ t('common.save', 'Save') }}
            </button>
          </div>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import modelPricingAPI, { type GlobalModelPricingRow, type ModelPricingSnapshot } from '@/api/admin/modelPricing'
import type { BillingMode } from '@/constants/channel'
import { mTokToPerToken, perTokenToMTok, toNullableNumber } from '@/components/admin/channel/types'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'

type PriceField = keyof ModelPricingSnapshot

const { t } = useI18n()
const appStore = useAppStore()

const rows = ref<GlobalModelPricingRow[]>([])
const loading = ref(false)
const saving = ref(false)
const errorMessage = ref('')
const searchQuery = ref('')
const showDialog = ref(false)
const creatingOverride = ref(false)
const editingRow = ref<GlobalModelPricingRow | null>(null)

const pagination = reactive({
  page: 1,
  page_size: 50,
  total: 0,
  pages: 0,
})

const filters = reactive({
  provider: '',
  source: '',
  billing_mode: '',
})

const form = reactive({
  model: '',
  provider: '',
  billing_mode: 'token' as BillingMode,
  input_price: null as number | string | null,
  output_price: null as number | string | null,
  cache_write_price: null as number | string | null,
  cache_read_price: null as number | string | null,
  image_output_price: null as number | string | null,
  per_request_price: null as number | string | null,
})

const columns = computed<Column[]>(() => [
  { key: 'model', label: t('admin.modelPricing.columns.model', 'Model'), sortable: true },
  { key: 'provider', label: t('admin.modelPricing.columns.provider', 'Provider'), sortable: true },
  { key: 'source', label: t('admin.modelPricing.columns.source', 'Source'), sortable: true },
  { key: 'billing_mode', label: t('admin.modelPricing.columns.mode', 'Mode'), sortable: true },
  { key: 'input_price', label: t('admin.modelPricing.columns.input', 'Input'), sortable: true },
  { key: 'output_price', label: t('admin.modelPricing.columns.output', 'Output'), sortable: true },
  { key: 'cache_write_price', label: t('admin.modelPricing.columns.cacheWrite', 'Cache W'), sortable: true },
  { key: 'cache_read_price', label: t('admin.modelPricing.columns.cacheRead', 'Cache R'), sortable: true },
  { key: 'image_output_price', label: t('admin.modelPricing.columns.imageOutput', 'Image Out'), sortable: true },
  { key: 'per_request_price', label: t('admin.modelPricing.columns.perRequest', 'Per Req'), sortable: true },
  { key: 'actions', label: t('common.actions', 'Actions'), sortable: false },
])

const providerOptions = computed(() => [
  { value: '', label: t('admin.modelPricing.filters.allProviders', 'All Providers') },
  { value: 'anthropic', label: 'Anthropic' },
  { value: 'openai', label: 'OpenAI' },
  { value: 'gemini', label: 'Gemini' },
  { value: 'google', label: 'Google' },
  { value: 'vertex_ai', label: 'Vertex AI' },
])

const sourceOptions = computed(() => [
  { value: '', label: t('admin.modelPricing.filters.allSources', 'All Sources') },
  { value: 'global_override', label: t('admin.modelPricing.sources.globalOverride', 'Global Override') },
  { value: 'litellm', label: 'LiteLLM' },
  { value: 'fallback', label: t('admin.modelPricing.sources.fallback', 'Fallback') },
])

const billingModeOptions = computed(() => [
  { value: '', label: t('admin.modelPricing.filters.allModes', 'All Modes') },
  { value: 'token', label: 'Token' },
  { value: 'per_request', label: t('admin.modelPricing.modes.perRequest', 'Per Request') },
  { value: 'image', label: t('admin.modelPricing.modes.image', 'Image') },
])

const editBillingModeOptions = computed(() => billingModeOptions.value.filter(option => option.value !== ''))

const dialogTitle = computed(() => (
  creatingOverride.value
    ? t('admin.modelPricing.addOverride', 'Add Override')
    : t('admin.modelPricing.editOverride', 'Edit Override')
))

let searchTimeout: ReturnType<typeof setTimeout> | undefined

async function loadPricing() {
  loading.value = true
  errorMessage.value = ''
  try {
    const response = await modelPricingAPI.list({
      page: pagination.page,
      page_size: pagination.page_size,
      search: searchQuery.value || undefined,
      provider: filters.provider || undefined,
      source: filters.source || undefined,
      billing_mode: filters.billing_mode || undefined,
    })
    rows.value = response.items
    pagination.total = response.total
    pagination.pages = response.pages
    pagination.page = response.page
    pagination.page_size = response.page_size
  } catch (error) {
    errorMessage.value = t('admin.modelPricing.failedToLoad', 'Failed to load model pricing')
    appStore.showError(errorMessage.value)
    console.error('Failed to load model pricing:', error)
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    reloadFirstPage()
  }, 300)
}

function reloadFirstPage() {
  pagination.page = 1
  loadPricing()
}

function handlePageChange(page: number) {
  pagination.page = page
  loadPricing()
}

function handlePageSizeChange(pageSize: number) {
  pagination.page_size = pageSize
  pagination.page = 1
  loadPricing()
}

function openCreateDialog() {
  creatingOverride.value = true
  editingRow.value = null
  Object.assign(form, {
    model: '',
    provider: '',
    billing_mode: 'token',
    input_price: null,
    output_price: null,
    cache_write_price: null,
    cache_read_price: null,
    image_output_price: null,
    per_request_price: null,
  })
  showDialog.value = true
}

function openEditDialog(row: GlobalModelPricingRow) {
  creatingOverride.value = false
  editingRow.value = row
  const override = row.override
  Object.assign(form, {
    model: row.model,
    provider: override?.provider ?? row.provider ?? '',
    billing_mode: (override?.billing_mode ?? row.billing_mode ?? 'token') as BillingMode,
    input_price: perTokenToMTok(override?.input_price ?? null),
    output_price: perTokenToMTok(override?.output_price ?? null),
    cache_write_price: perTokenToMTok(override?.cache_write_price ?? null),
    cache_read_price: perTokenToMTok(override?.cache_read_price ?? null),
    image_output_price: perTokenToMTok(override?.image_output_price ?? null),
    per_request_price: override?.per_request_price ?? null,
  })
  showDialog.value = true
}

function closeDialog() {
  showDialog.value = false
  editingRow.value = null
}

async function saveOverride() {
  if (!form.model.trim()) {
    appStore.showError(t('admin.modelPricing.modelRequired', 'Model is required'))
    return
  }
  saving.value = true
  try {
    await modelPricingAPI.upsertOverride({
      model: form.model.trim(),
      provider: form.provider.trim(),
      billing_mode: form.billing_mode,
      input_price: mTokToPerToken(form.input_price),
      output_price: mTokToPerToken(form.output_price),
      cache_write_price: mTokToPerToken(form.cache_write_price),
      cache_read_price: mTokToPerToken(form.cache_read_price),
      image_output_price: mTokToPerToken(form.image_output_price),
      per_request_price: toNullableNumber(form.per_request_price),
    })
    appStore.showSuccess(t('admin.modelPricing.saved', 'Model pricing override saved'))
    closeDialog()
    loadPricing()
  } catch (error) {
    appStore.showError(t('admin.modelPricing.failedToSave', 'Failed to save model pricing override'))
    console.error('Failed to save model pricing override:', error)
  } finally {
    saving.value = false
  }
}

async function clearOverride(row: GlobalModelPricingRow) {
  if (!row.override || saving.value) return
  saving.value = true
  try {
    await modelPricingAPI.deleteOverride(row.model)
    appStore.showSuccess(t('admin.modelPricing.cleared', 'Model pricing override cleared'))
    closeDialog()
    loadPricing()
  } catch (error) {
    appStore.showError(t('admin.modelPricing.failedToClear', 'Failed to clear model pricing override'))
    console.error('Failed to clear model pricing override:', error)
  } finally {
    saving.value = false
  }
}

function fallbackFor(field: PriceField): number | null {
  return editingRow.value?.fallback?.[field] ?? null
}

function fallbackMTok(field: PriceField): string {
  const value = perTokenToMTok(fallbackFor(field))
  return value == null ? '' : String(value)
}

function fallbackUSD(field: PriceField): string {
  const value = fallbackFor(field)
  return value == null ? '' : String(value)
}

function formatMTok(value: number | null | undefined): string {
  if (value == null) return '-'
  return `$${(value * 1_000_000).toLocaleString(undefined, { maximumFractionDigits: 6 })}`
}

function formatUSD(value: number | null | undefined): string {
  if (value == null) return '-'
  return `$${value.toLocaleString(undefined, { maximumFractionDigits: 6 })}`
}

function sourceLabel(source: string): string {
  switch (source) {
    case 'global_override':
      return t('admin.modelPricing.sources.globalOverride', 'Global Override')
    case 'litellm':
      return 'LiteLLM'
    case 'fallback':
      return t('admin.modelPricing.sources.fallback', 'Fallback')
    default:
      return source || '-'
  }
}

function billingModeLabel(mode: string): string {
  switch (mode) {
    case 'per_request':
      return t('admin.modelPricing.modes.perRequest', 'Per Request')
    case 'image':
      return t('admin.modelPricing.modes.image', 'Image')
    default:
      return 'Token'
  }
}

function sourceClass(source: string): string {
  if (source === 'global_override') return 'bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-300'
  if (source === 'litellm') return 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300'
  return 'bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-gray-300'
}

function providerClass(provider: string): string {
  switch ((provider || '').toLowerCase()) {
    case 'anthropic':
      return 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-300'
    case 'openai':
      return 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-300'
    case 'gemini':
    case 'google':
    case 'vertex_ai':
      return 'bg-sky-100 text-sky-800 dark:bg-sky-900/30 dark:text-sky-300'
    default:
      return 'bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-gray-300'
  }
}

onMounted(() => {
  loadPricing()
})
</script>

<style scoped>
.price-label {
  @apply block text-xs font-medium text-gray-500 dark:text-gray-400;
}

.price-label span {
  @apply font-normal text-gray-400;
}
</style>
