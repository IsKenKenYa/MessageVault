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
			Token string `json:"token"`
		} `json:"data"`
	}
	decodeJSON(t, res.Body.Bytes(), &envelope)
	if envelope.Data.Token == "" {
		t.Fatal("expected auth token")
	}
	return envelope.Data.Token
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
