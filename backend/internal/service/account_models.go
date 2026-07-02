package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	AccountExtraAvailableModels            = "available_models"
	AccountExtraAvailableModelsRefreshedAt = "available_models_refreshed_at"

	defaultAnthropicModelsURL = "https://api.anthropic.com/v1/models"
	anthropicVersionHeader    = "2023-06-01"

	DefaultAccountModelRefreshInterval = 24 * time.Hour
)

var (
	ErrAccountModelRefreshUnsupported        = infraerrors.New(http.StatusBadRequest, "ACCOUNT_MODEL_REFRESH_UNSUPPORTED", "only Anthropic OAuth, setup-token, and API key accounts support model refresh")
	ErrAccountModelRefreshMissingCredentials = infraerrors.New(http.StatusBadRequest, "ACCOUNT_MODEL_REFRESH_MISSING_CREDENTIALS", "account is missing Anthropic credentials")
)

type accountExtraUpdater interface {
	UpdateExtra(ctx context.Context, id int64, updates map[string]any) error
}

var anthropicModelsConfig = struct {
	sync.RWMutex
	client  *http.Client
	baseURL string
}{
	client:  http.DefaultClient,
	baseURL: defaultAnthropicModelsURL,
}

func SetAnthropicModelsHTTPClientForTest(client *http.Client, baseURL string) func() {
	anthropicModelsConfig.Lock()
	previousClient := anthropicModelsConfig.client
	previousBaseURL := anthropicModelsConfig.baseURL
	if client == nil {
		client = http.DefaultClient
	}
	anthropicModelsConfig.client = client
	anthropicModelsConfig.baseURL = strings.TrimSpace(baseURL)
	anthropicModelsConfig.Unlock()

	return func() {
		anthropicModelsConfig.Lock()
		anthropicModelsConfig.client = previousClient
		anthropicModelsConfig.baseURL = previousBaseURL
		anthropicModelsConfig.Unlock()
	}
}

func ClaudeAvailableModelsForAccount(account *Account) []claude.Model {
	if account == nil || account.Platform != PlatformAnthropic {
		return cloneClaudeModels(claude.DefaultModels)
	}

	cached := claudeModelsFromExtra(account.Extra)
	mapping := account.GetModelMapping()
	if len(mapping) > 0 && !account.IsOAuth() {
		models := claudeModelsFromMapping(mapping)
		if len(cached) > 0 {
			models = mergeClaudeModels(models, cached)
		}
		return models
	}
	if len(cached) > 0 {
		return cached
	}
	return cloneClaudeModels(claude.DefaultModels)
}

func CanRefreshClaudeModels(account *Account) bool {
	if account == nil || account.Platform != PlatformAnthropic {
		return false
	}
	if account.Type != AccountTypeOAuth && account.Type != AccountTypeSetupToken && account.Type != AccountTypeAPIKey {
		return false
	}
	return strings.TrimSpace(account.GetCredential("api_key")) != "" || strings.TrimSpace(account.GetCredential("access_token")) != ""
}

func ShouldAutoRefreshClaudeModels(account *Account, maxAge time.Duration) bool {
	if !CanRefreshClaudeModels(account) {
		return false
	}
	if len(claudeModelsFromExtra(account.Extra)) == 0 {
		return true
	}
	if maxAge <= 0 {
		maxAge = DefaultAccountModelRefreshInterval
	}
	raw, ok := account.Extra[AccountExtraAvailableModelsRefreshedAt]
	if !ok || raw == nil {
		return true
	}
	refreshedAt, err := time.Parse(time.RFC3339, strings.TrimSpace(fmt.Sprint(raw)))
	if err != nil {
		return true
	}
	return time.Since(refreshedAt) > maxAge
}

func RefreshClaudeAvailableModelsForAccount(ctx context.Context, account *Account, updater accountExtraUpdater) ([]claude.Model, error) {
	if account == nil || account.Platform != PlatformAnthropic {
		return nil, ErrAccountModelRefreshUnsupported
	}
	if account.Type != AccountTypeOAuth && account.Type != AccountTypeSetupToken && account.Type != AccountTypeAPIKey {
		return nil, ErrAccountModelRefreshUnsupported
	}
	if updater == nil {
		return nil, infraerrors.New(http.StatusInternalServerError, "ACCOUNT_MODEL_REFRESH_REPOSITORY_MISSING", "account repository is not configured")
	}

	models, err := fetchAnthropicModels(ctx, account)
	if err != nil {
		return nil, err
	}
	if len(models) == 0 {
		return nil, infraerrors.New(http.StatusBadGateway, "ACCOUNT_MODEL_REFRESH_EMPTY", "Anthropic returned no models")
	}

	refreshedAt := time.Now().UTC().Format(time.RFC3339)
	if err := updater.UpdateExtra(ctx, account.ID, map[string]any{
		AccountExtraAvailableModels:            models,
		AccountExtraAvailableModelsRefreshedAt: refreshedAt,
	}); err != nil {
		return nil, err
	}
	if account.Extra == nil {
		account.Extra = map[string]any{}
	}
	account.Extra[AccountExtraAvailableModels] = models
	account.Extra[AccountExtraAvailableModelsRefreshedAt] = refreshedAt
	return models, nil
}

