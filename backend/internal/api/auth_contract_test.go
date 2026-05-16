package api

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/IsKenKenYa/Commory/backend/internal/config"
	"github.com/IsKenKenYa/Commory/backend/internal/msglayer"
	"github.com/IsKenKenYa/Commory/backend/internal/storage"
)

func TestMobileAuthRegisterUsesCookieOnlyRefreshContract(t *testing.T) {
	handler := newTestMobileHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(mustJSON(t, map[string]string{
		"userName": "cookie-user",
		"email":    "cookie@example.com",
		"password": "passw0rd!",
	})))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("register status = %d, body = %s", res.Code, res.Body.String())
	}

	var envelope struct {
		Data map[string]any `json:"data"`
	}
	decodeJSON(t, res.Body.Bytes(), &envelope)
	if _, ok := envelope.Data["refreshToken"]; ok {
		t.Fatal("expected register response body to omit refreshToken")
	}
	if token := readRefreshCookieValue(t, res); token == "" {
		t.Fatal("expected refresh cookie to be set")
	}
	assertPersistentRefreshCookie(t, res)
}

func TestMobileAuthLoginUsesPersistentRefreshCookie(t *testing.T) {
	handler := newTestMobileHandler(t)
	registerMobileSession(t, handler, "login-cookie-user", "login-cookie@example.com")

	payload := mustJSON(t, map[string]string{
		"userName": "login-cookie-user",
		"password": "passw0rd!",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", res.Code, res.Body.String())
	}
	if token := readRefreshCookieValue(t, res); token == "" {
		t.Fatal("expected refresh cookie to be set on login")
	}
	assertPersistentRefreshCookie(t, res)
}

func TestMobileRefreshRetryAllowsImmediateRetry(t *testing.T) {
	handler := newTestMobileHandler(t)
	session := registerMobileSession(t, handler, "retry-user", "retry@example.com")

	firstRefresh := refreshMobileSession(t, handler, session.RefreshToken)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: refreshCookieName, Value: session.RefreshToken})
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusConflict {
		t.Fatalf("expected concurrent refresh replay to return 409, got %d", res.Code)
	}
	var envelope struct {
		Msg string `json:"msg"`
	}
	decodeJSON(t, res.Body.Bytes(), &envelope)
	if envelope.Msg != "ERR_REFRESH_TOKEN_RETRY" {
		t.Fatalf("expected retry message, got %q", envelope.Msg)
	}

	secondRefresh := refreshMobileSession(t, handler, firstRefresh.RefreshToken)
	if secondRefresh.Token == "" || secondRefresh.RefreshToken == "" {
		t.Fatal("expected immediate retry path to preserve rotated session")
	}
}

func TestAuthHandlerUsesPublicErrors(t *testing.T) {
	handler := newTestMobileHandler(t)

	badJSONReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString("{"))
	badJSONReq.Header.Set("Content-Type", "application/json")
	badJSONRes := httptest.NewRecorder()
	handler.ServeHTTP(badJSONRes, badJSONReq)
	if badJSONRes.Code != http.StatusBadRequest {
		t.Fatalf("bad json status = %d", badJSONRes.Code)
	}
	assertEnvelopeMessage(t, badJSONRes.Body.Bytes(), "ERR_INVALID_REQUEST")

	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(mustJSON(t, map[string]string{
		"userName": "missing-user",
		"password": "passw0rd!",
	})))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRes := httptest.NewRecorder()
	handler.ServeHTTP(loginRes, loginReq)
	if loginRes.Code != http.StatusUnauthorized {
		t.Fatalf("login status = %d", loginRes.Code)
	}
	assertEnvelopeMessage(t, loginRes.Body.Bytes(), "ERR_INVALID_CREDENTIALS")
}

