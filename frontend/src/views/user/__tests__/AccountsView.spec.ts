import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import AccountsView from '@/views/user/AccountsView.vue'

const {
  deleteConnectedAccount,
  exchangeConnectedAccountCode,
  generateConnectedAccountAuthUrl,
  getConnectedAccountModels,
  listConnectedAccounts,
  refreshConnectedAccount,
  refreshConnectedAccountModels,
  showError,
  showSuccess,
  showWarning,
} = vi.hoisted(() => ({
  deleteConnectedAccount: vi.fn(),
  exchangeConnectedAccountCode: vi.fn(),
  generateConnectedAccountAuthUrl: vi.fn(),
  getConnectedAccountModels: vi.fn(),
  listConnectedAccounts: vi.fn(),
  refreshConnectedAccount: vi.fn(),
  refreshConnectedAccountModels: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
  showWarning: vi.fn(),
}))

vi.mock('@/api/user', () => ({
  deleteConnectedAccount,
  exchangeConnectedAccountCode,
  generateConnectedAccountAuthUrl,
  getConnectedAccountModels,
  listConnectedAccounts,
  refreshConnectedAccount,
  refreshConnectedAccountModels,
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
    showWarning,
  }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => {
        if (!params) return key
        return `${key}:${JSON.stringify(params)}`
      },
    }),
  }
})

const AppLayoutStub = { template: '<div><slot /></div>' }
const DataTableStub = {
  props: ['data', 'loading'],
  template: `
    <div data-testid="accounts-table">
      <div v-if="loading" data-testid="loading">loading</div>
      <slot v-else-if="!data.length" name="empty" />
      <div v-for="row in data" :key="row.id" data-testid="account-row">
        <slot name="cell-name" :row="row" :value="row.name" />
        <slot name="cell-platform" :row="row" />
        <slot name="cell-status" :row="row" />
        <slot name="cell-group" :row="row" />
        <slot name="cell-actions" :row="row" />
      </div>
    </div>
  `,
}
const ConfirmDialogStub = {
  props: ['show', 'title', 'message'],
  emits: ['confirm', 'cancel'],
  template: `
    <section v-if="show" data-testid="confirm-dialog">
      <h2>{{ title }}</h2>
      <p>{{ message }}</p>
      <button data-testid="confirm-delete" @click="$emit('confirm')">confirm</button>
      <button data-testid="cancel-delete" @click="$emit('cancel')">cancel</button>
    </section>
  `,
}
const AccountTestModalStub = {
  props: ['show', 'account', 'scope'],
  emits: ['close'],
  template: `
    <section v-if="show" data-testid="account-test-modal">
      <span data-testid="account-test-modal-name">{{ account?.name }}</span>
      <span data-testid="account-test-modal-scope">{{ scope }}</span>
      <button data-testid="close-test-modal" @click="$emit('close')">close</button>
    </section>
  `,
}
const BaseDialogStub = {
  props: ['show', 'title'],
  emits: ['close'],
  template: `
    <section v-if="show" data-testid="base-dialog">
      <h2>{{ title }}</h2>
      <slot />
      <slot name="footer" />
      <button data-testid="close-base-dialog" @click="$emit('close')">close</button>
    </section>
  `,
}

function mountView() {
  return mount(AccountsView, {
    global: {
      stubs: {
        AccountTestModal: AccountTestModalStub,
        AppLayout: AppLayoutStub,
        BaseDialog: BaseDialogStub,
        ConfirmDialog: ConfirmDialogStub,
        DataTable: DataTableStub,
        Icon: true,
        PlatformTypeBadge: true,
        StatusBadge: true,
      },
    },
  })
}

