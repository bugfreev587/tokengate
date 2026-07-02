//go:build unit

package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAPIKeyService_BuildClaudeCodeConnect_AnthropicKeyUsesGatewayDiscoveryAndFamilyPins(t *testing.T) {
	modelsCalled := false
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer sk-claude-live", r.Header.Get("Authorization"))
		require.Equal(t, "/v1/models", r.URL.Path)
		require.Equal(t, "1000", r.URL.Query().Get("limit"))
		modelsCalled = true
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object": "list",
			"data": []map[string]string{
				{"id": "claude-opus-4-8"},
				{"id": "claude-fable-5"},
				{"id": "claude-sonnet-4-6"},
				{"id": "claude-haiku-4-5-20251001"},
			},
		})
	}))
	defer upstream.Close()

	groupID := int64(11)
	svc := &APIKeyService{
		apiKeyRepo: &apiKeyRepoStub{
			apiKey: &APIKey{
				ID:      99,
				UserID:  7,
				Key:     "sk-claude-live",
				Name:    "claude-prod",
				GroupID: &groupID,
				Group: &Group{
					ID:       groupID,
					Name:     "anthropic-default",
					Platform: PlatformAnthropic,
				},
			},
		},
	}

	payload, err := svc.BuildClaudeCodeConnect(context.Background(), 99, 7, ClaudeCodeConnectOptions{
		BaseURL:    upstream.URL,
		HTTPClient: upstream.Client(),
	})

	require.NoError(t, err)
	require.True(t, modelsCalled)
	require.True(t, payload.Supported)
	require.Empty(t, payload.Reason)
	require.Equal(t, upstream.URL, payload.BaseURL)
	require.Equal(t, "claude-prod", payload.KeyName)
	require.Equal(t, PlatformAnthropic, payload.Platform)
	require.Equal(t, "2.1.129", payload.MinimumVersions.GatewayDiscovery)
	require.Equal(t, "2.1.170", payload.MinimumVersions.FablePicker)
	require.Equal(t, "latest", payload.RecommendedClaudeCodeVersion)

	env := payload.Settings.Env
	require.Equal(t, upstream.URL, env["ANTHROPIC_BASE_URL"])
	require.Equal(t, "sk-claude-live", env["ANTHROPIC_AUTH_TOKEN"])
	require.Equal(t, "1", env["CLAUDE_CODE_ENABLE_GATEWAY_MODEL_DISCOVERY"])
	require.Equal(t, "claude-opus-4-8", env["ANTHROPIC_DEFAULT_OPUS_MODEL"])
	require.Equal(t, "claude-fable-5", env["ANTHROPIC_DEFAULT_FABLE_MODEL"])
	require.Equal(t, "claude-sonnet-4-6", env["ANTHROPIC_DEFAULT_SONNET_MODEL"])
	require.Equal(t, "claude-haiku-4-5-20251001", env["ANTHROPIC_DEFAULT_HAIKU_MODEL"])
	require.NotContains(t, env, "ANTHROPIC_CUSTOM_MODEL_OPTION")
	require.NotContains(t, env, "CLAUDE_CODE_ATTRIBUTION_HEADER")
	require.NotContains(t, env, "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC")
	require.Empty(t, payload.Settings.AvailableModels)
	require.False(t, payload.Settings.EnforceAvailableModels)

	require.Equal(t, "claude-opus-4-8", payload.Models.Opus)
	require.Equal(t, "claude-fable-5", payload.Models.Fable)
	require.Equal(t, "claude-sonnet-4-6", payload.Models.Sonnet)
	require.Equal(t, "claude-haiku-4-5-20251001", payload.Models.Haiku)
	require.Equal(t, []string{
		"claude-opus-4-8",
		"claude-fable-5",
		"claude-sonnet-4-6",
		"claude-haiku-4-5-20251001",
	}, payload.Models.Available)

	require.Equal(t, []string{
		"claude-opus-4-8",
		"claude-fable-5",
		"claude-sonnet-4-6",
		"claude-haiku-4-5-20251001",
	}, payload.OptionalPolicySettings.AvailableModels)
	require.False(t, payload.OptionalPolicySettings.EnforceAvailableModels)
}

func TestAPIKeyService_BuildClaudeCodeConnect_OpenAIOnlyKeyIsUnsupported(t *testing.T) {
	groupID := int64(12)
	svc := &APIKeyService{
		apiKeyRepo: &apiKeyRepoStub{
			apiKey: &APIKey{
				ID:      100,
				UserID:  7,
				Key:     "sk-openai-live",
				Name:    "openai-prod",
				GroupID: &groupID,
				Group: &Group{
					ID:       groupID,
					Name:     "openai-default",
					Platform: PlatformOpenAI,
				},
			},
		},
	}

	payload, err := svc.BuildClaudeCodeConnect(context.Background(), 100, 7, ClaudeCodeConnectOptions{
		BaseURL: "https://api.tokengate.to",
	})

	require.NoError(t, err)
	require.False(t, payload.Supported)
	require.Equal(t, "GROUP_NOT_ANTHROPIC_COMPATIBLE", payload.Reason)
	require.Contains(t, payload.Message, "Anthropic-compatible")
	require.Empty(t, payload.Settings.Env)
}
