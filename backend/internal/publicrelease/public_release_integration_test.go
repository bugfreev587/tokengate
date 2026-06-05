//go:build integration

package publicrelease

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/server/routes"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newPublicReleaseIntegrationRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	routes.RegisterCommonRoutes(router)
	routes.RegisterGatewayRoutes(
		router,
		&handler.Handlers{
			Gateway:       &handler.GatewayHandler{},
			OpenAIGateway: &handler.OpenAIGatewayHandler{},
		},
		middleware.APIKeyAuthMiddleware(func(c *gin.Context) {
			groupID := int64(1)
			c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{
				GroupID: &groupID,
				Group:   &service.Group{Platform: service.PlatformOpenAI},
			})
			c.Next()
		}),
		nil,
		nil,
		nil,
		nil,
		&config.Config{},
	)

	return router
}

func TestPublicReleaseHTTPIntegrationGatewayEntrypoints(t *testing.T) {
	server := httptest.NewServer(newPublicReleaseIntegrationRouter())
	t.Cleanup(server.Close)

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{name: "health", method: http.MethodGet, path: "/health"},
		{
			name:   "chat completions",
			method: http.MethodPost,
			path:   "/v1/chat/completions",
			body:   `{"model":"gpt-5","messages":[{"role":"user","content":"ping"}]}`,
		},
		{
			name:   "responses compact",
			method: http.MethodPost,
			path:   "/backend-api/codex/responses/compact",
			body:   `{"model":"gpt-5","input":"ping"}`,
		},
		{
			name:   "messages",
			method: http.MethodPost,
			path:   "/v1/messages",
			body:   `{"model":"claude-sonnet-4-5","max_tokens":32,"messages":[{"role":"user","content":"ping"}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, server.URL+tt.path, bytes.NewBufferString(tt.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			t.Cleanup(func() { _ = resp.Body.Close() })

			require.NotEqual(t, http.StatusNotFound, resp.StatusCode, "public API path %s should survive HTTP integration routing", tt.path)
		})
	}
}
