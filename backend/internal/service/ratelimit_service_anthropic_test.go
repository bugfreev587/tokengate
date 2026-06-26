package service

import (
	"context"
	"net/http"
	"strconv"
	"testing"
	"time"
)

type anthropic429ModelRepo struct {
	sessionWindowMockRepo
	rateLimitedCalls    []anthropic429AccountCall
	modelRateLimitCalls []anthropic429ModelCall
}

type anthropic429AccountCall struct {
	id      int64
	resetAt time.Time
}

type anthropic429ModelCall struct {
	id      int64
	model   string
	resetAt time.Time
}

func (r *anthropic429ModelRepo) SetRateLimited(_ context.Context, id int64, resetAt time.Time) error {
	r.rateLimitedCalls = append(r.rateLimitedCalls, anthropic429AccountCall{id: id, resetAt: resetAt})
	return nil
}

func (r *anthropic429ModelRepo) SetModelRateLimit(_ context.Context, id int64, model string, resetAt time.Time) error {
	r.modelRateLimitCalls = append(r.modelRateLimitCalls, anthropic429ModelCall{id: id, model: model, resetAt: resetAt})
	return nil
}

func TestHandleUpstreamErrorForModel_Anthropic429MarksOnlyRequestedModel(t *testing.T) {
	resetUnix := time.Now().Add(48 * time.Hour).Unix()
	resetAt := time.Unix(resetUnix, 0)
	fiveHourUnix := time.Now().Add(4 * time.Hour).Unix()
	fiveHourReset := time.Unix(fiveHourUnix, 0)

	headers := http.Header{}
	headers.Set("anthropic-ratelimit-unified-5h-utilization", "0.08")
	headers.Set("anthropic-ratelimit-unified-7d-utilization", "1.00")
	headers.Set("anthropic-ratelimit-unified-5h-reset", formatUnix(fiveHourUnix))
	headers.Set("anthropic-ratelimit-unified-7d-reset", formatUnix(resetUnix))

	repo := &anthropic429ModelRepo{}
	svc := newRateLimitServiceForTest(repo)
	account := &Account{ID: 42, Platform: PlatformAnthropic, Type: AccountTypeOAuth}

	shouldDisable := svc.HandleUpstreamErrorForModel(
		context.Background(),
		account,
		"claude-sonnet-4-5",
		http.StatusTooManyRequests,
		headers,
		[]byte(`{"type":"error","error":{"type":"rate_limit_error","message":"rate limited"}}`),
	)

	if shouldDisable {
		t.Fatal("expected model-specific Anthropic 429 to avoid disabling the whole account")
	}
	if len(repo.rateLimitedCalls) != 0 {
		t.Fatalf("expected no account-level SetRateLimited calls, got %d", len(repo.rateLimitedCalls))
	}
	if len(repo.modelRateLimitCalls) != 1 {
		t.Fatalf("expected one SetModelRateLimit call, got %d", len(repo.modelRateLimitCalls))
	}
	call := repo.modelRateLimitCalls[0]
	if call.id != account.ID {
		t.Errorf("expected account ID %d, got %d", account.ID, call.id)
	}
	if call.model != "claude-sonnet-4-5-20250929" {
		t.Errorf("expected normalized model key claude-sonnet-4-5-20250929, got %q", call.model)
	}
	if !call.resetAt.Equal(resetAt) {
		t.Errorf("expected model resetAt %v, got %v", resetAt, call.resetAt)
	}
	if len(repo.sessionWindowCalls) != 1 {
		t.Fatalf("expected one session window update, got %d", len(repo.sessionWindowCalls))
	}
	window := repo.sessionWindowCalls[0]
	if window.End == nil || !window.End.Equal(fiveHourReset) {
		t.Errorf("expected session window end %v, got %v", fiveHourReset, window.End)
	}
	if window.Status != "rejected" {
		t.Errorf("expected rejected session window status, got %q", window.Status)
	}
}

