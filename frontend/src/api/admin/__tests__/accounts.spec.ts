import { beforeEach, describe, expect, it, vi } from 'vitest'

const { get, post } = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  apiClient: {
    get,
    post,
  },
}))

import { getAvailableModels, refreshModels } from '@/api/admin/accounts'

describe('admin accounts api', () => {
  beforeEach(() => {
    get.mockReset()
    post.mockReset()
  })

  it('loads and refreshes available models for an account', async () => {
    get.mockResolvedValueOnce({
      data: [
        { id: 'claude-opus-4-8', display_name: 'Claude Opus 4.8' },
      ],
    })
    post.mockResolvedValueOnce({
      data: [
        { id: 'claude-sonnet-5', display_name: 'Claude Sonnet 5' },
      ],
    })

    await expect(getAvailableModels(42)).resolves.toEqual([
      { id: 'claude-opus-4-8', display_name: 'Claude Opus 4.8' },
    ])
    await expect(refreshModels(42)).resolves.toEqual([
      { id: 'claude-sonnet-5', display_name: 'Claude Sonnet 5' },
    ])

    expect(get).toHaveBeenCalledWith('/admin/accounts/42/models')
    expect(post).toHaveBeenCalledWith('/admin/accounts/42/models/refresh')
  })
})
