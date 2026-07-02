import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { nextTick } from 'vue'

import KeysView from '../KeysView.vue'

const {
  list,
  getDashboardApiKeysUsage,
  getAvailable,
  getUserGroupRates,
  getPublicSettings,
  showError,
  showSuccess,
} = vi.hoisted(() => ({
  list: vi.fn(),
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
  'keys.createKey': 'Create API Key',
  'keys.editKey': 'Edit API Key',
  'keys.nameLabel': 'Name',
  'keys.namePlaceholder': 'My API Key',
  'keys.groupLabel': 'Group',
  'keys.selectGroup': 'Select a group',
  'keys.searchGroup': 'Search groups...',
  'keys.groupSections.tokengate.title': 'TokenGate Capacity',
  'keys.groupSections.tokengate.description': 'Billed by TokenGate model usage',
  'keys.groupSections.connected.title': 'My Connected Accounts',
  'keys.groupSections.connected.description': 'Uses your own provider account',
  'keys.customKeyLabel': 'Custom key',
  'keys.enableCustomKey': 'Use custom key',
  'keys.statusLabel': 'Status',
  'keys.selectStatus': 'Select status',
  'common.name': 'Name',
  'common.status': 'Status',
  'common.actions': 'Actions',
  'common.refresh': 'Refresh',
}

vi.mock('@/api', () => ({
  keysAPI: {
    list,
    update: vi.fn(),
    toggleStatus: vi.fn(),
    create: vi.fn(),
    delete: vi.fn(),
    testConnection: vi.fn(),
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
  template: '<div><div v-for="row in data" :key="row.id"><slot name="cell-actions" :row="row" /></div></div>',
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

describe('KeysView group selector sections', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    list.mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      page_size: 10,
      pages: 0,
    })
    getDashboardApiKeysUsage.mockResolvedValue({ stats: {} })
    getAvailable.mockResolvedValue([
      {
        id: 1,
        name: 'default',
        description: 'Default TokenGate routing group',
        platform: 'anthropic',
        rate_multiplier: 1,
        is_exclusive: false,
        status: 'active',
        subscription_type: 'standard',
      },
      {
        id: 2,
        name: 'byo-anthropic-u2-a5',
        description: 'User-owned connected account capacity',
        platform: 'anthropic',
        rate_multiplier: 1,
        is_exclusive: true,
        status: 'active',
        subscription_type: 'standard',
        capacity_source: 'connected_account',
      },
    ])
    getUserGroupRates.mockResolvedValue({})
    getPublicSettings.mockResolvedValue({})
  })

  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('separates TokenGate groups from connected-account groups in the create API key selector', async () => {
    const wrapper = mount(KeysView, {
      attachTo: document.body,
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          TablePageLayout: TablePageLayoutStub,
          DataTable: DataTableStub,
          BaseDialog: BaseDialogStub,
          Pagination: true,
          ConfirmDialog: true,
          EmptyState: true,
          SearchInput: true,
          EndpointPopover: true,
          GroupBadge: { props: ['name'], template: '<span>{{ name }}</span>' },
          GroupOptionItem: { props: ['name'], template: '<div>{{ name }}</div>' },
          Icon: true,
          UseKeyModal: true,
          RouterLink: true,
        },
      },
    })

    await flushPromises()
    await nextTick()

    await wrapper.get('[data-tour="keys-create-btn"]').trigger('click')
    await nextTick()

    await wrapper.get('[data-tour="key-form-group"] button').trigger('click')
    await nextTick()

    const dropdownText = document.body.textContent || ''
    expect(dropdownText).toContain('TokenGate Capacity')
    expect(dropdownText).toContain('Billed by TokenGate model usage')
    expect(dropdownText).toContain('default')
    expect(dropdownText).toContain('My Connected Accounts')
    expect(dropdownText).toContain('Uses your own provider account')
    expect(dropdownText).toContain('byo-anthropic-u2-a5')
  })
})
