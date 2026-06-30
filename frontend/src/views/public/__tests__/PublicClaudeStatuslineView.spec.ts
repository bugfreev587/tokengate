import { mount, RouterLinkStub } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import PublicClaudeStatuslineView from '../PublicClaudeStatuslineView.vue'

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    cachedPublicSettings: null,
    siteName: 'TokenGate',
    siteLogo: '',
    fetchPublicSettings: vi.fn().mockResolvedValue(undefined),
  }),
}))

describe('PublicClaudeStatuslineView', () => {
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
})
