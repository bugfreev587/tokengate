package service

import (
	"math"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/shopspring/decimal"
)

const defaultBalanceRechargeMultiplier = 1.0
const defaultUSDCNYRate = 7.2

func normalizeBalanceRechargeMultiplier(multiplier float64) float64 {
	if math.IsNaN(multiplier) || math.IsInf(multiplier, 0) || multiplier <= 0 {
		return defaultBalanceRechargeMultiplier
	}
	return multiplier
}

func normalizeUSDCNYRate(rate float64) float64 {
	if math.IsNaN(rate) || math.IsInf(rate, 0) || rate <= 0 {
		return defaultUSDCNYRate
	}
	return rate
}

func calculatePaymentBaseAmount(pricingAmount float64, paymentCurrency string, usdCNYRate float64) float64 {
	currency, err := payment.NormalizePaymentCurrency(paymentCurrency)
	if err != nil {
		currency = payment.DefaultPaymentCurrency
	}
	amount := decimal.NewFromFloat(pricingAmount)
	if currency == "CNY" {
		return amount.
			Mul(decimal.NewFromFloat(normalizeUSDCNYRate(usdCNYRate))).
			Round(int32(payment.CurrencyMaxFractionDigits(currency))).
			InexactFloat64()
	}
	return amount.InexactFloat64()
}

func calculateGatewayRefundAmount(orderAmount, payAmount, refundAmount float64, currency string) float64 {
	if orderAmount <= 0 || payAmount <= 0 || refundAmount <= 0 {
		return 0
	}
	fractionDigits := int32(payment.CurrencyMaxFractionDigits(currency))
	if math.Abs(refundAmount-orderAmount) <= paymentAmountToleranceForCurrency(currency) {
		return decimal.NewFromFloat(payAmount).Round(fractionDigits).InexactFloat64()
	}
	return decimal.NewFromFloat(payAmount).
		Mul(decimal.NewFromFloat(refundAmount)).
		Div(decimal.NewFromFloat(orderAmount)).
		Round(fractionDigits).
		InexactFloat64()
}
