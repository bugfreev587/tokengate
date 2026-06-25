import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import LoginView from '@/views/auth/LoginView.vue'

const { getPublicSettingsMock, showErrorMock, showWarningMock, showSuccessMock, loginMock, pushMock } = vi.hoisted(() => ({
  getPublicSettingsMock: vi.fn(),
  showErrorMock: vi.fn(),
  showWarningMock: vi.fn(),
  showSuccessMock: vi.fn(),
  loginMock: vi.fn(),
  pushMock: vi.fn(),
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: pushMock,
    currentRoute: {
      value: {
        query: {},
      },
    },
  }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

vi.mock('@/components/layout', () => ({
  AuthLayout: defineComponent({
    template: '<main><slot /><footer><slot name="footer" /></footer></main>',
  }),
}))

vi.mock('@/stores', () => ({
  useAuthStore: () => ({
    login: (...args: any[]) => loginMock(...args),
    login2FA: vi.fn(),
  }),
  useAppStore: () => ({
    showError: (...args: any[]) => showErrorMock(...args),
    showWarning: (...args: any[]) => showWarningMock(...args),
    showSuccess: (...args: any[]) => showSuccessMock(...args),
  }),
}))

vi.mock('@/api/auth', async () => {
  const actual = await vi.importActual<typeof import('@/api/auth')>('@/api/auth')
  return {
    ...actual,
    getPublicSettings: (...args: any[]) => getPublicSettingsMock(...args),
    isTotp2FARequired: (response: any) => response?.requires_2fa === true,
    isWeChatWebOAuthEnabled: () => false,
  }
})

vi.mock('@/utils/oauthAffiliate', () => ({
  clearAllAffiliateReferralCodes: vi.fn(),
}))

const RouterLinkStub = defineComponent({
  props: {
    to: {
      type: [String, Object],
      required: true,
    },
  },
  template: '<a :href="typeof to === \'string\' ? to : to.path"><slot /></a>',
})

describe('LoginView password reset entry', () => {
  beforeEach(() => {
    getPublicSettingsMock.mockReset()
    showErrorMock.mockReset()
    showWarningMock.mockReset()
    showSuccessMock.mockReset()
    loginMock.mockReset()
    pushMock.mockReset()
    localStorage.clear()
    sessionStorage.clear()
  })

  it('shows forgot password link in backend mode when password reset is enabled', async () => {
    getPublicSettingsMock.mockResolvedValue({
      turnstile_enabled: false,
      turnstile_site_key: '',
      linuxdo_oauth_enabled: false,
      wechat_oauth_enabled: false,
      wechat_oauth_open_enabled: false,
      wechat_oauth_mp_enabled: false,
      wechat_oauth_mobile_enabled: false,
      oidc_oauth_enabled: false,
      oidc_oauth_provider_name: 'OIDC',
      github_oauth_enabled: false,
      google_oauth_enabled: false,
      backend_mode_enabled: true,
      password_reset_enabled: true,
      login_agreement_enabled: false,
      login_agreement_documents: [],
    })

    const wrapper = mount(LoginView, {
      global: {
        stubs: {
          RouterLink: RouterLinkStub,
          Icon: true,
          TurnstileWidget: true,
          LoginAgreementPrompt: true,
          EmailOAuthButtons: true,
          LinuxDoOAuthSection: true,
          WechatOAuthSection: true,
          OidcOAuthSection: true,
          TotpLoginModal: true,
        },
      },
    })

    await flushPromises()

    const forgotPasswordLink = wrapper.find('a[href="/forgot-password"]')
    expect(forgotPasswordLink.exists()).toBe(true)
    expect(forgotPasswordLink.text()).toBe('auth.forgotPassword')
  })
})
