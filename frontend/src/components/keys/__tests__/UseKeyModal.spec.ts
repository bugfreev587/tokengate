import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'

const { getClaudeCodeConnect } = vi.hoisted(() => ({
  getClaudeCodeConnect: vi.fn()
}))

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

vi.mock('@/api', () => ({
  keysAPI: {
    getClaudeCodeConnect
  }
}))

import UseKeyModal from '../UseKeyModal.vue'

describe('UseKeyModal', () => {
  beforeEach(() => {
    getClaudeCodeConnect.mockReset()
  })

  it('renders GPT-5.4 mini entry in OpenCode config', async () => {
    const wrapper = mount(UseKeyModal, {
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

    const opencodeTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.opencode')
    )

    expect(opencodeTab).toBeDefined()
    await opencodeTab!.trigger('click')
    await nextTick()

    const codeBlock = wrapper.find('pre code')
    expect(codeBlock.exists()).toBe(true)
    expect(codeBlock.text()).toContain('"name": "GPT-5.4 Mini"')
    expect(codeBlock.text()).not.toContain('"name": "GPT-5.4 Nano"')
  })

  it('keeps Claude Code visible but disabled for OpenAI keys without messages dispatch', () => {
    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKeyId: 7,
        apiKey: 'sk-test',
        baseUrl: 'https://example.com/v1',
        platform: 'openai',
        allowMessagesDispatch: false
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

    const claudeTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.claudeCode')
    )

    expect(claudeTab).toBeDefined()
    expect(claudeTab!.attributes('disabled')).toBeDefined()
    expect(getClaudeCodeConnect).not.toHaveBeenCalled()
  })

  it('renders Claude Code settings from the connect payload without local discovery blockers', async () => {
    getClaudeCodeConnect.mockResolvedValue({
      supported: true,
      base_url: 'https://api.tokengate.to',
      settings: {
        env: {
          ANTHROPIC_BASE_URL: 'https://api.tokengate.to',
          ANTHROPIC_AUTH_TOKEN: 'sk-live',
          CLAUDE_CODE_ENABLE_GATEWAY_MODEL_DISCOVERY: '1',
          ANTHROPIC_DEFAULT_FABLE_MODEL: 'claude-fable-5'
        }
      },
      models: {
        fable: 'claude-fable-5',
        available: ['claude-opus-4-8', 'claude-fable-5']
      }
    })

    const wrapper = mount(UseKeyModal, {
      props: {
        show: true,
        apiKeyId: 7,
        apiKey: 'sk-live',
        baseUrl: 'https://api.tokengate.to',
        platform: 'anthropic'
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

    await nextTick()
    await Promise.resolve()
    await nextTick()

    expect(getClaudeCodeConnect).toHaveBeenCalledWith(7)
    const codeBlocks = wrapper.findAll('pre code').map((node) => node.text())
    const combined = codeBlocks.join('\n')
    expect(combined).toContain('"CLAUDE_CODE_ENABLE_GATEWAY_MODEL_DISCOVERY": "1"')
    expect(combined).toContain('"ANTHROPIC_DEFAULT_FABLE_MODEL": "claude-fable-5"')
    expect(combined).not.toContain('CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC')
    expect(combined).not.toContain('ANTHROPIC_CUSTOM_MODEL_OPTION')
  })
})
