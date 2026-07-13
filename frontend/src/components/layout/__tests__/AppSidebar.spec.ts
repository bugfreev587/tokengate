import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../AppSidebar.vue')
const componentSource = readFileSync(componentPath, 'utf8')
const stylePath = resolve(dirname(fileURLToPath(import.meta.url)), '../../../style.css')
const styleSource = readFileSync(stylePath, 'utf8')

describe('AppSidebar custom SVG styles', () => {
  it('does not override uploaded SVG fill or stroke colors', () => {
    expect(componentSource).toContain('.sidebar-svg-icon {')
    expect(componentSource).toContain('color: currentColor;')
    expect(componentSource).toContain('display: block;')
    expect(componentSource).not.toContain('stroke: currentColor;')
    expect(componentSource).not.toContain('fill: none;')
  })
})

describe('AppSidebar header styles', () => {
  it('does not clip the version badge dropdown', () => {
    const sidebarHeaderBlockMatch = styleSource.match(/\.sidebar-header\s*\{[\s\S]*?\n {2}\}/)
    const sidebarBrandBlockMatch = componentSource.match(/\.sidebar-brand\s*\{[\s\S]*?\n\}/)

    expect(sidebarHeaderBlockMatch).not.toBeNull()
    expect(sidebarBrandBlockMatch).not.toBeNull()
    expect(sidebarHeaderBlockMatch?.[0]).not.toContain('@apply overflow-hidden;')
    expect(sidebarBrandBlockMatch?.[0]).not.toContain('overflow: hidden;')
  })
})

describe('AppSidebar user account navigation', () => {
  it('includes user-owned AI accounts under the shared My Account menu', () => {
    expect(componentSource).toContain("{ path: '/accounts', label: t('nav.accounts'), icon: GlobeIcon }")
  })
})

describe('AppSidebar Billing navigation', () => {
  it('defines Billing as an expand-only group with query-backed mode children', () => {
    expect(componentSource).toContain("path: '/purchase',")
    expect(componentSource).toContain("label: t('nav.buySubscription')")
    expect(componentSource).toContain('expandOnly: true')
    expect(componentSource).toContain("label: t('nav.usageBasedMode')")
    expect(componentSource).toContain("query: { tab: 'recharge' }")
    expect(componentSource).toContain("label: t('nav.byoSubMode')")
    expect(componentSource).toContain("query: { tab: 'subscription' }")
  })

  it('renders collapsible groups in regular-user and personal navigation', () => {
    expect(componentSource.match(/v-if="item\.children\?\.length"/g)?.length).toBe(3)
    expect(componentSource).toContain(':to="navTarget(child)"')
    expect(componentSource).toContain("'sidebar-link-active': isNavItemActive(child)")
  })
})
