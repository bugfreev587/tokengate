package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/antigravity"
	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
	"github.com/Wei-Shaw/sub2api/internal/pkg/geminicli"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type ConnectedAccountHandler struct {
	connectedAccountService *service.ConnectedAccountService
	accountTestService      *service.AccountTestService
	rateLimitService        *service.RateLimitService
}

func NewConnectedAccountHandler(
	connectedAccountService *service.ConnectedAccountService,
	accountTestService *service.AccountTestService,
	rateLimitService *service.RateLimitService,
) *ConnectedAccountHandler {
	return &ConnectedAccountHandler{
		connectedAccountService: connectedAccountService,
		accountTestService:      accountTestService,
		rateLimitService:        rateLimitService,
	}
}

type ConnectedAccountSummary struct {
	ID             int64      `json:"id"`
	Name           string     `json:"name"`
	Platform       string     `json:"platform"`
	Type           string     `json:"type"`
	Status         string     `json:"status"`
	Email          string     `json:"email,omitempty"`
	PlanType       string     `json:"plan_type,omitempty"`
	GroupID        *int64     `json:"group_id,omitempty"`
	GroupName      string     `json:"group_name,omitempty"`
	CapacitySource string     `json:"capacity_source"`
	LastUsedAt     *time.Time `json:"last_used_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type connectedOpenAIAuthURLRequest struct {
	ProxyID     *int64 `json:"proxy_id"`
	RedirectURI string `json:"redirect_uri"`
}

type connectedAnthropicAuthURLRequest struct {
	ProxyID *int64 `json:"proxy_id"`
}

type connectedGeminiAuthURLRequest struct {
	ProxyID     *int64 `json:"proxy_id"`
	RedirectURI string `json:"redirect_uri"`
	ProjectID   string `json:"project_id"`
	OAuthType   string `json:"oauth_type"`
	TierID      string `json:"tier_id"`
}

type connectedOpenAIExchangeCodeRequest struct {
	SessionID   string `json:"session_id" binding:"required"`
	Code        string `json:"code" binding:"required"`
	State       string `json:"state" binding:"required"`
	RedirectURI string `json:"redirect_uri"`
	ProxyID     *int64 `json:"proxy_id"`
	Name        string `json:"name"`
	Concurrency int    `json:"concurrency"`
	Priority    int    `json:"priority"`
}

type connectedAnthropicExchangeCodeRequest struct {
	SessionID   string `json:"session_id" binding:"required"`
	Code        string `json:"code" binding:"required"`
	ProxyID     *int64 `json:"proxy_id"`
	Name        string `json:"name"`
	Concurrency int    `json:"concurrency"`
	Priority    int    `json:"priority"`
}

type connectedGeminiExchangeCodeRequest struct {
	SessionID   string `json:"session_id" binding:"required"`
	Code        string `json:"code" binding:"required"`
	State       string `json:"state" binding:"required"`
	ProxyID     *int64 `json:"proxy_id"`
	Name        string `json:"name"`
	Concurrency int    `json:"concurrency"`
	Priority    int    `json:"priority"`
	OAuthType   string `json:"oauth_type"`
	TierID      string `json:"tier_id"`
}

type connectedAccountTestRequest struct {
	ModelID string `json:"model_id"`
	Prompt  string `json:"prompt"`
	Mode    string `json:"mode"`
}

func (h *ConnectedAccountHandler) List(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    c.DefaultQuery("sort_by", "created_at"),
		SortOrder: c.DefaultQuery("sort_order", "desc"),
	}
	accounts, result, err := h.connectedAccountService.List(c.Request.Context(), subject.UserID, params)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	out := make([]ConnectedAccountSummary, 0, len(accounts))
	for i := range accounts {
		out = append(out, connectedAccountSummaryFromService(&accounts[i]))
	}
	response.Paginated(c, out, result.Total, page, pageSize)
}

func (h *ConnectedAccountHandler) GenerateOpenAIAuthURL(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var req connectedOpenAIAuthURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req = connectedOpenAIAuthURLRequest{}
	}
	result, err := h.connectedAccountService.GenerateOpenAIAuthURL(c.Request.Context(), subject.UserID, req.ProxyID, req.RedirectURI)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ConnectedAccountHandler) GenerateAnthropicAuthURL(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var req connectedAnthropicAuthURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req = connectedAnthropicAuthURLRequest{}
	}
	result, err := h.connectedAccountService.GenerateAnthropicAuthURL(c.Request.Context(), subject.UserID, req.ProxyID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ConnectedAccountHandler) GenerateGeminiAuthURL(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var req connectedGeminiAuthURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req = connectedGeminiAuthURLRequest{}
	}
	result, err := h.connectedAccountService.GenerateGeminiAuthURL(c.Request.Context(), subject.UserID, req.ProxyID, req.RedirectURI, req.ProjectID, req.OAuthType, req.TierID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ConnectedAccountHandler) ExchangeOpenAICode(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var req connectedOpenAIExchangeCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	account, err := h.connectedAccountService.CreateOpenAIAccountFromOAuth(c.Request.Context(), service.CreateConnectedOpenAIAccountInput{
		UserID:      subject.UserID,
		SessionID:   req.SessionID,
		Code:        req.Code,
		State:       req.State,
		RedirectURI: req.RedirectURI,
		ProxyID:     req.ProxyID,
		Name:        req.Name,
		Concurrency: req.Concurrency,
		Priority:    req.Priority,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, connectedAccountSummaryFromService(account))
}

func (h *ConnectedAccountHandler) ExchangeAnthropicCode(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var req connectedAnthropicExchangeCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	account, err := h.connectedAccountService.CreateAnthropicAccountFromOAuth(c.Request.Context(), service.CreateConnectedAnthropicAccountInput{
		UserID:      subject.UserID,
		SessionID:   req.SessionID,
		Code:        req.Code,
		ProxyID:     req.ProxyID,
		Name:        req.Name,
		Concurrency: req.Concurrency,
		Priority:    req.Priority,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, connectedAccountSummaryFromService(account))
}

func (h *ConnectedAccountHandler) ExchangeGeminiCode(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var req connectedGeminiExchangeCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	account, err := h.connectedAccountService.CreateGeminiAccountFromOAuth(c.Request.Context(), service.CreateConnectedGeminiAccountInput{
		UserID:      subject.UserID,
		SessionID:   req.SessionID,
		Code:        req.Code,
		State:       req.State,
		ProxyID:     req.ProxyID,
		Name:        req.Name,
		Concurrency: req.Concurrency,
		Priority:    req.Priority,
		OAuthType:   req.OAuthType,
		TierID:      req.TierID,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, connectedAccountSummaryFromService(account))
}

func (h *ConnectedAccountHandler) GetAvailableModels(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || accountID <= 0 {
		response.BadRequest(c, "Invalid account ID")
		return
	}
	account, err := h.connectedAccountService.GetOwnedAccount(c.Request.Context(), subject.UserID, accountID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if service.ShouldAutoRefreshClaudeModels(account, service.DefaultAccountModelRefreshInterval) {
		if models, err := h.connectedAccountService.RefreshAccountModels(c.Request.Context(), subject.UserID, accountID); err == nil {
			response.Success(c, models)
			return
		}
	}
	response.Success(c, connectedAccountAvailableModels(account))
}

func (h *ConnectedAccountHandler) RefreshAvailableModels(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || accountID <= 0 {
		response.BadRequest(c, "Invalid account ID")
		return
	}
	models, err := h.connectedAccountService.RefreshAccountModels(c.Request.Context(), subject.UserID, accountID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, models)
}

func (h *ConnectedAccountHandler) Test(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || accountID <= 0 {
		response.BadRequest(c, "Invalid account ID")
		return
	}
	account, err := h.connectedAccountService.GetOwnedAccount(c.Request.Context(), subject.UserID, accountID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if h.accountTestService == nil {
		response.ErrorFrom(c, service.ErrConnectedAccountUnsupported)
		return
	}

	var req connectedAccountTestRequest
	_ = c.ShouldBindJSON(&req)
	if err := h.accountTestService.TestAccountConnection(c, account.ID, req.ModelID, req.Prompt, req.Mode); err != nil {
		return
	}
	if h.rateLimitService != nil {
		if _, err := h.rateLimitService.RecoverAccountAfterSuccessfulTest(c.Request.Context(), account.ID); err != nil {
			_ = c.Error(err)
		}
	}
}

func (h *ConnectedAccountHandler) Refresh(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || accountID <= 0 {
		response.BadRequest(c, "Invalid account ID")
		return
	}
	account, err := h.connectedAccountService.RefreshAccount(c.Request.Context(), subject.UserID, accountID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, connectedAccountSummaryFromService(account))
}

func (h *ConnectedAccountHandler) Delete(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || accountID <= 0 {
		response.BadRequest(c, "Invalid account ID")
		return
	}
	if err := h.connectedAccountService.Delete(c.Request.Context(), subject.UserID, accountID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func connectedAccountAvailableModels(account *service.Account) any {
	if account == nil {
		return []claude.Model{}
	}
	if account.IsOpenAI() {
		if account.IsOpenAIPassthroughEnabled() {
			return openai.DefaultModels
		}
		mapping := account.GetModelMapping()
		if len(mapping) == 0 {
			return openai.DefaultModels
		}
		models := make([]openai.Model, 0, len(mapping))
		for requestedModel := range mapping {
			if model, ok := findOpenAIModel(requestedModel); ok {
				models = append(models, model)
				continue
			}
			models = append(models, openai.Model{
				ID:          requestedModel,
				Object:      "model",
				Type:        "model",
				DisplayName: requestedModel,
			})
		}
		return models
	}
	if account.IsGemini() {
		if account.IsOAuth() {
			return geminicli.DefaultModels
		}
		mapping := account.GetModelMapping()
		if len(mapping) == 0 {
			return geminicli.DefaultModels
		}
		models := make([]geminicli.Model, 0, len(mapping))
		for requestedModel := range mapping {
			if model, ok := findGeminiModel(requestedModel); ok {
				models = append(models, model)
				continue
			}
			models = append(models, geminicli.Model{
				ID:          requestedModel,
				Type:        "model",
				DisplayName: requestedModel,
				CreatedAt:   "",
			})
		}
		return models
	}
	if account.Platform == service.PlatformAntigravity {
		return antigravity.DefaultModels()
	}
	return service.ClaudeAvailableModelsForAccount(account)
}

func findOpenAIModel(id string) (openai.Model, bool) {
	for _, model := range openai.DefaultModels {
		if model.ID == id {
			return model, true
		}
	}
	return openai.Model{}, false
}

func findGeminiModel(id string) (geminicli.Model, bool) {
	for _, model := range geminicli.DefaultModels {
		if model.ID == id {
			return model, true
		}
	}
	return geminicli.Model{}, false
}

func findClaudeModel(id string) (claude.Model, bool) {
	for _, model := range claude.DefaultModels {
		if model.ID == id {
			return model, true
		}
	}
	return claude.Model{}, false
}

func connectedAccountSummaryFromService(account *service.Account) ConnectedAccountSummary {
	if account == nil {
		return ConnectedAccountSummary{}
	}
	out := ConnectedAccountSummary{
		ID:             account.ID,
		Name:           account.Name,
		Platform:       account.Platform,
		Type:           account.Type,
		Status:         account.Status,
		Email:          connectedAccountEmail(account),
		PlanType:       credentialString(account.Credentials, "plan_type"),
		CapacitySource: service.CapacitySourceConnectedAccount,
		LastUsedAt:     account.LastUsedAt,
		CreatedAt:      account.CreatedAt,
		UpdatedAt:      account.UpdatedAt,
	}
	for _, group := range account.Groups {
		if group == nil || !group.IsUserOwnedConnectedAccount() {
			continue
		}
		groupID := group.ID
		out.GroupID = &groupID
		out.GroupName = group.Name
		out.CapacitySource = service.NormalizeCapacitySource(group.CapacitySource)
		break
	}
	return out
}

func connectedAccountEmail(account *service.Account) string {
	if account == nil {
		return ""
	}
	if email := credentialString(account.Credentials, "email"); email != "" {
		return email
	}
	return credentialString(account.Extra, "email_address")
}

func credentialString(credentials map[string]any, key string) string {
	if len(credentials) == 0 {
		return ""
	}
	value, ok := credentials[key]
	if !ok || value == nil {
		return ""
	}
	return strings.TrimSpace(toString(value))
}

func toString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return ""
	}
}
