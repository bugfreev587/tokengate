import { describe, expect, it } from 'vitest'
import {
  DEFAULT_PAYMENT_CURRENCY,
  PRICING_CURRENCY,
  convertPricingAmountToPaymentCurrency,
  formatPaymentAmount,
} from '../currency'

describe('formatPaymentAmount', () => {
  it('uses the currency default fraction digits', () => {
    expect(formatPaymentAmount(100, 'JPY', 'en-US')).not.toContain('.00')
    expect(formatPaymentAmount(100, 'KRW', 'en-US')).not.toContain('.00')
    expect(formatPaymentAmount(100, 'HKD', 'en-US')).toContain('.00')
  })
})

describe('payment currency conversion', () => {
  it('defaults display pricing to USD', () => {
    expect(DEFAULT_PAYMENT_CURRENCY).toBe('USD')
    expect(PRICING_CURRENCY).toBe('USD')
  })

  it('converts USD pricing amounts to CNY payment amounts with the configured rate', () => {
    expect(convertPricingAmountToPaymentCurrency(10, 'CNY', 7.2)).toBe(72)
    expect(convertPricingAmountToPaymentCurrency(10.25, 'CNY', 7.234)).toBe(74.15)
    expect(convertPricingAmountToPaymentCurrency(10, 'USD', 7.2)).toBe(10)
  })
})
