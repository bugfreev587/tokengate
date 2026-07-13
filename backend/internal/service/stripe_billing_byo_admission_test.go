//go:build unit

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStripeBillingServiceSyncBYOAccountEntitlementInvalidatesAuthCache(t *testing.T) {
	checker := &stripeBillingBYOCheckerStub{active: true}
	updater := &stripeBillingBYOUpdaterStub{}
	invalidator := &authCacheInvalidatorStub{}
	svc := &StripeBillingService{byoEntitlementChecker: checker, byoAccountUpdater: updater, authCacheInvalidator: invalidator}

	require.NoError(t, svc.syncBYOAccountEntitlement(context.Background(), 101))
	require.Equal(t, []stripeBillingBYOUpdateCall{{userID: 101, enabled: true}}, updater.calls)
	require.Equal(t, []int64{101}, invalidator.userIDs)
}

func TestStripeBillingServiceSyncBYOAccountEntitlementDoesNotInvalidateOnUpdateError(t *testing.T) {
	updater := &stripeBillingBYOUpdaterStub{err: errors.New("update failed")}
	invalidator := &authCacheInvalidatorStub{}
	svc := &StripeBillingService{byoEntitlementChecker: &stripeBillingBYOCheckerStub{}, byoAccountUpdater: updater, authCacheInvalidator: invalidator}

	require.Error(t, svc.syncBYOAccountEntitlement(context.Background(), 101))
	require.Empty(t, invalidator.userIDs)
}

func TestStripeBillingServiceSyncBYOAccountEntitlementDoesNotDisableOnLookupError(t *testing.T) {
	lookupErr := errors.New("subscription lookup failed")
	updater := &stripeBillingBYOUpdaterStub{}
	svc := &StripeBillingService{
		byoEntitlementChecker: &stripeBillingBYOCheckerStub{err: lookupErr},
		byoAccountUpdater:     updater,
	}

	require.ErrorIs(t, svc.syncBYOAccountEntitlement(context.Background(), 101), lookupErr)
	require.Empty(t, updater.calls)
}

type stripeBillingBYOCheckerStub struct {
	active bool
	err    error
}

func (s *stripeBillingBYOCheckerStub) HasActiveBYOSubscription(context.Context, int64) (bool, error) {
	return s.active, s.err
}

type stripeBillingBYOUpdateCall struct {
	userID  int64
	enabled bool
}

type stripeBillingBYOUpdaterStub struct {
	calls []stripeBillingBYOUpdateCall
	err   error
}

func (s *stripeBillingBYOUpdaterStub) SetOwnerBYOAccountEntitlement(_ context.Context, ownerUserID int64, enabled bool) (int64, error) {
	s.calls = append(s.calls, stripeBillingBYOUpdateCall{userID: ownerUserID, enabled: enabled})
	if s.err != nil {
		return 0, s.err
	}
	return 1, nil
}

func (s *stripeBillingBYOUpdaterStub) ReconcileBYOAccountEntitlements(context.Context) ([]int64, error) {
	return nil, nil
}
