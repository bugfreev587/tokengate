import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import LaunchReadinessView from '../LaunchReadinessView.vue'

const appStoreState = vi.hoisted(() => ({
  publicSettingsLoaded: true,
  cachedPublicSettings: null as null | Record<string, unknown>,
  fetchPublicSettings: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/stores', () => ({
  useAppStore: () => appStoreState,
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      locale: { value: 'en-US' },
    }),
  }
})

function mountView() {
  return mount(LaunchReadinessView, {
    global: {
      stubs: {
        RouterLink: {
          props: ['to'],
          template: '<a><slot /></a>',
        },
      },
    },
  })
}

describe('admin LaunchReadinessView', () => {
  beforeEach(() => {
    appStoreState.publicSettingsLoaded = true
    appStoreState.cachedPublicSettings = {
      site_name: 'TokenGate',
      site_subtitle: 'AI API Gateway',
      contact_info: 'support@example.com',
      registration_enabled: true,
      password_reset_enabled: true,
      email_verify_enabled: true,
      payment_enabled: true,
    }
    appStoreState.fetchPublicSettings.mockReset()
    appStoreState.fetchPublicSettings.mockResolvedValue(appStoreState.cachedPublicSettings)
    appStoreState.showSuccess.mockReset()
  })

  it('marks the public launch gate ready when critical settings are enabled', async () => {
    const wrapper = mountView()
    await flushPromises()

    const text = wrapper.text()
    expect(text).toContain('Production launch checklist')
    expect(text).toContain('Private beta gate')
    expect(text).toContain('Public self-serve gate')
    expect(text).toContain('Ready')
    expect(text).toContain('Registration open')
    expect(text).toContain('Password reset enabled')
    expect(text).toContain('Email verification enabled')
    expect(text).toContain('Payment enabled')
    expect(text).not.toContain('Blocked')
  })

  it('surfaces public release blockers from public settings', async () => {
    appStoreState.cachedPublicSettings = {
      site_name: 'TokenGate',
      contact_info: '',
      registration_enabled: false,
      password_reset_enabled: false,
      email_verify_enabled: false,
      payment_enabled: false,
    }

    const wrapper = mountView()
    await flushPromises()

    const text = wrapper.text()
    expect(text).toContain('Public self-serve gate')
    expect(text).toContain('Blocked')
    expect(text).toContain('Support contact missing')
    expect(text).toContain('Self-serve registration disabled')
    expect(text).toContain('Password reset disabled')
    expect(text).toContain('Email verification disabled')
    expect(text).toContain('Payment disabled')
  })

  it('refreshes public settings on first load and explicit refresh', async () => {
    appStoreState.publicSettingsLoaded = false

    const wrapper = mountView()
    await flushPromises()

    expect(appStoreState.fetchPublicSettings).toHaveBeenCalledWith(true)

    appStoreState.fetchPublicSettings.mockClear()
    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(appStoreState.fetchPublicSettings).toHaveBeenCalledWith(true)
  })
})
