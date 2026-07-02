package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

var (
	ErrConnectedAccountUserRequired = infraerrors.Unauthorized("CONNECTED_ACCOUNT_USER_REQUIRED", "user authentication is required")
	ErrConnectedAccountUnsupported  = infraerrors.New(http.StatusBadRequest, "CONNECTED_ACCOUNT_UNSUPPORTED", "connected account provider is not supported")
)

type ConnectedAccountRepository interface {
	Create(ctx context.Context, account *Account) error
	Delete(ctx context.Context, id int64) error
	UpdateCredentials(ctx context.Context, id int64, credentials map[string]any) error
	GetByIDAndOwnerUserID(ctx context.Context, id int64, ownerUserID int64) (*Account, error)
	ListByOwnerUserID(ctx context.Context, ownerUserID int64, params pagination.PaginationParams) ([]Account, *pagination.PaginationResult, error)
	BindGroups(ctx context.Context, accountID int64, groupIDs []int64) error
}

type ConnectedGroupRepository interface {
	Create(ctx context.Context, group *Group) error
	DeleteCascade(ctx context.Context, id int64) ([]int64, error)
}

type ConnectedOpenAIOAuthService interface {
	GenerateAuthURL(ctx context.Context, proxyID *int64, redirectURI, platform string) (*OpenAIAuthURLResult, error)
	ExchangeCode(ctx context.Context, input *OpenAIExchangeCodeInput) (*OpenAITokenInfo, error)
	RefreshAccountToken(ctx context.Context, account *Account) (*OpenAITokenInfo, error)
	BuildAccountCredentials(tokenInfo *OpenAITokenInfo) map[string]any
}

type ConnectedAnthropicOAuthService interface {
	GenerateAuthURL(ctx context.Context, proxyID *int64) (*GenerateAuthURLResult, error)
	ExchangeCode(ctx context.Context, input *ExchangeCodeInput) (*TokenInfo, error)
	RefreshAccountToken(ctx context.Context, account *Account) (*TokenInfo, error)
}

type ConnectedGeminiOAuthService interface {
	GenerateAuthURL(ctx context.Context, proxyID *int64, redirectURI, projectID, oauthType, tierID string) (*GeminiAuthURLResult, error)
	ExchangeCode(ctx context.Context, input *GeminiExchangeCodeInput) (*GeminiTokenInfo, error)
	RefreshAccountToken(ctx context.Context, account *Account) (*GeminiTokenInfo, error)
	BuildAccountCredentials(tokenInfo *GeminiTokenInfo) map[string]any
}

type ConnectedAccountService struct {
	accountRepo    ConnectedAccountRepository
	groupRepo      ConnectedGroupRepository
	openaiOAuth    ConnectedOpenAIOAuthService
	anthropicOAuth ConnectedAnthropicOAuthService
	geminiOAuth    ConnectedGeminiOAuthService
}

func NewConnectedAccountService(
	accountRepo ConnectedAccountRepository,
	groupRepo ConnectedGroupRepository,
	openaiOAuth ConnectedOpenAIOAuthService,
	anthropicOAuth ConnectedAnthropicOAuthService,
	geminiOAuth ConnectedGeminiOAuthService,
) *ConnectedAccountService {
	return &ConnectedAccountService{
		accountRepo:    accountRepo,
		groupRepo:      groupRepo,
		openaiOAuth:    openaiOAuth,
		anthropicOAuth: anthropicOAuth,
		geminiOAuth:    geminiOAuth,
	}
}

type CreateConnectedOpenAIAccountInput struct {
	UserID      int64
	SessionID   string
	Code        string
	State       string
	RedirectURI string
	ProxyID     *int64
	Name        string
	Concurrency int
	Priority    int
}

type CreateConnectedAnthropicAccountInput struct {
	UserID      int64
	SessionID   string
	Code        string
	ProxyID     *int64
	Name        string
	Concurrency int
	Priority    int
}

