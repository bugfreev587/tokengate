import { beforeEach, describe, expect, it, vi } from 'vitest'

const post = vi.fn()

vi.mock('@/api/client', () => ({
  apiClient: {
    post
  }
}))

describe('auth session api', () => {
  beforeEach(() => {
    post.mockReset()
    post.mockResolvedValue({ data: {} })
    localStorage.clear()
  })

  it('posts logout even when only the HttpOnly refresh cookie is available', async () => {
    const { logout } = await import('@/api/auth')

    await logout()

    expect(post).toHaveBeenCalledWith('/auth/logout', {})
  })

  it('posts prepareSessionCookies to the session-cookie endpoint', async () => {
    const { prepareSessionCookies } = await import('@/api/auth')

    await prepareSessionCookies()

    expect(post).toHaveBeenCalledWith('/auth/session-cookie')
  })
})
