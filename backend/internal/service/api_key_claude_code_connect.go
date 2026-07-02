package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	claudeCodeConnectModelsLimit = "1000"
	claudeCodeConnectTimeout     = 10 * time.Second
)

type ClaudeCodeConnectOptions struct {
	BaseURL    string
	HTTPClient *http.Client
}

type ClaudeCodeConnectPayload struct {
	Supported                    bool                                  `json:"supported"`
	Reason                       string                                `json:"reason,omitempty"`
	Message                      string                                `json:"message,omitempty"`
	BaseURL                      string                                `json:"base_url,omitempty"`
	KeyName                      string                                `json:"key_name,omitempty"`
	GroupID                      *int64                                `json:"group_id,omitempty"`
	GroupName                    string                                `json:"group_name,omitempty"`
	Platform                     string                                `json:"platform,omitempty"`
	Settings                     ClaudeCodeSettings                    `json:"settings"`
	OptionalPolicySettings       ClaudeCodePolicySettings              `json:"optional_policy_settings,omitempty"`
	OptionalEnv                  map[string]ClaudeCodeOptionalEnvEntry `json:"optional_env,omitempty"`
	Models                       ClaudeCodeConnectModels               `json:"models,omitempty"`
	MinimumVersions              ClaudeCodeMinimumVersions             `json:"minimum_versions,omitempty"`
	RecommendedClaudeCodeVersion string                                `json:"recommended_claude_code_version,omitempty"`
}

type ClaudeCodeSettings struct {
	Env                    map[string]string `json:"env,omitempty"`
	AvailableModels        []string          `json:"availableModels,omitempty"`
	EnforceAvailableModels bool              `json:"enforceAvailableModels,omitempty"`
}

type ClaudeCodePolicySettings struct {
	AvailableModels        []string `json:"availableModels,omitempty"`
	EnforceAvailableModels bool     `json:"enforceAvailableModels"`
}

type ClaudeCodeOptionalEnvEntry struct {
	Value          string `json:"value"`
	DefaultEnabled bool   `json:"default_enabled"`
	Reason         string `json:"reason"`
}

type ClaudeCodeConnectModels struct {
	Default   string   `json:"default,omitempty"`
	Opus      string   `json:"opus,omitempty"`
	Sonnet    string   `json:"sonnet,omitempty"`
	Haiku     string   `json:"haiku,omitempty"`
	Fable     string   `json:"fable,omitempty"`
	Available []string `json:"available,omitempty"`
}

type ClaudeCodeMinimumVersions struct {
	GatewayDiscovery       string `json:"gateway_discovery"`
	FablePicker            string `json:"fable_picker"`
	EnforceAvailableModels string `json:"enforce_available_models"`
	Sonnet5BuiltinAlias    string `json:"sonnet_5_builtin_alias"`
}

func (s *APIKeyService) BuildClaudeCodeConnect(ctx context.Context, id int64, userID int64, opts ClaudeCodeConnectOptions) (*ClaudeCodeConnectPayload, error) {
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

	payload := newClaudeCodeConnectPayload(apiKey, baseURL)
	if !apiKeySupportsClaudeCode(apiKey) {
		payload.Supported = false
		payload.Reason = "GROUP_NOT_ANTHROPIC_COMPATIBLE"
		payload.Message = "This API key is not assigned to an Anthropic-compatible group. Select an Anthropic-compatible group to use Claude Code."
		return payload, nil
	}

	client := opts.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: claudeCodeConnectTimeout}
	}

	models, err := fetchClaudeCodeConnectModels(ctx, client, baseURL, apiKey.Key)
	if err != nil {
		return nil, err
	}
	claudeModels := filterClaudeCodeModelIDs(models)
	if len(claudeModels) == 0 {
		payload.Supported = false
		payload.Reason = "NO_CLAUDE_MODELS_VISIBLE"
		payload.Message = "TokenGate /v1/models returned no Claude-compatible models for this API key."
		return payload, nil
	}

	families := selectClaudeCodeFamilies(claudeModels)
	env := map[string]string{
		"ANTHROPIC_BASE_URL":                         baseURL,
		"ANTHROPIC_AUTH_TOKEN":                       apiKey.Key,
		"CLAUDE_CODE_ENABLE_GATEWAY_MODEL_DISCOVERY": "1",
	}
	if families.Opus != "" {
		env["ANTHROPIC_DEFAULT_OPUS_MODEL"] = families.Opus
	}
	if families.Sonnet != "" {
		env["ANTHROPIC_DEFAULT_SONNET_MODEL"] = families.Sonnet
	}
	if families.Haiku != "" {
		env["ANTHROPIC_DEFAULT_HAIKU_MODEL"] = families.Haiku
	}
	if families.Fable != "" {
		env["ANTHROPIC_DEFAULT_FABLE_MODEL"] = families.Fable
	}

	payload.Supported = true
	payload.Settings.Env = env
	payload.OptionalPolicySettings = ClaudeCodePolicySettings{
		AvailableModels:        append([]string(nil), claudeModels...),
		EnforceAvailableModels: false,
	}
	payload.OptionalEnv = map[string]ClaudeCodeOptionalEnvEntry{
		"CLAUDE_CODE_ATTRIBUTION_HEADER": {
			Value:          "0",
			DefaultEnabled: false,
			Reason:         "Omit Claude Code's system-prompt attribution block only when TokenGate explicitly wants that gateway behavior.",
		},
	}
	payload.Models = families
	payload.Models.Available = append([]string(nil), claudeModels...)
	payload.Models.Default = firstClaudeCodeModel(families.Opus, families.Sonnet, families.Fable, families.Haiku)
	return payload, nil
}

