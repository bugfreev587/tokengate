import { beforeEach, describe, expect, it, vi } from 'vitest'

const { del, get, post } = vi.hoisted(() => ({
  del: vi.fn(),
  get: vi.fn(),
  post: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  apiClient: {
    delete: del,
    get,
    post,
  },
}))

import {
  deleteConnectedAccount,
  exchangeOpenAIConnectedAccountCode,
  generateOpenAIConnectedAccountAuthUrl,
  listConnectedAccounts,
  refreshOpenAIConnectedAccount,
  type ConnectedAccountSummary,
} from '@/api/user'

describe('user connected accounts api', () => {
  beforeEach(() => {
    del.mockReset()
    get.mockReset()
    post.mockReset()
  })

  it('lists user-owned connected accounts with pagination params', async () => {
    const account: ConnectedAccountSummary = {
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
    }
    get.mockResolvedValue({
      data: {
        items: [account],
        total: 1,
        page: 2,
        page_size: 10,
        pages: 1,
      },
    })

    const result = await listConnectedAccounts(2, 10)

    expect(get).toHaveBeenCalledWith('/user/accounts', {
      params: {
        page: 2,
        page_size: 10,
      },
    })
    expect(result.items).toEqual([account])
  })

  it('generates an OpenAI OAuth URL with optional redirect URI', async () => {
    post.mockResolvedValue({
      data: {
        auth_url: 'https://auth.example.com/start',
        session_id: 'session-123',
      },
    })

    const result = await generateOpenAIConnectedAccountAuthUrl({
      redirect_uri: 'https://tokengate.example.com/accounts',
    })

    expect(post).toHaveBeenCalledWith('/user/accounts/openai/auth-url', {
      redirect_uri: 'https://tokengate.example.com/accounts',
    })
    expect(result.session_id).toBe('session-123')
  })

  it('exchanges OpenAI OAuth code and refreshes/deletes owned accounts', async () => {
    const account: ConnectedAccountSummary = {
      id: 12,
      name: 'OpenAI Main',
      platform: 'openai',
      type: 'oauth',
      status: 'active',
      capacity_source: 'connected_account',
      created_at: '2026-07-01T09:00:00Z',
      updated_at: '2026-07-01T09:00:00Z',
    }
    post.mockResolvedValueOnce({ data: account })
    post.mockResolvedValueOnce({ data: account })
    del.mockResolvedValueOnce({ data: { deleted: true } })

    await expect(exchangeOpenAIConnectedAccountCode({
      session_id: 'session-123',
      code: 'oauth-code',
      state: 'state-123',
      name: 'OpenAI Main',
    })).resolves.toEqual(account)
    await expect(refreshOpenAIConnectedAccount(12)).resolves.toEqual(account)
    await expect(deleteConnectedAccount(12)).resolves.toEqual({ deleted: true })

    expect(post).toHaveBeenNthCalledWith(1, '/user/accounts/openai/exchange-code', {
      session_id: 'session-123',
      code: 'oauth-code',
      state: 'state-123',
      name: 'OpenAI Main',
    })
    expect(post).toHaveBeenNthCalledWith(2, '/user/accounts/12/refresh')
    expect(del).toHaveBeenCalledWith('/user/accounts/12')
  })
})
