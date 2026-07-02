//go:build unit

package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type authCookieRefreshTokenCacheStub struct {
	tokens  map[string]*service.RefreshTokenData
	deleted []string
}

func newAuthCookieRefreshTokenCacheStub() *authCookieRefreshTokenCacheStub {
	return &authCookieRefreshTokenCacheStub{
		tokens: make(map[string]*service.RefreshTokenData),
	}
}

func (s *authCookieRefreshTokenCacheStub) StoreRefreshToken(_ context.Context, tokenHash string, data *service.RefreshTokenData, _ time.Duration) error {
	cloned := *data
	s.tokens[tokenHash] = &cloned
	return nil
}

func (s *authCookieRefreshTokenCacheStub) GetRefreshToken(_ context.Context, tokenHash string) (*service.RefreshTokenData, error) {
	data, ok := s.tokens[tokenHash]
	if !ok {
		return nil, service.ErrRefreshTokenNotFound
	}
	cloned := *data
	return &cloned, nil
}

func (s *authCookieRefreshTokenCacheStub) DeleteRefreshToken(_ context.Context, tokenHash string) error {
	s.deleted = append(s.deleted, tokenHash)
	delete(s.tokens, tokenHash)
	return nil
}

func (s *authCookieRefreshTokenCacheStub) DeleteUserRefreshTokens(context.Context, int64) error {
	return nil
}

func (s *authCookieRefreshTokenCacheStub) DeleteTokenFamily(context.Context, string) error {
	return nil
}

func (s *authCookieRefreshTokenCacheStub) AddToUserTokenSet(context.Context, int64, string, time.Duration) error {
	return nil
}

func (s *authCookieRefreshTokenCacheStub) AddToFamilyTokenSet(context.Context, string, string, time.Duration) error {
	return nil
}

func (s *authCookieRefreshTokenCacheStub) GetUserTokenHashes(context.Context, int64) ([]string, error) {
	return nil, nil
}

func (s *authCookieRefreshTokenCacheStub) GetFamilyTokenHashes(context.Context, string) ([]string, error) {
	return nil, nil
}

func (s *authCookieRefreshTokenCacheStub) IsTokenInFamily(context.Context, string, string) (bool, error) {
	return false, nil
}

func newAuthCookieTestHandler(t *testing.T) (*AuthHandler, *service.AuthService, *authCookieRefreshTokenCacheStub, *service.User) {
	t.Helper()

	cfg := &config.Config{}
	cfg.JWT.Secret = "test-jwt-secret-32bytes-long!!!"
	cfg.JWT.AccessTokenExpireMinutes = 15
	cfg.JWT.RefreshTokenExpireDays = 7

	user := &service.User{
		ID:           42,
		Email:        "user@example.com",
		Role:         service.RoleUser,
		Status:       service.StatusActive,
		TokenVersion: 1,
	}
	cache := newAuthCookieRefreshTokenCacheStub()
	authService := service.NewAuthService(
		nil,
		&userHandlerRepoStub{user: user},
		nil,
		cache,
		cfg,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	handler := &AuthHandler{
		cfg:         cfg,
		authService: authService,
	}

	return handler, authService, cache, user
}

func TestAuthHandlerRespondWithTokenPairSetsHttpOnlyAuthCookies(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{}
	cfg.JWT.Secret = "test-jwt-secret-32bytes-long!!!"
	cfg.JWT.AccessTokenExpireMinutes = 15
	cfg.JWT.RefreshTokenExpireDays = 7

	handler := &AuthHandler{
		cfg: cfg,
		authService: service.NewAuthService(
			nil,
			nil,
			nil,
			&userHandlerRefreshTokenCacheStub{},
			cfg,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
		),
	}

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	c.Request.Header.Set("X-Forwarded-Proto", "https")

	handler.respondWithTokenPair(c, &service.User{
		ID:           42,
		Email:        "user@example.com",
		Role:         service.RoleUser,
		Status:       service.StatusActive,
		TokenVersion: 1,
	})

	require.Equal(t, http.StatusOK, recorder.Code)
	cookies := recorder.Result().Cookies()

	accessCookie := findCookie(cookies, "tokengate_access_token")
	require.NotNil(t, accessCookie)
	require.NotEmpty(t, accessCookie.Value)
	require.Equal(t, "/", accessCookie.Path)
	require.True(t, accessCookie.HttpOnly)
	require.True(t, accessCookie.Secure)
	require.Equal(t, http.SameSiteLaxMode, accessCookie.SameSite)
	require.Equal(t, 15*60, accessCookie.MaxAge)

	refreshCookie := findCookie(cookies, "tokengate_refresh_token")
	require.NotNil(t, refreshCookie)
	require.NotEmpty(t, refreshCookie.Value)
	require.Equal(t, "/", refreshCookie.Path)
	require.True(t, refreshCookie.HttpOnly)
	require.True(t, refreshCookie.Secure)
	require.Equal(t, http.SameSiteLaxMode, refreshCookie.SameSite)
	require.Equal(t, 7*24*60*60, refreshCookie.MaxAge)
}

func TestAuthHandlerRefreshTokenUsesHttpOnlyRefreshCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, authService, _, user := newAuthCookieTestHandler(t)
	tokenPair, err := authService.GenerateTokenPair(context.Background(), user, "")
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	c.Request.Header.Set("X-Forwarded-Proto", "https")
	c.Request.AddCookie(&http.Cookie{
		Name:  "tokengate_refresh_token",
		Value: tokenPair.RefreshToken,
	})

	handler.RefreshToken(c)

	require.Equal(t, http.StatusOK, recorder.Code)
	cookies := recorder.Result().Cookies()
	require.NotNil(t, findCookie(cookies, "tokengate_access_token"))
	require.NotNil(t, findCookie(cookies, "tokengate_refresh_token"))
}

func TestAuthHandlerLogoutRevokesRefreshCookieAndClearsAuthCookies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, authService, cache, user := newAuthCookieTestHandler(t)
	tokenPair, err := authService.GenerateTokenPair(context.Background(), user, "")
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	c.Request.Header.Set("X-Forwarded-Proto", "https")
	c.Request.AddCookie(&http.Cookie{
		Name:  "tokengate_refresh_token",
		Value: tokenPair.RefreshToken,
	})

	handler.Logout(c)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.NotEmpty(t, cache.deleted)

	cookies := recorder.Result().Cookies()
	accessCookie := findCookie(cookies, "tokengate_access_token")
	require.NotNil(t, accessCookie)
	require.Equal(t, -1, accessCookie.MaxAge)

	refreshCookie := findCookie(cookies, "tokengate_refresh_token")
	require.NotNil(t, refreshCookie)
	require.Equal(t, -1, refreshCookie.MaxAge)
}
