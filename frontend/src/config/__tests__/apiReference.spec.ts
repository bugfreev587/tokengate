import { describe, expect, it } from 'vitest'

import { tokenGateApiSidebarGroups } from '../apiReference'

describe('tokenGateApiSidebarGroups', () => {
  it('links to the CLI setup guide from the public docs sidebar', () => {
    const guideItems = tokenGateApiSidebarGroups
      .find((group) => group.title === 'Guides')
      ?.items ?? []

    expect(guideItems).toContainEqual({
      title: 'CLI setup',
      href: '/docs/cli',
    })
  })
})
