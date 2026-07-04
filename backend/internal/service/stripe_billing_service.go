package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	stripeBillingProvider = "stripe"
	stripeBillingMode     = "subscription"
)

var (
	ErrStripeBillingUnavailable = infraerrors.ServiceUnavailable("STRIPE_BILLING_UNAVAILABLE", "stripe billing provider is unavailable")
	ErrPlanNotStripeBilling     = infraerrors.BadRequest("PLAN_NOT_STRIPE_BILLING", "plan is not configured for Stripe Billing")
	ErrStripeBillingInvalid     = infraerrors.BadRequest("STRIPE_BILLING_INVALID", "invalid Stripe Billing request")
)

type stripeBillingProviderResolver interface {
	GetSubscriptionBillingProvider(ctx context.Context, environment string) (payment.SubscriptionBillingProvider, *payment.InstanceSelection, error)
}

type stripeBillingPlanStore interface {
	GetPlan(ctx context.Context, id int64) (*dbent.SubscriptionPlan, error)
}

// StripeBillingService owns Stripe Billing subscription Checkout, Portal, and
// webhook lifecycle synchronization.
type StripeBillingService struct {
	providerResolver    stripeBillingProviderResolver
	planStore           stripeBillingPlanStore
	userRepo            UserRepository
	userSubRepo         UserSubscriptionRepository
	billingCacheService *BillingCacheService
	now                 func() time.Time
}

type CreateStripeSubscriptionCheckoutInput struct {
	UserID           int64
	UserEmail        string
	PlanID           int64
	SuccessURL       string
	CancelURL        string
	StripeCustomerID string
}

type CreateStripeSubscriptionCheckoutOutput struct {
	SessionID  string `json:"session_id"`
	URL        string `json:"url"`
	CustomerID string `json:"customer_id,omitempty"`
}

type CreateStripeBillingPortalInput struct {
	UserID    int64
	ReturnURL string
}

type CreateStripeBillingPortalOutput struct {
	SessionID string `json:"session_id"`
	URL       string `json:"url"`
}

func NewStripeBillingService(paymentService *PaymentService, configService *PaymentConfigService, userRepo UserRepository, userSubRepo UserSubscriptionRepository, billingCacheService *BillingCacheService) *StripeBillingService {
	return &StripeBillingService{
		providerResolver:    paymentService,
		planStore:           configService,
		userRepo:            userRepo,
		userSubRepo:         userSubRepo,
		billingCacheService: billingCacheService,
		now:                 time.Now,
	}
}

func (s *StripeBillingService) CreateSubscriptionCheckout(ctx context.Context, input CreateStripeSubscriptionCheckoutInput) (*CreateStripeSubscriptionCheckoutOutput, error) {
	if input.UserID <= 0 || input.PlanID <= 0 || strings.TrimSpace(input.SuccessURL) == "" || strings.TrimSpace(input.CancelURL) == "" {
		return nil, ErrStripeBillingInvalid
	}

	user, err := s.resolveStripeBillingUser(ctx, input.UserID, input.UserEmail)
	if err != nil {
		return nil, err
	}
	environment := stripeEnvironmentForUser(user)

	plan, err := s.planStore.GetPlan(ctx, input.PlanID)
	if err != nil {
		return nil, err
	}
	priceID := stripeBillingPlanPriceID(plan, environment)
	if !isStripeBillingPlan(plan) || priceID == "" {
		return nil, ErrPlanNotStripeBilling
	}

	provider, selection, err := s.provider(ctx, environment)
	if err != nil {
		return nil, err
	}

	customerID := strings.TrimSpace(input.StripeCustomerID)
	if customerID == "" {
		customerID = s.findReusableStripeCustomerID(ctx, input.UserID, environment)
	}
	customerEmail := strings.TrimSpace(input.UserEmail)
	if customerEmail == "" && user != nil {
		customerEmail = strings.TrimSpace(user.Email)
	}

	metadata := map[string]string{
		"tokengate_user_id":            strconv.FormatInt(input.UserID, 10),
		"tokengate_plan_id":            strconv.FormatInt(plan.ID, 10),
		"tokengate_group_id":           strconv.FormatInt(plan.GroupID, 10),
		"tokengate_stripe_environment": environment,
	}
	if selection != nil && strings.TrimSpace(selection.InstanceID) != "" {
		metadata["tokengate_stripe_provider_instance_id"] = strings.TrimSpace(selection.InstanceID)
	}

	req := payment.CreateSubscriptionCheckoutRequest{
		PriceID:       priceID,
		CustomerID:    customerID,
		CustomerEmail: customerEmail,
		SuccessURL:    input.SuccessURL,
		CancelURL:     input.CancelURL,
		TrialDays:     int64(plan.StripeTrialDays),
		Metadata:      metadata,
	}

	res, err := provider.CreateSubscriptionCheckout(ctx, req)
	if err != nil {
		return nil, err
	}
	return &CreateStripeSubscriptionCheckoutOutput{
		SessionID:  res.SessionID,
		URL:        res.URL,
		CustomerID: res.CustomerID,
	}, nil
}