type CreateConnectedGeminiAccountInput struct {
	UserID      int64
	SessionID   string
	Code        string
	State       string
	ProxyID     *int64
	Name        string
	Concurrency int
	Priority    int
	ProjectID   string
	OAuthType   string
	TierID      string
}

func (s *ConnectedAccountService) GenerateOpenAIAuthURL(ctx context.Context, userID int64, proxyID *int64, redirectURI string) (*OpenAIAuthURLResult, error) {
	if err := validateConnectedAccountUser(userID); err != nil {
		return nil, err
	}
	if s == nil || s.openaiOAuth == nil {
		return nil, ErrConnectedAccountUnsupported
	}
	return s.openaiOAuth.GenerateAuthURL(ctx, proxyID, redirectURI, PlatformOpenAI)
}

func (s *ConnectedAccountService) GenerateAnthropicAuthURL(ctx context.Context, userID int64, proxyID *int64) (*GenerateAuthURLResult, error) {
	if err := validateConnectedAccountUser(userID); err != nil {
		return nil, err
	}
	if s == nil || s.anthropicOAuth == nil {
		return nil, ErrConnectedAccountUnsupported
	}
	return s.anthropicOAuth.GenerateAuthURL(ctx, proxyID)
}

func (s *ConnectedAccountService) GenerateGeminiAuthURL(ctx context.Context, userID int64, proxyID *int64, redirectURI, projectID, oauthType, tierID string) (*GeminiAuthURLResult, error) {
	if err := validateConnectedAccountUser(userID); err != nil {
		return nil, err
	}
	if s == nil || s.geminiOAuth == nil {
		return nil, ErrConnectedAccountUnsupported
	}
	oauthType = normalizeConnectedGeminiOAuthType(oauthType)
	return s.geminiOAuth.GenerateAuthURL(ctx, proxyID, redirectURI, strings.TrimSpace(projectID), oauthType, strings.TrimSpace(tierID))
}

func (s *ConnectedAccountService) CreateOpenAIAccountFromOAuth(ctx context.Context, input CreateConnectedOpenAIAccountInput) (*Account, error) {
	if err := validateConnectedAccountUser(input.UserID); err != nil {
		return nil, err
	}
	if s == nil || s.accountRepo == nil || s.groupRepo == nil || s.openaiOAuth == nil {
		return nil, ErrConnectedAccountUnsupported
	}

	tokenInfo, err := s.openaiOAuth.ExchangeCode(ctx, &OpenAIExchangeCodeInput{
		SessionID:   input.SessionID,
		Code:        input.Code,
		State:       input.State,
		RedirectURI: input.RedirectURI,
		ProxyID:     input.ProxyID,
	})
	if err != nil {
		return nil, err
	}

	credentials := s.openaiOAuth.BuildAccountCredentials(tokenInfo)
	extra := map[string]any{}
	if tokenInfo != nil && strings.TrimSpace(tokenInfo.PrivacyMode) != "" {
		extra["privacy_mode"] = strings.TrimSpace(tokenInfo.PrivacyMode)
	}
	name := strings.TrimSpace(input.Name)
	if name == "" && tokenInfo != nil {
		name = strings.TrimSpace(tokenInfo.Email)
	}
	if name == "" {
		name = "OpenAI OAuth Account"
	}

	return s.createConnectedAccount(ctx, createConnectedAccountParams{
		UserID:      input.UserID,
		Name:        name,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Credentials: credentials,
		Extra:       extra,
		ProxyID:     input.ProxyID,
		Concurrency: input.Concurrency,
		Priority:    input.Priority,
	})
}

