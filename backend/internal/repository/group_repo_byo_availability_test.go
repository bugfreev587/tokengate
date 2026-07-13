//go:build unit

package repository

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestApplyBYOAccountAvailability(t *testing.T) {
	ownerID := int64(42)
	makeGroup := func() service.Group {
		return service.Group{ID: 10, OwnerUserID: &ownerID, CapacitySource: service.CapacitySourceConnectedAccount}
	}

	t.Run("enabled when at least one account is schedulable", func(t *testing.T) {
		group := makeGroup()
		applyBYOAccountAvailability(&group, groupAccountCounts{Total: 1, Active: 1})
		require.NotNil(t, group.BYOEnabled)
		require.True(t, *group.BYOEnabled)
		require.Empty(t, group.BYODisabledReason)
	})

	t.Run("subscription inactive reason wins for disabled BYO accounts", func(t *testing.T) {
		group := makeGroup()
		applyBYOAccountAvailability(&group, groupAccountCounts{Total: 1, SubscriptionInactive: true})
		require.NotNil(t, group.BYOEnabled)
		require.False(t, *group.BYOEnabled)
		require.Equal(t, service.BYOAccountDisabledReasonSubscriptionInactive, group.BYODisabledReason)
	})

	t.Run("subscription inactive wins even when schedulable was changed operationally", func(t *testing.T) {
		group := makeGroup()
		applyBYOAccountAvailability(&group, groupAccountCounts{Total: 2, Active: 1, SubscriptionInactive: true})
		require.False(t, *group.BYOEnabled)
		require.Equal(t, service.BYOAccountDisabledReasonSubscriptionInactive, group.BYODisabledReason)
	})

	t.Run("group without accounts reports account missing", func(t *testing.T) {
		group := makeGroup()
		applyBYOAccountAvailability(&group, groupAccountCounts{})
		require.False(t, *group.BYOEnabled)
		require.Equal(t, service.BYOAccountDisabledReasonNoAccount, group.BYODisabledReason)
	})

	t.Run("manual account disable stays distinct", func(t *testing.T) {
		group := makeGroup()
		applyBYOAccountAvailability(&group, groupAccountCounts{Total: 1})
		require.False(t, *group.BYOEnabled)
		require.Equal(t, service.BYOAccountDisabledReasonAccountDisabled, group.BYODisabledReason)
	})
}
