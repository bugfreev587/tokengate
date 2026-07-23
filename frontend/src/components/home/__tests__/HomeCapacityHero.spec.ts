import { mount, RouterLinkStub } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import en from '@/i18n/locales/en'
import HomeCapacityHero from '../HomeCapacityHero.vue'

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

describe('HomeCapacityHero', () => {
  it('presents Usage-based and BYO as equal pricing choices', () => {
    const wrapper = mount(HomeCapacityHero, {
      props: {
        siteSubtitle: 'Subscription-native AI API gateway',
        docUrl: '/docs',
        docUrlExternal: false
      },
      global: {
        stubs: {
          RouterLink: RouterLinkStub,
          Icon: { template: '<span />' }
        }
      }
    })

    expect(wrapper.get('[data-mode="usage-based"]').text()).toContain('Usage-based')
    expect(wrapper.get('[data-mode="byo"]').text()).toContain('BYO')
    expect(wrapper.get('[data-mode="byo"]').text()).toContain(
      'No TokenGate prepaid balance deduction'
    )

    const pricingLinks = wrapper
      .findAllComponents(RouterLinkStub)
      .filter((link) => link.props('to') === '/pricing')

    expect(pricingLinks).toHaveLength(3)
    expect(wrapper.text()).not.toMatch(/\$19|7-day|minimum top-up/i)
  })

  it('uses the configured internal documentation route', () => {
    const wrapper = mount(HomeCapacityHero, {
      props: {
        siteSubtitle: 'Subscription-native AI API gateway',
        docUrl: '/docs/start',
        docUrlExternal: false
      },
      global: {
        stubs: {
          RouterLink: RouterLinkStub,
          Icon: { template: '<span />' }
        }
      }
    })

    expect(
      wrapper.findAllComponents(RouterLinkStub).some((link) => link.props('to') === '/docs/start')
    ).toBe(true)
  })
})
