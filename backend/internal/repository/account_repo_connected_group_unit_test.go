//go:build unit

package repository

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAccountRepositoryListByGroupFiltersConnectedGroupOwner(t *testing.T) {
	ctx := context.Background()
	client := newAccountCredentialSQLiteClient(t)
	accountRepo := newAccountRepositoryWithSQL(client, nil, nil)
	groupRepo := newGroupRepositoryWithSQL(client, nil)
	ownerID := int64(42)
	otherOwnerID := int64(77)

	group := &service.Group{
		Name:             "byo-openai-u42-a1",
		Platform:         service.PlatformOpenAI,
		RateMultiplier:   1,
		IsExclusive:      true,
		Status:           service.StatusActive,
		OwnerUserID:      &ownerID,
		CapacitySource:   service.CapacitySourceConnectedAccount,
		SubscriptionType: service.SubscriptionTypeStandard,
	}
	require.NoError(t, groupRepo.Create(ctx, group))

	ownedAccount := &service.Account{
		Name:               "owned",
		Platform:           service.PlatformOpenAI,
		Type:               service.AccountTypeOAuth,
		OwnerUserID:        &ownerID,
		Credentials:        map[string]any{"access_token": "owned"},
		Concurrency:        1,
		Priority:           50,
		Status:             service.StatusActive,
		Schedulable:        true,
		AutoPauseOnExpired: true,
	}
	otherAccount := &service.Account{
		Name:               "other",
		Platform:           service.PlatformOpenAI,
		Type:               service.AccountTypeOAuth,
		OwnerUserID:        &otherOwnerID,
		Credentials:        map[string]any{"access_token": "other"},
		Concurrency:        1,
		Priority:           50,
		Status:             service.StatusActive,
		Schedulable:        true,
		AutoPauseOnExpired: true,
	}
	require.NoError(t, accountRepo.Create(ctx, ownedAccount))
	require.NoError(t, accountRepo.Create(ctx, otherAccount))
	require.NoError(t, accountRepo.BindGroups(ctx, ownedAccount.ID, []int64{group.ID}))
	require.NoError(t, accountRepo.BindGroups(ctx, otherAccount.ID, []int64{group.ID}))

	got, err := accountRepo.ListByGroup(ctx, group.ID)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, ownedAccount.ID, got[0].ID)
	require.NotNil(t, got[0].OwnerUserID)
	require.Equal(t, ownerID, *got[0].OwnerUserID)
}
