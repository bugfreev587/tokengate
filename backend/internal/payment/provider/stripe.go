package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	stripe "github.com/stripe/stripe-go/v85"
	"github.com/stripe/stripe-go/v85/webhook"
)

// Stripe constants.
const (
	stripeEventPaymentSuccess              = "payment_intent.succeeded"
	stripeEventPaymentFailed               = "payment_intent.payment_failed"
	stripeEventCheckoutSessionCompleted    = "checkout.session.completed"
	stripeEventCustomerSubscriptionCreated = "customer.subscription.created"
	stripeEventCustomerSubscriptionUpdated = "customer.subscription.updated"
	stripeEventCustomerSubscriptionDeleted = "customer.subscription.deleted"
	stripeEventInvoicePaymentSucceeded     = "invoice.payment_succeeded"
	stripeEventInvoicePaymentFailed        = "invoice.payment_failed"
)

// Stripe implements the payment.CancelableProvider interface for Stripe payments.
type Stripe struct {
	instanceID string
	config     map[string]string

	mu          sync.Mutex
	initialized bool
	sc          *stripe.Client
}

// NewStripe creates a new Stripe provider instance.
func NewStripe(instanceID string, config map[string]string) (*Stripe, error) {
	if config["secretKey"] == "" {
		return nil, fmt.Errorf("stripe config missing required key: secretKey")
	}
	cfg := cloneStringMap(config)
	cfg["currency"] = payment.StripePaymentCurrency
	return &Stripe{
		instanceID: instanceID,
		config:     cfg,
	}, nil
}

func (s *Stripe) ensureInit() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.initialized {
		s.sc = stripe.NewClient(s.config["secretKey"])
		s.initialized = true
	}
}

// GetPublishableKey returns the publishable key for frontend use.
func (s *Stripe) GetPublishableKey() string {
	return s.config["publishableKey"]
}

func (s *Stripe) Name() string        { return "Stripe" }
func (s *Stripe) ProviderKey() string { return payment.TypeStripe }
func (s *Stripe) SupportedTypes() []payment.PaymentType {
	return []payment.PaymentType{payment.TypeStripe}
}

func (s *Stripe) MerchantIdentityMetadata() map[string]string {
	if s == nil {
		return nil
	}
	return map[string]string{"currency": s.currency()}
}

func (s *Stripe) currency() string {
	return payment.StripePaymentCurrency
}

// stripePaymentMethodTypes maps our PaymentType to Stripe payment_method_types.
var stripePaymentMethodTypes = map[string][]string{
	payment.TypeCard:   {"card"},
	payment.TypeAlipay: {"alipay"},
	payment.TypeWxpay:  {"wechat_pay"},
	payment.TypeLink:   {"link"},
}

// CreatePayment creates a Stripe PaymentIntent.
func (s *Stripe) CreatePayment(ctx context.Context, req payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	s.ensureInit()

	currency := s.currency()
	amountInMinorUnit, err := payment.AmountToMinorUnit(req.Amount, currency)
	if err != nil {
		return nil, fmt.Errorf("stripe create payment: %w", err)
	}

	// Collect all Stripe payment_method_types from the instance's configured sub-methods
	methods := resolveStripeMethodTypes(req.InstanceSubMethods)

	pmTypes := make([]*string, len(methods))
	for i, m := range methods {
		pmTypes[i] = stripe.String(m)
	}

	params := &stripe.PaymentIntentCreateParams{
		Amount:             stripe.Int64(amountInMinorUnit),
		Currency:           stripe.String(strings.ToLower(currency)),
		PaymentMethodTypes: pmTypes,
		Description:        stripe.String(req.Subject),
		Metadata:           map[string]string{"orderId": req.OrderID},
	}

	// WeChat Pay requires payment_method_options with client type
	if hasStripeMethod(methods, "wechat_pay") {
		params.PaymentMethodOptions = &stripe.PaymentIntentCreatePaymentMethodOptionsParams{
			WeChatPay: &stripe.PaymentIntentCreatePaymentMethodOptionsWeChatPayParams{
				Client: stripe.String("web"),
			},
		}
	}

	params.SetIdempotencyKey(fmt.Sprintf("pi-%s", req.OrderID))
	params.Context = ctx

	pi, err := s.sc.V1PaymentIntents.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("stripe create payment: %w", err)
	}

	return &payment.CreatePaymentResponse{
		TradeNo:      pi.ID,
		ClientSecret: pi.ClientSecret,
		Currency:     currency,
	}, nil
}

