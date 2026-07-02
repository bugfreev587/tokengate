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
	svc := NewConnectedAccountService(accountRepo, groupRepo, oauth, nil, nil)

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

func TestConnectedAccountServiceOpenAIUsesCodexRedirectURI(t *testing.T) {
	ctx := context.Background()
	accountRepo := newConnectedAccountRepoFake()
	groupRepo := newConnectedGroupRepoFake()
	oauth := &connectedOpenAIOAuthFake{
		tokenInfo: &OpenAITokenInfo{
			AccessToken:  "access-secret",
			RefreshToken: "refresh-secret",
			ExpiresAt:    time.Now().Add(time.Hour).Unix(),
			Email:        "owner@example.com",
		},
	}
	svc := NewConnectedAccountService(accountRepo, groupRepo, oauth, nil, nil)

	_, err := svc.GenerateOpenAIAuthURL(ctx, 42, nil, "https://api.tokengate.to/accounts")
	require.NoError(t, err)
	require.Empty(t, oauth.generateRedirectURI)

	_, err = svc.CreateOpenAIAccountFromOAuth(ctx, CreateConnectedOpenAIAccountInput{
		UserID:      42,
		SessionID:   "session",
		Code:        "code",
		State:       "state",
		RedirectURI: "https://api.tokengate.to/accounts",
	})
	require.NoError(t, err)
	require.NotNil(t, oauth.exchangeInput)
	require.Empty(t, oauth.exchangeInput.RedirectURI)
}

func TestConnectedAccountServiceCreateAnthropicAccountCreatesPrivateGroup(t *testing.T) {
	ctx := context.Background()
	accountRepo := newConnectedAccountRepoFake()
	groupRepo := newConnectedGroupRepoFake()
	oauth := &connectedAnthropicOAuthFake{
		tokenInfo: &TokenInfo{
			AccessToken:  "anthropic-access",
			TokenType:    "Bearer",
			ExpiresIn:    3600,
			ExpiresAt:    time.Now().Add(time.Hour).Unix(),
			RefreshToken: "anthropic-refresh",
			Scope:        "org:create_api_key user:profile",
			OrgUUID:      "org-uuid",
			AccountUUID:  "account-uuid",
			EmailAddress: "claude-owner@example.com",
		},
	}
	svc := NewConnectedAccountService(accountRepo, groupRepo, nil, oauth, nil)

	account, err := svc.CreateAnthropicAccountFromOAuth(ctx, CreateConnectedAnthropicAccountInput{
		UserID:    42,
		SessionID: "session",
		Code:      "code",
	})
	require.NoError(t, err)
	require.NotNil(t, account)
	require.Equal(t, "claude-owner@example.com", account.Name)
	require.Equal(t, PlatformAnthropic, account.Platform)
	require.Equal(t, AccountTypeOAuth, account.Type)
	require.True(t, account.CredentialsEncrypted)
	require.Equal(t, "anthropic-access", account.Credentials["access_token"])
	require.Equal(t, "claude-owner@example.com", account.Extra["email_address"])
	require.Equal(t, "org-uuid", account.Extra["org_uuid"])
	require.Equal(t, "account-uuid", account.Extra["account_uuid"])

	require.Len(t, groupRepo.groups, 1)
	var group *Group
	for _, stored := range groupRepo.groups {
		group = stored
	}
	require.NotNil(t, group)
	require.Equal(t, PlatformAnthropic, group.Platform)
	require.Equal(t, CapacitySourceConnectedAccount, group.CapacitySource)
	require.NotNil(t, group.OwnerUserID)
	require.Equal(t, int64(42), *group.OwnerUserID)
	require.Equal(t, []int64{group.ID}, accountRepo.boundGroups[account.ID])
}

