import { mount, RouterLinkStub } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import en from '@/i18n/locales/en'
import HomeView from '../HomeView.vue'

const t = (path: string) => {
  const value = path.split('.').reduce<unknown>((current, segment) => {
    if (current && typeof current === 'object' && segment in current) {
      return (current as Record<string, unknown>)[segment]
    }
    return undefined
  }, en)

  return typeof value === 'string' ? value : path
}

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t })
}))

vi.mock('@/stores', () => ({
  useAuthStore: () => ({
    isAuthenticated: false,
    isAdmin: false,
    user: null,
    checkAuth: vi.fn()
  }),
  useAppStore: () => ({
    cachedPublicSettings: null,
    siteName: 'TokenGate',
    siteLogo: '',
    siteSubtitle: '',
    docUrl: '/docs',
    publicSettingsLoaded: true,
    fetchPublicSettings: vi.fn()
  })
}))

vi.mock('@/components/common/LocaleSwitcher.vue', () => ({
  default: {
    name: 'LocaleSwitcher',
    template: '<span />'
  }
}))

describe('HomeView', () => {
  beforeEach(() => {
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: vi.fn().mockImplementation((query: string) => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn()
      }))
    })
  })

  it('shows Claude Code CLI and Codex CLI setup guidance on the default landing page', () => {
    const wrapper = mount(HomeView, {
      global: {
        stubs: {
          RouterLink: RouterLinkStub,
          LocaleSwitcher: {
            template: '<span />'
          },
          Icon: {
            template: '<span />'
          }
        }
      }
    })

    const text = wrapper.text()

    expect(text).toContain('Claude Code CLI')
    expect(text).toContain('Codex CLI')
    expect(text).toContain('ANTHROPIC_BASE_URL')
    expect(text).toContain('ANTHROPIC_AUTH_TOKEN')
    expect(text).toContain('[model_providers.tokengate]')
    expect(text).toContain('wire_api = "responses"')
  })
})
