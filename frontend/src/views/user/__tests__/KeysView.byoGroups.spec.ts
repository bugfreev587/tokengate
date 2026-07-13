import { describe, expect, it } from 'vitest'

import source from '../KeysView.vue?raw'

describe('KeysView BYO group selector wiring', () => {
  it('passes capacity source metadata into selected badges and dropdown options', () => {
    expect(source).toContain('capacitySource: getGroupCapacitySource(group)')
    expect(source).toContain('byoEnabled: getGroupBYOEnabled(group)')
    expect(source).toContain('byoDisabledReason: getGroupBYODisabledReason(group)')
    expect(source).toContain(':capacity-source="(option as unknown as GroupOption).capacitySource"')
    expect(source).toContain(':capacity-source="row.group.capacity_source"')
    expect(source).toContain(':byo-enabled="getEffectiveGroupBYOEnabled(row.group)"')
    expect(source).toContain('selectedFormGroupBYODisabled')
  })

  it('warns after saving or switching a key to subscription-inactive BYO capacity', () => {
    expect(source).toContain('const warnIfBYOSubscriptionRequired')
    expect(source).toContain("appStore.showWarning(t('keys.byoSubscriptionRequiredAfterBind'))")
    expect(source.match(/warnIfBYOSubscriptionRequired\(/g)).toHaveLength(2)
  })
})