func TestConnectedAccountServiceCreateGeminiAccountCreatesPrivateGroup(t *testing.T) {
	ctx := context.Background()
	accountRepo := newConnectedAccountRepoFake()
	groupRepo := newConnectedGroupRepoFake()
	oauth := &connectedGeminiOAuthFake{
		tokenInfo: &GeminiTokenInfo{
			AccessToken:  "gemini-access",
			RefreshToken: "gemini-refresh",
			TokenType:    "Bearer",
			ExpiresIn:    3600,
			ExpiresAt:    time.Now().Add(time.Hour).Unix(),
			Scope:        "https://www.googleapis.com/auth/cloud-platform",
			ProjectID:    "project-123",
			OAuthType:    "google_one",
			TierID:       GeminiTierGoogleAIPro,
			Extra: map[string]any{
				"drive_storage_limit": int64(2 * TB),
			},
		},
	}
	svc := NewConnectedAccountService(accountRepo, groupRepo, nil, nil, oauth)

	account, err := svc.CreateGeminiAccountFromOAuth(ctx, CreateConnectedGeminiAccountInput{
		UserID:      42,
		SessionID:   "session",
		Code:        "code",
		State:       "state",
		OAuthType:   "google_one",
		TierID:      GeminiTierGoogleAIPro,
		Name:        "My Gemini",
		Concurrency: 3,
		Priority:    11,
	})
	require.NoError(t, err)
	require.NotNil(t, account)
	require.Equal(t, "My Gemini", account.Name)
	require.Equal(t, PlatformGemini, account.Platform)
	require.Equal(t, AccountTypeOAuth, account.Type)
	require.True(t, account.CredentialsEncrypted)
	require.Equal(t, 3, account.Concurrency)
	require.Equal(t, 11, account.Priority)
	require.Equal(t, "gemini-access", account.Credentials["access_token"])
	require.Equal(t, "project-123", account.Credentials["project_id"])
	require.Equal(t, "google_one", account.Credentials["oauth_type"])
	require.Equal(t, GeminiTierGoogleAIPro, account.Credentials["tier_id"])
	require.Equal(t, int64(2*TB), account.Extra["drive_storage_limit"])

	require.Len(t, groupRepo.groups, 1)
	var group *Group
	for _, stored := range groupRepo.groups {
		group = stored
	}
	require.NotNil(t, group)
	require.Equal(t, PlatformGemini, group.Platform)
	require.Equal(t, CapacitySourceConnectedAccount, group.CapacitySource)
	require.NotNil(t, group.OwnerUserID)
	require.Equal(t, int64(42), *group.OwnerUserID)
	require.Equal(t, []int64{group.ID}, accountRepo.boundGroups[account.ID])
}

func TestConnectedAccountServiceListUsesOwnerScope(t *testing.T) {
	ctx := context.Background()
	accountRepo := newConnectedAccountRepoFake()
	svc := NewConnectedAccountService(accountRepo, newConnectedGroupRepoFake(), &connectedOpenAIOAuthFake{}, nil, nil)
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
	svc := NewConnectedAccountService(accountRepo, groupRepo, &connectedOpenAIOAuthFake{}, nil, nil)
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
	tokenInfo           *OpenAITokenInfo
	refreshedInfo       *OpenAITokenInfo
	authURLResult       *OpenAIAuthURLResult
	generateRedirectURI string
	exchangeInput       *OpenAIExchangeCodeInput
}

type connectedAnthropicOAuthFake struct {
	tokenInfo     *TokenInfo
	refreshedInfo *TokenInfo
	authURLResult *GenerateAuthURLResult
}

func (f *connectedAnthropicOAuthFake) GenerateAuthURL(context.Context, *int64) (*GenerateAuthURLResult, error) {
	if f.authURLResult != nil {
		return f.authURLResult, nil
	}
	return &GenerateAuthURLResult{AuthURL: "https://claude.example/auth", SessionID: "session"}, nil
}

func (f *connectedAnthropicOAuthFake) ExchangeCode(context.Context, *ExchangeCodeInput) (*TokenInfo, error) {
	return f.tokenInfo, nil
}

func (f *connectedAnthropicOAuthFake) RefreshAccountToken(context.Context, *Account) (*TokenInfo, error) {
	return f.refreshedInfo, nil
}

type connectedGeminiOAuthFake struct {
	tokenInfo     *GeminiTokenInfo
	refreshedInfo *GeminiTokenInfo
	authURLResult *GeminiAuthURLResult
}

func (f *connectedGeminiOAuthFake) GenerateAuthURL(context.Context, *int64, string, string, string, string) (*GeminiAuthURLResult, error) {
	if f.authURLResult != nil {
		return f.authURLResult, nil
	}
	return &GeminiAuthURLResult{AuthURL: "https://gemini.example/auth", SessionID: "session", State: "state"}, nil
}

func (f *connectedGeminiOAuthFake) ExchangeCode(context.Context, *GeminiExchangeCodeInput) (*GeminiTokenInfo, error) {
	return f.tokenInfo, nil
}

func (f *connectedGeminiOAuthFake) RefreshAccountToken(context.Context, *Account) (*GeminiTokenInfo, error) {
	return f.refreshedInfo, nil
}

func (f *connectedGeminiOAuthFake) BuildAccountCredentials(tokenInfo *GeminiTokenInfo) map[string]any {
	return (&GeminiOAuthService{}).BuildAccountCredentials(tokenInfo)
}

func (f *connectedOpenAIOAuthFake) GenerateAuthURL(_ context.Context, _ *int64, redirectURI string, _ string) (*OpenAIAuthURLResult, error) {
	f.generateRedirectURI = redirectURI
	if f.authURLResult != nil {
		return f.authURLResult, nil
	}
	return &OpenAIAuthURLResult{AuthURL: "https://auth.example", SessionID: "session"}, nil
}

func (f *connectedOpenAIOAuthFake) ExchangeCode(_ context.Context, input *OpenAIExchangeCodeInput) (*OpenAITokenInfo, error) {
	if input != nil {
		cp := *input
		f.exchangeInput = &cp
	}
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
