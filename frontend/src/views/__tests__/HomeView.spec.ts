import { mount, RouterLinkStub } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import en from '@/i18n/locales/en'
import HomeView from '../HomeView.vue'

const { appState, authState } = vi.hoisted(() => ({
  authState: {
    isAuthenticated: false,
    isAdmin: false,
    user: null as null | { username?: string; email?: string },
    checkAuth: vi.fn()
  },
  appState: {
    cachedPublicSettings: null as null | Record<string, unknown>,
    siteName: 'TokenGate',
    siteLogo: '',
    siteSubtitle: '',
    docUrl: '/docs',
    publicSettingsLoaded: true,
    fetchPublicSettings: vi.fn()
  }
}))

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
  useAuthStore: () => authState,
  useAppStore: () => appState
}))

vi.mock('@/components/common/LocaleSwitcher.vue', () => ({
  default: {
    name: 'LocaleSwitcher',
    template: '<span data-locale-switcher />'
  }
}))

const mountHome = () =>
  mount(HomeView, {
    global: {
      stubs: {
        RouterLink: RouterLinkStub,
        LocaleSwitcher: { template: '<span data-locale-switcher />' },
        Icon: { template: '<span />' }
      }
    }
  })

describe('HomeView', () => {
  beforeEach(() => {
    authState.isAuthenticated = false
    authState.isAdmin = false
    authState.user = null
    authState.checkAuth.mockClear()

    appState.cachedPublicSettings = null
    appState.siteName = 'TokenGate'
    appState.siteLogo = ''
    appState.siteSubtitle = ''
    appState.docUrl = '/docs'
    appState.publicSettingsLoaded = true
    appState.fetchPublicSettings.mockClear()

    localStorage.clear()
    document.documentElement.classList.remove('dark')

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

  it('renders both capacity modes and the compact product journey', () => {
    const wrapper = mountHome()
    const text = wrapper.text()

    expect(wrapper.get('[data-mode="usage-based"]').text()).toContain('Usage-based')
    expect(wrapper.get('[data-mode="byo"]').text()).toContain('BYO')
    expect(text).toContain('One TokenGate control layer')
    expect(text).toContain('Explicit routing')
    expect(text).toContain('Three steps from capacity to API')
    expect(text).toContain('Claude Code')
    expect(text).toContain('Codex CLI')
    expect(text).not.toContain('ANTHROPIC_BASE_URL')
    expect(text).not.toContain('[model_providers.tokengate]')

    expect(wrapper.getComponent('[data-mode="usage-based"]').props('to')).toBe('/pricing')
    expect(wrapper.getComponent('[data-mode="byo"]').props('to')).toBe('/pricing')
  })

  it('uses the configured internal Docs route', () => {
    appState.docUrl = '/docs/start'
    const wrapper = mountHome()

    const routes = wrapper.findAllComponents(RouterLinkStub).map((link) => link.props('to'))
    expect(routes).toContain('/docs/start')
  })

  it('uses secure external Docs links when configured', () => {
    appState.docUrl = 'https://docs.example.com/token-gate'
    const wrapper = mountHome()
    const links = wrapper.findAll('a[href="https://docs.example.com/token-gate"]')

    expect(links.length).toBeGreaterThanOrEqual(2)
    expect(links.every((link) => link.attributes('target') === '_blank')).toBe(true)
    expect(links.every((link) => link.attributes('rel') === 'noopener noreferrer')).toBe(true)
  })

  it('shows Login when signed out and Dashboard when signed in', () => {
    const signedOut = mountHome()
    expect(signedOut.text()).toContain('Login')
    expect(signedOut.findAllComponents(RouterLinkStub).some((link) => link.props('to') === '/login')).toBe(
      true
    )

    authState.isAuthenticated = true
    authState.user = { username: 'river' }
    const signedIn = mountHome()

    expect(signedIn.text()).toContain('Dashboard')
    expect(
      signedIn.findAllComponents(RouterLinkStub).some((link) => link.props('to') === '/dashboard')
    ).toBe(true)
  })

  it('renders configured home HTML instead of the default homepage', () => {
    appState.cachedPublicSettings = {
      home_content: '<main data-custom-home>Custom home</main>'
    }
    const wrapper = mountHome()

    expect(wrapper.find('[data-custom-home]').exists()).toBe(true)
    expect(wrapper.find('[data-mode="usage-based"]').exists()).toBe(false)
  })

  it('renders a configured home URL in an iframe instead of the default homepage', () => {
    appState.cachedPublicSettings = {
      home_content: 'https://example.com/custom-home'
    }
    const wrapper = mountHome()

    expect(wrapper.get('iframe').attributes('src')).toBe('https://example.com/custom-home')
    expect(wrapper.find('[data-mode="byo"]').exists()).toBe(false)
  })
})
