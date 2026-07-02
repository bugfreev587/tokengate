import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent } from 'vue'
import AccountTestModal from '../AccountTestModal.vue'

const { getAvailableModelsMock, refreshModelsMock, getConnectedAccountModelsMock, refreshConnectedAccountModelsMock } = vi.hoisted(() => ({
  getAvailableModelsMock: vi.fn(),
  refreshModelsMock: vi.fn(),
  getConnectedAccountModelsMock: vi.fn(),
  refreshConnectedAccountModelsMock: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      getAvailableModels: getAvailableModelsMock,
      refreshModels: refreshModelsMock
    }
  }
}))

vi.mock('@/api/user', () => ({
  getConnectedAccountModels: getConnectedAccountModelsMock,
  refreshConnectedAccountModels: refreshConnectedAccountModelsMock
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn()
  })
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    apiBaseUrl: '',
    showError: vi.fn(),
    showSuccess: vi.fn()
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

const BaseDialogStub = defineComponent({
  name: 'BaseDialog',
  props: { show: { type: Boolean, default: false } },
  template: '<div v-if="show"><slot /><slot name="footer" /></div>'
})

const SelectStub = defineComponent({
  name: 'SelectStub',
  props: {
    modelValue: { type: [String, Number, Boolean, null], default: '' },
    options: { type: Array, default: () => [] },
    valueKey: { type: String, default: 'value' },
    labelKey: { type: String, default: 'label' }
  },
  emits: ['update:modelValue'],
  template: `
    <select
      v-bind="$attrs"
      :value="modelValue"
      @change="$emit('update:modelValue', $event.target.value)"
    >
      <option
        v-for="option in options"
        :key="option[valueKey]"
        :value="option[valueKey]"
      >
        {{ option[labelKey] }}
      </option>
    </select>
  `
})

const TextAreaStub = defineComponent({
  name: 'TextArea',
  props: {
    modelValue: { type: String, default: '' }
  },
  emits: ['update:modelValue'],
  template: `
    <textarea
      v-bind="$attrs"
      :value="modelValue"
      @input="$emit('update:modelValue', $event.target.value)"
    />
  `
})

function createStreamResponse(lines: string[] = []) {
  const encoder = new TextEncoder()
  const chunks = lines.map((line) => encoder.encode(line))
  let index = 0

  return {
    ok: true,
    body: {
      getReader: () => ({
        read: vi.fn().mockImplementation(async () => {
          if (index < chunks.length) {
            return { done: false, value: chunks[index++] }
          }
          return { done: true, value: undefined }
        })
      })
    }
  } as Response
}

function buildAccount() {
  return {
    id: 1,
    name: 'OpenAI OAuth',
    platform: 'openai',
    type: 'oauth',
    status: 'active',
    credentials: {},
    extra: {},
    concurrency: 1,
    priority: 1,
    proxy_id: null,
    auto_pause_on_expired: false
  } as any
}

describe('AccountTestModal', () => {
  const originalFetch = global.fetch

  beforeEach(() => {
    getAvailableModelsMock.mockReset()
    refreshModelsMock.mockReset()
    getConnectedAccountModelsMock.mockReset()
    refreshConnectedAccountModelsMock.mockReset()
    getAvailableModelsMock.mockResolvedValue([
      { id: 'gpt-5.4', display_name: 'GPT-5.4' }
    ])
    refreshModelsMock.mockResolvedValue([
      { id: 'claude-sonnet-5', display_name: 'Claude Sonnet 5' }
    ])
    getConnectedAccountModelsMock.mockResolvedValue([
      { id: 'gpt-5.4', display_name: 'GPT-5.4' }
    ])
    refreshConnectedAccountModelsMock.mockResolvedValue([
      { id: 'claude-sonnet-5', display_name: 'Claude Sonnet 5' }
    ])
    global.fetch = vi.fn().mockImplementation(() => Promise.resolve(createStreamResponse()))
    localStorage.setItem('auth_token', 'test-token')
  })

  afterEach(() => {
    global.fetch = originalFetch
    localStorage.clear()
  })

  it('posts compact mode for OpenAI compact probe', async () => {
    const wrapper = mount(AccountTestModal, {
      props: {
        show: false,
        account: buildAccount()
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Select: SelectStub,
          TextArea: TextAreaStub,
          Icon: true
        }
      }
    })

    await wrapper.setProps({ show: true })
    await flushPromises()
    ;(wrapper.vm as any).selectedModelId = 'gpt-5.4'
    ;(wrapper.vm as any).testMode = 'compact'
    await (wrapper.vm as any).startTest()
    await flushPromises()

    expect(global.fetch).toHaveBeenCalledTimes(1)
    const [, options] = (global.fetch as any).mock.calls[0]
    expect(JSON.parse(options.body)).toMatchObject({
      model_id: 'gpt-5.4',
      mode: 'compact'
    })
  })

  it('uses user connected account endpoints for user-scoped probes', async () => {
    const wrapper = mount(AccountTestModal, {
      props: {
        show: false,
        account: buildAccount(),
        scope: 'user'
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Select: SelectStub,
          TextArea: TextAreaStub,
          Icon: true
        }
      }
    })

    await wrapper.setProps({ show: true })
    await flushPromises()
    ;(wrapper.vm as any).selectedModelId = 'gpt-5.4'
    await (wrapper.vm as any).startTest()
    await flushPromises()

    expect(getConnectedAccountModelsMock).toHaveBeenCalledWith(1)
    expect(getAvailableModelsMock).not.toHaveBeenCalled()
    const [url] = (global.fetch as any).mock.calls[0]
    expect(url).toContain('/user/accounts/1/test')
  })

  it('defaults to testing all models and probes each available model', async () => {
    getAvailableModelsMock.mockResolvedValueOnce([
      { id: 'gpt-5.4', display_name: 'GPT-5.4' },
      { id: 'gpt-5.4-mini', display_name: 'GPT-5.4 Mini' }
    ])
    ;(global.fetch as any).mockImplementation(() => Promise.resolve(createStreamResponse([
      'data: {"type":"test_complete","success":true}\n'
    ])))

    const wrapper = mount(AccountTestModal, {
      props: {
        show: false,
        account: buildAccount()
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Select: SelectStub,
          TextArea: TextAreaStub,
          Icon: true
        }
      }
    })

    await wrapper.setProps({ show: true })
    await flushPromises()

    expect((wrapper.vm as any).selectedModelId).toBe('__all_models__')
    expect((wrapper.vm as any).availableModels[0]).toMatchObject({
      id: '__all_models__',
      display_name: 'admin.accounts.testAllModelsConnection'
    })

    await (wrapper.vm as any).startTest()
    await flushPromises()

    expect(global.fetch).toHaveBeenCalledTimes(2)
    const bodies = (global.fetch as any).mock.calls.map(([, options]: any[]) => JSON.parse(options.body))
    expect(bodies.map((body: any) => body.model_id)).toEqual(['gpt-5.4', 'gpt-5.4-mini'])
  })

  it('refreshes models through the user endpoint in user scope', async () => {
    const account = {
      ...buildAccount(),
      name: 'Claude Main',
      platform: 'anthropic',
      type: 'oauth'
    }
    getConnectedAccountModelsMock.mockResolvedValueOnce([
      { id: 'claude-opus-4-8', display_name: 'Claude Opus 4.8' }
    ])
    refreshConnectedAccountModelsMock.mockResolvedValueOnce([
      { id: 'claude-sonnet-5', display_name: 'Claude Sonnet 5' }
    ])
    const wrapper = mount(AccountTestModal, {
      props: {
        show: true,
        account,
        scope: 'user'
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Select: SelectStub,
          TextArea: TextAreaStub,
          Icon: true
        }
      }
    })

    await flushPromises()
    await wrapper.get('button[aria-label="admin.accounts.refreshModels"]').trigger('click')
    await flushPromises()

    expect(refreshConnectedAccountModelsMock).toHaveBeenCalledWith(1)
    expect(refreshModelsMock).not.toHaveBeenCalled()
    expect((wrapper.vm as any).availableModels).toEqual([
      expect.objectContaining({
        id: '__all_models__',
        display_name: 'admin.accounts.testAllModelsConnection'
      }),
      { id: 'claude-sonnet-5', display_name: 'Claude Sonnet 5' }
    ])
    expect((wrapper.vm as any).selectedModelId).toBe('__all_models__')
  })
})