func (s *ConnectedAccountService) CreateAnthropicAccountFromOAuth(ctx context.Context, input CreateConnectedAnthropicAccountInput) (*Account, error) {
	if err := validateConnectedAccountUser(input.UserID); err != nil {
		return nil, err
	}
	if s == nil || s.accountRepo == nil || s.groupRepo == nil || s.anthropicOAuth == nil {
		return nil, ErrConnectedAccountUnsupported
	}

	tokenInfo, err := s.anthropicOAuth.ExchangeCode(ctx, &ExchangeCodeInput{
		SessionID: input.SessionID,
		Code:      input.Code,
		ProxyID:   input.ProxyID,
	})
	if err != nil {
		return nil, err
	}

	credentials := BuildClaudeAccountCredentials(tokenInfo)
	extra := buildAnthropicConnectedAccountExtra(tokenInfo)
	name := strings.TrimSpace(input.Name)
	if name == "" && tokenInfo != nil {
		name = strings.TrimSpace(tokenInfo.EmailAddress)
	}
	if name == "" {
		name = "Anthropic OAuth Account"
	}

	return s.createConnectedAccount(ctx, createConnectedAccountParams{
		UserID:      input.UserID,
		Name:        name,
		Platform:    PlatformAnthropic,
		Type:        AccountTypeOAuth,
		Credentials: credentials,
		Extra:       extra,
		ProxyID:     input.ProxyID,
		Concurrency: input.Concurrency,
		Priority:    input.Priority,
	})
}

func (s *ConnectedAccountService) CreateGeminiAccountFromOAuth(ctx context.Context, input CreateConnectedGeminiAccountInput) (*Account, error) {
	if err := validateConnectedAccountUser(input.UserID); err != nil {
		return nil, err
	}
	if s == nil || s.accountRepo == nil || s.groupRepo == nil || s.geminiOAuth == nil {
		return nil, ErrConnectedAccountUnsupported
	}

	oauthType := normalizeConnectedGeminiOAuthType(input.OAuthType)
	tokenInfo, err := s.geminiOAuth.ExchangeCode(ctx, &GeminiExchangeCodeInput{
		SessionID: input.SessionID,
		State:     input.State,
		Code:      input.Code,
		ProxyID:   input.ProxyID,
		OAuthType: oauthType,
		TierID:    strings.TrimSpace(input.TierID),
	})
	if err != nil {
		return nil, err
	}

	credentials := s.geminiOAuth.BuildAccountCredentials(tokenInfo)
	extra := map[string]any{}
	if tokenInfo != nil && len(tokenInfo.Extra) > 0 {
		for k, v := range tokenInfo.Extra {
			extra[k] = v
		}
	}
	name := strings.TrimSpace(input.Name)
	if name == "" && tokenInfo != nil && strings.TrimSpace(tokenInfo.ProjectID) != "" {
		name = strings.TrimSpace(tokenInfo.ProjectID)
	}
	if name == "" {
		name = "Gemini OAuth Account"
	}

	return s.createConnectedAccount(ctx, createConnectedAccountParams{
		UserID:      input.UserID,
		Name:        name,
		Platform:    PlatformGemini,
		Type:        AccountTypeOAuth,
		Credentials: credentials,
		Extra:       extra,
		ProxyID:     input.ProxyID,
		Concurrency: input.Concurrency,
		Priority:    input.Priority,
	})
}

func (s *ConnectedAccountService) List(ctx context.Context, userID int64, params pagination.PaginationParams) ([]Account, *pagination.PaginationResult, error) {
	if err := validateConnectedAccountUser(userID); err != nil {
		return nil, nil, err
	}
	if s == nil || s.accountRepo == nil {
		return nil, nil, ErrConnectedAccountUnsupported
	}
	return s.accountRepo.ListByOwnerUserID(ctx, userID, params)
}

func (s *ConnectedAccountService) GetOwnedAccount(ctx context.Context, userID int64, accountID int64) (*Account, error) {
	if err := validateConnectedAccountUser(userID); err != nil {
		return nil, err
	}
	if s == nil || s.accountRepo == nil {
		return nil, ErrConnectedAccountUnsupported
	}
	return s.accountRepo.GetByIDAndOwnerUserID(ctx, accountID, userID)
}

