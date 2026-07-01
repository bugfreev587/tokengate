<template>
  <AppLayout>
    <div class="mx-auto flex max-w-6xl flex-col gap-6">
      <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div class="space-y-1">
          <h1 class="text-2xl font-semibold tracking-tight text-gray-900 dark:text-white">
            {{ t('userAccounts.title') }}
          </h1>
          <p class="max-w-2xl text-sm text-gray-500 dark:text-gray-400">
            {{ t('userAccounts.description') }}
          </p>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <button
            type="button"
            class="btn btn-secondary"
            :disabled="loading"
            :title="t('common.refresh')"
            @click="loadAccounts"
          >
            <Icon name="refresh" size="md" :class="{ 'animate-spin': loading }" />
          </button>
          <button
            type="button"
            class="btn btn-primary"
            data-testid="connect-openai"
            :disabled="generatingAuthUrl"
            @click="startOpenAIConnection"
          >
            <Icon name="link" size="md" class="mr-2" />
            {{ generatingAuthUrl ? t('common.loading') : t('userAccounts.connectOpenAI') }}
          </button>
        </div>
      </div>

      <section
        v-if="authPanelOpen"
        class="rounded-lg border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800"
      >
        <div class="grid gap-5 lg:grid-cols-[1.1fr_0.9fr]">
          <div class="space-y-4">
            <div class="flex items-center gap-3">
              <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300">
                <Icon name="globe" size="md" />
              </div>
              <div>
                <h2 class="text-base font-semibold text-gray-900 dark:text-white">
                  {{ t('userAccounts.openAIConnectTitle') }}
                </h2>
                <p class="text-sm text-gray-500 dark:text-gray-400">
                  {{ t('userAccounts.openAIConnectSubtitle') }}
                </p>
              </div>
            </div>

            <div
              v-if="authUrl"
              class="rounded-lg border border-emerald-200 bg-emerald-50 p-4 dark:border-emerald-800 dark:bg-emerald-900/20"
            >
              <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <code class="min-w-0 break-all text-xs text-emerald-900 dark:text-emerald-100">
                  {{ authUrl }}
                </code>
                <a
                  class="btn btn-secondary flex-shrink-0"
                  :href="authUrl"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  <Icon name="externalLink" size="sm" class="mr-2" />
                  {{ t('userAccounts.openAuthorization') }}
                </a>
              </div>
            </div>

            <div
              v-else
              class="rounded-lg border border-gray-200 bg-gray-50 p-4 text-sm text-gray-600 dark:border-dark-700 dark:bg-dark-900/60 dark:text-gray-300"
            >
              {{ generatingAuthUrl ? t('common.loading') : t('userAccounts.oauthReadyState') }}
            </div>
          </div>

          <form class="grid gap-3" @submit.prevent="finishOpenAIOAuth">
            <div>
              <label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
                {{ t('userAccounts.accountName') }}
              </label>
              <input
                v-model="accountName"
                data-testid="account-name"
                class="input w-full"
                :placeholder="t('userAccounts.accountNamePlaceholder')"
              />
            </div>
            <div>
              <label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
                {{ t('userAccounts.oauthCode') }}
              </label>
              <input
                v-model="oauthCode"
                data-testid="oauth-code"
                class="input w-full font-mono text-sm"
                autocomplete="off"
              />
            </div>
            <div>
              <label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
                {{ t('userAccounts.oauthState') }}
              </label>
              <input
                v-model="oauthState"
                data-testid="oauth-state"
                class="input w-full font-mono text-sm"
                autocomplete="off"
              />
            </div>
            <button
              type="button"
              data-testid="finish-openai-oauth"
              class="btn btn-primary mt-1"
              :disabled="exchangingCode || !oauthCode.trim() || !oauthState.trim() || !sessionId"
              @click="finishOpenAIOAuth"
            >
              <Icon name="check" size="md" class="mr-2" />
              {{ exchangingCode ? t('common.submitting') : t('userAccounts.finishConnection') }}
            </button>
          </form>
        </div>
      </section>

      <div
        v-if="loadError"
        class="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700 dark:border-red-800 dark:bg-red-900/20 dark:text-red-300"
      >
        {{ loadError }}
      </div>

      <div class="overflow-hidden rounded-lg border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-800">
        <DataTable
          :columns="columns"
          :data="accounts"
          :loading="loading"
          row-key="id"
          :estimate-row-height="76"
        >
          <template #cell-name="{ row }">
            <div class="flex min-w-[180px] flex-col">
              <span class="font-medium text-gray-900 dark:text-white">{{ row.name }}</span>
              <span
                v-if="row.email"
                class="max-w-[240px] truncate text-xs text-gray-500 dark:text-gray-400"
                :title="row.email"
              >
                {{ row.email }}
              </span>
            </div>
          </template>

          <template #cell-platform="{ row }">
            <PlatformTypeBadge
              :platform="row.platform"
              :type="row.type"
              :plan-type="row.plan_type"
            />
          </template>

          <template #cell-status="{ row }">
            <StatusBadge :status="row.status" :label="statusLabel(row.status)" />
          </template>

          <template #cell-group="{ row }">
            <div class="flex min-w-[180px] flex-col gap-1">
              <span class="text-sm font-medium text-gray-800 dark:text-gray-100">
                {{ row.group_name || t('userAccounts.privateGroupPending') }}
              </span>
              <span class="inline-flex w-fit items-center gap-1 rounded-md bg-emerald-50 px-2 py-0.5 text-xs font-medium text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300">
                <Icon name="shield" size="xs" />
                {{ t('userAccounts.capacitySourceConnected') }}
              </span>
            </div>
          </template>

          <template #cell-last_used="{ row }">
            <span class="text-sm text-gray-600 dark:text-gray-300">
              {{ formatDateTime(row.last_used_at) }}
            </span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center justify-end gap-1">
              <button
                type="button"
                class="rounded-lg p-2 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 disabled:opacity-50 dark:hover:bg-dark-700 dark:hover:text-gray-200"
                :data-testid="`refresh-account-${row.id}`"
                :title="t('userAccounts.refresh')"
                :disabled="actionLoadingId === row.id"
                @click="refreshAccount(row)"
              >
                <Icon name="refresh" size="sm" />
              </button>
              <button
                type="button"
                class="rounded-lg p-2 text-red-500 transition-colors hover:bg-red-50 hover:text-red-700 disabled:opacity-50 dark:hover:bg-red-900/20 dark:hover:text-red-300"
                :data-testid="`delete-account-${row.id}`"
                :title="t('common.delete')"
                :disabled="actionLoadingId === row.id"
                @click="openDeleteDialog(row)"
              >
                <Icon name="trash" size="sm" />
              </button>
            </div>
          </template>

          <template #empty>
            <div class="flex flex-col items-center gap-3 py-4 text-center">
              <div class="flex h-12 w-12 items-center justify-center rounded-lg bg-gray-100 text-gray-500 dark:bg-dark-700 dark:text-gray-300">
                <Icon name="inbox" size="lg" />
              </div>
              <div>
                <p class="font-medium text-gray-900 dark:text-white">{{ t('userAccounts.emptyTitle') }}</p>
                <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('userAccounts.emptyDescription') }}</p>
              </div>
            </div>
          </template>
        </DataTable>
      </div>
    </div>

    <ConfirmDialog
      :show="!!deleteTarget"
      :title="t('userAccounts.deleteTitle')"
      :message="t('userAccounts.deleteMessage', { name: deleteTarget?.name || '' })"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      :danger="true"
      @confirm="confirmDelete"
      @cancel="deleteTarget = null"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  deleteConnectedAccount,
  exchangeOpenAIConnectedAccountCode,
  generateOpenAIConnectedAccountAuthUrl,
  listConnectedAccounts,
  refreshOpenAIConnectedAccount,
  type ConnectedAccountSummary,
} from '@/api/user'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import DataTable from '@/components/common/DataTable.vue'
import PlatformTypeBadge from '@/components/common/PlatformTypeBadge.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import type { Column } from '@/components/common/types'
import Icon from '@/components/icons/Icon.vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import { useAppStore } from '@/stores/app'

