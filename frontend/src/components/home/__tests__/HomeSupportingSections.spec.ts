import { mount, RouterLinkStub } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import en from '@/i18n/locales/en'
import HomeClosingCta from '../HomeClosingCta.vue'
import HomeHowItWorks from '../HomeHowItWorks.vue'
import HomePublicFooter from '../HomePublicFooter.vue'
import HomeSharedCapabilities from '../HomeSharedCapabilities.vue'

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

const global = {
  stubs: {
    RouterLink: RouterLinkStub,
    Icon: { template: '<span />' }
  }
}

describe('homepage supporting sections', () => {
  it('explains the shared control layer and compact client flow', () => {
    const shared = mount(HomeSharedCapabilities, { global })
    const flow = mount(HomeHowItWorks, { global })

    expect(shared.text()).toContain('One TokenGate control layer')
    expect(shared.text()).toContain('Explicit routing')
    expect(flow.text()).toContain('Three steps from capacity to API')
    expect(flow.text()).toContain('Claude Code')
    expect(flow.text()).toContain('Codex CLI')
    expect(flow.text()).not.toContain('ANTHROPIC_BASE_URL')
  })

  it('routes the closing action and footer navigation', () => {
    const closing = mount(HomeClosingCta, { global })
    const footer = mount(HomePublicFooter, {
      props: {
        siteName: 'TokenGate',
        currentYear: 2026,
        docUrl: '/docs',
        docUrlExternal: false
      },
      global
    })

    expect(closing.getComponent(RouterLinkStub).props('to')).toBe('/pricing')

    const footerRoutes = footer.findAllComponents(RouterLinkStub).map((link) => link.props('to'))
    expect(footerRoutes).toContain('/pricing')
    expect(footerRoutes).toContain('/docs')
    expect(footerRoutes).toContain('/support')
    expect(footerRoutes).toContain('/legal/privacy')
  })
})
