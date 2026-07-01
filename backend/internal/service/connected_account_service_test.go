//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

func TestConnectedAccountServiceCreateOpenAIAccountCreatesPrivateGroup(t *testing.T) {
	ctx := context.Background()
	accountRepo := newConnectedAccountRepoFake()
	groupRepo := newConnectedGroupRepoFake()
	oauth := &connectedOpenAIOAuthFake{
		tokenInfo: &OpenAITokenInfo{
			AccessToken:  "access-secret",
			RefreshToken: "refresh-secret",
			ExpiresAt:    time.Now().Add(time.Hour).Unix(),
			Email:        "owner@example.com",
			ClientID:     "client-id",
		},
	}
	svc := NewConnectedAccountService(accountRepo, groupRepo, oauth)

	account, err := svc.CreateOpenAIAccountFromOAuth(ctx, CreateConnectedOpenAIAccountInput{
		UserID:    42,
		SessionID: "session",
		Code:      "code",
		State:     "state",
	})
	require.NoError(t, err)
	require.NotNil(t, account)
	require.Equal(t, "owner@example.com", account.Name)
	require.Equal(t, PlatformOpenAI, account.Platform)
	require.Equal(t, AccountTypeOAuth, account.Type)
	require.True(t, account.CredentialsEncrypted)
	require.NotNil(t, account.OwnerUserID)
	require.Equal(t, int64(42), *account.OwnerUserID)
	require.Equal(t, "access-secret", account.Credentials["access_token"])

	require.Len(t, groupRepo.groups, 1)
	var group *Group
	for _, stored := range groupRepo.groups {
		group = stored
	}
	require.NotNil(t, group)
	require.Equal(t, PlatformOpenAI, group.Platform)
	require.Equal(t, StatusActive, group.Status)
	require.Equal(t, SubscriptionTypeStandard, group.SubscriptionType)
	require.Equal(t, CapacitySourceConnectedAccount, group.CapacitySource)
	require.True(t, group.IsExclusive)
	require.NotNil(t, group.OwnerUserID)
	require.Equal(t, int64(42), *group.OwnerUserID)

	require.Equal(t, []int64{group.ID}, accountRepo.boundGroups[account.ID])
	require.Len(t, account.Groups, 1)
	require.Equal(t, group.ID, account.Groups[0].ID)
}

func TestConnectedAccountServiceListUsesOwnerScope(t *testing.T) {
	ctx := context.Background()
	accountRepo := newConnectedAccountRepoFake()
	svc := NewConnectedAccountService(accountRepo, newConnectedGroupRepoFake(), &connectedOpenAIOAuthFake{})
	ownerID := int64(42)
	otherOwnerID := int64(77)
	require.NoError(t, accountRepo.Create(ctx, &Account{Name: "mine", OwnerUserID: &ownerID, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Status: StatusActive}))
	require.NoError(t, accountRepo.Create(ctx, &Account{Name: "other", OwnerUserID: &otherOwnerID, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Status: StatusActive}))

	accounts, result, err := svc.List(ctx, ownerID, pagination.PaginationParams{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), result.Total)
	require.Len(t, accounts, 1)
	require.Equal(t, "mine", accounts[0].Name)
}