func TestAuthMiddlewareRejectsRevokedSessionAndThrottlesLastSeen(t *testing.T) {
	handler, store := newFileBackedTestServer(t, false)

	user := createAuthTestUser(t, store, "session-user")
	oldSessionID := "session-old"
	recentSessionID := "session-recent"
	createAuthTestSession(t, store, user.ID, oldSessionID, time.Now().UTC().Add(-10*time.Minute))
	createAuthTestSession(t, store, user.ID, recentSessionID, time.Now().UTC().Add(-2*time.Minute))
	revokedSessionID := "session-revoked"
	createAuthTestSession(t, store, user.ID, revokedSessionID, time.Now().UTC().Add(-10*time.Minute))
	if err := store.RevokeSession(context.Background(), revokedSessionID); err != nil {
		t.Fatalf("revoke session: %v", err)
	}

	oldBefore, _ := store.GetSession(context.Background(), oldSessionID)
	recentBefore, _ := store.GetSession(context.Background(), recentSessionID)

	oldReq := httptest.NewRequest(http.MethodGet, "/api/user/info", nil)
	oldReq.Header.Set("Authorization", "Bearer "+buildTestAccessToken(t, "test-secret", user.ID, oldSessionID, time.Now().Add(time.Hour)))
	oldRes := httptest.NewRecorder()
	handler.ServeHTTP(oldRes, oldReq)
	if oldRes.Code != http.StatusOK {
		t.Fatalf("old session status = %d, body = %s", oldRes.Code, oldRes.Body.String())
	}

	recentReq := httptest.NewRequest(http.MethodGet, "/api/user/info", nil)
	recentReq.Header.Set("Authorization", "Bearer "+buildTestAccessToken(t, "test-secret", user.ID, recentSessionID, time.Now().Add(time.Hour)))
	recentRes := httptest.NewRecorder()
	handler.ServeHTTP(recentRes, recentReq)
	if recentRes.Code != http.StatusOK {
		t.Fatalf("recent session status = %d, body = %s", recentRes.Code, recentRes.Body.String())
	}

	revokedReq := httptest.NewRequest(http.MethodGet, "/api/user/info", nil)
	revokedReq.Header.Set("Authorization", "Bearer "+buildTestAccessToken(t, "test-secret", user.ID, revokedSessionID, time.Now().Add(time.Hour)))
	revokedRes := httptest.NewRecorder()
	handler.ServeHTTP(revokedRes, revokedReq)
	if revokedRes.Code != http.StatusUnauthorized {
		t.Fatalf("revoked session status = %d, body = %s", revokedRes.Code, revokedRes.Body.String())
	}
	assertEnvelopeMessage(t, revokedRes.Body.Bytes(), "ERR_UNAUTHORIZED")

	oldAfter, _ := store.GetSession(context.Background(), oldSessionID)
	recentAfter, _ := store.GetSession(context.Background(), recentSessionID)
	if !oldAfter.LastSeenAt.After(oldBefore.LastSeenAt) {
		t.Fatal("expected stale session last_seen_at to be updated")
	}
	if !recentAfter.LastSeenAt.Equal(recentBefore.LastSeenAt) {
		t.Fatal("expected recent session last_seen_at to remain unchanged inside throttle window")
	}
}

func TestUserInfoMissingUserUsesUnauthorizedEnvelope(t *testing.T) {
	handler, store := newFileBackedTestServer(t, false)

	missingUserID := "user_missing"
	sessionID := "session-missing-user"
	lastSeenAt := time.Now().UTC().Add(-10 * time.Minute)
	createAuthTestSession(t, store, missingUserID, sessionID, lastSeenAt)

	req := httptest.NewRequest(http.MethodGet, "/api/user/info", nil)
	req.Header.Set("Authorization", "Bearer "+buildTestAccessToken(t, "test-secret", missingUserID, sessionID, time.Now().Add(time.Hour)))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("user info status = %d, body = %s", res.Code, res.Body.String())
	}
	assertEnvelopeMessage(t, res.Body.Bytes(), "ERR_UNAUTHORIZED")
	if strings.Contains(res.Body.String(), "not found") {
		t.Fatalf("expected user info response to hide storage error, got %s", res.Body.String())
	}
}

