//go:build unit

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestUserSubscriptionRepositoryHasActiveBYOSubscription(t *testing.T) {
	_, client := newAPIKeyRepoSQLite(t)
	ctx := context.Background()
	repo := &userSubscriptionRepository{client: client}
	group, err := client.Group.Create().SetName("byo-plan").SetStatus(service.StatusActive).Save(ctx)
	require.NoError(t, err)

	activeUser := mustCreateAPIKeyRepoUser(t, ctx, client, "byo-active@test.com")
	nonStripeUser := mustCreateAPIKeyRepoUser(t, ctx, client, "byo-nonstripe@test.com")
	expiredUser := mustCreateAPIKeyRepoUser(t, ctx, client, "byo-expired@test.com")

	now := time.Now()
	_, err = client.UserSubscription.Create().
		SetUserID(activeUser.ID).
		SetGroupID(group.ID).
		SetStartsAt(now.Add(-time.Hour)).
		SetExpiresAt(now.Add(time.Hour)).
		SetStatus(service.SubscriptionStatusActive).
		SetStripeSubscriptionID("sub_active").
		Save(ctx)
	require.NoError(t, err)
	_, err = client.UserSubscription.Create().
		SetUserID(nonStripeUser.ID).
		SetGroupID(group.ID).
		SetStartsAt(now.Add(-time.Hour)).
		SetExpiresAt(now.Add(time.Hour)).
		SetStatus(service.SubscriptionStatusActive).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.UserSubscription.Create().
		SetUserID(expiredUser.ID).
		SetGroupID(group.ID).
		SetStartsAt(now.Add(-2 * time.Hour)).
		SetExpiresAt(now.Add(-time.Hour)).
		SetStatus(service.SubscriptionStatusActive).
		SetStripeSubscriptionID("sub_expired").
		Save(ctx)
	require.NoError(t, err)

	active, err := repo.HasActiveBYOSubscription(ctx, activeUser.ID)
	require.NoError(t, err)
	require.True(t, active)
	active, err = repo.HasActiveBYOSubscription(ctx, nonStripeUser.ID)
	require.NoError(t, err)
	require.False(t, active)
	active, err = repo.HasActiveBYOSubscription(ctx, expiredUser.ID)
	require.NoError(t, err)
	require.False(t, active)
}
