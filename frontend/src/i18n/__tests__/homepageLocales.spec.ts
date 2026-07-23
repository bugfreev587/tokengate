import { describe, expect, it } from 'vitest'

import en from '@/i18n/locales/en'
import zh from '@/i18n/locales/zh'

const enHome = en.home as Record<string, any>
const zhHome = zh.home as Record<string, any>

describe('homepage locale contract', () => {
  it('presents Usage-based and BYO as equal English capacity modes', () => {
    expect(enHome.hero?.eyebrow).toBe('One gateway. Two ways to run.')
    expect(enHome.modes?.usageBased?.label).toBe('Usage-based')
    expect(enHome.modes?.byo?.label).toBe('BYO')
    expect(enHome.modes?.byo?.points?.balance).toContain('No TokenGate prepaid balance deduction')
    expect(enHome.closing?.button).toBe('Compare modes on Pricing')
  })

  it('uses natural Chinese terminology for both modes', () => {
    expect(zhHome.modes?.usageBased?.label).toBe('按量付费')
    expect(zhHome.modes?.byo?.label).toBe('BYO')
    expect(zhHome.modes?.byo?.description).toContain('自己的 AI 服务账号')
    expect(zhHome.modes?.byo?.points?.balance).toContain('不会扣除 TokenGate 预付余额')
  })
})