func TestConnectedAccountServiceDeleteRemovesOnlyOwnedConnectedGroup(t *testing.T) {
	ctx := context.Background()
	accountRepo := newConnectedAccountRepoFake()
	groupRepo := newConnectedGroupRepoFake()
	svc := NewConnectedAccountService(accountRepo, groupRepo, &connectedOpenAIOAuthFake{})
	ownerID := int64(42)
	group := &Group{
		ID:             10,
		Name:           "byo-openai-u42-a1",
		Platform:       PlatformOpenAI,
		Status:         StatusActive,
		OwnerUserID:    &ownerID,
		CapacitySource: CapacitySourceConnectedAccount,
	}
	groupRepo.groups[group.ID] = group
	account := &Account{
		ID:          1,
		Name:        "mine",
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Status:      StatusActive,
		OwnerUserID: &ownerID,
		Groups:      []*Group{group},
	}
	accountRepo.accounts[account.ID] = account

	require.NoError(t, svc.Delete(ctx, ownerID, account.ID))
	require.Contains(t, groupRepo.deletedGroupIDs, group.ID)
	require.Contains(t, accountRepo.deletedAccountIDs, account.ID)

	otherOwnerID := int64(77)
	otherAccount := &Account{ID: 2, OwnerUserID: &otherOwnerID, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	accountRepo.accounts[otherAccount.ID] = otherAccount
	err := svc.Delete(ctx, ownerID, otherAccount.ID)
	require.ErrorIs(t, err, ErrAccountNotFound)
}

type connectedOpenAIOAuthFake struct {
	tokenInfo     *OpenAITokenInfo
	refreshedInfo *OpenAITokenInfo
	authURLResult *OpenAIAuthURLResult
}

func (f *connectedOpenAIOAuthFake) GenerateAuthURL(context.Context, *int64, string, string) (*OpenAIAuthURLResult, error) {
	if f.authURLResult != nil {
		return f.authURLResult, nil
	}
	return &OpenAIAuthURLResult{AuthURL: "https://auth.example", SessionID: "session"}, nil
}

func (f *connectedOpenAIOAuthFake) ExchangeCode(context.Context, *OpenAIExchangeCodeInput) (*OpenAITokenInfo, error) {
	return f.tokenInfo, nil
}

func (f *connectedOpenAIOAuthFake) RefreshAccountToken(context.Context, *Account) (*OpenAITokenInfo, error) {
	return f.refreshedInfo, nil
}

func (f *connectedOpenAIOAuthFake) BuildAccountCredentials(tokenInfo *OpenAITokenInfo) map[string]any {
	return (&OpenAIOAuthService{}).BuildAccountCredentials(tokenInfo)
}

type connectedAccountRepoFake struct {
	nextID            int64
	accounts          map[int64]*Account
	boundGroups       map[int64][]int64
	deletedAccountIDs []int64
}

func newConnectedAccountRepoFake() *connectedAccountRepoFake {
	return &connectedAccountRepoFake{
		nextID:      1,
		accounts:    make(map[int64]*Account),
		boundGroups: make(map[int64][]int64),
	}
}

func (r *connectedAccountRepoFake) Create(_ context.Context, account *Account) error {
	account.ID = r.nextID
	r.nextID++
	cp := *account
	r.accounts[account.ID] = &cp
	return nil
}

func (r *connectedAccountRepoFake) Delete(_ context.Context, id int64) error {
	if _, ok := r.accounts[id]; !ok {
		return ErrAccountNotFound
	}
	r.deletedAccountIDs = append(r.deletedAccountIDs, id)
	delete(r.accounts, id)
	return nil
}

func (r *connectedAccountRepoFake) UpdateCredentials(_ context.Context, id int64, credentials map[string]any) error {
	account, ok := r.accounts[id]
	if !ok {
		return ErrAccountNotFound
	}
	account.Credentials = credentials
	return nil
}

func (r *connectedAccountRepoFake) GetByIDAndOwnerUserID(_ context.Context, id int64, ownerUserID int64) (*Account, error) {
	account, ok := r.accounts[id]
	if !ok || account.OwnerUserID == nil || *account.OwnerUserID != ownerUserID {
		return nil, ErrAccountNotFound
	}
	cp := *account
	return &cp, nil
}

func (r *connectedAccountRepoFake) ListByOwnerUserID(_ context.Context, ownerUserID int64, params pagination.PaginationParams) ([]Account, *pagination.PaginationResult, error) {
	out := make([]Account, 0)
	for _, account := range r.accounts {
		if account.OwnerUserID != nil && *account.OwnerUserID == ownerUserID {
			out = append(out, *account)
		}
	}
	return out, &pagination.PaginationResult{Total: int64(len(out)), Page: params.Page, PageSize: params.PageSize, Pages: 1}, nil
}

func (r *connectedAccountRepoFake) BindGroups(_ context.Context, accountID int64, groupIDs []int64) error {
	if _, ok := r.accounts[accountID]; !ok {
		return ErrAccountNotFound
	}
	r.boundGroups[accountID] = append([]int64(nil), groupIDs...)
	return nil
}

type connectedGroupRepoFake struct {
	nextID          int64
	groups          map[int64]*Group
	deletedGroupIDs []int64
}

func newConnectedGroupRepoFake() *connectedGroupRepoFake {
	return &connectedGroupRepoFake{
		nextID: 1,
		groups: make(map[int64]*Group),
	}
}

func (r *connectedGroupRepoFake) Create(_ context.Context, group *Group) error {
	if group.ID == 0 {
		group.ID = r.nextID
		r.nextID++
	}
	cp := *group
	r.groups[group.ID] = &cp
	return nil
}

func (r *connectedGroupRepoFake) DeleteCascade(_ context.Context, id int64) ([]int64, error) {
	if _, ok := r.groups[id]; !ok {
		return nil, ErrGroupNotFound
	}
	r.deletedGroupIDs = append(r.deletedGroupIDs, id)
	delete(r.groups, id)
	return nil, nil
}
