package provider

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v85/webhook"
)

const stripeTestWebhookSecret = "whsec_test_tokengate"

func TestStripeImplementsSubscriptionBillingProvider(t *testing.T) {
	var _ payment.SubscriptionBillingProvider = (*Stripe)(nil)
}

func TestStripeVerifyNotificationParsesCheckoutSessionCompleted(t *testing.T) {
	p := newTestStripeProvider(t)

	raw := `{
		"id":"evt_checkout_completed",
		"object":"event",
		"api_version":"2026-03-25.dahlia",
		"type":"checkout.session.completed",
		"data":{
			"object":{
				"id":"cs_test_123",
				"object":"checkout.session",
				"customer":"cus_123",
				"subscription":"sub_123",
				"metadata":{
					"tokengate_user_id":"101",
					"tokengate_plan_id":"202",
					"tokengate_group_id":"303"
				}
			}
		}
	}`

	n := verifySignedStripeNotification(t, p, raw)
	require.NotNil(t, n)
	require.Equal(t, payment.ProviderStatusSuccess, n.Status)
	require.Equal(t, "cs_test_123", n.TradeNo)
	require.Equal(t, "checkout.session.completed", n.Metadata["stripe_event_type"])
	require.Equal(t, "evt_checkout_completed", n.Metadata["stripe_event_id"])
	require.Equal(t, "cs_test_123", n.Metadata["stripe_session_id"])
	require.Equal(t, "cus_123", n.Metadata["stripe_customer_id"])
	require.Equal(t, "sub_123", n.Metadata["stripe_subscription_id"])
	require.Equal(t, "101", n.Metadata["tokengate_user_id"])
	require.Equal(t, "202", n.Metadata["tokengate_plan_id"])
	require.Equal(t, "303", n.Metadata["tokengate_group_id"])
}

func TestStripeVerifyNotificationParsesSubscriptionLifecycleEvents(t *testing.T) {
	p := newTestStripeProvider(t)

	raw := `{
		"id":"evt_subscription_updated",
		"object":"event",
		"api_version":"2026-03-25.dahlia",
		"type":"customer.subscription.updated",
		"data":{
			"object":{
				"id":"sub_123",
				"object":"subscription",
				"customer":"cus_123",
				"status":"trialing",
				"current_period_start":1800000000,
				"current_period_end":1802592000,
				"trial_start":1800000000,
				"trial_end":1800604800,
				"cancel_at_period_end":true,
				"items":{
					"data":[{
						"price":{"id":"price_byo_monthly"}
					}]
				},
				"metadata":{
					"tokengate_user_id":"101",
					"tokengate_plan_id":"202",
					"tokengate_group_id":"303"
				}
			}
		}
	}`

	n := verifySignedStripeNotification(t, p, raw)
	require.NotNil(t, n)
	require.Equal(t, payment.ProviderStatusSuccess, n.Status)
	require.Equal(t, "sub_123", n.TradeNo)
	require.Equal(t, "customer.subscription.updated", n.Metadata["stripe_event_type"])
	require.Equal(t, "evt_subscription_updated", n.Metadata["stripe_event_id"])
	require.Equal(t, "sub_123", n.Metadata["stripe_subscription_id"])
	require.Equal(t, "cus_123", n.Metadata["stripe_customer_id"])
	require.Equal(t, "trialing", n.Metadata["stripe_status"])
	require.Equal(t, "price_byo_monthly", n.Metadata["stripe_price_id"])
	require.Equal(t, "1800000000", n.Metadata["current_period_start"])
	require.Equal(t, "1802592000", n.Metadata["current_period_end"])
	require.Equal(t, "1800000000", n.Metadata["trial_start"])
	require.Equal(t, "1800604800", n.Metadata["trial_end"])
	require.Equal(t, "true", n.Metadata["cancel_at_period_end"])
	require.Equal(t, "101", n.Metadata["tokengate_user_id"])
}

func TestStripeVerifyNotificationParsesInvoicePaymentFailed(t *testing.T) {
	p := newTestStripeProvider(t)

	raw := `{
		"id":"evt_invoice_failed",
		"object":"event",
		"api_version":"2026-03-25.dahlia",
		"type":"invoice.payment_failed",
		"data":{
			"object":{
				"id":"in_123",
				"object":"invoice",
				"customer":"cus_123",
				"subscription":"sub_123",
				"status":"open",
				"metadata":{
					"tokengate_user_id":"101",
					"tokengate_plan_id":"202"
				}
			}
		}
	}`

	n := verifySignedStripeNotification(t, p, raw)
	require.NotNil(t, n)
	require.Equal(t, payment.ProviderStatusFailed, n.Status)
	require.Equal(t, "in_123", n.TradeNo)
	require.Equal(t, "invoice.payment_failed", n.Metadata["stripe_event_type"])
	require.Equal(t, "evt_invoice_failed", n.Metadata["stripe_event_id"])
	require.Equal(t, "in_123", n.Metadata["stripe_invoice_id"])
	require.Equal(t, "cus_123", n.Metadata["stripe_customer_id"])
	require.Equal(t, "sub_123", n.Metadata["stripe_subscription_id"])
	require.Equal(t, "open", n.Metadata["stripe_status"])
	require.Equal(t, "101", n.Metadata["tokengate_user_id"])
}

func newTestStripeProvider(t *testing.T) *Stripe {
	t.Helper()

	p, err := NewStripe("stripe-test", map[string]string{
		"secretKey":      "sk_test_tokengate",
		"publishableKey": "pk_test_tokengate",
		"webhookSecret":  stripeTestWebhookSecret,
		"currency":       "USD",
	})
	require.NoError(t, err)
	return p
}

func verifySignedStripeNotification(t *testing.T, p *Stripe, raw string) *payment.PaymentNotification {
	t.Helper()

	signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload:   []byte(raw),
		Secret:    stripeTestWebhookSecret,
		Timestamp: time.Now(),
	})
	n, err := p.VerifyNotification(context.Background(), string(signed.Payload), map[string]string{
		"stripe-signature": signed.Header,
	})
	require.NoError(t, err)
	return n
}
