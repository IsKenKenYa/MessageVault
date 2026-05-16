package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/IsKenKenYa/Commory/backend/internal/config"
	"github.com/IsKenKenYa/Commory/backend/internal/msglayer"
	"github.com/IsKenKenYa/Commory/backend/internal/storage"
)

func TestMobileAuthUploadListAndExportContract(t *testing.T) {
	handler := newTestMobileHandler(t)

	token := registerMobileUser(t, handler, "mobile-user", "mobile@example.com")
	raw := readFixture(t, "msglayer", "examples", "export.sms-call.json")

	uploadReq := httptest.NewRequest(http.MethodPost, "/api/imports/upload", bytes.NewReader(raw))
	uploadReq.Header.Set("Authorization", "Bearer "+token)
	uploadReq.Header.Set("Content-Type", "application/json")
	uploadRes := httptest.NewRecorder()
	handler.ServeHTTP(uploadRes, uploadReq)
	if uploadRes.Code != http.StatusCreated {
		t.Fatalf("upload status = %d, body = %s", uploadRes.Code, uploadRes.Body.String())
	}
	var uploadEnvelope struct {
		Code int `json:"code"`
		Data struct {
			ImportID string `json:"import_id"`
		} `json:"data"`
	}
	decodeJSON(t, uploadRes.Body.Bytes(), &uploadEnvelope)
	if uploadEnvelope.Data.ImportID == "" {
		t.Fatal("expected import_id in mobile upload response")
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/imports", nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listRes := httptest.NewRecorder()
	handler.ServeHTTP(listRes, listReq)
	if listRes.Code != http.StatusOK {
		t.Fatalf("list status = %d, body = %s", listRes.Code, listRes.Body.String())
	}
	var listEnvelope struct {
		Code int                     `json:"code"`
		Data []storage.ImportSummary `json:"data"`
	}
	decodeJSON(t, listRes.Body.Bytes(), &listEnvelope)
	if len(listEnvelope.Data) != 1 {
		t.Fatalf("expected one import, got %d", len(listEnvelope.Data))
	}

	exportReq := httptest.NewRequest(http.MethodGet, "/api/imports/"+uploadEnvelope.Data.ImportID+"/export", nil)
	exportReq.Header.Set("Authorization", "Bearer "+token)
	exportRes := httptest.NewRecorder()
	handler.ServeHTTP(exportRes, exportReq)
	if exportRes.Code != http.StatusOK {
		t.Fatalf("export status = %d, body = %s", exportRes.Code, exportRes.Body.String())
	}
	if !bytes.Contains(exportRes.Body.Bytes(), []byte(`"version"`)) {
		t.Fatal("expected raw MsgLayer export body")
	}
}

func TestMobileImportExportRequiresOwningUser(t *testing.T) {
	handler := newTestMobileHandler(t)
	ownerToken := registerMobileUser(t, handler, "owner", "owner@example.com")
	otherToken := registerMobileUser(t, handler, "other", "other@example.com")

	raw := readFixture(t, "msglayer", "examples", "export.minimal.json")
	uploadReq := httptest.NewRequest(http.MethodPost, "/api/imports/upload", bytes.NewReader(raw))
	uploadReq.Header.Set("Authorization", "Bearer "+ownerToken)
	uploadReq.Header.Set("Content-Type", "application/json")
	uploadRes := httptest.NewRecorder()
	handler.ServeHTTP(uploadRes, uploadReq)
	if uploadRes.Code != http.StatusCreated {
		t.Fatalf("upload status = %d, body = %s", uploadRes.Code, uploadRes.Body.String())
	}
	var uploadEnvelope struct {
		Data struct {
			ImportID string `json:"import_id"`
		} `json:"data"`
	}
	decodeJSON(t, uploadRes.Body.Bytes(), &uploadEnvelope)

	exportReq := httptest.NewRequest(http.MethodGet, "/api/imports/"+uploadEnvelope.Data.ImportID+"/export", nil)
	exportReq.Header.Set("Authorization", "Bearer "+otherToken)
	exportRes := httptest.NewRecorder()
	handler.ServeHTTP(exportRes, exportReq)
	if exportRes.Code != http.StatusNotFound {
		t.Fatalf("expected non-owner export to be hidden, got %d", exportRes.Code)
	}
}

func TestMobileSetupProbeContract(t *testing.T) {
	handler := newTestMobileHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/setup", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("setup status = %d, body = %s", res.Code, res.Body.String())
	}
	var envelope struct {
		Data struct {
			DatabaseType string `json:"database_type"`
		} `json:"data"`
	}
	decodeJSON(t, res.Body.Bytes(), &envelope)
	if envelope.Data.DatabaseType == "" {
		t.Fatal("expected setup status database_type")
	}
}

