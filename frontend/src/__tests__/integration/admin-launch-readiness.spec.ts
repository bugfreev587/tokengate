import { describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { RouterView } from 'vue-router'
import { defineComponent, h } from 'vue'

const authStore = vi.hoisted(() => ({
  checkAuth: vi.fn(),
  isAuthenticated: true,
  isAdmin: true,
  isSimpleMode: false,
  hasPendingAuthSession: false,
}))

const appStore = vi.hoisted(() => ({
  siteName: 'TokenGate',
  backendModeEnabled: false,
  cachedPublicSettings: {
    site_name: 'TokenGate',
    site_subtitle: 'AI API Gateway',
    contact_info: 'support@example.com',
    registration_enabled: true,
    password_reset_enabled: true,
    email_verify_enabled: true,
    payment_enabled: true,
    backend_mode_enabled: false,
  } as Record<string, unknown>,
  publicSettingsLoaded: true,
  fetchPublicSettings: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => authStore,
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => appStore,
}))

vi.mock('@/stores', () => ({
  useAppStore: () => appStore,
}))

vi.mock('@/stores/adminSettings', () => ({
  useAdminSettingsStore: () => ({
    customMenuItems: [],
  }),
}))

vi.mock('@/i18n', () => ({
  i18n: {
    global: {
      t: (key: string) => (key === 'nav.launchReadiness' ? 'Launch Readiness' : key),
    },
  },
}))

vi.mock('@/composables/useNavigationLoading', () => ({
  useNavigationLoadingState: () => ({
    startNavigation: vi.fn(),
    endNavigation: vi.fn(),
    isLoading: { value: false },
  }),
}))

vi.mock('@/composables/useRoutePrefetch', () => ({
  useRoutePrefetch: () => ({
    triggerPrefetch: vi.fn(),
    cancelPendingPrefetch: vi.fn(),
    resetPrefetchState: vi.fn(),
  }),
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

const TestApp = defineComponent({
  setup() {
    return () => h(RouterView)
  },
})

describe('admin launch readiness integration', () => {
  it('allows an admin to navigate to the launch readiness dashboard', async () => {
    vi.stubGlobal('scrollTo', vi.fn())

    const { default: router } = await import('@/router')
    const wrapper = mount(TestApp, {
      global: {
        plugins: [router],
      },
    })

    await router.push('/admin/launch-readiness')
    await router.isReady()
    await flushPromises()

    expect(router.currentRoute.value.name).toBe('AdminLaunchReadiness')
    expect(authStore.checkAuth).toHaveBeenCalled()
    expect(wrapper.text()).toContain('Production launch checklist')
    expect(wrapper.text()).toContain('Public self-serve gate')
    expect(wrapper.text()).toContain('Ready')
  })
})