func (s *ConnectedAccountService) RefreshAccount(ctx context.Context, userID int64, accountID int64) (*Account, error) {
	if err := validateConnectedAccountUser(userID); err != nil {
		return nil, err
	}
	if s == nil || s.accountRepo == nil {
		return nil, ErrConnectedAccountUnsupported
	}
	account, err := s.accountRepo.GetByIDAndOwnerUserID(ctx, accountID, userID)
	if err != nil {
		return nil, err
	}
	if account.Type != AccountTypeOAuth {
		return nil, ErrConnectedAccountUnsupported
	}

	var newCredentials map[string]any
	switch account.Platform {
	case PlatformOpenAI:
		if s.openaiOAuth == nil {
			return nil, ErrConnectedAccountUnsupported
		}
		tokenInfo, err := s.openaiOAuth.RefreshAccountToken(ctx, account)
		if err != nil {
			return nil, err
		}
		newCredentials = s.openaiOAuth.BuildAccountCredentials(tokenInfo)
	case PlatformAnthropic:
		if s.anthropicOAuth == nil {
			return nil, ErrConnectedAccountUnsupported
		}
		tokenInfo, err := s.anthropicOAuth.RefreshAccountToken(ctx, account)
		if err != nil {
			return nil, err
		}
		newCredentials = BuildClaudeAccountCredentials(tokenInfo)
	case PlatformGemini:
		if s.geminiOAuth == nil {
			return nil, ErrConnectedAccountUnsupported
		}
		tokenInfo, err := s.geminiOAuth.RefreshAccountToken(ctx, account)
		if err != nil {
			return nil, err
		}
		newCredentials = s.geminiOAuth.BuildAccountCredentials(tokenInfo)
	default:
		return nil, ErrConnectedAccountUnsupported
	}

	for k, v := range account.Credentials {
		if _, ok := newCredentials[k]; !ok {
			newCredentials[k] = v
		}
	}
	if err := s.accountRepo.UpdateCredentials(ctx, account.ID, newCredentials); err != nil {
		return nil, err
	}
	account.Credentials = newCredentials
	account.CredentialsEncrypted = true
	return account, nil
}

func (s *ConnectedAccountService) RefreshOpenAIAccount(ctx context.Context, userID int64, accountID int64) (*Account, error) {
	return s.RefreshAccount(ctx, userID, accountID)
}

func (s *ConnectedAccountService) Delete(ctx context.Context, userID int64, accountID int64) error {
	if err := validateConnectedAccountUser(userID); err != nil {
		return err
	}
	if s == nil || s.accountRepo == nil || s.groupRepo == nil {
		return ErrConnectedAccountUnsupported
	}
	account, err := s.accountRepo.GetByIDAndOwnerUserID(ctx, accountID, userID)
	if err != nil {
		return err
	}
	for _, group := range account.Groups {
		if isOwnedConnectedAccountGroup(group, userID) {
			if _, err := s.groupRepo.DeleteCascade(ctx, group.ID); err != nil {
				return err
			}
		}
	}
	return s.accountRepo.Delete(ctx, account.ID)
}

func validateConnectedAccountUser(userID int64) error {
	if userID <= 0 {
		return ErrConnectedAccountUserRequired
	}
	return nil
}

type createConnectedAccountParams struct {
	UserID      int64
	Name        string
	Platform    string
	Type        string
	Credentials map[string]any
	Extra       map[string]any
	ProxyID     *int64
	Concurrency int
	Priority    int
}

