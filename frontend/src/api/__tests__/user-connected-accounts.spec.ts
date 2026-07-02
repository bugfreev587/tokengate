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
  exchangeConnectedAccountCode,
  deleteConnectedAccount,
  generateConnectedAccountAuthUrl,
  generateOpenAIConnectedAccountAuthUrl,
  getConnectedAccountModels,
  listConnectedAccounts,
  refreshConnectedAccount,
  refreshConnectedAccountModels,
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

  it('generates an OpenAI OAuth URL with the callback copy-page redirect URI', async () => {
    post.mockResolvedValue({
      data: {
        auth_url: 'https://auth.example.com/start',
        session_id: 'session-123',
      },
    })

    const result = await generateOpenAIConnectedAccountAuthUrl({
      proxy_id: 7,
      redirect_uri: 'https://tokengate.example.com/auth/callback',
    })

    expect(post).toHaveBeenCalledWith('/user/accounts/openai/auth-url', {
      proxy_id: 7,
      redirect_uri: 'https://tokengate.example.com/auth/callback',
    })
    expect(result.session_id).toBe('session-123')
  })

  it('generates provider OAuth URLs for Anthropic and Gemini', async () => {
    post.mockResolvedValueOnce({
      data: {
        auth_url: 'https://claude.example.com/start',
        session_id: 'claude-session',
      },
    })
    post.mockResolvedValueOnce({
      data: {
        auth_url: 'https://google.example.com/start',
        session_id: 'gemini-session',
        state: 'gemini-state',
      },
    })

    await expect(generateConnectedAccountAuthUrl('anthropic', {})).resolves.toEqual({
      auth_url: 'https://claude.example.com/start',
      session_id: 'claude-session',
    })
    await expect(generateConnectedAccountAuthUrl('gemini', {
      oauth_type: 'google_one',
      tier_id: 'google_ai_pro',
    })).resolves.toEqual({
      auth_url: 'https://google.example.com/start',
      session_id: 'gemini-session',
      state: 'gemini-state',
    })

    expect(post).toHaveBeenNthCalledWith(1, '/user/accounts/anthropic/auth-url', {})
    expect(post).toHaveBeenNthCalledWith(2, '/user/accounts/gemini/auth-url', {
      oauth_type: 'google_one',
      tier_id: 'google_ai_pro',
    })
  })

  it('exchanges provider OAuth code and refreshes/deletes owned accounts', async () => {
    const account: ConnectedAccountSummary = {
      id: 12,
      name: 'Gemini Main',
      platform: 'gemini',
      type: 'oauth',
      status: 'active',
      capacity_source: 'connected_account',
      created_at: '2026-07-01T09:00:00Z',
      updated_at: '2026-07-01T09:00:00Z',
    }
    post.mockResolvedValueOnce({ data: account })
    post.mockResolvedValueOnce({ data: account })
    del.mockResolvedValueOnce({ data: { deleted: true } })

    await expect(exchangeConnectedAccountCode('gemini', {
      session_id: 'session-123',
      code: 'oauth-code',
      state: 'state-123',
      oauth_type: 'google_one',
      name: 'Gemini Main',
    })).resolves.toEqual(account)
    await expect(refreshConnectedAccount(12)).resolves.toEqual(account)
    await expect(deleteConnectedAccount(12)).resolves.toEqual({ deleted: true })

    expect(post).toHaveBeenNthCalledWith(1, '/user/accounts/gemini/exchange-code', {
      session_id: 'session-123',
      code: 'oauth-code',
      state: 'state-123',
      oauth_type: 'google_one',
      name: 'Gemini Main',
    })
    expect(post).toHaveBeenNthCalledWith(2, '/user/accounts/12/refresh')
    expect(del).toHaveBeenCalledWith('/user/accounts/12')
  })

  it('exchanges OpenAI OAuth codes with the callback copy-page redirect URI', async () => {
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

    await expect(exchangeConnectedAccountCode('openai', {
      session_id: 'session-123',
      code: 'oauth-code',
      state: 'state-123',
      redirect_uri: 'https://tokengate.example.com/auth/callback',
      name: 'OpenAI Main',
    })).resolves.toEqual(account)

    expect(post).toHaveBeenCalledWith('/user/accounts/openai/exchange-code', {
      session_id: 'session-123',
      code: 'oauth-code',
      state: 'state-123',
      redirect_uri: 'https://tokengate.example.com/auth/callback',
      name: 'OpenAI Main',
    })
  })

  it('loads available models for an owned connected account', async () => {
    get.mockResolvedValueOnce({
      data: [
        { id: 'gpt-5.4', display_name: 'GPT-5.4' },
      ],
    })

    await expect(getConnectedAccountModels(12)).resolves.toEqual([
      { id: 'gpt-5.4', display_name: 'GPT-5.4' },
    ])

    expect(get).toHaveBeenCalledWith('/user/accounts/12/models')
  })

  it('refreshes available models for an owned connected account', async () => {
    post.mockResolvedValueOnce({
      data: [
        { id: 'claude-sonnet-5', display_name: 'Claude Sonnet 5' },
      ],
    })

    await expect(refreshConnectedAccountModels(12)).resolves.toEqual([
      { id: 'claude-sonnet-5', display_name: 'Claude Sonnet 5' },
    ])

    expect(post).toHaveBeenCalledWith('/user/accounts/12/models/refresh')
  })
})
