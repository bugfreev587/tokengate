//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAPIKeyServiceCanBindUserOwnedConnectedAccountGroup(t *testing.T) {
	ownerID := int64(42)
	group := &Group{
		ID:             99,
		IsExclusive:    true,
		OwnerUserID:    &ownerID,
		CapacitySource: CapacitySourceConnectedAccount,
	}
	svc := &APIKeyService{}

	require.True(t, svc.canUserBindGroup(context.Background(), &User{ID: ownerID}, group))
	require.False(t, svc.canUserBindGroup(context.Background(), &User{ID: 77}, group))
}

func TestAPIKeyServiceCanBindInactiveBYOGroup(t *testing.T) {
	ownerID := int64(42)
	byoEnabled := false
	group := &Group{
		ID: 99, Status: StatusActive, IsExclusive: true, OwnerUserID: &ownerID,
		CapacitySource: CapacitySourceConnectedAccount, BYOEnabled: &byoEnabled,
		BYODisabledReason: BYOAccountDisabledReasonSubscriptionInactive,
	}
	svc := &APIKeyService{}
	require.True(t, svc.canUserBindGroup(context.Background(), &User{ID: ownerID}, group))
}

func TestAPIKeyServiceCanSwitchFromBYOToAvailableTokenGateGroup(t *testing.T) {
	svc := &APIKeyService{}
	group := &Group{ID: 101, Status: StatusActive, IsExclusive: false, CapacitySource: CapacitySourceTokenGate}
	require.True(t, svc.canUserBindGroup(context.Background(), &User{ID: 42}, group))
}

func TestAPIKeyServiceAvailableGroupsIncludesOnlyOwnedConnectedAccountGroups(t *testing.T) {
	ownerID := int64(42)
	otherOwnerID := int64(77)
	ownedGroup := &Group{
		ID:             99,
		IsExclusive:    true,
		OwnerUserID:    &ownerID,
		CapacitySource: CapacitySourceConnectedAccount,
	}
	otherOwnedGroup := &Group{
		ID:             100,
		IsExclusive:    true,
		OwnerUserID:    &otherOwnerID,
		CapacitySource: CapacitySourceConnectedAccount,
	}
	publicGroup := &Group{ID: 101, IsExclusive: false, CapacitySource: CapacitySourceTokenGate}
	svc := &APIKeyService{}
	user := &User{ID: ownerID}

	require.True(t, svc.canUserBindGroupInternal(user, ownedGroup, nil))
	require.False(t, svc.canUserBindGroupInternal(user, otherOwnedGroup, nil))
	require.True(t, svc.canUserBindGroupInternal(user, publicGroup, nil))
}
