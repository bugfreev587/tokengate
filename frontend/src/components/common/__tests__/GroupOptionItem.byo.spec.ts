import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import GroupOptionItem from '../GroupOptionItem.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => {
      const messages: Record<string, string> = {
        'keys.byoGroupDisabledLabel': 'Needs subscription',
        'keys.byoGroupLabel': 'BYO account',
        'keys.byoGroupDescription': 'Uses your connected account',
        'keys.byoGroupDisabled.subscriptionInactive': 'Subscribe to enable this BYO account',
        'keys.byoGroupDisabled.accountDisabled': 'Account disabled',
        'keys.byoGroupDisabled.accountMissing': 'Account missing',
      }
      return messages[key] ?? key
    },
  }),
}))

describe('GroupOptionItem BYO status', () => {
  it('shows a subscription warning for disabled BYO groups', () => {
    const wrapper = mount(GroupOptionItem, {
      props: {
        name: 'byo-openai-u1-a2',
        platform: 'openai',
        capacitySource: 'connected_account',
        byoEnabled: false,
        byoDisabledReason: 'subscription_inactive',
      },
      global: {
        stubs: {
          PlatformIcon: true,
        },
      },
    })

    expect(wrapper.text()).toContain('Needs subscription')
    expect(wrapper.text()).toContain('Subscribe to enable this BYO account')
  })
})
