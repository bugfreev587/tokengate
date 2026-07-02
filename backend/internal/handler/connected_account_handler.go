package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type ConnectedAccountHandler struct {
	connectedAccountService *service.ConnectedAccountService
}

func NewConnectedAccountHandler(connectedAccountService *service.ConnectedAccountService) *ConnectedAccountHandler {
	return &ConnectedAccountHandler{connectedAccountService: connectedAccountService}
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