func TestPasskeyInvalidResponseUsesPublicError(t *testing.T) {
	handler, store := newFileBackedTestServer(t, true)
	user := createAuthTestUser(t, store, "passkey-user")
	sessionID := "session-passkey"
	createAuthTestSession(t, store, user.ID, sessionID, time.Now().UTC().Add(-10*time.Minute))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/passkey/register/finish", bytes.NewReader(mustJSON(t, map[string]any{
		"challengeId": "chal_missing",
		"response":    map[string]any{},
	})))
	req.Header.Set("Authorization", "Bearer "+buildTestAccessToken(t, "test-secret", user.ID, sessionID, time.Now().Add(time.Hour)))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("passkey finish status = %d, body = %s", res.Code, res.Body.String())
	}
	assertEnvelopeMessage(t, res.Body.Bytes(), "ERR_PASSKEY_INVALID_RESPONSE")
}

func newFileBackedTestServer(t *testing.T, withPasskey bool) (http.Handler, storage.Provider) {
	t.Helper()
	store, err := storage.NewFileStoreProvider(filepath.Join(t.TempDir(), "commory.json"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	if err := store.Init(context.Background()); err != nil {
		t.Fatal(err)
	}
	validator, err := msglayer.NewValidator(repoPath("msglayer", "schema", "v0.1", "root.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{AuthSecret: "test-secret", TLS: false}
	if withPasskey {
		cfg.PasskeyRPName = "Commory"
		cfg.PasskeyRPID = "localhost"
		cfg.PasskeyOrigin = "http://localhost"
	}
	server := NewServer(cfg, store, validator)
	return server.Handler(), store
}

func createAuthTestUser(t *testing.T, store storage.Provider, userName string) storage.UserRecord {
	t.Helper()
	user := storage.UserRecord{
		ID:           "user_" + userName,
		UserName:     userName,
		Email:        userName + "@example.com",
		PasswordHash: "hash",
		PasswordSalt: "salt",
		Roles:        []string{"R_USER"},
		Buttons:      []string{"view"},
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	if _, err := store.CreateUser(context.Background(), user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func createAuthTestSession(t *testing.T, store storage.Provider, userID, sessionID string, lastSeenAt time.Time) {
	t.Helper()
	if err := store.CreateSession(context.Background(), storage.SessionRecord{
		ID:         sessionID,
		UserID:     userID,
		CreatedAt:  lastSeenAt,
		LastSeenAt: lastSeenAt,
	}); err != nil {
		t.Fatalf("create session: %v", err)
	}
}

func buildTestAccessToken(t *testing.T, secret, userID, sessionID string, expiresAt time.Time) string {
	t.Helper()
	headerJSON, err := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		t.Fatal(err)
	}
	payloadJSON, err := json.Marshal(map[string]any{
		"sub": userID,
		"sid": sessionID,
		"exp": expiresAt.Unix(),
	})
	if err != nil {
		t.Fatal(err)
	}
	header := base64.RawURLEncoding.EncodeToString(headerJSON)
	payload := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signingInput := header + "." + payload
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(signingInput))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return signingInput + "." + signature
}

func assertEnvelopeMessage(t *testing.T, body []byte, want string) {
	t.Helper()
	var envelope struct {
		Msg string `json:"msg"`
	}
	decodeJSON(t, body, &envelope)
	if envelope.Msg != want {
		t.Fatalf("expected msg %q, got %q", want, envelope.Msg)
	}
}

func assertPersistentRefreshCookie(t *testing.T, res *httptest.ResponseRecorder) {
	t.Helper()
	cookieHeader := res.Header().Get("Set-Cookie")
	if !strings.Contains(cookieHeader, "Max-Age=604800") {
		t.Fatalf("expected persistent refresh cookie Max-Age, got %q", cookieHeader)
	}
	if !strings.Contains(cookieHeader, "Expires=") {
		t.Fatalf("expected persistent refresh cookie Expires, got %q", cookieHeader)
	}
}