func (s *StripeBillingService) CreateBillingPortal(ctx context.Context, input CreateStripeBillingPortalInput) (*CreateStripeBillingPortalOutput, error) {
	if input.UserID <= 0 || strings.TrimSpace(input.ReturnURL) == "" {
		return nil, ErrStripeBillingInvalid
	}
	user, err := s.resolveStripeBillingUser(ctx, input.UserID, "")
	if err != nil {
		return nil, err
	}
	environment := stripeEnvironmentForUser(user)
	customerID := s.findReusableStripeCustomerID(ctx, input.UserID, environment)
	if customerID == "" {
		return nil, infraerrors.NotFound("STRIPE_CUSTOMER_NOT_FOUND", "Stripe customer not found for user")
	}
	provider, _, err := s.provider(ctx, environment)
	if err != nil {
		return nil, err
	}
	res, err := provider.CreateBillingPortal(ctx, payment.CreateBillingPortalRequest{
		CustomerID: customerID,
		ReturnURL:  input.ReturnURL,
	})
	if err != nil {
		return nil, err
	}
	return &CreateStripeBillingPortalOutput{SessionID: res.SessionID, URL: res.URL}, nil
}

func (s *StripeBillingService) HandleNotification(ctx context.Context, n *payment.PaymentNotification) (bool, error) {
	if n == nil {
		return false, nil
	}

	switch n.Metadata["stripe_event_type"] {
	case "checkout.session.completed":
		return true, nil
	case "customer.subscription.created", "customer.subscription.updated", "customer.subscription.deleted":
		return true, s.handleSubscriptionNotification(ctx, n)
	case "invoice.payment_succeeded":
		return true, s.handleInvoicePaymentSucceeded(ctx, n)
	case "invoice.payment_failed":
		return true, s.handleInvoicePaymentFailed(ctx, n)
	default:
		return false, nil
	}
}

func (s *StripeBillingService) handleSubscriptionNotification(ctx context.Context, n *payment.PaymentNotification) error {
	subscriptionID := strings.TrimSpace(n.Metadata["stripe_subscription_id"])
	if subscriptionID == "" {
		return ErrStripeBillingInvalid
	}

	existing, err := s.userSubRepo.GetByStripeSubscriptionID(ctx, subscriptionID)
	if err != nil && err != ErrSubscriptionNotFound {
		return err
	}

	if existing == nil {
		userID, groupID, parseErr := stripeBillingUserAndGroup(n.Metadata)
		if parseErr != nil {
			return parseErr
		}
		existing = &UserSubscription{
			UserID:  userID,
			GroupID: groupID,
		}
	}

	s.applySubscriptionMetadata(existing, n.Metadata)
	if existing.ID == 0 {
		if err := s.userSubRepo.Create(ctx, existing); err != nil {
			return err
		}
	} else if err := s.userSubRepo.Update(ctx, existing); err != nil {
		return err
	}
	s.invalidateSubscriptionCache(ctx, existing.UserID, existing.GroupID)
	return nil
}

