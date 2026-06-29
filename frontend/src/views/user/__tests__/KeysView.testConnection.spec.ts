import { describe, expect, it, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { nextTick } from 'vue'

import KeysView from '../KeysView.vue'

const {
  list,
  testConnection,
  getDashboardApiKeysUsage,
  getAvailable,
  getUserGroupRates,
  getPublicSettings,
  showError,
  showSuccess,
} = vi.hoisted(() => ({
  list: vi.fn(),
  testConnection: vi.fn(),
  getDashboardApiKeysUsage: vi.fn(),
  getAvailable: vi.fn(),
  getUserGroupRates: vi.fn(),
  getPublicSettings: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

const messages: Record<string, string> = {
  'keys.searchPlaceholder': 'Search name or key...',
  'keys.allGroups': 'All Groups',
  'keys.allStatus': 'All Status',
  'keys.apiKey': 'API Key',
  'keys.group': 'Group',
  'keys.usage': 'Usage',
  'keys.rateLimitColumn': 'Rate Limit',
  'keys.expiresAt': 'Expires',
  'keys.created': 'Created',
  'keys.today': 'Today',
  'keys.total': 'Last 30d',
  'keys.noExpiration': 'Never',
  'keys.status.active': 'Active',
  'keys.useKey': 'Use Key',
  'keys.importToCcSwitch': 'Import to CCS',
  'keys.disable': 'Disable',
  'keys.enable': 'Enable',
  'keys.testConnection': 'Test connection',
  'keys.testingConnection': 'Testing...',
  'keys.connectionTestTitle': 'Connection test',
  'keys.connectionTestSuccess': 'Connection OK',
  'keys.connectionTestFailed': 'Connection failed',
  'keys.connectionModelsVisible': 'Models visible',
  'keys.connectionModelsTested': 'Models tested',
  'common.name': 'Name',
  'common.status': 'Status',
  'common.actions': 'Actions',
  'common.edit': 'Edit',
  'common.delete': 'Delete',
  'common.close': 'Close',
  'common.refresh': 'Refresh',
}

vi.mock('@/api', () => ({
  keysAPI: {
    list,
    testConnection,
    update: vi.fn(),
    toggleStatus: vi.fn(),
    create: vi.fn(),
    delete: vi.fn(),
  },
  authAPI: {
    getPublicSettings,
  },
  usageAPI: {
    getDashboardApiKeysUsage,
  },
  userGroupsAPI: {
    getAvailable,
    getUserGroupRates,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError, showSuccess }),
}))

vi.mock('@/stores/onboarding', () => ({
  useOnboardingStore: () => ({
    isCurrentStep: vi.fn(() => false),
    nextStep: vi.fn(),
  }),
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({ copyToClipboard: vi.fn().mockResolvedValue(true) }),
}))

vi.mock('@/composables/usePersistedPageSize', () => ({
  getPersistedPageSize: () => 10,
}))

vi.mock('@/utils/ccswitchImport', () => ({
  buildCcSwitchImportDeeplink: vi.fn(() => 'ccswitch://import'),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => messages[key] ?? key,
    }),
  }
})

const AppLayoutStub = { template: '<div><slot /></div>' }
const TablePageLayoutStub = {
  template: '<div><slot name="filters" /><slot name="actions" /><slot name="table" /><slot name="pagination" /></div>',
}
const DataTableStub = {
  props: ['data'],
  template: `
    <div>
      <div v-for="row in data" :key="row.id" data-testid="api-key-row">
        <slot name="cell-actions" :row="row" />
      </div>
    </div>
  `,
}
const BaseDialogStub = {
  props: ['show', 'title'],
  template: `
    <section v-if="show" data-testid="base-dialog">
      <h2>{{ title }}</h2>
      <slot />
      <slot name="footer" />
    </section>
  `,
}

describe('KeysView test connection action', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    list.mockResolvedValue({
      items: [
        {
          id: 1,
          user_id: 7,
          key: 'sk-live-test',
          name: 'canary-key',
          group_id: 9,
          group: {
            id: 9,
            name: 'openai-default',
            platform: 'openai',
            subscription_type: 'standard',
            rate_multiplier: 1,
          },
          status: 'active',
          quota: 0,
          quota_used: 0,
          rate_limit_5h: 0,
          rate_limit_1d: 0,
          rate_limit_7d: 0,
          usage_5h: 0,
          usage_1d: 0,
          usage_7d: 0,
          ip_whitelist: [],
          ip_blacklist: [],
          expires_at: null,
          created_at: '2026-01-01T00:00:00Z',
          updated_at: '2026-01-01T00:00:00Z',
          last_used_at: null,
        },
      ],
      total: 1,
      page: 1,
      page_size: 10,
      pages: 1,
    })
    getDashboardApiKeysUsage.mockResolvedValue({ stats: {} })
    getAvailable.mockResolvedValue([])
    getUserGroupRates.mockResolvedValue({})
    getPublicSettings.mockResolvedValue({})
    testConnection.mockResolvedValue({
      api_key_id: 1,
      key_name: 'canary-key',
      group_id: 9,
      group_name: 'openai-default',
      platform: 'openai',
      base_url: 'https://api.tokengate.test',
      success: true,
      models_visible: true,
      visible_model_count: 1,
      tested_model_count: 1,
      skipped_model_count: 0,
      truncated: false,
      message: 'Tested 1 visible models successfully',
      results: [
        {
          model: 'gpt-4.1-mini',
          provider: 'openai',
          endpoint: '/v1/chat/completions',
          status: 'success',
          http_status: 200,
          latency_ms: 24,
        },
      ],
    })
  })

  it('runs a live connection test from the row actions and shows model results', async () => {
    const wrapper = mount(KeysView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          TablePageLayout: TablePageLayoutStub,
          DataTable: DataTableStub,
          BaseDialog: BaseDialogStub,
          Pagination: true,
          ConfirmDialog: true,
          EmptyState: true,
          Select: true,
          SearchInput: true,
          EndpointPopover: true,
          GroupBadge: true,
          GroupOptionItem: true,
          Icon: true,
          UseKeyModal: true,
          RouterLink: true,
          Teleport: true,
        },
      },
    })

    await flushPromises()
    await nextTick()

    await wrapper.get('[data-testid="test-connection-button-1"]').trigger('click')
    await flushPromises()
    await nextTick()

    expect(testConnection).toHaveBeenCalledWith(1, { max_models: 12 })
    expect(wrapper.text()).toContain('Connection OK')
    expect(wrapper.text()).toContain('gpt-4.1-mini')
    expect(wrapper.text()).toContain('/v1/chat/completions')
  })
})
