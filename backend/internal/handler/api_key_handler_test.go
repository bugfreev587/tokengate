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
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyHandler_TestConnection_UsesForwardedGatewayBaseURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	groupID := int64(8)
	repo := &apiKeyHandlerRepoStub{
		apiKey: &service.APIKey{
			ID:      44,
			UserID:  12,
			Key:     "sk-live",
			Name:    "canary",
			GroupID: &groupID,
			Group: &service.Group{
				ID:       groupID,
				Name:     "openai-default",
				Platform: service.PlatformOpenAI,
			},
		},
	}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer sk-live", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/models":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]string{{"id": "gpt-4.1-mini"}},
			})
		case "/v1/chat/completions":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"choices": []map[string]any{{"message": map[string]string{"content": "ok"}}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer upstream.Close()

	svc := service.NewAPIKeyService(repo, nil, nil, nil, nil, nil, nil)
	router := gin.New()
	handler := NewAPIKeyHandler(svc)
	router.POST("/api/v1/keys/:id/test-connection", func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 12})
		handler.TestConnection(c)
	})

	body := bytes.NewBufferString(`{"max_models":1}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/keys/44/test-connection", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-Proto", "http")
	req.Header.Set("X-Forwarded-Host", upstream.Listener.Addr().String())
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var envelope struct {
		Code int `json:"code"`
		Data struct {
			Success          bool   `json:"success"`
			BaseURL          string `json:"base_url"`
			TestedModelCount int    `json:"tested_model_count"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	require.Equal(t, 0, envelope.Code)
	require.True(t, envelope.Data.Success)
	require.Equal(t, upstream.URL, envelope.Data.BaseURL)
	require.Equal(t, 1, envelope.Data.TestedModelCount)
}

