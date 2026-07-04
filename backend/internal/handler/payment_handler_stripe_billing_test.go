package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestPaymentHandlerCreateStripeSubscriptionCheckout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stripeBilling := &paymentStripeBillingServiceStub{
		checkoutOut: &service.CreateStripeSubscriptionCheckoutOutput{
			SessionID:  "cs_test_123",
			URL:        "https://checkout.stripe.test/session",
			CustomerID: "cus_test_123",
		},
	}
	h := NewPaymentHandler(nil, nil, nil, stripeBilling)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	body := bytes.NewBufferString(`{
		"plan_id": 42,
		"success_url": "https://app.tokengate.to/subscriptions?checkout=success",
		"cancel_url": "https://app.tokengate.to/subscriptions?checkout=cancel"
	}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/payment/subscriptions/checkout", body)
	c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: 101})

	h.CreateStripeSubscriptionCheckout(c)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, int64(101), stripeBilling.checkoutIn.UserID)
	require.Equal(t, int64(42), stripeBilling.checkoutIn.PlanID)
	require.Equal(t, "https://app.tokengate.to/subscriptions?checkout=success", stripeBilling.checkoutIn.SuccessURL)
	require.Equal(t, "https://app.tokengate.to/subscriptions?checkout=cancel", stripeBilling.checkoutIn.CancelURL)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	data := payload["data"].(map[string]any)
	require.Equal(t, "cs_test_123", data["session_id"])
	require.Equal(t, "https://checkout.stripe.test/session", data["url"])
	require.Equal(t, "cus_test_123", data["customer_id"])
}

type paymentStripeBillingServiceStub struct {
	checkoutIn  service.CreateStripeSubscriptionCheckoutInput
	checkoutOut *service.CreateStripeSubscriptionCheckoutOutput
	checkoutErr error
	portalIn    service.CreateStripeBillingPortalInput
	portalOut   *service.CreateStripeBillingPortalOutput
	portalErr   error
}

func (s *paymentStripeBillingServiceStub) CreateSubscriptionCheckout(_ context.Context, in service.CreateStripeSubscriptionCheckoutInput) (*service.CreateStripeSubscriptionCheckoutOutput, error) {
	s.checkoutIn = in
	return s.checkoutOut, s.checkoutErr
}

func (s *paymentStripeBillingServiceStub) CreateBillingPortal(_ context.Context, in service.CreateStripeBillingPortalInput) (*service.CreateStripeBillingPortalOutput, error) {
	s.portalIn = in
	return s.portalOut, s.portalErr
}

func (s *paymentStripeBillingServiceStub) HandleNotification(context.Context, *payment.PaymentNotification) (bool, error) {
	return false, nil
}