// CreateSubscriptionCheckout creates a hosted Stripe Checkout Session for an
// auto-renewing subscription.
func (s *Stripe) CreateSubscriptionCheckout(ctx context.Context, req payment.CreateSubscriptionCheckoutRequest) (*payment.CreateSubscriptionCheckoutResponse, error) {
	s.ensureInit()

	if strings.TrimSpace(req.PriceID) == "" {
		return nil, fmt.Errorf("stripe create subscription checkout: priceID is required")
	}
	if strings.TrimSpace(req.SuccessURL) == "" {
		return nil, fmt.Errorf("stripe create subscription checkout: successURL is required")
	}
	if strings.TrimSpace(req.CancelURL) == "" {
		return nil, fmt.Errorf("stripe create subscription checkout: cancelURL is required")
	}

	metadata := cloneStringMap(req.Metadata)
	params := &stripe.CheckoutSessionCreateParams{
		Mode:                    stripe.String(stripe.CheckoutSessionModeSubscription),
		SuccessURL:              stripe.String(req.SuccessURL),
		CancelURL:               stripe.String(req.CancelURL),
		PaymentMethodCollection: stripe.String(stripe.CheckoutSessionPaymentMethodCollectionAlways),
		LineItems: []*stripe.CheckoutSessionCreateLineItemParams{
			{
				Price:    stripe.String(req.PriceID),
				Quantity: stripe.Int64(1),
			},
		},
		Metadata: metadata,
		SubscriptionData: &stripe.CheckoutSessionCreateSubscriptionDataParams{
			Metadata: metadata,
		},
	}
	if strings.TrimSpace(req.CustomerID) != "" {
		params.Customer = stripe.String(req.CustomerID)
	} else if strings.TrimSpace(req.CustomerEmail) != "" {
		params.CustomerEmail = stripe.String(req.CustomerEmail)
	}
	if req.TrialDays > 0 {
		params.SubscriptionData.TrialPeriodDays = stripe.Int64(req.TrialDays)
	}
	if metadata["tokengate_user_id"] != "" && metadata["tokengate_plan_id"] != "" {
		params.SetIdempotencyKey(fmt.Sprintf("sub-checkout-%s-%s", metadata["tokengate_user_id"], metadata["tokengate_plan_id"]))
	}
	params.Context = ctx

	session, err := s.sc.V1CheckoutSessions.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("stripe create subscription checkout: %w", err)
	}

	customerID := ""
	if session.Customer != nil {
		customerID = session.Customer.ID
	}
	return &payment.CreateSubscriptionCheckoutResponse{
		SessionID:  session.ID,
		URL:        session.URL,
		CustomerID: customerID,
	}, nil
}

// CreateBillingPortal creates a hosted Stripe customer portal session.
func (s *Stripe) CreateBillingPortal(ctx context.Context, req payment.CreateBillingPortalRequest) (*payment.CreateBillingPortalResponse, error) {
	s.ensureInit()

	if strings.TrimSpace(req.CustomerID) == "" {
		return nil, fmt.Errorf("stripe create billing portal: customerID is required")
	}
	if strings.TrimSpace(req.ReturnURL) == "" {
		return nil, fmt.Errorf("stripe create billing portal: returnURL is required")
	}

	params := &stripe.BillingPortalSessionCreateParams{
		Customer:  stripe.String(req.CustomerID),
		ReturnURL: stripe.String(req.ReturnURL),
	}
	params.Context = ctx

	session, err := s.sc.V1BillingPortalSessions.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("stripe create billing portal: %w", err)
	}

	return &payment.CreateBillingPortalResponse{
		SessionID: session.ID,
		URL:       session.URL,
	}, nil
}

