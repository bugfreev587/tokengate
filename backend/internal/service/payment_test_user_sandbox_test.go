//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

type captureEnvironmentLoadBalancer struct {
	env string
}

func (l *captureEnvironmentLoadBalancer) GetInstanceConfig(context.Context, int64) (map[string]string, error) {
	return map[string]string{}, nil
}

func (l *captureEnvironmentLoadBalancer) SelectInstance(ctx context.Context, _ string, paymentType payment.PaymentType, _ payment.Strategy, _ float64) (*payment.InstanceSelection, error) {
	l.env = payment.ProviderEnvironmentFromContext(ctx)
	return &payment.InstanceSelection{
		InstanceID:     "22",
		ProviderKey:    payment.TypeStripe,
		Environment:    l.env,
		SupportedTypes: string(paymentType),
		Config:         map[string]string{"currency": "USD"},
	}, nil
}

func TestSelectCreateOrderInstancePassesSandboxEnvironmentForTestUserStripePayments(t *testing.T) {
	lb := &captureEnvironmentLoadBalancer{}
	svc := &PaymentService{loadBalancer: lb}

	sel, err := svc.selectCreateOrderInstance(context.Background(), CreateOrderRequest{
		PaymentType:       payment.TypeStripe,
		StripeEnvironment: payment.ProviderEnvironmentSandbox,
	}, &PaymentConfig{LoadBalanceStrategy: payment.DefaultLoadBalanceStrategy}, 19.99)

	require.NoError(t, err)
	require.Equal(t, payment.ProviderEnvironmentSandbox, lb.env)
	require.Equal(t, payment.ProviderEnvironmentSandbox, sel.Environment)
}