func TestMobileRefreshRotatesRefreshToken(t *testing.T) {
	handler := newTestMobileHandler(t)
	session := registerMobileSession(t, handler, "refresh-user", "refresh@example.com")

	firstRefresh := refreshMobileSession(t, handler, session.RefreshToken)
	if firstRefresh.RefreshToken == "" || firstRefresh.RefreshToken == session.RefreshToken {
		t.Fatal("expected refresh token rotation")
	}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewReader(mustJSON(t, map[string]string{
		"refreshToken": session.RefreshToken,
	})))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected old refresh token to be rejected, got %d", res.Code)
	}
	var envelope struct {
		Msg string `json:"msg"`
	}
	decodeJSON(t, res.Body.Bytes(), &envelope)
	if envelope.Msg != "ERR_REFRESH_TOKEN_REPLAYED" {
		t.Fatalf("expected replay error message, got %q", envelope.Msg)
	}
}

func TestMobileLogoutRevokesRefreshToken(t *testing.T) {
	handler := newTestMobileHandler(t)
	session := registerMobileSession(t, handler, "logout-user", "logout@example.com")

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/auth/logout", bytes.NewReader(mustJSON(t, map[string]string{
		"refreshToken": session.RefreshToken,
	})))
	logoutReq.Header.Set("Content-Type", "application/json")
	logoutRes := httptest.NewRecorder()
	handler.ServeHTTP(logoutRes, logoutReq)
	if logoutRes.Code != http.StatusOK {
		t.Fatalf("logout status = %d, body = %s", logoutRes.Code, logoutRes.Body.String())
	}

	refreshReq := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewReader(mustJSON(t, map[string]string{
		"refreshToken": session.RefreshToken,
	})))
	refreshReq.Header.Set("Content-Type", "application/json")
	refreshRes := httptest.NewRecorder()
	handler.ServeHTTP(refreshRes, refreshReq)
	if refreshRes.Code != http.StatusUnauthorized {
		t.Fatalf("expected revoked refresh token to fail, got %d", refreshRes.Code)
	}
}

func TestMobileSetupInitializeRejectsRepeat(t *testing.T) {
	handler := newTestMobileHandler(t)
	payload := mustJSON(t, map[string]string{
		"userName":        "admin",
		"password":        "passw0rd!",
		"confirmPassword": "passw0rd!",
		"usageMode":       "personal",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/setup", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("setup init status = %d, body = %s", res.Code, res.Body.String())
	}

	repeatReq := httptest.NewRequest(http.MethodPost, "/api/setup", bytes.NewReader(payload))
	repeatReq.Header.Set("Content-Type", "application/json")
	repeatRes := httptest.NewRecorder()
	handler.ServeHTTP(repeatRes, repeatReq)
	if repeatRes.Code != http.StatusBadRequest {
		t.Fatalf("expected repeated setup to fail, got %d", repeatRes.Code)
	}
}

func TestMobilePasskeyEnrollmentRoutesRequireAuth(t *testing.T) {
	handler := newTestMobileHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/passkey/register/begin", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthenticated passkey begin to fail with 401, got %d", res.Code)
	}

	session := registerMobileSession(t, handler, "passkey-user", "passkey@example.com")
	authReq := httptest.NewRequest(http.MethodPost, "/api/auth/passkey/register/begin", nil)
	authReq.Header.Set("Authorization", "Bearer "+session.Token)
	authRes := httptest.NewRecorder()
	handler.ServeHTTP(authRes, authReq)
	if authRes.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected authenticated request to reach passkey handler, got %d", authRes.Code)
	}
}

func TestMobileSessionsDeleteOthersKeepsCurrentSession(t *testing.T) {
	handler := newTestMobileHandler(t)
	current := registerMobileSession(t, handler, "session-user", "session@example.com")
	other := loginMobileSession(t, handler, "session-user")

	sessions := listMobileSessions(t, handler, current.Token)
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}
	currentCount := 0
	for _, session := range sessions {
		if session.Current {
			currentCount++
		}
	}
	if currentCount != 1 {
		t.Fatalf("expected exactly one current session marker, got %d", currentCount)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/sessions", nil)
	req.Header.Set("Authorization", "Bearer "+current.Token)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("delete others status = %d, body = %s", res.Code, res.Body.String())
	}

	if _, _, code := tryRefreshMobileSession(t, handler, other.RefreshToken); code != http.StatusUnauthorized {
		t.Fatalf("expected revoked secondary session refresh to fail, got %d", code)
	}
	if _, refreshed, code := tryRefreshMobileSession(t, handler, current.RefreshToken); code != http.StatusOK || refreshed.RefreshToken == "" {
		t.Fatalf("expected current session refresh to survive, got status=%d refresh=%q", code, refreshed.RefreshToken)
	}
}

