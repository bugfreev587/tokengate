package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	apiKeyConnectionDefaultMaxModels = 12
	apiKeyConnectionHardMaxModels    = 20
	apiKeyConnectionTimeout          = 45 * time.Second
	apiKeyConnectionBodyLimit        = 512 * 1024

	APIKeyConnectionStatusSuccess = "success"
	APIKeyConnectionStatusError   = "error"
	APIKeyConnectionStatusSkipped = "skipped"

	chatGPTCodexUnsupportedModelPhrase = "not supported when using codex with a chatgpt account"
)

// APIKeyConnectionTestOptions configures a live TokenGate gateway probe.
type APIKeyConnectionTestOptions struct {
	BaseURL    string
	MaxModels  int
	Models     []string
	HTTPClient *http.Client
}

// APIKeyConnectionTestResult is returned to the API key owner after a live probe.
type APIKeyConnectionTestResult struct {
	APIKeyID          int64                             `json:"api_key_id"`
	KeyName           string                            `json:"key_name"`
	GroupID           *int64                            `json:"group_id,omitempty"`
	GroupName         string                            `json:"group_name,omitempty"`
	Platform          string                            `json:"platform,omitempty"`
	BaseURL           string                            `json:"base_url"`
	Success           bool                              `json:"success"`
	ModelsVisible     bool                              `json:"models_visible"`
	VisibleModelCount int                               `json:"visible_model_count"`
	TestedModelCount  int                               `json:"tested_model_count"`
	SkippedModelCount int                               `json:"skipped_model_count"`
	Truncated         bool                              `json:"truncated"`
	Message           string                            `json:"message"`
	Results           []APIKeyConnectionTestModelResult `json:"results"`
}

// APIKeyConnectionTestModelResult contains the outcome for one gateway endpoint/model pair.
type APIKeyConnectionTestModelResult struct {
	Model      string `json:"model"`
	Provider   string `json:"provider"`
	Endpoint   string `json:"endpoint"`
	Status     string `json:"status"`
	HTTPStatus int    `json:"http_status,omitempty"`
	LatencyMs  int    `json:"latency_ms"`
	Message    string `json:"message,omitempty"`
}

// TestConnection sends a small live request through the TokenGate gateway using this API key.
func (s *APIKeyService) TestConnection(ctx context.Context, id int64, userID int64, opts APIKeyConnectionTestOptions) (*APIKeyConnectionTestResult, error) {
	apiKey, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if apiKey.UserID != userID {
		return nil, ErrInsufficientPerms
	}

	baseURL, err := normalizeConnectionTestBaseURL(opts.BaseURL)
	if err != nil {
		return nil, err
	}

	result := newAPIKeyConnectionResult(apiKey, baseURL)
	client := opts.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: apiKeyConnectionTimeout}
	}

	models, probe := listConnectionTestModels(ctx, client, baseURL, apiKey.Key, result.Platform)
	if probe.Status != APIKeyConnectionStatusSuccess {
		result.Results = append(result.Results, probe)
		result.Message = probe.Message
		return result, nil
	}

	result.ModelsVisible = len(models) > 0
	result.VisibleModelCount = len(models)
	selected := selectConnectionTestModels(models, opts.Models, opts.MaxModels)
	result.Truncated = len(selected) < len(models)
	if len(selected) == 0 {
		result.Message = "No visible models are available for this API key"
		return result, nil
	}

	for _, model := range selected {
		modelResult := probeConnectionTestModel(ctx, client, baseURL, apiKey.Key, result.Platform, model)
		result.Results = append(result.Results, modelResult)
		if modelResult.Status == APIKeyConnectionStatusSkipped {
			result.SkippedModelCount++
			continue
		}
		result.TestedModelCount++
	}
	result.Success = connectionTestResultsSuccessful(result.Results)
	result.Message = connectionTestSummary(result)
	return result, nil
}

func newAPIKeyConnectionResult(apiKey *APIKey, baseURL string) *APIKeyConnectionTestResult {
	result := &APIKeyConnectionTestResult{
		APIKeyID: apiKey.ID,
		KeyName:  apiKey.Name,
		GroupID:  apiKey.GroupID,
		BaseURL:  baseURL,
		Results:  []APIKeyConnectionTestModelResult{},
	}
	if apiKey.Group != nil {
		result.GroupName = apiKey.Group.Name
		result.Platform = apiKey.Group.Platform
	}
	return result
}

func normalizeConnectionTestBaseURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", infraerrors.BadRequest("INVALID_BASE_URL", "base url is required")
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", infraerrors.BadRequest("INVALID_BASE_URL", "base url must be an absolute HTTP URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", infraerrors.BadRequest("INVALID_BASE_URL", "base url must use http or https")
	}
	path := strings.TrimRight(parsed.EscapedPath(), "/")
	path = strings.TrimSuffix(path, "/api/v1")
	return parsed.Scheme + "://" + parsed.Host + path, nil
}

func listConnectionTestModels(ctx context.Context, client *http.Client, baseURL, apiKey, platform string) ([]string, APIKeyConnectionTestModelResult) {
	endpoint, googleAuth := connectionTestModelsEndpoint(platform)
	result := APIKeyConnectionTestModelResult{
		Model:    "*",
		Provider: platform,
		Endpoint: endpoint,
		Status:   APIKeyConnectionStatusError,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+endpoint, nil)
	if err != nil {
		result.Message = "Failed to create model list request"
		return nil, result
	}
	applyConnectionTestAuth(req, apiKey, googleAuth)

	start := time.Now()
	resp, err := client.Do(req)
	result.LatencyMs = int(time.Since(start) / time.Millisecond)
	if err != nil {
		result.Message = sanitizeConnectionTestMessage(err.Error())
		return nil, result
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, apiKeyConnectionBodyLimit))
	result.HTTPStatus = resp.StatusCode
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		result.Message = connectionTestHTTPMessage(resp.StatusCode, body)
		return nil, result
	}

	models := parseConnectionTestModels(body)
	result.Status = APIKeyConnectionStatusSuccess
	return models, result
}

func connectionTestModelsEndpoint(platform string) (endpoint string, googleAuth bool) {
	switch platform {
	case PlatformGemini:
		return "/v1beta/models", true
	case PlatformAntigravity:
		return "/antigravity/models", false
	default:
		return "/v1/models", false
	}
}

func parseConnectionTestModels(body []byte) []string {
	var openAIStyle struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &openAIStyle); err == nil && len(openAIStyle.Data) > 0 {
		models := make([]string, 0, len(openAIStyle.Data))
		for _, item := range openAIStyle.Data {
			if model := strings.TrimSpace(item.ID); model != "" {
				models = append(models, model)
			}
		}
		return dedupeConnectionTestStrings(models)
	}

	var geminiStyle struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.Unmarshal(body, &geminiStyle); err == nil && len(geminiStyle.Models) > 0 {
		models := make([]string, 0, len(geminiStyle.Models))
		for _, item := range geminiStyle.Models {
			model := strings.TrimSpace(strings.TrimPrefix(item.Name, "models/"))
			if model != "" {
				models = append(models, model)
			}
		}
		return dedupeConnectionTestStrings(models)
	}

	return nil
}

func selectConnectionTestModels(visible []string, requested []string, maxModels int) []string {
	if maxModels <= 0 {
		maxModels = apiKeyConnectionDefaultMaxModels
	}
	if maxModels > apiKeyConnectionHardMaxModels {
		maxModels = apiKeyConnectionHardMaxModels
	}

	source := visible
	if len(requested) > 0 {
		allowed := make(map[string]struct{}, len(visible))
		for _, model := range visible {
			allowed[model] = struct{}{}
		}
		source = make([]string, 0, len(requested))
		for _, model := range requested {
			model = strings.TrimSpace(model)
			if _, ok := allowed[model]; ok {
				source = append(source, model)
			}
		}
	}
	source = dedupeConnectionTestStrings(source)
	if len(source) > maxModels {
		return source[:maxModels]
	}
	return source
}

func probeConnectionTestModel(ctx context.Context, client *http.Client, baseURL, apiKey, platform, model string) APIKeyConnectionTestModelResult {
	endpoint, body, googleAuth, skipped := buildConnectionTestProbe(platform, model)
	result := APIKeyConnectionTestModelResult{
		Model:    model,
		Provider: platform,
		Endpoint: endpoint,
		Status:   APIKeyConnectionStatusError,
	}
	if skipped != "" {
		result.Status = APIKeyConnectionStatusSkipped
		result.Message = skipped
		return result
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+endpoint, bytes.NewReader(body))
	if err != nil {
		result.Message = "Failed to create model test request"
		return result
	}
	req.Header.Set("Content-Type", "application/json")
	applyConnectionTestAuth(req, apiKey, googleAuth)
	if endpoint == "/v1/messages" || endpoint == "/antigravity/v1/messages" {
		req.Header.Set("anthropic-version", "2023-06-01")
	}

	start := time.Now()
	resp, err := client.Do(req)
	result.LatencyMs = int(time.Since(start) / time.Millisecond)
	if err != nil {
		result.Message = sanitizeConnectionTestMessage(err.Error())
		return result
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, apiKeyConnectionBodyLimit))
	result.HTTPStatus = resp.StatusCode
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if isChatGPTCodexUnsupportedModelResponse(resp.StatusCode, respBody) {
			result.Status = APIKeyConnectionStatusSkipped
			result.Message = connectionTestJSONErrorMessage(respBody)
			return result
		}
		result.Message = connectionTestHTTPMessage(resp.StatusCode, respBody)
		return result
	}
	result.Status = APIKeyConnectionStatusSuccess
	return result
}