func (s *StripeBillingService) handleInvoicePaymentSucceeded(ctx context.Context, n *payment.PaymentNotification) error {
	sub, err := s.userSubRepo.GetByStripeSubscriptionID(ctx, strings.TrimSpace(n.Metadata["stripe_subscription_id"]))
	if err != nil {
		return nil
	}
	sub.Status = SubscriptionStatusActive
	sub.StripeStatus = "active"
	sub.PastDueSince = nil
	if err := s.userSubRepo.Update(ctx, sub); err != nil {
		return err
	}
	s.invalidateSubscriptionCache(ctx, sub.UserID, sub.GroupID)
	return nil
}

func (s *StripeBillingService) handleInvoicePaymentFailed(ctx context.Context, n *payment.PaymentNotification) error {
	sub, err := s.userSubRepo.GetByStripeSubscriptionID(ctx, strings.TrimSpace(n.Metadata["stripe_subscription_id"]))
	if err != nil {
		return nil
	}
	sub.Status = SubscriptionStatusSuspended
	sub.StripeStatus = "past_due"
	if sub.PastDueSince == nil {
		now := s.currentTime()
		sub.PastDueSince = &now
	}
	if err := s.userSubRepo.Update(ctx, sub); err != nil {
		return err
	}
	s.invalidateSubscriptionCache(ctx, sub.UserID, sub.GroupID)
	return nil
}

func (s *StripeBillingService) provider(ctx context.Context, environment string) (payment.SubscriptionBillingProvider, *payment.InstanceSelection, error) {
	if s == nil || s.providerResolver == nil {
		return nil, nil, ErrStripeBillingUnavailable
	}
	provider, selection, err := s.providerResolver.GetSubscriptionBillingProvider(ctx, environment)
	if err != nil {
		return nil, nil, err
	}
	if provider == nil {
		return nil, nil, ErrStripeBillingUnavailable
	}
	return provider, selection, nil
}

func (s *StripeBillingService) findReusableStripeCustomerID(ctx context.Context, userID int64, environment string) string {
	if s == nil || s.userSubRepo == nil {
		return ""
	}
	environment = payment.NormalizeProviderEnvironment(environment)
	subs, err := s.userSubRepo.ListByUserID(ctx, userID)
	if err != nil {
		return ""
	}
	for i := range subs {
		if payment.NormalizeProviderEnvironment(subs[i].StripeEnvironment) != environment {
			continue
		}
		if customerID := strings.TrimSpace(subs[i].StripeCustomerID); customerID != "" {
			return customerID
		}
	}
	return ""
}

func (s *StripeBillingService) resolveStripeBillingUser(ctx context.Context, userID int64, fallbackEmail string) (*User, error) {
	if s == nil || s.userRepo == nil {
		return &User{ID: userID, Email: strings.TrimSpace(fallbackEmail)}, nil
	}
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user != nil && strings.TrimSpace(user.Email) == "" {
		user.Email = strings.TrimSpace(fallbackEmail)
	}
	return user, nil
}

