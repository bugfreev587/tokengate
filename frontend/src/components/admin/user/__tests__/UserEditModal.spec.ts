import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { AdminUser } from '@/types'
import UserEditModal from '../UserEditModal.vue'

const { updateUser, updateUserAttributeValues, showError, showSuccess } = vi.hoisted(() => ({
  updateUser: vi.fn(),
  updateUserAttributeValues: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: {
      update: updateUser
    },
    userAttributes: {
      updateUserAttributeValues
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn()
  })
}))

const user: AdminUser = {
  id: 7,
  email: 'user@example.com',
  username: 'user',
  role: 'user',
  balance: 0,
  concurrency: 5,
  rpm_limit: 0,
  status: 'active',
  allowed_groups: [],
  balance_notify_enabled: false,
  balance_notify_threshold: null,
  balance_notify_extra_emails: [],
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
  notes: ''
}

describe('UserEditModal', () => {
  beforeEach(() => {
    updateUser.mockReset()
    updateUserAttributeValues.mockReset()
    showError.mockReset()
    showSuccess.mockReset()
    updateUser.mockResolvedValue({ ...user, role: 'admin' })
  })

  it('submits role changes when promoting a user to admin', async () => {
    const wrapper = mount(UserEditModal, {
      props: {
        show: true,
        user
      },
      global: {
        stubs: {
          BaseDialog: {
            template: '<div><slot /><slot name="footer" /></div>'
          },
          UserAttributeForm: {
            template: '<div />'
          },
          Icon: true
        }
      }
    })

    await wrapper.get('[data-test="user-role-select"]').setValue('admin')
    await wrapper.get('form').trigger('submit.prevent')

    expect(updateUser).toHaveBeenCalledWith(
      7,
      expect.objectContaining({
        email: 'user@example.com',
        role: 'admin'
      })
    )
  })
})
