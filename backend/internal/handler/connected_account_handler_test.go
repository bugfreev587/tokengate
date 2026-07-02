//go:build unit

package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestConnectedAccountHandlerListRedactsCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ownerID := int64(42)
	accountRepo := newHandlerConnectedAccountRepoFake()
	groupRepo := newHandlerConnectedGroupRepoFake()
	group := &service.Group{
		ID:             9,
		Name:           "byo-openai-u42-a1",
		Platform:       service.PlatformOpenAI,
		OwnerUserID:    &ownerID,
		CapacitySource: service.CapacitySourceConnectedAccount,
	}
	accountRepo.accounts[1] = &service.Account{
		ID:          1,
		Name:        "owner@example.com",
		Platform:    service.PlatformOpenAI,
		Type:        service.AccountTypeOAuth,
		Status:      service.StatusActive,
		OwnerUserID: &ownerID,
		Credentials: map[string]any{
			"access_token":  "access-secret",
			"refresh_token": "refresh-secret",
			"email":         "owner@example.com",
		},
		Groups:    []*service.Group{group},
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	svc := service.NewConnectedAccountService(accountRepo, groupRepo, &handlerConnectedOpenAIOAuthFake{}, nil, nil)
	h := NewConnectedAccountHandler(svc, nil, nil)
	router := gin.New()
	router.GET("/api/v1/user/accounts", func(c *gin.Context) {
		c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: ownerID})
		h.List(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/accounts?page=1&page_size=20", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var envelope response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &envelope))
	data, ok := envelope.Data.(map[string]any)
	require.True(t, ok)
	items, ok := data["items"].([]any)
	require.True(t, ok)
	require.Len(t, items, 1)
	item, ok := items[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, float64(1), item["id"])
	require.Equal(t, "owner@example.com", item["email"])
	require.NotContains(t, item, "credentials")
	raw := w.Body.String()
	require.NotContains(t, raw, "access-secret")
	require.NotContains(t, raw, "refresh-secret")
}

func TestConnectedAccountHandlerExchangeOpenAICodeCreatesAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ownerID := int64(42)
	accountRepo := newHandlerConnectedAccountRepoFake()
	groupRepo := newHandlerConnectedGroupRepoFake()
	oauth := &handlerConnectedOpenAIOAuthFake{
		tokenInfo: &service.OpenAITokenInfo{
			AccessToken:  "access-secret",
			RefreshToken: "refresh-secret",
			ExpiresAt:    time.Now().Add(time.Hour).Unix(),
			Email:        "owner@example.com",
		},
	}
	svc := service.NewConnectedAccountService(accountRepo, groupRepo, oauth, nil, nil)
	h := NewConnectedAccountHandler(svc, nil, nil)
	router := gin.New()
	router.POST("/api/v1/user/accounts/openai/exchange-code", func(c *gin.Context) {
		c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: ownerID})
		h.ExchangeOpenAICode(c)
	})

	body := []byte(`{"session_id":"session","code":"code","state":"state"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/accounts/openai/exchange-code", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var envelope response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &envelope))
	item, ok := envelope.Data.(map[string]any)
	require.True(t, ok)
	require.Equal(t, "owner@example.com", item["name"])
	require.Equal(t, "owner@example.com", item["email"])
	require.Equal(t, service.CapacitySourceConnectedAccount, item["capacity_source"])
	require.Equal(t, float64(1), item["group_id"])
	require.NotContains(t, w.Body.String(), "access-secret")
	require.Len(t, accountRepo.accounts, 1)
	require.Len(t, groupRepo.groups, 1)
}

func TestConnectedAccountHandlerExchangeAnthropicCodeCreatesAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ownerID := int64(42)
	accountRepo := newHandlerConnectedAccountRepoFake()
	groupRepo := newHandlerConnectedGroupRepoFake()
	oauth := &handlerConnectedAnthropicOAuthFake{
		tokenInfo: &service.TokenInfo{
			AccessToken:  "anthropic-access",
			RefreshToken: "anthropic-refresh",
			ExpiresAt:    time.Now().Add(time.Hour).Unix(),
			EmailAddress: "claude-owner@example.com",
		},
	}
	svc := service.NewConnectedAccountService(accountRepo, groupRepo, nil, oauth, nil)
	h := NewConnectedAccountHandler(svc, nil, nil)
	router := gin.New()
	router.POST("/api/v1/user/accounts/anthropic/exchange-code", func(c *gin.Context) {
		c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: ownerID})
		h.ExchangeAnthropicCode(c)
	})

	body := []byte(`{"session_id":"session","code":"code","name":"Claude Main"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/accounts/anthropic/exchange-code", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var envelope response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &envelope))
	item, ok := envelope.Data.(map[string]any)
	require.True(t, ok)
	require.Equal(t, "Claude Main", item["name"])
	require.Equal(t, service.PlatformAnthropic, item["platform"])
	require.Equal(t, "claude-owner@example.com", item["email"])
	require.Equal(t, float64(1), item["group_id"])
	require.NotContains(t, w.Body.String(), "anthropic-access")
	require.NotContains(t, w.Body.String(), "anthropic-refresh")
	require.Len(t, accountRepo.accounts, 1)
	require.Len(t, groupRepo.groups, 1)
}

func TestConnectedAccountHandlerExchangeGeminiCodeCreatesAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ownerID := int64(42)
	accountRepo := newHandlerConnectedAccountRepoFake()
	groupRepo := newHandlerConnectedGroupRepoFake()
	oauth := &handlerConnectedGeminiOAuthFake{
		tokenInfo: &service.GeminiTokenInfo{
			AccessToken:  "gemini-access",
			RefreshToken: "gemini-refresh",
			TokenType:    "Bearer",
			ExpiresAt:    time.Now().Add(time.Hour).Unix(),
			ProjectID:    "project-123",
			OAuthType:    "google_one",
			TierID:       service.GeminiTierGoogleAIPro,
		},
	}
	svc := service.NewConnectedAccountService(accountRepo, groupRepo, nil, nil, oauth)
	h := NewConnectedAccountHandler(svc, nil, nil)
	router := gin.New()
	router.POST("/api/v1/user/accounts/gemini/exchange-code", func(c *gin.Context) {
		c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: ownerID})
		h.ExchangeGeminiCode(c)
	})

	body := []byte(`{"session_id":"session","code":"code","state":"state","oauth_type":"google_one","name":"Gemini Main"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/accounts/gemini/exchange-code", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var envelope response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &envelope))
	item, ok := envelope.Data.(map[string]any)
	require.True(t, ok)
	require.Equal(t, "Gemini Main", item["name"])
	require.Equal(t, service.PlatformGemini, item["platform"])
	require.Equal(t, float64(1), item["group_id"])
	require.NotContains(t, w.Body.String(), "gemini-access")
	require.NotContains(t, w.Body.String(), "gemini-refresh")
	require.Len(t, accountRepo.accounts, 1)
	require.Len(t, groupRepo.groups, 1)
}

func TestConnectedAccountHandlerGetAvailableModelsRequiresOwnership(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ownerID := int64(42)
	otherOwnerID := int64(99)
	accountRepo := newHandlerConnectedAccountRepoFake()
	groupRepo := newHandlerConnectedGroupRepoFake()
	accountRepo.accounts[12] = &service.Account{
		ID:          12,
		Name:        "Other account",
		Platform:    service.PlatformOpenAI,
		Type:        service.AccountTypeOAuth,
		Status:      service.StatusActive,
		OwnerUserID: &otherOwnerID,
	}
	svc := service.NewConnectedAccountService(accountRepo, groupRepo, &handlerConnectedOpenAIOAuthFake{}, nil, nil)
	h := NewConnectedAccountHandler(svc, nil, nil)
	router := gin.New()
	router.GET("/api/v1/user/accounts/:id/models", func(c *gin.Context) {
		c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: ownerID})
		h.GetAvailableModels(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/accounts/12/models", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestConnectedAccountHandlerTestRequiresOwnership(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ownerID := int64(42)
	otherOwnerID := int64(99)
	accountRepo := newHandlerConnectedAccountRepoFake()
	groupRepo := newHandlerConnectedGroupRepoFake()
	accountRepo.accounts[12] = &service.Account{
		ID:          12,
		Name:        "Other account",
		Platform:    service.PlatformOpenAI,
		Type:        service.AccountTypeOAuth,
		Status:      service.StatusActive,
		OwnerUserID: &otherOwnerID,
	}
	svc := service.NewConnectedAccountService(accountRepo, groupRepo, &handlerConnectedOpenAIOAuthFake{}, nil, nil)
	h := NewConnectedAccountHandler(svc, nil, nil)
	router := gin.New()
	router.POST("/api/v1/user/accounts/:id/test", func(c *gin.Context) {
		c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: ownerID})
		h.Test(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/accounts/12/test", bytes.NewReader([]byte(`{"model_id":"gpt-5.4"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

type handlerConnectedAccountRepoFake struct {
	nextID      int64
	accounts    map[int64]*service.Account
	boundGroups map[int64][]int64
}

func newHandlerConnectedAccountRepoFake() *handlerConnectedAccountRepoFake {
	return &handlerConnectedAccountRepoFake{
		nextID:      1,
		accounts:    make(map[int64]*service.Account),
		boundGroups: make(map[int64][]int64),
	}
}

func (r *handlerConnectedAccountRepoFake) Create(_ context.Context, account *service.Account) error {
	account.ID = r.nextID
	r.nextID++
	cp := *account
	r.accounts[account.ID] = &cp
	return nil
}

func (r *handlerConnectedAccountRepoFake) Delete(_ context.Context, id int64) error {
	if _, ok := r.accounts[id]; !ok {
		return service.ErrAccountNotFound
	}
	delete(r.accounts, id)
	return nil
}

func (r *handlerConnectedAccountRepoFake) UpdateCredentials(_ context.Context, id int64, credentials map[string]any) error {
	account, ok := r.accounts[id]
	if !ok {
		return service.ErrAccountNotFound
	}
	account.Credentials = credentials
	return nil
}

func (r *handlerConnectedAccountRepoFake) GetByIDAndOwnerUserID(_ context.Context, id int64, ownerUserID int64) (*service.Account, error) {
	account, ok := r.accounts[id]
	if !ok || account.OwnerUserID == nil || *account.OwnerUserID != ownerUserID {
		return nil, service.ErrAccountNotFound
	}
	cp := *account
	return &cp, nil
}

func (r *handlerConnectedAccountRepoFake) ListByOwnerUserID(_ context.Context, ownerUserID int64, params pagination.PaginationParams) ([]service.Account, *pagination.PaginationResult, error) {
	out := make([]service.Account, 0)
	for _, account := range r.accounts {
		if account.OwnerUserID != nil && *account.OwnerUserID == ownerUserID {
			out = append(out, *account)
		}
	}
	return out, &pagination.PaginationResult{Total: int64(len(out)), Page: params.Page, PageSize: params.PageSize, Pages: 1}, nil
}

func (r *handlerConnectedAccountRepoFake) BindGroups(_ context.Context, accountID int64, groupIDs []int64) error {
	if _, ok := r.accounts[accountID]; !ok {
		return service.ErrAccountNotFound
	}
	r.boundGroups[accountID] = append([]int64(nil), groupIDs...)
	return nil
}

type handlerConnectedGroupRepoFake struct {
	nextID int64
	groups map[int64]*service.Group
}

func newHandlerConnectedGroupRepoFake() *handlerConnectedGroupRepoFake {
	return &handlerConnectedGroupRepoFake{
		nextID: 1,
		groups: make(map[int64]*service.Group),
	}
}

func (r *handlerConnectedGroupRepoFake) Create(_ context.Context, group *service.Group) error {
	if group.ID == 0 {
		group.ID = r.nextID
		r.nextID++
	}
	cp := *group
	r.groups[group.ID] = &cp
	return nil
}

func (r *handlerConnectedGroupRepoFake) DeleteCascade(_ context.Context, id int64) ([]int64, error) {
	delete(r.groups, id)
	return nil, nil
}

type handlerConnectedOpenAIOAuthFake struct {
	tokenInfo *service.OpenAITokenInfo
}

func (f *handlerConnectedOpenAIOAuthFake) GenerateAuthURL(context.Context, *int64, string, string) (*service.OpenAIAuthURLResult, error) {
	return &service.OpenAIAuthURLResult{AuthURL: "https://auth.example", SessionID: "session"}, nil
}

func (f *handlerConnectedOpenAIOAuthFake) ExchangeCode(context.Context, *service.OpenAIExchangeCodeInput) (*service.OpenAITokenInfo, error) {
	return f.tokenInfo, nil
}

func (f *handlerConnectedOpenAIOAuthFake) RefreshAccountToken(context.Context, *service.Account) (*service.OpenAITokenInfo, error) {
	return f.tokenInfo, nil
}

func (f *handlerConnectedOpenAIOAuthFake) BuildAccountCredentials(tokenInfo *service.OpenAITokenInfo) map[string]any {
	return (&service.OpenAIOAuthService{}).BuildAccountCredentials(tokenInfo)
}

type handlerConnectedAnthropicOAuthFake struct {
	tokenInfo *service.TokenInfo
}

func (f *handlerConnectedAnthropicOAuthFake) GenerateAuthURL(context.Context, *int64) (*service.GenerateAuthURLResult, error) {
	return &service.GenerateAuthURLResult{AuthURL: "https://claude.example/auth", SessionID: "session"}, nil
}

func (f *handlerConnectedAnthropicOAuthFake) ExchangeCode(context.Context, *service.ExchangeCodeInput) (*service.TokenInfo, error) {
	return f.tokenInfo, nil
}

func (f *handlerConnectedAnthropicOAuthFake) RefreshAccountToken(context.Context, *service.Account) (*service.TokenInfo, error) {
	return f.tokenInfo, nil
}

type handlerConnectedGeminiOAuthFake struct {
	tokenInfo *service.GeminiTokenInfo
}

func (f *handlerConnectedGeminiOAuthFake) GenerateAuthURL(context.Context, *int64, string, string, string, string) (*service.GeminiAuthURLResult, error) {
	return &service.GeminiAuthURLResult{AuthURL: "https://gemini.example/auth", SessionID: "session", State: "state"}, nil
}

func (f *handlerConnectedGeminiOAuthFake) ExchangeCode(context.Context, *service.GeminiExchangeCodeInput) (*service.GeminiTokenInfo, error) {
	return f.tokenInfo, nil
}

func (f *handlerConnectedGeminiOAuthFake) RefreshAccountToken(context.Context, *service.Account) (*service.GeminiTokenInfo, error) {
	return f.tokenInfo, nil
}

func (f *handlerConnectedGeminiOAuthFake) BuildAccountCredentials(tokenInfo *service.GeminiTokenInfo) map[string]any {
	return (&service.GeminiOAuthService{}).BuildAccountCredentials(tokenInfo)
}
