import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import AccountsView from '@/views/user/AccountsView.vue'

const {
  deleteConnectedAccount,
  exchangeConnectedAccountCode,
  generateConnectedAccountAuthUrl,
  listConnectedAccounts,
  refreshConnectedAccount,
  showError,
  showSuccess,
} = vi.hoisted(() => ({
  deleteConnectedAccount: vi.fn(),
  exchangeConnectedAccountCode: vi.fn(),
  generateConnectedAccountAuthUrl: vi.fn(),
  listConnectedAccounts: vi.fn(),
  refreshConnectedAccount: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/api/user', () => ({
  deleteConnectedAccount,
  exchangeConnectedAccountCode,
  generateConnectedAccountAuthUrl,
  listConnectedAccounts,
  refreshConnectedAccount,
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
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

function mountView() {
  return mount(AccountsView, {
    global: {
      stubs: {
        AppLayout: AppLayoutStub,
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
      created_at: '2026-07-01T09:00:00Z',
      updated_at: '2026-07-01T09:00:00Z',
    })
    refreshConnectedAccount.mockResolvedValue({})
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
})
