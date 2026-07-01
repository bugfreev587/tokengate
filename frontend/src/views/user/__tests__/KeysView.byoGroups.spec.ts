import { describe, expect, it } from 'vitest'

import source from '../KeysView.vue?raw'

describe('KeysView BYO group selector wiring', () => {
  it('passes capacity source metadata into selected badges and dropdown options', () => {
    expect(source).toContain('capacitySource: group.capacity_source')
    expect(source).toContain(':capacity-source="(option as unknown as GroupOption).capacitySource"')
    expect(source).toContain(':capacity-source="row.group.capacity_source"')
  })
})