func TestAnthropicRateLimitModelFallback(t *testing.T) {
	tests := []struct {
		name      string
		model     string
		wantModel string
		wantOK    bool
	}{
		{name: "sonnet short id falls back to opus", model: "claude-sonnet-4-5", wantModel: "claude-opus-4-7", wantOK: true},
		{name: "sonnet long id falls back to opus", model: "claude-sonnet-4-5-20250929", wantModel: "claude-opus-4-7", wantOK: true},
		{name: "opus does not fall back", model: "claude-opus-4-7", wantOK: false},
		{name: "haiku does not fall back", model: "claude-haiku-4-5-20251001", wantOK: false},
		{name: "empty model does not fall back", model: "", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := anthropicRateLimitModelFallback(tt.model)
			if ok != tt.wantOK {
				t.Fatalf("expected ok=%v, got %v", tt.wantOK, ok)
			}
			if got != tt.wantModel {
				t.Fatalf("expected fallback %q, got %q", tt.wantModel, got)
			}
		})
	}
}

func TestCalculateAnthropic429ResetTime_Only5hExceeded(t *testing.T) {
	headers := http.Header{}
	headers.Set("anthropic-ratelimit-unified-5h-utilization", "1.02")
	headers.Set("anthropic-ratelimit-unified-5h-reset", "1770998400")
	headers.Set("anthropic-ratelimit-unified-7d-utilization", "0.32")
	headers.Set("anthropic-ratelimit-unified-7d-reset", "1771549200")

	result := calculateAnthropic429ResetTime(headers)
	assertAnthropicResult(t, result, 1770998400)

	if result.fiveHourReset == nil || !result.fiveHourReset.Equal(time.Unix(1770998400, 0)) {
		t.Errorf("expected fiveHourReset=1770998400, got %v", result.fiveHourReset)
	}
}

func TestCalculateAnthropic429ResetTime_Only7dExceeded(t *testing.T) {
	headers := http.Header{}
	headers.Set("anthropic-ratelimit-unified-5h-utilization", "0.50")
	headers.Set("anthropic-ratelimit-unified-5h-reset", "1770998400")
	headers.Set("anthropic-ratelimit-unified-7d-utilization", "1.05")
	headers.Set("anthropic-ratelimit-unified-7d-reset", "1771549200")

	result := calculateAnthropic429ResetTime(headers)
	assertAnthropicResult(t, result, 1771549200)

	// fiveHourReset should still be populated for session window calculation
	if result.fiveHourReset == nil || !result.fiveHourReset.Equal(time.Unix(1770998400, 0)) {
		t.Errorf("expected fiveHourReset=1770998400, got %v", result.fiveHourReset)
	}
}

func TestCalculateAnthropic429ResetTime_BothExceeded(t *testing.T) {
	headers := http.Header{}
	headers.Set("anthropic-ratelimit-unified-5h-utilization", "1.10")
	headers.Set("anthropic-ratelimit-unified-5h-reset", "1770998400")
	headers.Set("anthropic-ratelimit-unified-7d-utilization", "1.02")
	headers.Set("anthropic-ratelimit-unified-7d-reset", "1771549200")

	result := calculateAnthropic429ResetTime(headers)
	assertAnthropicResult(t, result, 1771549200)
}

func TestCalculateAnthropic429ResetTime_NoPerWindowHeaders(t *testing.T) {
	headers := http.Header{}
	headers.Set("anthropic-ratelimit-unified-reset", "1771549200")

	result := calculateAnthropic429ResetTime(headers)
	if result != nil {
		t.Errorf("expected nil result when no per-window headers, got resetAt=%v", result.resetAt)
	}
}

func TestCalculateAnthropic429ResetTime_NoHeaders(t *testing.T) {
	result := calculateAnthropic429ResetTime(http.Header{})
	if result != nil {
		t.Errorf("expected nil result for empty headers, got resetAt=%v", result.resetAt)
	}
}

func TestCalculateAnthropic429ResetTime_SurpassedThreshold(t *testing.T) {
	headers := http.Header{}
	headers.Set("anthropic-ratelimit-unified-5h-surpassed-threshold", "true")
	headers.Set("anthropic-ratelimit-unified-5h-reset", "1770998400")
	headers.Set("anthropic-ratelimit-unified-7d-surpassed-threshold", "false")
	headers.Set("anthropic-ratelimit-unified-7d-reset", "1771549200")

	result := calculateAnthropic429ResetTime(headers)
	assertAnthropicResult(t, result, 1770998400)
}

