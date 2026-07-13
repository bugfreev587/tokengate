package repository

import (
	"context"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestGroupEntityToService_PreservesMessagesDispatchModelConfig(t *testing.T) {
	group := &dbent.Group{
		ID:                    1,
		Name:                  "openai-dispatch",
		Platform:              service.PlatformOpenAI,
		Status:                service.StatusActive,
		SubscriptionType:      service.SubscriptionTypeStandard,
		RateMultiplier:        1,
		AllowMessagesDispatch: true,
		DefaultMappedModel:    "gpt-5.4",
		MessagesDispatchModelConfig: service.OpenAIMessagesDispatchModelConfig{
			OpusMappedModel:   "gpt-5.4-nano",
			SonnetMappedModel: "gpt-5.3-codex",
			HaikuMappedModel:  "gpt-5.4-mini",
			ExactModelMappings: map[string]string{
				"claude-sonnet-4.5": "gpt-5.4-nano",
			},
		},
	}

	got := groupEntityToService(group)
	require.NotNil(t, got)
	require.Equal(t, group.MessagesDispatchModelConfig, got.MessagesDispatchModelConfig)
}

func TestGroupEntityToService_PreservesOwnerAndCapacitySource(t *testing.T) {
	ownerID := int64(42)
	group := &dbent.Group{
		ID:                  2,
		Name:                "byo-openai",
		Platform:            service.PlatformOpenAI,
		Status:              service.StatusActive,
		SubscriptionType:    service.SubscriptionTypeStandard,
		RateMultiplier:      1,
		OwnerUserID:         &ownerID,
		CapacitySource:      service.CapacitySourceConnectedAccount,
		DefaultValidityDays: 30,
	}

	got := groupEntityToService(group)
	require.NotNil(t, got)
	require.NotNil(t, got.OwnerUserID)
	require.Equal(t, ownerID, *got.OwnerUserID)
	require.Equal(t, service.CapacitySourceConnectedAccount, got.CapacitySource)
	require.True(t, got.IsUserOwnedConnectedAccount())
}

func TestAPIKeyRepository_GetByKeyForAuth_PreservesMessagesDispatchModelConfig_SQLite(t *testing.T) {
	repo, client := newAPIKeyRepoSQLite(t)
	ctx := context.Background()
	user := mustCreateAPIKeyRepoUser(t, ctx, client, "getbykey-auth-dispatch-unit@test.com")

	group, err := client.Group.Create().
		SetName("g-auth-dispatch-unit").
		SetPlatform(service.PlatformOpenAI).
		SetStatus(service.StatusActive).
		SetSubscriptionType(service.SubscriptionTypeStandard).
		SetRateMultiplier(1).
		SetAllowMessagesDispatch(true).
		SetDefaultMappedModel("gpt-5.4").
		SetMessagesDispatchModelConfig(service.OpenAIMessagesDispatchModelConfig{
			OpusMappedModel:   "gpt-5.4-nano",
			SonnetMappedModel: "gpt-5.3-codex",
			HaikuMappedModel:  "gpt-5.4-mini",
			ExactModelMappings: map[string]string{
				"claude-sonnet-4.5": "gpt-5.4-nano",
			},
		}).
		Save(ctx)
	require.NoError(t, err)

	key := &service.APIKey{
		UserID:  user.ID,
		Key:     "sk-getbykey-auth-dispatch-unit",
		Name:    "Dispatch Key Unit",
		GroupID: &group.ID,
		Status:  service.StatusActive,
	}
	require.NoError(t, repo.Create(ctx, key))

	got, err := repo.GetByKeyForAuth(ctx, key.Key)
	require.NoError(t, err)
	require.Equal(t, key.Name, got.Name)
	require.NotNil(t, got.Group)
	require.Equal(t, group.MessagesDispatchModelConfig, got.Group.MessagesDispatchModelConfig)
}

func TestAPIKeyRepository_GetByKeyForAuth_PreservesBYOCapacityMetadata_SQLite(t *testing.T) {
	repo, client := newAPIKeyRepoSQLite(t)
	ctx := context.Background()
	user := mustCreateAPIKeyRepoUser(t, ctx, client, "getbykey-auth-byo-unit@test.com")

	group, err := client.Group.Create().
		SetName("g-auth-byo-unit").
		SetPlatform(service.PlatformOpenAI).
		SetStatus(service.StatusActive).
		SetSubscriptionType(service.SubscriptionTypeStandard).
		SetRateMultiplier(1).
		SetOwnerUserID(user.ID).
		SetCapacitySource(service.CapacitySourceConnectedAccount).
		Save(ctx)
	require.NoError(t, err)

	account, err := client.Account.Create().
		SetName("BYO OpenAI Unit").
		SetPlatform(service.PlatformOpenAI).
		SetType("oauth").
		SetStatus(service.StatusActive).
		SetSchedulable(false).
		SetExtra(map[string]any{service.BYOAccountDisabledReasonKey: service.BYOAccountDisabledReasonSubscriptionInactive}).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.AccountGroup.Create().SetAccountID(account.ID).SetGroupID(group.ID).Save(ctx)
	require.NoError(t, err)

	key := &service.APIKey{UserID: user.ID, Key: "sk-getbykey-auth-byo-unit", Name: "BYO Key Unit", GroupID: &group.ID, Status: service.StatusActive}
	require.NoError(t, repo.Create(ctx, key))

	got, err := repo.GetByKeyForAuth(ctx, key.Key)
	require.NoError(t, err)
	require.NotNil(t, got.User)
	require.NotNil(t, got.Group)
	require.NotNil(t, got.Group.OwnerUserID)
	require.Equal(t, user.ID, *got.Group.OwnerUserID)
	require.Equal(t, service.CapacitySourceConnectedAccount, got.Group.CapacitySource)
	require.True(t, service.IsUserOwnedConnectedAccountCapacity(got.User, got.Group))
	require.NotNil(t, got.Group.BYOEnabled)
	require.False(t, *got.Group.BYOEnabled)
	require.Equal(t, service.BYOAccountDisabledReasonSubscriptionInactive, got.Group.BYODisabledReason)
}
