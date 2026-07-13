//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSubscriptionExpiryServiceReconcilesBYOEntitlementsEvenWhenNoStatusRowsChanged(t *testing.T) {
	repo := &subscriptionExpiryRepoStub{}
	updater := &subscriptionExpiryBYOUpdaterStub{reconcileOwners: []int64{101}}
	svc := NewSubscriptionExpiryService(repo, time.Minute, updater)

	svc.runOnce()

	require.Equal(t, int64(0), repo.batchUpdated)
	require.Equal(t, 1, updater.reconcileCalls)
}

func TestSubscriptionExpiryServiceRetriesReconciliationAfterFailure(t *testing.T) {
	repo := &subscriptionExpiryRepoStub{}
	updater := &subscriptionExpiryBYOUpdaterStub{
		reconcileOwners: []int64{101},
		reconcileErrors: []error{errors.New("temporary database failure"), nil},
	}
	svc := NewSubscriptionExpiryService(repo, time.Minute, updater)

	svc.runOnce()
	svc.runOnce()

	require.Equal(t, 2, updater.reconcileCalls)
}

func TestSubscriptionExpiryServiceNilRunOnceIsSafe(t *testing.T) {
	var svc *SubscriptionExpiryService
	require.NotPanics(t, svc.runOnce)
}

type subscriptionExpiryRepoStub struct {
	UserSubscriptionRepository
	batchUpdated int64
}

func (r *subscriptionExpiryRepoStub) BatchUpdateExpiredStatus(context.Context) (int64, error) {
	return r.batchUpdated, nil
}

type subscriptionExpiryBYOUpdateCall struct {
	userID  int64
	enabled bool
}

type subscriptionExpiryBYOUpdaterStub struct {
	calls           []subscriptionExpiryBYOUpdateCall
	reconcileCalls  int
	reconcileOwners []int64
	reconcileErrors []error
}

func (s *subscriptionExpiryBYOUpdaterStub) SetOwnerBYOAccountEntitlement(_ context.Context, ownerUserID int64, enabled bool) (int64, error) {
	s.calls = append(s.calls, subscriptionExpiryBYOUpdateCall{userID: ownerUserID, enabled: enabled})
	return 1, nil
}

func (s *subscriptionExpiryBYOUpdaterStub) ReconcileBYOAccountEntitlements(context.Context) ([]int64, error) {
	call := s.reconcileCalls
	s.reconcileCalls++
	if call < len(s.reconcileErrors) && s.reconcileErrors[call] != nil {
		return nil, s.reconcileErrors[call]
	}
	return append([]int64(nil), s.reconcileOwners...), nil
}
