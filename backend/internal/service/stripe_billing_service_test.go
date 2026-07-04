package service

import (
	"context"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestStripeBillingServiceCreateCheckoutPassesPlanTrialAndMetadata(t *testing.T) {
	priceID := "price_byo_monthly"
	provider := &stripeBillingProviderStub{}
	repo := newStripeBillingMemorySubRepo()
	svc := newStripeBillingServiceForTest(provider, &dbent.SubscriptionPlan{
		ID:              42,
		GroupID:         7,
		StripePriceID:   &priceID,
		StripeTrialDays: 7,
		BillingProvider: "stripe",
		BillingMode:     "subscription",
	}, repo)

	out, err := svc.CreateSubscriptionCheckout(context.Background(), CreateStripeSubscriptionCheckoutInput{
		UserID:           101,
		UserEmail:        "user@example.com",
		PlanID:           42,
		SuccessURL:       "https://app.tokengate.to/subscriptions?checkout=success",
		CancelURL:        "https://app.tokengate.to/subscriptions?checkout=cancel",
		StripeCustomerID: "cus_existing",
	})

	require.NoError(t, err)
	require.Equal(t, "cs_test_123", out.SessionID)
	require.Equal(t, "https://checkout.stripe.test/session", out.URL)
	require.Equal(t, "price_byo_monthly", provider.checkoutReq.PriceID)
	require.Equal(t, "cus_existing", provider.checkoutReq.CustomerID)
	require.Equal(t, int64(7), provider.checkoutReq.TrialDays)
	require.Equal(t, "101", provider.checkoutReq.Metadata["tokengate_user_id"])
	require.Equal(t, "42", provider.checkoutReq.Metadata["tokengate_plan_id"])
	require.Equal(t, "7", provider.checkoutReq.Metadata["tokengate_group_id"])
}

func TestStripeBillingServiceCreateCheckoutUsesSandboxPriceAndCustomerForTestUser(t *testing.T) {
	livePriceID := "price_live_byo_monthly"
	sandboxPriceID := "price_sandbox_byo_monthly"
	provider := &stripeBillingProviderStub{}
	repo := newStripeBillingMemorySubRepo()
	require.NoError(t, repo.Create(context.Background(), &UserSubscription{
		UserID:            101,
		GroupID:           7,
		StripeCustomerID:  "cus_live_101",
		StripeEnvironment: payment.ProviderEnvironmentLive,
	}))
	require.NoError(t, repo.Create(context.Background(), &UserSubscription{
		UserID:            101,
		GroupID:           7,
		StripeCustomerID:  "cus_sandbox_101",
		StripeEnvironment: payment.ProviderEnvironmentSandbox,
	}))
	svc := newStripeBillingServiceForTest(provider, &dbent.SubscriptionPlan{
		ID:                   42,
		GroupID:              7,
		StripePriceID:        &livePriceID,
		StripeSandboxPriceID: &sandboxPriceID,
		StripeTrialDays:      7,
		BillingProvider:      "stripe",
		BillingMode:          "subscription",
	}, repo)
	svc.userRepo = stripeBillingUserRepoStub{user: &User{
		ID:         101,
		Email:      "test-user@example.com",
		IsTestUser: true,
	}}

	out, err := svc.CreateSubscriptionCheckout(context.Background(), CreateStripeSubscriptionCheckoutInput{
		UserID:     101,
		PlanID:     42,
		SuccessURL: "https://app.tokengate.to/subscriptions?checkout=success",
		CancelURL:  "https://app.tokengate.to/subscriptions?checkout=cancel",
	})

	require.NoError(t, err)
	require.Equal(t, "cs_test_123", out.SessionID)
	require.Equal(t, payment.ProviderEnvironmentSandbox, provider.resolvedEnvironment)
	require.Equal(t, sandboxPriceID, provider.checkoutReq.PriceID)
	require.Equal(t, "cus_sandbox_101", provider.checkoutReq.CustomerID)
	require.Equal(t, "test-user@example.com", provider.checkoutReq.CustomerEmail)
	require.Equal(t, payment.ProviderEnvironmentSandbox, provider.checkoutReq.Metadata["tokengate_stripe_environment"])
}

func TestStripeBillingServiceCreateCheckoutRequiresStripeBillingPlan(t *testing.T) {
	priceID := "price_byo_monthly"
	provider := &stripeBillingProviderStub{}
	svc := newStripeBillingServiceForTest(provider, &dbent.SubscriptionPlan{
		ID:              42,
		GroupID:         7,
		StripePriceID:   &priceID,
		StripeTrialDays: 7,
		BillingProvider: "internal",
		BillingMode:     "fixed_period",
	}, newStripeBillingMemorySubRepo())

	_, err := svc.CreateSubscriptionCheckout(context.Background(), CreateStripeSubscriptionCheckoutInput{
		UserID:     101,
		UserEmail:  "user@example.com",
		PlanID:     42,
		SuccessURL: "https://app.tokengate.to/subscriptions?checkout=success",
		CancelURL:  "https://app.tokengate.to/subscriptions?checkout=cancel",
	})

	require.Error(t, err)
	require.Equal(t, "PLAN_NOT_STRIPE_BILLING", infraerrors.Reason(err))
	require.False(t, provider.checkoutCalled)
}

func TestStripeBillingServiceHandleSubscriptionNotificationCreatesAndReplaysIdempotently(t *testing.T) {
	provider := &stripeBillingProviderStub{}
	repo := newStripeBillingMemorySubRepo()
	svc := newStripeBillingServiceForTest(provider, nil, repo)
	svc.now = func() time.Time { return time.Unix(1800000100, 0).UTC() }

	notification := &payment.PaymentNotification{
		TradeNo: "sub_123",
		Status:  payment.ProviderStatusSuccess,
		Metadata: map[string]string{
			"stripe_event_type":      "customer.subscription.updated",
			"stripe_subscription_id": "sub_123",
			"stripe_customer_id":     "cus_123",
			"stripe_price_id":        "price_byo_monthly",
			"stripe_status":          "trialing",
			"current_period_start":   "1800000000",
			"current_period_end":     "1802592000",
			"trial_start":            "1800000000",
			"trial_end":              "1800604800",
			"cancel_at_period_end":   "false",
			"tokengate_user_id":      "101",
			"tokengate_plan_id":      "42",
			"tokengate_group_id":     "7",
		},
	}

	handled, err := svc.HandleNotification(context.Background(), notification)
	require.NoError(t, err)
	require.True(t, handled)
	handled, err = svc.HandleNotification(context.Background(), notification)
	require.NoError(t, err)
	require.True(t, handled)

	require.Equal(t, 1, repo.createCount, "webhook replay must not create duplicate subscriptions")
	sub, err := repo.GetByStripeSubscriptionID(context.Background(), "sub_123")
	require.NoError(t, err)
	require.Equal(t, int64(101), sub.UserID)
	require.Equal(t, int64(7), sub.GroupID)
	require.Equal(t, SubscriptionStatusActive, sub.Status)
	require.Equal(t, "trialing", sub.StripeStatus)
	require.True(t, sub.TrialUsed)
	require.False(t, sub.CancelAtPeriodEnd)
	require.Equal(t, time.Unix(1802592000, 0).UTC(), sub.ExpiresAt)
	require.NotNil(t, sub.TrialEnd)
}

func TestStripeBillingServiceInvoicePaymentFailedSuspendsExistingSubscription(t *testing.T) {
	repo := newStripeBillingMemorySubRepo()
	now := time.Unix(1800000100, 0).UTC()
	existing := &UserSubscription{
		UserID:               101,
		GroupID:              7,
		StartsAt:             now.Add(-24 * time.Hour),
		ExpiresAt:            now.Add(30 * 24 * time.Hour),
		Status:               SubscriptionStatusActive,
		StripeCustomerID:     "cus_123",
		StripeSubscriptionID: "sub_123",
		StripePriceID:        "price_byo_monthly",
		StripeStatus:         "active",
	}
	require.NoError(t, repo.Create(context.Background(), existing))

	svc := newStripeBillingServiceForTest(&stripeBillingProviderStub{}, nil, repo)
	svc.now = func() time.Time { return now }

	handled, err := svc.HandleNotification(context.Background(), &payment.PaymentNotification{
		TradeNo: "in_123",
		Status:  payment.ProviderStatusFailed,
		Metadata: map[string]string{
			"stripe_event_type":      "invoice.payment_failed",
			"stripe_invoice_id":      "in_123",
			"stripe_subscription_id": "sub_123",
			"stripe_customer_id":     "cus_123",
			"stripe_status":          "open",
		},
	})

	require.NoError(t, err)
	require.True(t, handled)
	sub, err := repo.GetByStripeSubscriptionID(context.Background(), "sub_123")
	require.NoError(t, err)
	require.Equal(t, SubscriptionStatusSuspended, sub.Status)
	require.Equal(t, "past_due", sub.StripeStatus)
	require.NotNil(t, sub.PastDueSince)
	require.Equal(t, now, *sub.PastDueSince)
	require.Equal(t, 1, repo.createCount)
}

func newStripeBillingServiceForTest(provider payment.SubscriptionBillingProvider, plan *dbent.SubscriptionPlan, repo *stripeBillingMemorySubRepo) *StripeBillingService {
	return &StripeBillingService{
		providerResolver: stripeBillingProviderResolverStub{provider: provider},
		planStore:        stripeBillingPlanStoreStub{plan: plan},
		userSubRepo:      repo,
		now:              time.Now,
	}
}

type stripeBillingProviderResolverStub struct {
	provider payment.SubscriptionBillingProvider
}

func (s stripeBillingProviderResolverStub) GetSubscriptionBillingProvider(_ context.Context, environment string) (payment.SubscriptionBillingProvider, *payment.InstanceSelection, error) {
	environment = payment.NormalizeProviderEnvironment(environment)
	if stub, ok := s.provider.(*stripeBillingProviderStub); ok {
		stub.resolvedEnvironment = environment
	}
	return s.provider, &payment.InstanceSelection{
		InstanceID:  "stripe-" + environment,
		ProviderKey: payment.TypeStripe,
		Environment: environment,
	}, nil
}

type stripeBillingUserRepoStub struct {
	UserRepository
	user *User
	err  error
}

func (s stripeBillingUserRepoStub) GetByID(context.Context, int64) (*User, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.user == nil {
		return &User{}, nil
	}
	cp := *s.user
	return &cp, nil
}

type stripeBillingPlanStoreStub struct {
	plan *dbent.SubscriptionPlan
	err  error
}

func (s stripeBillingPlanStoreStub) GetPlan(context.Context, int64) (*dbent.SubscriptionPlan, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.plan == nil {
		return nil, infraerrors.NotFound("PLAN_NOT_FOUND", "subscription plan not found")
	}
	return s.plan, nil
}

type stripeBillingMemorySubRepo struct {
	UserSubscriptionRepository

	nextID      int64
	createCount int
	updateCount int
	subs        map[int64]*UserSubscription
}

func newStripeBillingMemorySubRepo() *stripeBillingMemorySubRepo {
	return &stripeBillingMemorySubRepo{
		nextID: 1,
		subs:   make(map[int64]*UserSubscription),
	}
}

func (r *stripeBillingMemorySubRepo) Create(_ context.Context, sub *UserSubscription) error {
	r.createCount++
	cp := *sub
	if cp.ID == 0 {
		cp.ID = r.nextID
		r.nextID++
	}
	r.subs[cp.ID] = &cp
	sub.ID = cp.ID
	return nil
}

func (r *stripeBillingMemorySubRepo) Update(_ context.Context, sub *UserSubscription) error {
	r.updateCount++
	if _, ok := r.subs[sub.ID]; !ok {
		return ErrSubscriptionNotFound
	}
	cp := *sub
	r.subs[sub.ID] = &cp
	return nil
}

func (r *stripeBillingMemorySubRepo) GetByStripeSubscriptionID(_ context.Context, stripeSubscriptionID string) (*UserSubscription, error) {
	for _, sub := range r.subs {
		if sub.StripeSubscriptionID == stripeSubscriptionID {
			cp := *sub
			return &cp, nil
		}
	}
	return nil, ErrSubscriptionNotFound
}

func (r *stripeBillingMemorySubRepo) ListByStripeCustomerID(_ context.Context, stripeCustomerID string) ([]UserSubscription, error) {
	out := make([]UserSubscription, 0)
	for _, sub := range r.subs {
		if sub.StripeCustomerID == stripeCustomerID {
			out = append(out, *sub)
		}
	}
	return out, nil
}

func (r *stripeBillingMemorySubRepo) ListByUserID(_ context.Context, userID int64) ([]UserSubscription, error) {
	out := make([]UserSubscription, 0)
	for _, sub := range r.subs {
		if sub.UserID == userID {
			out = append(out, *sub)
		}
	}
	return out, nil
}

func (r *stripeBillingMemorySubRepo) ListActiveByUserID(_ context.Context, userID int64) ([]UserSubscription, error) {
	out := make([]UserSubscription, 0)
	now := time.Now()
	for _, sub := range r.subs {
		if sub.UserID == userID && sub.Status == SubscriptionStatusActive && now.Before(sub.ExpiresAt) {
			out = append(out, *sub)
		}
	}
	return out, nil
}

type stripeBillingProviderStub struct {
	resolvedEnvironment string
	checkoutCalled      bool
	checkoutReq         payment.CreateSubscriptionCheckoutRequest
	portalReq           payment.CreateBillingPortalRequest
}

func (s *stripeBillingProviderStub) Name() string { return "stripe" }
func (s *stripeBillingProviderStub) ProviderKey() string {
	return payment.TypeStripe
}
func (s *stripeBillingProviderStub) SupportedTypes() []payment.PaymentType {
	return []payment.PaymentType{payment.TypeStripe}
}
func (s *stripeBillingProviderStub) CreatePayment(context.Context, payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	return nil, nil
}
func (s *stripeBillingProviderStub) QueryOrder(context.Context, string) (*payment.QueryOrderResponse, error) {
	return nil, nil
}
func (s *stripeBillingProviderStub) VerifyNotification(context.Context, string, map[string]string) (*payment.PaymentNotification, error) {
	return nil, nil
}
func (s *stripeBillingProviderStub) Refund(context.Context, payment.RefundRequest) (*payment.RefundResponse, error) {
	return nil, nil
}
func (s *stripeBillingProviderStub) CreateSubscriptionCheckout(_ context.Context, req payment.CreateSubscriptionCheckoutRequest) (*payment.CreateSubscriptionCheckoutResponse, error) {
	s.checkoutCalled = true
	s.checkoutReq = req
	return &payment.CreateSubscriptionCheckoutResponse{
		SessionID:  "cs_test_123",
		URL:        "https://checkout.stripe.test/session",
		CustomerID: req.CustomerID,
	}, nil
}
func (s *stripeBillingProviderStub) CreateBillingPortal(_ context.Context, req payment.CreateBillingPortalRequest) (*payment.CreateBillingPortalResponse, error) {
	s.portalReq = req
	return &payment.CreateBillingPortalResponse{
		SessionID: "bps_test_123",
		URL:       "https://billing.stripe.test/session",
	}, nil
}
