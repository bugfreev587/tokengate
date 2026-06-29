//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type stubGlobalModelPricingOverrideRepo struct {
	byModel map[string]*GlobalModelPricingOverride
}

func (s *stubGlobalModelPricingOverrideRepo) GetByModel(_ context.Context, model string) (*GlobalModelPricingOverride, error) {
	if s == nil {
		return nil, nil
	}
	if override, ok := s.byModel[model]; ok {
		cp := *override
		return &cp, nil
	}
	return nil, nil
}

func (s *stubGlobalModelPricingOverrideRepo) List(_ context.Context) ([]GlobalModelPricingOverride, error) {
	if s == nil {
		return nil, nil
	}
	items := make([]GlobalModelPricingOverride, 0, len(s.byModel))
	for _, override := range s.byModel {
		items = append(items, *override)
	}
	return items, nil
}

func (s *stubGlobalModelPricingOverrideRepo) Upsert(_ context.Context, override *GlobalModelPricingOverride) error {
	if s.byModel == nil {
		s.byModel = map[string]*GlobalModelPricingOverride{}
	}
	cp := *override
	s.byModel[override.Model] = &cp
	return nil
}

func (s *stubGlobalModelPricingOverrideRepo) DeleteByModel(_ context.Context, model string) error {
	delete(s.byModel, model)
	return nil
}

func TestResolve_WithGlobalOverride_TokenFlat(t *testing.T) {
	bs := newTestBillingServiceForResolver()
	r := NewModelPricingResolverWithGlobalOverrides(
		&ChannelService{},
		bs,
		&stubGlobalModelPricingOverrideRepo{byModel: map[string]*GlobalModelPricingOverride{
			"claude-sonnet-4": {
				Model:      "claude-sonnet-4",
				InputPrice: testPtrFloat64(9e-6),
			},
		}},
	)

	resolved := r.Resolve(context.Background(), PricingInput{
		Model:   "claude-sonnet-4",
		GroupID: nil,
	})

	require.NotNil(t, resolved)
	require.Equal(t, BillingModeToken, resolved.Mode)
	require.Equal(t, PricingSourceGlobalOverride, resolved.Source)
	require.NotNil(t, resolved.BasePricing)
	require.InDelta(t, 9e-6, resolved.BasePricing.InputPricePerToken, 1e-12)
	require.InDelta(t, 15e-6, resolved.BasePricing.OutputPricePerToken, 1e-12)
}

func TestResolve_WithGlobalOverride_ImageMode(t *testing.T) {
	bs := newTestBillingServiceForResolver()
	r := NewModelPricingResolverWithGlobalOverrides(
		&ChannelService{},
		bs,
		&stubGlobalModelPricingOverrideRepo{byModel: map[string]*GlobalModelPricingOverride{
			"gpt-image-test": {
				Model:           "gpt-image-test",
				BillingMode:     BillingModeImage,
				PerRequestPrice: testPtrFloat64(0.12),
			},
		}},
	)

	resolved := r.Resolve(context.Background(), PricingInput{
		Model:   "gpt-image-test",
		GroupID: nil,
	})

	require.NotNil(t, resolved)
	require.Equal(t, BillingModeImage, resolved.Mode)
	require.Equal(t, PricingSourceGlobalOverride, resolved.Source)
	require.InDelta(t, 0.12, resolved.DefaultPerRequestPrice, 1e-12)
}

func TestResolve_WithChannelOverrideStillWinsOverGlobalOverride(t *testing.T) {
	r := newResolverWithChannel(t, []ChannelModelPricing{{
		Platform:    "anthropic",
		Models:      []string{"claude-sonnet-4"},
		BillingMode: BillingModeToken,
		InputPrice:  testPtrFloat64(11e-6),
	}})
	r.globalOverrideRepo = &stubGlobalModelPricingOverrideRepo{byModel: map[string]*GlobalModelPricingOverride{
		"claude-sonnet-4": {
			Model:      "claude-sonnet-4",
			InputPrice: testPtrFloat64(9e-6),
		},
	}}

	resolved := r.Resolve(context.Background(), PricingInput{
		Model:   "claude-sonnet-4",
		GroupID: groupIDPtr(),
	})

	require.NotNil(t, resolved)
	require.Equal(t, BillingModeToken, resolved.Mode)
	require.Equal(t, PricingSourceChannel, resolved.Source)
	require.NotNil(t, resolved.BasePricing)
	require.InDelta(t, 11e-6, resolved.BasePricing.InputPricePerToken, 1e-12)
}

func TestGatewayResolveChannelPricingReturnsGlobalOverride(t *testing.T) {
	bs := newTestBillingServiceForResolver()
	resolver := NewModelPricingResolverWithGlobalOverrides(
		nil,
		bs,
		&stubGlobalModelPricingOverrideRepo{byModel: map[string]*GlobalModelPricingOverride{
			"claude-sonnet-4": {
				Model:      "claude-sonnet-4",
				InputPrice: testPtrFloat64(9e-6),
			},
		}},
	)
	svc := &GatewayService{resolver: resolver}
	apiKey := &APIKey{Group: &Group{ID: 100}}

	resolved := svc.resolveChannelPricing(context.Background(), "claude-sonnet-4", apiKey)

	require.NotNil(t, resolved)
	require.Equal(t, PricingSourceGlobalOverride, resolved.Source)
	require.NotNil(t, resolved.BasePricing)
	require.InDelta(t, 9e-6, resolved.BasePricing.InputPricePerToken, 1e-12)
}