const OPENAI_OAUTH_STORAGE_KEY = 'tokengate:user-openai-connected-account-oauth'

const { t } = useI18n()
const appStore = useAppStore()

const accounts = ref<ConnectedAccountSummary[]>([])
const loading = ref(false)
const loadError = ref('')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

const authPanelOpen = ref(false)
const generatingAuthUrl = ref(false)
const exchangingCode = ref(false)
const authUrl = ref('')
const sessionId = ref('')
const redirectUri = ref('')
const oauthCode = ref('')
const oauthState = ref('')
const accountName = ref('')

const actionLoadingId = ref<number | null>(null)
const deleteTarget = ref<ConnectedAccountSummary | null>(null)

const columns = computed<Column[]>(() => [
  { key: 'name', label: t('userAccounts.columns.name') },
  { key: 'platform', label: t('userAccounts.columns.platform') },
  { key: 'status', label: t('userAccounts.columns.status') },
  { key: 'group', label: t('userAccounts.columns.group') },
  { key: 'last_used', label: t('userAccounts.columns.lastUsed') },
  { key: 'actions', label: t('common.actions'), class: 'text-right' },
])

async function loadAccounts() {
  loading.value = true
  loadError.value = ''
  try {
    const result = await listConnectedAccounts(page.value, pageSize.value)
    accounts.value = result.items
    total.value = result.total
  } catch (error) {
    loadError.value = errorMessage(error, t('userAccounts.loadFailed'))
    appStore.showError(loadError.value)
  } finally {
    loading.value = false
  }
}

async function startOpenAIConnection() {
  generatingAuthUrl.value = true
  authPanelOpen.value = true
  loadError.value = ''
  try {
    const resolvedRedirectUri = resolveRedirectUri()
    const result = await generateOpenAIConnectedAccountAuthUrl({
      redirect_uri: resolvedRedirectUri,
    })
    authUrl.value = result.auth_url
    sessionId.value = result.session_id
    redirectUri.value = resolvedRedirectUri
    oauthState.value = parseOAuthState(result.auth_url) || oauthState.value
    persistOAuthSession()
  } catch (error) {
    appStore.showError(errorMessage(error, t('userAccounts.authUrlFailed')))
  } finally {
    generatingAuthUrl.value = false
  }
}

