package service

import (
	"context"
	"fmt"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	paymentprovider "github.com/Wei-Shaw/sub2api/internal/payment/provider"
)

// GetSubscriptionBillingProvider returns the configured provider that supports
// hosted subscription billing in the requested provider environment.
func (s *PaymentService) GetSubscriptionBillingProvider(ctx context.Context, environment string) (payment.SubscriptionBillingProvider, *payment.InstanceSelection, error) {
	if s == nil || s.loadBalancer == nil {
		return nil, nil, ErrStripeBillingUnavailable
	}

	environment = payment.NormalizeProviderEnvironment(environment)
	selection, err := s.loadBalancer.SelectInstance(
		payment.WithProviderEnvironment(ctx, environment),
		payment.TypeStripe,
		payment.TypeStripe,
		payment.Strategy(payment.DefaultLoadBalanceStrategy),
		0,
	)
	if err != nil {
		return nil, nil, err
	}
	if selection == nil || selection.ProviderKey != payment.TypeStripe {
		return nil, nil, ErrStripeBillingUnavailable
	}

	provider, err := paymentprovider.CreateProvider(selection.ProviderKey, selection.InstanceID, selection.Config)
	if err != nil {
		return nil, nil, err
	}
	billingProvider, ok := provider.(payment.SubscriptionBillingProvider)
	if !ok {
		return nil, nil, fmt.Errorf("%w: provider %s does not support subscription billing", ErrStripeBillingUnavailable, provider.ProviderKey())
	}
	return billingProvider, selection, nil
}

func (s *PaymentService) ResolveStripeEnvironmentForUser(ctx context.Context, userID int64) string {
	if s == nil || s.userRepo == nil || userID <= 0 {
		return payment.ProviderEnvironmentLive
	}
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return payment.ProviderEnvironmentLive
	}
	return stripeEnvironmentForUser(user)
}
