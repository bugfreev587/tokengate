import { mount, RouterLinkStub } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import PublicClaudeStatuslineView from '../PublicClaudeStatuslineView.vue'

const routerReplace = vi.hoisted(() => vi.fn())

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    cachedPublicSettings: null,
    siteName: 'TokenGate',
    siteLogo: '',
    fetchPublicSettings: vi.fn().mockResolvedValue(undefined),
  }),
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({
    replace: routerReplace,
  }),
}))

describe('PublicClaudeStatuslineView', () => {
  beforeEach(() => {
    routerReplace.mockReset()
    window.history.replaceState(null, '', '/docs/cli/statusline')
  })

  it('documents TokenGate Claude Code statusline setup and mode selection', () => {
    const wrapper = mount(PublicClaudeStatuslineView, {
      global: {
        stubs: {
          RouterLink: RouterLinkStub,
        },
      },
    })

    const text = wrapper.text()

    expect(text).toContain('Claude Code statusline')
    expect(text).toContain('TOKENGATE_STATUSLINE_MODE')
    expect(text).toContain('--mode tokengate')
    expect(text).toContain('--mode claude')
    expect(text).toContain('TokenGate unavailable')
    expect(text).toContain('$12.34 30d')
    expect(text).toContain('https://api.tokengate.to')
    expect(text).toContain('statusLine')
    expect(text).toContain('tokengate')
    expect(text).toContain('claude')
  })

  it('links API sidebar entries back to the API reference page', () => {
    const wrapper = mount(PublicClaudeStatuslineView, {
      global: {
        stubs: {
          RouterLink: RouterLinkStub,
        },
      },
    })

    const imageGenerationLink = wrapper
      .findAllComponents(RouterLinkStub)
      .find((link) => link.text().includes('Image generation'))

    expect(imageGenerationLink?.props('to')).toBe('/docs#images-api')
  })

  it('redirects stale API hashes on the statusline route to the API reference page', () => {
    window.history.replaceState(null, '', '/docs/cli/statusline#images-api')

    mount(PublicClaudeStatuslineView, {
      global: {
        stubs: {
          RouterLink: RouterLinkStub,
        },
      },
    })

    expect(routerReplace).toHaveBeenCalledWith('/docs#images-api')
  })
})