func buildConnectionTestProbe(platform, model string) (endpoint string, body []byte, googleAuth bool, skipped string) {
	switch platform {
	case PlatformOpenAI:
		if isOpenAIImageModel(model) {
			return "/v1/chat/completions", nil, false, "Image model quick probe is skipped"
		}
		return "/v1/chat/completions", mustJSON(map[string]any{
			"model": model,
			"messages": []map[string]string{
				{"role": "user", "content": "Reply with exactly: ok"},
			},
			"stream": false,
		}), false, ""
	case PlatformGemini:
		cleanModel := strings.TrimPrefix(model, "models/")
		return "/v1beta/models/" + url.PathEscape(cleanModel) + ":generateContent", mustJSON(map[string]any{
			"contents": []map[string]any{
				{"parts": []map[string]string{{"text": "Reply with exactly: ok"}}},
			},
			"generationConfig": map[string]int{"maxOutputTokens": 8},
		}), true, ""
	case PlatformAntigravity:
		if isGeminiLikeModel(model) {
			cleanModel := strings.TrimPrefix(model, "models/")
			return "/antigravity/v1beta/models/" + url.PathEscape(cleanModel) + ":generateContent", mustJSON(map[string]any{
				"contents": []map[string]any{
					{"parts": []map[string]string{{"text": "Reply with exactly: ok"}}},
				},
				"generationConfig": map[string]int{"maxOutputTokens": 8},
			}), true, ""
		}
		return "/antigravity/v1/messages", mustJSON(map[string]any{
			"model":      model,
			"max_tokens": 8,
			"messages": []map[string]string{
				{"role": "user", "content": "Reply with exactly: ok"},
			},
		}), false, ""
	default:
		return "/v1/messages", mustJSON(map[string]any{
			"model":      model,
			"max_tokens": 8,
			"messages": []map[string]string{
				{"role": "user", "content": "Reply with exactly: ok"},
			},
		}), false, ""
	}
}

func applyConnectionTestAuth(req *http.Request, apiKey string, googleAuth bool) {
	if googleAuth {
		req.Header.Set("x-goog-api-key", apiKey)
		return
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
}

func connectionTestResultsSuccessful(results []APIKeyConnectionTestModelResult) bool {
	seenProbe := false
	for _, result := range results {
		if result.Status == APIKeyConnectionStatusSkipped {
			continue
		}
		seenProbe = true
		if result.Status != APIKeyConnectionStatusSuccess {
			return false
		}
	}
	return seenProbe
}

func connectionTestSummary(result *APIKeyConnectionTestResult) string {
	if result.Success {
		if result.Truncated {
			return fmt.Sprintf("Tested %d of %d visible models successfully", result.TestedModelCount, result.VisibleModelCount)
		}
		return fmt.Sprintf("Tested %d visible models successfully", result.TestedModelCount)
	}
	return "One or more model probes failed"
}

func connectionTestHTTPMessage(status int, body []byte) string {
	snippet := strings.TrimSpace(string(body))
	if snippet == "" {
		return fmt.Sprintf("HTTP %d", status)
	}
	return sanitizeConnectionTestMessage(fmt.Sprintf("HTTP %d: %s", status, snippet))
}

func isChatGPTCodexUnsupportedModelResponse(status int, body []byte) bool {
	if status != http.StatusBadRequest {
		return false
	}
	return strings.Contains(strings.ToLower(string(body)), chatGPTCodexUnsupportedModelPhrase)
}

func connectionTestJSONErrorMessage(body []byte) string {
	var payload struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &payload); err == nil && strings.TrimSpace(payload.Error.Message) != "" {
		return sanitizeConnectionTestMessage(payload.Error.Message)
	}
	return "Model is not supported when using Codex with a ChatGPT account"
}

func sanitizeConnectionTestMessage(message string) string {
	message = strings.TrimSpace(strings.ReplaceAll(message, "\n", " "))
	if len(message) > 500 {
		return message[:500] + "..."
	}
	return message
}

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func dedupeConnectionTestStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func isGeminiLikeModel(model string) bool {
	model = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(model), "models/"))
	return strings.HasPrefix(model, "gemini-") || strings.HasPrefix(model, "learnlm-")
}
