export const PRICING_CURRENCY = 'USD'
export const DEFAULT_PAYMENT_CURRENCY = PRICING_CURRENCY
export const DEFAULT_USD_CNY_RATE = 7.2

export function normalizePaymentCurrency(currency?: string | null): string {
  const normalized = String(currency || '').trim().toUpperCase()
  return /^[A-Z]{3}$/.test(normalized) ? normalized : DEFAULT_PAYMENT_CURRENCY
}

function paymentCurrencyFractionDigits(currency: string): number {
  try {
    return new Intl.NumberFormat(undefined, {
      style: 'currency',
      currency,
    }).resolvedOptions().maximumFractionDigits ?? 2
  } catch {
    return 2
  }
}

export function formatPaymentAmount(amount: number, currency?: string | null, locale?: string): string {
  const normalized = normalizePaymentCurrency(currency)
  const fractionDigits = paymentCurrencyFractionDigits(normalized)
  try {
    return new Intl.NumberFormat(locale || undefined, {
      style: 'currency',
      currency: normalized,
      currencyDisplay: 'narrowSymbol',
      minimumFractionDigits: fractionDigits,
      maximumFractionDigits: fractionDigits,
    }).format(Number.isFinite(amount) ? amount : 0)
  } catch {
    return `${normalized} ${(Number.isFinite(amount) ? amount : 0).toFixed(fractionDigits)}`
  }
}

export function convertPricingAmountToPaymentCurrency(
  amount: number,
  currency?: string | null,
  usdCnyRate = DEFAULT_USD_CNY_RATE,
): number {
  const normalized = normalizePaymentCurrency(currency)
  const safeAmount = Number.isFinite(amount) ? amount : 0
  const safeRate = Number.isFinite(usdCnyRate) && usdCnyRate > 0 ? usdCnyRate : DEFAULT_USD_CNY_RATE
  if (normalized !== 'CNY') {
    return safeAmount
  }
  return Math.round((safeAmount * safeRate + Number.EPSILON) * 100) / 100
}