func TestAPIKeyHandler_ClaudeCodeConnect_UsesForwardedGatewayBaseURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	groupID := int64(8)
	repo := &apiKeyHandlerRepoStub{
		apiKey: &service.APIKey{
			ID:      45,
			UserID:  12,
			Key:     "sk-claude-live",
			Name:    "claude",
			GroupID: &groupID,
			Group: &service.Group{
				ID:       groupID,
				Name:     "anthropic-default",
				Platform: service.PlatformAnthropic,
			},
		},
	}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer sk-claude-live", r.Header.Get("Authorization"))
		require.Equal(t, "/v1/models", r.URL.Path)
		require.Equal(t, "1000", r.URL.Query().Get("limit"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]string{
				{"id": "claude-opus-4-8"},
				{"id": "claude-fable-5"},
				{"id": "claude-sonnet-4-6"},
			},
		})
	}))
	defer upstream.Close()

	svc := service.NewAPIKeyService(repo, nil, nil, nil, nil, nil, nil)
	router := gin.New()
	handler := NewAPIKeyHandler(svc)
	router.GET("/api/v1/keys/:id/claude-code/connect", func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 12})
		handler.ClaudeCodeConnect(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/keys/45/claude-code/connect", nil)
	req.Header.Set("X-Forwarded-Proto", "http")
	req.Header.Set("X-Forwarded-Host", upstream.Listener.Addr().String())
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var envelope struct {
		Code int `json:"code"`
		Data struct {
			Supported bool   `json:"supported"`
			BaseURL   string `json:"base_url"`
			Settings  struct {
				Env map[string]string `json:"env"`
			} `json:"settings"`
			Models struct {
				Fable string   `json:"fable"`
				Items []string `json:"available"`
			} `json:"models"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	require.Equal(t, 0, envelope.Code)
	require.True(t, envelope.Data.Supported)
	require.Equal(t, upstream.URL, envelope.Data.BaseURL)
	require.Equal(t, upstream.URL, envelope.Data.Settings.Env["ANTHROPIC_BASE_URL"])
	require.Equal(t, "sk-claude-live", envelope.Data.Settings.Env["ANTHROPIC_AUTH_TOKEN"])
	require.Equal(t, "1", envelope.Data.Settings.Env["CLAUDE_CODE_ENABLE_GATEWAY_MODEL_DISCOVERY"])
	require.Equal(t, "claude-fable-5", envelope.Data.Models.Fable)
	require.Contains(t, envelope.Data.Models.Items, "claude-fable-5")
}

type apiKeyHandlerRepoStub struct {
	apiKey *service.APIKey
}

func (s *apiKeyHandlerRepoStub) Create(ctx context.Context, key *service.APIKey) error {
	panic("unexpected Create call")
}

func (s *apiKeyHandlerRepoStub) GetByID(ctx context.Context, id int64) (*service.APIKey, error) {
	if s.apiKey == nil || s.apiKey.ID != id {
		return nil, service.ErrAPIKeyNotFound
	}
	clone := *s.apiKey
	return &clone, nil
}

func (s *apiKeyHandlerRepoStub) GetKeyAndOwnerID(ctx context.Context, id int64) (string, int64, error) {
	panic("unexpected GetKeyAndOwnerID call")
}

func (s *apiKeyHandlerRepoStub) GetByKey(ctx context.Context, key string) (*service.APIKey, error) {
	panic("unexpected GetByKey call")
}

func (s *apiKeyHandlerRepoStub) GetByKeyForAuth(ctx context.Context, key string) (*service.APIKey, error) {
	panic("unexpected GetByKeyForAuth call")
}

func (s *apiKeyHandlerRepoStub) Update(ctx context.Context, key *service.APIKey) error {
	panic("unexpected Update call")
}

func (s *apiKeyHandlerRepoStub) Delete(ctx context.Context, id int64) error {
	panic("unexpected Delete call")
}

func (s *apiKeyHandlerRepoStub) ListByUserID(ctx context.Context, userID int64, params pagination.PaginationParams, filters service.APIKeyListFilters) ([]service.APIKey, *pagination.PaginationResult, error) {
	panic("unexpected ListByUserID call")
}

func (s *apiKeyHandlerRepoStub) VerifyOwnership(ctx context.Context, userID int64, apiKeyIDs []int64) ([]int64, error) {
	panic("unexpected VerifyOwnership call")
}

func (s *apiKeyHandlerRepoStub) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	panic("unexpected CountByUserID call")
}

func (s *apiKeyHandlerRepoStub) ExistsByKey(ctx context.Context, key string) (bool, error) {
	panic("unexpected ExistsByKey call")
}

func (s *apiKeyHandlerRepoStub) ListByGroupID(ctx context.Context, groupID int64, params pagination.PaginationParams) ([]service.APIKey, *pagination.PaginationResult, error) {
	panic("unexpected ListByGroupID call")
}

func (s *apiKeyHandlerRepoStub) SearchAPIKeys(ctx context.Context, userID int64, keyword string, limit int) ([]service.APIKey, error) {
	panic("unexpected SearchAPIKeys call")
}

func (s *apiKeyHandlerRepoStub) ClearGroupIDByGroupID(ctx context.Context, groupID int64) (int64, error) {
	panic("unexpected ClearGroupIDByGroupID call")
}

func (s *apiKeyHandlerRepoStub) UpdateGroupIDByUserAndGroup(ctx context.Context, userID, oldGroupID, newGroupID int64) (int64, error) {
	panic("unexpected UpdateGroupIDByUserAndGroup call")
}

func (s *apiKeyHandlerRepoStub) CountByGroupID(ctx context.Context, groupID int64) (int64, error) {
	panic("unexpected CountByGroupID call")
}

func (s *apiKeyHandlerRepoStub) ListKeysByUserID(ctx context.Context, userID int64) ([]string, error) {
	panic("unexpected ListKeysByUserID call")
}

func (s *apiKeyHandlerRepoStub) ListKeysByGroupID(ctx context.Context, groupID int64) ([]string, error) {
	panic("unexpected ListKeysByGroupID call")
}

func (s *apiKeyHandlerRepoStub) IncrementQuotaUsed(ctx context.Context, id int64, amount float64) (float64, error) {
	panic("unexpected IncrementQuotaUsed call")
}

func (s *apiKeyHandlerRepoStub) UpdateLastUsed(ctx context.Context, id int64, usedAt time.Time) error {
	panic("unexpected UpdateLastUsed call")
}

func (s *apiKeyHandlerRepoStub) IncrementRateLimitUsage(ctx context.Context, id int64, cost float64) error {
	panic("unexpected IncrementRateLimitUsage call")
}

func (s *apiKeyHandlerRepoStub) ResetRateLimitWindows(ctx context.Context, id int64) error {
	panic("unexpected ResetRateLimitWindows call")
}

func (s *apiKeyHandlerRepoStub) GetRateLimitData(ctx context.Context, id int64) (*service.APIKeyRateLimitData, error) {
	panic("unexpected GetRateLimitData call")
}