// QueryOrder retrieves a PaymentIntent by ID.
func (s *Stripe) QueryOrder(ctx context.Context, tradeNo string) (*payment.QueryOrderResponse, error) {
	s.ensureInit()

	pi, err := s.sc.V1PaymentIntents.Retrieve(ctx, tradeNo, nil)
	if err != nil {
		return nil, fmt.Errorf("stripe query order: %w", err)
	}

	status := payment.ProviderStatusPending
	switch pi.Status {
	case stripe.PaymentIntentStatusSucceeded:
		status = payment.ProviderStatusPaid
	case stripe.PaymentIntentStatusCanceled:
		status = payment.ProviderStatusFailed
	}

	currency := stripeIntentCurrency(pi.Currency, s.currency())
	return &payment.QueryOrderResponse{
		TradeNo: pi.ID,
		Status:  status,
		Amount:  payment.MinorUnitToAmount(pi.Amount, currency),
		Metadata: map[string]string{
			"currency": currency,
		},
	}, nil
}

// VerifyNotification verifies a Stripe webhook event.
func (s *Stripe) VerifyNotification(_ context.Context, rawBody string, headers map[string]string) (*payment.PaymentNotification, error) {
	s.ensureInit()

	webhookSecret := s.config["webhookSecret"]
	if webhookSecret == "" {
		return nil, fmt.Errorf("stripe webhookSecret not configured")
	}

	sig := headers["stripe-signature"]
	if sig == "" {
		return nil, fmt.Errorf("stripe notification missing stripe-signature header")
	}

	event, err := webhook.ConstructEvent([]byte(rawBody), sig, webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("stripe verify notification: %w", err)
	}

	switch event.Type {
	case stripeEventPaymentSuccess:
		return parseStripePaymentIntent(&event, payment.ProviderStatusSuccess, rawBody)
	case stripeEventPaymentFailed:
		return parseStripePaymentIntent(&event, payment.ProviderStatusFailed, rawBody)
	case stripeEventCheckoutSessionCompleted:
		return parseStripeCheckoutSession(&event, rawBody)
	case stripeEventCustomerSubscriptionCreated, stripeEventCustomerSubscriptionUpdated, stripeEventCustomerSubscriptionDeleted:
		return parseStripeSubscription(&event, rawBody)
	case stripeEventInvoicePaymentSucceeded:
		return parseStripeInvoice(&event, payment.ProviderStatusSuccess, rawBody)
	case stripeEventInvoicePaymentFailed:
		return parseStripeInvoice(&event, payment.ProviderStatusFailed, rawBody)
	}

	return nil, nil
}

func parseStripePaymentIntent(event *stripe.Event, status string, rawBody string) (*payment.PaymentNotification, error) {
	var pi stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
		return nil, fmt.Errorf("stripe parse payment_intent: %w", err)
	}
	currency := stripeIntentCurrency(pi.Currency, payment.StripePaymentCurrency)
	return &payment.PaymentNotification{
		TradeNo: pi.ID,
		OrderID: pi.Metadata["orderId"],
		Amount:  payment.MinorUnitToAmount(pi.Amount, currency),
		Status:  status,
		RawData: rawBody,
		Metadata: map[string]string{
			"currency": currency,
		},
	}, nil
}

