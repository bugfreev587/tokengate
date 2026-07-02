import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import AccountActionMenu from '../AccountActionMenu.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

describe('AccountActionMenu', () => {
  it('shows Refresh Models below Refresh Token for Anthropic OAuth accounts', async () => {
    const wrapper = mount(AccountActionMenu, {
      props: {
        show: true,
        account: {
          id: 42,
          name: 'Claude Main',
          platform: 'anthropic',
          type: 'oauth',
          status: 'active'
        } as any,
        position: { top: 10, left: 10 }
      },
      global: {
        stubs: {
          Icon: true,
          Teleport: true
        }
      }
    })

    const labels = wrapper.findAll('button').map((button) => button.text())
    expect(labels.indexOf('admin.accounts.refreshModels')).toBeGreaterThan(
      labels.indexOf('admin.accounts.refreshToken')
    )

    const refreshModelsButton = wrapper.findAll('button').find((button) => button.text() === 'admin.accounts.refreshModels')
    expect(refreshModelsButton).toBeTruthy()
    await refreshModelsButton!.trigger('click')

    expect(wrapper.emitted('refresh-models')?.[0]?.[0]).toMatchObject({ id: 42 })
  })
})
