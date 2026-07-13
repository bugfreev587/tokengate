//go:build unit

package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyAuthBYOAdmissionPrecedesBalance(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ownerID := int64(42)
	enabled := false
	user := &service.User{ID: ownerID, Role: service.RoleUser, Status: service.StatusActive, Balance: 100, Concurrency: 1}
	group := &service.Group{
		ID: 7, Name: "byo-openai", Status: service.StatusActive, OwnerUserID: &ownerID,
		CapacitySource: service.CapacitySourceConnectedAccount, BYOEnabled: &enabled,
		BYODisabledReason: service.BYOAccountDisabledReasonSubscriptionInactive,
	}
	key := &service.APIKey{ID: 8, UserID: ownerID, Key: "sk-byo-inactive", Status: service.StatusActive, User: user, Group: group, GroupID: &group.ID}
	repo := &stubApiKeyRepo{getByKey: func(context.Context, string) (*service.APIKey, error) { return key, nil }}
	cfg := &config.Config{RunMode: config.RunModeStandard}
	router := newAuthTestRouter(service.NewAPIKeyService(repo, nil, nil, nil, nil, nil, cfg), nil, cfg)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/t", nil)
	request.Header.Set("x-api-key", key.Key)
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusForbidden, recorder.Code)
	var body ErrorResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, "SUBSCRIPTION_REQUIRED", body.Code)
}

func TestAPIKeyAuthActiveBYOSkipsTokenGateBalance(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ownerID := int64(42)
	enabled := true
	user := &service.User{ID: ownerID, Role: service.RoleUser, Status: service.StatusActive, Balance: 0, Concurrency: 1}
	group := &service.Group{ID: 7, Status: service.StatusActive, OwnerUserID: &ownerID, CapacitySource: service.CapacitySourceConnectedAccount, BYOEnabled: &enabled}
	key := &service.APIKey{ID: 8, UserID: ownerID, Key: "sk-byo-active", Status: service.StatusActive, User: user, Group: group, GroupID: &group.ID}
	repo := &stubApiKeyRepo{getByKey: func(context.Context, string) (*service.APIKey, error) { return key, nil }}
	cfg := &config.Config{RunMode: config.RunModeStandard}
	router := newAuthTestRouter(service.NewAPIKeyService(repo, nil, nil, nil, nil, nil, cfg), nil, cfg)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/t", nil)
	request.Header.Set("x-api-key", key.Key)
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestGoogleAPIKeyAuthReturnsStructuredBYOSubscriptionReason(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ownerID := int64(42)
	enabled := false
	user := &service.User{ID: ownerID, Role: service.RoleUser, Status: service.StatusActive, Balance: 100, Concurrency: 1}
	group := &service.Group{
		ID: 7, Status: service.StatusActive, OwnerUserID: &ownerID,
		CapacitySource: service.CapacitySourceConnectedAccount, BYOEnabled: &enabled,
		BYODisabledReason: service.BYOAccountDisabledReasonSubscriptionInactive,
	}
	key := &service.APIKey{ID: 8, UserID: ownerID, Key: "sk-google-byo-inactive", Status: service.StatusActive, User: user, Group: group, GroupID: &group.ID}
	apiKeyService := newTestAPIKeyService(fakeAPIKeyRepo{getByKey: func(context.Context, string) (*service.APIKey, error) { return key, nil }})
	router := gin.New()
	router.Use(APIKeyAuthWithSubscriptionGoogle(apiKeyService, nil, &config.Config{RunMode: config.RunModeStandard}))
	router.GET("/v1beta/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/v1beta/test", nil)
	request.Header.Set("x-goog-api-key", key.Key)
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusForbidden, recorder.Code)
	var body struct {
		Error struct {
			Details []struct {
				Reason string `json:"reason"`
			} `json:"details"`
		} `json:"error"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Len(t, body.Error.Details, 1)
	require.Equal(t, "SUBSCRIPTION_REQUIRED", body.Error.Details[0].Reason)
}

func TestEvaluateBYOAdmissionKeepsOperationalFailuresDistinct(t *testing.T) {
	ownerID := int64(42)
	enabled := false
	key := &service.APIKey{
		User:  &service.User{ID: ownerID},
		Group: &service.Group{OwnerUserID: &ownerID, CapacitySource: service.CapacitySourceConnectedAccount, BYOEnabled: &enabled, BYODisabledReason: service.BYOAccountDisabledReasonAccountDisabled},
	}

	admission := evaluateBYOAdmission(key)
	require.NotNil(t, admission)
	require.Equal(t, http.StatusServiceUnavailable, admission.Status)
	require.Equal(t, "CONNECTED_ACCOUNT_ERROR", admission.Code)
}
