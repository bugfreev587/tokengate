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

type ConnectedAccountService struct {
	accountRepo ConnectedAccountRepository
	groupRepo   ConnectedGroupRepository
	openaiOAuth ConnectedOpenAIOAuthService
}

func NewConnectedAccountService(accountRepo ConnectedAccountRepository, groupRepo ConnectedGroupRepository, openaiOAuth ConnectedOpenAIOAuthService) *ConnectedAccountService {
	return &ConnectedAccountService{
		accountRepo: accountRepo,
		groupRepo:   groupRepo,
		openaiOAuth: openaiOAuth,
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

func (s *ConnectedAccountService) GenerateOpenAIAuthURL(ctx context.Context, userID int64, proxyID *int64, redirectURI string) (*OpenAIAuthURLResult, error) {
	if err := validateConnectedAccountUser(userID); err != nil {
		return nil, err
	}
	if s == nil || s.openaiOAuth == nil {
		return nil, ErrConnectedAccountUnsupported
	}
	return s.openaiOAuth.GenerateAuthURL(ctx, proxyID, redirectURI, PlatformOpenAI)
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
	name := strings.TrimSpace(input.Name)
	if name == "" && tokenInfo != nil {
		name = strings.TrimSpace(tokenInfo.Email)
	}
	if name == "" {
		name = "OpenAI OAuth Account"
	}
	concurrency := input.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}
	priority := input.Priority
	if priority <= 0 {
		priority = 50
	}

	ownerUserID := input.UserID
	account := &Account{
		Name:                 name,
		Platform:             PlatformOpenAI,
		Type:                 AccountTypeOAuth,
		OwnerUserID:          &ownerUserID,
		Credentials:          credentials,
		CredentialsEncrypted: true,
		ProxyID:              input.ProxyID,
		Concurrency:          concurrency,
		Priority:             priority,
		Status:               StatusActive,
		Schedulable:          true,
		AutoPauseOnExpired:   true,
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

func (s *ConnectedAccountService) List(ctx context.Context, userID int64, params pagination.PaginationParams) ([]Account, *pagination.PaginationResult, error) {
	if err := validateConnectedAccountUser(userID); err != nil {
		return nil, nil, err
	}
	if s == nil || s.accountRepo == nil {
		return nil, nil, ErrConnectedAccountUnsupported
	}
	return s.accountRepo.ListByOwnerUserID(ctx, userID, params)
}

func (s *ConnectedAccountService) RefreshOpenAIAccount(ctx context.Context, userID int64, accountID int64) (*Account, error) {
	if err := validateConnectedAccountUser(userID); err != nil {
		return nil, err
	}
	if s == nil || s.accountRepo == nil || s.openaiOAuth == nil {
		return nil, ErrConnectedAccountUnsupported
	}
	account, err := s.accountRepo.GetByIDAndOwnerUserID(ctx, accountID, userID)
	if err != nil {
		return nil, err
	}
	if account.Platform != PlatformOpenAI || account.Type != AccountTypeOAuth {
		return nil, ErrConnectedAccountUnsupported
	}
	tokenInfo, err := s.openaiOAuth.RefreshAccountToken(ctx, account)
	if err != nil {
		return nil, err
	}
	newCredentials := s.openaiOAuth.BuildAccountCredentials(tokenInfo)
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