func TestCalculateAnthropic429ResetTime_UtilizationExactlyOne(t *testing.T) {
	headers := http.Header{}
	headers.Set("anthropic-ratelimit-unified-5h-utilization", "1.0")
	headers.Set("anthropic-ratelimit-unified-5h-reset", "1770998400")
	headers.Set("anthropic-ratelimit-unified-7d-utilization", "0.5")
	headers.Set("anthropic-ratelimit-unified-7d-reset", "1771549200")

	result := calculateAnthropic429ResetTime(headers)
	assertAnthropicResult(t, result, 1770998400)
}

func TestCalculateAnthropic429ResetTime_NeitherExceeded_UsesShorter(t *testing.T) {
	headers := http.Header{}
	headers.Set("anthropic-ratelimit-unified-5h-utilization", "0.95")
	headers.Set("anthropic-ratelimit-unified-5h-reset", "1770998400") // sooner
	headers.Set("anthropic-ratelimit-unified-7d-utilization", "0.80")
	headers.Set("anthropic-ratelimit-unified-7d-reset", "1771549200") // later

	result := calculateAnthropic429ResetTime(headers)
	assertAnthropicResult(t, result, 1770998400)
}

func TestCalculateAnthropic429ResetTime_Only5hResetHeader(t *testing.T) {
	headers := http.Header{}
	headers.Set("anthropic-ratelimit-unified-5h-utilization", "1.05")
	headers.Set("anthropic-ratelimit-unified-5h-reset", "1770998400")

	result := calculateAnthropic429ResetTime(headers)
	assertAnthropicResult(t, result, 1770998400)
}

func TestCalculateAnthropic429ResetTime_Only7dResetHeader(t *testing.T) {
	headers := http.Header{}
	headers.Set("anthropic-ratelimit-unified-7d-utilization", "1.03")
	headers.Set("anthropic-ratelimit-unified-7d-reset", "1771549200")

	result := calculateAnthropic429ResetTime(headers)
	assertAnthropicResult(t, result, 1771549200)

	if result.fiveHourReset != nil {
		t.Errorf("expected fiveHourReset=nil when no 5h headers, got %v", result.fiveHourReset)
	}
}

func TestIsAnthropicWindowExceeded(t *testing.T) {
	tests := []struct {
		name     string
		headers  http.Header
		window   string
		expected bool
	}{
		{
			name:     "utilization above 1.0",
			headers:  makeHeader("anthropic-ratelimit-unified-5h-utilization", "1.02"),
			window:   "5h",
			expected: true,
		},
		{
			name:     "utilization exactly 1.0",
			headers:  makeHeader("anthropic-ratelimit-unified-5h-utilization", "1.0"),
			window:   "5h",
			expected: true,
		},
		{
			name:     "utilization below 1.0",
			headers:  makeHeader("anthropic-ratelimit-unified-5h-utilization", "0.99"),
			window:   "5h",
			expected: false,
		},
		{
			name:     "surpassed-threshold true",
			headers:  makeHeader("anthropic-ratelimit-unified-7d-surpassed-threshold", "true"),
			window:   "7d",
			expected: true,
		},
		{
			name:     "surpassed-threshold True (case insensitive)",
			headers:  makeHeader("anthropic-ratelimit-unified-7d-surpassed-threshold", "True"),
			window:   "7d",
			expected: true,
		},
		{
			name:     "surpassed-threshold false",
			headers:  makeHeader("anthropic-ratelimit-unified-7d-surpassed-threshold", "false"),
			window:   "7d",
			expected: false,
		},
		{
			name:     "no headers",
			headers:  http.Header{},
			window:   "5h",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isAnthropicWindowExceeded(tc.headers, tc.window)
			if got != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

// assertAnthropicResult is a test helper that verifies the result is non-nil and
// has the expected resetAt unix timestamp.
func assertAnthropicResult(t *testing.T, result *anthropic429Result, wantUnix int64) {
	t.Helper()
	if result == nil {
		t.Fatal("expected non-nil result")
		return // unreachable, but satisfies staticcheck SA5011
	}
	want := time.Unix(wantUnix, 0)
	if !result.resetAt.Equal(want) {
		t.Errorf("expected resetAt=%v, got %v", want, result.resetAt)
	}
}

func makeHeader(key, value string) http.Header {
	h := http.Header{}
	h.Set(key, value)
	return h
}

func formatUnix(ts int64) string {
	return strconv.FormatInt(ts, 10)
}