func newClaudeCodeConnectPayload(apiKey *APIKey, baseURL string) *ClaudeCodeConnectPayload {
	payload := &ClaudeCodeConnectPayload{
		Supported:                    false,
		BaseURL:                      baseURL,
		Settings:                     ClaudeCodeSettings{Env: map[string]string{}},
		MinimumVersions:              defaultClaudeCodeMinimumVersions(),
		RecommendedClaudeCodeVersion: "latest",
	}
	if apiKey == nil {
		return payload
	}
	payload.KeyName = apiKey.Name
	payload.GroupID = apiKey.GroupID
	if apiKey.Group != nil {
		payload.GroupName = apiKey.Group.Name
		payload.Platform = apiKey.Group.Platform
	}
	return payload
}

func defaultClaudeCodeMinimumVersions() ClaudeCodeMinimumVersions {
	return ClaudeCodeMinimumVersions{
		GatewayDiscovery:       "2.1.129",
		FablePicker:            "2.1.170",
		EnforceAvailableModels: "2.1.175",
		Sonnet5BuiltinAlias:    "2.1.197",
	}
}

func apiKeySupportsClaudeCode(apiKey *APIKey) bool {
	if apiKey == nil || apiKey.Group == nil {
		return false
	}
	switch apiKey.Group.Platform {
	case PlatformAnthropic, PlatformAntigravity:
		return true
	case PlatformOpenAI:
		return apiKey.Group.AllowMessagesDispatch
	default:
		return false
	}
}

func fetchClaudeCodeConnectModels(ctx context.Context, client *http.Client, baseURL, apiKey string) ([]string, error) {
	endpoint, err := url.Parse(strings.TrimRight(baseURL, "/") + "/v1/models")
	if err != nil {
		return nil, infraerrors.BadRequest("INVALID_BASE_URL", "base url must be an absolute HTTP URL")
	}
	query := endpoint.Query()
	query.Set("limit", claudeCodeConnectModelsLimit)
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create claude code models request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, infraerrors.Newf(http.StatusBadGateway, "CLAUDE_CODE_MODELS_FETCH_FAILED", "failed to fetch TokenGate models: %s", sanitizeConnectionTestMessage(err.Error()))
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, apiKeyConnectionBodyLimit))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, infraerrors.Newf(http.StatusBadGateway, "CLAUDE_CODE_MODELS_FETCH_FAILED", "TokenGate /v1/models returned %d: %s", resp.StatusCode, sanitizeConnectionTestMessage(string(body)))
	}
	return parseConnectionTestModels(body), nil
}

func filterClaudeCodeModelIDs(models []string) []string {
	out := make([]string, 0, len(models))
	seen := map[string]struct{}{}
	for _, model := range models {
		model = strings.TrimSpace(model)
		lower := strings.ToLower(model)
		if !strings.HasPrefix(lower, "claude") && !strings.HasPrefix(lower, "anthropic") {
			continue
		}
		if _, ok := seen[model]; ok {
			continue
		}
		seen[model] = struct{}{}
		out = append(out, model)
	}
	return out
}

func selectClaudeCodeFamilies(models []string) ClaudeCodeConnectModels {
	return ClaudeCodeConnectModels{
		Opus:   selectClaudeCodeFamilyModel(models, "opus"),
		Sonnet: selectClaudeCodeFamilyModel(models, "sonnet"),
		Haiku:  selectClaudeCodeFamilyModel(models, "haiku"),
		Fable:  selectClaudeCodeFamilyModel(models, "fable"),
	}
}

func selectClaudeCodeFamilyModel(models []string, family string) string {
	candidates := make([]string, 0, len(models))
	needle := "-" + family + "-"
	for _, model := range models {
		if strings.Contains(strings.ToLower(model), needle) {
			candidates = append(candidates, model)
		}
	}
	if len(candidates) == 0 {
		return ""
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		return claudeCodeModelRank(candidates[i]) > claudeCodeModelRank(candidates[j])
	})
	return candidates[0]
}

func claudeCodeModelRank(model string) int {
	lower := strings.ToLower(model)
	score := 0
	switch {
	case strings.Contains(lower, "-fable-5"):
		score += 50000
	case strings.Contains(lower, "-opus-4-8"):
		score += 40800
	case strings.Contains(lower, "-opus-4-7"):
		score += 40700
	case strings.Contains(lower, "-opus-4-6"):
		score += 40600
	case strings.Contains(lower, "-opus-4-5"):
		score += 40500
	case strings.Contains(lower, "-sonnet-5"):
		score += 50000
	case strings.Contains(lower, "-sonnet-4-6"):
		score += 40600
	case strings.Contains(lower, "-sonnet-4-5"):
		score += 40500
	case strings.Contains(lower, "-haiku-4-5"):
		score += 40500
	}
	if strings.Contains(lower, "thinking") {
		score -= 10
	}
	return score
}

func firstClaudeCodeModel(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