func TestMobileSessionDeleteRequiresOwnership(t *testing.T) {
	handler := newTestMobileHandler(t)
	owner := registerMobileSession(t, handler, "owner-user", "owner-user@example.com")
	other := registerMobileSession(t, handler, "other-user", "other-user@example.com")

	otherSessions := listMobileSessions(t, handler, other.Token)
	if len(otherSessions) != 1 {
		t.Fatalf("expected 1 other session, got %d", len(otherSessions))
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/sessions/"+otherSessions[0].ID, nil)
	req.Header.Set("Authorization", "Bearer "+owner.Token)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusNotFound {
		t.Fatalf("expected cross-user revoke to be hidden, got %d", res.Code)
	}
}

func newTestMobileHandler(t *testing.T) http.Handler {
	t.Helper()
	store, err := storage.NewSQLiteProvider(filepath.Join(t.TempDir(), "commory.json"))
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
	server := NewServer(config.Config{AuthSecret: "test-secret", TLS: false}, store, validator)
	return server.Handler()
}

func registerMobileUser(t *testing.T, handler http.Handler, userName, email string) string {
	t.Helper()
	return registerMobileSession(t, handler, userName, email).Token
}

type mobileAuthSession struct {
	Token        string
	RefreshToken string
}

type mobileSessionRecord struct {
	ID      string `json:"id"`
	Current bool   `json:"current"`
}

func registerMobileSession(t *testing.T, handler http.Handler, userName, email string) mobileAuthSession {
	t.Helper()
	body := map[string]string{
		"userName": userName,
		"email":    email,
		"password": "passw0rd!",
	}
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("register status = %d, body = %s", res.Code, res.Body.String())
	}
	var envelope struct {
		Data struct {
			Token        string `json:"token"`
			RefreshToken string `json:"refreshToken"`
		} `json:"data"`
	}
	decodeJSON(t, res.Body.Bytes(), &envelope)
	if envelope.Data.Token == "" || envelope.Data.RefreshToken == "" {
		t.Fatal("expected auth token pair")
	}
	return mobileAuthSession{Token: envelope.Data.Token, RefreshToken: envelope.Data.RefreshToken}
}

func refreshMobileSession(t *testing.T, handler http.Handler, refreshToken string) mobileAuthSession {
	t.Helper()
	res, session, code := tryRefreshMobileSession(t, handler, refreshToken)
	if code != http.StatusOK {
		t.Fatalf("refresh status = %d, body = %s", code, res)
	}
	return session
}

func tryRefreshMobileSession(t *testing.T, handler http.Handler, refreshToken string) (string, mobileAuthSession, int) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewReader(mustJSON(t, map[string]string{
		"refreshToken": refreshToken,
	})))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	var envelope struct {
		Data struct {
			Token        string `json:"token"`
			RefreshToken string `json:"refreshToken"`
		} `json:"data"`
	}
	_ = json.Unmarshal(res.Body.Bytes(), &envelope)
	return res.Body.String(), mobileAuthSession{
		Token:        envelope.Data.Token,
		RefreshToken: envelope.Data.RefreshToken,
	}, res.Code
}

func loginMobileSession(t *testing.T, handler http.Handler, userName string) mobileAuthSession {
	t.Helper()
	payload := mustJSON(t, map[string]string{
		"userName": userName,
		"password": "passw0rd!",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", res.Code, res.Body.String())
	}
	var envelope struct {
		Data struct {
			Token        string `json:"token"`
			RefreshToken string `json:"refreshToken"`
		} `json:"data"`
	}
	decodeJSON(t, res.Body.Bytes(), &envelope)
	if envelope.Data.Token == "" || envelope.Data.RefreshToken == "" {
		t.Fatal("expected login token pair")
	}
	return mobileAuthSession{Token: envelope.Data.Token, RefreshToken: envelope.Data.RefreshToken}
}

func listMobileSessions(t *testing.T, handler http.Handler, token string) []mobileSessionRecord {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("list sessions status = %d, body = %s", res.Code, res.Body.String())
	}
	var envelope struct {
		Data []mobileSessionRecord `json:"data"`
	}
	decodeJSON(t, res.Body.Bytes(), &envelope)
	return envelope.Data
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func readFixture(t *testing.T, parts ...string) []byte {
	t.Helper()
	data, err := os.ReadFile(repoPath(parts...))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func decodeJSON(t *testing.T, data []byte, target any) {
	t.Helper()
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatalf("decode json: %v\n%s", err, string(data))
	}
}

func repoPath(parts ...string) string {
	_, file, _, _ := runtime.Caller(0)
	base := filepath.Join(filepath.Dir(file), "..", "..", "..")
	all := append([]string{base}, parts...)
	return filepath.Join(all...)
}
