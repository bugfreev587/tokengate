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

func TestAPIKeyService_TestConnection_OpenAIHappyPath(t *testing.T) {
	modelsCalled := false
	chatCalled := false

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer sk-live-test", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/v1/models":
			modelsCalled = true
			_ = json.NewEncoder(w).Encode(map[string]any{
				"object": "list",
				"data": []map[string]string{
					{"id": "gpt-4.1-mini"},
				},
			})
		case "/v1/chat/completions":
			chatCalled = true
			var body map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			require.Equal(t, "gpt-4.1-mini", body["model"])
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":      "chatcmpl-test",
				"object":  "chat.completion",
				"choices": []map[string]any{{"message": map[string]string{"content": "ok"}}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer upstream.Close()

	groupID := int64(9)
	svc := &APIKeyService{
		apiKeyRepo: &apiKeyRepoStub{
			apiKey: &APIKey{
				ID:      42,
				UserID:  7,
				Key:     "sk-live-test",
				Name:    "prod-key",
				GroupID: &groupID,
				Group: &Group{
					ID:       groupID,
					Name:     "openai-default",
					Platform: PlatformOpenAI,
				},
			},
		},
	}

	result, err := svc.TestConnection(context.Background(), 42, 7, APIKeyConnectionTestOptions{
		BaseURL:    upstream.URL,
		HTTPClient: upstream.Client(),
		MaxModels:  1,
	})

	require.NoError(t, err)
	require.True(t, modelsCalled)
	require.True(t, chatCalled)
	require.True(t, result.Success)
	require.True(t, result.ModelsVisible)
	require.Equal(t, int64(42), result.APIKeyID)
	require.Equal(t, "prod-key", result.KeyName)
	require.Equal(t, int64(9), *result.GroupID)
	require.Equal(t, "openai-default", result.GroupName)
	require.Equal(t, PlatformOpenAI, result.Platform)
	require.Equal(t, 1, result.VisibleModelCount)
	require.Equal(t, 1, result.TestedModelCount)
	require.Len(t, result.Results, 1)
	require.Equal(t, "gpt-4.1-mini", result.Results[0].Model)
	require.Equal(t, "success", result.Results[0].Status)
	require.Equal(t, "/v1/chat/completions", result.Results[0].Endpoint)
	require.GreaterOrEqual(t, result.Results[0].LatencyMs, 0)
}

func TestAPIKeyService_TestConnection_OpenAIChatGPTAccountUnsupportedModelIsSkipped(t *testing.T) {
	probedModels := make(map[string]bool)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer sk-live-test", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/v1/models":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"object": "list",
				"data": []map[string]string{
					{"id": "gpt-5.4"},
					{"id": "gpt-5.3-codex"},
				},
			})
		case "/v1/chat/completions":
			var body map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			model, _ := body["model"].(string)
			probedModels[model] = true
			if model == "gpt-5.3-codex" {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error": map[string]string{
						"message": "The 'gpt-5.3-codex' model is not supported when using Codex with a ChatGPT account.",
						"type":    "invalid_request_error",
					},
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"choices": []map[string]any{{"message": map[string]string{"content": "ok"}}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer upstream.Close()

	groupID := int64(9)
	svc := &APIKeyService{
		apiKeyRepo: &apiKeyRepoStub{
			apiKey: &APIKey{
				ID:      42,
				UserID:  7,
				Key:     "sk-live-test",
				Name:    "prod-key",
				GroupID: &groupID,
				Group: &Group{
					ID:       groupID,
					Name:     "openai-default",
					Platform: PlatformOpenAI,
				},
			},
		},
	}

	result, err := svc.TestConnection(context.Background(), 42, 7, APIKeyConnectionTestOptions{
		BaseURL:    upstream.URL,
		HTTPClient: upstream.Client(),
		MaxModels:  2,
	})

	require.NoError(t, err)
	require.True(t, probedModels["gpt-5.4"])
	require.True(t, probedModels["gpt-5.3-codex"])
	require.True(t, result.Success)
	require.Equal(t, 2, result.VisibleModelCount)
	require.Equal(t, 1, result.TestedModelCount)
	require.Equal(t, 1, result.SkippedModelCount)
	require.Len(t, result.Results, 2)
	require.Equal(t, APIKeyConnectionStatusSuccess, result.Results[0].Status)
	require.Equal(t, "gpt-5.3-codex", result.Results[1].Model)
	require.Equal(t, APIKeyConnectionStatusSkipped, result.Results[1].Status)
	require.Equal(t, http.StatusBadRequest, result.Results[1].HTTPStatus)
	require.Contains(t, result.Results[1].Message, "ChatGPT account")
}