func parseStripeCheckoutSession(event *stripe.Event, rawBody string) (*payment.PaymentNotification, error) {
	obj, err := stripeEventObject(event)
	if err != nil {
		return nil, fmt.Errorf("stripe parse checkout session: %w", err)
	}

	metadata := stripeBaseEventMetadata(event, obj)
	sessionID := stripeObjectString(obj, "id")
	metadata["stripe_session_id"] = sessionID
	metadata["stripe_customer_id"] = stripeObjectID(obj["customer"])
	metadata["stripe_subscription_id"] = stripeObjectID(obj["subscription"])

	return &payment.PaymentNotification{
		TradeNo:  sessionID,
		Status:   payment.ProviderStatusSuccess,
		RawData:  rawBody,
		Metadata: metadata,
	}, nil
}

func parseStripeSubscription(event *stripe.Event, rawBody string) (*payment.PaymentNotification, error) {
	obj, err := stripeEventObject(event)
	if err != nil {
		return nil, fmt.Errorf("stripe parse subscription: %w", err)
	}

	metadata := stripeBaseEventMetadata(event, obj)
	subscriptionID := stripeObjectString(obj, "id")
	metadata["stripe_subscription_id"] = subscriptionID
	metadata["stripe_customer_id"] = stripeObjectID(obj["customer"])
	metadata["stripe_status"] = stripeObjectString(obj, "status")
	metadata["stripe_price_id"] = stripeSubscriptionPriceID(obj)
	metadata["current_period_start"] = stripeObjectInt64String(obj, "current_period_start")
	metadata["current_period_end"] = stripeObjectInt64String(obj, "current_period_end")
	metadata["trial_start"] = stripeObjectInt64String(obj, "trial_start")
	metadata["trial_end"] = stripeObjectInt64String(obj, "trial_end")
	metadata["cancel_at_period_end"] = stripeObjectBoolString(obj, "cancel_at_period_end")

	return &payment.PaymentNotification{
		TradeNo:  subscriptionID,
		Status:   payment.ProviderStatusSuccess,
		RawData:  rawBody,
		Metadata: metadata,
	}, nil
}

func parseStripeInvoice(event *stripe.Event, status string, rawBody string) (*payment.PaymentNotification, error) {
	obj, err := stripeEventObject(event)
	if err != nil {
		return nil, fmt.Errorf("stripe parse invoice: %w", err)
	}

	metadata := stripeBaseEventMetadata(event, obj)
	invoiceID := stripeObjectString(obj, "id")
	metadata["stripe_invoice_id"] = invoiceID
	metadata["stripe_customer_id"] = stripeObjectID(obj["customer"])
	metadata["stripe_subscription_id"] = stripeObjectID(obj["subscription"])
	metadata["stripe_status"] = stripeObjectString(obj, "status")

	return &payment.PaymentNotification{
		TradeNo:  invoiceID,
		Status:   status,
		RawData:  rawBody,
		Metadata: metadata,
	}, nil
}