async function finishOpenAIOAuth() {
  if (!sessionId.value || !oauthCode.value.trim() || !oauthState.value.trim()) {
    appStore.showError(t('userAccounts.oauthFieldsRequired'))
    return
  }
  exchangingCode.value = true
  try {
    await exchangeOpenAIConnectedAccountCode({
      session_id: sessionId.value,
      code: oauthCode.value.trim(),
      state: oauthState.value.trim(),
      redirect_uri: redirectUri.value || resolveRedirectUri(),
      name: accountName.value.trim() || undefined,
    })
    appStore.showSuccess(t('userAccounts.connectedSuccess'))
    resetOAuthState()
    await loadAccounts()
  } catch (error) {
    appStore.showError(errorMessage(error, t('userAccounts.exchangeFailed')))
  } finally {
    exchangingCode.value = false
  }
}

async function refreshAccount(account: ConnectedAccountSummary) {
  actionLoadingId.value = account.id
  try {
    await refreshOpenAIConnectedAccount(account.id)
    appStore.showSuccess(t('userAccounts.refreshSuccess'))
    await loadAccounts()
  } catch (error) {
    appStore.showError(errorMessage(error, t('userAccounts.refreshFailed')))
  } finally {
    actionLoadingId.value = null
  }
}

function openDeleteDialog(account: ConnectedAccountSummary) {
  deleteTarget.value = account
}

async function confirmDelete() {
  if (!deleteTarget.value) return
  const account = deleteTarget.value
  actionLoadingId.value = account.id
  try {
    await deleteConnectedAccount(account.id)
    appStore.showSuccess(t('userAccounts.deleteSuccess'))
    deleteTarget.value = null
    await loadAccounts()
  } catch (error) {
    appStore.showError(errorMessage(error, t('userAccounts.deleteFailed')))
  } finally {
    actionLoadingId.value = null
  }
}

function resetOAuthState() {
  authPanelOpen.value = false
  authUrl.value = ''
  sessionId.value = ''
  redirectUri.value = ''
  oauthCode.value = ''
  oauthState.value = ''
  accountName.value = ''
  if (typeof window !== 'undefined') {
    window.sessionStorage.removeItem(OPENAI_OAUTH_STORAGE_KEY)
  }
}

function restoreOAuthState() {
  if (typeof window === 'undefined') return
  const raw = window.sessionStorage.getItem(OPENAI_OAUTH_STORAGE_KEY)
  if (raw) {
    try {
      const parsed = JSON.parse(raw) as {
        authUrl?: string
        sessionId?: string
        redirectUri?: string
        oauthState?: string
      }
      authUrl.value = parsed.authUrl || ''
      sessionId.value = parsed.sessionId || ''
      redirectUri.value = parsed.redirectUri || ''
      oauthState.value = parsed.oauthState || parseOAuthState(parsed.authUrl || '') || ''
      authPanelOpen.value = !!sessionId.value
    } catch {
      window.sessionStorage.removeItem(OPENAI_OAUTH_STORAGE_KEY)
    }
  }

  const currentURL = new URL(window.location.href)
  oauthCode.value = currentURL.searchParams.get('code') || oauthCode.value
  oauthState.value = currentURL.searchParams.get('state') || oauthState.value
  if (oauthCode.value || oauthState.value) {
    authPanelOpen.value = true
  }
}

function persistOAuthSession() {
  if (typeof window === 'undefined') return
  window.sessionStorage.setItem(
    OPENAI_OAUTH_STORAGE_KEY,
    JSON.stringify({
      authUrl: authUrl.value,
      sessionId: sessionId.value,
      redirectUri: redirectUri.value,
      oauthState: oauthState.value,
    })
  )
}

function parseOAuthState(url: string): string {
  if (!url) return ''
  try {
    return new URL(url).searchParams.get('state') || ''
  } catch {
    const match = url.match(/[?&]state=([^&]+)/)
    return match?.[1] ? decodeURIComponent(match[1]) : ''
  }
}

function resolveRedirectUri(): string {
  if (typeof window === 'undefined') {
    return '/accounts'
  }
  return `${window.location.origin}/accounts`
}

function statusLabel(status: string): string {
  switch (status) {
    case 'active':
      return t('common.active')
    case 'inactive':
      return t('common.inactive')
    case 'error':
      return t('common.error')
    default:
      return status || t('common.unknown')
  }
}

function formatDateTime(value?: string | null): string {
  if (!value) {
    return t('common.time.never')
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return t('common.unknown')
  }
  return date.toLocaleString()
}

function errorMessage(error: unknown, fallback: string): string {
  if (error && typeof error === 'object') {
    const maybe = error as { message?: string; response?: { data?: { detail?: string; message?: string } } }
    return maybe.response?.data?.detail || maybe.response?.data?.message || maybe.message || fallback
  }
  return fallback
}

onMounted(async () => {
  restoreOAuthState()
  await loadAccounts()
})
</script>
