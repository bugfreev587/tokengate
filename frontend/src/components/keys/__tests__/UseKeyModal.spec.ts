import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn().mockResolvedValue(true)
  })
}))

import UseKeyModal from '../UseKeyModal.vue'

describe('UseKeyModal', () => {
  const mountOpenAI = () => mount(UseKeyModal, {
    props: {
      show: true,
      apiKey: 'sk-test',
      baseUrl: 'https://example.com/v1',
      platform: 'openai'
    },
    global: {
      stubs: {
        BaseDialog: {
          template: '<div><slot /><slot name="footer" /></div>'
        },
        Icon: {
          template: '<span />'
        }
      }
    }
  })

  it('uses the current broadly available Codex model in config.toml', () => {
    const wrapper = mountOpenAI()
    const codeBlocks = wrapper.findAll('pre code').map((code) => code.text()).join('\n')

    expect(codeBlocks).toContain('model = "gpt-5.6-terra"')
    expect(codeBlocks).toContain('review_model = "gpt-5.6-terra"')
    expect(codeBlocks).not.toContain('gpt-5.4')
  })

  it('renders current Codex models in OpenCode config', async () => {
    const wrapper = mountOpenAI()

    const opencodeTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.opencode')
    )

    expect(opencodeTab).toBeDefined()
    await opencodeTab!.trigger('click')
    await nextTick()

    const codeBlock = wrapper.find('pre code')
    expect(codeBlock.exists()).toBe(true)
    expect(codeBlock.text()).toContain('"name": "GPT-6 Astra"')
    expect(codeBlock.text()).toContain('"name": "GPT-5.6 Sol"')
    expect(codeBlock.text()).toContain('"name": "GPT-5.6 Terra"')
    expect(codeBlock.text()).toContain('"name": "GPT-5.6 Luna"')
    expect(codeBlock.text()).not.toContain('GPT-5.4')
    expect(codeBlock.text()).not.toContain('GPT-5.3 Codex')
  })
})
