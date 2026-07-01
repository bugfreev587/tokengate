import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import GroupBadge from '@/components/common/GroupBadge.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => ({
        'keys.byoGroupLabel': 'BYO account',
      }[key] ?? key),
    }),
  }
})

describe('GroupBadge BYO capacity source', () => {
  it('shows a BYO label instead of a token billing multiplier', () => {
    const wrapper = mount(GroupBadge, {
      props: {
        name: 'My OpenAI',
        platform: 'openai',
        rateMultiplier: 2,
        capacitySource: 'connected_account',
      },
      global: {
        stubs: {
          PlatformIcon: true,
        },
      },
    })

    expect(wrapper.text()).toContain('My OpenAI')
    expect(wrapper.text()).toContain('BYO account')
    expect(wrapper.text()).not.toContain('2x')
  })
})