func stripeEventObject(event *stripe.Event) (map[string]any, error) {
	var obj map[string]any
	if err := json.Unmarshal(event.Data.Raw, &obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func stripeBaseEventMetadata(event *stripe.Event, obj map[string]any) map[string]string {
	metadata := map[string]string{
		"stripe_event_type": string(event.Type),
		"stripe_event_id":   event.ID,
		"stripe_object_id":  stripeObjectString(obj, "id"),
	}
	for k, v := range stripeMetadata(obj) {
		metadata[k] = v
	}
	return metadata
}

func stripeMetadata(obj map[string]any) map[string]string {
	raw, ok := obj["metadata"].(map[string]any)
	if !ok {
		return nil
	}

	metadata := make(map[string]string, len(raw))
	for k, v := range raw {
		if s, ok := v.(string); ok {
			metadata[k] = s
		}
	}
	return metadata
}

func stripeObjectString(obj map[string]any, key string) string {
	if obj == nil {
		return ""
	}
	if s, ok := obj[key].(string); ok {
		return s
	}
	return ""
}

func stripeObjectID(raw any) string {
	switch v := raw.(type) {
	case string:
		return v
	case map[string]any:
		if s, ok := v["id"].(string); ok {
			return s
		}
	}
	return ""
}

func stripeObjectInt64String(obj map[string]any, key string) string {
	if obj == nil {
		return ""
	}
	switch v := obj[key].(type) {
	case float64:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case json.Number:
		return v.String()
	case string:
		return v
	default:
		return ""
	}
}

func stripeObjectBoolString(obj map[string]any, key string) string {
	if obj == nil {
		return ""
	}
	if v, ok := obj[key].(bool); ok {
		return strconv.FormatBool(v)
	}
	return ""
}

func stripeSubscriptionPriceID(obj map[string]any) string {
	items, ok := obj["items"].(map[string]any)
	if !ok {
		return ""
	}
	data, ok := items["data"].([]any)
	if !ok || len(data) == 0 {
		return ""
	}
	first, ok := data[0].(map[string]any)
	if !ok {
		return ""
	}
	return stripeObjectID(first["price"])
}

// Refund creates a Stripe refund.
func (s *Stripe) Refund(ctx context.Context, req payment.RefundRequest) (*payment.RefundResponse, error) {
	s.ensureInit()

	amountInMinorUnit, err := payment.AmountToMinorUnit(req.Amount, s.currency())
	if err != nil {
		return nil, fmt.Errorf("stripe refund: %w", err)
	}

	params := &stripe.RefundCreateParams{
		PaymentIntent: stripe.String(req.TradeNo),
		Amount:        stripe.Int64(amountInMinorUnit),
		Reason:        stripe.String(string(stripe.RefundReasonRequestedByCustomer)),
	}
	params.Context = ctx

	r, err := s.sc.V1Refunds.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("stripe refund: %w", err)
	}

	refundStatus := payment.ProviderStatusPending
	if r.Status == stripe.RefundStatusSucceeded {
		refundStatus = payment.ProviderStatusSuccess
	}

	return &payment.RefundResponse{
		RefundID: r.ID,
		Status:   refundStatus,
	}, nil
}

func stripeIntentCurrency(raw stripe.Currency, fallback string) string {
	currency, err := payment.NormalizePaymentCurrency(string(raw))
	if err != nil || currency == payment.DefaultPaymentCurrency && strings.TrimSpace(string(raw)) == "" {
		normalizedFallback, fallbackErr := payment.NormalizePaymentCurrency(fallback)
		if fallbackErr == nil {
			return normalizedFallback
		}
		return payment.DefaultPaymentCurrency
	}
	return currency
}

// resolveStripeMethodTypes converts instance supported_types (comma-separated)
// into Stripe API payment_method_types. Falls back to ["card"] if empty.
func resolveStripeMethodTypes(instanceSubMethods string) []string {
	if instanceSubMethods == "" {
		return []string{"card"}
	}
	var methods []string
	for _, t := range strings.Split(instanceSubMethods, ",") {
		t = strings.TrimSpace(t)
		if mapped, ok := stripePaymentMethodTypes[t]; ok {
			methods = append(methods, mapped...)
		}
	}
	if len(methods) == 0 {
		return []string{"card"}
	}
	return methods
}

// hasStripeMethod checks if the given Stripe method list contains the target method.
func hasStripeMethod(methods []string, target string) bool {
	for _, m := range methods {
		if m == target {
			return true
		}
	}
	return false
}

// CancelPayment cancels a pending PaymentIntent.
func (s *Stripe) CancelPayment(ctx context.Context, tradeNo string) error {
	s.ensureInit()

	_, err := s.sc.V1PaymentIntents.Cancel(ctx, tradeNo, nil)
	if err != nil {
		return fmt.Errorf("stripe cancel payment: %w", err)
	}
	return nil
}

// Ensure interface compliance.
var (
	_ payment.Provider                    = (*Stripe)(nil)
	_ payment.CancelableProvider          = (*Stripe)(nil)
	_ payment.SubscriptionBillingProvider = (*Stripe)(nil)
	_ payment.MerchantIdentityProvider    = (*Stripe)(nil)
)
