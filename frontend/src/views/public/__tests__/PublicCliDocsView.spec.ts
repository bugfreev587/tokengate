import { mount, RouterLinkStub } from '@vue/test-utils'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it, vi } from 'vitest'

import PublicCliDocsView from '../PublicCliDocsView.vue'

const viewSource = readFileSync(
  resolve(process.cwd(), 'src/views/public/PublicCliDocsView.vue'),
  'utf8',
)

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    cachedPublicSettings: null,
    siteName: 'TokenGate',
    siteLogo: '',
    fetchPublicSettings: vi.fn().mockResolvedValue(undefined),
  }),
}))

describe('PublicCliDocsView', () => {
  it('shows TokenGate setup guidance for Claude Code CLI and Codex CLI', () => {
    const wrapper = mount(PublicCliDocsView, {
      global: {
        stubs: {
          RouterLink: RouterLinkStub,
        },
      },
    })

    const text = wrapper.text()

    expect(text).toContain('Use Claude Code CLI and Codex CLI with TokenGate')
    expect(text).toContain('Claude Code CLI')
    expect(text).toContain('Codex CLI')
    expect(text).toContain('ANTHROPIC_BASE_URL')
    expect(text).toContain('ANTHROPIC_AUTH_TOKEN')
    expect(text).toContain('TOKENGATE_API_KEY')
    expect(text).toContain('https://api.tokengate.to')
    expect(text).toContain('https://api.tokengate.to/v1')
    expect(text).toContain('wire_api = "responses"')
  })

  it('keeps docs code text readable when selected', () => {
    expect(viewSource).toContain('.tg-code-line::selection')
    expect(viewSource).toContain('.tg-inline-code::selection')
    expect(viewSource).toContain('color: #111827;')
  })
})
