import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const routerPath = resolve(dirname(fileURLToPath(import.meta.url)), '../index.ts')
const routerSource = readFileSync(routerPath, 'utf8')

describe('user accounts route', () => {
  it('registers the My Account accounts page as an authenticated user route', () => {
    expect(routerSource).toContain("path: '/accounts'")
    expect(routerSource).toContain("name: 'UserAccounts'")
    expect(routerSource).toContain("component: () => import('@/views/user/AccountsView.vue')")
    expect(routerSource).toContain("titleKey: 'userAccounts.title'")
  })
})
