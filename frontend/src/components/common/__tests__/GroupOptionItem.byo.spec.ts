import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import GroupOptionItem from '@/components/common/GroupOptionItem.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => ({
        'keys.byoGroupLabel': 'BYO account',
        'keys.byoGroupDescription': 'Uses your connected AI account. TokenGate token balance is not charged.',
      }[key] ?? key),
    }),
  }
})

describe('GroupOptionItem BYO capacity source', () => {
  it('explains that connected-account groups do not use TokenGate balance billing', () => {
    const wrapper = mount(GroupOptionItem, {
      props: {
        name: 'My OpenAI',
        platform: 'openai',
        rateMultiplier: 2,
        capacitySource: 'connected_account',
      },
      global: {
        stubs: {
          GroupBadge: {
            props: ['name'],
            template: '<span>{{ name }}</span>',
          },
        },
      },
    })

    expect(wrapper.text()).toContain('My OpenAI')
    expect(wrapper.text()).toContain('BYO account')
    expect(wrapper.text()).toContain('TokenGate token balance is not charged')
    expect(wrapper.text()).not.toContain('2x')
  })
})
