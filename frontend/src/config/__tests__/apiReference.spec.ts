import { describe, expect, it } from 'vitest'

import {
  tokenGateApiEndpoint,
  tokenGateApiEndpoints,
  tokenGateApiSidebarGroups,
} from '../apiReference'

describe('tokenGateApiSidebarGroups', () => {
  it('links to the CLI setup guide from the public docs sidebar', () => {
    const guideItems = tokenGateApiSidebarGroups
      .find((group) => group.title === 'Guides')
      ?.items ?? []

    expect(guideItems).toContainEqual({
      title: 'CLI setup',
      href: '/docs/cli',
    })
    expect(guideItems).toContainEqual({
      title: 'Claude Code statusline',
      href: '/docs/cli/statusline',
    })
  })

  it('uses current model ids in user API key examples', () => {
    expect(JSON.stringify(tokenGateApiEndpoint)).toContain('gpt-5.6-terra')
    expect(JSON.stringify(tokenGateApiEndpoint)).not.toContain('gpt-5.4')

    const anthropicMessages = tokenGateApiEndpoints.find((endpoint) => endpoint.id === 'anthropic-messages')
    expect(anthropicMessages).toBeDefined()
    expect(JSON.stringify(anthropicMessages)).toContain('claude-sonnet-5')
    expect(JSON.stringify(anthropicMessages)).not.toContain('claude-sonnet-4.6')
  })
})