func (s *StripeBillingService) applySubscriptionMetadata(sub *UserSubscription, metadata map[string]string) {
	now := s.currentTime()
	stripeStatus := strings.TrimSpace(metadata["stripe_status"])
	mappedStatus := mapStripeSubscriptionStatus(stripeStatus)

	sub.StripeSubscriptionID = firstNonEmptyStripeBilling(metadata["stripe_subscription_id"], sub.StripeSubscriptionID)
	sub.StripeCustomerID = firstNonEmptyStripeBilling(metadata["stripe_customer_id"], sub.StripeCustomerID)
	sub.StripePriceID = firstNonEmptyStripeBilling(metadata["stripe_price_id"], sub.StripePriceID)
	sub.StripeStatus = firstNonEmptyStripeBilling(stripeStatus, sub.StripeStatus)
	if environment := strings.TrimSpace(metadata["tokengate_stripe_environment"]); environment != "" {
		sub.StripeEnvironment = payment.NormalizeProviderEnvironment(environment)
	} else if strings.TrimSpace(sub.StripeEnvironment) == "" {
		sub.StripeEnvironment = payment.ProviderEnvironmentLive
	}
	sub.StripeProviderInstanceID = firstNonEmptyStripeBilling(metadata["tokengate_stripe_provider_instance_id"], sub.StripeProviderInstanceID)
	sub.Status = mappedStatus
	sub.CancelAtPeriodEnd = parseBoolString(metadata["cancel_at_period_end"])

	if t := parseUnixTimestamp(metadata["current_period_start"]); t != nil {
		sub.CurrentPeriodStart = t
		sub.StartsAt = *t
	} else if sub.StartsAt.IsZero() {
		sub.StartsAt = now
	}
	if t := parseUnixTimestamp(metadata["current_period_end"]); t != nil {
		sub.CurrentPeriodEnd = t
		sub.ExpiresAt = *t
	}
	if t := parseUnixTimestamp(metadata["trial_start"]); t != nil {
		sub.TrialStart = t
		sub.TrialUsed = true
	}
	if t := parseUnixTimestamp(metadata["trial_end"]); t != nil {
		sub.TrialEnd = t
		sub.TrialUsed = true
		if sub.ExpiresAt.IsZero() {
			sub.ExpiresAt = *t
		}
	}
	if sub.ExpiresAt.IsZero() || mappedStatus == SubscriptionStatusExpired && metadata["stripe_event_type"] == "customer.subscription.deleted" {
		sub.ExpiresAt = now
	}

	if mappedStatus == SubscriptionStatusSuspended {
		if sub.PastDueSince == nil {
			sub.PastDueSince = &now
		}
	} else {
		sub.PastDueSince = nil
	}
}

func (s *StripeBillingService) invalidateSubscriptionCache(ctx context.Context, userID, groupID int64) {
	if s == nil || s.billingCacheService == nil || userID <= 0 || groupID <= 0 {
		return
	}
	_ = s.billingCacheService.InvalidateSubscription(ctx, userID, groupID)
}

func (s *StripeBillingService) currentTime() time.Time {
	if s != nil && s.now != nil {
		return s.now().UTC()
	}
	return time.Now().UTC()
}

func isStripeBillingPlan(plan *dbent.SubscriptionPlan) bool {
	if plan == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(plan.BillingProvider), stripeBillingProvider) &&
		strings.EqualFold(strings.TrimSpace(plan.BillingMode), stripeBillingMode)
}

func derefStripePlanString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func stripeBillingPlanPriceID(plan *dbent.SubscriptionPlan, environment string) string {
	if payment.NormalizeProviderEnvironment(environment) == payment.ProviderEnvironmentSandbox {
		return strings.TrimSpace(derefStripePlanString(plan.StripeSandboxPriceID))
	}
	return strings.TrimSpace(derefStripePlanString(plan.StripePriceID))
}

func stripeBillingUserAndGroup(metadata map[string]string) (int64, int64, error) {
	userID, err := strconv.ParseInt(strings.TrimSpace(metadata["tokengate_user_id"]), 10, 64)
	if err != nil || userID <= 0 {
		return 0, 0, fmt.Errorf("%w: tokengate_user_id is required", ErrStripeBillingInvalid)
	}
	groupID, err := strconv.ParseInt(strings.TrimSpace(metadata["tokengate_group_id"]), 10, 64)
	if err != nil || groupID <= 0 {
		return 0, 0, fmt.Errorf("%w: tokengate_group_id is required", ErrStripeBillingInvalid)
	}
	return userID, groupID, nil
}

func mapStripeSubscriptionStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "active", "trialing":
		return SubscriptionStatusActive
	case "past_due", "incomplete":
		return SubscriptionStatusSuspended
	case "canceled", "unpaid", "incomplete_expired":
		return SubscriptionStatusExpired
	default:
		return SubscriptionStatusSuspended
	}
}

func parseUnixTimestamp(raw string) *time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "0" {
		return nil
	}
	sec, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || sec <= 0 {
		return nil
	}
	t := time.Unix(sec, 0).UTC()
	return &t
}

func parseBoolString(raw string) bool {
	v, err := strconv.ParseBool(strings.TrimSpace(raw))
	return err == nil && v
}

func firstNonEmptyStripeBilling(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
