package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
	"github.com/stretchr/testify/require"
)

type accountModelsRepoStub struct {
	updates map[string]any
}

func (r *accountModelsRepoStub) UpdateExtra(_ context.Context, _ int64, updates map[string]any) error {
	r.updates = updates
	return nil
}

func TestClaudeAvailableModelsForAccountUsesCachedModelsBeforeDefaults(t *testing.T) {
	account := &Account{
		Platform: PlatformAnthropic,
		Type:     AccountTypeOAuth,
		Extra: map[string]any{
			"available_models": []any{
				map[string]any{
					"id":           "claude-mythos-5",
					"type":         "model",
					"display_name": "Claude Mythos 5",
					"created_at":   "2026-07-02T00:00:00Z",
				},
			},
		},
	}

	models := ClaudeAvailableModelsForAccount(account)

	require.Equal(t, []claude.Model{{
		ID:          "claude-mythos-5",
		Type:        "model",
		DisplayName: "Claude Mythos 5",
		CreatedAt:   "2026-07-02T00:00:00Z",
	}}, models)
}

func TestRefreshClaudeAvailableModelsForAccountFetchesPagesAndStoresExtra(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.URL.String())
		require.Equal(t, "Bearer oauth-token", r.Header.Get("Authorization"))
		require.Equal(t, "2023-06-01", r.Header.Get("anthropic-version"))

		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("after_id") {
		case "":
			_, _ = w.Write([]byte(`{
				"data": [
					{"id": "claude-sonnet-5", "type": "model", "display_name": "Claude Sonnet 5", "created_at": "2026-07-01T00:00:00Z"}
				],
				"has_more": true,
				"last_id": "claude-sonnet-5"
			}`))
		case "claude-sonnet-5":
			_, _ = w.Write([]byte(`{
				"data": [
					{"id": "claude-mythos-5", "type": "model", "display_name": "Claude Mythos 5", "created_at": "2026-07-02T00:00:00Z"}
				],
				"has_more": false
			}`))
		default:
			t.Fatalf("unexpected after_id %q", r.URL.Query().Get("after_id"))
		}
	}))
	defer server.Close()
	restore := SetAnthropicModelsHTTPClientForTest(server.Client(), server.URL)
	defer restore()

	repo := &accountModelsRepoStub{}
	models, err := RefreshClaudeAvailableModelsForAccount(context.Background(), &Account{
		ID:       42,
		Platform: PlatformAnthropic,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token": "oauth-token",
		},
	}, repo)

	require.NoError(t, err)
	require.Equal(t, []string{"/?limit=100", "/?after_id=claude-sonnet-5&limit=100"}, requests)
	require.Len(t, models, 2)
	require.Equal(t, "claude-sonnet-5", models[0].ID)
	require.Equal(t, "claude-mythos-5", models[1].ID)
	require.NotEmpty(t, repo.updates["available_models_refreshed_at"])

	raw, err := json.Marshal(repo.updates["available_models"])
	require.NoError(t, err)
	require.JSONEq(t, `[
		{"id": "claude-sonnet-5", "type": "model", "display_name": "Claude Sonnet 5", "created_at": "2026-07-01T00:00:00Z"},
		{"id": "claude-mythos-5", "type": "model", "display_name": "Claude Mythos 5", "created_at": "2026-07-02T00:00:00Z"}
	]`, string(raw))
}

func TestShouldAutoRefreshClaudeModelsUsesCacheAge(t *testing.T) {
	account := &Account{
		Platform: PlatformAnthropic,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token": "oauth-token",
		},
		Extra: map[string]any{},
	}

	require.True(t, ShouldAutoRefreshClaudeModels(account, time.Hour))

	account.Extra[AccountExtraAvailableModels] = []any{
		map[string]any{"id": "claude-sonnet-5", "type": "model", "display_name": "Claude Sonnet 5"},
	}
	account.Extra[AccountExtraAvailableModelsRefreshedAt] = time.Now().UTC().Format(time.RFC3339)
	require.False(t, ShouldAutoRefreshClaudeModels(account, time.Hour))

	account.Extra[AccountExtraAvailableModelsRefreshedAt] = time.Now().UTC().Add(-2 * time.Hour).Format(time.RFC3339)
	require.True(t, ShouldAutoRefreshClaudeModels(account, time.Hour))
}