func fetchAnthropicModels(ctx context.Context, account *Account) ([]claude.Model, error) {
	client, baseURL := currentAnthropicModelsClient()
	if client == nil {
		client = http.DefaultClient
	}

	var models []claude.Model
	afterID := ""
	for {
		page, hasMore, lastID, err := fetchAnthropicModelsPage(ctx, client, baseURL, account, afterID)
		if err != nil {
			return nil, err
		}
		models = mergeClaudeModels(models, page)
		if !hasMore || strings.TrimSpace(lastID) == "" {
			break
		}
		afterID = strings.TrimSpace(lastID)
	}
	return models, nil
}

func currentAnthropicModelsClient() (*http.Client, string) {
	anthropicModelsConfig.RLock()
	defer anthropicModelsConfig.RUnlock()
	baseURL := strings.TrimSpace(anthropicModelsConfig.baseURL)
	if baseURL == "" {
		baseURL = defaultAnthropicModelsURL
	}
	return anthropicModelsConfig.client, baseURL
}

func fetchAnthropicModelsPage(ctx context.Context, client *http.Client, baseURL string, account *Account, afterID string) ([]claude.Model, bool, string, error) {
	endpoint, err := url.Parse(baseURL)
	if err != nil {
		return nil, false, "", err
	}
	query := endpoint.Query()
	query.Set("limit", "100")
	if strings.TrimSpace(afterID) != "" {
		query.Set("after_id", strings.TrimSpace(afterID))
	}
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, false, "", err
	}
	req.Header.Set("anthropic-version", anthropicVersionHeader)
	if err := setAnthropicModelsAuthHeader(req, account); err != nil {
		return nil, false, "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, false, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, false, "", infraerrors.Newf(http.StatusBadGateway, "ACCOUNT_MODEL_REFRESH_FAILED", "Anthropic models API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload struct {
		Data    []claude.Model `json:"data"`
		HasMore bool           `json:"has_more"`
		LastID  string         `json:"last_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, false, "", fmt.Errorf("decode Anthropic models response: %w", err)
	}
	return normalizeClaudeModels(payload.Data), payload.HasMore, payload.LastID, nil
}

func setAnthropicModelsAuthHeader(req *http.Request, account *Account) error {
	if account == nil {
		return ErrAccountModelRefreshMissingCredentials
	}
	apiKey := strings.TrimSpace(account.GetCredential("api_key"))
	accessToken := strings.TrimSpace(account.GetCredential("access_token"))
	if account.Type == AccountTypeAPIKey && apiKey != "" {
		req.Header.Set("x-api-key", apiKey)
		return nil
	}
	if account.IsOAuth() && accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
		return nil
	}
	if apiKey != "" {
		req.Header.Set("x-api-key", apiKey)
		return nil
	}
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
		return nil
	}
	return ErrAccountModelRefreshMissingCredentials
}

func claudeModelsFromExtra(extra map[string]any) []claude.Model {
	if len(extra) == 0 {
		return nil
	}
	raw, ok := extra[AccountExtraAvailableModels]
	if !ok || raw == nil {
		return nil
	}
	switch value := raw.(type) {
	case []claude.Model:
		return normalizeClaudeModels(value)
	case []any:
		data, err := json.Marshal(value)
		if err != nil {
			return nil
		}
		var models []claude.Model
		if err := json.Unmarshal(data, &models); err != nil {
			return nil
		}
		return normalizeClaudeModels(models)
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return nil
		}
		var models []claude.Model
		if err := json.Unmarshal(data, &models); err != nil {
			return nil
		}
		return normalizeClaudeModels(models)
	}
}

func claudeModelsFromMapping(mapping map[string]string) []claude.Model {
	if len(mapping) == 0 {
		return nil
	}
	ids := make([]string, 0, len(mapping))
	for id := range mapping {
		ids = append(ids, id)
	}
	slices.Sort(ids)

	models := make([]claude.Model, 0, len(ids))
	for _, id := range ids {
		if model, ok := defaultClaudeModelByID(id); ok {
			models = append(models, model)
			continue
		}
		models = append(models, claude.Model{
			ID:          id,
			Type:        "model",
			DisplayName: id,
		})
	}
	return models
}

func defaultClaudeModelByID(id string) (claude.Model, bool) {
	for _, model := range claude.DefaultModels {
		if model.ID == id {
			return model, true
		}
	}
	return claude.Model{}, false
}

func normalizeClaudeModels(models []claude.Model) []claude.Model {
	if len(models) == 0 {
		return nil
	}
	out := make([]claude.Model, 0, len(models))
	seen := make(map[string]struct{}, len(models))
	for _, model := range models {
		model.ID = strings.TrimSpace(model.ID)
		if model.ID == "" {
			continue
		}
		if _, ok := seen[model.ID]; ok {
			continue
		}
		if strings.TrimSpace(model.Type) == "" {
			model.Type = "model"
		}
		if strings.TrimSpace(model.DisplayName) == "" {
			model.DisplayName = model.ID
		}
		seen[model.ID] = struct{}{}
		out = append(out, model)
	}
	return out
}

func cloneClaudeModels(models []claude.Model) []claude.Model {
	if len(models) == 0 {
		return []claude.Model{}
	}
	out := make([]claude.Model, len(models))
	copy(out, models)
	return out
}

func mergeClaudeModels(first []claude.Model, rest []claude.Model) []claude.Model {
	if len(first) == 0 {
		return normalizeClaudeModels(rest)
	}
	out := normalizeClaudeModels(first)
	seen := make(map[string]struct{}, len(out))
	for _, model := range out {
		seen[model.ID] = struct{}{}
	}
	for _, model := range normalizeClaudeModels(rest) {
		if _, ok := seen[model.ID]; ok {
			continue
		}
		seen[model.ID] = struct{}{}
		out = append(out, model)
	}
	return out
}
