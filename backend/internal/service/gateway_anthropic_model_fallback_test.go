package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestGatewayService_AnthropicSonnet429RetriesWithOpusOnSameAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	resetUnix := time.Now().Add(48 * time.Hour).Unix()
	headers429 := http.Header{}
	headers429.Set("Content-Type", "application/json")
	headers429.Set("anthropic-ratelimit-unified-5h-utilization", "0.10")
	headers429.Set("anthropic-ratelimit-unified-5h-reset", formatUnix(time.Now().Add(4*time.Hour).Unix()))
	headers429.Set("anthropic-ratelimit-unified-7d-utilization", "1.00")
	headers429.Set("anthropic-ratelimit-unified-7d-reset", formatUnix(resetUnix))
	headers429.Set("anthropic-ratelimit-unified-7d-surpassed-threshold", "true")
	upstream := &httpUpstreamRecorder{responses: []*http.Response{
		{
			StatusCode: http.StatusTooManyRequests,
			Header:     headers429,
			Body:       io.NopCloser(strings.NewReader(`{"type":"error","error":{"type":"rate_limit_error","message":"sonnet exhausted"}}`)),
		},
		{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}, "x-request-id": []string{"rid-opus"}},
			Body: io.NopCloser(strings.NewReader(`{
				"id":"msg_1",
				"type":"message",
				"role":"assistant",
				"model":"claude-opus-4-7",
				"content":[{"type":"text","text":"ok"}],
				"usage":{"input_tokens":3,"output_tokens":2}
			}`)),
		},
	}}
	repo := &anthropic429ModelRepo{}
	svc := &GatewayService{
		cfg:                 &config.Config{},
		httpUpstream:        upstream,
		rateLimitService:    newRateLimitServiceForTest(repo),
		tlsFPProfileService: &TLSFingerprintProfileService{},
	}
	account := &Account{
		ID:          42,
		Name:        "claude",
		Platform:    PlatformAnthropic,
		Type:        AccountTypeOAuth,
		Concurrency: 1,
		Credentials: map[string]any{
			"access_token": "oauth-token",
		},
		Status:      StatusActive,
		Schedulable: true,
	}
	body := []byte(`{"model":"claude-sonnet-4-5","stream":false,"max_tokens":16,"metadata":{"user_id":"u"},"messages":[{"role":"user","content":"hi"}]}`)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("User-Agent", "claude-cli/2.1.191")

	result, err := svc.Forward(context.Background(), c, account, &ParsedRequest{
		Body:           body,
		Model:          "claude-sonnet-4-5",
		Stream:         false,
		MetadataUserID: "u",
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, upstream.bodies, 2)
	require.Equal(t, "claude-sonnet-4-5-20250929", gjson.GetBytes(upstream.bodies[0], "model").String())
	require.Equal(t, "claude-opus-4-7", gjson.GetBytes(upstream.bodies[1], "model").String())
	require.Equal(t, "claude-sonnet-4-5", result.Model)
	require.Equal(t, "claude-opus-4-7", result.UpstreamModel)
	require.Empty(t, repo.rateLimitedCalls)
	require.Len(t, repo.modelRateLimitCalls, 1)
	require.Equal(t, "claude-sonnet-4-5-20250929", repo.modelRateLimitCalls[0].model)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "claude-sonnet-4-5", gjson.Get(rec.Body.String(), "model").String())
}