func (s *ConnectedAccountService) createConnectedAccount(ctx context.Context, params createConnectedAccountParams) (*Account, error) {
	if s == nil || s.accountRepo == nil || s.groupRepo == nil {
		return nil, ErrConnectedAccountUnsupported
	}
	name := strings.TrimSpace(params.Name)
	if name == "" {
		name = defaultConnectedAccountName(params.Platform)
	}
	concurrency := params.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}
	priority := params.Priority
	if priority <= 0 {
		priority = 50
	}

	ownerUserID := params.UserID
	account := &Account{
		Name:                 name,
		Platform:             params.Platform,
		Type:                 params.Type,
		OwnerUserID:          &ownerUserID,
		Credentials:          params.Credentials,
		CredentialsEncrypted: true,
		Extra:                params.Extra,
		ProxyID:              params.ProxyID,
		Concurrency:          concurrency,
		Priority:             priority,
		Status:               StatusActive,
		Schedulable:          true,
		AutoPauseOnExpired:   true,
	}
	if account.Credentials == nil {
		account.Credentials = map[string]any{}
	}
	if account.Extra == nil {
		account.Extra = map[string]any{}
	}
	if err := s.accountRepo.Create(ctx, account); err != nil {
		return nil, err
	}

	group := newConnectedAccountGroup(ownerUserID, account)
	if err := s.groupRepo.Create(ctx, group); err != nil {
		_ = s.accountRepo.Delete(ctx, account.ID)
		return nil, err
	}
	if err := s.accountRepo.BindGroups(ctx, account.ID, []int64{group.ID}); err != nil {
		_, _ = s.groupRepo.DeleteCascade(ctx, group.ID)
		_ = s.accountRepo.Delete(ctx, account.ID)
		return nil, err
	}
	account.GroupIDs = []int64{group.ID}
	account.Groups = []*Group{group}
	return account, nil
}

func defaultConnectedAccountName(platform string) string {
	switch platform {
	case PlatformAnthropic:
		return "Anthropic OAuth Account"
	case PlatformGemini:
		return "Gemini OAuth Account"
	case PlatformOpenAI:
		return "OpenAI OAuth Account"
	default:
		return "Connected OAuth Account"
	}
}

func buildAnthropicConnectedAccountExtra(tokenInfo *TokenInfo) map[string]any {
	extra := map[string]any{}
	if tokenInfo == nil {
		return extra
	}
	if strings.TrimSpace(tokenInfo.OrgUUID) != "" {
		extra["org_uuid"] = strings.TrimSpace(tokenInfo.OrgUUID)
	}
	if strings.TrimSpace(tokenInfo.AccountUUID) != "" {
		extra["account_uuid"] = strings.TrimSpace(tokenInfo.AccountUUID)
	}
	if strings.TrimSpace(tokenInfo.EmailAddress) != "" {
		extra["email_address"] = strings.TrimSpace(tokenInfo.EmailAddress)
	}
	return extra
}

func normalizeConnectedGeminiOAuthType(oauthType string) string {
	oauthType = strings.TrimSpace(oauthType)
	if oauthType == "" {
		return "google_one"
	}
	return oauthType
}

func newConnectedAccountGroup(ownerUserID int64, account *Account) *Group {
	platform := PlatformOpenAI
	accountID := int64(0)
	if account != nil {
		if strings.TrimSpace(account.Platform) != "" {
			platform = strings.TrimSpace(account.Platform)
		}
		accountID = account.ID
	}
	return &Group{
		Name:             fmt.Sprintf("byo-%s-u%d-a%d", platform, ownerUserID, accountID),
		Description:      "User-owned connected account capacity",
		Platform:         platform,
		RateMultiplier:   1,
		IsExclusive:      true,
		Status:           StatusActive,
		OwnerUserID:      &ownerUserID,
		CapacitySource:   CapacitySourceConnectedAccount,
		SubscriptionType: SubscriptionTypeStandard,
	}
}

func isOwnedConnectedAccountGroup(group *Group, ownerUserID int64) bool {
	if group == nil || group.OwnerUserID == nil {
		return false
	}
	return *group.OwnerUserID == ownerUserID && group.IsUserOwnedConnectedAccount()
}