describe('User AccountsView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    window.sessionStorage.clear()
    window.history.replaceState({}, '', '/accounts')
    listConnectedAccounts.mockResolvedValue({
      items: [
        {
          id: 12,
          name: 'OpenAI Main',
          platform: 'openai',
          type: 'oauth',
          status: 'active',
          email: 'owner@example.com',
          plan_type: 'plus',
          group_id: 44,
          group_name: 'byo-openai-u7-a12',
          capacity_source: 'connected_account',
          created_at: '2026-07-01T09:00:00Z',
          updated_at: '2026-07-01T09:00:00Z',
          last_used_at: null,
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1,
    })
    generateConnectedAccountAuthUrl.mockResolvedValue({
      auth_url: 'https://auth.example.com/start?state=state-123',
      session_id: 'session-123',
      state: 'state-123',
    })
    exchangeConnectedAccountCode.mockResolvedValue({
      id: 13,
      name: 'OpenAI Added',
      platform: 'openai',
      type: 'oauth',
      status: 'active',
      capacity_source: 'connected_account',
      byo_enabled: false,
      byo_disabled_reason: 'subscription_inactive',
      created_at: '2026-07-01T09:00:00Z',
      updated_at: '2026-07-01T09:00:00Z',
    })
    refreshConnectedAccount.mockResolvedValue({})
    getConnectedAccountModels.mockResolvedValue([
      { id: 'claude-opus-4-8', type: 'model', display_name: 'Claude Opus 4.8', created_at: '2026-07-01T00:00:00Z' },
    ])
    refreshConnectedAccountModels.mockResolvedValue([
      { id: 'claude-sonnet-5', type: 'model', display_name: 'Claude Sonnet 5', created_at: '2026-07-01T00:00:00Z' },
    ])
    deleteConnectedAccount.mockResolvedValue({ deleted: true })
  })

  it('loads and renders user-owned connected accounts', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(listConnectedAccounts).toHaveBeenCalledWith(1, 20)
    expect(wrapper.text()).toContain('OpenAI Main')
    expect(wrapper.text()).toContain('owner@example.com')
    expect(wrapper.text()).toContain('byo-openai-u7-a12')
  })

  it('lets users choose a provider before generating an OAuth URL', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-testid="connect-account"]').trigger('click')
    expect(wrapper.text()).toContain('OpenAI')
    expect(wrapper.text()).toContain('Anthropic')
    expect(wrapper.text()).toContain('Gemini')

    await wrapper.get('[data-testid="provider-option-gemini"]').trigger('click')
    await flushPromises()

    expect(generateConnectedAccountAuthUrl).toHaveBeenCalledWith('gemini', {
      redirect_uri: expect.any(String),
      oauth_type: 'google_one',
    })
    expect(wrapper.text()).toContain('https://auth.example.com/start')

    await wrapper.get('[data-testid="oauth-code"]').setValue('oauth-code')
    await wrapper.get('[data-testid="oauth-state"]').setValue('state-123')
    await wrapper.get('[data-testid="account-name"]').setValue('Gemini Main')
    await wrapper.get('[data-testid="finish-oauth"]').trigger('click')
    await flushPromises()

    expect(exchangeConnectedAccountCode).toHaveBeenCalledWith('gemini', {
      session_id: 'session-123',
      code: 'oauth-code',
      state: 'state-123',
      redirect_uri: expect.any(String),
      oauth_type: 'google_one',
      name: 'Gemini Main',
    })
    expect(showSuccess).toHaveBeenCalled()
    expect(showWarning).toHaveBeenCalledWith('userAccounts.subscriptionRequiredAfterConnect')
    expect(listConnectedAccounts).toHaveBeenCalledTimes(2)
  })

  it('uses the OAuth callback copy page for OpenAI connections', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-testid="connect-account"]').trigger('click')
    await wrapper.get('[data-testid="provider-option-openai"]').trigger('click')
    await flushPromises()

    expect(generateConnectedAccountAuthUrl).toHaveBeenCalledWith('openai', {
      redirect_uri: expect.stringMatching(/\/auth\/callback$/),
    })

    await wrapper.get('[data-testid="oauth-code"]').setValue('oauth-code')
    await wrapper.get('[data-testid="oauth-state"]').setValue('state-123')
    await wrapper.get('[data-testid="finish-oauth"]').trigger('click')
    await flushPromises()

    const exchangePayload = exchangeConnectedAccountCode.mock.calls[0]?.[1]
    expect(exchangeConnectedAccountCode).toHaveBeenCalledWith('openai', expect.any(Object))
    expect(exchangePayload).toMatchObject({
      redirect_uri: expect.stringMatching(/\/auth\/callback$/),
    })
  })

  it('does not reopen stale OAuth drafts on normal page entry', async () => {
    window.sessionStorage.setItem(
      'tokengate:user-connected-account-oauth',
      JSON.stringify({
        provider: 'openai',
        authUrl: 'https://auth.example.com/stale-openai-draft',
        sessionId: 'stale-session',
        redirectUri: 'https://api.tokengate.to/auth/callback',
        oauthState: 'stale-state',
      })
    )

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).not.toContain('https://auth.example.com/stale-openai-draft')
    expect(wrapper.find('[data-testid="cancel-oauth"]').exists()).toBe(false)
    expect(window.sessionStorage.getItem('tokengate:user-connected-account-oauth')).toBeNull()
  })

  it('lets users cancel an in-progress OAuth draft', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-testid="connect-account"]').trigger('click')
    await wrapper.get('[data-testid="provider-option-openai"]').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('https://auth.example.com/start')
    expect(window.sessionStorage.getItem('tokengate:user-connected-account-oauth')).not.toBeNull()

    await wrapper.get('[data-testid="cancel-oauth"]').trigger('click')
    await flushPromises()

    expect(wrapper.text()).not.toContain('https://auth.example.com/start')
    expect(wrapper.find('[data-testid="cancel-oauth"]').exists()).toBe(false)
    expect(window.sessionStorage.getItem('tokengate:user-connected-account-oauth')).toBeNull()
  })

  it('extracts OpenAI code and state from a pasted localhost callback URL', async () => {
    generateConnectedAccountAuthUrl.mockResolvedValueOnce({
      auth_url: 'https://auth.example.com/start',
      session_id: 'session-parse-callback',
    })

    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-testid="connect-account"]').trigger('click')
    await wrapper.get('[data-testid="provider-option-openai"]').trigger('click')
    await flushPromises()

    await wrapper.get('[data-testid="oauth-code"]').setValue(
      'http://localhost:1455/auth/callback?code=ac_test-code&scope=openid+profile+email+offline_access&state=state-from-callback'
    )
    await flushPromises()

    expect((wrapper.get('[data-testid="oauth-code"]').element as HTMLInputElement).value).toBe('ac_test-code')
    expect((wrapper.get('[data-testid="oauth-state"]').element as HTMLInputElement).value).toBe('state-from-callback')

    await wrapper.get('[data-testid="finish-oauth"]').trigger('click')
    await flushPromises()

    expect(exchangeConnectedAccountCode).toHaveBeenCalledWith('openai', {
      session_id: 'session-parse-callback',
      code: 'ac_test-code',
      state: 'state-from-callback',
      redirect_uri: expect.stringMatching(/\/auth\/callback$/),
      name: undefined,
    })
  })

  it('refreshes and deletes an account from row actions', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-testid="refresh-account-12"]').trigger('click')
    await flushPromises()
    expect(refreshConnectedAccount).toHaveBeenCalledWith(12)

    await wrapper.get('[data-testid="delete-account-12"]').trigger('click')
    await wrapper.get('[data-testid="confirm-delete"]').trigger('click')
    await flushPromises()

    expect(deleteConnectedAccount).toHaveBeenCalledWith(12)
    expect(showSuccess).toHaveBeenCalled()
  })

  it('opens a direct connected account test modal from row actions', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-testid="test-account-12"]').trigger('click')

    expect(wrapper.get('[data-testid="account-test-modal-name"]').text()).toBe('OpenAI Main')
    expect(wrapper.get('[data-testid="account-test-modal-scope"]').text()).toBe('user')
  })

  it('shows and refreshes models for an Anthropic connected account', async () => {
    listConnectedAccounts.mockResolvedValueOnce({
      items: [
        {
          id: 12,
          name: 'Claude Main',
          platform: 'anthropic',
          type: 'oauth',
          status: 'active',
          email: 'owner@example.com',
          group_id: 44,
          group_name: 'byo-anthropic-u7-a12',
          capacity_source: 'connected_account',
          created_at: '2026-07-01T09:00:00Z',
          updated_at: '2026-07-01T09:00:00Z',
          last_used_at: null,
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1,
    })
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-testid="models-account-12"]').trigger('click')
    await flushPromises()

    expect(getConnectedAccountModels).toHaveBeenCalledWith(12)
    expect(wrapper.get('[data-testid="connected-account-models-modal"]').text()).toContain('Claude Opus 4.8')

    await wrapper.get('[data-testid="refresh-connected-models"]').trigger('click')
    await flushPromises()

    expect(refreshConnectedAccountModels).toHaveBeenCalledWith(12)
    expect(wrapper.get('[data-testid="connected-account-models-modal"]').text()).toContain('Claude Sonnet 5')
    expect(showSuccess).toHaveBeenCalledWith('userAccounts.modelsRefreshed')
  })
})
